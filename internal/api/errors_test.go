package api

import (
	"strings"
	"testing"
)

// Hint は MT のメッセージで出し分ける。同じ 403 / 404 でも打つ手が違う
// （実地テストで、素の Forbidden にロールの助言が出て遠回りした）。
func TestHintDependsOnTheMessage(t *testing.T) {
	cases := []struct {
		status  int
		message string
		want    string
	}{
		{403, "Forbidden", "Data API"},
		{403, "", "Data API"},
		{403, "Only a system administrator can do this.", "system administrator's profile"},
		{403, "Do not have permission to retrieve the list of templates.", "their role"},
		{404, "Unknown endpoint", "no such endpoint"},
		{404, "", "no such endpoint"},
		{404, "Site not found", "does not exist"},
	}
	for _, c := range cases {
		e := &Error{Status: c.status, Message: c.message}
		if h := e.Hint(); !strings.Contains(h, c.want) {
			t.Errorf("%d %q: Hint = %q, want it to contain %q", c.status, c.message, h, c.want)
		}
	}
}

func TestUnknownEndpoint(t *testing.T) {
	if !(&Error{Status: 404, Message: "Unknown endpoint"}).UnknownEndpoint() {
		t.Error("Unknown endpoint は UnknownEndpoint であるべき")
	}
	if (&Error{Status: 404, Message: "File not found"}).UnknownEndpoint() {
		t.Error("File not found は対象が無いだけ")
	}
	if (&Error{Status: 403, Message: "Unknown endpoint"}).UnknownEndpoint() {
		t.Error("404 以外は UnknownEndpoint ではない")
	}
}
