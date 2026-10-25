package cmd

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// baseURLExample は聞くときに見せる例です。README とスキルの例に合わせて
// あります（別の形を見せると、どれが正しいのか分からなくなる）。
const baseURLExample = "https://example.com/cgi-bin/mt/mt-data-api.cgi"

// newProfileCmd は接続先の管理コマンドです。
func (a *App) newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile",
		Aliases: []string{"profiles"},
		Short:   i18n.T("接続先プロファイルの管理", "Manage connection profiles"),
		Long: i18n.T(
			"MT の接続先を名前で切り替えます。\n\n"+
				"設定は config.json、認証情報は credentials.json に分けて保存されます。\n"+
				"前者は接続先の住所録、後者は 0600 の秘密です。",
			"Switch between MT servers by name.\n\n"+
				"Settings live in config.json and credentials in credentials.json.\n"+
				"The former is an address book; the latter is a 0600 secret."),
	}
	cmd.AddCommand(
		a.newProfileListCmd(),
		a.newProfileAddCmd(),
		a.newProfileSetCmd(),
		a.newProfileUseCmd(),
		a.newProfileShowCmd(),
		a.newProfileRemoveCmd(),
	)
	return cmd
}

func (a *App) newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   i18n.T("プロファイルの一覧", "List the profiles"),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := a.env.Store.Load()
			if err != nil {
				return err
			}
			if len(f.Profiles) == 0 {
				fmt.Fprintf(a.env.Out, i18n.T("プロファイルがありません（%s）\n", "No profiles yet (%s)\n"), a.env.Store.ConfigPath())
				fmt.Fprintln(a.env.Out, i18n.T(
					"mtb profile add <名前> --base-url <URL> で登録してください",
					"Register one with: mtb profile add <name> --base-url <URL>"))
				return nil
			}
			return a.writeProfileTable(a.env.Out, f)
		},
	}
}

