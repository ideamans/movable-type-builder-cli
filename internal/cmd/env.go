package cmd

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/extension"
	"github.com/ideamans/movable-type-builder-cli/internal/prompt"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// Env は CLI が外の世界に触れる口をすべて集めたものです。
//
// ここを差し替えるだけでテストが成立するように、標準入出力・設定ファイル・
// HTTP 送信・時刻・対話入力のいずれもフィールドとして持たせています。
// コマンドの実装から os や net/http を直接触らないこと。
type Env struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer

	// Store は設定と認証情報の置き場です。
	Store *config.Store
	// Prompt は対話入力です。
	Prompt prompt.Prompter
	// Now は現在時刻です。
	Now func() time.Time
	// Sleep は待ち time.Sleep です。再構築の呼び直しの間隔に使います。
	// テストでは何もしない関数に差し替えて、待たずに回します。
	Sleep func(d time.Duration)
	// NewDoer は接続先ごとの HTTP 送信手段を作ります。
	NewDoer func(insecure bool, timeout time.Duration) api.Doer
	// Spec は埋め込み済みの API 定義です。
	Spec *spec.Spec
	// ReadFile は `--entry @file.json` や添付ファイルの読み込みです。
	// "-" は標準入力を指します。
	ReadFile func(name string) ([]byte, error)
	// WriteFile は --out での書き出しです。
	WriteFile func(name string, data []byte) error
	// ListFiles は dir 以下のファイルを再帰的に列挙します。返すのは dir
	// からの相対パス（スラッシュ区切り・辞書順）です。--dir の一括
	// アップロードだけがこれを使います。
	ListFiles func(dir string) ([]string, error)
	// Getenv は環境変数の読み取りです。未設定なら Store のものを使います。
	// Store を持たない経路（カタログ生成器）のための逃げ道です。
	Getenv func(key string) string
	// Extensions は見つけた拡張です。ツリーに足せるかは cmd.New が決めます。
	Extensions []*extension.Extension
	// ExtensionErrors は拡張を読み込めなかった理由です。起動のたびに
	// 標準エラーへ出します（黙って消えると、コマンドが無い理由が分からない）。
	ExtensionErrors []error
	// ExtensionsDir は mtb extensions install が入れる先です。
	ExtensionsDir string
	// RemoveAll はディレクトリを丸ごと消します。拡張の入れ直しと外すときに使います。
	RemoveAll func(path string) error
	// DotEnvPath は読み込んだ .env の場所です。読んでいなければ空。
	// 表示のためだけに持ちます（mtb profile show）。
	DotEnvPath string
}

// DefaultEnv は実際の OS を相手にする Env を作ります。
func DefaultEnv() (*Env, error) {
	environ, dotEnvPath := environWithDotEnv()
	store, err := config.Open(environ, os.UserHomeDir)
	if err != nil {
		return nil, err
	}
	env := &Env{
		DotEnvPath: dotEnvPath,
		In:         os.Stdin,
		Out:        os.Stdout,
		Err:        os.Stderr,
		Store:      store,
		Prompt:     prompt.NewTerminal(os.Stdin, os.Stderr),
		Now:        time.Now,
		Sleep:      time.Sleep,
		NewDoer: func(insecure bool, timeout time.Duration) api.Doer {
			return api.NewHTTPClient(insecure, timeout)
		},
		Spec: spec.MustLoad(),
		// 親ディレクトリが無ければ作ります。templates pull --dir も --out も
		// 「まだ無い場所」に書き出すのが普通の使い方で、そのたびに mkdir を
		// 強いるのは無駄です。
		WriteFile: func(name string, data []byte) error {
			if dir := filepath.Dir(name); dir != "" && dir != "." {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
			}
			return os.WriteFile(name, data, 0o644)
		},
	}
	env.ListFiles = func(dir string) ([]string, error) {
		var out []string
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			out = append(out, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			return nil, err
		}
		sort.Strings(out)
		return out, nil
	}
	env.ReadFile = func(name string) ([]byte, error) {
		if name == "-" {
			return io.ReadAll(env.In)
		}
		return os.ReadFile(name)
	}
	env.RemoveAll = os.RemoveAll
	if dir, err := config.DefaultDir(environ, os.UserHomeDir); err == nil {
		env.ExtensionsDir = filepath.Join(dir, "extensions")
	}
	env.Extensions, env.ExtensionErrors = discoverExtensions(env.ExtensionsDir, environ("MTB_EXTENSIONS_PATH"))
	return env, nil
}

// discoverExtensions は拡張を探して読み込みます。
//
// MTB_EXTENSIONS_PATH（開発中の拡張のディレクトリを並べたもの）を先に、
// インストール先を後に見ます。同じ名前なら先に見つけた方を使うので、
// 入れたものを手元の作りかけで上書きして試せます。
func discoverExtensions(installedDir, pathList string) ([]*extension.Extension, []error) {
	type candidate struct{ dir, source string }
	var candidates []candidate
	for _, dir := range filepath.SplitList(pathList) {
		if dir != "" {
			candidates = append(candidates, candidate{dir, extension.SourcePath})
		}
	}
	if installedDir != "" {
		entries, _ := os.ReadDir(installedDir)
		for _, e := range entries {
			if e.IsDir() {
				candidates = append(candidates, candidate{filepath.Join(installedDir, e.Name()), extension.SourceInstalled})
			}
		}
	}
	var (
		found []*extension.Extension
		errs  []error
		seen  = map[string]bool{}
	)
	for _, c := range candidates {
		e, err := extension.Load(os.ReadFile, c.dir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		e.Source = c.source
		found = append(found, e)
	}
	return found, errs
}

// environWithDotEnv はカレントディレクトリの .env をプロセスの環境変数の
// 後ろに重ねます。読み込んだ場所も返します（表示のため）。
//
// **プロセスに書き出しはしません。** os.Setenv で撒くと、この CLI が
// 起動する他のプロセスにまで漏れます。読み取りの層として重ねるだけです。
func environWithDotEnv() (config.Environ, string) {
	data, err := os.ReadFile(config.DotEnvFile)
	if err != nil {
		return config.OSEnv, ""
	}
	vars := config.ParseDotEnv(data)
	if len(vars) == 0 {
		return config.OSEnv, ""
	}
	path, err := filepath.Abs(config.DotEnvFile)
	if err != nil {
		path = config.DotEnvFile
	}
	return config.DotEnv(config.OSEnv, vars), path
}

// getenv は環境変数の読み取りです。Store と同じ経路を使うので、
// テストで withEnvVars を使えば表示言語もそこで決まります。
func (e *Env) getenv(key string) string {
	if e.Getenv != nil {
		return e.Getenv(key)
	}
	if e.Store != nil {
		return e.Store.Env(key)
	}
	return ""
}

// sleep は Env.Sleep が未設定でも動くようにするための補助です。
func (a *App) sleep(d time.Duration) {
	if a.env != nil && a.env.Sleep != nil {
		a.env.Sleep(d)
		return
	}
	time.Sleep(d)
}

// now は Env.Now が未設定でも動くようにするための補助です。
func (e *Env) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now()
}
