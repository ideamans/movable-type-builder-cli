package cmd

import (
	"net/url"
	"strings"
	"testing"
)

func TestListBuildsQuery(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"totalResults":1,"items":[{"id":"1"}]}`))
	out := h.mustRun("sites", "entries", "list", "--site-id", "3", "--limit", "5", "--status", "Publish")

	sent := h.last()
	if sent.Method != "GET" {
		t.Errorf("method = %s", sent.Method)
	}
	u, err := url.Parse(sent.URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/mt/mt-data-api.cgi/v7/sites/3/entries" {
		t.Errorf("path = %s", u.Path)
	}
	if u.Query().Get("limit") != "5" || u.Query().Get("status") != "Publish" {
		t.Errorf("query = %v", u.Query())
	}
	// 指定していないパラメータは送らない（MT 側の既定に任せる）
	if _, ok := u.Query()["offset"]; ok {
		t.Error("指定していない offset が送られています")
	}
	if !strings.Contains(out, `"totalResults": 1`) {
		t.Errorf("整形された JSON になっていません: %s", out)
	}
}

func TestPositionalArgsFillPathParams(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"42"}`))
	h.mustRun("sites", "entries", "get", "3", "42")

	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/3/entries/42" {
		t.Errorf("path = %s", got)
	}
}

func TestProfileDefaultSiteIDIsUsed(t *testing.T) {
	h := newHarness(t)
	if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "7")); err != nil {
		t.Fatal(err)
	}
	h.reply(okJSON(`{"totalResults":0,"items":[]}`))
	h.mustRun("sites", "entries", "list")

	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/7/entries" {
		t.Errorf("path = %s（既定サイトIDが効いていません）", got)
	}
}

func TestExplicitSiteIDBeatsProfileDefault(t *testing.T) {
	h := newHarness(t)
	if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "7")); err != nil {
		t.Fatal(err)
	}
	h.reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1")

	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/1/entries" {
		t.Errorf("path = %s", got)
	}
}

func TestMissingPathParamIsExplained(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "entries", "list")
	if err == nil {
		t.Fatal("サイトIDが無いのに成功しています")
	}
	for _, want := range []string{"--site-id", "mtb profile set --site-id"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("エラーに %q がありません: %v", want, err)
		}
	}
}

func TestCreateSendsJSONBody(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"1"}`))
	h.mustRun("sites", "entries", "create", "--site-id", "1",
		"--entry", `{"title":"新しい記事","body":"本文"}`, "--publish", "0")

	sent := h.last()
	if sent.Method != "POST" {
		t.Errorf("method = %s", sent.Method)
	}
	values, err := url.ParseQuery(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("entry") != `{"title":"新しい記事","body":"本文"}` {
		t.Errorf("entry = %q", values.Get("entry"))
	}
	if values.Get("publish") != "0" {
		t.Errorf("publish = %q", values.Get("publish"))
	}
}

// JSON の書き間違いは MT の 400 より手元で気づく方が直しやすい。
func TestInvalidJSONBodyIsRejectedLocally(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "entries", "create", "--site-id", "1", "--entry", `{title: 新しい記事}`)
	if err == nil {
		t.Fatal("不正な JSON が通っています")
	}
	if !strings.Contains(err.Error(), "JSON") {
		t.Errorf("エラーが分かりにくいです: %v", err)
	}
	if len(h.doer.sent) != 0 {
		t.Error("不正な JSON なのに送信しています")
	}
}

func TestJSONBodyFromFile(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"1"}`))
	h.files["entry.json"] = `{"title":"ファイルから"}`
	h.mustRun("sites", "entries", "create", "--site-id", "1", "--entry", "@entry.json")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("entry") != `{"title":"ファイルから"}` {
		t.Errorf("entry = %q", values.Get("entry"))
	}
}

