package cmd

import (
	"net/url"
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// logsStub はログ一覧の応答を作ります。
func logsStub(ids ...string) stub {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, `{"id":`+id+`,"level":"ERROR","message":"Error during publishing: `+id+`"}`)
	}
	return okJSON(`{"totalResults":` + itoa(len(ids)) + `,"items":[` + strings.Join(items, ",") + `]}`)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// --verify-publish を付けると、送信の前後でエラーログを見比べる。
// 増えていれば「保存はできたが書き出せていない」として終了コード 8。
func TestVerifyPublishReportsNewErrors(t *testing.T) {
	h := newHarness(t)
	h.reply(
		logsStub("100"),                  // 送信前の目印
		okJSON(`{"id":"7","title":"x"}`), // 記事の更新そのもの
		logsStub("102", "101", "100"),    // 送信後（新しい順）
	)

	err := h.run("sites", "entries", "update", "4", "7", "--entry", `{"title":"x"}`, "--verify-publish")
	if err == nil {
		t.Fatal("書き出しに失敗しているのにエラーになっていません")
	}
	if got := ExitCode(err); got != ExitPublishFailed {
		t.Errorf("終了コード = %d, want %d", got, ExitPublishFailed)
	}
	msg := err.Error()
	for _, want := range []string{"2 件のエラー", "publishing: 101", "publishing: 102"} {
		if !strings.Contains(msg, want) {
			t.Errorf("エラー文に %q がありません:\n%s", want, msg)
		}
	}
	// 起きた順（古い順）に並べ直していること。
	if strings.Index(msg, "publishing: 101") > strings.Index(msg, "publishing: 102") {
		t.Errorf("ログが起きた順に並んでいません:\n%s", msg)
	}
	// 応答そのものは捨てない。作成したもののIDが失われると次に進めなくなる。
	if !strings.Contains(h.out.String(), `"id"`) {
		t.Errorf("応答が出力されていません: %q", h.out.String())
	}
}

// ログが増えていなければ、そのまま成功する。
func TestVerifyPublishSucceedsWhenNoNewErrors(t *testing.T) {
	h := newHarness(t)
	h.reply(
		logsStub("100"),
		okJSON(`{"id":"7"}`),
		logsStub("100"), // 送信前と同じ。増えていない
	)
	if err := h.run("sites", "entries", "update", "4", "7", "--entry", `{"title":"x"}`, "--verify-publish"); err != nil {
		t.Fatalf("増えていないのに失敗しました: %v", err)
	}
	if len(h.doer.sent) != 3 {
		t.Errorf("送信回数 = %d, want 3", len(h.doer.sent))
	}
}

// もともとエラーが1件も無いサイトでも成立する（目印は 0 になる）。
func TestVerifyPublishWorksWithNoPriorErrors(t *testing.T) {
	h := newHarness(t)
	h.reply(
		logsStub(),
		okJSON(`{"id":"7"}`),
		logsStub("1"),
	)
	err := h.run("sites", "entries", "update", "4", "7", "--entry", `{"title":"x"}`, "--verify-publish")
	if ExitCode(err) != ExitPublishFailed {
		t.Fatalf("終了コード = %d (%v), want %d", ExitCode(err), err, ExitPublishFailed)
	}
}

// ログを読めなくても、成功した保存をエラーにはしない。確かめられなかったと伝えるだけ。
func TestVerifyPublishDoesNotFailWhenLogIsUnreadable(t *testing.T) {
	h := newHarness(t)
	h.reply(
		stub{status: 403, body: `{"error":{"code":403,"message":"Do not have permission."}}`, contentType: "application/json"},
		okJSON(`{"id":"7"}`),
	)
	if err := h.run("sites", "entries", "update", "4", "7", "--entry", `{"title":"x"}`, "--verify-publish"); err != nil {
		t.Fatalf("ログが読めないだけで失敗しました: %v", err)
	}
	if !strings.Contains(h.errOut.String(), "確かめられません") {
		t.Errorf("確かめられなかった旨が出ていません: %q", h.errOut.String())
	}
	if len(h.doer.sent) != 2 {
		t.Errorf("送信回数 = %d, want 2（目印の取得と本体だけ）", len(h.doer.sent))
	}
}

// フラグが無ければ、余計なリクエストは1つも増えない。
func TestVerifyPublishOffByDefault(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"id":"7"}`))
	if err := h.run("sites", "entries", "update", "4", "7", "--entry", `{"title":"x"}`); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1", len(h.doer.sent))
	}
}

// 再構築を伴わない操作では、フラグを付けてもログを見に行かない。
func TestVerifyPublishSkipsNonPublishingOperations(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"id":"3"}`))
	if err := h.run("sites", "categories", "create", "4", "--category", `{"label":"x"}`, "--verify-publish"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1（カテゴリの作成は再構築を伴わない）", len(h.doer.sent))
	}
}

// 目印を取るときのクエリが、エラーだけ・新しい順・1件になっていること。
func TestVerifyPublishAsksForNewestErrorOnly(t *testing.T) {
	h := newHarness(t)
	h.reply(logsStub("100"), okJSON(`{"id":"7"}`), logsStub("100"))
	if err := h.run("sites", "entries", "update", "4", "7", "--entry", `{"title":"x"}`, "--verify-publish"); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(h.doer.sent[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(u.Path, "/sites/4/logs") {
		t.Errorf("目印の取得先 = %q", u.Path)
	}
	q := u.Query()
	for k, want := range map[string]string{"level": "error", "sortBy": "created_on", "sortOrder": "descend", "limit": "1"} {
		if q.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, q.Get(k), want)
		}
	}
	// 後から見るときは多めに取る（1件の保存で複数のアーカイブが失敗しうる）。
	u2, err := url.Parse(h.doer.sent[2].URL)
	if err != nil {
		t.Fatal(err)
	}
	if got := u2.Query().Get("limit"); got != itoa(verifyLimit) {
		t.Errorf("確認時の limit = %q, want %q", got, itoa(verifyLimit))
	}
}

// publishingOps の綴りが定義と合っていること。
//
// メソッドとパスの文字列で引いているので、MT 側の定義が変わったりキーを
// 打ち間違えたりすると、**何も起きないまま検査が無効になります**。
func TestPublishingOpsExistInSpec(t *testing.T) {
	s := spec.MustLoad()
	have := map[string]bool{}
	for _, c := range s.Commands {
		for _, op := range c.Ops {
			have[op.Method+" "+op.Path] = true
		}
	}
	for key := range publishingOps {
		if !have[key] {
			t.Errorf("publishingOps の %q は API 定義にありません", key)
		}
	}
}

// 同じく、再構築の操作も定義に実在すること。
func TestRebuildOpsExistInSpec(t *testing.T) {
	s := spec.MustLoad()
	have := map[string]bool{}
	for _, c := range s.Commands {
		for _, op := range c.Ops {
			have[op.Method+" "+op.Path] = true
		}
	}
	for key := range rebuildOps {
		if !have[key] {
			t.Errorf("rebuildOps の %q は API 定義にありません", key)
		}
	}
}
