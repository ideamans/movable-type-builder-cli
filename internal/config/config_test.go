package config_test

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
)

func newStore(t *testing.T, env map[string]string) (*config.Store, *config.MemStorage) {
	t.Helper()
	mem := config.NewMemStorage()
	return config.New(mem, config.MapEnv(env)), mem
}

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		in          string
		version     string
		wantBase    string
		wantVersion string
	}{
		{"https://example.com/mt/mt-data-api.cgi", "", "https://example.com/mt/mt-data-api.cgi", "v7"},
		{"https://example.com/mt/mt-data-api.cgi/", "", "https://example.com/mt/mt-data-api.cgi", "v7"},
		// リファレンスに載っている URL をそのまま貼っても通ること
		{"https://example.com/mt/mt-data-api.cgi/v7", "", "https://example.com/mt/mt-data-api.cgi", "v7"},
		{"https://example.com/mt/mt-data-api.cgi/v4", "v7", "https://example.com/mt/mt-data-api.cgi", "v4"},
		{"  https://example.com/a.cgi  ", "v3", "https://example.com/a.cgi", "v3"},
		// v で始まるだけのパスをバージョンと誤認しないこと
		{"https://example.com/video", "", "https://example.com/video", "v7"},
	}
	for _, tt := range tests {
		base, version := config.NormalizeBaseURL(tt.in, tt.version)
		if base != tt.wantBase || version != tt.wantVersion {
			t.Errorf("NormalizeBaseURL(%q, %q) = %q, %q; want %q, %q",
				tt.in, tt.version, base, version, tt.wantBase, tt.wantVersion)
		}
	}
}

