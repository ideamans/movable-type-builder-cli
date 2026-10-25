package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// ダンプは「2つの MT のどこが違うか」を機械で出すためのものです。
//
// ステージングでデザインを組んで本番へ手で反映する、という進め方が普通に
// あります。そのとき一番時間を食うのは反映そのものではなく、**どこが違うのかを
// 突き止めること**です。管理画面を2つ開いて見比べるしかない。
//
// ディレクトリに種類ごとの YAML を出すので、あとは普通の diff で読めます。
// プロファイルを2つ用意して両方ダンプすれば、それが差分の一覧になります。
//
//	mtb dump --profile stg  --dir ./dump/stg
//	mtb dump --profile prod --dir ./dump/prod
//	diff -ru ./dump/prod ./dump/stg
const (
	dumpDirFlag  = "dir"
	dumpHashFlag = "asset-hash"
)

func (a *App) newDumpCmd() *cobra.Command {
	var (
		dir       string
		assetHash bool
		sites     string
	)
	cmd := &cobra.Command{
		Use:   "dump",
		Short: i18n.T("MT の中身をディレクトリに YAML で書き出す（差分を取るため）", "Write the contents of an MT to a directory as YAML, for diffing"),
		Long: i18n.T(
			"接続先の中身を、種類ごとの YAML にして書き出します。2つの環境で\n"+
				"実行して diff を取ると、どこが違うかが分かります。\n\n"+
				"  mtb dump --profile stg  --dir ./dump/stg\n"+
				"  mtb dump --profile prod --dir ./dump/prod\n"+
				"  diff -ru ./dump/prod ./dump/stg\n\n"+
				"**ID・時刻・作成者は書き出しません。** 環境ごとに違うので、残すと\n"+
				"中身が同じでも全行が差分になります。並べ替えと突き合わせは\n"+
				"basename や name などの自然キーで行います。\n\n"+
				"統計・ログ・コメントは対象外です（環境で違って当たり前のもの）。",
			"Write the contents of the target server as one YAML file per object type.\n"+
				"Run it against two environments and diff the directories to see what differs.\n\n"+
				"  mtb dump --profile stg  --dir ./dump/stg\n"+
				"  mtb dump --profile prod --dir ./dump/prod\n"+
				"  diff -ru ./dump/prod ./dump/stg\n\n"+
				"**IDs, timestamps and authors are left out.** They differ between\n"+
				"environments, so keeping them would make every line differ even when the\n"+
				"content is identical. Objects are ordered and matched by natural keys\n"+
				"such as basename and name.\n\n"+
				"Statistics, logs and comments are out of scope: they are expected to differ."),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runDump(context.Background(), dir, sites, assetHash)
		},
	}
	f := cmd.Flags()
	f.StringVar(&dir, dumpDirFlag, "", i18n.T("書き出す先のディレクトリ（必須）", "directory to write into (required)"))
	f.StringVar(&sites, "sites", "", i18n.T("対象のサイトIDをカンマ区切りで（既定は全サイト）", "comma-separated site IDs (default: every site)"))
	f.BoolVar(&assetHash, dumpHashFlag, false, i18n.T(
		"アセットの中身を取得して MD5 を記録する（1ファイル1リクエスト増えます）",
		"download each asset and record its MD5 (one extra request per file)"))
	return cmd
}

func (a *App) runDump(ctx context.Context, dir, sites string, assetHash bool) error {
	if dir == "" {
		return usagef("%s", i18n.T("--dir が要ります", "--dir is required"))
	}
	client, err := a.client()
	if err != nil {
		return err
	}

	for _, kind := range append(append([]dumpKind{}, systemKinds...), envKinds...) {
		if err := a.dumpKindTo(ctx, client, kind, dumpSite{}, path.Join(dir, "system"), false); err != nil {
			return err
		}
	}
	if err := a.dumpEnvironment(ctx, client, path.Join(dir, "system")); err != nil {
		return err
	}

	targets, err := a.dumpSites(ctx, client, sites)
	if err != nil {
		return err
	}
	for _, site := range targets {
		base := path.Join(dir, "sites", site.dirName())
		// 運ぶ設定と、そのサーバがどこにあるかを分けて書きます。混ぜると
		// 「差分だが運んではいけない」行が紛れ込んで、判断できなくなります。
		portable, environment := splitSiteSettings(site.Settings)
		if err := a.writeDumpFile(path.Join(base, "site.yaml"), []dumpItem{{Key: site.Name, Object: portable}}); err != nil {
			return err
		}
		if err := a.writeDumpFile(path.Join(base, "site-environment.yaml"), []dumpItem{{Key: site.Name, Object: environment}}); err != nil {
			return err
		}
		for _, kind := range siteKinds {
			if err := a.dumpKindTo(ctx, client, kind, site, base, assetHash); err != nil {
				return err
			}
		}
	}
	fmt.Fprintf(a.env.Err, i18n.T("%d サイトを %s に書き出しました\n", "Wrote %d sites to %s\n"), len(targets), dir)
	return nil
}

