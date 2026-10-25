// Package cmd は mtb のコマンドツリーを組み立てます。
//
// ツリーの大半は internal/spec に埋め込んだ OpenAPI 定義から実行時に作られ、
// 手書きなのは auth / profile / api / llm の4つだけです。「公式の定義に無い
// コマンドは存在しない」状態を保つのが狙いで、MT 側が増えたらリファレンスを
// 取り込み直すだけで CLI が追いつきます。
package cmd

import (
	"errors"
	"strings"
	"time"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"
	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/extension"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// PluginVersion はこのCLIのリリースバージョンです。
// plugins/movable-type-builder-cli/.claude-plugin/plugin.json の version と
// 一致していることをテストが、git タグと一致していることをリリース
// ワークフローが検査します。
const PluginVersion = "0.1.0"

// Version はビルド時に -ldflags で上書きされるバージョン文字列です。
var Version = PluginVersion

// globals は全コマンド共通のフラグです。
type globals struct {
	profile        string
	baseURL        string
	apiVersion     string
	clientID       string
	format         string
	compact        bool
	outFile        string
	extract        string
	timeout        time.Duration
	insecure       bool
	methodOverride bool
	verbose        bool
	dryRun         bool
	assumeYes      bool
	lang           string
	verifyPublish  bool
	auditNote      string
}

// App はコマンドツリーと、その実行に必要な依存をまとめたものです。
type App struct {
	env  *Env
	g    globals
	root *cobra.Command

	resolved  *config.Resolved
	apiClient *api.Client

	// extensionOf は拡張が足したコマンドと、その出どころです。
	extensionOf map[*cobra.Command]*extension.Extension
	// loadedExtensions はツリーに足せた拡張、extensionErrors は足せなかった理由です。
	loadedExtensions []*extension.Extension
	extensionErrors  []error
}

// New はコマンドツリーを組み立てます。実行はしません。
// args は --lang を先読みするためだけに受け取ります。ヘルプの文言は
// ここで確定してしまうので、cobra のフラグ解析を待っていては間に合いません。
func New(env *Env, args ...string) *App {
	a := &App{env: env}
	a.g.lang = langFromArgs(args)
	i18n.SetLang(i18n.Detect(a.g.lang, env.getenv))

	if env.Spec == nil {
		env.Spec = spec.MustLoad()
	}

	a.root = &cobra.Command{
		Use:   "mtb",
		Short: i18n.T("Movable Type でサイトを組み立てる CLI", "A CLI for building Movable Type sites"),
		Long: i18n.T(
			"Movable Type のサイトを、コマンドラインから組み立てます。\n"+
				"接続先はプロファイルで切り替え、認証は mtb auth login で済ませます。\n\n"+
				"リソース操作のコマンドは Data API (v7) の公式定義から作られているので、\n"+
				"リファレンスのパスとコマンド名が1対1で対応します。それに加えて、\n"+
				"API に無い「まとめて送る・組み立てる・取り出す」を mtb が受け持ちます\n"+
				"（--dir・--set・--extract・templates pull/push など）。",
			"Build Movable Type sites from the command line.\n"+
				"Switch between servers with profiles; sign in once with mtb auth login.\n\n"+
				"The resource commands are generated from the official Data API (v7) definition,\n"+
				"so command names map one-to-one onto the paths in the reference. On top of that,\n"+
				"mtb supplies the bulk, compose and extract operations the API does not have\n"+
				"(--dir, --set, --extract, templates pull/push and so on)."),
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// フラグの綴り間違いは「呼び出し方の誤り」として終了コードを分けます。
	a.root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &usageError{err}
	})
	a.root.SetIn(env.In)
	a.root.SetOut(env.Out)
	a.root.SetErr(env.Err)

	f := a.root.PersistentFlags()
	f.StringVar(&a.g.profile, "profile", "", i18n.T("使うプロファイル名（既定は config.json の current）", "profile to use (defaults to the current profile in config.json)"))
	f.StringVar(&a.g.baseURL, "base-url", "", i18n.T("mt-data-api.cgi までの URL（プロファイルより優先）", "URL up to mt-data-api.cgi (overrides the profile)"))
	f.StringVar(&a.g.apiVersion, "api-version", "", i18n.T("エンドポイントのバージョン（既定 "+config.DefaultAPIVersion+"）", "endpoint version (default "+config.DefaultAPIVersion+")"))
	f.StringVar(&a.g.clientID, "client-id", "", i18n.T("認証時に名乗る clientId（既定 "+config.DefaultClientID+"）", "clientId to present when signing in (default "+config.DefaultClientID+")"))
	f.StringVar(&a.g.format, "format", "auto", i18n.T("出力形式 auto|json|raw", "output format: auto|json|raw"))
	f.BoolVar(&a.g.compact, "compact", false, i18n.T("JSON を1行で出す", "print JSON on one line"))
	f.StringVar(&a.g.outFile, "out", "", i18n.T("応答をファイルに書き出す", "write the response to a file"))
	f.StringVar(&a.g.extract, "extract", "", i18n.T(
		"応答から1つの項目だけを生の値で取り出す（text、items.0.name のように辿れます）",
		"take a single field out of the response as its raw value (dotted paths like items.0.name work)"))
	f.DurationVar(&a.g.timeout, "timeout", 60*time.Second, i18n.T("1リクエストのタイムアウト", "timeout for one request"))
	f.BoolVar(&a.g.insecure, "insecure", false, i18n.T("TLS 証明書を検証しない", "do not verify the TLS certificate"))
	f.BoolVar(&a.g.methodOverride, "method-override", false, i18n.T("PUT/DELETE を POST + __method で送る", "send PUT/DELETE as POST with __method"))
	f.BoolVarP(&a.g.verbose, "verbose", "v", false, i18n.T("送受信の要約を標準エラーに出す", "print a summary of each request to stderr"))
	f.BoolVar(&a.g.dryRun, "dry-run", false, i18n.T("送信せずに、送るはずだった内容を表示する", "show what would be sent, without sending it"))
	f.BoolVarP(&a.g.assumeYes, "yes", "y", false, i18n.T("破壊的な操作の確認を省く", "skip the confirmation for destructive operations"))
	f.StringVar(&a.g.lang, "lang", a.g.lang, i18n.T("表示言語 ja|en（既定は環境変数から判定）", "display language: ja|en (detected from the environment by default)"))
	f.StringVar(&a.g.auditNote, auditNoteFlag, "", i18n.T(
		"何の作業かの説明を X-MTB-Audit-Note ヘッダで送る（サーバが監査ログを取っていれば一緒に残る。既定は MTB_AUDIT_NOTE）",
		"describe what this work is for, sent as the X-MTB-Audit-Note header (kept with the server's audit log if it has one; defaults to MTB_AUDIT_NOTE)"))
	f.BoolVar(&a.g.verifyPublish, verifyPublishFlag, false, i18n.T(
		"保存できたかだけでなく、静的ファイルを書き出せたかまで MT のログで確かめる（失敗なら終了コード 8）",
		"check the MT log to confirm the static files were published, not just saved (exit code 8 if they were not)"))

	a.root.AddCommand(
		a.newAuthCmd(),
		a.newProfileCmd(),
		a.newAPICmd(),
		a.newDumpCmd(),
		a.newExtensionsCmd(),
	)
	a.addGeneratedCommands()
	a.addTemplateSyncCommands()
	// 拡張は最後に足します。手書きのコマンドも含めた全部と、名前が
	// ぶつからないかを確かめるためです。
	a.addExtensionCommands()
	a.warnExtensionErrors()
	llmcmd.AddTo(a.root, LLMConfig(a.loadedExtensions...))
	return a
}

