package config_test

import (
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
)

func TestParseDotEnv(t *testing.T) {
	vars := config.ParseDotEnv([]byte(`
# 接続先
MTB_PROFILE=stg
export MTB_SITE_ID=3
MTB_BASE_URL = https://example.com/cgi-bin/mt/mt-data-api.cgi
MTB_USERNAME="alice"
QUOTED='single'
EMPTY=
# 行末コメントは値の一部。パスワードの # を落とすと 401 の原因になる
MTB_PASSWORD=pa#ss word
NOEQUALS
=novalue
`))

	for _, tc := range []struct{ key, want string }{
		{"MTB_PROFILE", "stg"},
		{"MTB_SITE_ID", "3"},
		{"MTB_BASE_URL", "https://example.com/cgi-bin/mt/mt-data-api.cgi"},
		{"MTB_USERNAME", "alice"},
		{"QUOTED", "single"},
		{"EMPTY", ""},
		{"MTB_PASSWORD", "pa#ss word"},
	} {
		if got := vars[tc.key]; got != tc.want {
			t.Errorf("%s = %q, want %q", tc.key, got, tc.want)
		}
	}
	for _, key := range []string{"NOEQUALS", "", "# 接続先"} {
		if _, ok := vars[key]; ok {
			t.Errorf("%q は取り込まないはずです", key)
		}
	}
}

// プロセスの環境変数が勝つこと。MTB_PROFILE=stg と打ったものを、
// 置いてあるファイルが黙って覆したら、どこを叩いているのか分からなくなる。
func TestDotEnvDoesNotOverrideProcessEnv(t *testing.T) {
	base := config.MapEnv(map[string]string{"MTB_PROFILE": "prod"})
	env := config.DotEnv(base, map[string]string{
		"MTB_PROFILE": "stg",
		"MTB_SITE_ID": "3",
	})

	if got := env("MTB_PROFILE"); got != "prod" {
		t.Errorf("MTB_PROFILE = %q, want prod（プロセスの環境変数が勝つ）", got)
	}
	if got := env("MTB_SITE_ID"); got != "3" {
		t.Errorf("MTB_SITE_ID = %q, want 3（打っていないものは .env から）", got)
	}
	if got := env("MTB_BASE_URL"); got != "" {
		t.Errorf("どちらにも無いものは空のはずです: %q", got)
	}
}

// .env のプロファイル指定が、実際に接続先の解決まで届くこと。
func TestDotEnvProfileReachesResolve(t *testing.T) {
	env := config.DotEnv(config.MapEnv(nil), config.ParseDotEnv([]byte("MTB_PROFILE=stg\n")))
	store := config.New(config.NewMemStorage(), env)
	if err := store.Save(&config.File{
		Current: "prod",
		Profiles: map[string]*config.Profile{
			"prod": {BaseURL: "https://prod.example.com/mt-data-api.cgi"},
			"stg":  {BaseURL: "https://stg.example.com/mt-data-api.cgi"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	r, err := store.Resolve(config.Overrides{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "stg" {
		t.Errorf("プロファイル = %q, want stg（.env が current より優先）", r.Name)
	}
	// --profile は .env より強い。
	r, err = store.Resolve(config.Overrides{Profile: "prod"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "prod" {
		t.Errorf("--profile prod が効いていません: %q", r.Name)
	}
}
