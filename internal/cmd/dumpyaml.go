package cmd

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ダンプの YAML は「差分を読むため」だけに出します。人が編集して戻すことは
// 想定していません。だから見た目より**同じ入力なら必ず同じバイト列**が出ることを
// 優先しています。
//
// 決めごとは3つ。
//
//   - **キーは名前順**。API が返す順は保証されないので、そのまま出すと環境間で
//     並びが揺れて、中身が同じでも差分が出ます。
//   - **複数行の文字列はリテラルブロック（`|`）**。テンプレート本文は 42 件中
//     40 件が複数行で、最長 1566 行あります。1行の引用符付き文字列にすると
//     差分が「1行まるごと変わった」としか出ず、読めません。
//   - **タブと行末の空白を含む文字列は引用**。YAML のブロックスカラーはこれらを
//     復元できないので、リテラルにすると値が変わってしまいます。
//
// yaml.v3 を使うのは、この3つ目のような細かい取り決めを自前で間違えないためです。

// yamlBytes は JSON 由来の値を決定的な YAML にします。
func yamlBytes(v any) ([]byte, error) {
	node, err := yamlNode(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// yamlNode は値を yaml.Node に写します。マップは名前順に並べます。
func yamlNode(v any) (*yaml.Node, error) {
	switch t := v.(type) {
	case map[string]any:
		n := &yaml.Node{Kind: yaml.MappingNode}
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child, err := yamlNode(t[k])
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, child)
		}
		return n, nil
	case []any:
		n := &yaml.Node{Kind: yaml.SequenceNode}
		for _, item := range t {
			child, err := yamlNode(item)
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, child)
		}
		return n, nil
	case string:
		return stringNode(t), nil
	case nil:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	case json.Number:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: t.String()}, nil
	}
	// bool と数値はそのまま任せます。
	n := &yaml.Node{}
	if err := n.Encode(v); err != nil {
		return nil, err
	}
	return n, nil
}

// stringNode は文字列の書き方を決めます。
func stringNode(s string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
	if strings.Contains(s, "\n") && canBeLiteral(s) {
		n.Style = yaml.LiteralStyle
	}
	return n
}

// canBeLiteral はリテラルブロックで書いても値が変わらないかを返します。
//
// ブロックスカラーは行末の空白とタブによる字下げを復元できません。
// テンプレート本文にはどちらも普通に現れるので、ここを見ないと
// 「ダンプすると中身が変わる」という一番まずい壊れ方をします。
func canBeLiteral(s string) bool {
	if strings.Contains(s, "\t") {
		return false
	}
	for _, line := range strings.Split(s, "\n") {
		if line != strings.TrimRight(line, " ") {
			return false
		}
		if strings.HasPrefix(line, " ") && strings.TrimSpace(line) == "" {
			return false
		}
	}
	return true
}