// Root は組み立て済みのコマンドツリーを返します。
func (a *App) Root() *cobra.Command { return a.root }

// Execute はコマンドを実行します。
func (a *App) Execute(args []string) error {
	a.root.SetArgs(args)
	return a.root.Execute()
}

// Root はカタログ生成器のための入口です。実行せずにツリーだけを返します。
// 表示言語は、呼び出し側が i18n.SetLang で決めたものをそのまま使います
// （生成器を動かす環境の言語で生成物が入れ替わらないように）。
func Root() *cobra.Command {
	lang := string(i18n.Current())
	return New(&Env{
		Spec: spec.MustLoad(),
		Getenv: func(key string) string {
			if key == "MTB_LANG" {
				return lang
			}
			return ""
		},
	}).Root()
}

// resolve は接続先を決めます。1回のプロセスで1度だけ解決します。
func (a *App) resolve() (*config.Resolved, error) {
	if a.resolved != nil {
		return a.resolved, nil
	}
	if a.env.Store == nil {
		return nil, errors.New(i18n.T("設定の読み書きができません", "cannot read or write the configuration"))
	}
	r, err := a.env.Store.Resolve(config.Overrides{
		Profile:    a.g.profile,
		BaseURL:    a.g.baseURL,
		APIVersion: a.g.apiVersion,
		ClientID:   a.g.clientID,
	})
	if err != nil {
		return nil, err
	}
	a.resolved = r
	return r, nil
}

