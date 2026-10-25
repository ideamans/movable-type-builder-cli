// Package extension は mtb の拡張（エクステンション）を読み込みます。
//
// MT のプラグインは Data API にエンドポイントを足せますが、mtb のコマンドは
// 公式の定義から作るので、足されたエンドポイントにはコマンドがありません
// （mtb api で叩くしかない）。拡張は、その足りない定義を mtb に持ち込む
// ためのものです。
//
// 拡張は**宣言だけで、コードを持ちません**。中身はディレクトリ1つで、
//
//	mtb-extension.json   マニフェスト（名前・版・要る MT プラグイン）
//	spec.json            追加するコマンドの定義（internal/spec と同じ形）
//	llm.md               mtb llm に足す章（任意）
//	plugins/<名前>/       Claude Code のプラグイン（任意）
//
// コマンドは埋め込みの定義と同じ道具（newOperationCmd）で組むので、
// --set / --fields / --dry-run / --all がそのまま効きます。コードを
// 読み込まないので、拡張を入れても任意のコードは走りません。
package extension

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// ManifestFile は拡張のディレクトリに置くマニフェストの名前です。
const ManifestFile = "mtb-extension.json"

const (
	defaultSpecFile = "spec.json"
	defaultLLMFile  = "llm.md"
)

var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Manifest は mtb-extension.json の中身です。
type Manifest struct {
	// Name は拡張の名前です（英小文字・数字・ハイフン）。
	// インストール先のディレクトリ名にもなります。
	Name    string `json:"name"`
	Version string `json:"version"`
	// Description は英語の説明、DescriptionJa は日本語の説明です。
	Description   string `json:"description"`
	DescriptionJa string `json:"descriptionJa,omitempty"`
	// Requires はこの拡張が MT 側に求めるものです。
	Requires *Requires `json:"requires,omitempty"`
	// Spec・LLM・ClaudePlugin は拡張のディレクトリからの相対パスです。
	// Spec と LLM は省略すると spec.json と llm.md を探します。
	Spec         string `json:"spec,omitempty"`
	LLM          string `json:"llm,omitempty"`
	ClaudePlugin string `json:"claudePlugin,omitempty"`
}

// Requires は拡張が動くのに要る MT 側の条件です。
type Requires struct {
	// MTPlugin は、拡張のエンドポイントを生やす MT プラグインです。
	MTPlugin *MTPlugin `json:"mtPlugin,omitempty"`
	// Endpoints は MT の /endpoints に現れるはずのエンドポイント ID です。
	// mtb extensions check がこれでサーバ側の有無を確かめます。
	Endpoints []string `json:"endpoints,omitempty"`
}

// MTPlugin は MT プラグイン1つです。
type MTPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	URL     string `json:"url,omitempty"`
}

// Label は「名前 版」です。版が無ければ名前だけです。
func (p *MTPlugin) Label() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + " " + p.Version
}

// RequiredPlugin は拡張が要る MT プラグインです。無ければ nil です。
func (m *Manifest) RequiredPlugin() *MTPlugin {
	if m.Requires == nil {
		return nil
	}
	return m.Requires.MTPlugin
}

// Describe は表示言語に合わせた説明です。
func (m *Manifest) Describe() string {
	if m.DescriptionJa == "" {
		return m.Description
	}
	return i18n.T(m.DescriptionJa, m.Description)
}

// Extension は読み込み済みの拡張1つです。
type Extension struct {
	Manifest
	// Dir は拡張のディレクトリです。
	Dir string
	// Source は見つけた場所の種類です（installed か path）。
	Source string
	// Spec は追加するコマンドの定義です。
	Spec *spec.Spec
	// LLMDoc は mtb llm に足す章です。無ければ空です。
	LLMDoc string
}

// 拡張を見つけた場所の種類です。
const (
	// SourceInstalled は mtb extensions install で入れたものです。
	SourceInstalled = "installed"
	// SourcePath は MTB_EXTENSIONS_PATH で指したものです（開発用）。
	SourcePath = "path"
)

// ClaudePluginDir は Claude Code のプラグインのディレクトリです。無ければ空です。
func (e *Extension) ClaudePluginDir() string {
	if e.ClaudePlugin == "" {
		return ""
	}
	return filepath.Join(e.Dir, e.ClaudePlugin)
}

// ReadFunc はファイル1つを読みます。無いときは fs.ErrNotExist を返すこと。
type ReadFunc func(name string) ([]byte, error)

// Load は dir にある拡張を読み込み、中身を検査します。
func Load(read ReadFunc, dir string) (*Extension, error) {
	raw, err := read(filepath.Join(dir, ManifestFile))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf(i18n.T(
				"%s に %s がありません（拡張のディレクトリではありません）",
				"%s has no %s (it is not an extension directory)"), dir, ManifestFile)
		}
		return nil, err
	}
	e := &Extension{Dir: dir}
	if err := json.Unmarshal(raw, &e.Manifest); err != nil {
		return nil, fmt.Errorf(i18n.T("%s を読めません: %w", "cannot read %s: %w"), ManifestFile, err)
	}
	if err := e.Manifest.validate(); err != nil {
		return nil, err
	}

	specFile := e.Manifest.Spec
	if specFile == "" {
		specFile = defaultSpecFile
	}
	raw, err = read(filepath.Join(dir, specFile))
	if err != nil {
		return nil, fmt.Errorf(i18n.T("拡張 %s の定義 %s を読めません: %w", "extension %s: cannot read the definition %s: %w"), e.Name, specFile, err)
	}
	s := &spec.Spec{}
	if err := json.Unmarshal(raw, s); err != nil {
		return nil, fmt.Errorf(i18n.T("拡張 %s の定義 %s を読めません: %w", "extension %s: cannot read the definition %s: %w"), e.Name, specFile, err)
	}
	if err := Normalize(s); err != nil {
		return nil, fmt.Errorf(i18n.T("拡張 %s の定義: %w", "definition of extension %s: %w"), e.Name, err)
	}
	e.Spec = s

	llmFile := e.Manifest.LLM
	if raw, err := read(filepath.Join(dir, firstNonEmpty(llmFile, defaultLLMFile))); err == nil {
		e.LLMDoc = string(raw)
	} else if llmFile != "" || !errors.Is(err, fs.ErrNotExist) {
		// 明示した章が無いのは書き損じ。既定の llm.md が無いのは構わない。
		return nil, fmt.Errorf(i18n.T("拡張 %s の章 %s を読めません: %w", "extension %s: cannot read the chapter %s: %w"), e.Name, firstNonEmpty(llmFile, defaultLLMFile), err)
	}

	if e.ClaudePlugin != "" {
		manifest := filepath.Join(e.ClaudePluginDir(), ".claude-plugin", "plugin.json")
		if _, err := read(manifest); err != nil {
			return nil, fmt.Errorf(i18n.T(
				"拡張 %s の claudePlugin に %s がありません",
				"claudePlugin of extension %s has no %s"), e.Name, manifest)
		}
	}
	return e, nil
}

