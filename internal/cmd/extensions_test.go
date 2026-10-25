package cmd

import (
	"net/url"
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/extension"
)

const testExtManifest = `{"name":"entry-author","version":"0.1.0","description":"Change the author","descriptionJa":"作成者を変える",
 "requires":{"mtPlugin":{"name":"EntryAuthor","version":"0.1.0"},"endpoints":["update_entry_author"]},
 "claudePlugin":"plugins/mtb-entry-author"}`

const testExtSpec = `{"groups":[{"command":["sites","entries","author"],"short":"Author"}],
 "commands":[{"command":["sites","entries","author","update"],"short":"Change the author",
  "ops":[{"method":"PUT","path":"/sites/{site_id}/entries/{entry_id}/author",
   "pathParams":[{"name":"site_id","type":"integer","required":true},{"name":"entry_id","type":"integer","required":true}],
   "body":[{"name":"authorId","type":"integer"},{"name":"authorName"}]}]}]}`

// withExtensionFiles は拡張のディレクトリ dir を手元のファイルとして置きます。
func withExtensionFiles(dir, manifest, specJSON string) harnessOption {
	return func(h *harness) {
		h.files[dir+"/mtb-extension.json"] = manifest
		h.files[dir+"/spec.json"] = specJSON
		h.files[dir+"/llm.md"] = "# 使い方\n\n```bash\n# コメントは見出しではない\n```\n"
		h.files[dir+"/plugins/mtb-entry-author/.claude-plugin/plugin.json"] = `{"name":"mtb-entry-author"}`
		h.files[dir+"/.git/HEAD"] = "ref: refs/heads/main"
	}
}

// loadExtension は手元に置いた拡張を、起動時に見つけたものとして Env に載せます。
func (h *harness) loadExtension(dir, source string) *harness {
	h.t.Helper()
	e, err := extension.Load(h.env.ReadFile, dir)
	if err != nil {
		h.t.Fatal(err)
	}
	e.Source = source
	h.env.Extensions = append(h.env.Extensions, e)
	return h
}

func newExtensionHarness(t *testing.T, opts ...harnessOption) *harness {
	opts = append([]harnessOption{withExtensionFiles("/src/ext", testExtManifest, testExtSpec)}, opts...)
	return newHarness(t, opts...).loadExtension("/src/ext", extension.SourceInstalled)
}

func TestExtensionCommandSendsLikeAGeneratedOne(t *testing.T) {
	h := newExtensionHarness(t).reply(okJSON(`{"id":42,"author":{"id":3,"displayName":"Editor"}}`))
	out := h.mustRun("sites", "entries", "author", "update", "1", "42", "--author-name", "editor", "--fields", "author")

	sent := h.last()
	if sent.Method != "PUT" {
		t.Errorf("method = %s", sent.Method)
	}
	if got := pathOf(t, sent.URL); got != "/mt/mt-data-api.cgi/v7/sites/1/entries/42/author" {
		t.Errorf("path = %s", got)
	}
	form, err := url.ParseQuery(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	if form.Get("authorName") != "editor" {
		t.Errorf("body = %s", sent.Body)
	}
	if _, ok := form["authorId"]; ok {
		t.Errorf("指定していない authorId が送られています: %s", sent.Body)
	}
	// 生成コマンドと同じ道具（--fields の手元での絞り込み）が効くこと
	if strings.Contains(out, `"id": 42`) || !strings.Contains(out, "Editor") {
		t.Errorf("--fields が効いていません: %s", out)
	}
}

func TestExtensionCommandWorksThroughTheSingularAlias(t *testing.T) {
	h := newExtensionHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entry", "author", "update", "1", "42", "--author-id", "3")
	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/1/entries/42/author" {
		t.Errorf("path = %s", got)
	}
}

func TestExtensionHelpSaysWhereTheCommandComesFrom(t *testing.T) {
	h := newExtensionHarness(t)
	out := h.mustRun("sites", "entries", "author", "update", "--help")
	for _, want := range []string{"PUT /sites/{site_id}/entries/{entry_id}/author", "拡張 entry-author", "EntryAuthor 0.1.0 プラグイン", "--author-name"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help に %q がありません:\n%s", want, out)
		}
	}
}

