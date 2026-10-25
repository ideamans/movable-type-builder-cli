package cmd

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestOutputFormats(t *testing.T) {
	const body = `{"totalResults":1,"items":[{"id":"1","title":"記事"}]}`

	t.Run("既定は整形した JSON", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(body))
		out := h.mustRun("sites", "entries", "list", "--site-id", "1")
		if !strings.Contains(out, "\n  \"totalResults\": 1") {
			t.Errorf("整形されていません:\n%s", out)
		}
	})

	t.Run("compact は1行", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(body))
		out := h.mustRun("sites", "entries", "list", "--site-id", "1", "--compact")
		if strings.Count(strings.TrimSpace(out), "\n") != 0 {
			t.Errorf("1行になっていません:\n%s", out)
		}
		if !strings.Contains(out, `"totalResults":1`) {
			t.Errorf("出力 = %s", out)
		}
	})

	t.Run("raw はそのまま", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(body))
		out := h.mustRun("sites", "entries", "list", "--site-id", "1", "--format", "raw")
		if strings.TrimSpace(out) != body {
			t.Errorf("出力 = %s", out)
		}
	})

	t.Run("知らない形式は拒む", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(body))
		err := h.run("sites", "entries", "list", "--site-id", "1", "--format", "yaml")
		if err == nil || !strings.Contains(err.Error(), "auto / json / raw") {
			t.Fatalf("err = %v", err)
		}
	})
}

// JSON を返さないエンドポイントを整形しようとしない。
func TestNonJSONResponseIsPassedThrough(t *testing.T) {
	h := newHarness(t).reply(stub{status: 200, body: "id,title\n1,hello\n", contentType: "text/csv"})
	out := h.mustRun("sites", "entries", "export", "--site-id", "1")
	if out != "id,title\n1,hello\n" {
		t.Errorf("出力 = %q", out)
	}
}

// ベースURLを間違えて管理画面の HTML を掴んだときも、握りつぶさずに見せる。
func TestUnexpectedHTMLIsShown(t *testing.T) {
	h := newHarness(t).reply(stub{status: 200, body: "<html><body>Movable Type</body></html>", contentType: "text/html"})
	out := h.mustRun("sites", "entries", "list", "--site-id", "1")
	if !strings.Contains(out, "<html>") {
		t.Errorf("出力 = %q", out)
	}
}

func TestOutToFile(t *testing.T) {
	h := newHarness(t).reply(stub{status: 200, body: "backup-data", contentType: "application/octet-stream"})
	h.mustRun("sites", "backup", "1", "--out", "backup.tar.gz")

	got, ok := h.written["backup.tar.gz"]
	if !ok {
		t.Fatalf("書き出されたファイル = %v", h.written)
	}
	if string(got) != "backup-data" {
		t.Errorf("中身 = %q", got)
	}
	if h.out.Len() != 0 {
		t.Errorf("--out 指定時に標準出力へ本文が出ています: %q", h.out.String())
	}
	if !strings.Contains(h.errOut.String(), "backup.tar.gz") {
		t.Errorf("書き出し先が案内されていません: %q", h.errOut.String())
	}
}

func TestVerboseTracesToStderr(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1", "--verbose")

	trace := h.errOut.String()
	if !strings.Contains(trace, "> GET") || !strings.Contains(trace, "< 200") {
		t.Errorf("トレースが出ていません: %q", trace)
	}
	// トレースは標準出力を汚さない（パイプ処理を壊さないため）
	if strings.Contains(h.out.String(), "> GET") {
		t.Error("トレースが標準出力に出ています")
	}
}

func TestInsecureAndTimeoutReachTransport(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1", "--insecure", "--timeout", "5s")
	if !h.lastInsecure {
		t.Error("--insecure が伝わっていません")
	}
	if h.lastTimeout.String() != "5s" {
		t.Errorf("timeout = %s", h.lastTimeout)
	}
}