// writeProfileTable は一覧を1つ書き出します。追加のときにも同じものを
// 見せるので、描き方を1か所に持ちます。
func (a *App) writeProfileTable(out io.Writer, f *config.File) error {
	creds, err := a.env.Store.LoadCredentials()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\tNAME\tENDPOINT\tSITE\tAUTH")
	for _, name := range f.Names() {
		p := f.Profiles[name]
		marker := " "
		if name == f.Current {
			marker = "*"
		}
		base, version := config.NormalizeBaseURL(p.BaseURL, p.APIVersion)
		auth := i18n.T("未サインイン", "signed out")
		if c, ok := creds[name]; ok && c != nil && (c.SessionID != "" || c.AccessToken != "") {
			auth = i18n.T("サインイン済み", "signed in")
			if c.Expired(a.env.now()) && c.SessionID == "" {
				auth = i18n.T("要サインイン", "sign-in required")
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s/%s\t%s\t%s\n", marker, name, base, version, dash(p.DefaultSiteID), auth)
	}
	return w.Flush()
}

func (a *App) newProfileAddCmd() *cobra.Command {
	var (
		username string
		siteID   string
		use      bool
	)
	cmd := &cobra.Command{
		Use:   i18n.T("add <名前>", "add <name>"),
		Short: i18n.T("プロファイルを追加する", "Add a profile"),
		Long: i18n.T(
			"接続先を登録します。--base-url は mt-data-api.cgi までの URL です。\n"+
				"末尾に /v7 が付いていても取り除いて解釈します。\n\n"+
				"名前は省けます。省くと、いまあるプロファイルを一覧してから\n"+
				"名前と URL を聞きます（対話できない場面ではエラーになります）。\n\n"+
				"登録できたら、そのままサインインするか聞きます。\n\n"+
				"--base-url / --api-version / --client-id / --insecure は共通フラグを使います。\n\n"+
				"例: mtb profile add prod --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi",
			"Register a server. --base-url is the URL up to mt-data-api.cgi.\n"+
				"A trailing /v7 is stripped before it is stored.\n\n"+
				"The name is optional. Leave it off and the existing profiles are listed,\n"+
				"then you are asked for the name and the URL (where there is no terminal\n"+
				"to prompt on, this is an error instead).\n\n"+
				"Once it is registered you are asked whether to sign in straight away.\n\n"+
				"--base-url / --api-version / --client-id / --insecure are the global flags.\n\n"+
				"Example: mtb profile add prod --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi"),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := a.env.Store.Load()
			if err != nil {
				return err
			}
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				if name, err = a.askProfileName(f); err != nil {
					return err
				}
			}
			if _, exists := f.Profiles[name]; exists {
				return fmt.Errorf(i18n.T(
					"プロファイル %q はすでにあります。変更は mtb profile set --profile %s を使ってください",
					"profile %q already exists; change it with mtb profile set --profile %s"), name, name)
			}
			baseURL := a.g.baseURL
			if baseURL == "" && a.env.Prompt != nil && a.env.Prompt.Interactive() {
				// 例を添える。「URL」とだけ聞かれても、サイトのトップと
				// mt-data-api.cgi までのどちらを入れるのか分からない。
				// ここを取り違えたときの症状は全コマンドの 404 なので、
				// 入力の前に見せておく価値がある。
				baseURL, err = a.env.Prompt.Line(i18n.T(
					"Data API の URL（例: "+baseURLExample+"）",
					"Data API URL (e.g. "+baseURLExample+")"), "")
				if err != nil {
					return err
				}
			}
			if baseURL == "" {
				return errors.New(i18n.T(
					"接続先の URL が要ります（--base-url でも指定できます）",
					"the connection URL is required (you can also pass --base-url)"))
			}
			base, version := config.NormalizeBaseURL(baseURL, a.g.apiVersion)
			f.Profiles[name] = &config.Profile{
				BaseURL:            base,
				APIVersion:         version,
				ClientID:           a.g.clientID,
				Username:           username,
				DefaultSiteID:      siteID,
				InsecureSkipVerify: a.g.insecure,
			}
			if use || f.Current == "" {
				f.Current = name
			}
			if err := a.env.Store.Save(f); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Out, i18n.T("プロファイル %s を追加しました（%s/%s）\n", "Added profile %s (%s/%s)\n"), name, base, version)
			if f.Current == name {
				fmt.Fprintln(a.env.Out, i18n.T("既定のプロファイルにしました", "Made it the default profile"))
			}
			// 登録の次はサインイン、というのが普通の流れです。ここで
			// 「次は mtb auth login を実行してください」と突き放すと、
			// 打ち直しが1回増えるだけで誰も得をしません。
			return a.offerSignIn(cmd, name)
		},
	}
	f := cmd.Flags()
	f.StringVar(&username, "username", "", i18n.T("サインイン名（サインイン時の入力を省くため）", "sign-in name (so you are not asked for it at sign-in)"))
	f.StringVar(&siteID, "site-id", "", i18n.T("既定のサイトID", "default site ID"))
	f.BoolVar(&use, "use", false, i18n.T("追加したプロファイルを既定にする", "make the new profile the default"))
	return cmd
}