func TestExtensionNotFoundPointsAtTheMTPlugin(t *testing.T) {
	h := newExtensionHarness(t).reply(stub{status: 404, body: `{"error":{"code":404,"message":"Unknown endpoint"}}`, contentType: "application/json"})
	err := h.run("sites", "entries", "author", "update", "1", "42", "--author-id", "3")
	if err == nil {
		t.Fatal("404 なのに成功しました")
	}
	if !strings.Contains(err.Error(), "EntryAuthor 0.1.0 プラグインが入っていない") || !strings.Contains(err.Error(), "mtb extensions check entry-author") {
		t.Errorf("MT プラグインの助言がありません: %v", err)
	}
	if code := ExitCode(err); code != ExitNotFound {
		t.Errorf("終了コード = %d（404 の %d のままであるべき）", code, ExitNotFound)
	}
}

func TestExtensionThatCollidesIsNotLoadedAtAll(t *testing.T) {
	cases := map[string]string{
		// 生成コマンドとぶつかる
		"生成コマンド": strings.Replace(testExtSpec, `"author","update"]`, `"update"]`, 1),
		// 手書きのコマンドとぶつかる
		"手書きのコマンド": `{"commands":[
			{"command":["sites","entries","author","update"],"short":"x","ops":[{"method":"PUT","path":"/x"}]},
			{"command":["profile","list"],"short":"x","ops":[{"method":"GET","path":"/y"}]}],
			"groups":[{"command":["sites","entries","author"]}]}`,
		// 単数形の別名とぶつかる
		"別名": `{"commands":[{"command":["site"],"short":"x","ops":[{"method":"GET","path":"/x"}]}]}`,
	}
	for name, specJSON := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, withExtensionFiles("/src/ext", testExtManifest, specJSON)).loadExtension("/src/ext", extension.SourceInstalled)
			h.mustRun("extensions", "list")
			if !strings.Contains(h.errOut.String(), "ぶつかる") {
				t.Errorf("警告が出ていません: %q", h.errOut.String())
			}
			if !strings.Contains(h.out.String(), "読み込めない") {
				t.Errorf("一覧に読み込めないと出ていません:\n%s", h.out.String())
			}
			// 一部だけ入ってはいけない（ぶつからなかったコマンドも足さない）
			if findCommand(New(h.env).Root(), []string{"sites", "entries", "author"}) != nil {
				t.Error("ぶつかった拡張のコマンドが一部だけ入っています")
			}
		})
	}
}

func TestExtensionWithoutAParentGroupIsRejected(t *testing.T) {
	specJSON := strings.Replace(testExtSpec, `"groups":[{"command":["sites","entries","author"],"short":"Author"}],`, "", 1)
	h := newHarness(t, withExtensionFiles("/src/ext", testExtManifest, specJSON))
	h.files["/src/ext/llm.md"] = ""
	err := h.run("extensions", "validate", "/src/ext")
	if err == nil || !strings.Contains(err.Error(), "groups") {
		t.Errorf("親グループの無い定義が通りました: %v", err)
	}
}

func TestExtensionsListAndShow(t *testing.T) {
	h := newExtensionHarness(t)
	out := h.mustRun("extensions", "list")
	for _, want := range []string{"entry-author", "0.1.0", "installed", "読み込み済み", "作成者を変える"} {
		if !strings.Contains(out, want) {
			t.Errorf("一覧に %q がありません:\n%s", want, out)
		}
	}
	out = h.mustRun("extensions", "show", "entry-author")
	for _, want := range []string{"mtb sites entries author update", "EntryAuthor 0.1.0", "claude --plugin-dir /src/ext/plugins/mtb-entry-author"} {
		if !strings.Contains(out, want) {
			t.Errorf("詳細に %q がありません:\n%s", want, out)
		}
	}
	if err := h.run("extensions", "show", "nothing"); ExitCode(err) != ExitUsage {
		t.Errorf("無い拡張: %v", err)
	}
}

