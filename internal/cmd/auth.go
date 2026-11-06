package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// newAuthCmd は認証まわりのコマンドです。
//
// 「気軽に」の中身は3つです。login は足りないものだけを聞く、パスワードは
// 保存せずセッションIDだけを持つ、アクセストークンは失効したら黙って
// 取り直す。日常の操作で auth を意識しなくて済むのが目標です。
func (a *App) newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: i18n.T("認証（サインイン・状態確認・サインアウト）", "Authentication (sign in, status, sign out)"),
	}
	cmd.AddCommand(a.newAuthLoginCmd(), a.newAuthStatusCmd(), a.newAuthRefreshCmd(), a.newAuthLogoutCmd())
	return cmd
}

func (a *App) newAuthLoginCmd() *cobra.Command {
	var (
		username      string
		password      string
		passwordStdin bool
		noRemember    bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: i18n.T("ユーザー名とパスワードでサインインする", "Sign in with a username and password"),
		Long: i18n.T(
			"ユーザー名とパスワードで認証し、セッションIDとアクセストークンを保存します。\n\n"+
				"パスワードは MT の「Webサービスパスワード」です。サインイン用の\n"+
				"パスワードとは別物で、MT のユーザー情報の編集画面で発行します（MT 9 では\n"+
				"発行したときにしか表示されません。分からなくなったら削除して発行し直します）。\n\n"+
				"パスワードは保存しません。保存するのはセッションIDとアクセストークンで、\n"+
				"アクセストークンが切れたらセッションIDから自動で取り直します。\n\n"+
				"対話できない環境では --username と --password-stdin を使ってください。\n\n"+
				"  pass show mt/prod | mtb auth login --username alice --password-stdin\n\n"+
				"パスワードの渡し方は3つあります。--password-stdin はどこにも残りません。\n"+
				"MTB_PASSWORD は同じホストの他のプロセスから見えます。--password は\n"+
				"プロセス一覧に残るので、他に手が無いときだけにしてください。",
			"Authenticate with a username and password, then store the session ID and access token.\n\n"+
				"The password is the MT Web Services Password. It is separate from the\n"+
				"password you sign in to MT with, and is issued on your user profile screen\n"+
				"in MT (MT 9 shows it only once, when issued; if it is lost, delete it and\n"+
				"issue a new one).\n\n"+
				"The password itself is never stored. What is stored is the session ID and\n"+
				"the access token; when the access token expires it is renewed automatically\n"+
				"from the session ID.\n\n"+
				"Where there is no terminal to prompt on, use --username with --password-stdin.\n\n"+
				"  pass show mt/prod | mtb auth login --username alice --password-stdin\n\n"+
				"There are three ways to hand over the password. --password-stdin leaves no\n"+
				"trace. MTB_PASSWORD is visible to other processes on the same host.\n"+
				"--password shows up in the process list, so keep it for when nothing else works."),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.signIn(signInOptions{
				username:      username,
				password:      password,
				passwordStdin: passwordStdin,
				noRemember:    noRemember,
			})
		},
	}
	cmd.Flags().StringVar(&username, "username", "", i18n.T("サインイン名", "sign-in name"))
	cmd.Flags().StringVar(&password, "password", "", i18n.T("Webサービスパスワード（プロセス一覧に残ります。--password-stdin を推奨）", "Web Services Password (this stays in the process list; prefer --password-stdin)"))
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, i18n.T("Webサービスパスワードを標準入力から読む", "read the Web Services Password from standard input"))
	cmd.Flags().BoolVar(&noRemember, "no-remember", false, i18n.T("永続セッションを作らない（再サインインが頻繁になります）", "do not create a persistent session (you will sign in more often)"))
	return cmd
}

// signInOptions は mtb auth login の入力です。
type signInOptions struct {
	username      string
	password      string
	passwordStdin bool
	noRemember    bool
}

// signIn はサインインの本体です。**mtb profile add からも呼びます**——
// 接続先を登録した直後にサインインするのが普通の流れで、そこで
// 「次は mtb auth login を実行してください」と突き放す理由がありません。
func (a *App) signIn(opts signInOptions) error {
	username, password := opts.username, opts.password
	r, err := a.resolve()
	if err != nil {
		return err
	}

	if username == "" {
		username = r.Username
	}
	if username == "" && a.env.Prompt != nil && a.env.Prompt.Interactive() {
		username, err = a.env.Prompt.Line(i18n.T("ユーザー名", "Username"), "")
		if err != nil {
			return err
		}
	}
	if username == "" {
		return errors.New(i18n.T(
			"ユーザー名が要ります。--username で指定してください",
			"a username is required; pass it with --username"))
	}

	if opts.passwordStdin {
		if password != "" {
			return usagef("%s", i18n.T(
				"--password と --password-stdin は同時に使えません",
				"--password and --password-stdin cannot be used together"))
		}
		if password, err = a.readPasswordFromStdin(); err != nil {
			return err
		}
	}
	if password == "" {
		password = a.env.Store.Env("MTB_PASSWORD")
	}
	if password == "" && a.env.Prompt != nil && a.env.Prompt.Interactive() {
		password, err = a.env.Prompt.Secret(i18n.T("Webサービスパスワード", "Web Services Password"))
		if err != nil {
			return err
		}
	}
	if password == "" {
		return errors.New(i18n.T(
			"パスワードが要ります。--password-stdin か MTB_PASSWORD で渡してください",
			"a password is required; pass it with --password-stdin or via MTB_PASSWORD"))
	}

	client, err := a.client()
	if err != nil {
		return err
	}
	cred, err := client.Login(context.Background(), username, password, !opts.noRemember)
	if err != nil {
		return err
	}

	// 次回の入力を省けるように、ユーザー名だけプロファイルに控えます。
	if r.FromFile && r.Username != username {
		if err := a.updateProfile(r.Name, func(p *config.Profile) { p.Username = username }); err != nil {
			return err
		}
	}

	fmt.Fprintf(a.env.Out, i18n.T(
		"%s にサインインしました（プロファイル %s / ユーザー %s）\n",
		"Signed in to %s (profile %s, user %s)\n"),
		client.Endpoint(), r.Name, username)
	a.printExpiry(cred)
	return nil
}

