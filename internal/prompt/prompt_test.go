package prompt

import (
	"errors"
	"strings"
	"testing"
)

// 端末でないときに入力を求めたら、黙って待たずにエラーにする。
// エージェントや CI から実行したときに固まらないための約束です。
func TestTerminalWithoutTTY(t *testing.T) {
	term := NewTerminal(strings.NewReader("alice\n"), &strings.Builder{})
	if term.Interactive() {
		t.Fatal("パイプ入力が端末と判定されています")
	}
	if _, err := term.Line("ユーザー名", ""); !errors.Is(err, ErrNotInteractive) {
		t.Errorf("Line err = %v", err)
	}
	if _, err := term.Secret("パスワード"); !errors.Is(err, ErrNotInteractive) {
		t.Errorf("Secret err = %v", err)
	}
	if _, err := term.Confirm("消しますか", false); !errors.Is(err, ErrNotInteractive) {
		t.Errorf("Confirm err = %v", err)
	}
}

func TestParseYesNo(t *testing.T) {
	tests := []struct {
		in  string
		def bool
		out bool
	}{
		{"y", false, true},
		{"Yes", false, true},
		{"n", true, false},
		{"NO", true, false},
		{"", true, true},
		{"", false, false},
		{"わからない", false, false},
	}
	for _, tt := range tests {
		if got := parseYesNo(tt.in, tt.def); got != tt.out {
			t.Errorf("parseYesNo(%q, %v) = %v, want %v", tt.in, tt.def, got, tt.out)
		}
	}
}

func TestScripted(t *testing.T) {
	s := &Scripted{Lines: []string{"alice", ""}, Secrets: []string{"pw"}, Confirms: []bool{true}}

	if got, _ := s.Line("ユーザー名", "既定"); got != "alice" {
		t.Errorf("Line = %q", got)
	}
	// 空の答えは既定値になる
	if got, _ := s.Line("ホスト", "example.com"); got != "example.com" {
		t.Errorf("Line = %q", got)
	}
	// 答えを使い切ったら既定値
	if got, _ := s.Line("追加", "fallback"); got != "fallback" {
		t.Errorf("Line = %q", got)
	}
	if got, _ := s.Secret("パスワード"); got != "pw" {
		t.Errorf("Secret = %q", got)
	}
	if got, _ := s.Confirm("いいですか", false); !got {
		t.Error("Confirm = false")
	}
	if len(s.Asked) != 5 {
		t.Errorf("聞いた項目 = %v", s.Asked)
	}
}

func TestScriptedNotInteractive(t *testing.T) {
	s := &Scripted{NotInteractive: true}
	if s.Interactive() {
		t.Error("Interactive = true")
	}
	if _, err := s.Line("x", ""); !errors.Is(err, ErrNotInteractive) {
		t.Errorf("err = %v", err)
	}
}
