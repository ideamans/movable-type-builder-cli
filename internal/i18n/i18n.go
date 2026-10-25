// Package i18n は、利用者に見せる文言を日本語と英語で切り替えます。
//
// 対訳表を別ファイルに持たず、呼び出し箇所に日英を並べて書く形にしています。
//
//	fmt.Fprintln(out, i18n.T("サインインしました", "Signed in."))
//
// 表を持つとキーと本文がずれても気づけませんが、この形ならずれようがなく、
// その場で両方を読めます。文言は 160 ほどなので、表を管理する手間の方が高くつきます。
//
// **LLM 向けの出力（mtb llm の埋め込みリファレンス）は対象外**です。
// あちらは読み手が言語を選ばないので、日本語のままにしてあります。
package i18n

import (
	"strings"
	"sync"
)

// Lang は表示に使う言語です。
type Lang string

const (
	// JA は日本語です。
	JA Lang = "ja"
	// EN は英語です。判別できないときの既定でもあります。
	EN Lang = "en"
)

var (
	mu      sync.RWMutex
	current = EN
)

// SetLang は表示言語を決めます。
func SetLang(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	current = l
}

// Current はいまの表示言語です。
func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// T は表示言語に合わせて日本語か英語を返します。
func T(ja, en string) string {
	if Current() == JA {
		return ja
	}
	return en
}

// Detect は環境から表示言語を決めます。優先順位は
// 明示指定 → MTB_LANG → LC_ALL → LC_MESSAGES → LANG です。
//
// 判別できないときは英語にします。CLI の出力を読む相手が日本語話者だと
// 決めつけない方が安全で、日本語がほしい人は MTB_LANG=ja と書けます。
func Detect(explicit string, getenv func(string) string) Lang {
	candidates := []string{explicit}
	if getenv != nil {
		for _, key := range []string{"MTB_LANG", "LC_ALL", "LC_MESSAGES", "LANG"} {
			candidates = append(candidates, getenv(key))
		}
	}
	for _, v := range candidates {
		if l, ok := parse(v); ok {
			return l
		}
	}
	return EN
}

// parse は "ja_JP.UTF-8" のようなロケール表記から言語を取り出します。
// 空や C / POSIX は「指定なし」として扱い、次の候補に譲ります。
func parse(v string) (Lang, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	switch strings.ToUpper(v) {
	case "C", "POSIX", "C.UTF-8":
		return "", false
	}
	lower := strings.ToLower(v)
	if i := strings.IndexAny(lower, "._@"); i >= 0 {
		lower = lower[:i]
	}
	lower = strings.ReplaceAll(lower, "-", "_")
	switch {
	// jp は国コードで言語コードではないが、--lang jp と打つ人は多い。
	// 黙って英語になると、なぜ日本語にならないのか分からない。
	case lower == "ja" || lower == "jp" || lower == "japanese" || strings.HasPrefix(lower, "ja_"):
		return JA, true
	case lower == "en" || lower == "english" || strings.HasPrefix(lower, "en_"):
		return EN, true
	}
	// 知らない言語は英語にします（未訳の日本語を出すより読める）。
	return EN, true
}

// Error は、文言を「表示するとき」に解決するエラーを作ります。
//
// パッケージ変数のエラーを T() で作ると、言語を決める前（init 時）に
// 文言が固まってしまいます。errors.Is で比べる相手は値そのものなので、
// 文言を後から解決してもエラーの同一性は変わりません。
func Error(ja, en string) error { return &lazyError{ja: ja, en: en} }

type lazyError struct{ ja, en string }

func (e *lazyError) Error() string { return T(e.ja, e.en) }
