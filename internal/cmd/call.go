package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// newCallCmd はエンドポイント ID で呼ぶ汎用コマンドです。
//
// MT のプラグインは Data API にエンドポイントを足せますが、mtb の拡張
// （mtb extensions）が出ていなければコマンドはありません。mtb api で叩けるものの、
// 利用者がパスを組み立てる必要があり、パラメータも分かりません。そこで、
// **サーバ自身に聞く**: 経路とメソッドは /endpoints から、パラメータはサーバの
// OpenAPI 定義（GET /）から取ります。プラグインが openapi_handler を持っていれば、
// そのエンドポイントの説明もサーバが返します。
//
// 埋め込みの定義には頼らないので、どのプラグインのエンドポイントでも呼べます。
// その代わり、呼ぶたびに /endpoints を1回読みます（--describe は定義も読む）。
func (a *App) newCallCmd() *cobra.Command {
	var (
		o        rawOptions
		paths    []string
		describe bool
	)
	cmd := &cobra.Command{
		Use:   i18n.T("call <エンドポイントID> [パスの値...]", "call <endpoint-id> [path-values...]"),
		Short: i18n.T("エンドポイント ID で任意のエンドポイントを呼ぶ（拡張の無いプラグイン向け）", "Call any endpoint by its endpoint ID (for plugins without an extension)"),
		Long: i18n.T(
			"MT のプラグインが足したエンドポイントを、mtb の拡張が無くても呼びます。\n"+
				"経路とメソッドはサーバの /endpoints から引き、:site_id などのパスの値は\n"+
				"位置引数（経路に出てくる順）か --path 名前=値 で埋めます。\n\n"+
				"  mtb endpoints list --fields id,verb,route,component   # ID を調べる\n"+
				"  mtb call update_entry_author --describe                # パラメータをサーバに聞く\n"+
				"  mtb call update_entry_author 1 42 --form authorName=editor\n"+
				"  mtb call list_comments 1 --query limit=5 --all\n\n"+
				"本文・クエリ・ファイルの渡し方と --all・--fields・--set は mtb api と同じです。\n"+
				"--dry-run でも経路を引くためにサーバへ1回問い合わせます。\n"+
				"そのプラグインの拡張があるなら、そちらのコマンドを使ってください。",
			"Calls an endpoint an MT plugin added, even without an mtb extension for it.\n"+
				"The route and method come from the server's /endpoints; path values such as\n"+
				":site_id are filled from the positional arguments (in route order) or --path name=value.\n\n"+
				"  mtb endpoints list --fields id,verb,route,component   # find the ID\n"+
				"  mtb call update_entry_author --describe                # ask the server for the parameters\n"+
				"  mtb call update_entry_author 1 42 --form authorName=editor\n"+
				"  mtb call list_comments 1 --query limit=5 --all\n\n"+
				"Body, query and file parameters and --all, --fields and --set work as in mtb api.\n"+
				"Even with --dry-run, the server is asked once to look up the route.\n"+
				"If the plugin has an extension, prefer its commands."),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ep, err := a.lookupEndpoint(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if describe {
				return a.describeEndpoint(cmd.Context(), ep)
			}
			named := map[string]string{}
			for _, p := range paths {
				k, v, ok := strings.Cut(p, "=")
				if !ok || k == "" {
					return usagef(i18n.T("--path は 名前=値 の形で書いてください: %s", "--path must be name=value: %s"), p)
				}
				named[strings.TrimPrefix(k, ":")] = v
			}
			path, err := a.fillRoute(ep, args[1:], named)
			if err != nil {
				return err
			}
			return a.sendRaw(cmd, ep.Verb, path, &o)
		},
	}
	o.addFlags(cmd)
	f := cmd.Flags()
	f.StringArrayVar(&paths, "path", nil, i18n.T("パスの値を名前で渡す（site_id=1、繰り返し可。位置引数より優先）", "pass a path value by name (site_id=1; repeatable; beats the positional arguments)"))
	f.BoolVar(&describe, "describe", false, i18n.T("呼ばずに、サーバの定義からパラメータを表示する", "do not call; show the parameters from the server's definition"))
	return cmd
}

// serverEndpoint は /endpoints の1件です。
type serverEndpoint struct {
	ID        string          `json:"id"`
	Route     string          `json:"route"`
	Verb      string          `json:"verb"`
	Version   json.RawMessage `json:"version"`
	Component struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"component"`
}

func (e *serverEndpoint) version() int {
	s := strings.Trim(string(e.Version), `"`)
	n, _ := strconv.Atoi(s)
	return n
}

