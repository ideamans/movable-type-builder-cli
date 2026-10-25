package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// テンプレートをローカルのディレクトリと往復させます。
//
// これは**エンドポイントではありません**。Data API に「まとめて取る／
// まとめて戻す」は無いので、mtb が list と update を回します。ページを
// 繰る --all や、ディレクトリを送る --dir と同じ立場のものです。
//
// 置き場所は MT 自身のテーマ構成に寄せてあります（本文は1テンプレート
// 1ファイル、種別や出力ファイル名はマニフェストが持つ）。テーマの
// templates/*.mtml と theme.yaml の関係と同じで、そちらへ移りたく
// なったときに考え方を変えずに済みます。
const (
	manifestFile  = "mtb-templates.json"
	templateExt   = ".mtml"
	syncDirFlag   = "dir"
	syncForceFlag = "force"
)

// manifest はローカルのディレクトリと MT のテンプレートの対応表です。
//
// **ファイル名ではなく id で対応を取ります。** MT 側で名前を変えても、
// 手元でファイル名を変えても、対応が切れないようにするためです。
type manifest struct {
	SiteID    string          `json:"siteId"`
	Templates []*manifestItem `json:"templates"`
}

type manifestItem struct {
	// ID が空なら push で新規作成し、採番された ID を書き戻します。
	ID   string `json:"id,omitempty"`
	File string `json:"file"`
	Name string `json:"name"`
	Type string `json:"type"`
	// 以下は種別によって空です。空のものは送りません。
	OutputFile string `json:"outputFile,omitempty"`
	BuildType  string `json:"buildType,omitempty"`
	LinkToFile string `json:"linkToFile,omitempty"`
	// ArchiveMaps はアーカイブマッピングです。**これが URL の設計そのもの**
	// なので、テンプレートだけ往復させても記事の出力先は戻りません。
	// 持てるのは archiveTypes に挙げた種別だけです。
	ArchiveMaps []*manifestMap `json:"archiveMaps,omitempty"`
}

// manifestMap はアーカイブマッピング1件です。
type manifestMap struct {
	// ID が空なら push で新規作成し、採番された ID を書き戻します。
	ID           string `json:"id,omitempty"`
	ArchiveType  string `json:"archiveType"`
	FileTemplate string `json:"fileTemplate,omitempty"`
	BuildType    string `json:"buildType,omitempty"`
	IsPreferred  bool   `json:"isPreferred,omitempty"`
}

// archiveTypes はアーカイブマッピングを持てるテンプレートの種別です。
// 全件に問い合わせると 30 件超のテンプレートで毎回その数だけ往復するので、
// 持ちうるものだけに絞ります。
var archiveTypes = map[string]bool{
	"individual": true,
	"page":       true,
	"archive":    true,
	"ct":         true,
	"ct_archive": true,
}

// apiTemplatemap は API から受け取るアーカイブマッピングです。
type apiTemplatemap struct {
	ID           json.RawMessage `json:"id"`
	ArchiveType  string          `json:"archiveType"`
	FileTemplate string          `json:"fileTemplate"`
	BuildType    string          `json:"buildType"`
	IsPreferred  bool            `json:"isPreferred"`
}

func (m apiTemplatemap) id() string { return strings.Trim(string(m.ID), `"`) }

// toManifest はマニフェストに書く形にします。
func (m apiTemplatemap) toManifest() *manifestMap {
	return &manifestMap{
		ID: m.id(), ArchiveType: m.ArchiveType, FileTemplate: m.FileTemplate,
		BuildType: m.BuildType, IsPreferred: m.IsPreferred,
	}
}

// sameMap は送る必要が無いかを返します。
func sameMap(want *manifestMap, got apiTemplatemap) bool {
	return want.ArchiveType == got.ArchiveType &&
		want.FileTemplate == got.FileTemplate &&
		want.BuildType == got.BuildType &&
		want.IsPreferred == got.IsPreferred
}

