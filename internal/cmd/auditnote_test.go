package cmd

import (
	"net/url"
	"strings"
	"testing"
)

func TestRequestsCarryUserAgentAndAuditNote(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{"MTB_AUDIT_NOTE": "stg から本番への移行"})).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1")

	sent := h.last()
	if ua := sent.Header.Get("User-Agent"); !strings.HasPrefix(ua, "mtb/") {
		t.Errorf("User-Agent = %q（Web サーバのログで mtb を見分けられません）", ua)
	}
	raw := sent.Header.Get("X-MTB-Audit-Note")
	for _, r := range raw {
		if r > 0x7e {
			t.Fatalf("ヘッダに ASCII 以外が入っています: %q", raw)
		}
	}
	if got, err := url.PathUnescape(raw); err != nil || got != "stg から本番への移行" {
		t.Errorf("X-MTB-Audit-Note = %q → %q, %v", raw, got, err)
	}
}

func TestAuditNoteFlagBeatsTheEnvironment(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{"MTB_AUDIT_NOTE": "環境変数"})).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1", "--audit-note", "フラグ")
	if got, _ := url.PathUnescape(h.last().Header.Get("X-MTB-Audit-Note")); got != "フラグ" {
		t.Errorf("note = %q（打ったフラグが環境変数に負けています）", got)
	}
}

func TestNoAuditNoteMeansNoHeader(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1")
	if _, ok := h.last().Header["X-Mtb-Audit-Note"]; ok {
		t.Error("説明が無いのにヘッダを送っています")
	}
}