// dumpSite は書き出す対象のサイト1つ分です。
type dumpSite struct {
	ID       string
	Name     string
	URL      string
	Path     string
	Settings map[string]any
}

// dirName はサイトのディレクトリ名です。**ID は使いません**——環境をまたぐと
// 違う番号になるので、突き合わせられなくなります。
func (s dumpSite) dirName() string {
	name := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < 0x20 {
			return '_'
		}
		return r
	}, strings.TrimSpace(s.Name))
	if name == "" {
		return "site-" + s.ID
	}
	return name
}

func (a *App) dumpSites(ctx context.Context, client *api.Client, only string) ([]dumpSite, error) {
	wanted := map[string]bool{}
	for _, id := range strings.Split(only, ",") {
		if id = strings.TrimSpace(id); id != "" {
			wanted[id] = true
		}
	}
	items, err := a.fetchDumpItems(ctx, client, "/sites", nil)
	if err != nil {
		return nil, err
	}
	var out []dumpSite
	for _, obj := range items {
		id := scalarString(obj["id"])
		if len(wanted) > 0 && !wanted[id] {
			continue
		}
		name, _ := obj["name"].(string)
		siteURL, _ := obj["url"].(string)
		sitePath, _ := obj["sitePath"].(string)
		settings, _ := clean(obj).(map[string]any)
		out = append(out, dumpSite{ID: id, Name: name, URL: siteURL, Path: sitePath, Settings: settings})
	}
	return out, nil
}

func (a *App) dumpKindTo(ctx context.Context, client *api.Client, kind dumpKind, site dumpSite, dir string, assetHash bool) error {
	p := kind.Path
	if strings.Contains(p, "%s") {
		p = fmt.Sprintf(p, site.ID)
	}
	raw, err := a.fetchDumpItems(ctx, client, p, kind.Query)
	if err != nil {
		// そのMTに無い機能（プラグイン由来のもの）は黙って飛ばします。
		if e, ok := api.AsError(err); ok && e.NotFound() {
			return nil
		}
		return err
	}

	parents := map[string]string{}
	if kind.Hierarchical {
		parents = parentKeys(raw, kind)
	}

	items := make([]dumpItem, 0, len(raw))
	for _, obj := range raw {
		key := kind.naturalKey(obj)
		if kind.Hierarchical {
			key = parents[scalarString(obj["id"])]
			obj = withParentKey(obj, parents)
		}
		if kind.File == "plugins" {
			obj = withoutKeys(obj, pluginNoise...)
		}
		if kind.File == "content-types" {
			// コンテンツタイプはフィールドと一緒でないと意味が無い。
			// dataLabel は「どのフィールドを名前に使うか」の指定で、値は
			// そのフィールドの uniqueID。環境ごとに違う値なので、
			// **指している名前**に畳んでから落とす。
			obj = a.withContentFields(ctx, client, site.ID, obj)
		}
		if kind.File == "assets" {
			// **filePath も url も環境ごとに違います。** サーバの絶対パスと
			// ホスト名が入っているので、残すと全件が差分になります。
			// サイトルートからの相対パスに畳むと、それがそのまま鍵になります
			// （同じ名前のファイルが別フォルダにあっても区別できる）。
			rel := siteRelativePath(obj, site)
			obj = withoutKeys(obj, "filePath", "url")
			obj["path"] = rel
			key = rel
			if assetHash {
				if sum, err := a.assetMD5(ctx, absoluteURL(obj, site)); err == nil {
					obj["md5"] = sum
				}
			}
		}
		cleaned, _ := clean(obj).(map[string]any)
		if cleaned == nil {
			cleaned = map[string]any{}
		}
		body, _ := yamlBytes(cleaned)
		items = append(items, dumpItem{Key: key, Object: cleaned, tiebreak: string(body)})
	}
	sortObjects(items)
	return a.writeDumpFile(path.Join(dir, kind.File+".yaml"), items)
}

// parentKeys は階層のあるオブジェクトに、根からの道を付けます。
// **親を ID で持ったままにすると、環境間で必ず差分になります。**
func parentKeys(items []map[string]any, kind dumpKind) map[string]string {
	self := map[string]string{}
	parent := map[string]string{}
	for _, obj := range items {
		id := scalarString(obj["id"])
		self[id] = kind.naturalKey(obj)
		parent[id] = scalarString(obj["parent"])
	}
	out := map[string]string{}
	for id := range self {
		var parts []string
		for cur, guard := id, 0; cur != "" && cur != "0" && guard < 32; guard++ {
			name, ok := self[cur]
			if !ok {
				break
			}
			parts = append([]string{name}, parts...)
			cur = parent[cur]
		}
		out[id] = strings.Join(parts, "/")
	}
	return out
}

