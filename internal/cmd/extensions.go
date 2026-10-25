package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/extension"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// 拡張（エクステンション）は、MT のプラグインが足した Data API の
// エンドポイントに、生成コマンドと同じ形のコマンドを与えるものです。
// 仕組みは internal/extension のパッケージ文書にあります。
//
// 「プラグイン」と呼ばないのは、トップレベルの plugins（と別名の plugin）が
// MT のプラグインを操作するコマンドとしてすでにあるためです。MT 側は
// プラグイン、mtb 側は拡張、と呼び分けます。

// extensionConflict は拡張のコマンドが既存のコマンドとぶつかったことです。
type extensionConflict struct {
	ext  string
	path string
	with string
}

func (c *extensionConflict) Error() string {
	return fmt.Sprintf(i18n.T(
		"拡張 %s の %s は、既存の %s とぶつかるので読み込みません",
		"extension %s: %s collides with the existing %s, so the extension is not loaded"), c.ext, c.path, c.with)
}

// addExtensionCommands は読み込み済みの拡張のコマンドをツリーに足します。
//
// 足す場所は生成コマンドと同じ木の中です（利用者はその操作を探しに来る
// 場所で探す）。名前がぶつかる拡張は**丸ごと**読み込みません。一部だけ
// 入ると、どのコマンドが拡張のものなのか分からなくなるためです。
func (a *App) addExtensionCommands() {
	a.extensionOf = map[*cobra.Command]*extension.Extension{}
	for _, e := range a.env.Extensions {
		if err := a.attachExtension(e); err != nil {
			a.extensionErrors = append(a.extensionErrors, err)
			continue
		}
		a.loadedExtensions = append(a.loadedExtensions, e)
	}
}

// attachExtension は拡張1つ分を確かめてから足します。
func (a *App) attachExtension(e *extension.Extension) error {
	if err := a.checkExtension(e, ""); err != nil {
		return err
	}
	groups := map[string]*cobra.Command{}
	for _, g := range sortedGroups(e.Spec) {
		key := strings.Join(g.Command, " ")
		if existing := findCommand(a.root, g.Command); existing != nil {
			groups[key] = existing
			continue
		}
		c := &cobra.Command{
			Use:     g.Command[len(g.Command)-1],
			Short:   groupShort(g),
			Aliases: aliasesFor(g.Command[len(g.Command)-1]),
		}
		a.attachUnder(groups, g.Command, c)
		groups[key] = c
		a.extensionOf[c] = e
	}
	for _, sc := range e.Spec.Commands {
		c := a.newExtensionOperationCmd(e, sc)
		a.attachUnder(groups, sc.Command, c)
		a.extensionOf[c] = e
	}
	return nil
}

// attachUnder は親（拡張が足したグループか既存のコマンド）の下に付けます。
func (a *App) attachUnder(groups map[string]*cobra.Command, p []string, c *cobra.Command) {
	parentPath := p[:len(p)-1]
	if len(parentPath) == 0 {
		a.root.AddCommand(c)
		return
	}
	if parent, ok := groups[strings.Join(parentPath, " ")]; ok {
		parent.AddCommand(c)
		return
	}
	findCommand(a.root, parentPath).AddCommand(c)
}

// checkExtension は拡張を足せるかを、ツリーを変えずに確かめます。
// replacing が空でなければ、その名前の拡張が足したものとはぶつかっても
// 構いません（入れ直しのとき）。
func (a *App) checkExtension(e *extension.Extension, replacing string) error {
	ours := func(c *cobra.Command) bool {
		owner, ok := a.extensionOf[c]
		return ok && replacing != "" && owner.Name == replacing
	}
	declared := map[string]bool{}
	for _, g := range e.Spec.Groups {
		declared[strings.Join(g.Command, " ")] = true
	}
	for _, g := range e.Spec.Groups {
		if c := findCommand(a.root, g.Command); c != nil && c.Runnable() && !ours(c) {
			return &extensionConflict{ext: e.Name, path: strings.Join(g.Command, " "), with: c.CommandPath()}
		}
	}
	seen := map[string]bool{}
	for _, sc := range e.Spec.Commands {
		key := strings.Join(sc.Command, " ")
		if seen[key] || declared[key] {
			return fmt.Errorf(i18n.T("拡張 %s に %s が2回あります", "extension %s defines %s twice"), e.Name, key)
		}
		seen[key] = true

		parentPath := sc.Command[:len(sc.Command)-1]
		if len(parentPath) > 0 && !declared[strings.Join(parentPath, " ")] {
			if findCommand(a.root, parentPath) == nil {
				return fmt.Errorf(i18n.T(
					"拡張 %s の %s の親 %s がありません（groups に書いてください）",
					"extension %s: the parent of %s, %s, does not exist (declare it in groups)"),
					e.Name, key, strings.Join(parentPath, " "))
			}
		}
		parent := a.root
		if len(parentPath) > 0 {
			parent = findCommand(a.root, parentPath)
		}
		if parent == nil {
			continue // 拡張自身が足すグループの下。中身はまだ空。
		}
		name := sc.Command[len(sc.Command)-1]
		for _, child := range parent.Commands() {
			if (child.Name() == name || child.HasAlias(name)) && !ours(child) {
				return &extensionConflict{ext: e.Name, path: key, with: child.CommandPath()}
			}
		}
	}
	return nil
}