func TestExtensionsInstallCopiesTheDirectory(t *testing.T) {
	h := newHarness(t, withExtensionFiles("/src/ext", testExtManifest, testExtSpec))
	out := h.mustRun("extensions", "install", "/src/ext")

	dest := "/config/extensions/entry-author"
	for _, rel := range []string{"mtb-extension.json", "spec.json", "llm.md", "plugins/mtb-entry-author/.claude-plugin/plugin.json"} {
		if _, ok := h.written[dest+"/"+rel]; !ok {
			t.Errorf("%s を写していません", rel)
		}
	}
	if _, ok := h.written[dest+"/.git/HEAD"]; ok {
		t.Error(".git まで写しています")
	}
	if len(h.removed) != 1 || h.removed[0] != dest {
		t.Errorf("先に消すのは入れる先だけのはず: %v", h.removed)
	}
	if !strings.Contains(out, "mtb sites entries author update") || !strings.Contains(out, "mtb extensions check entry-author") {
		t.Errorf("入れたあとに足したコマンドと確認の仕方を見せていません:\n%s", out)
	}
}

func TestExtensionsInstallRefusesToOverwriteWithoutForce(t *testing.T) {
	h := newExtensionHarness(t)
	err := h.run("extensions", "install", "/src/ext")
	if ExitCode(err) != ExitUsage || !strings.Contains(err.Error(), "--force") {
		t.Errorf("上書きを止めていません: %v", err)
	}
	if len(h.written) != 0 {
		t.Errorf("止めたのに書いています: %v", h.written)
	}
	// --force なら、自分の古いコマンドとはぶつからずに入れ直せる
	h.mustRun("extensions", "install", "/src/ext", "--force")
	if len(h.written) == 0 {
		t.Error("--force でも入れていません")
	}
}

func TestExtensionsInstallDryRunWritesNothing(t *testing.T) {
	h := newHarness(t, withExtensionFiles("/src/ext", testExtManifest, testExtSpec))
	h.mustRun("extensions", "install", "/src/ext", "--dry-run")
	if len(h.written) != 0 || len(h.removed) != 0 {
		t.Errorf("--dry-run なのに書いています: %v %v", h.written, h.removed)
	}
}

func TestExtensionsUninstall(t *testing.T) {
	h := newExtensionHarness(t)
	h.mustRun("extensions", "uninstall", "entry-author")
	if len(h.removed) != 1 || h.removed[0] != "/src/ext" {
		t.Errorf("removed = %v", h.removed)
	}

	// MTB_EXTENSIONS_PATH から読んだもの（作りかけ）は消さない
	h = newHarness(t, withExtensionFiles("/src/ext", testExtManifest, testExtSpec)).loadExtension("/src/ext", extension.SourcePath)
	err := h.run("extensions", "uninstall", "entry-author")
	if ExitCode(err) != ExitUsage || len(h.removed) != 0 {
		t.Errorf("開発中の拡張を消そうとしました: %v %v", err, h.removed)
	}
}

func TestExtensionsCheckReadsTheServerEndpoints(t *testing.T) {
	h := newExtensionHarness(t).reply(okJSON(`{"items":[{"id":"list_entries"},{"id":"update_entry_author"}]}`))
	out := h.mustRun("extensions", "check")
	if !strings.Contains(out, "entry-author: OK") {
		t.Errorf("out = %s", out)
	}
	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/endpoints" {
		t.Errorf("path = %s", got)
	}

	h = newExtensionHarness(t).reply(okJSON(`{"items":[{"id":"list_entries"}]}`))
	err := h.run("extensions", "check", "entry-author")
	if err == nil {
		t.Fatal("足りないのに成功しました")
	}
	if !strings.Contains(h.out.String(), "update_entry_author") || !strings.Contains(h.out.String(), "EntryAuthor 0.1.0") {
		t.Errorf("何が足りないか出ていません: %s", h.out.String())
	}
}

