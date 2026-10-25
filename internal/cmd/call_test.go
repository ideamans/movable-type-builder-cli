package cmd

import (
	"strings"
	"testing"
)

// /endpoints の応答。update_thing は版違いで2つ並ぶ（新しい版を採る）。
const testEndpoints = `{"totalResults":4,"items":[
 {"id":"update_thing","route":"/sites/:site_id/things/:thing_id","verb":"PUT","version":1,"component":{"id":"thing","name":"Thing"}},
 {"id":"update_thing","route":"/sites/:site_id/things/:thing_id","verb":"PUT","version":"3","component":{"id":"thing","name":"Thing"}},
 {"id":"list_things","route":"/sites/:site_id/things","verb":"GET","version":1,"component":{"id":"thing","name":"Thing"}},
 {"id":"split_brain","route":"/a","verb":"GET","version":1},
 {"id":"split_brain","route":"/b","verb":"POST","version":2},
 {"id":"delete_thing","route":"/things/:thing_id","verb":"DELETE","version":1}]}`

func TestCallResolvesTheRouteFromTheServer(t *testing.T) {
	h := newHarness(t).reply(okJSON(testEndpoints), okJSON(`{"id":42}`))
	h.mustRun("call", "update_thing", "1", "42", "--form", "name=x", "--set", "thing.label=ラベル")

	if len(h.doer.sent) != 2 {
		t.Fatalf("送信 %d 回（/endpoints と本体の2回のはず）", len(h.doer.sent))
	}
	if got := pathOf(t, h.doer.sent[0].URL); got != "/mt/mt-data-api.cgi/v7/endpoints" {
		t.Errorf("1回目 = %s", got)
	}
	sent := h.last()
	if sent.Method != "PUT" {
		t.Errorf("method = %s（/endpoints の verb を使うはず）", sent.Method)
	}
	if got := pathOf(t, sent.URL); got != "/mt/mt-data-api.cgi/v7/sites/1/things/42" {
		t.Errorf("path = %s", got)
	}
	if !strings.Contains(sent.Body, "name=x") || !strings.Contains(sent.Body, "thing=") {
		t.Errorf("body = %s（--form と --set が mtb api と同じように効くはず）", sent.Body)
	}
}

func TestCallPathByNameBeatsPositional(t *testing.T) {
	h := newHarness(t).reply(okJSON(testEndpoints), okJSON(`{}`))
	h.mustRun("call", "update_thing", "42", "--path", "site_id=7")
	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/7/things/42" {
		t.Errorf("path = %s（名前で渡した site_id が先に埋まり、位置引数は残りに入るはず）", got)
	}
}

func TestCallUsesTheProfileDefaultSite(t *testing.T) {
	h := newHarness(t)
	if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "5")); err != nil {
		t.Fatal(err)
	}
	h.reply(okJSON(testEndpoints), okJSON(`{}`))
	h.mustRun("call", "list_things")
	if got := pathOf(t, h.last().URL); got != "/mt/mt-data-api.cgi/v7/sites/5/things" {
		t.Errorf("path = %s", got)
	}
}

func TestCallRejectsBadPathValues(t *testing.T) {
	cases := map[string][]string{
		"足りない":         {"call", "update_thing", "1"},
		"多すぎる":         {"call", "list_things", "1", "2"},
		"経路に無い名前":      {"call", "list_things", "1", "--path", "entry_id=3"},
		"名前=値 になっていない": {"call", "list_things", "--path", "site_id"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t).reply(okJSON(testEndpoints))
			err := h.run(args...)
			if ExitCode(err) != ExitUsage {
				t.Errorf("err = %v（呼び出し方の誤りのはず）", err)
			}
			if len(h.doer.sent) != 1 {
				t.Errorf("/endpoints 以外を送っています（%d 回）", len(h.doer.sent))
			}
		})
	}
}

func TestCallUnknownEndpointSuggestsSimilarOnes(t *testing.T) {
	h := newHarness(t).reply(okJSON(testEndpoints))
	err := h.run("call", "thing")
	if ExitCode(err) != ExitUsage || !strings.Contains(err.Error(), "update_thing") || !strings.Contains(err.Error(), "mtb endpoints list") {
		t.Errorf("err = %v", err)
	}
}

func TestCallRefusesAnAmbiguousEndpoint(t *testing.T) {
	h := newHarness(t).reply(okJSON(testEndpoints))
	err := h.run("call", "split_brain")
	if ExitCode(err) != ExitUsage || !strings.Contains(err.Error(), "mtb api") {
		t.Errorf("err = %v（メソッドと経路が違うものは選べない）", err)
	}
}

func TestCallDeleteNeedsConfirmation(t *testing.T) {
	h := newHarness(t).reply(okJSON(testEndpoints))
	if err := h.run("call", "delete_thing", "9"); err == nil {
		t.Fatal("確認なしで DELETE を送りました")
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("DELETE を送っています")
	}
	h = newHarness(t).reply(okJSON(testEndpoints), okJSON(`{}`))
	h.mustRun("call", "delete_thing", "9", "--yes")
	if h.last().Method != "DELETE" {
		t.Errorf("method = %s", h.last().Method)
	}
}

func TestCallDryRunShowsTheResolvedRequest(t *testing.T) {
	h := newHarness(t).reply(okJSON(testEndpoints))
	out := h.mustRun("call", "update_thing", "1", "42", "--form", "name=x", "--dry-run")
	if !strings.Contains(out, "PUT https://example.com/mt/mt-data-api.cgi/v7/sites/1/things/42") || !strings.Contains(out, "name=x") {
		t.Errorf("out = %s", out)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("--dry-run なのに本体を送っています（%d 回）", len(h.doer.sent))
	}
}

func TestCallDescribeReadsTheServerDefinition(t *testing.T) {
	openapi := `{"paths":{"/sites/{site_id}/things/{thing_id}":{"put":{"summary":"Update a thing",
	 "parameters":[{"name":"saveRevision","in":"query","description":"Save a revision"}],
	 "requestBody":{"content":{"application/x-www-form-urlencoded":{"schema":{"properties":{
	   "thing":{"$ref":"#/components/schemas/thing"},"label":{"type":"string","description":"A label"}}}}}}}}}}`
	h := newHarness(t).reply(okJSON(testEndpoints), okJSON(openapi))
	out := h.mustRun("call", "update_thing", "--describe")
	for _, want := range []string{"PUT /sites/:site_id/things/:thing_id", "Thing", "v3", "Update a thing", "site_id thing_id", "saveRevision", "label (string) — A label", "thing (object)"} {
		if !strings.Contains(out, want) {
			t.Errorf("--describe に %q がありません:\n%s", want, out)
		}
	}
	if len(h.doer.sent) != 2 || pathOf(t, h.last().URL) != "/mt/mt-data-api.cgi/v7/" {
		t.Errorf("サーバの定義（GET /）を読んでいません")
	}
}