func TestUploadSendsMultipart(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"1"}`))
	h.files["photo.jpg"] = "binary"
	h.mustRun("sites", "assets", "upload", "--site-id", "1", "--file", "photo.jpg", "--path", "/2026")

	sent := h.last()
	if !strings.HasPrefix(sent.Header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("Content-Type = %q", sent.Header.Get("Content-Type"))
	}
	if !strings.Contains(sent.Body, `filename="photo.jpg"`) {
		t.Errorf("ファイル名が入っていません: %s", sent.Body)
	}
	if !strings.Contains(sent.Body, "binary") {
		t.Error("ファイルの中身が入っていません")
	}
}

// 末尾のIDがあってもなくても同じ操作、というエンドポイントを1コマンドに畳んでいる。
func TestMergedOperationsPickByArgument(t *testing.T) {
	t.Run("IDあり", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{}`))
		h.mustRun("plugins", "enable", "MyPlugin")
		if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/plugins/MyPlugin/enable" {
			t.Errorf("path = %s", got)
		}
	})
	t.Run("IDなし", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{}`))
		h.mustRun("plugins", "enable")
		if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/plugins/enable" {
			t.Errorf("path = %s", got)
		}
	})
	t.Run("記事プレビューはIDの有無で別のエンドポイント", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{}`), okJSON(`{}`))
		h.mustRun("sites", "entries", "preview", "1", "--entry", "{}")
		if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/1/entries/preview" {
			t.Errorf("path = %s", got)
		}
		h.mustRun("sites", "entries", "preview", "1", "42", "--entry", "{}")
		if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/1/entries/42/preview" {
			t.Errorf("path = %s", got)
		}
	})
}

// DELETE は対話できない環境では黙って実行しない。
func TestDeleteRequiresConfirmation(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "entries", "delete", "1", "42")
	if err == nil {
		t.Fatal("確認なしで削除が通っています")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("--yes を案内すべきです: %v", err)
	}
	// どのサーバに向けた削除なのかが確認に出ること。prod と stg の両方に
	// サインインしたまま作業していると、打つコマンドは同じなので、
	// ここに出ている文字列だけが取り違えに気づける手がかりになる。
	for _, want := range []string{"prod", "https://example.com/mt/mt-data-api.cgi"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("確認に接続先 %q が出ていません: %v", want, err)
		}
	}
	if len(h.doer.sent) != 0 {
		t.Error("確認前に送信しています")
	}
}

func TestDeleteWithYes(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"42"}`))
	h.mustRun("sites", "entries", "delete", "1", "42", "--yes")
	if h.last().Method != "DELETE" {
		t.Errorf("method = %s", h.last().Method)
	}
}

func TestDeleteInteractiveDeclined(t *testing.T) {
	h := newHarness(t)
	h.prompt.NotInteractive = false
	h.prompt.Confirms = []bool{false}
	err := h.run("sites", "entries", "delete", "1", "42")
	if err == nil || !strings.Contains(err.Error(), "中止") {
		t.Fatalf("err = %v", err)
	}
	if len(h.doer.sent) != 0 {
		t.Error("断ったのに送信しています")
	}
}

func TestDryRunDoesNotSend(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("sites", "entries", "create", "--site-id", "1", "--entry", `{"title":"x"}`, "--dry-run")

	if len(h.doer.sent) != 0 {
		t.Error("--dry-run なのに送信しています")
	}
	for _, want := range []string{"POST", "/v7/sites/1/entries", `entry={"title":"x"}`} {
		if !strings.Contains(out, want) {
			t.Errorf("出力に %q がありません: %s", want, out)
		}
	}
}

func TestDryRunOnDeleteMentionsYes(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("sites", "entries", "delete", "1", "42", "--dry-run")
	if !strings.Contains(out, "--yes") {
		t.Errorf("出力に --yes の案内がありません: %s", out)
	}
}

func TestExtraQueryAndForm(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1", "--query", "customField=1", "--query", "another=2")

	u, err := url.Parse(h.last().URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("customField") != "1" || u.Query().Get("another") != "2" {
		t.Errorf("query = %v", u.Query())
	}
}

func TestExtraPairMustBeKeyValue(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "entries", "list", "--site-id", "1", "--query", "broken")
	if err == nil || !strings.Contains(err.Error(), "key=value") {
		t.Fatalf("err = %v", err)
	}
}

func TestMethodOverrideFlagReachesClient(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "delete", "1", "42", "--yes", "--method-override")
	sent := h.last()
	if sent.Method != "POST" || !strings.Contains(sent.Body, "__method=DELETE") {
		t.Errorf("%s body=%q", sent.Method, sent.Body)
	}
}

// /assets/upload はサイトIDをパスではなく本文で受ける。ここでも既定サイトを効かせる。
func TestDefaultSiteIDFillsBodyParam(t *testing.T) {
	h := newHarness(t)
	if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "7")); err != nil {
		t.Fatal(err)
	}
	h.files["a.png"] = "x"
	h.reply(okJSON(`{}`))
	h.mustRun("assets", "upload", "--file", "a.png")

	if !strings.Contains(h.last().Body, "site_id") || !strings.Contains(h.last().Body, "7") {
		t.Errorf("本文に site_id が入っていません: %s", h.last().Body)
	}
}

func TestAPIErrorCarriesHint(t *testing.T) {
	h := newHarness(t).reply(stub{status: 403, body: `{"error":{"code":403,"message":"Do not have permission"}}`, contentType: "application/json"})
	err := h.run("sites", "entries", "list", "--site-id", "1")
	if err == nil {
		t.Fatal("エラーになるはずです")
	}
	if !strings.Contains(err.Error(), "権限") {
		t.Errorf("助言が足りません: %v", err)
	}
}

func pathOf(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u.Path
}

// --form / --query の値も @ファイル名 で渡せること。
// 長い JSON をシェルに直接書かずに済ませるための逃げ道で、mtb api と
// 挙動を揃えてある（揃っていないと「api では通るのに」で詰まる）。
func TestExtraFormReadsFile(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":2}`))
	h.files["blog.json"] = `{"url":"blog/"}`
	h.mustRun("sites", "update", "2", "--form", "blog=@blog.json")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("blog") != `{"url":"blog/"}` {
		t.Errorf("blog = %q", values.Get("blog"))
	}
}

