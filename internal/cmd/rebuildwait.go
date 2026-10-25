package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// waitFlag は再構築が終わるまで呼び続けるフラグです。
//
// `GET /publish/entries` は1回で終わりません。応答の restIds に「まだ残って
// いる分」が入るので、空になるまで呼び直すのが本来の使い方です。ところが
// 応答の status は最初から "Rebuilding" で、**1回だけ呼んで終わると成功した
// ように見えます**。
//
// **さらに、呼び直すときは startTime も返さなければなりません。** 渡さないと
// MT は毎回「新しい再構築」として受け付け直し、restIds がいつまでも減りません
// （実測: 渡さないと12回呼んでも6件のまま、渡すと2回で空になった）。
// エラーにはならず Rebuilding が返り続けるので、自力では気づけません。
//
// このフラグを付けると、mtb が startTime を引き継ぎながら restIds を追い、
// それでも減らなくなったら止まって「進んでいない」と伝えます。
const waitFlag = "wait"

// rebuildOps は restIds を返す（＝繰り返し呼ぶ）操作です。
var rebuildOps = map[string]bool{
	"GET /publish/entries":     true,
	"GET /publish/contentData": true,
}

// isRebuild は繰り返し呼ぶ再構築の操作かを返します。
func isRebuild(op *spec.Operation) bool {
	return rebuildOps[op.Method+" "+op.Path]
}

// rebuildStatus は再構築の応答です。
type rebuildStatus struct {
	Status    string `json:"status"`
	RestIDs   string `json:"restIds"`
	StartTime string `json:"startTime"`
}

const (
	// rebuildMaxRounds は暴走よけです。
	rebuildMaxRounds = 100
	// rebuildStallRounds は「減らない」と判断するまでの回数です。MT は
	// 1回ごとに一定量だけ進めるので、1回の足踏みでは決めつけません。
	rebuildStallRounds = 3
	// rebuildInterval は呼び直しの間隔です。
	rebuildInterval = 2 * time.Second
)

// waitRebuild は restIds が空になるまで呼び直します。
//
// 返すのは最後の応答です。減らないまま rebuildStallRounds 回続いたら、
// 何件残っているかを添えてエラーにします。
func (a *App) waitRebuild(ctx context.Context, client *api.Client, req api.Request) (*api.Response, error) {
	resp, err := client.Do(ctx, req)
	if err != nil {
		return resp, err
	}

	prev := -1
	stalled := 0
	for round := 0; round < rebuildMaxRounds; round++ {
		st, ok := parseRebuild(resp.Body)
		if !ok {
			// restIds を持たない応答は、繰り返す対象ではありません。
			return resp, nil
		}
		rest := splitIDs(st.RestIDs)
		if len(rest) == 0 {
			return resp, nil
		}
		if prev >= 0 {
			if len(rest) >= prev {
				stalled++
			} else {
				stalled = 0
			}
		}
		if stalled >= rebuildStallRounds {
			return resp, &publishError{fmt.Errorf(i18n.T(
				"再構築が進んでいません（%d 件が残ったままです: %s）。"+
					"MT は受け付けたことしか返さないので、これは待っても終わりません。"+
					"対象を保存し直すと同期で再構築されます。"+
					"ただし値が変わらない更新では再構築されないので、実際に変わる値を送ってください。",
				"the rebuild is not progressing (%d entries still queued: %s). "+
					"MT only reports that it accepted the request, so waiting will not finish this. "+
					"Saving the objects again rebuilds them synchronously — "+
					"but an update that changes nothing does not rebuild, so send a value that actually differs."),
				len(rest), strings.Join(rest, ","))}
		}
		prev = len(rest)

		a.sleep(rebuildInterval)

		next := req
		next.Query = cloneValues(req.Query)
		next.Query.Set("ids", strings.Join(rest, ","))
		if st.StartTime != "" {
			next.Query.Set("startTime", st.StartTime)
		}
		resp, err = client.Do(ctx, next)
		if err != nil {
			return resp, err
		}
	}
	return resp, &publishError{fmt.Errorf(i18n.T(
		"再構築が %d 回呼んでも終わりませんでした",
		"the rebuild did not finish after %d rounds"), rebuildMaxRounds)}
}

// parseRebuild は再構築の応答を読みます。restIds を持たなければ false。
func parseRebuild(body []byte) (rebuildStatus, bool) {
	var st rebuildStatus
	if err := json.Unmarshal(body, &st); err != nil {
		return st, false
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(body, &probe); err != nil {
		return st, false
	}
	_, ok := probe["restIds"]
	return st, ok
}

// splitIDs はカンマ区切りの ID を、並び順によらず比べられる形にします。
// MT は毎回ばらばらの順で返すので、そのままでは増減が判定できません。
func splitIDs(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	sort.Strings(out)
	return out
}
