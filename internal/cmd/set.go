package cmd

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// setFlag と setJSONFlag はリソース本体を部品から組み立てるフラグです。
//
//	--set template.name=メインページ --set template.text=@index.html
//
// これが無いと、HTML をテンプレートに入れるだけで jq が要ります。
// 「JSON を操作したい」のではなく「ファイルと項目をやりとりしたい」だけ
// なので、問い合わせ言語を増やすのではなく、指定の形を足しました。
const (
	setFlag     = "set"
	setJSONFlag = "set-json"
)

// applySets は --set / --set-json をリソース本体に流し込みます。
//
// すでに --template などで本体が渡っていれば、その上に重ねます。
// ファイルを土台にして一部だけ差し替える、という使い方ができます。
func (a *App) applySets(cmd *cobra.Command, op *spec.Operation, form map[string]string) error {
	// 定義のあるコマンドは、そのコマンドが取る本体かどうかを検査します。
	return a.applySetsTo(cmd, form, func(target string) bool { return objectParam(op, target) != nil })
}

// applySetsTo は指定先の検査だけを差し替えられる applySets です。
//
// **mtb api には定義がありません。** プラグインが足したエンドポイントを叩く
// ためのものなので、どんな名前の本体を取るかは事前に分かりません。
// そこでは名前を検査せず、打たれたとおりに組み立てます。
func (a *App) applySetsTo(cmd *cobra.Command, form map[string]string, valid func(string) bool) error {
	pairs, err := cmd.Flags().GetStringArray(setFlag)
	if err != nil {
		return nil // そのコマンドに無ければ何もしない
	}
	jsonPairs, _ := cmd.Flags().GetStringArray(setJSONFlag)
	if len(pairs) == 0 && len(jsonPairs) == 0 {
		return nil
	}

	// 対象ごとに1つの入れ物を作り、既存の JSON があればそこから始めます。
	docs := map[string]map[string]any{}
	order := []string{}
	start := func(flag string) (map[string]any, error) {
		if doc, ok := docs[flag]; ok {
			return doc, nil
		}
		doc := map[string]any{}
		if raw, ok := form[flag]; ok && strings.TrimSpace(raw) != "" {
			if err := json.Unmarshal([]byte(raw), &doc); err != nil {
				return nil, usagef(i18n.T(
					"--%s は JSON オブジェクトではないので --%s と混ぜられません",
					"--%s is not a JSON object, so it cannot be combined with --%s"), flag, setFlag)
			}
		}
		docs[flag] = doc
		order = append(order, flag)
		return doc, nil
	}

	assign := func(flagName string, raw string, asJSON bool) error {
		target, path, value, err := a.splitSet(flagName, raw)
		if err != nil {
			return err
		}
		if !valid(target) {
			return usagef(i18n.T(
				"--%s %s: このコマンドに --%s はありません（リソース本体を取るフラグを指定してください）",
				"--%s %s: this command has no --%s (name a flag that takes a resource body)"),
				flagName, raw, target)
		}
		doc, err := start(target)
		if err != nil {
			return err
		}
		var v any = value
		if asJSON {
			if err := json.Unmarshal([]byte(value), &v); err != nil {
				return usagef(i18n.T(
					"--%s %s: 値が JSON として不正です",
					"--%s %s: the value is not valid JSON"), flagName, raw)
			}
		}
		return setPath(doc, path, v, flagName, raw)
	}

	for _, raw := range pairs {
		if err := assign(setFlag, raw, false); err != nil {
			return err
		}
	}
	for _, raw := range jsonPairs {
		if err := assign(setJSONFlag, raw, true); err != nil {
			return err
		}
	}

	sort.Strings(order)
	for _, flag := range order {
		payload, err := marshalNoEscape(docs[flag])
		if err != nil {
			return err
		}
		form[flag] = payload
	}
	return nil
}

// marshalNoEscape は < > & を \uXXXX に置き換えずに JSON を書きます。
//
// 既定の json.Marshal は HTML を意識して逃がしますが、ここで運ぶのは
// MT のテンプレートです。<$mt:BlogName$> のような中身が全部 \u003c に
// なると、--dry-run で目視できず、エージェントには同じ内容が数倍の
// トークンになります（内容は同じでも、読めないものを送らない）。
func marshalNoEscape(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

// splitSet は `template.name=値` を対象・項目の並び・値に分けます。
// 値は @ファイル名 と - を読み込みます（他のフラグと同じ作法です）。
func (a *App) splitSet(flagName, raw string) (target string, path []string, value string, err error) {
	left, right, ok := strings.Cut(raw, "=")
	if !ok {
		return "", nil, "", usagef(i18n.T(
			"--%s は <フラグ>.<項目>=<値> の形で指定してください: %q",
			"--%s takes <flag>.<field>=<value>: %q"), flagName, raw)
	}
	target, rest, ok := strings.Cut(left, ".")
	if !ok || target == "" || rest == "" {
		return "", nil, "", usagef(i18n.T(
			"--%s は <フラグ>.<項目>=<値> の形で指定してください: %q",
			"--%s takes <flag>.<field>=<value>: %q"), flagName, raw)
	}
	for _, part := range strings.Split(rest, ".") {
		if part == "" {
			return "", nil, "", usagef(i18n.T(
				"--%s の項目名が空です: %q", "--%s has an empty field name: %q"), flagName, raw)
		}
		path = append(path, part)
	}
	value, err = a.readValue(flagName+" "+left, right)
	return target, path, value, err
}

// setPath は入れ子をたどって値を置きます。途中に値が入っていれば
// 上書きせずに断ります（打ち間違いを黙って飲み込まないため）。
func setPath(doc map[string]any, path []string, value any, flagName, raw string) error {
	for _, key := range path[:len(path)-1] {
		next, ok := doc[key]
		if !ok {
			child := map[string]any{}
			doc[key] = child
			doc = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok {
			return usagef(i18n.T(
				"--%s %s: %s には値が入っているので、その下に項目を作れません",
				"--%s %s: %s already holds a value, so no field can be nested under it"), flagName, raw, key)
		}
		doc = child
	}
	doc[path[len(path)-1]] = value
	return nil
}

// objectParam は名前の合うリソース本体のパラメータを返します。
func objectParam(op *spec.Operation, flag string) *spec.Param {
	for _, p := range op.Body {
		if p.Flag == flag && (p.Type == "object" || p.Type == "array") {
			return p
		}
	}
	return nil
}