// 公式定義に無い本文パラメータを補正しているので、--blog が生えていること。
func TestSiteUpdateHasBlogFlag(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":2}`))
	h.files["blog.json"] = `{"url":"blog/"}`
	h.mustRun("sites", "update", "2", "--blog", "@blog.json")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("blog") != `{"url":"blog/"}` {
		t.Errorf("blog = %q", values.Get("blog"))
	}
}

func TestExtraFormMissingFileIsUsageError(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "update", "2", "--form", "blog=@missing.json")
	if err == nil {
		t.Fatal("エラーになるはずです")
	}
	if got := ExitCode(err); got != ExitUsage {
		t.Errorf("ExitCode = %d, want %d", got, ExitUsage)
	}
	if len(h.doer.sent) != 0 {
		t.Error("読めないファイルなのに送信しています")
	}
}

// パスに入る値が数値でないと、URL は組み立てられて 404 になる。
// 原因が値の間違いだと分かりにくいので手元で弾く。
func TestNonIntegerPathParamIsRejected(t *testing.T) {
	h := newHarness(t)
	err := h.run("sites", "entries", "get", "1 40", "42")
	if err == nil {
		t.Fatal("数値でない site-id が通っています")
	}
	if got := ExitCode(err); got != ExitUsage {
		t.Errorf("ExitCode = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(err.Error(), "--site-id") {
		t.Errorf("どのフラグかが分かりません: %v", err)
	}
	if len(h.doer.sent) != 0 {
		t.Error("送信しています")
	}
}

// 文字列のパスパラメータ（プラグインIDなど）は数値でなくてよい。
func TestStringPathParamAllowsNonInteger(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("plugins", "get", "MyPlugin")
	if !strings.HasSuffix(pathOf(t, h.last().URL), "/plugins/MyPlugin") {
		t.Errorf("path = %s", h.last().URL)
	}
}

// 空の更新は「何もしない」ではない。MT は指定しなかった関連を外すことがある。
func TestEmptyUpdateWarns(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":2}`))
	h.mustRun("sites", "pages", "update", "1", "2", "--page", "{}")

	if !strings.Contains(h.errOut.String(), "空のオブジェクト") {
		t.Errorf("警告が出ていません: %q", h.errOut.String())
	}
	// 警告は出すが、送信は止めない（意図して送る場面もある）
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d", len(h.doer.sent))
	}
}

// 作成時の空オブジェクトは MT のプレビューが要求する正規の使い方なので黙る。
func TestEmptyCreateDoesNotWarn(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "preview", "1", "42", "--entry", "{}")
	if strings.Contains(h.errOut.String(), "空のオブジェクト") {
		t.Errorf("余計な警告が出ています: %q", h.errOut.String())
	}
}