// newExtensionOperationCmd は拡張のコマンド1つです。中身は生成コマンドと
// 同じで、出どころの説明と 404 のときの助言だけを足します。
func (a *App) newExtensionOperationCmd(e *extension.Extension, sc *spec.Command) *cobra.Command {
	c := a.newOperationCmd(sc)
	c.Long += "\n" + extensionNote(e)
	run := c.RunE
	c.RunE = func(cmd *cobra.Command, args []string) error {
		err := run(cmd, args)
		if apiErr, ok := api.AsError(err); ok && apiErr.UnknownEndpoint() {
			return fmt.Errorf("%w\n%s", err, missingPluginHint(e))
		}
		return err
	}
	return c
}

// extensionNote は --help に足す「この操作はどこから来たか」です。
func extensionNote(e *extension.Extension) string {
	note := fmt.Sprintf(i18n.T(
		"この操作は拡張 %s（%s）が足したもので、MT 本体の Data API にはありません。\n",
		"This operation comes from the extension %s (%s); it is not part of MT's own Data API.\n"),
		e.Name, e.Version)
	if p := e.RequiredPlugin(); p != nil {
		note += fmt.Sprintf(i18n.T(
			"MT 側に %s プラグインが要ります。\n",
			"MT needs the %s plugin installed.\n"), p.Label())
	}
	return note
}

// missingPluginHint は拡張のコマンドが「Unknown endpoint」の 404 を受けたときの
// 助言です。拡張のエンドポイントは MT 本体に無いので、それは「MT 側に
// プラグインが入っていない」ことを意味します。プラグインが返す 404
// （File not found など）には付けません（入っているのに疑わせることになる）。
func missingPluginHint(e *extension.Extension) string {
	if p := e.RequiredPlugin(); p != nil {
		return fmt.Sprintf(i18n.T(
			"このコマンドは拡張 %s のものです。MT 側に %s プラグインが入っていない可能性があります（mtb extensions check %s で確かめられます）。",
			"This command comes from the extension %s. MT may not have the %s plugin installed (check with: mtb extensions check %s)."),
			e.Name, p.Label(), e.Name)
	}
	return fmt.Sprintf(i18n.T(
		"このコマンドは拡張 %s のものです。MT 側にエンドポイントがあるか mtb extensions check %s で確かめてください。",
		"This command comes from the extension %s. Check that MT has the endpoint with: mtb extensions check %s"),
		e.Name, e.Name)
}

// sortedGroups は親が子より先に来る順に並べます。
func sortedGroups(s *spec.Spec) []*spec.Group {
	out := append([]*spec.Group(nil), s.Groups...)
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].Command) < len(out[j].Command) })
	return out
}

