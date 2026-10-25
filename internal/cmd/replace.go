package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// replaceFlag は同名のアセットを消してから送るフラグです。
//
// **MT はアセットを上書きできません。** 同じ名前で送ると 409 を返し、
// Data API に上書きの手段はありません（overwrite=1 は管理画面の二段階
// アップロード用の別経路、auto_rename_if_exists はファイル名をハッシュに
// 変えてしまう）。テーマの静的ファイルは名前が変わっては困るので、
// 消してから送るしかありません。
//
// デザインの差し替えは日常の操作です。そのたびに「消してから送る」を
// 手で書かせるなら、この CLI は仕事をしていません。
const replaceFlag = "replace"

// hasUploadPath は「サイトの中の場所を指定してアセットを置く」操作かを返します。
// 記事や利用者の取り込みはファイルを置かないので対象外です。
//
// **アセットのアップロードに限ります。** 拡張は file と path を取る別の
// エンドポイント（テーマやサイトのファイルを置くもの）を足します。そこに
// --replace が付くと、同名の「アセット」を探して消しに行ってしまいます。
func hasUploadPath(sc *spec.Command) bool {
	for _, op := range sc.Ops {
		if fileParam(op) != nil && hasPathParam(op) && strings.HasSuffix(op.Path, "/assets/upload") {
			return true
		}
	}
	return false
}

// assetReplacer は送る前に同名のアセットを消します。
type assetReplacer struct {
	app     *App
	client  *api.Client
	siteID  string
	siteURL string
	removed int
}

// newAssetReplacer は差し替えの準備をします。サイトのURLを1度だけ引くのは、
// 「同じ名前の別の場所にあるアセット」を消さないためです。ファイル名だけで
// 消すと、別フォルダの同名ファイルを巻き込みます。
func (a *App) newAssetReplacer(ctx context.Context, client *api.Client, siteID string) (*assetReplacer, error) {
	if siteID == "" {
		return nil, usagef("%s", i18n.T(
			"--replace にはサイトIDが要ります", "--replace needs a site ID"))
	}
	resp, err := client.Do(ctx, api.Request{
		Method: http.MethodGet,
		Path:   "/sites/" + siteID,
		Query:  url.Values{"fields": []string{"url"}},
	})
	if err != nil {
		return nil, fmt.Errorf(i18n.T(
			"--replace: サイトのURLを取得できません: %w",
			"--replace: cannot read the site URL: %w"), err)
	}
	// サイトのURLは url。siteUrl という項目は無い（実機で空が返る）。
	var site struct {
		SiteURL string `json:"url"`
	}
	if err := json.Unmarshal(resp.Body, &site); err != nil || site.SiteURL == "" {
		return nil, fmt.Errorf(i18n.T(
			"--replace: サイト（ID %s）のURLが分かりません",
			"--replace: cannot determine the URL of site %s"), siteID)
	}
	return &assetReplacer{app: a, client: client, siteID: siteID, siteURL: site.SiteURL}, nil
}

// remove は送り先にある同名のアセットを消します。無ければ何もしません。
func (r *assetReplacer) remove(ctx context.Context, dest, filename string) error {
	want := r.expectedURL(dest, filename)
	resp, err := r.client.Do(ctx, api.Request{
		Method: http.MethodGet,
		Path:   "/sites/" + r.siteID + "/assets",
		Query: url.Values{
			"search":       []string{filename},
			"searchFields": []string{"file_name"},
			"fields":       []string{"id,url"},
			"limit":        []string{"100"},
		},
	})
	if err != nil {
		return err
	}
	// id は数値で返ることも文字列で返ることもあります（MT はここが
	// 揃っていない）。どちらでも読めるように生のまま受けます。
	var page struct {
		Items []struct {
			ID  json.RawMessage `json:"id"`
			URL string          `json:"url"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &page); err != nil {
		return err
	}
	for _, item := range page.Items {
		if item.URL != want {
			continue
		}
		id := strings.Trim(string(item.ID), `"`)
		if _, err := r.client.Do(ctx, api.Request{
			Method: http.MethodDelete,
			Path:   "/sites/" + r.siteID + "/assets/" + id,
		}); err != nil {
			return fmt.Errorf(i18n.T(
				"--replace: 既存のアセット %s を消せません: %w",
				"--replace: cannot delete the existing asset %s: %w"), id, err)
		}
		r.removed++
		if r.app.g.verbose {
			fmt.Fprintf(r.app.env.Err, i18n.T(
				"既存のアセット %s を消しました（%s）\n",
				"deleted the existing asset %s (%s)\n"), id, want)
		}
		return nil
	}
	return nil
}

// expectedURL は送り先のURLを組み立てます。ファイル名だけで一致を見ると
// 別フォルダの同名ファイルを消してしまうので、URL まで合わせます。
func (r *assetReplacer) expectedURL(dest, filename string) string {
	base := strings.TrimSuffix(r.siteURL, "/")
	dest = strings.Trim(dest, "/")
	if dest == "" {
		return base + "/" + filename
	}
	return base + "/" + path.Join(dest, filename)
}
