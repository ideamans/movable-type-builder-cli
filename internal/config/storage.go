package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Storage は設定ディレクトリ1つ分のファイル読み書きです。
// 実体はディレクトリですが、テストではメモリに差し替えます。
type Storage interface {
	// Read はファイルを読みます。無いときは fs.ErrNotExist を返します。
	Read(name string) ([]byte, error)
	// Write はファイルを書きます。途中で落ちても壊れた設定が残らないよう、
	// 実装は原子的に置き換えることが期待されます。
	Write(name string, data []byte, mode fs.FileMode) error
	// Location は人に見せるための場所（パスなど）を返します。
	Location(name string) string
}

// DirStorage はディレクトリ上の実装です。
type DirStorage struct {
	Dir string
}

// NewDirStorage はディレクトリ上の Storage を作ります。
func NewDirStorage(dir string) *DirStorage { return &DirStorage{Dir: dir} }

// Read はファイルを読みます。
func (s *DirStorage) Read(name string) ([]byte, error) {
	return os.ReadFile(s.Location(name))
}

// Write は一時ファイルに書いてから rename します。書き込み中に中断しても
// 元のファイルが壊れないようにするためです。
func (s *DirStorage) Write(name string, data []byte, mode fs.FileMode) error {
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	path := s.Location(name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	// WriteFile の mode は既存ファイルには効かないので明示的に締め直します。
	if err := os.Chmod(tmp, mode); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Location はファイルのフルパスです。
func (s *DirStorage) Location(name string) string { return filepath.Join(s.Dir, name) }

// MemStorage はメモリ上の実装です。テスト専用ですが、実装が分かれると
// 挙動もずれるため本体と同じパッケージに置いています。
type MemStorage struct {
	mu    sync.Mutex
	files map[string][]byte
	modes map[string]fs.FileMode
}

// NewMemStorage は空のメモリストレージを作ります。
func NewMemStorage() *MemStorage {
	return &MemStorage{files: map[string][]byte{}, modes: map[string]fs.FileMode{}}
}

// Read はファイルを読みます。
func (s *MemStorage) Read(name string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	buf, ok := s.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	return append([]byte(nil), buf...), nil
}

// Write はファイルを書きます。
func (s *MemStorage) Write(name string, data []byte, mode fs.FileMode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[name] = append([]byte(nil), data...)
	s.modes[name] = mode
	return nil
}

// Location は人に見せる場所です。
func (s *MemStorage) Location(name string) string { return "mem:" + name }

// Mode は書き込み時に指定されたパーミッションを返します。認証情報が
// 0600 で保存されていることをテストが確かめるために使います。
func (s *MemStorage) Mode(name string) (fs.FileMode, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.modes[name]
	return m, ok
}

// Names は保存済みのファイル名を名前順で返します。
func (s *MemStorage) Names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.files))
	for name := range s.files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Environ は環境変数の読み取りです。
type Environ func(key string) string

// OSEnv はプロセスの環境変数を読みます。
func OSEnv(key string) string { return os.Getenv(key) }

// MapEnv は固定の値を返す Environ を作ります。テスト用です。
func MapEnv(m map[string]string) Environ {
	return func(key string) string { return m[key] }
}

// DefaultDir は設定ディレクトリを決めます。優先順位は
// MTB_CONFIG_DIR → XDG_CONFIG_HOME/mtb → ~/.config/mtb です。
func DefaultDir(env Environ, home func() (string, error)) (string, error) {
	if dir := env("MTB_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	if dir := env("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "mtb"), nil
	}
	h, err := home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".config", "mtb"), nil
}

func isNotExist(err error) bool { return errors.Is(err, fs.ErrNotExist) }