// findCommand は名前の並びでツリーを辿ります（別名は辿りません）。
func findCommand(root *cobra.Command, p []string) *cobra.Command {
	cur := root
	for _, name := range p {
		var next *cobra.Command
		for _, c := range cur.Commands() {
			if c.Name() == name {
				next = c
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

// extensionCommandPaths は拡張が足した末端コマンドのパスです。
func (a *App) extensionCommandPaths(e *extension.Extension) []string {
	var out []string
	for c, owner := range a.extensionOf {
		if owner == e && c.Runnable() {
			out = append(out, c.CommandPath())
		}
	}
	sort.Strings(out)
	return out
}

func (a *App) findExtension(name string) (*extension.Extension, error) {
	for _, e := range a.env.Extensions {
		if e.Name == name {
			return e, nil
		}
	}
	return nil, &usageError{fmt.Errorf(i18n.T(
		"拡張 %s はありません（mtb extensions list で一覧できます）",
		"there is no extension %s (see: mtb extensions list)"), name)}
}

func (a *App) isLoaded(e *extension.Extension) bool {
	for _, l := range a.loadedExtensions {
		if l == e {
			return true
		}
	}
	return false
}

// newExtensionsCmd は拡張の管理コマンドです。
func (a *App) newExtensionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extensions",
		Aliases: []string{"extension", "ext"},
		Short:   i18n.T("mtb の拡張（MT プラグインが足したエンドポイントのコマンド）の管理", "Manage mtb extensions (commands for endpoints added by MT plugins)"),
		Long: i18n.T(
			"拡張は、MT のプラグインが Data API に足したエンドポイントに、\n"+
				"生成コマンドと同じ形のコマンドを与えます（--set・--fields・--dry-run が効きます）。\n\n"+
				"拡張は定義（JSON）と説明（Markdown）だけでできていて、コードは走りません。\n"+
				"入れる先は設定ディレクトリの extensions/ です。開発中のものは\n"+
				"MTB_EXTENSIONS_PATH にディレクトリを並べれば、入れずに読み込めます。\n\n"+
				"MT のプラグインそのものを操作するのは mtb plugins です（別物）。",
			"An extension gives the endpoints an MT plugin adds to the Data API commands of\n"+
				"the same shape as the generated ones (--set, --fields and --dry-run work).\n\n"+
				"An extension is only a definition (JSON) and a chapter (Markdown); no code runs.\n"+
				"Extensions are installed under extensions/ in the config directory. To load one\n"+
				"you are developing without installing it, list its directory in MTB_EXTENSIONS_PATH.\n\n"+
				"To operate MT's own plugins, use mtb plugins (a different thing)."),
	}
	cmd.AddCommand(
		a.newExtensionsListCmd(),
		a.newExtensionsShowCmd(),
		a.newExtensionsValidateCmd(),
		a.newExtensionsInstallCmd(),
		a.newExtensionsUninstallCmd(),
		a.newExtensionsCheckCmd(),
	)
	return cmd
}

func (a *App) newExtensionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   i18n.T("拡張の一覧", "List the extensions"),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(a.env.Extensions) == 0 {
				fmt.Fprintf(a.env.Out, i18n.T("拡張はありません（%s）\n", "No extensions (%s)\n"), a.env.ExtensionsDir)
				fmt.Fprintln(a.env.Out, i18n.T(
					"mtb extensions install <ディレクトリ> で入れられます",
					"Install one with: mtb extensions install <directory>"))
				return nil
			}
			w := tabwriter.NewWriter(a.env.Out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tVERSION\tSOURCE\tCOMMANDS\tSTATUS\tDESCRIPTION")
			for _, e := range a.env.Extensions {
				status := i18n.T("読み込み済み", "loaded")
				if !a.isLoaded(e) {
					status = i18n.T("読み込めない", "not loaded")
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", e.Name, e.Version, e.Source, len(e.Spec.Commands), status, e.Describe())
			}
			return w.Flush()
		},
	}
}

func (a *App) newExtensionsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: i18n.T("拡張の中身（コマンド・要る MT プラグイン・Claude Code のプラグイン）", "Show an extension (its commands, the MT plugin it needs, its Claude Code plugin)"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := a.findExtension(args[0])
			if err != nil {
				return err
			}
			a.writeExtension(e)
			return nil
		},
	}
}