func (m *Manifest) validate() error {
	if !namePattern.MatchString(m.Name) {
		return fmt.Errorf(i18n.T(
			"%s の name %q は英小文字・数字・ハイフンで書いてください",
			"%s: name %q must be lowercase letters, digits and hyphens"), ManifestFile, m.Name)
	}
	if m.Version == "" {
		return fmt.Errorf(i18n.T("拡張 %s に version がありません", "extension %s has no version"), m.Name)
	}
	if m.Description == "" {
		return fmt.Errorf(i18n.T("拡張 %s に description がありません", "extension %s has no description"), m.Name)
	}
	if m.Requires != nil && m.Requires.MTPlugin != nil && m.Requires.MTPlugin.Name == "" {
		return fmt.Errorf(i18n.T("拡張 %s の requires.mtPlugin に name がありません", "requires.mtPlugin of extension %s has no name"), m.Name)
	}
	for _, p := range []string{m.Spec, m.LLM, m.ClaudePlugin} {
		if p != "" && (path.IsAbs(p) || strings.HasPrefix(path.Clean(p), "..")) {
			return fmt.Errorf(i18n.T(
				"拡張 %s のパス %q は拡張のディレクトリの中を指してください",
				"extension %s: path %q must point inside the extension directory"), m.Name, p)
		}
	}
	return nil
}

var methods = map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true}

var types = map[string]bool{
	"string": true, "integer": true, "number": true, "boolean": true,
	"object": true, "array": true, "file": true,
}

// Normalize は手で書いた定義の省略を補い、組み立てられない定義を弾きます。
//
// 拡張の定義は人が手で書くので、生成器が埋める項目（フラグ名・型）は
// 省けるようにしてあります。フラグ名は API の名前をケバブケースにしたもの、
// 型は string です。
func Normalize(s *spec.Spec) error {
	if len(s.Commands) == 0 {
		return errors.New(i18n.T("commands が空です", "commands is empty"))
	}
	for _, g := range s.Groups {
		if len(g.Command) == 0 {
			return errors.New(i18n.T("command の無い group があります", "a group has no command"))
		}
	}
	for _, c := range s.Commands {
		where := strings.Join(c.Command, " ")
		if len(c.Command) == 0 {
			return errors.New(i18n.T("command の無いコマンドがあります", "a command has no command path"))
		}
		if len(c.Ops) == 0 {
			return fmt.Errorf(i18n.T("%s: ops が空です", "%s: ops is empty"), where)
		}
		for _, op := range c.Ops {
			op.Method = strings.ToUpper(op.Method)
			if !methods[op.Method] {
				return fmt.Errorf(i18n.T("%s: method %q は使えません", "%s: method %q is not supported"), where, op.Method)
			}
			if !strings.HasPrefix(op.Path, "/") {
				return fmt.Errorf(i18n.T("%s: path は / で始めてください（%q）", "%s: path must start with / (%q)"), where, op.Path)
			}
			for _, group := range [][]*spec.Param{op.PathParams, op.Query, op.Body} {
				for _, p := range group {
					if p.Name == "" {
						return fmt.Errorf(i18n.T("%s: name の無いパラメータがあります", "%s: a parameter has no name"), where)
					}
					if p.Flag == "" {
						p.Flag = Kebab(p.Name)
					}
					if p.Type == "" {
						p.Type = "string"
					}
					if !types[p.Type] {
						return fmt.Errorf(i18n.T("%s: %s の type %q は使えません", "%s: %s has an unsupported type %q"), where, p.Name, p.Type)
					}
				}
			}
			for _, p := range op.PathParams {
				if !strings.Contains(op.Path, "{"+p.Name+"}") {
					return fmt.Errorf(i18n.T(
						"%s: パスパラメータ %s が path %s にありません",
						"%s: path parameter %s does not appear in path %s"), where, p.Name, op.Path)
				}
			}
			if len(op.Body) > 0 && op.BodyKind == "" {
				op.BodyKind = "form"
			}
		}
	}
	return nil
}

// Kebab は API の名前（authorId・author_id）をフラグ名（author-id）にします。
func Kebab(name string) string {
	var b strings.Builder
	prev := rune(0)
	for _, r := range name {
		switch {
		case r == '_' || r == ' ':
			b.WriteRune('-')
		case unicode.IsUpper(r):
			if b.Len() > 0 && prev != '_' && !unicode.IsUpper(prev) {
				b.WriteRune('-')
			}
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(r)
		}
		prev = r
	}
	return b.String()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
