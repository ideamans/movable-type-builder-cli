package i18n_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// ユーザーに見せる文言を持つパッケージ。ここに日本語のリテラルが
// i18n.T / i18n.Error の外に現れたら、その文言は英語環境でも日本語のまま
// 出ます。目視では気づけないので機械で見張ります。
var userFacing = []string{
	"internal/cmd",
	"internal/api",
	"internal/config",
	"internal/prompt",
	"internal/spec",
	"internal/extension",
	"cmd/mtb",
}

func TestNoUntranslatedUserFacingText(t *testing.T) {
	fset := token.NewFileSet()
	for _, dir := range userFacing {
		root := filepath.Join("..", "..", dir)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			for _, lit := range untranslated(file) {
				t.Errorf("%s: 日本語の文字列が i18n.T / i18n.Error の外にあります: %s",
					fset.Position(lit.Pos()), lit.Value)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// untranslated は i18n.T / i18n.Error に包まれていない日本語リテラルを返します。
func untranslated(file *ast.File) []*ast.BasicLit {
	wrapped := map[token.Pos]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.CallExpr:
			if !isI18nCall(v.Fun) {
				return true
			}
			// 包まれた中身はすべて訳し分け済みとみなします。
			ast.Inspect(v, func(inner ast.Node) bool {
				if lit, ok := inner.(*ast.BasicLit); ok {
					wrapped[lit.Pos()] = true
				}
				return true
			})
		case *ast.KeyValueExpr:
			// whatJa / whatEn のように、言語ごとの値を持つ構造体の項目。
			// 表示のたびに i18n.T で選ぶので、リテラル自体は素で書きます。
			key, ok := v.Key.(*ast.Ident)
			if !ok || !(strings.HasSuffix(key.Name, "Ja") || strings.HasSuffix(key.Name, "En")) {
				return true
			}
			if lit, ok := v.Value.(*ast.BasicLit); ok {
				wrapped[lit.Pos()] = true
			}
		}
		return true
	})

	var out []*ast.BasicLit
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || wrapped[lit.Pos()] || !hasJapanese(lit.Value) {
			return true
		}
		out = append(out, lit)
		return true
	})
	return out
}

func isI18nCall(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "i18n" && (sel.Sel.Name == "T" || sel.Sel.Name == "Error")
}

func hasJapanese(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return true
		}
	}
	return false
}

// verbPattern は fmt の書式指定子です。
var verbPattern = regexp.MustCompile(`%[-+# 0]*[\d.*]*[a-zA-Z%]`)

// TestFormatVerbsMatchAcrossLanguages は、対になる和文と英文で書式指定子の
// 並びが一致することを見ます。
//
// ずれても片方の言語では正しく出るので、テストがその言語しか見ていなければ
// 気づけません。実際に出るのは %!s(MISSING) や取り違えた値で、しかも
// 「英語環境の利用者だけ」に出ます。順番の入れ替えは特に踏みやすい
// （日本語の語順で訳すと %s と %d が入れ替わる）。
func TestFormatVerbsMatchAcrossLanguages(t *testing.T) {
	fset := token.NewFileSet()
	for _, dir := range userFacing {
		root := filepath.Join("..", "..", dir)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) != 2 || !isI18nCall(call.Fun) {
					return true
				}
				// 定数以外を渡している呼び出し（spec.Param.Describe など）は
				// 見られないので飛ばします。
				ja, ok1 := constString(call.Args[0])
				en, ok2 := constString(call.Args[1])
				if !ok1 || !ok2 {
					return true
				}
				a, b := verbsOf(ja), verbsOf(en)
				if strings.Join(a, " ") != strings.Join(b, " ") {
					t.Errorf("%s: 書式指定子がずれています\n  ja %v: %s\n  en %v: %s",
						fset.Position(call.Pos()), a, ja, b, en)
				}
				// pflag はバッククォートで囲まれた語をフラグの引数名として
				// 取り出すので、説明文に混ぜると --profile current のような
				// 表示になります。
				if strings.Contains(ja, "`") || strings.Contains(en, "`") {
					t.Errorf("%s: 文言にバッククォートがあります（pflag が引数名と解釈します）", fset.Position(call.Pos()))
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// constString は文字列リテラル（+ で連結したものを含む）を畳み込みます。
func constString(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(v.Value)
		return s, err == nil
	case *ast.BinaryExpr:
		if v.Op != token.ADD {
			return "", false
		}
		left, ok1 := constString(v.X)
		right, ok2 := constString(v.Y)
		return left + right, ok1 && ok2
	}
	return "", false
}

func verbsOf(s string) []string {
	var out []string
	for _, m := range verbPattern.FindAllString(s, -1) {
		if m != "%%" {
			out = append(out, m)
		}
	}
	return out
}