func TestLLMIncludesTheExtensionChapter(t *testing.T) {
	h := newExtensionHarness(t)
	out := h.mustRun("llm")
	for _, want := range []string{"# 拡張: entry-author", "`mtb sites entries author update`", "PUT /sites/{site_id}/entries/{entry_id}/author", "EntryAuthor 0.1.0 プラグインが要る", "## 使い方"} {
		if !strings.Contains(out, want) {
			t.Errorf("mtb llm に %q がありません", want)
		}
	}
	// 拡張の章はコマンドカタログより前、手書きの章より後
	if i, j := strings.Index(out, "# 拡張: entry-author"), strings.Index(out, "# コマンドカタログ"); i < 0 || j < 0 || i > j {
		t.Errorf("章の順番が違います（拡張 %d、カタログ %d）", i, j)
	}
	// コードブロックの中のシェルのコメントは見出しにしない
	if strings.Contains(out, "## コメントは見出しではない") {
		t.Error("コードブロックの中まで見出しを下げています")
	}
}

func TestNoExtensionsMeansNoChange(t *testing.T) {
	h := newHarness(t)
	h.mustRun("extensions", "list")
	if !strings.Contains(h.out.String(), "拡張はありません") {
		t.Errorf("out = %s", h.out.String())
	}
	if h.errOut.Len() != 0 {
		t.Errorf("拡張が無いのに警告が出ています: %s", h.errOut.String())
	}
}

// file と path を取る拡張のコマンドには --dir が付き、--replace は付かない。
// --replace は同名の「アセット」を消してから送るもので、テーマやサイトの
// ファイルを置くコマンドに付くと、関係の無いアセットを消しに行く。
func TestExtensionFileUploadGetsDirButNotReplace(t *testing.T) {
	specJSON := `{"groups":[{"command":["themes","files"]}],
	 "commands":[{"command":["themes","files","upload"],"short":"Upload",
	  "ops":[{"method":"POST","path":"/themes/{theme_id}/files/upload",
	   "pathParams":[{"name":"theme_id","required":true}],
	   "body":[{"name":"file","type":"file"},{"name":"path"}],"bodyKind":"multipart"}]}]}`
	h := newHarness(t, withExtensionFiles("/src/ext", testExtManifest, specJSON)).loadExtension("/src/ext", extension.SourceInstalled)
	c := findCommand(New(h.env).Root(), []string{"themes", "files", "upload"})
	if c == nil {
		t.Fatal("拡張のコマンドがありません")
	}
	if c.Flags().Lookup(dirFlag) == nil {
		t.Error("--dir が付いていません")
	}
	if c.Flags().Lookup(replaceFlag) != nil {
		t.Error("アセットのアップロードではないのに --replace が付いています")
	}
	// 本体のアセットのアップロードには付いたまま
	if a := findCommand(New(h.env).Root(), []string{"sites", "assets", "upload"}); a == nil || a.Flags().Lookup(replaceFlag) == nil {
		t.Error("sites assets upload から --replace が消えています")
	}
}

// --strict は同梱の Claude Code のプラグインの問題も失敗にする（既定は警告）。
func TestExtensionsValidateStrictFailsOnClaudePluginProblems(t *testing.T) {
	h := newHarness(t, withExtensionFiles("/src/ext", testExtManifest, testExtSpec))
	var gotDir, gotVersion string
	h.env.CheckClaudePlugin = func(dir, version string) []string {
		gotDir, gotVersion = dir, version
		return []string{"plugin.json version 0.0.9 != 0.1.0"}
	}
	if err := h.run("extensions", "validate", "/src/ext"); err != nil {
		t.Fatalf("既定では警告で通るはず: %v", err)
	}
	if !strings.Contains(h.errOut.String(), "0.0.9") {
		t.Errorf("警告が出ていません: %q", h.errOut.String())
	}
	if gotDir != "/src/ext/plugins/mtb-entry-author" || gotVersion != "0.1.0" {
		t.Errorf("検査に渡したもの = %q %q", gotDir, gotVersion)
	}
	if err := h.run("extensions", "validate", "/src/ext", "--strict"); err == nil || !strings.Contains(err.Error(), "--strict") {
		t.Errorf("--strict で失敗していません: %v", err)
	}
}