// apiTemplate は API から受け取るテンプレートです。
type apiTemplate struct {
	ID         json.RawMessage `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Text       string          `json:"text"`
	OutputFile string          `json:"outputFile"`
	BuildType  string          `json:"buildType"`
	LinkToFile string          `json:"linkToFile"`
}

func (t apiTemplate) id() string { return strings.Trim(string(t.ID), `"`) }

// addTemplateSyncCommands は生成されたツリーの sites templates に
// pull / push を足します。利用者はテンプレートの操作をここで探すので、
// 別の場所に置くより見つけやすい。名前が MT 側の新しいエンドポイントと
// ぶつかれば TestNoDuplicateNamesAmongSiblings が落ちます。
func (a *App) addTemplateSyncCommands() {
	parent, _, err := a.root.Find([]string{"sites", "templates"})
	if err != nil || parent == nil {
		return
	}
	parent.AddCommand(a.newTemplatePullCmd(), a.newTemplatePushCmd())
}

func (a *App) newTemplatePullCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   i18n.T("pull <サイトID>", "pull <site-id>"),
		Short: i18n.T("テンプレートをディレクトリに書き出す", "Write the templates out to a directory"),
		Long: i18n.T(
			"サイトのテンプレートを1つ1ファイルで書き出し、種別や出力ファイル名を\n"+
				manifestFile+" に記録します。Git で管理して push で戻せます。\n\n"+
				"**手元のファイルは上書きされます。** 書き換えたまま pull すると\n"+
				"消えるので、先に commit しておくか --dry-run で確かめてください。\n\n"+
				"アーカイブマッピングは対象外です（sites templates templatemaps で扱います）。",
			"Write each template of a site to its own file, recording the type and\n"+
				"output file name in "+manifestFile+". Keep it in Git and send it back with push.\n\n"+
				"**Local files are overwritten.** If you have edits you have not committed,\n"+
				"commit them first or check with --dry-run.\n\n"+
				"Archive mappings are out of scope (use sites templates templatemaps)."),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			siteID, err := a.syncSiteID(args)
			if err != nil {
				return err
			}
			return a.pullTemplates(context.Background(), siteID, dir)
		},
	}
	cmd.Flags().StringVar(&dir, syncDirFlag, ".", i18n.T("書き出す先のディレクトリ", "directory to write into"))
	return cmd
}

func (a *App) newTemplatePushCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   i18n.T("push <サイトID>", "push <site-id>"),
		Short: i18n.T("ディレクトリのテンプレートを MT に戻す", "Send the templates in a directory back to MT"),
		Long: i18n.T(
			manifestFile+" に載っているものを MT に戻します。ID があれば更新し、\n"+
				"無ければ新規に作って採番された ID を書き戻します。\n\n"+
				"**MT にあって手元に無いテンプレートには触れません。** 消したいときは\n"+
				"sites templates delete を使ってください（取り違えると戻せないので、\n"+
				"まとめて消すことはしません）。",
			"Send what is listed in "+manifestFile+" back to MT. Entries with an ID are\n"+
				"updated; entries without one are created and the new ID is written back.\n\n"+
				"**Templates that exist on MT but not here are left alone.** To remove one,\n"+
				"use sites templates delete (deleting in bulk is too easy to get wrong)."),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			siteID, err := a.syncSiteID(args)
			if err != nil {
				return err
			}
			return a.pushTemplates(context.Background(), siteID, dir)
		},
	}
	cmd.Flags().StringVar(&dir, syncDirFlag, ".", i18n.T("読み込むディレクトリ", "directory to read from"))
	return cmd
}

// syncSiteID は位置引数かプロファイルの既定サイトから対象を決めます。
func (a *App) syncSiteID(args []string) (string, error) {
	if len(args) == 1 && args[0] != "" {
		return args[0], nil
	}
	r, err := a.resolve()
	if err != nil {
		return "", err
	}
	if r.SiteID == "" {
		return "", usagef("%s", i18n.T(
			"サイトIDが要ります（mtb profile set --site-id <ID> で既定にもできます）",
			"a site ID is required (or set a default with mtb profile set --site-id <ID>)"))
	}
	return r.SiteID, nil
}

func (a *App) pullTemplates(ctx context.Context, siteID, dir string) error {
	client, err := a.client()
	if err != nil {
		return err
	}
	items, err := a.fetchTemplates(ctx, client, siteID)
	if err != nil {
		return err
	}

	m := &manifest{SiteID: siteID}
	used := map[string]bool{}
	for _, t := range items {
		file := uniqueFileName(t.Name, t.id(), used)
		item := &manifestItem{
			ID: t.id(), File: file, Name: t.Name, Type: t.Type,
			OutputFile: t.OutputFile, BuildType: t.BuildType, LinkToFile: t.LinkToFile,
		}
		if archiveTypes[t.Type] {
			maps, err := a.fetchTemplatemaps(ctx, client, siteID, t.id())
			if err != nil {
				return err
			}
			for _, tm := range maps {
				item.ArchiveMaps = append(item.ArchiveMaps, tm.toManifest())
			}
		}
		m.Templates = append(m.Templates, item)
		if a.g.dryRun {
			fmt.Fprintf(a.env.Out, "%s\n", path.Join(dir, file))
			continue
		}
		if err := a.env.WriteFile(path.Join(dir, file), []byte(t.Text)); err != nil {
			return err
		}
	}
	if a.g.dryRun {
		fmt.Fprintf(a.env.Out, "%s\n", path.Join(dir, manifestFile))
		fmt.Fprintf(a.env.Out, i18n.T("（%d 件）\n", "(%d templates)\n"), len(items))
		return nil
	}
	if err := a.writeManifest(dir, m); err != nil {
		return err
	}
	fmt.Fprintf(a.env.Err, i18n.T("%d 件を %s に書き出しました\n", "Wrote %d templates to %s\n"), len(items), dir)
	return nil
}