func TestCredentialExpiry(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	tests := []struct {
		name string
		cred config.Credential
		want bool
	}{
		{"トークンが無い", config.Credential{}, true},
		{"失効時刻が不明なら使ってみる", config.Credential{AccessToken: "t"}, false},
		{"まだ余裕がある", config.Credential{AccessToken: "t", ExpiresAt: now.Add(10 * time.Minute).Unix()}, false},
		{"すでに切れている", config.Credential{AccessToken: "t", ExpiresAt: now.Add(-time.Second).Unix()}, true},
		// 送信中に切れるのを避けるため、直前は切れている扱いにする
		{"30秒後に切れるなら切れている扱い", config.Credential{AccessToken: "t", ExpiresAt: now.Add(30 * time.Second).Unix()}, true},
	}
	for _, tt := range tests {
		if got := tt.cred.Expired(now); got != tt.want {
			t.Errorf("%s: Expired = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestResolvePrecedence(t *testing.T) {
	store, _ := newStore(t, map[string]string{
		"MTB_BASE_URL": "https://env.example.com/api.cgi",
		"MTB_SITE_ID":  "9",
	})
	if err := store.Save(&config.File{
		Current: "prod",
		Profiles: map[string]*config.Profile{
			"prod": {BaseURL: "https://prod.example.com/api.cgi", DefaultSiteID: "1"},
			"stg":  {BaseURL: "https://stg.example.com/api.cgi", DefaultSiteID: "2", APIVersion: "v4"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	t.Run("環境変数はプロファイルより強い", func(t *testing.T) {
		r, err := store.Resolve(config.Overrides{})
		if err != nil {
			t.Fatal(err)
		}
		if r.BaseURL != "https://env.example.com/api.cgi" {
			t.Errorf("BaseURL = %q", r.BaseURL)
		}
		if r.SiteID != "9" {
			t.Errorf("SiteID = %q, want 9", r.SiteID)
		}
		if r.Name != "prod" {
			t.Errorf("Name = %q, want prod", r.Name)
		}
	})

	t.Run("フラグは環境変数より強い", func(t *testing.T) {
		r, err := store.Resolve(config.Overrides{BaseURL: "https://flag.example.com/api.cgi", SiteID: "5"})
		if err != nil {
			t.Fatal(err)
		}
		if r.BaseURL != "https://flag.example.com/api.cgi" || r.SiteID != "5" {
			t.Errorf("解決結果が違います: %+v", r)
		}
	})

	t.Run("プロファイル指定で切り替わる", func(t *testing.T) {
		bare, _ := newStore(t, nil)
		if err := bare.Save(&config.File{
			Current: "prod",
			Profiles: map[string]*config.Profile{
				"prod": {BaseURL: "https://prod.example.com/api.cgi"},
				"stg":  {BaseURL: "https://stg.example.com/api.cgi", APIVersion: "v4"},
			},
		}); err != nil {
			t.Fatal(err)
		}
		r, err := bare.Resolve(config.Overrides{Profile: "stg"})
		if err != nil {
			t.Fatal(err)
		}
		if r.BaseURL != "https://stg.example.com/api.cgi" || r.APIVersion != "v4" {
			t.Errorf("解決結果が違います: %+v", r)
		}
	})
}

func TestResolveErrors(t *testing.T) {
	t.Run("接続先がどこにも無い", func(t *testing.T) {
		store, _ := newStore(t, nil)
		_, err := store.Resolve(config.Overrides{})
		if !errors.Is(err, config.ErrNoProfile) {
			t.Fatalf("err = %v, want ErrNoProfile", err)
		}
	})

	t.Run("名指ししたプロファイルが無い", func(t *testing.T) {
		store, _ := newStore(t, nil)
		_, err := store.Resolve(config.Overrides{Profile: "nope"})
		if !errors.Is(err, config.ErrNoSuchProfile) {
			t.Fatalf("err = %v, want ErrNoSuchProfile", err)
		}
	})

	// current が壊れていても、環境変数だけで動ける余地は残す
	t.Run("current が壊れていても環境変数で動く", func(t *testing.T) {
		store, _ := newStore(t, map[string]string{"MTB_BASE_URL": "https://env.example.com/api.cgi"})
		if err := store.Save(&config.File{Current: "消えたプロファイル"}); err != nil {
			t.Fatal(err)
		}
		r, err := store.Resolve(config.Overrides{})
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if r.FromFile {
			t.Error("FromFile が true になっています")
		}
	})
}

func TestCredentialsAreWrittenPrivately(t *testing.T) {
	store, mem := newStore(t, nil)
	cred := &config.Credential{SessionID: "s", AccessToken: "a", Username: "alice"}
	if err := store.PutCredential("prod", cred); err != nil {
		t.Fatal(err)
	}

	mode, ok := mem.Mode(config.CredentialsFile)
	if !ok {
		t.Fatal("credentials.json が書かれていません")
	}
	if mode != fs.FileMode(0o600) {
		t.Errorf("credentials.json のパーミッションが %v です。0600 であるべきです", mode)
	}

	got, err := store.Credential("prod")
	if err != nil {
		t.Fatal(err)
	}
	if got.SessionID != "s" || got.Username != "alice" {
		t.Errorf("読み戻せていません: %+v", got)
	}

	// 認証情報が config.json 側に漏れていないこと
	buf, err := mem.Read(config.ConfigFile)
	if err == nil && strings.Contains(string(buf), "sessionId") {
		t.Error("config.json にセッションIDが書かれています")
	}
}

func TestDeleteCredential(t *testing.T) {
	store, _ := newStore(t, nil)
	if err := store.PutCredential("a", &config.Credential{AccessToken: "x"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutCredential("b", &config.Credential{AccessToken: "y"}); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteCredential("a"); err != nil {
		t.Fatal(err)
	}
	m, err := store.LoadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m["a"]; ok {
		t.Error("a が残っています")
	}
	if _, ok := m["b"]; !ok {
		t.Error("b まで消えています")
	}
	// 無いものを消してもエラーにしない
	if err := store.DeleteCredential("missing"); err != nil {
		t.Errorf("存在しないプロファイルの削除でエラー: %v", err)
	}
}

func TestLoadOnMissingFilesReturnsEmpty(t *testing.T) {
	store, _ := newStore(t, nil)
	f, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(f.Profiles) != 0 {
		t.Errorf("Profiles = %v", f.Profiles)
	}
	creds, err := store.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if len(creds) != 0 {
		t.Errorf("creds = %v", creds)
	}
}

func TestDefaultDir(t *testing.T) {
	home := func() (string, error) { return "/home/tester", nil }
	tests := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"MTB_CONFIG_DIR": "/tmp/x"}, "/tmp/x"},
		{map[string]string{"XDG_CONFIG_HOME": "/cfg"}, "/cfg/mtb"},
		{nil, "/home/tester/.config/mtb"},
	}
	for _, tt := range tests {
		got, err := config.DefaultDir(config.MapEnv(tt.env), home)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("DefaultDir(%v) = %q, want %q", tt.env, got, tt.want)
		}
	}
}