// 入れ物ごと壊す操作は --yes だけでは実行させない。
// 総当たりテストで sites delete が --yes ひとつでサイトを丸ごと消し、
// 記事もテンプレートも巻き添えになった。
func TestSweepingOperationNeedsConfirm(t *testing.T) {
	t.Run("--confirm が無ければ送らない", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{"id":2,"name":"アイデアマンズ ブログ"}`))
		err := h.run("sites", "delete", "2", "--yes")
		if err == nil {
			t.Fatal("確認なしで消せています")
		}
		if got := ExitCode(err); got != ExitUsage {
			t.Errorf("ExitCode = %d, want %d", got, ExitUsage)
		}
		// 何が失われるかと、打つべき名前が示されること
		for _, want := range []string{"--confirm", "アイデアマンズ ブログ", "記事"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("エラーに %q がありません: %v", want, err)
			}
		}
		// 名前を引く GET だけで、DELETE は送っていない
		if len(h.doer.sent) != 1 || h.doer.sent[0].Method != "GET" {
			t.Errorf("送信内容 = %+v", h.doer.sent)
		}
	})

	t.Run("名前が違えば送らない", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{"id":2,"name":"アイデアマンズ ブログ"}`))
		err := h.run("sites", "delete", "2", "--yes", "--confirm", "べつのサイト")
		if err == nil || ExitCode(err) != ExitUsage {
			t.Fatalf("err = %v", err)
		}
		if len(h.doer.sent) != 1 {
			t.Errorf("DELETE を送っています: %+v", h.doer.sent)
		}
	})

	t.Run("名前が合えば実行する", func(t *testing.T) {
		h := newHarness(t).reply(
			okJSON(`{"id":2,"name":"アイデアマンズ ブログ"}`),
			okJSON(`{"id":2}`))
		h.mustRun("sites", "delete", "2", "--yes", "--confirm", "アイデアマンズ ブログ")
		if h.last().Method != "DELETE" {
			t.Errorf("method = %s", h.last().Method)
		}
	})

	t.Run("--dry-run は名前を聞かない", func(t *testing.T) {
		h := newHarness(t)
		out := h.mustRun("sites", "delete", "2", "--dry-run")
		if len(h.doer.sent) != 0 {
			t.Error("送信しています")
		}
		if !strings.Contains(out, "DELETE") {
			t.Errorf("出力 = %s", out)
		}
	})

	// 1件の削除まで重くしない。普通の delete は --yes だけで通る。
	t.Run("普通の削除は --yes だけ", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{"id":42}`))
		h.mustRun("sites", "entries", "delete", "1", "42", "--yes")
		if h.last().Method != "DELETE" {
			t.Errorf("method = %s", h.last().Method)
		}
		if cmd, _, err := New(h.env).Root().Find([]string{"sites", "entries", "delete"}); err == nil {
			if cmd.Flags().Lookup(confirmFlag) != nil {
				t.Error("普通の削除に --confirm が生えています")
			}
		}
	})
}

// permutate の4つは「並べ替えた後のリソース配列」を取るのに、公式定義では
// string になっている。string のままだと @ファイル名 が展開されず、
// リテラルが MT に届いて 500 malformed JSON string になる。
func TestPermutateAcceptsFileInput(t *testing.T) {
	h := newHarness(t).reply(okJSON(`[{"id":11}]`))
	h.files["folders.json"] = `[{"id":11,"label":"会社情報"}]`
	h.mustRun("sites", "folders", "permutate", "1", "--folders", "@folders.json")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("folders") != `[{"id":11,"label":"会社情報"}]` {
		t.Errorf("folders = %q（ファイルが展開されていません）", values.Get("folders"))
	}
}

// 文字列の項目に @ファイル名 を渡しても展開しない。ただし本当にファイルが
// あるなら渡し方の間違いなので教える。
func TestWarnsWhenStringFlagLooksLikeFile(t *testing.T) {
	t.Run("本文の文字列項目", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{}`))
		h.files["note.txt"] = "x"
		h.mustRun("sites", "assets", "upload", "1", "--file", "note.txt", "--path", "@note.txt")
		if !strings.Contains(h.errOut.String(), "--path") {
			t.Errorf("警告が出ていません: %q", h.errOut.String())
		}
	})

	t.Run("クエリの文字列項目", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{}`))
		h.files["note.txt"] = "x"
		h.mustRun("sites", "entries", "list", "1", "--search", "@note.txt")
		if !strings.Contains(h.errOut.String(), "--search") {
			t.Errorf("警告が出ていません: %q", h.errOut.String())
		}
	})

	// MT には @SITE_LOGO のようなタグ記法があるので、ファイルが無ければ黙る。
	t.Run("MTのタグ記法には反応しない", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{}`))
		h.mustRun("sites", "entries", "list", "1", "--search", "@ADD_TO_SITE_NAV")
		if h.errOut.Len() != 0 {
			t.Errorf("余計な警告が出ています: %q", h.errOut.String())
		}
	})
}