func (a *App) pushTemplates(ctx context.Context, siteID, dir string) error {
	m, err := a.readManifest(dir)
	if err != nil {
		return err
	}
	if m.SiteID != "" && m.SiteID != siteID {
		return usagef(i18n.T(
			"%s はサイト %s のものです。サイト %s に戻すなら %s の siteId を直してください",
			"%s belongs to site %s, not site %s; change siteId in %s if that is intended"),
			manifestFile, m.SiteID, siteID, manifestFile)
	}
	if len(m.Templates) == 0 {
		return usagef(i18n.T("%s にテンプレートがありません", "%s lists no templates"), manifestFile)
	}

	client, err := a.client()
	if err != nil {
		return err
	}

	// 変わっていないものは送りません。1つ直しただけで全件にリビジョンが
	// 積まれると、MT 側の更新履歴が「いつ何を変えたか」の役に立たなくなります。
	current, err := a.fetchTemplates(ctx, client, siteID)
	if err != nil {
		return err
	}
	live := map[string]apiTemplate{}
	for _, t := range current {
		live[t.id()] = t
	}

	var changed []*manifestItem
	texts := map[*manifestItem]string{}
	for _, item := range m.Templates {
		text, err := a.env.ReadFile(path.Join(dir, item.File))
		if err != nil {
			return fmt.Errorf(i18n.T("%s: %w", "%s: %w"), item.File, err)
		}
		texts[item] = string(text)
		if t, ok := live[item.ID]; ok && item.ID != "" && sameTemplate(item, t, string(text)) {
			continue
		}
		changed = append(changed, item)
	}
	// **本文が変わっていなくても、アーカイブマッピングだけ変わっていることが
	// あります。** マッピングは URL の設計そのものなので、本文の差分が無いことを
	// 理由に飛ばすと、出力先が戻らないまま「変わっていません」と言うことになります。
	liveMaps, pendingMaps, err := a.planTemplatemaps(ctx, client, siteID, m)
	if err != nil {
		return err
	}

	if len(changed) == 0 && pendingMaps == 0 {
		fmt.Fprintf(a.env.Err, i18n.T(
			"変わっているものはありません（%d 件を確認）\n",
			"Nothing has changed (%d templates checked)\n"), len(m.Templates))
		return a.reportUntracked(current, m)
	}

	if !a.g.dryRun {
		ok, err := a.confirm(fmt.Sprintf(i18n.T(
			"%sに対して、テンプレート %d 件とアーカイブマッピング %d 件を書き換えます。よろしいですか？",
			"On %s, %d templates and %d archive mappings will be rewritten. Continue?"),
			a.target(), len(changed), pendingMaps))
		if err != nil {
			return err
		}
		if !ok {
			return errAborted()
		}
	}

	created := 0
	for i, item := range changed {
		body := map[string]any{"name": item.Name, "text": texts[item]}
		// 種別は作るときだけ送ります。既存のテンプレートの種別を
		// 変えるのは、システムテンプレートでは通りません。
		if item.ID == "" {
			body["type"] = item.Type
		}
		for k, v := range map[string]string{
			"outputFile": item.OutputFile, "buildType": item.BuildType, "linkToFile": item.LinkToFile,
		} {
			if v != "" {
				body[k] = v
			}
		}
		payload, err := marshalNoEscape(body)
		if err != nil {
			return err
		}

		req := api.Request{
			Method: http.MethodPut,
			Path:   "/sites/" + siteID + "/templates/" + item.ID,
			Form:   url.Values{"template": []string{payload}},
		}
		if item.ID == "" {
			req.Method = http.MethodPost
			req.Path = "/sites/" + siteID + "/templates"
		}
		if a.g.dryRun {
			fmt.Fprintf(a.env.Out, "%s %s  %s\n", req.Method, req.Path, item.File)
			continue
		}
		if a.g.verbose {
			fmt.Fprintf(a.env.Err, "%d/%d %s\n", i+1, len(changed), item.File)
		}
		resp, err := client.Do(ctx, req)
		if err != nil {
			fmt.Fprintf(a.env.Err, i18n.T(
				"%d/%d %s で止まりました\n", "stopped at %d/%d %s\n"), i+1, len(changed), item.File)
			return a.describeError(err, resp)
		}
		if item.ID == "" {
			var t apiTemplate
			if err := json.Unmarshal(resp.Body, &t); err == nil {
				item.ID = t.id()
			}
			created++
		}
	}

	// アーカイブマッピングは本文とは別のエンドポイントなので、あとから合わせます。
	// **本文が変わっていなくてもマッピングだけ変わっていることがある**ので、
	// 対象は changed ではなくマニフェスト全体です。
	mapped := 0
	for _, item := range m.Templates {
		if len(item.ArchiveMaps) == 0 || item.ID == "" {
			continue
		}
		n, err := a.pushTemplatemaps(ctx, client, siteID, item, liveMaps[item.ID])
		if err != nil {
			return err
		}
		mapped += n
	}

	if a.g.dryRun {
		return nil
	}
	if mapped > 0 {
		if err := a.writeManifest(dir, m); err != nil {
			return err
		}
		fmt.Fprintf(a.env.Err, i18n.T(
			"アーカイブマッピングを %d 件そろえました\n",
			"Reconciled %d archive mappings\n"), mapped)
	}
	if created > 0 {
		if err := a.writeManifest(dir, m); err != nil {
			return err
		}
	}
	if len(changed) > 0 {
		fmt.Fprintf(a.env.Err, i18n.T(
			"%d 件を戻しました（うち新規 %d 件、変わっていない %d 件は送っていません）\n",
			"Sent %d templates (%d of them new; %d unchanged were not sent)\n"),
			len(changed), created, len(m.Templates)-len(changed))
	}
	return a.reportUntracked(current, m)
}

