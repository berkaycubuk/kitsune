package seo_test

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/checks/seo"
)

func loadFixture(t *testing.T, name string) *checks.PageContext {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return &checks.PageContext{
		RequestedURL: "https://example.com/widgets",
		FinalURL:     "https://example.com/widgets",
		StatusCode:   200,
		Headers:      http.Header{},
		HTML:         body,
		Doc:          doc,
	}
}

func resultIDs(rs []checks.Result) map[string]checks.Severity {
	out := map[string]checks.Severity{}
	for _, r := range rs {
		out[r.ID] = r.Severity
	}
	return out
}

func TestGoodPage(t *testing.T) {
	ctx := loadFixture(t, "good.html")

	cases := []struct {
		c       checks.Check
		wantID  string
		wantSev checks.Severity
	}{
		{seo.TitleCheck{}, "seo.title.present", checks.SeverityInfo},
		{seo.DescriptionCheck{}, "seo.description.present", checks.SeverityInfo},
		{seo.CanonicalCheck{}, "seo.canonical.present", checks.SeverityInfo},
		{seo.LangCheck{}, "seo.lang.present", checks.SeverityInfo},
		{seo.ViewportCheck{}, "seo.viewport.present", checks.SeverityInfo},
		{seo.OpenGraphCheck{}, "seo.open_graph.present", checks.SeverityInfo},
		{seo.HeadingsCheck{}, "seo.headings.h1_present", checks.SeverityInfo},
		{seo.ImageAltCheck{}, "seo.image_alt.ok", checks.SeverityInfo},
	}
	for _, tc := range cases {
		t.Run(tc.c.ID(), func(t *testing.T) {
			got := resultIDs(tc.c.Run(ctx))
			sev, ok := got[tc.wantID]
			if !ok {
				t.Fatalf("expected result id %q in %v", tc.wantID, got)
			}
			if sev != tc.wantSev {
				t.Errorf("severity for %q: got %s, want %s", tc.wantID, sev, tc.wantSev)
			}
		})
	}
}

func TestEmptyDocument(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader([]byte("<html><body></body></html>")))
	ctx := &checks.PageContext{
		RequestedURL: "https://example.com",
		FinalURL:     "https://example.com",
		StatusCode:   200,
		Headers:      http.Header{},
		Doc:          doc,
	}
	rs := seo.TitleCheck{}.Run(ctx)
	if len(rs) != 1 || rs[0].ID != "seo.title.missing" {
		t.Fatalf("expected seo.title.missing, got %+v", rs)
	}
	rs = seo.HeadingsCheck{}.Run(ctx)
	got := resultIDs(rs)
	if _, ok := got["seo.headings.h1_missing"]; !ok {
		t.Fatalf("expected seo.headings.h1_missing, got %v", got)
	}
}
