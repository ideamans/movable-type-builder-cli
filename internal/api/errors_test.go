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
		// 文面に system administrator を含んでも、閉じてあるなら管理者の助言は出さない
		{403, "This endpoint (update_site_data_api) is turned off in the MtbPro plugin settings. A system administrator can turn it on in the admin screen.", "not a permission problem"},
		// サーバの設定で無効（mtb-pro のファイル送信など）。ロールの助言を出さない
		{403, `File upload is disabled. Add "MtbProAllowFileUpload 1" to mt-config.cgi to enable it.`, "mt-config.cgi"},
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