// withParentKey は parent を親の道に差し替えます。
func withParentKey(obj map[string]any, parents map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range obj {
		out[k] = v
	}
	if p := scalarString(obj["parent"]); p != "" && p != "0" {
		out["parent"] = parents[p]
	} else {
		delete(out, "parent")
	}
	return out
}

// assetMD5 はアセットの中身を取ってハッシュにします。
//
// **アセットの資源は中身について何も教えてくれません。** 持っているのは
// 大きさと寸法だけで、同じ大きさの別の画像は見分けが付きません。実物を
// 取ってこないと「同じファイルか」は言えないので、URL から素で取ります
// （公開ファイルなので認証は要りません）。1ファイル1リクエスト増えるため
// 既定では行いません。
func (a *App) assetMD5(ctx context.Context, u string) (string, error) {
	if u == "" {
		return "", errors.New("no url")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	resp, err := a.env.NewDoer(a.g.insecure, a.g.timeout).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("%s: %d", u, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sum := md5.Sum(body)
	return hex.EncodeToString(sum[:]), nil
}

// fetchDumpItems は一覧を全件取ります。
func (a *App) fetchDumpItems(ctx context.Context, client *api.Client, p string, query map[string]string) ([]map[string]any, error) {
	q := url.Values{}
	for k, v := range query {
		q.Set(k, v)
	}
	resp, err := a.fetchAll(ctx, client, api.Request{Method: http.MethodGet, Path: p, Query: q}, 0)
	if err != nil {
		return nil, err
	}
	var page struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (a *App) writeDumpFile(name string, items []dumpItem) error {
	doc := make([]any, 0, len(items))
	for _, item := range items {
		doc = append(doc, item.Object)
	}
	body, err := yamlBytes(doc)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		body = []byte("[]\n")
	}
	return a.env.WriteFile(name, body)
}

// scalarString は数値でも文字列でも来る ID を文字列にします。
func scalarString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%d", int64(t))
	case json.Number:
		return t.String()
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}

// siteRelativePath はアセットの置き場所をサイトルートからの相対にします。
// url が引ければそちらを、駄目なら filePath を使います。
func siteRelativePath(obj map[string]any, site dumpSite) string {
	if u, _ := obj["url"].(string); u != "" && site.URL != "" {
		if rel, ok := strings.CutPrefix(u, site.URL); ok {
			return strings.TrimPrefix(rel, "/")
		}
	}
	if p, _ := obj["filePath"].(string); p != "" && site.Path != "" {
		if rel, ok := strings.CutPrefix(p, site.Path); ok {
			return strings.TrimPrefix(rel, "/")
		}
	}
	name, _ := obj["filename"].(string)
	return name
}

// absoluteURL はハッシュを取るための実物の URL です。
func absoluteURL(obj map[string]any, site dumpSite) string {
	if rel, _ := obj["path"].(string); rel != "" && site.URL != "" {
		return strings.TrimSuffix(site.URL, "/") + "/" + rel
	}
	u, _ := obj["url"].(string)
	return u
}

// withoutKeys は指定した項目を落とした複製を返します。
func withoutKeys(obj map[string]any, keys ...string) map[string]any {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	out := make(map[string]any, len(obj))
	for k, v := range obj {
		if !drop[k] {
			out[k] = v
		}
	}
	return out
}

// withContentFields はコンテンツタイプにフィールド定義を埋め込み、
// dataLabel を「どのフィールドか」が分かる形に畳みます。
func (a *App) withContentFields(ctx context.Context, client *api.Client, siteID string, obj map[string]any) map[string]any {
	id := scalarString(obj["id"])
	if id == "" {
		return obj
	}
	fields, err := a.fetchDumpItems(ctx, client, "/sites/"+siteID+"/contentTypes/"+id+"/fields", nil)
	if err != nil {
		return obj
	}
	byUID := map[string]string{}
	for _, f := range fields {
		if uid, _ := f["uniqueID"].(string); uid != "" {
			byUID[uid], _ = f["label"].(string)
		}
	}
	out := withoutKeys(obj, "dataLabel")
	if uid, _ := obj["dataLabel"].(string); uid != "" {
		if label := byUID[uid]; label != "" {
			out["dataLabelField"] = label
		}
	}
	items := make([]any, 0, len(fields))
	for _, f := range fields {
		items = append(items, f)
	}
	out["fields"] = sortIfKeyed(items)
	return out
}

// splitSiteSettings はサイト設定を「運ぶもの」と「その環境のもの」に分けます。
func splitSiteSettings(settings map[string]any) (portable, environment map[string]any) {
	portable, environment = map[string]any{}, map[string]any{}
	for k, v := range settings {
		if environmentKeys[k] {
			environment[k] = v
		} else {
			portable[k] = v
		}
	}
	return portable, environment
}
