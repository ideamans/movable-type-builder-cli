package config

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// Store は設定と認証情報の読み書きの入口です。
type Store struct {
	storage Storage
	env     Environ
}

// New は Storage と Environ を差し替えられる Store を作ります。
func New(storage Storage, env Environ) *Store {
	if env == nil {
		env = func(string) string { return "" }
	}
	return &Store{storage: storage, env: env}
}

// Open は既定の設定ディレクトリを使う Store を作ります。
func Open(env Environ, home func() (string, error)) (*Store, error) {
	dir, err := DefaultDir(env, home)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("設定ディレクトリを決められません: %w", "cannot determine the config directory: %w"), err)
	}
	return New(NewDirStorage(dir), env), nil
}

// ConfigPath は config.json の場所です（表示用）。
func (s *Store) ConfigPath() string { return s.storage.Location(ConfigFile) }

// CredentialsPath は credentials.json の場所です（表示用）。
func (s *Store) CredentialsPath() string { return s.storage.Location(CredentialsFile) }

// Load は config.json を読みます。無ければ空の設定を返します。
func (s *Store) Load() (*File, error) {
	f := &File{Profiles: map[string]*Profile{}}
	buf, err := s.storage.Read(ConfigFile)
	if isNotExist(err) {
		return f, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, f); err != nil {
		return nil, fmt.Errorf(i18n.T("%s を読めません: %w", "cannot read %s: %w"), s.ConfigPath(), err)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]*Profile{}
	}
	return f, nil
}

// Save は config.json を書きます。
func (s *Store) Save(f *File) error {
	return s.writeJSON(ConfigFile, f, 0o644)
}

// LoadCredentials は credentials.json を読みます。無ければ空を返します。
func (s *Store) LoadCredentials() (map[string]*Credential, error) {
	m := map[string]*Credential{}
	buf, err := s.storage.Read(CredentialsFile)
	if isNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil, fmt.Errorf(i18n.T("%s を読めません: %w", "cannot read %s: %w"), s.CredentialsPath(), err)
	}
	return m, nil
}

// SaveCredentials は credentials.json を 0600 で書きます。
func (s *Store) SaveCredentials(m map[string]*Credential) error {
	return s.writeJSON(CredentialsFile, m, 0o600)
}

// Credential は1プロファイル分の認証情報を返します。無ければ空を返します。
func (s *Store) Credential(profile string) (*Credential, error) {
	m, err := s.LoadCredentials()
	if err != nil {
		return nil, err
	}
	if c, ok := m[profile]; ok && c != nil {
		return c, nil
	}
	return &Credential{}, nil
}

// PutCredential は1プロファイル分の認証情報を保存します。
func (s *Store) PutCredential(profile string, c *Credential) error {
	m, err := s.LoadCredentials()
	if err != nil {
		return err
	}
	m[profile] = c
	return s.SaveCredentials(m)
}

// DeleteCredential は1プロファイル分の認証情報を消します。
func (s *Store) DeleteCredential(profile string) error {
	m, err := s.LoadCredentials()
	if err != nil {
		return err
	}
	if _, ok := m[profile]; !ok {
		return nil
	}
	delete(m, profile)
	return s.SaveCredentials(m)
}

func (s *Store) writeJSON(name string, v any, mode fs.FileMode) error {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return s.storage.Write(name, append(buf, '\n'), mode)
}

// Resolve は「フラグ → 環境変数 → プロファイル → 既定値」の順で接続先を決めます。
//
// プロファイルを名指ししたのに設定に無い場合はエラーにします。既定
// （config.json の current）が壊れているだけなら、環境変数だけで動ける
// 余地を残すため素通しします。
func (s *Store) Resolve(o Overrides) (*Resolved, error) {
	f, err := s.Load()
	if err != nil {
		return nil, err
	}

	named := firstNonEmpty(o.Profile, s.env("MTB_PROFILE"))
	name := firstNonEmpty(named, f.Current)
	r := &Resolved{Name: name}

	if name != "" {
		p, ok := f.Profiles[name]
		switch {
		case ok:
			r.FromFile = true
			r.BaseURL = p.BaseURL
			r.APIVersion = p.APIVersion
			r.ClientID = p.ClientID
			r.Username = p.Username
			r.SiteID = p.DefaultSiteID
			r.Insecure = p.InsecureSkipVerify
		case named != "":
			return nil, fmt.Errorf(i18n.T(
				"%w: %q。%s を確認するか mtb profile add %s で作ってください",
				"%w: %q. Check %s, or create it with: mtb profile add %s"),
				ErrNoSuchProfile, name, s.ConfigPath(), name)
		}
	}

	r.BaseURL = firstNonEmpty(o.BaseURL, s.env("MTB_BASE_URL"), r.BaseURL)
	r.APIVersion = firstNonEmpty(o.APIVersion, s.env("MTB_API_VERSION"), r.APIVersion, DefaultAPIVersion)
	r.ClientID = firstNonEmpty(o.ClientID, s.env("MTB_CLIENT_ID"), r.ClientID, DefaultClientID)
	r.Username = firstNonEmpty(s.env("MTB_USERNAME"), r.Username)
	r.SiteID = firstNonEmpty(o.SiteID, s.env("MTB_SITE_ID"), r.SiteID)
	if r.Name == "" {
		r.Name = "(env)"
	}

	if r.BaseURL == "" {
		return nil, fmt.Errorf(i18n.T(
			"%w。mtb profile add <名前> --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi で登録するか、MTB_BASE_URL を設定してください",
			"%w. Register one with: mtb profile add <name> --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi, or set MTB_BASE_URL"),
			ErrNoProfile)
	}
	r.BaseURL, r.APIVersion = NormalizeBaseURL(r.BaseURL, r.APIVersion)
	return r, nil
}

// Env は Store が使っている環境変数の読み取りを返します。
// 認証情報以外の環境変数（MTB_PASSWORD など）をコマンド側が読むために使います。
func (s *Store) Env(key string) string { return s.env(key) }
