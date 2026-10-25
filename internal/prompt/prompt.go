// Package prompt は対話入力を抽象化します。
//
// 端末は外部依存の中でもテストしづらい部類なので、Prompter インタフェースに
// 閉じ込めてあります。テストは Scripted に答えを積んで渡します。
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// ErrNotInteractive は対話できない場面で入力を求められたときのエラーです。
// 「エージェントが実行して黙って止まる」を避けるため、フラグや環境変数で
// 与えるべきだと分かる文言にしてあります。
var ErrNotInteractive = i18n.Error("対話入力ができません", "cannot read interactive input")

// Prompter は対話入力です。
type Prompter interface {
	// Line は1行を読みます。def は空入力のときの既定値です。
	Line(label, def string) (string, error)
	// Secret はエコーせずに1行を読みます。
	Secret(label string) (string, error)
	// Confirm は yes/no を読みます。
	Confirm(label string, def bool) (bool, error)
	// Interactive は対話できる状況かを返します。
	Interactive() bool
}

// Terminal は実際の端末から読む実装です。
type Terminal struct {
	In  io.Reader
	Out io.Writer
	// Fd は端末判定とエコー抑止に使うファイルディスクリプタです。
	Fd int

	reader *bufio.Reader
}

// NewTerminal は標準入出力を使う Prompter を作ります。
func NewTerminal(in io.Reader, out io.Writer) *Terminal {
	fd := -1
	if f, ok := in.(*os.File); ok {
		fd = int(f.Fd())
	}
	return &Terminal{In: in, Out: out, Fd: fd}
}

// Interactive は入力が端末かどうかです。
func (t *Terminal) Interactive() bool { return t.Fd >= 0 && term.IsTerminal(t.Fd) }

// Line は1行を読みます。
func (t *Terminal) Line(label, def string) (string, error) {
	if !t.Interactive() {
		return "", fmt.Errorf("%w: %s", ErrNotInteractive, label)
	}
	if def != "" {
		fmt.Fprintf(t.Out, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(t.Out, "%s: ", label)
	}
	if t.reader == nil {
		t.reader = bufio.NewReader(t.In)
	}
	line, err := t.reader.ReadString('\n')
	if err != nil && (line == "" || !errors.Is(err, io.EOF)) {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

// Secret はエコーせずに1行を読みます。
func (t *Terminal) Secret(label string) (string, error) {
	if !t.Interactive() {
		return "", fmt.Errorf("%w: %s", ErrNotInteractive, label)
	}
	fmt.Fprintf(t.Out, "%s: ", label)
	buf, err := term.ReadPassword(t.Fd)
	fmt.Fprintln(t.Out)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf)), nil
}

// Confirm は yes/no を読みます。
func (t *Terminal) Confirm(label string, def bool) (bool, error) {
	hint := "y/N"
	if def {
		hint = "Y/n"
	}
	answer, err := t.Line(fmt.Sprintf("%s (%s)", label, hint), "")
	if err != nil {
		return false, err
	}
	return parseYesNo(answer, def), nil
}

// Scripted は答えをあらかじめ積んでおく実装です。テスト用です。
type Scripted struct {
	// Lines は Line が順に返す答えです。
	Lines []string
	// Secrets は Secret が順に返す答えです。
	Secrets []string
	// Confirms は Confirm が順に返す答えです。
	Confirms []bool
	// NotInteractive を立てると Interactive が false になり、入力要求は
	// ErrNotInteractive になります。
	NotInteractive bool
	// Asked は聞かれたラベルの記録です。
	Asked []string
}

// Interactive は対話できるかを返します。
func (s *Scripted) Interactive() bool { return !s.NotInteractive }

// Line は積んである答えを順に返します。
func (s *Scripted) Line(label, def string) (string, error) {
	s.Asked = append(s.Asked, label)
	if s.NotInteractive {
		return "", fmt.Errorf("%w: %s", ErrNotInteractive, label)
	}
	if len(s.Lines) == 0 {
		return def, nil
	}
	answer := s.Lines[0]
	s.Lines = s.Lines[1:]
	if answer == "" {
		return def, nil
	}
	return answer, nil
}

// Secret は積んである答えを順に返します。
func (s *Scripted) Secret(label string) (string, error) {
	s.Asked = append(s.Asked, label)
	if s.NotInteractive {
		return "", fmt.Errorf("%w: %s", ErrNotInteractive, label)
	}
	if len(s.Secrets) == 0 {
		return "", nil
	}
	answer := s.Secrets[0]
	s.Secrets = s.Secrets[1:]
	return answer, nil
}

// Confirm は積んである答えを順に返します。
func (s *Scripted) Confirm(label string, def bool) (bool, error) {
	s.Asked = append(s.Asked, label)
	if s.NotInteractive {
		return false, fmt.Errorf("%w: %s", ErrNotInteractive, label)
	}
	if len(s.Confirms) == 0 {
		return def, nil
	}
	answer := s.Confirms[0]
	s.Confirms = s.Confirms[1:]
	return answer, nil
}

func parseYesNo(s string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return def
	}
}