// readPasswordFromStdin は標準入力からパスワードを読みます。
//
// 末尾の改行だけを落とします。`echo pw |` は改行を付けますが、
// パスワードの先頭や途中の空白は落としません（消してよいものか
// こちらでは判断できないため）。
func (a *App) readPasswordFromStdin() (string, error) {
	if a.env.In == nil {
		return "", errors.New(i18n.T("標準入力を読めません", "cannot read standard input"))
	}
	buf, err := io.ReadAll(a.env.In)
	if err != nil {
		return "", err
	}
	password := strings.TrimRight(string(buf), "\r\n")
	if password == "" {
		return "", errors.New(i18n.T(
			"標準入力が空です",
			"standard input was empty"))
	}
	return password, nil
}

func (a *App) newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: i18n.T("いまの接続先と認証状態を表示する", "Show the current server and authentication state"),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := a.resolve()
			if err != nil {
				return err
			}
			fmt.Fprintf(a.env.Out, i18n.T("プロファイル: %s\n", "Profile: %s\n"), r.Name)
			fmt.Fprintf(a.env.Out, i18n.T("エンドポイント: %s/%s\n", "Endpoint: %s/%s\n"), r.BaseURL, r.APIVersion)
			if r.SiteID != "" {
				fmt.Fprintf(a.env.Out, i18n.T("既定サイトID: %s\n", "Default site ID: %s\n"), r.SiteID)
			}
			if token := a.env.Store.Env("MTB_ACCESS_TOKEN"); token != "" {
				fmt.Fprintln(a.env.Out, i18n.T(
					"認証: MTB_ACCESS_TOKEN（設定ファイルは使いません）",
					"Authentication: MTB_ACCESS_TOKEN (the configuration file is not used)"))
				return nil
			}
			cred, err := a.env.Store.Credential(r.Name)
			if err != nil {
				return err
			}
			if cred.SessionID == "" && cred.AccessToken == "" {
				fmt.Fprintln(a.env.Out, i18n.T(
					"認証: 未サインイン（mtb auth login）",
					"Authentication: not signed in (mtb auth login)"))
				return nil
			}
			if cred.Username != "" {
				fmt.Fprintf(a.env.Out, i18n.T("ユーザー: %s\n", "User: %s\n"), cred.Username)
			}
			fmt.Fprintf(a.env.Out, i18n.T("セッション: %s\n", "Session: %s\n"), presence(cred.SessionID != ""))
			a.printExpiry(cred)
			return nil
		},
	}
}

func (a *App) newAuthRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: i18n.T("アクセストークンを取り直す", "Renew the access token"),
		Long: i18n.T(
			"セッションIDから新しいアクセストークンを取得します。\n"+
				"通常は失効時に自動で行われるので、明示的に叩く必要はありません。",
			"Obtain a new access token from the session ID.\n"+
				"This normally happens automatically on expiry, so you rarely need to run it."),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			cred, err := client.Refresh(context.Background())
			if err != nil {
				return err
			}
			fmt.Fprintln(a.env.Out, i18n.T("アクセストークンを更新しました", "Renewed the access token"))
			a.printExpiry(cred)
			return nil
		},
	}
}

func (a *App) newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: i18n.T("サインアウトしてセッションを無効にする", "Sign out and invalidate the session"),
		Long: i18n.T(
			"MT 側のセッションを無効にし、保存してある認証情報を消します。\n"+
				"同じセッションから発行された他のアクセストークンもまとめて無効になります。",
			"Invalidate the session on the MT side and delete the stored credentials.\n"+
				"Every other access token issued from the same session is invalidated too."),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := a.resolve()
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			if err := client.Logout(context.Background()); err != nil {
				return err
			}
			if err := a.env.Store.DeleteCredential(r.Name); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Out, i18n.T("プロファイル %s からサインアウトしました\n", "Signed out of profile %s\n"), r.Name)
			return nil
		},
	}
}

func (a *App) printExpiry(cred *config.Credential) {
	if cred == nil || cred.AccessToken == "" {
		return
	}
	if cred.ExpiresAt == 0 {
		fmt.Fprintln(a.env.Out, i18n.T(
			"アクセストークン: あり（失効時刻は不明）",
			"Access token: present (expiry unknown)"))
		return
	}
	left := cred.ExpiresIn(a.env.now())
	fmt.Fprintf(a.env.Out, i18n.T(
		"アクセストークン: %s まで有効（残り %s）\n",
		"Access token: valid until %s (%s left)\n"),
		time.Unix(cred.ExpiresAt, 0).Format(time.RFC3339), left.Round(time.Second))
}

func presence(ok bool) string {
	if ok {
		return i18n.T("あり", "yes")
	}
	return i18n.T("なし", "no")
}