// writeExtension は1つ分の詳細です。install の後にも同じものを見せます。
func (a *App) writeExtension(e *extension.Extension) {
	out := a.env.Out
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", i18n.T("名前", "Name"), e.Name)
	fmt.Fprintf(w, "%s\t%s\n", i18n.T("版", "Version"), e.Version)
	fmt.Fprintf(w, "%s\t%s\n", i18n.T("説明", "Description"), e.Describe())
	fmt.Fprintf(w, "%s\t%s (%s)\n", i18n.T("場所", "Location"), e.Dir, e.Source)
	if p := e.RequiredPlugin(); p != nil {
		label := p.Label()
		if p.URL != "" {
			label += "  " + p.URL
		}
		fmt.Fprintf(w, "%s\t%s\n", i18n.T("要る MT プラグイン", "MT plugin needed"), label)
	}
	w.Flush()

	if a.isLoaded(e) {
		fmt.Fprintln(out, "\n"+i18n.T("コマンド:", "Commands:"))
		for _, p := range a.extensionCommandPaths(e) {
			fmt.Fprintln(out, "  "+p)
		}
	}
	if dir := e.ClaudePluginDir(); dir != "" {
		fmt.Fprintln(out, "\n"+i18n.T("Claude Code のプラグイン:", "Claude Code plugin:"))
		fmt.Fprintln(out, "  "+dir)
		fmt.Fprintf(out, i18n.T(
			"  試すなら: claude --plugin-dir %s\n",
			"  To try it: claude --plugin-dir %s\n"), dir)
	}
	if e.LLMDoc != "" {
		fmt.Fprintln(out, "\n"+i18n.T(
			"使い方の章は mtb llm に入っています。",
			"Its chapter is included in mtb llm."))
	}
}

func (a *App) newExtensionsValidateCmd() *cobra.Command {
	var strict bool
	cmd := &cobra.Command{
		Use:   "validate <directory>",
		Short: i18n.T("拡張のディレクトリを検査する（入れはしない）", "Check an extension directory without installing it"),
		Long: i18n.T(
			"マニフェストと定義を読み、既存のコマンドとぶつからないかまで確かめます。\n"+
				"同梱の Claude Code のプラグインも見ます（版が拡張と一致するか、スキルの書式）。\n"+
				"拡張を作るときと、拡張の CI で使います。\n\n"+
				"--strict を付けると、Claude Code のプラグインの問題も失敗にします（既定は警告）。",
			"Reads the manifest and the definition and checks that nothing collides with\n"+
				"the existing commands. Also checks the bundled Claude Code plugin (its version\n"+
				"matches the extension's, and the skills are well formed). Use it while writing\n"+
				"an extension and in its CI.\n\n"+
				"With --strict, problems in the Claude Code plugin fail too (they warn by default)."),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := a.loadExtensionDir(args[0])
			if err != nil {
				return err
			}
			if err := a.checkExtension(e, e.Name); err != nil {
				return err
			}
			problems := a.lintExtension(e)
			for _, p := range problems {
				fmt.Fprintln(a.env.Err, i18n.T("警告:", "warning:"), p)
			}
			if strict && len(problems) > 0 {
				return fmt.Errorf(i18n.T(
					"%s: %d 件の問題があります（--strict）",
					"%s: %d problems (--strict)"), e.Name, len(problems))
			}
			fmt.Fprintf(a.env.Out, i18n.T(
				"%s %s: 問題ありません（コマンド %d）\n",
				"%s %s: OK (%d commands)\n"), e.Name, e.Version, len(e.Spec.Commands))
			return nil
		},
	}
	cmd.Flags().BoolVar(&strict, "strict", false, i18n.T(
		"Claude Code のプラグインの問題も失敗にする（拡張の CI 向け）",
		"fail on problems in the Claude Code plugin too (for an extension's CI)"))
	return cmd
}

// lintExtension は読み込みには差し支えない問題を集めます。
func (a *App) lintExtension(e *extension.Extension) []string {
	var problems []string
	for _, sc := range e.Spec.Commands {
		if strings.TrimSpace(sc.Short) == "" {
			problems = append(problems, fmt.Sprintf(i18n.T(
				"%s に short（一覧に出る1行）がありません",
				"%s has no short (the line shown in listings)"), strings.Join(sc.Command, " ")))
		}
	}
	if dir := e.ClaudePluginDir(); dir != "" && a.env.CheckClaudePlugin != nil {
		for _, p := range a.env.CheckClaudePlugin(dir, e.Version) {
			problems = append(problems, "Claude Code plugin: "+p)
		}
	}
	return problems
}