// storeTokens は解決済みプロファイルに紐づく認証情報の置き場です。
type storeTokens struct {
	store   *config.Store
	profile string
	cached  *config.Credential
}

func (s *storeTokens) Credential() (*config.Credential, error) {
	if s.cached != nil {
		return s.cached, nil
	}
	c, err := s.store.Credential(s.profile)
	if err != nil {
		return nil, err
	}
	s.cached = c
	return c, nil
}

func (s *storeTokens) Save(c *config.Credential) error {
	s.cached = c
	return s.store.PutCredential(s.profile, c)
}

// client は接続先に対応するクライアントを返します。
//
// MTB_ACCESS_TOKEN があるときはそれだけを使い、設定ファイルには一切
// 書き戻しません。CI やコンテナから使うときの経路です。
func (a *App) client() (*api.Client, error) {
	if a.apiClient != nil {
		return a.apiClient, nil
	}
	r, err := a.resolve()
	if err != nil {
		return nil, err
	}

	var tokens api.TokenStore
	if token := a.env.Store.Env("MTB_ACCESS_TOKEN"); token != "" {
		tokens = &api.MemoryTokenStore{Cred: &config.Credential{AccessToken: token}}
	} else {
		tokens = &storeTokens{store: a.env.Store, profile: r.Name}
	}

	c := &api.Client{
		BaseURL:        r.BaseURL,
		APIVersion:     r.APIVersion,
		ClientID:       r.ClientID,
		HTTP:           a.env.NewDoer(r.Insecure || a.g.insecure, a.g.timeout),
		Tokens:         tokens,
		Now:            a.env.now,
		MethodOverride: a.g.methodOverride,
		UserAgent:      "mtb/" + Version,
		AuditNote:      a.auditNote(),
	}
	if a.g.verbose {
		c.Trace = a.env.Err
	}
	a.apiClient = c
	return c, nil
}

// auditNoteFlag は作業の説明を送るフラグです。
const auditNoteFlag = "audit-note"

// auditNote は送る作業の説明です。フラグが環境変数に勝ちます。
//
// エージェントが mtb を何十回と叩く作業では、あとで監査ログを見直すときに
// 「どの依頼の操作だったか」が分からないと追えません。1回の作業の間は
// MTB_AUDIT_NOTE に入れておけば、全部の呼び出しに同じ説明が付きます。
func (a *App) auditNote() string {
	if a.g.auditNote != "" {
		return a.g.auditNote
	}
	return a.env.getenv("MTB_AUDIT_NOTE")
}

// SetLang は表示言語を決めます。cmd.New より前に文言を出す経路
// （設定の読み込み失敗や非推奨の --llm）のために公開しています。
func SetLang(args []string, getenv func(string) string) {
	i18n.SetLang(i18n.Detect(langFromArgs(args), getenv))
}

// langFromArgs は --lang を cobra より先に拾います。ヘルプの文言は
// コマンドを組み立てる時点で確定するため、フラグ解析を待てません。
func langFromArgs(args []string) string {
	for i, a := range args {
		switch {
		case a == "--lang" && i+1 < len(args):
			return args[i+1]
		case strings.HasPrefix(a, "--lang="):
			return strings.TrimPrefix(a, "--lang=")
		}
	}
	return ""
}
