package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/prompt"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"成功", nil, ExitOK},
		{"未認証", api.ErrNotAuthenticated, ExitAuth},
		{"401", &api.Error{Status: http.StatusUnauthorized}, ExitAuth},
		{"403", &api.Error{Status: http.StatusForbidden}, ExitForbidden},
		{"404", &api.Error{Status: http.StatusNotFound}, ExitNotFound},
		{"400", &api.Error{Status: http.StatusBadRequest}, ExitUsage},
		{"500", &api.Error{Status: http.StatusInternalServerError}, ExitServer},
		{"使い方の誤り", usagef("bad"), ExitUsage},
		{"対話できない", prompt.ErrNotInteractive, ExitUsage},
		{"接続先未設定", config.ErrNoProfile, ExitUsage},
		{"プロファイルが無い", config.ErrNoSuchProfile, ExitUsage},
		{"分類できない", errors.New("boom"), ExitError},
	}
	for _, tt := range tests {
		if got := ExitCode(tt.err); got != tt.want {
			t.Errorf("%s: ExitCode = %d, want %d", tt.name, got, tt.want)
		}
	}
}

// 包まれていても取り出せること（describeError が助言を付けて包むため）。
func TestExitCodeThroughWrapping(t *testing.T) {
	wrapped := errors.Join(errors.New("前置き"), &api.Error{Status: http.StatusForbidden})
	if got := ExitCode(wrapped); got != ExitForbidden {
		t.Errorf("ExitCode = %d, want %d", got, ExitForbidden)
	}
}