// lookupEndpoint は ID からエンドポイントを引きます。
//
// 同じ ID が版違いで並ぶことがある（authenticate の v1 と v3 など。経路は同じ）。
// そのときは新しい版を採る。メソッドか経路が違うものが並んだら、どれを呼ぶべきか
// 分からないので断る。
func (a *App) lookupEndpoint(ctx context.Context, id string) (*serverEndpoint, error) {
	all, err := a.serverEndpoints(ctx)
	if err != nil {
		return nil, err
	}
	var hits []*serverEndpoint
	for _, e := range all {
		if e.ID == id {
			hits = append(hits, e)
		}
	}
	if len(hits) == 0 {
		var similar []string
		for _, e := range all {
			if strings.Contains(e.ID, id) || strings.Contains(id, e.ID) {
				similar = append(similar, e.ID)
			}
		}
		sort.Strings(similar)
		similar = uniqueStrings(similar)
		msg := fmt.Sprintf(i18n.T(
			"エンドポイント %s はこの MT にありません（mtb endpoints list --fields id,verb,route で一覧できます）",
			"this MT has no endpoint %s (list them with: mtb endpoints list --fields id,verb,route)"), id)
		if len(similar) > 0 {
			if len(similar) > 8 {
				similar = similar[:8]
			}
			msg += "\n" + i18n.T("似た ID: ", "similar IDs: ") + strings.Join(similar, ", ")
		}
		return nil, &usageError{fmt.Errorf("%s", msg)}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].version() > hits[j].version() })
	for _, h := range hits[1:] {
		if h.Verb != hits[0].Verb || h.Route != hits[0].Route {
			return nil, &usageError{fmt.Errorf(i18n.T(
				"エンドポイント %s は複数あり、どれを呼ぶか決められません（%s %s と %s %s）。mtb api でパスを指定してください",
				"endpoint %s is ambiguous (%s %s and %s %s); call it with mtb api and the path"),
				id, hits[0].Verb, hits[0].Route, h.Verb, h.Route)}
		}
	}
	return hits[0], nil
}