func (a *App) newProfileSetCmd() *cobra.Command {
	var (
		username string
		siteID   string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: i18n.T("いまのプロファイルの設定を変える", "Change the settings of the current profile"),
		Long: i18n.T(
			"--profile を付けなければ既定のプロファイルを変更します。\n"+
				"指定したフラグだけを書き換えます。\n\n"+
				"--base-url / --api-version / --client-id / --insecure は共通フラグを使います。",
			"Without --profile this changes the default profile.\n"+
				"Only the flags you pass are rewritten.\n\n"+
				"--base-url / --api-version / --client-id / --insecure are the global flags."),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := a.resolve()
			if err != nil {
				return err
			}
			if !r.FromFile {
				return errors.New(i18n.T(
					"設定ファイルにプロファイルがありません。mtb profile add で作ってください",
					"the configuration file has no profiles; create one with mtb profile add"))
			}
			changed := false
			err = a.updateProfile(r.Name, func(p *config.Profile) {
				if cmd.Flags().Changed("base-url") {
					p.BaseURL, p.APIVersion = config.NormalizeBaseURL(a.g.baseURL, p.APIVersion)
					changed = true
				}
				if cmd.Flags().Changed("api-version") {
					p.APIVersion = a.g.apiVersion
					changed = true
				}
				if cmd.Flags().Changed("client-id") {
					p.ClientID = a.g.clientID
					changed = true
				}
				if cmd.Flags().Changed("username") {
					p.Username = username
					changed = true
				}
				if cmd.Flags().Changed("site-id") {
					p.DefaultSiteID = siteID
					changed = true
				}
				if cmd.Flags().Changed("insecure") {
					p.InsecureSkipVerify = a.g.insecure
					changed = true
				}
			})
			if err != nil {
				return err
			}
			if !changed {
				return errors.New(i18n.T(
					"変更するフラグを指定してください（--site-id など）",
					"pass at least one flag to change (--site-id and so on)"))
			}
			fmt.Fprintf(a.env.Out, i18n.T("プロファイル %s を更新しました\n", "Updated profile %s\n"), r.Name)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&username, "username", "", i18n.T("サインイン名", "sign-in name"))
	f.StringVar(&siteID, "site-id", "", i18n.T("既定のサイトID", "default site ID"))
	return cmd
}

func (a *App) newProfileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   i18n.T("use <名前>", "use <name>"),
		Short: i18n.T("既定のプロファイルを切り替える", "Switch the default profile"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := a.env.Store.Load()
			if err != nil {
				return err
			}
			if _, ok := f.Profiles[args[0]]; !ok {
				return fmt.Errorf(i18n.T("プロファイル %q がありません（%s）", "no such profile: %q (%s)"), args[0], joinNames(f.Names()))
			}
			f.Current = args[0]
			if err := a.env.Store.Save(f); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Out, i18n.T("既定のプロファイルを %s にしました\n", "The default profile is now %s\n"), args[0])
			return nil
		},
	}
}

func (a *App) newProfileShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: i18n.T("いま使う接続先の解決結果を表示する", "Show the resolved connection settings"),
		Long: i18n.T(
			"フラグ・環境変数・プロファイルを重ねた結果を表示します。\n"+
				"「どのサーバを叩こうとしているのか」が分からなくなったときに使います。\n"+
				"カレントディレクトリの .env を読んでいれば、その場所も出ます。",
			"Show the result of layering flags, environment variables and the profile.\n"+
				"Use it when you have lost track of which server you are about to call.\n"+
				"If a .env in the current directory was read, its path is shown too."),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := a.resolve()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(a.env.Out, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, i18n.T("プロファイル\t%s\n", "Profile\t%s\n"), r.Name)
			fmt.Fprintf(w, i18n.T("ベースURL\t%s\n", "Base URL\t%s\n"), r.BaseURL)
			fmt.Fprintf(w, i18n.T("APIバージョン\t%s\n", "API version\t%s\n"), r.APIVersion)
			fmt.Fprintf(w, i18n.T("クライアントID\t%s\n", "Client ID\t%s\n"), r.ClientID)
			fmt.Fprintf(w, i18n.T("既定サイトID\t%s\n", "Default site ID\t%s\n"), dash(r.SiteID))
			fmt.Fprintf(w, i18n.T("ユーザー名\t%s\n", "Username\t%s\n"), dash(r.Username))
			fmt.Fprintf(w, i18n.T("設定ファイル\t%s\n", "Configuration file\t%s\n"), a.env.Store.ConfigPath())
			fmt.Fprintf(w, i18n.T("認証情報\t%s\n", "Credentials\t%s\n"), a.env.Store.CredentialsPath())
			// .env は置いてあるだけで接続先を変えられるので、ここに出す。
			// 「どのサーバを叩こうとしているのか」を答えるのがこのコマンドの
			// 役目なのに、根拠のひとつが見えないのでは答えたことにならない。
			if a.env.DotEnvPath != "" {
				fmt.Fprintf(w, i18n.T("読み込んだ .env\t%s\n", "Loaded .env\t%s\n"), a.env.DotEnvPath)
			}
			return w.Flush()
		},
	}
}

