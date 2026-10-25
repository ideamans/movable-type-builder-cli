package i18n_test

import (
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		explicit string
		vars     map[string]string
		want     i18n.Lang
	}{
		{"明示指定が最優先", "ja", map[string]string{"LANG": "en_US.UTF-8"}, i18n.JA},
		{"MTB_LANG", "", map[string]string{"MTB_LANG": "ja", "LANG": "en_US.UTF-8"}, i18n.JA},
		{"LC_ALL は LANG より強い", "", map[string]string{"LC_ALL": "ja_JP.UTF-8", "LANG": "en_US"}, i18n.JA},
		{"LC_MESSAGES", "", map[string]string{"LC_MESSAGES": "ja_JP.UTF-8"}, i18n.JA},
		{"LANG", "", map[string]string{"LANG": "ja_JP.UTF-8"}, i18n.JA},
		{"英語ロケール", "", map[string]string{"LANG": "en_US.UTF-8"}, i18n.EN},
		{"何も無ければ英語", "", nil, i18n.EN},
		// C / POSIX は「指定なし」。次の候補に譲る
		{"C は無視して次を見る", "", map[string]string{"LC_ALL": "C", "LANG": "ja_JP.UTF-8"}, i18n.JA},
		{"POSIX も同じ", "", map[string]string{"LC_ALL": "POSIX", "LANG": "ja_JP.UTF-8"}, i18n.JA},
		{"全部 C なら英語", "", map[string]string{"LC_ALL": "C", "LANG": "C"}, i18n.EN},
		// 表記ゆれ
		{"ハイフン区切り", "", map[string]string{"LANG": "ja-JP"}, i18n.JA},
		{"言語だけ", "", map[string]string{"LANG": "ja"}, i18n.JA},
		{"大文字", "", map[string]string{"LANG": "JA_JP.UTF-8"}, i18n.JA},
		{"@ 付き", "", map[string]string{"LANG": "ja_JP@cjknarrow"}, i18n.JA},
		// jp は国コードだが --lang jp と打つ人は多い。黙って英語になると、
		// なぜ日本語にならないのか分からないまま終わる。
		{"--lang jp", "jp", nil, i18n.JA},
		{"--lang japanese", "japanese", nil, i18n.JA},
		{"--lang english", "english", nil, i18n.EN},
		// 知らない言語は英語（未訳の日本語より読める）
		{"知らない言語は英語", "", map[string]string{"LANG": "fr_FR.UTF-8"}, i18n.EN},
	}
	for _, tt := range tests {
		if got := i18n.Detect(tt.explicit, env(tt.vars)); got != tt.want {
			t.Errorf("%s: Detect = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestT(t *testing.T) {
	defer i18n.SetLang(i18n.EN)

	i18n.SetLang(i18n.JA)
	if got := i18n.T("こんにちは", "hello"); got != "こんにちは" {
		t.Errorf("ja: %q", got)
	}
	i18n.SetLang(i18n.EN)
	if got := i18n.T("こんにちは", "hello"); got != "hello" {
		t.Errorf("en: %q", got)
	}
}

// 既定は英語。読み手を日本語話者だと決めつけない。
func TestDefaultIsEnglish(t *testing.T) {
	if i18n.Detect("", nil) != i18n.EN {
		t.Error("既定が英語ではありません")
	}
}