// Movable Type は日本語を ASCII エスケープした JSON を返す。そのまま出すと
// 人が読めず、エージェントにとっては同じ内容が数倍のトークンになる。
func TestJapaneseIsReadableInOutput(t *testing.T) {
	const name = "アイデアマンズ ブログ"
	escaped := `{"items":[{"id":2,"name":"` + asciiEscape(name) + `"}]}`
	if !strings.Contains(escaped, `\u`) {
		t.Fatalf("テストデータがエスケープされていません: %s", escaped)
	}

	t.Run("整形", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(escaped))
		out := h.mustRun("sites", "list")
		if !strings.Contains(out, name) {
			t.Errorf("日本語が読める形になっていません:\n%s", out)
		}
		if strings.Contains(out, `\u`) {
			t.Errorf("エスケープが残っています:\n%s", out)
		}
	})

	t.Run("compact", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(escaped))
		out := h.mustRun("sites", "list", "--compact")
		if !strings.Contains(out, name) {
			t.Errorf("出力 = %s", out)
		}
	})

	// raw は無加工。バイト列そのものが要る用途のための逃げ道。
	t.Run("raw は触らない", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(escaped))
		out := h.mustRun("sites", "list", "--format", "raw")
		if !strings.Contains(out, `\u`) {
			t.Errorf("raw が加工されています:\n%s", out)
		}
	})
}

// MT は GET にしか fields を用意していない。作成・更新でも同じ --fields が
// 使えないと「使えたり使えなかったり」になるので、手元で絞る。
func TestClientSideFields(t *testing.T) {
	t.Run("単一リソース", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{"id":"42","title":"記事","status":"Draft","body":"長い本文","blog":{"id":1}}`))
		out := h.mustRun("sites", "entries", "create", "--site-id", "1", "--entry", "{}", "--fields", "id,title")

		var got map[string]any
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		if len(got) != 2 || got["id"] != "42" || got["title"] != "記事" {
			t.Errorf("絞り込み結果 = %v", got)
		}
	})

	t.Run("一覧は items の各要素を絞り totalResults は残す", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{"totalResults":2,"items":[{"id":"1","title":"あ","body":"x"},{"id":"2","title":"い","body":"y"}]}`))
		// sites logs export はサーバ側 fields 非対応の GET
		out := h.mustRun("sites", "entries", "update", "1", "42", "--entry", "{}", "--fields", "id,title")

		var got struct {
			TotalResults int                      `json:"totalResults"`
			Items        []map[string]interface{} `json:"items"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		if got.TotalResults != 2 || len(got.Items) != 2 {
			t.Fatalf("got = %+v", got)
		}
		for _, item := range got.Items {
			if len(item) != 2 || item["body"] != nil {
				t.Errorf("item = %v", item)
			}
		}
	})

	// サーバが fields に対応しているコマンドでは、手元では絞らずサーバに渡す。
	t.Run("GET はサーバ側に渡す", func(t *testing.T) {
		h := newHarness(t).reply(okJSON(`{"items":[{"id":"1"}],"totalResults":1}`))
		h.mustRun("sites", "entries", "list", "--site-id", "1", "--fields", "id,title")

		u, err := url.Parse(h.last().URL)
		if err != nil {
			t.Fatal(err)
		}
		if u.Query().Get("fields") != "id,title" {
			t.Errorf("サーバに fields が渡っていません: %v", u.Query())
		}
	})

	// JSON でない応答は絞りようがないので素通し。
	t.Run("非JSONは素通し", func(t *testing.T) {
		h := newHarness(t).reply(stub{status: 200, body: "id,title\n1,x\n", contentType: "text/csv"})
		out := h.mustRun("sites", "entries", "update", "1", "42", "--entry", "{}", "--fields", "id")
		if out != "id,title\n1,x\n" {
			t.Errorf("出力 = %q", out)
		}
	})
}

// 属性名を間違えると {} が出るだけで何が起きたか分からない。
// その応答が実際に持っている名前を標準エラーで教える。
func TestClientSideFieldsWarnsOnNoMatch(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"status":"success"}`))
	out := h.mustRun("sites", "permissions", "grant", "--site-id", "2",
		"--user-id", "2", "--role-id", "2", "--fields", "id,permissions")

	if strings.TrimSpace(out) != "{}" {
		t.Errorf("stdout = %q", out)
	}
	warn := h.errOut.String()
	for _, want := range []string{"id,permissions", "使える属性", "status"} {
		if !strings.Contains(warn, want) {
			t.Errorf("警告に %q がありません: %q", want, warn)
		}
	}
}

