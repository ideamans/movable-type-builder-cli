package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// 書き出し先の親ディレクトリが無ければ作る。
//
// templates pull --dir も --out も「まだ無い場所」に書き出すのが普通の
// 使い方で、pull は最初の1回が no such file or directory で落ちていました。
func TestDefaultEnvWriteFileCreatesParentDirectories(t *testing.T) {
	env, err := DefaultEnv()
	if err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	name := filepath.Join(base, "mt", "templates", "ページ開始.mtml")
	if err := env.WriteFile(name, []byte("<mt:Ignore>")); err != nil {
		t.Fatalf("親ディレクトリを作れていません: %v", err)
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "<mt:Ignore>" {
		t.Errorf("中身 = %q", got)
	}
}

// すでにあるディレクトリでも、これまでどおり上書きできる。
func TestDefaultEnvWriteFileOverwrites(t *testing.T) {
	env, err := DefaultEnv()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "a.txt")
	if err := env.WriteFile(name, []byte("1")); err != nil {
		t.Fatal(err)
	}
	if err := env.WriteFile(name, []byte("2")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "2" {
		t.Errorf("中身 = %q, want 2", got)
	}
}
