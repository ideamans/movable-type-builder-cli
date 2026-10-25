package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"sort"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
)

// 環境の記録。**運べないが、違っていると分かるだけで価値があるもの**です。
//
// サイトの中身は運べても、その前提となるサーバの設定は API から変えられません
// （mt-config.cgi と管理画面の領分）。それでも**差分として見えれば**、移行の前に
// 気づけます。実際に検証したときは、片方にだけ EmailAddressMain があり、
// サイトを完全に再現しても移行先ではメール送信が 500 で落ちる状態でした。
//
// ここに出るものが違ったら、**運ぶ前に人が揃える**必要があります。
//
//   - プラグイン — 無いものに依存したテンプレートは公開で落ちる
//     （ブロックエディタもプラグイン）
//   - テキストフィルタ — format: markdown の記事は、その書式が無い MT では読めない
//   - エンドポイント — その MT が持っていない機能は、叩いた時点で 404 になる
//   - サーバのバージョン
var envKinds = []dumpKind{
	{File: "plugins", Path: "/plugins", Key: []string{"name"}},
	{File: "text-filters", Path: "/textFilters", Key: []string{"key"}},
}

// pluginNoise はプラグインの応答のうち、環境の比較に関係しないものです。
// icon は SVG が丸ごと入っていて、差分を読めなくします。
var pluginNoise = []string{"icon", "description", "authorLink", "authorName", "documentLink", "pluginLink", "configLink"}

// dumpEnvironment はサーバのバージョンと、持っているエンドポイントを書きます。
//
// エンドポイントは 266 個あり多く見えますが、**その MT に無い機能を移行先で
// 使おうとしていないか**は、この差分でしか分かりません（叩くまで 404 で気づけない）。
func (a *App) dumpEnvironment(ctx context.Context, client *api.Client, dir string) error {
	ver, err := a.fetchDumpObject(ctx, client, "/version")
	if err == nil {
		if err := a.writeDumpFile(path.Join(dir, "server.yaml"), []dumpItem{{Object: clean(ver).(map[string]any)}}); err != nil {
			return err
		}
	}
	eps, err := a.fetchDumpItems(ctx, client, "/endpoints", nil)
	if err != nil {
		return nil // 版によっては無い
	}
	routes := make([]string, 0, len(eps))
	for _, e := range eps {
		verb, _ := e["verb"].(string)
		route, _ := e["route"].(string)
		if verb != "" && route != "" {
			routes = append(routes, verb+" "+route)
		}
	}
	sort.Strings(routes)
	items := make([]any, 0, len(routes))
	for _, r := range routes {
		items = append(items, r)
	}
	body, err := yamlBytes(items)
	if err != nil {
		return err
	}
	return a.env.WriteFile(path.Join(dir, "endpoints.yaml"), body)
}

// fetchDumpObject は一覧ではない応答を1つ取ります。
func (a *App) fetchDumpObject(ctx context.Context, client *api.Client, p string) (map[string]any, error) {
	resp, err := client.Do(ctx, api.Request{Method: http.MethodGet, Path: p})
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(resp.Body, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}