func (a *App) serverEndpoints(ctx context.Context) ([]*serverEndpoint, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(ctx, api.Request{Method: http.MethodGet, Path: "/endpoints"})
	if err != nil {
		return nil, a.describeError(err, resp)
	}
	var body struct {
		Items []*serverEndpoint `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return nil, fmt.Errorf(i18n.T("/endpoints の応答を読めません: %w", "cannot read the /endpoints response: %w"), err)
	}
	return body.Items, nil
}

var routeVar = regexp.MustCompile(`:(\w+)`)

// fillRoute は経路の :名前 を埋めます。名前で渡したもの → 位置引数（経路の順）→
// プロファイルの既定サイト（site_id だけ）の順で決めます。
func (a *App) fillRoute(ep *serverEndpoint, positional []string, named map[string]string) (string, error) {
	vars := routeVar.FindAllStringSubmatch(ep.Route, -1)
	values := map[string]string{}
	next := 0
	for _, m := range vars {
		name := m[1]
		if v, ok := named[name]; ok {
			values[name] = v
			continue
		}
		if next < len(positional) {
			values[name] = positional[next]
			next++
		}
	}
	if next < len(positional) {
		return "", usagef(i18n.T(
			"パスの値が多すぎます（%s が取るのは %d 個）",
			"too many path values (%s takes %d)"), ep.Route, len(vars))
	}
	for name := range named {
		if !strings.Contains(ep.Route, ":"+name) {
			return "", usagef(i18n.T("--path %s は経路 %s にありません", "--path %s is not in the route %s"), name, ep.Route)
		}
	}
	var missing []string
	for _, m := range vars {
		name := m[1]
		if values[name] != "" {
			continue
		}
		if name == "site_id" {
			if r, err := a.resolve(); err == nil && r.SiteID != "" {
				values[name] = r.SiteID
				continue
			}
		}
		missing = append(missing, name)
	}
	if len(missing) > 0 {
		return "", usagef(i18n.T(
			"経路 %s のパスの値が足りません: %s（位置引数か --path 名前=値 で渡してください）",
			"the route %s needs path values: %s (give them positionally or with --path name=value)"),
			ep.Route, strings.Join(missing, ", "))
	}
	return routeVar.ReplaceAllStringFunc(ep.Route, func(s string) string {
		return values[s[1:]]
	}), nil
}

// describeEndpoint はサーバの OpenAPI 定義から、そのエンドポイントの
// パラメータを表示します。説明文はサーバが返したものをそのまま出します。
func (a *App) describeEndpoint(ctx context.Context, ep *serverEndpoint) error {
	out := a.env.Out
	fmt.Fprintf(out, "%s %s\n", ep.Verb, ep.Route)
	fmt.Fprintf(out, i18n.T("ID: %s（%s、v%d）\n", "ID: %s (%s, v%d)\n"), ep.ID, firstNonEmptyString(ep.Component.Name, ep.Component.ID), ep.version())

	c, err := a.client()
	if err != nil {
		return err
	}
	resp, err := c.Do(ctx, api.Request{Method: http.MethodGet, Path: "/"})
	if err != nil {
		return a.describeError(err, resp)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
			Parameters  []struct {
				Name        string `json:"name"`
				In          string `json:"in"`
				Required    bool   `json:"required"`
				Description string `json:"description"`
				Ref         string `json:"$ref"`
			} `json:"parameters"`
			RequestBody struct {
				Content map[string]struct {
					Schema struct {
						Properties map[string]struct {
							Type        string `json:"type"`
							Description string `json:"description"`
							Ref         string `json:"$ref"`
						} `json:"properties"`
					} `json:"schema"`
				} `json:"content"`
			} `json:"requestBody"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(resp.Body, &doc); err != nil {
		return fmt.Errorf(i18n.T("サーバの OpenAPI 定義を読めません: %w", "cannot read the server's OpenAPI definition: %w"), err)
	}
	op, ok := doc.Paths[routeVar.ReplaceAllString(ep.Route, "{$1}")][strings.ToLower(ep.Verb)]
	if !ok {
		fmt.Fprintln(out, "\n"+i18n.T(
			"サーバの定義にこのエンドポイントの説明がありません（プラグインが openapi_handler を持っていない）。",
			"The server's definition does not describe this endpoint (the plugin has no openapi_handler)."))
		return nil
	}
	if op.Summary != "" {
		fmt.Fprintf(out, "\n%s\n", op.Summary)
	}
	if d := strings.TrimSpace(op.Description); d != "" {
		fmt.Fprintf(out, "\n%s\n", d)
	}

	var pathVars []string
	for _, m := range routeVar.FindAllStringSubmatch(ep.Route, -1) {
		pathVars = append(pathVars, m[1])
	}
	if len(pathVars) > 0 {
		fmt.Fprintf(out, "\n%s %s\n", i18n.T("パスの値（位置引数の順）:", "Path values (positional order):"), strings.Join(pathVars, " "))
	}
	var query []string
	for _, p := range op.Parameters {
		if p.In != "query" {
			continue
		}
		query = append(query, paramLine(p.Name, "", p.Required, p.Description))
	}
	if len(query) > 0 {
		fmt.Fprintln(out, "\n"+i18n.T("クエリ（--query 名前=値）:", "Query (--query name=value):"))
		for _, l := range query {
			fmt.Fprintln(out, l)
		}
	}
	for mime, content := range op.RequestBody.Content {
		flag := "--form"
		if strings.HasPrefix(mime, "multipart/") {
			flag = "--form / --file"
		}
		fmt.Fprintf(out, i18n.T("\n本文（%s 名前=値、%s）:\n", "\nBody (%s name=value, %s):\n"), flag, mime)
		names := make([]string, 0, len(content.Schema.Properties))
		for n := range content.Schema.Properties {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			p := content.Schema.Properties[n]
			kind := p.Type
			if p.Ref != "" {
				kind = "object"
			}
			fmt.Fprintln(out, paramLine(n, kind, false, p.Description))
		}
	}
	return nil
}

func paramLine(name, kind string, required bool, desc string) string {
	line := "  " + name
	var notes []string
	if kind != "" {
		notes = append(notes, kind)
	}
	if required {
		notes = append(notes, i18n.T("必須", "required"))
	}
	if len(notes) > 0 {
		line += " (" + strings.Join(notes, ", ") + ")"
	}
	if d := oneLine(desc); d != "" {
		line += " — " + d
	}
	return line
}

func uniqueStrings(s []string) []string {
	out := s[:0]
	for i, v := range s {
		if i == 0 || v != s[i-1] {
			out = append(out, v)
		}
	}
	return out
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