func (a *App) newExtensionsInstallCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "install <directory>",
		Short: i18n.T("拡張を入れる", "Install an extension"),
		Long: i18n.T(
			"拡張のディレクトリを設定ディレクトリの extensions/<名前>/ に写します。\n"+
				"Git のリポジトリなら、clone してからそのディレクトリを渡してください。\n\n"+
				"MT 側のプラグインは別に入れる必要があります。入っているかは\n"+
				"mtb extensions check <名前> で確かめられます。",
			"Copies the extension directory to extensions/<name>/ in the config directory.\n"+
				"For a Git repository, clone it first and pass that directory.\n\n"+
				"The MT plugin on the server has to be installed separately. Check whether it\n"+
				"is there with: mtb extensions check <name>"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			e, err := a.loadExtensionDir(src)
			if err != nil {
				return err
			}
			dest := filepath.Join(a.env.ExtensionsDir, e.Name)
			for _, existing := range a.env.Extensions {
				if existing.Name == e.Name && !force {
					return &usageError{fmt.Errorf(i18n.T(
						"拡張 %s はすでにあります（%s）。入れ直すなら --force を付けてください",
						"extension %s is already there (%s); add --force to reinstall it"), e.Name, existing.Dir)}
				}
			}
			if err := a.checkExtension(e, e.Name); err != nil {
				return err
			}
			files, err := a.env.ListFiles(src)
			if err != nil {
				return err
			}
			if a.g.dryRun {
				fmt.Fprintf(a.env.Out, i18n.T("%s へ %d 個のファイルを写します（--dry-run）\n", "would copy to %s: %d files (--dry-run)\n"), dest, len(files))
				return nil
			}
			if err := a.env.RemoveAll(dest); err != nil {
				return err
			}
			copied := 0
			for _, rel := range files {
				if skipWhenCopying(rel) {
					continue
				}
				data, err := a.env.ReadFile(filepath.Join(src, rel))
				if err != nil {
					return err
				}
				if err := a.env.WriteFile(filepath.Join(dest, rel), data); err != nil {
					return err
				}
				copied++
			}
			fmt.Fprintf(a.env.Out, i18n.T("%s %s を %s に入れました（%d ファイル）\n\n", "Installed %s %s to %s (%d files)\n\n"), e.Name, e.Version, dest, copied)

			installed := *e
			installed.Dir = dest
			installed.Source = extension.SourceInstalled
			// 入れたばかりの拡張のコマンドを見せるため、このプロセスの
			// ツリーにも足します（入れ直しなら古い方と差し替わります）。
			a.forgetExtension(e.Name)
			if err := a.attachExtension(&installed); err == nil {
				a.loadedExtensions = append(a.loadedExtensions, &installed)
				a.env.Extensions = append(a.env.Extensions, &installed)
			}
			a.writeExtension(&installed)
			if p := e.RequiredPlugin(); p != nil {
				fmt.Fprintf(a.env.Out, "\n"+i18n.T(
					"MT 側に %s プラグインが要ります。入っているかは mtb extensions check %s で確かめられます。\n",
					"MT needs the %s plugin. Check whether it is installed with: mtb extensions check %s\n"),
					p.Label(), e.Name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, i18n.T("同じ名前の拡張があっても入れ直す", "reinstall even if an extension of the same name exists"))
	return cmd
}

// forgetExtension はこのプロセスのツリーから拡張を外します（入れ直しのため）。
func (a *App) forgetExtension(name string) {
	for c, owner := range a.extensionOf {
		if owner.Name == name {
			if c.Parent() != nil {
				c.Parent().RemoveCommand(c)
			}
			delete(a.extensionOf, c)
		}
	}
	keep := a.env.Extensions[:0]
	for _, e := range a.env.Extensions {
		if e.Name != name {
			keep = append(keep, e)
		}
	}
	a.env.Extensions = keep
}

// skipWhenCopying は拡張を入れるときに写さないものです。リポジトリの
// 管理情報は要らず、.claude-plugin は Claude Code のプラグインに要るので残す。
func skipWhenCopying(rel string) bool {
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".git" || seg == ".DS_Store" || seg == "node_modules" {
			return true
		}
	}
	return false
}

