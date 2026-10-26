package geo_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/checks/geo"
)

func TestDateSources(t *testing.T) {
	cases := []struct {
		name, head, body, want string
	}{
		{"meta tags", `<meta property="article:published_time" content="2026-01-02">`, "", "geo.date.present"},
		{"JSON-LD", `<script type="application/ld+json">{"@context":"https://schema.org","@graph":[{"@type":"BlogPosting","headline":"Hi","datePublished":"2026-01-02"}]}</script>`, "", "geo.date.present"},
		{"time element", "", `<article><time datetime="2026-01-02">2 Jan</time></article>`, "geo.date.present"},
		{"none", `<script type="application/ld+json">{"@type":"Person","name":"Ada"}</script>`, "<p>Hi</p>", "geo.date.missing"},
	}
	for _, tc := range cases {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><head>" + tc.head + "</head><body>" + tc.body + "</body></html>"))
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, r := range (geo.AuthorDateCheck{}).Run(&checks.PageContext{Doc: doc}) {
			ids = append(ids, r.ID)
		}
		if !strings.Contains(strings.Join(ids, " "), tc.want) {
			t.Errorf("%s: got %v, want %s", tc.name, ids, tc.want)
		}
	}
}
