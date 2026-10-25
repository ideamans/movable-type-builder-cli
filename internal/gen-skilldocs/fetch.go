package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// runFetch は movabletype.jp から原本を取り直します。
//
// ネットワークに出るので go generate からは呼びません。MT 側のドキュメントが
// 更新されたときに手で走らせるもの。HTML はキャッシュに残すので、
// パーサだけ直したいときは取得を飛ばせます。
func runFetch() error {
	base := "https://www.movabletype.jp/documentation/appendices"
	dir := *cache
	if dir == "" {
		var err error
		if dir, err = os.MkdirTemp("", "mtdocs-"); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "HTML キャッシュ:", dir)
	}

	for _, kind := range []string{"tags", "modifiers"} {
		indexHTML, err := get(base + "/" + kind + "/")
		if err != nil {
			return fmt.Errorf("%s の索引: %w", kind, err)
		}
		entries := parseIndex(indexHTML)
		if len(entries) == 0 {
			return fmt.Errorf("%s の索引から1件も取れませんでした（ページ構造が変わった可能性）", kind)
		}
		fmt.Fprintf(os.Stderr, "%s: 索引 %d 件\n", kind, len(entries))

		if err := os.MkdirAll(filepath.Join(dir, kind), 0o755); err != nil {
			return err
		}
		if err := fetchAllPages(base+"/"+kind, filepath.Join(dir, kind), entries); err != nil {
			return err
		}
		if err := parseAll(kind, filepath.Join(dir, kind), entries); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "原本を更新しました。go generate ./... で参照ファイルを作り直してください")
	return nil
}

type indexEntry struct {
	Name   string   `json:"name"`
	Slug   string   `json:"slug"`
	Labels []string `json:"labels"`
	Desc   string   `json:"desc"`
}

var (
	reItem   = regexp.MustCompile(`(?s)<li[^>]*>\s*<h2 class="entry-title">(.*?)</h2>\s*<p>(.*?)</p>`)
	reLink   = regexp.MustCompile(`(?s)href="[^"]*/([^"/]+)\.html"[^>]*>(.*?)</a>`)
	reLabel  = regexp.MustCompile(`(?s)<a[^>]*class="(?:block|function|re|new)[^"]*"[^>]*>(.*?)</a>`)
	reTags   = regexp.MustCompile(`<[^>]+>`)
	reSec    = `(?s)<section class="[^"]*\b%s\b[^"]*">(.*?)</section>`
	rePara   = regexp.MustCompile(`(?s)<p>(.*?)</p>`)
	rePre    = regexp.MustCompile(`(?s)<pre[^>]*>(?:<code[^>]*>)?(.*?)(?:</code>)?</pre>`)
	reDefRow = regexp.MustCompile(`(?s)<dt>(.*?)</dt>\s*<dd>(.*?)</dd>`)
	reLi     = regexp.MustCompile(`(?s)<li>(.*?)</li>`)
	reImg    = regexp.MustCompile(`<img[^>]*>`)
	reBr     = regexp.MustCompile(`<br\s*/?>`)
	reCode   = regexp.MustCompile(`</?(code|var)>`)
)

func strip(s string, keepCode bool) string {
	s = reImg.ReplaceAllString(s, "")
	s = reBr.ReplaceAllString(s, " ")
	if keepCode {
		s = reCode.ReplaceAllString(s, "`")
	}
	s = reTags.ReplaceAllString(s, "")
	return flat(html.UnescapeString(s))
}

func parseIndex(body string) []indexEntry {
	var out []indexEntry
	for _, m := range reItem.FindAllStringSubmatch(body, -1) {
		link := reLink.FindStringSubmatch(m[1])
		if link == nil {
			continue
		}
		var labels []string
		for _, l := range reLabel.FindAllStringSubmatch(m[1], -1) {
			labels = append(labels, strip(l[1], false))
		}
		out = append(out, indexEntry{
			Name: strip(link[2], false), Slug: link[1],
			Labels: labels, Desc: strip(m[2], false),
		})
	}
	return out
}