func (a *App) newExtensionsUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall <name>",
		Aliases: []string{"remove", "rm"},
		Short:   i18n.T("拡張を外す", "Uninstall an extension"),
		Long: i18n.T(
			"extensions/<名前>/ を消します。MT 側のプラグインには触りません。\n"+
				"MTB_EXTENSIONS_PATH で読み込んでいるものは消しません（そちらの設定から外してください）。",
			"Deletes extensions/<name>/. The MT plugin on the server is left alone.\n"+
				"Extensions loaded through MTB_EXTENSIONS_PATH are not deleted (remove them from that variable)."),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := a.findExtension(args[0])
			if err != nil {
				return err
			}
			if e.Source != extension.SourceInstalled {
				return &usageError{fmt.Errorf(i18n.T(
					"拡張 %s は MTB_EXTENSIONS_PATH から読み込んでいます（%s）。消すのではなく、変数から外してください",
					"extension %s is loaded through MTB_EXTENSIONS_PATH (%s); remove it from the variable instead"), e.Name, e.Dir)}
			}
			if a.g.dryRun {
				fmt.Fprintf(a.env.Out, i18n.T("%s を消します（--dry-run）\n", "would delete %s (--dry-run)\n"), e.Dir)
				return nil
			}
			if err := a.env.RemoveAll(e.Dir); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Out, i18n.T("拡張 %s を外しました（%s）\n", "Uninstalled extension %s (%s)\n"), e.Name, e.Dir)
			return nil
		},
	}
}

func (a *App) newExtensionsCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [name]",
		Short: i18n.T("拡張が要るエンドポイントが MT にあるか確かめる", "Check that MT has the endpoints an extension needs"),
		Long: i18n.T(
			"MT の /endpoints を読み、拡張のマニフェストにある requires.endpoints が\n"+
				"揃っているかを見ます。名前を省くと全部の拡張を見ます。\n"+
				"足りなければ終了コード 1 で終わります（MT 側にプラグインが入っていない）。",
			"Reads MT's /endpoints and checks that the requires.endpoints listed in each\n"+
				"extension's manifest are all there. Leave off the name to check every extension.\n"+
				"Exits with code 1 if anything is missing (the MT plugin is not installed)."),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targets := a.env.Extensions
			if len(args) == 1 {
				e, err := a.findExtension(args[0])
				if err != nil {
					return err
				}
				targets = []*extension.Extension{e}
			}
			if len(targets) == 0 {
				fmt.Fprintln(a.env.Out, i18n.T("拡張はありません", "No extensions"))
				return nil
			}
			have, err := a.serverEndpointIDs(cmd)
			if err != nil {
				return err
			}
			missing := 0
			for _, e := range targets {
				var want []string
				if e.Requires != nil {
					want = e.Requires.Endpoints
				}
				if len(want) == 0 {
					fmt.Fprintf(a.env.Out, i18n.T(
						"%s: マニフェストに requires.endpoints が無いので確かめられません\n",
						"%s: cannot check; the manifest has no requires.endpoints\n"), e.Name)
					continue
				}
				var absent []string
				for _, id := range want {
					if !have[id] {
						absent = append(absent, id)
					}
				}
				if len(absent) == 0 {
					fmt.Fprintf(a.env.Out, i18n.T("%s: OK（%s）\n", "%s: OK (%s)\n"), e.Name, strings.Join(want, ", "))
					continue
				}
				missing++
				fmt.Fprintf(a.env.Out, i18n.T("%s: 足りません: %s\n", "%s: missing: %s\n"), e.Name, strings.Join(absent, ", "))
				if p := e.RequiredPlugin(); p != nil {
					fmt.Fprintf(a.env.Out, i18n.T(
						"  MT 側に %s プラグインを入れてください。\n",
						"  Install the %s plugin on MT.\n"), p.Label())
				}
			}
			if missing > 0 {
				return errors.New(i18n.T(
					"MT 側に足りないエンドポイントがあります",
					"MT is missing endpoints that an extension needs"))
			}
			return nil
		},
	}
}

// serverEndpointIDs は MT にあるエンドポイントの ID の集合です。
func (a *App) serverEndpointIDs(cmd *cobra.Command) (map[string]bool, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(cmd.Context(), api.Request{Method: http.MethodGet, Path: "/endpoints", Query: map[string][]string{"limit": {"10000"}}})
	if err != nil {
		return nil, a.describeError(err, resp)
	}
	var body struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return nil, err
	}
	have := map[string]bool{}
	for _, it := range body.Items {
		have[it.ID] = true
	}
	return have, nil
}

// loadExtensionDir は手元のディレクトリから拡張を読みます。
func (a *App) loadExtensionDir(dir string) (*extension.Extension, error) {
	e, err := extension.Load(a.env.ReadFile, dir)
	if err != nil {
		return nil, err
	}
	e.Source = extension.SourcePath
	return e, nil
}