func (a *App) newProfileRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     i18n.T("remove <名前>", "remove <name>"),
		Aliases: []string{"rm"},
		Short:   i18n.T("プロファイルと、その認証情報を消す", "Delete a profile and its credentials"),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			f, err := a.env.Store.Load()
			if err != nil {
				return err
			}
			if _, ok := f.Profiles[name]; !ok {
				return fmt.Errorf(i18n.T("プロファイル %q がありません", "no such profile: %q"), name)
			}
			ok, err := a.confirm(fmt.Sprintf(i18n.T(
				"プロファイル %s と保存済みの認証情報を消します。よろしいですか？",
				"This deletes profile %s and its stored credentials. Continue?"), name))
			if err != nil {
				return err
			}
			if !ok {
				return errAborted()
			}
			delete(f.Profiles, name)
			if f.Current == name {
				f.Current = ""
				if names := f.Names(); len(names) > 0 {
					f.Current = names[0]
				}
			}
			if err := a.env.Store.Save(f); err != nil {
				return err
			}
			if err := a.env.Store.DeleteCredential(name); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Out, i18n.T("プロファイル %s を削除しました\n", "Deleted profile %s\n"), name)
			return nil
		},
	}
}

// updateProfile は1つのプロファイルを読み書きします。
func (a *App) updateProfile(name string, mutate func(*config.Profile)) error {
	f, err := a.env.Store.Load()
	if err != nil {
		return err
	}
	p, ok := f.Profiles[name]
	if !ok {
		return fmt.Errorf(i18n.T("プロファイル %q がありません", "no such profile: %q"), name)
	}
	mutate(p)
	return a.env.Store.Save(f)
}

// offerSignIn は追加したプロファイルで、そのままサインインするか聞きます。
//
// **対話できない場面では聞きません。** エージェントや CI では黙って止まる
// のが一番たちが悪いので、次にやることだけ伝えて終わります。
func (a *App) offerSignIn(cmd *cobra.Command, name string) error {
	if a.env.Prompt == nil || !a.env.Prompt.Interactive() {
		fmt.Fprintln(a.env.Out, i18n.T("次は mtb auth login --profile ", "Next: mtb auth login --profile ")+name)
		return nil
	}
	ok, err := a.env.Prompt.Confirm(i18n.T(
		"このままサインインしますか？", "Sign in now?"), true)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(a.env.Out, i18n.T("次は mtb auth login --profile ", "Next: mtb auth login --profile ")+name)
		return nil
	}
	// 追加したプロファイルを相手にします。--profile が付いていても、
	// いま作ったものにサインインするのが意図のはずです。
	a.g.profile = name
	a.resolved = nil
	return a.signIn(signInOptions{})
}

// askProfileName は名前を省いたときに聞きます。
//
// 先に既存の一覧を見せるのは、名前を思いつくためだけではありません。
// 「もう prod がある」「いま向いているのは stg だ」が同時に分かるので、
// 同じサーバを二重に登録したり、本番に向いたまま作業を続けたりしにくくなります。
func (a *App) askProfileName(f *config.File) (string, error) {
	if len(f.Profiles) > 0 {
		fmt.Fprintln(a.env.Out, i18n.T("いまあるプロファイル:", "Profiles you already have:"))
		if err := a.writeProfileTable(a.env.Out, f); err != nil {
			return "", err
		}
		fmt.Fprintln(a.env.Out)
	}
	if a.env.Prompt == nil || !a.env.Prompt.Interactive() {
		return "", errors.New(i18n.T(
			"プロファイル名が要ります（例: mtb profile add prod）",
			"a profile name is required (for example: mtb profile add prod)"))
	}
	name, err := a.env.Prompt.Line(i18n.T(
		"プロファイル名（例: prod）",
		"Profile name (e.g. prod)"), "")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New(i18n.T("プロファイル名が要ります", "a profile name is required"))
	}
	return name, nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return i18n.T("登録なし", "none registered")
	}
	sort.Strings(names)
	out := i18n.T("登録済み: ", "registered: ")
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}