// fetchTemplatemaps はテンプレート1件のアーカイブマッピングを取ります。
func (a *App) fetchTemplatemaps(ctx context.Context, client *api.Client, siteID, templateID string) ([]apiTemplatemap, error) {
	resp, err := client.Do(ctx, api.Request{
		Method: http.MethodGet,
		Path:   "/sites/" + siteID + "/templates/" + templateID + "/templatemaps",
		Query:  url.Values{"limit": []string{"100"}},
	})
	if err != nil {
		return nil, err
	}
	var page struct {
		Items []apiTemplatemap `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

// planTemplatemaps は、いま MT にあるマッピングと、送る必要のある件数を返します。
// 確認の前に数えたいので、送信そのものとは分けてあります。
func (a *App) planTemplatemaps(ctx context.Context, client *api.Client, siteID string, m *manifest) (map[string]map[string]apiTemplatemap, int, error) {
	live := map[string]map[string]apiTemplatemap{}
	pending := 0
	for _, item := range m.Templates {
		if len(item.ArchiveMaps) == 0 || item.ID == "" {
			// ID がまだ無いもの（新規作成）のマッピングは全部が新規です。
			pending += len(item.ArchiveMaps)
			continue
		}
		maps, err := a.fetchTemplatemaps(ctx, client, siteID, item.ID)
		if err != nil {
			return nil, 0, err
		}
		byID := map[string]apiTemplatemap{}
		for _, tm := range maps {
			byID[tm.id()] = tm
		}
		live[item.ID] = byID
		for _, want := range item.ArchiveMaps {
			if got, ok := byID[want.ID]; ok && want.ID != "" && sameMap(want, got) {
				continue
			}
			pending++
		}
	}
	return live, pending, nil
}

// pushTemplatemaps はマニフェストのマッピングを MT に合わせます。
// 送った件数を返します。**MT にあって手元に無いものは消しません**
// （テンプレート本体と同じ扱い。消すなら templatemaps delete を明示的に）。
func (a *App) pushTemplatemaps(ctx context.Context, client *api.Client, siteID string, item *manifestItem, live map[string]apiTemplatemap) (int, error) {
	sent := 0
	for _, want := range item.ArchiveMaps {
		if got, ok := live[want.ID]; ok && want.ID != "" && sameMap(want, got) {
			continue
		}
		body := map[string]any{"archiveType": want.ArchiveType, "isPreferred": want.IsPreferred}
		for k, v := range map[string]string{"fileTemplate": want.FileTemplate, "buildType": want.BuildType} {
			if v != "" {
				body[k] = v
			}
		}
		payload, err := marshalNoEscape(body)
		if err != nil {
			return sent, err
		}
		req := api.Request{
			Method: http.MethodPut,
			Path:   "/sites/" + siteID + "/templates/" + item.ID + "/templatemaps/" + want.ID,
			Form:   url.Values{"templatemap": []string{payload}},
		}
		if want.ID == "" {
			req.Method = http.MethodPost
			req.Path = "/sites/" + siteID + "/templates/" + item.ID + "/templatemaps"
		}
		if a.g.dryRun {
			fmt.Fprintf(a.env.Out, "%s %s  %s\n", req.Method, req.Path, want.ArchiveType)
			sent++
			continue
		}
		resp, err := client.Do(ctx, req)
		if err != nil {
			return sent, a.describeError(err, resp)
		}
		if want.ID == "" {
			var tm apiTemplatemap
			if err := json.Unmarshal(resp.Body, &tm); err == nil {
				want.ID = tm.id()
			}
		}
		sent++
	}
	return sent, nil
}

// sameTemplate は送る必要が無いかを返します。本文だけでなく、名前や
// 出力ファイル名の変更も拾います。
func sameTemplate(item *manifestItem, t apiTemplate, text string) bool {
	return t.Text == text &&
		t.Name == item.Name &&
		t.OutputFile == item.OutputFile &&
		t.BuildType == item.BuildType &&
		t.LinkToFile == item.LinkToFile
}

// reportUntracked は MT にあって手元に無いものを教えます。消しはしません。
// 手元が正だと思い込んでいると、MT 側の変更を黙って踏み潰すことになります。
func (a *App) reportUntracked(items []apiTemplate, m *manifest) error {
	known := map[string]bool{}
	for _, item := range m.Templates {
		known[item.ID] = true
	}
	var extra []string
	for _, t := range items {
		if !known[t.id()] {
			extra = append(extra, t.Name)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		fmt.Fprintf(a.env.Err, i18n.T(
			"MT にあって手元に無いテンプレートが %d 件あります（触れていません）: %s\n",
			"%d templates exist on MT but not here (left alone): %s\n"), len(extra), strings.Join(extra, ", "))
	}
	return nil
}

// fetchTemplates はサイトのテンプレートを本文まで取ります。
// 一覧が本文を返すので1往復で済みます。
func (a *App) fetchTemplates(ctx context.Context, client *api.Client, siteID string) ([]apiTemplate, error) {
	resp, err := a.fetchAll(ctx, client, api.Request{
		Method: http.MethodGet,
		Path:   "/sites/" + siteID + "/templates",
		Query:  url.Values{"fields": []string{"id,name,type,text,outputFile,buildType,linkToFile"}},
	}, 0)
	if err != nil {
		return nil, a.describeError(err, resp)
	}
	var page struct {
		Items []apiTemplate `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (a *App) writeManifest(dir string, m *manifest) error {
	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return a.env.WriteFile(path.Join(dir, manifestFile), append(buf, '\n'))
}

func (a *App) readManifest(dir string) (*manifest, error) {
	buf, err := a.env.ReadFile(path.Join(dir, manifestFile))
	if err != nil {
		return nil, usagef(i18n.T(
			"%s を読めません（先に mtb sites templates pull を実行してください）: %v",
			"cannot read %s (run mtb sites templates pull first): %v"), manifestFile, err)
	}
	m := &manifest{}
	if err := json.Unmarshal(buf, m); err != nil {
		return nil, fmt.Errorf(i18n.T("%s が壊れています: %w", "%s is malformed: %w"), manifestFile, err)
	}
	return m, nil
}

// uniqueFileName はテンプレート名からファイル名を作ります。
// 名前は日本語でもよく、/ が入ることもあるので落とします。
func uniqueFileName(name, id string, used map[string]bool) string {
	base := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < 0x20 {
			return '_'
		}
		return r
	}, strings.TrimSpace(name))
	if base == "" || strings.Trim(base, "._") == "" {
		base = "template-" + id
	}
	file := base + templateExt
	for i := 2; used[file]; i++ {
		file = fmt.Sprintf("%s-%d%s", base, i, templateExt)
	}
	used[file] = true
	return file
}
