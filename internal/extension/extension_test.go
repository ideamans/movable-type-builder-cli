package extension

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// files は拡張のディレクトリを模したものです。
type files map[string]string

func (f files) read(name string) ([]byte, error) {
	if s, ok := f[name]; ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
}

const manifest = `{"name":"entry-author","version":"0.1.0","description":"Change the author",
 "requires":{"mtPlugin":{"name":"EntryAuthor","version":"0.1.0"},"endpoints":["update_entry_author"]}}`

const specJSON = `{"groups":[{"command":["sites","entries","author"],"short":"Author"}],
 "commands":[{"command":["sites","entries","author","update"],"short":"Change the author",
  "ops":[{"method":"put","path":"/sites/{site_id}/entries/{entry_id}/author",
   "pathParams":[{"name":"site_id","type":"integer","required":true},{"name":"entry_id","type":"integer","required":true}],
   "body":[{"name":"authorId","type":"integer"},{"name":"authorName"}]}]}]}`

func valid() files {
	return files{"ext/mtb-extension.json": manifest, "ext/spec.json": specJSON}
}

func TestLoadFillsWhatAHandWrittenDefinitionLeavesOut(t *testing.T) {
	e, err := Load(valid().read, "ext")
	if err != nil {
		t.Fatal(err)
	}
	op := e.Spec.Commands[0].Ops[0]
	if op.Method != "PUT" {
		t.Errorf("method = %s（大文字にそろえていません）", op.Method)
	}
	if op.BodyKind != "form" {
		t.Errorf("bodyKind = %q（本文があるのに既定の form になっていません）", op.BodyKind)
	}
	if got := op.PathParams[0].Flag; got != "site-id" {
		t.Errorf("flag = %q", got)
	}
	if got := op.Body[1]; got.Flag != "author-name" || got.Type != "string" {
		t.Errorf("authorName = %+v（フラグ名と型を補っていません）", got)
	}
	if p := e.RequiredPlugin(); p == nil || p.Label() != "EntryAuthor 0.1.0" {
		t.Errorf("RequiredPlugin = %+v", p)
	}
	if e.LLMDoc != "" {
		t.Errorf("llm.md が無いのに章があります: %q", e.LLMDoc)
	}
}

func TestLoadReadsTheChapter(t *testing.T) {
	f := valid()
	f["ext/llm.md"] = "## 使い方"
	e, err := Load(f.read, "ext")
	if err != nil {
		t.Fatal(err)
	}
	if e.LLMDoc != "## 使い方" {
		t.Errorf("LLMDoc = %q", e.LLMDoc)
	}
}

func TestLoadRejects(t *testing.T) {
	cases := map[string]struct {
		edit func(files)
		want string
	}{
		"マニフェストが無い": {func(f files) { delete(f, "ext/mtb-extension.json") }, "mtb-extension.json"},
		"名前に大文字": {func(f files) {
			f["ext/mtb-extension.json"] = strings.Replace(manifest, "entry-author", "EntryAuthor", 1)
		}, "name"},
		"版が無い": {func(f files) {
			f["ext/mtb-extension.json"] = strings.Replace(manifest, `"version":"0.1.0",`, "", 1)
		}, "version"},
		"定義が無い":        {func(f files) { delete(f, "ext/spec.json") }, "spec.json"},
		"コマンドが空":       {func(f files) { f["ext/spec.json"] = `{"commands":[]}` }, "commands"},
		"パスに無いパスパラメータ": {func(f files) { f["ext/spec.json"] = strings.Replace(specJSON, "{entry_id}", "{id}", 1) }, "entry_id"},
		"知らないメソッド":     {func(f files) { f["ext/spec.json"] = strings.Replace(specJSON, `"put"`, `"PATCH"`, 1) }, "PATCH"},
		"知らない型":        {func(f files) { f["ext/spec.json"] = strings.Replace(specJSON, `"type":"integer"}`, `"type":"int"}`, 1) }, "int"},
		"明示した章が無い": {func(f files) {
			f["ext/mtb-extension.json"] = strings.Replace(manifest, `"description"`, `"llm":"guide.md","description"`, 1)
		}, "guide.md"},
		"外を指すパス": {func(f files) {
			f["ext/mtb-extension.json"] = strings.Replace(manifest, `"description"`, `"spec":"../x.json","description"`, 1)
		}, "../x.json"},
		"Claude のプラグインが無い": {func(f files) {
			f["ext/mtb-extension.json"] = strings.Replace(manifest, `"description"`, `"claudePlugin":"plugins/p","description"`, 1)
		}, "plugin.json"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f := valid()
			c.edit(f)
			_, err := Load(f.read, "ext")
			if err == nil {
				t.Fatal("通ってしまいました")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("エラーに %q がありません: %v", c.want, err)
			}
		})
	}
}

func TestKebab(t *testing.T) {
	for in, want := range map[string]string{
		"authorId":     "author-id",
		"author_id":    "author-id",
		"saveRevision": "save-revision",
		"id":           "id",
		"URL":          "url",
	} {
		if got := Kebab(in); got != want {
			t.Errorf("Kebab(%q) = %q, want %q", in, got, want)
		}
	}
}