func fetchAllPages(base, dir string, entries []indexEntry) error {
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failed []string
	for _, e := range entries {
		path := filepath.Join(dir, e.Slug+".html")
		if st, err := os.Stat(path); err == nil && st.Size() > 2000 {
			continue
		}
		wg.Add(1)
		go func(slug, path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			body, err := get(base + "/" + slug + ".html")
			if err != nil {
				mu.Lock()
				failed = append(failed, slug)
				mu.Unlock()
				return
			}
			_ = os.WriteFile(path, []byte(body), 0o644)
			time.Sleep(150 * time.Millisecond)
		}(e.Slug, path)
	}
	wg.Wait()
	if len(failed) > 0 {
		sort.Strings(failed)
		return fmt.Errorf("%d 件を取得できませんでした: %s", len(failed), strings.Join(failed[:min(5, len(failed))], ", "))
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func get(url string) (string, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "ideamans-mtdocs-fetch/1.0")
	client := &http.Client{Timeout: 45 * time.Second}
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			last = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			last = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			last = fmt.Errorf("%s: %s", url, resp.Status)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		return string(body), nil
	}
	return "", last
}

func sectionOf(body, class string) string {
	re := regexp.MustCompile(fmt.Sprintf(reSec, class))
	if m := re.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return ""
}

func paragraphs(sec string, limit int) []string {
	var out []string
	for _, m := range rePara.FindAllStringSubmatch(sec, -1) {
		if p := strip(m[1], true); p != "" {
			out = append(out, p)
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func codeBlocks(sec string, limit, maxLen int) []string {
	var out []string
	for _, m := range rePre.FindAllStringSubmatch(sec, -1) {
		c := strings.TrimSpace(html.UnescapeString(reTags.ReplaceAllString(m[1], "")))
		if c != "" {
			out = append(out, trunc(c, maxLen))
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

// parseAll は取得した HTML から**事実だけ**を取り出して原本に書きます。
// 説明・詳細・使用例（マニュアルの文章）は取り出しません。要約は
// summaries.json に手で書きます（新しいタグが増えたら、ここで名前を出す）。
func parseAll(kind, dir string, entries []indexEntry) error {
	switch kind {
	case "tags":
		var rows []mtdocsTag
		for _, e := range entries {
			body, err := os.ReadFile(filepath.Join(dir, e.Slug+".html"))
			if err != nil {
				continue
			}
			h := string(body)
			t := mtdocsTag{
				Name: e.Name, Slug: e.Slug, Labels: e.Labels,
				Type: "FUNCTION", Deprecated: contains(e.Labels, "DEPRECATED"),
			}
			if contains(e.Labels, "BLOCK") {
				t.Type = "BLOCK"
			}
			for _, m := range reDefRow.FindAllStringSubmatch(sectionOf(h, "entry-modifier"), -1) {
				t.Attributes = appendAttributes(t.Attributes, strip(m[1], true))
			}
			t.Group = groupOf(t.Name)
			rows = append(rows, t)
		}
		return writeJSON(filepath.Join(*srcDir, "tags.json"), rows)
	case "modifiers":
		var rows []mtdocsMod
		for _, e := range entries {
			rows = append(rows, mtdocsMod{Name: e.Name, Slug: e.Slug})
		}
		return writeJSON(filepath.Join(*srcDir, "modifiers.json"), rows)
	}
	return nil
}

// reAttr は属性の書き方1つ（name="value" か name）です。
var (
	reAttr      = regexp.MustCompile(`^[A-Za-z_][\w:.@-]*(="[^"\n]*")?$`)
	reAttrSplit = regexp.MustCompile(`\n|,\s+([A-Za-z_][\w:.@-]*=")`)
)

// appendAttributes は見出し（dt）から属性の書き方だけを取り出して足します。
// マニュアルの見出しには、本文の説明や使用例まで入っていることがある
// （unique の見出しに段落とサンプルが続いていた）。書き方の形をしていない
// 行は捨てる。そうしないとマニュアルの文章がそのまま原本に入る。
func appendAttributes(list []string, dt string) []string {
	dt = strings.ReplaceAll(dt, "`", "")
	dt = reAttrSplit.ReplaceAllString(dt, "\n$1")
	for _, piece := range strings.Split(dt, "\n") {
		piece = strings.TrimSpace(piece)
		if piece == "" || !reAttr.MatchString(piece) || contains(list, piece) {
			continue
		}
		list = append(list, piece)
	}
	return list
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func writeJSON(path string, v any) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0o644)
}

// mtdocsTag / mtdocsMod は原本の形。internal/mtdocs の型と対応させること。
type mtdocsTag struct {
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	Type       string   `json:"type"`
	Deprecated bool     `json:"deprecated"`
	Group      string   `json:"group"`
	Labels     []string `json:"labels"`
	Attributes []string `json:"attributes"`
}

type mtdocsMod struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
