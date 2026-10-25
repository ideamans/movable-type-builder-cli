package cmd

import (
	"net/url"
	"strings"
	"testing"
)

// rebuildStub は publish entries の応答を作ります。
func rebuildStub(restIDs string) stub {
	return okJSON(`{"status":"Rebuilding","restIds":"` + restIDs + `","startTime":"2026-08-29T06:14:08+00:00"}`)
}

// --wait を付けると restIds が空になるまで呼び直す。
func TestWaitRebuildRepeatsUntilRestIsEmpty(t *testing.T) {
	h := newHarness(t)
	h.reply(
		rebuildStub("3,1,2"),
		rebuildStub("2,3"),
		rebuildStub("3"),
		rebuildStub(""),
	)
	if err := h.run("publish", "entries", "--ids", "1,2,3", "--wait"); err != nil {
		t.Fatalf("終わるはずなのに失敗しました: %v", err)
	}
	if len(h.doer.sent) != 4 {
		t.Fatalf("送信回数 = %d, want 4", len(h.doer.sent))
	}
	// 2回目以降は、残っている ID だけを送る。
	for i, want := range []string{"1,2,3", "1,2,3", "2,3", "3"} {
		u, err := url.Parse(h.doer.sent[i].URL)
		if err != nil {
			t.Fatal(err)
		}
		if got := u.Query().Get("ids"); got != want {
			t.Errorf("%d 回目の ids = %q, want %q", i+1, got, want)
		}
	}
	// 待ちを挟んでいること（呼び直しの回数ぶん）。
	if len(h.slept) != 3 {
		t.Errorf("待ちの回数 = %d, want 3", len(h.slept))
	}
}

// restIds が減らないなら、待っても終わらないので止めて理由を伝える。
func TestWaitRebuildStopsWhenNotProgressing(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < 10; i++ {
		h.reply(rebuildStub("1,2,3"))
	}
	err := h.run("publish", "entries", "--ids", "1,2,3", "--wait")
	if err == nil {
		t.Fatal("進んでいないのにエラーになっていません")
	}
	if got := ExitCode(err); got != ExitPublishFailed {
		t.Errorf("終了コード = %d, want %d", got, ExitPublishFailed)
	}
	for _, want := range []string{"進んでいません", "3 件", "保存し直す"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("エラー文に %q がありません:\n%s", want, err.Error())
		}
	}
	// 足踏みを見きわめるまでで止まること。延々と叩き続けない。
	if len(h.doer.sent) > rebuildStallRounds+2 {
		t.Errorf("送信回数 = %d、足踏みの判定に対して多すぎます", len(h.doer.sent))
	}
	// 止まっても最後の応答は出す。
	if !strings.Contains(h.out.String(), "restIds") {
		t.Errorf("応答が出力されていません: %q", h.out.String())
	}
}

// 並び順が毎回変わっても、件数が減っていれば進んでいると見なす。
// MT は restIds をばらばらの順で返すので、文字列の比較では判定できない。
func TestWaitRebuildIgnoresIDOrder(t *testing.T) {
	h := newHarness(t)
	h.reply(
		rebuildStub("3,1,2"),
		rebuildStub("2,3,1"), // 同じ3件。並びだけ違う
		rebuildStub("1,3,2"), // まだ同じ
		rebuildStub("2,1,3"), // ここで足踏みと判定
	)
	err := h.run("publish", "entries", "--ids", "1,2,3", "--wait")
	if ExitCode(err) != ExitPublishFailed {
		t.Fatalf("並びが違うだけで進んだと誤判定しています: %v", err)
	}
}

// startTime は次の呼び出しに引き継ぐ。
func TestWaitRebuildCarriesStartTime(t *testing.T) {
	h := newHarness(t)
	h.reply(rebuildStub("2"), rebuildStub(""))
	if err := h.run("publish", "entries", "--ids", "1,2", "--wait"); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(h.doer.sent[1].URL)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("startTime"); got != "2026-08-29T06:14:08+00:00" {
		t.Errorf("startTime = %q", got)
	}
}

// --wait を付けなければ1回で終わる（これまでどおり）。
func TestWithoutWaitCallsOnce(t *testing.T) {
	h := newHarness(t)
	h.reply(rebuildStub("1,2,3"))
	if err := h.run("publish", "entries", "--ids", "1,2,3"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1", len(h.doer.sent))
	}
}

// --wait は再構築のコマンドにだけ生やす。
func TestWaitFlagOnlyOnRebuildCommands(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "entries", "list", "4", "--wait")
	if err == nil {
		t.Fatal("--wait が一覧コマンドにも生えています")
	}
	if got := ExitCode(err); got != ExitUsage {
		t.Errorf("終了コード = %d, want %d", got, ExitUsage)
	}
}

// restIds を持たない応答は、繰り返す対象ではないのでそのまま返す。
func TestWaitRebuildPassesThroughUnknownShape(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"status":"Complete"}`))
	if err := h.run("publish", "entries", "--ids", "1", "--wait"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1", len(h.doer.sent))
	}
}
