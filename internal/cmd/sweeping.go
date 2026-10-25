package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// sweepingOps は「入れ物ごと壊す」操作です。--yes だけでは実行させません。
//
// 総当たりテストで sites delete が --yes ひとつでサイトを丸ごと消し、
// 記事・カテゴリ・テンプレートが巻き添えになりました。1件の削除と
// サイトの削除が同じ重みなのは危ないので、対象の名前を打たせます。
// エージェントにとっては「名前を調べてから消す」という一手が入り、
// 取り違えたIDで消えることがなくなります。
var sweepingOps = map[string]sweepingOp{
	"DELETE /sites/{site_id}": {
		whatJa: "サイトと、その中の記事・ページ・カテゴリ・テンプレートのすべて",
		whatEn: "the site and every entry, page, category and template in it",
		idKey:  "site_id",
	},
	"DELETE /sites/{site_id}/logs": {
		whatJa: "このサイトのログ全件",
		whatEn: "every log record of this site",
		idKey:  "site_id",
	},
	// テーマの適用はテンプレートの総入れ替えです。refresh_templates と
	// 同じ重さなのに、これまで確認が入っていませんでした。テーマを
	// 本番サイトに間違って当てるのは、この防護が想定している事故そのものです。
	"POST /sites/{site_id}/themes/{theme_id}/apply": {
		whatJa: "このサイトのテンプレート（テーマのもので総入れ替えになります）",
		whatEn: "the templates of this site (they are replaced wholesale by the theme's)",
		idKey:  "site_id",
	},
	"POST /sites/{site_id}/refresh_templates": {
		whatJa: "このサイトのテンプレート（初期状態に戻ります）",
		whatEn: "the templates of this site (they are reset to their initial state)",
		idKey:  "site_id",
	},
}

type sweepingOp struct {
	// whatJa / whatEn は何が失われるかの説明です。この表は init の時点で
	// 作られ、表示言語はまだ決まっていないので、両方を持って表示時に選びます。
	whatJa string
	whatEn string
	// idKey は対象を特定するパスパラメータ名です。
	idKey string
}

// what は表示言語に合わせた説明です。
func (s sweepingOp) what() string { return i18n.T(s.whatJa, s.whatEn) }

// confirmFlag は対象の名前を打ち込ませるフラグです。
const confirmFlag = "confirm"

// isSweeping は入れ物ごと壊す操作かを返します。
func isSweeping(op *spec.Operation) (sweepingOp, bool) {
	s, ok := sweepingOps[op.Method+" "+op.Path]
	return s, ok
}

// verifySweeping は --confirm が対象の名前と一致することを確かめます。
// 一致しなければ何も送りません。
func (a *App) verifySweeping(ctx context.Context, s sweepingOp, siteID, given string) error {
	name, err := a.siteName(ctx, siteID)
	if err != nil {
		return err
	}
	if given == "" {
		return usagef(i18n.T(
			"この操作は %s を失います。実行するなら --%s %q を付けてください（%s サイトID %s）",
			"this operation destroys %s. To go ahead, add --%s %q (on %s, site ID %s)"),
			s.what(), confirmFlag, name, a.target(), siteID)
	}
	if given != name {
		return usagef(i18n.T(
			"--%s が対象の名前と一致しません。対象は %q です（サイトID %s）",
			"--%s does not match the name of the target, which is %q (site ID %s)"),
			confirmFlag, name, siteID)
	}
	return nil
}

// siteName は確認用にサイト名を引きます。
func (a *App) siteName(ctx context.Context, siteID string) (string, error) {
	client, err := a.client()
	if err != nil {
		return "", err
	}
	resp, err := client.Do(ctx, api.Request{Method: http.MethodGet, Path: "/sites/" + siteID})
	if err != nil {
		return "", fmt.Errorf(i18n.T("対象サイトを確認できません: %w", "cannot look up the target site: %w"), err)
	}
	var site struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(resp.Body, &site); err != nil || site.Name == "" {
		return "", fmt.Errorf(i18n.T("対象サイト（ID %s）の名前を取得できません", "cannot read the name of the target site (ID %s)"), siteID)
	}
	return site.Name, nil
}
