package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

// pageStub は totalResults 件のうち offset から limit 件を返す一覧応答を作ります。
func pageStub(total, offset, limit int) stub {
	items := []string{}
	for i := offset; i < offset+limit && i < total; i++ {
		items = append(items, fmt.Sprintf(`{"id":"%d"}`, i+1))
	}
	return okJSON(fmt.Sprintf(`{"totalResults":%d,"items":[%s]}`, total, strings.Join(items, ",")))
}

// --all は limit と offset を動かしながら全ページを取り、1つに畳む。
func TestAllFetchesEveryPage(t *testing.T) {
	const total = 261
	h := newHarness(t)
	for off := 0; off < total; off += defaultPageSize {
		h.reply(pageStub(total, off, defaultPageSize))
	}
	out := h.mustRun("sites", "entries", "list", "4", "--all")

	var got struct {
		TotalResults int               `json:"totalResults"`
		Items        []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v: %s", err, out[:200])
	}
	if got.TotalResults != total || len(got.Items) != total {
		t.Errorf("totalResults=%d items=%d, want %d", got.TotalResults, len(got.Items), total)
	}
	if len(h.doer.sent) != 3 {
		t.Errorf("送信回数 = %d, want 3", len(h.doer.sent))
	}
	// offset が正しく進んでいること
	for i, want := range []string{"0", "100", "200"} {
		u, err := url.Parse(h.doer.sent[i].URL)
		if err != nil {
			t.Fatal(err)
		}
		if u.Query().Get("offset") != want {
			t.Errorf("%d 回目の offset = %q, want %q", i+1, u.Query().Get("offset"), want)
		}
		if u.Query().Get("limit") != "100" {
			t.Errorf("%d 回目の limit = %q", i+1, u.Query().Get("limit"))
		}
	}
}

// --limit を渡したら、それが1ページの件数になる。
func TestAllUsesLimitAsPageSize(t *testing.T) {
	const total = 12
	h := newHarness(t)
	for off := 0; off < total; off += 5 {
		h.reply(pageStub(total, off, 5))
	}
	h.mustRun("sites", "entries", "list", "4", "--all", "--limit", "5")

	if len(h.doer.sent) != 3 {
		t.Fatalf("送信回数 = %d, want 3", len(h.doer.sent))
	}
	u, err := url.Parse(h.doer.sent[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("limit") != "5" {
		t.Errorf("limit = %q", u.Query().Get("limit"))
	}
}

// 途中で空のページが返ったら止まる（件数が動いても無限に回らない）。
func TestAllStopsOnEmptyPage(t *testing.T) {
	h := newHarness(t).reply(
		okJSON(`{"totalResults":999,"items":[{"id":"1"}]}`),
		okJSON(`{"totalResults":999,"items":[]}`),
	)
	out := h.mustRun("sites", "entries", "list", "4", "--all")

	if len(h.doer.sent) != 2 {
		t.Errorf("送信回数 = %d, want 2", len(h.doer.sent))
	}
	if !strings.Contains(out, `"id": "1"`) && !strings.Contains(out, `"id":"1"`) {
		t.Errorf("出力 = %s", out)
	}
}

// 他のクエリは各ページに引き継がれる。
func TestAllKeepsOtherQueryParams(t *testing.T) {
	h := newHarness(t).reply(pageStub(1, 0, 100))
	h.mustRun("sites", "entries", "list", "4", "--all", "--status", "Draft", "--sort-by", "title")

	u, err := url.Parse(h.doer.sent[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("status") != "Draft" || u.Query().Get("sortBy") != "title" {
		t.Errorf("query = %v", u.Query())
	}
}

// ページを繰れないコマンドには --all を生やさない。
func TestAllOnlyOnPaginatedCommands(t *testing.T) {
	root := testRoot(t)
	for _, tc := range []struct {
		path []string
		want bool
	}{
		{[]string{"sites", "entries", "list"}, true},
		{[]string{"sites", "list"}, true},
		{[]string{"sites", "entries", "get"}, false},
		{[]string{"sites", "entries", "create"}, false},
		{[]string{"sites", "stats", "provider"}, false},
	} {
		cmd, _, err := root.Find(tc.path)
		if err != nil {
			t.Fatalf("%v: %v", tc.path, err)
		}
		got := cmd.Flags().Lookup(allFlag) != nil
		if got != tc.want {
			t.Errorf("mtb %s の --all = %v, want %v", strings.Join(tc.path, " "), got, tc.want)
		}
	}
}

// --all の totalResults は「実際に集めた件数」。サーバの申告をそのまま
// 載せると、/search のように limit で打ち切られると 0 を返すエンドポイントで
// 取れた件数と食い違う。
func TestAllReportsCollectedCount(t *testing.T) {
	h := newHarness(t).reply(
		okJSON(`{"totalResults":0,"items":[{"id":"1"},{"id":"2"}]}`),
		okJSON(`{"totalResults":0,"items":[{"id":"3"}]}`),
		okJSON(`{"totalResults":0,"items":[]}`),
	)
	out := h.mustRun("search", "--search", "x", "--all", "--limit", "2")

	var got struct {
		TotalResults int               `json:"totalResults"`
		Items        []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.TotalResults != 3 || len(got.Items) != 3 {
		t.Errorf("totalResults=%d items=%d, want 3/3", got.TotalResults, len(got.Items))
	}
}

// 途中で打ち切ったときは、取り切れていないことを伝える。
func TestAllWarnsWhenShort(t *testing.T) {
	h := newHarness(t).reply(
		okJSON(`{"totalResults":50,"items":[{"id":"1"}]}`),
		okJSON(`{"totalResults":50,"items":[]}`),
	)
	h.mustRun("sites", "entries", "list", "4", "--all")
	if !strings.Contains(h.errOut.String(), "50") {
		t.Errorf("警告が出ていません: %q", h.errOut.String())
	}
}