// 一致していれば黙っていること（毎回出ると邪魔になる）。
func TestClientSideFieldsQuietOnMatch(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"42","title":"記事"}`))
	h.mustRun("sites", "entries", "create", "--site-id", "1", "--entry", "{}", "--fields", "id")
	if h.errOut.Len() != 0 {
		t.Errorf("余計な警告が出ています: %q", h.errOut.String())
	}
}

// 1つの項目を生の値で取り出せること。これが無いと、テンプレートの
// text をファイルへ落とすのに jq が要る。
func TestExtractField(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"id":"42","text":"<!doctype html>\n<html><$mt:BlogName$></html>\n"}`))

	out := h.mustRun("sites", "templates", "get", "1", "42", "--extract", "text")

	want := "<!doctype html>\n<html><$mt:BlogName$></html>\n"
	if out != want {
		t.Errorf("取り出した値 = %q, want %q", out, want)
	}
}

// 文字列以外は JSON のまま出す（jq -r と同じ振る舞い）。
func TestExtractNonString(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"items":[{"name":"a"},{"name":"b"}],"totalResults":2}`))

	out := h.mustRun("sites", "templates", "list", "1", "--extract", "totalResults")
	if strings.TrimSpace(out) != "2" {
		t.Errorf("out = %q", out)
	}
}

// 配列の添字をたどれること。
func TestExtractWalksArrays(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"items":[{"name":"a"},{"name":"b"}]}`))

	// 端末に出すときだけ末尾に改行が付きます（--out はバイトそのまま）。
	if out := h.mustRun("sites", "templates", "list", "1", "--extract", "items.1.name"); out != "b\n" {
		t.Errorf("out = %q, want \"b\\n\"", out)
	}
}

// 綴り違いはよくある。何が取れるのかをその場で見せる。
func TestExtractUnknownFieldListsWhatIsThere(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"id":"42","text":"x"}`))

	err := h.run("sites", "templates", "get", "1", "42", "--extract", "txet")
	if err == nil {
		t.Fatal("エラーになるはずです")
	}
	for _, want := range []string{"txet", "id, text"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("エラーに %q がありません: %v", want, err)
		}
	}
}

// raw は「無加工」という意味なので、取り出しとは両立しない。
func TestExtractConflictsWithRaw(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"text":"x"}`))

	err := h.run("sites", "templates", "get", "1", "42", "--extract", "text", "--format", "raw")
	if err == nil {
		t.Fatal("エラーになるはずです")
	}
	if code := ExitCode(err); code != ExitUsage {
		t.Errorf("exit = %d, want %d", code, ExitUsage)
	}
}

// --out と組み合わせると、そのままファイルになる。
func TestExtractToFile(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"text":"<html></html>"}`))

	h.mustRun("sites", "templates", "get", "1", "42", "--extract", "text", "--out", "index.html")
	if got := string(h.written["index.html"]); got != "<html></html>" {
		t.Errorf("書き出し = %q", got)
	}
}

// サーバ側で絞る endpoint でも、1つも一致しなければ知らせる。
//
// 作成直後に --fields id で {} が返ると、作成が失敗したのか属性名を
// 間違えたのか区別が付きません。手元で絞る経路には案内があるので、
// サーバ側の経路にも同じ知らせを出します。
func TestFieldsWarnsWhenServerFilteredEverythingOut(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{}`))
	if err := h.run("sites", "get", "6", "--fields", "nosuchfield"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.errOut.String(), "1つも一致しませんでした") {
		t.Errorf("知らせが出ていません: %q", h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), "nosuchfield") {
		t.Errorf("外した属性名が出ていません: %q", h.errOut.String())
	}
}

// 一覧で全件が空オブジェクトになったときも知らせる。
func TestFieldsWarnsWhenEveryItemIsEmpty(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"totalResults":2,"items":[{},{}]}`))
	if err := h.run("sites", "entries", "list", "6", "--fields", "nosuchfield"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.errOut.String(), "1つも一致しませんでした") {
		t.Errorf("知らせが出ていません: %q", h.errOut.String())
	}
}

// 対象が0件なだけのときは知らせない。--fields のせいではない。
func TestFieldsQuietWhenListIsEmpty(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"totalResults":0,"items":[]}`))
	if err := h.run("sites", "entries", "list", "6", "--fields", "id"); err != nil {
		t.Fatal(err)
	}
	if h.errOut.String() != "" {
		t.Errorf("余計な知らせが出ています: %q", h.errOut.String())
	}
}

// 一致していれば黙っている。
func TestFieldsQuietWhenMatched(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"id":"6","name":"SaasHub"}`))
	if err := h.run("sites", "get", "6", "--fields", "id"); err != nil {
		t.Fatal(err)
	}
	if h.errOut.String() != "" {
		t.Errorf("余計な知らせが出ています: %q", h.errOut.String())
	}
}
