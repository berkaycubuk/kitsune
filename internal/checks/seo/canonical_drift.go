package seo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// CanonicalDriftCheck — flag mismatch between canonical and final URL.
type CanonicalDriftCheck struct{}

func (CanonicalDriftCheck) ID() string       { return "seo.canonical_drift" }
func (CanonicalDriftCheck) Category() string { return checks.CategorySEO }

func (CanonicalDriftCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developers.google.com/search/docs/crawling-indexing/canonicalization"

	href := strings.TrimSpace(ctx.Doc.Find(`head > link[rel="canonical"]`).AttrOr("href", ""))
	if href == "" {
		return nil
	}
	canon, err := url.Parse(href)
	if err != nil || !canon.IsAbs() {
		return nil // already covered by CanonicalCheck
	}
	final, err := url.Parse(ctx.FinalURL)
	if err != nil {
		return nil
	}

	var issues []string
	if !strings.EqualFold(canon.Scheme, final.Scheme) {
		issues = append(issues, fmt.Sprintf("scheme %s → %s", final.Scheme, canon.Scheme))
	}
	if !strings.EqualFold(canon.Host, final.Host) {
		issues = append(issues, fmt.Sprintf("host %s → %s", final.Host, canon.Host))
	}
	if normalizePath(canon.Path) != normalizePath(final.Path) {
		issues = append(issues, fmt.Sprintf("path %s → %s", final.Path, canon.Path))
	} else if canon.Path != final.Path {
		// Same after slash-normalisation = trailing-slash drift.
		issues = append(issues, "trailing-slash differs")
	}
	if canon.RawQuery != final.RawQuery {
		issues = append(issues, "query string differs")
	}

	if len(issues) == 0 {
		return nil
	}
	sev := checks.SeverityInfo
	for _, i := range issues {
		if strings.HasPrefix(i, "host") || strings.HasPrefix(i, "scheme") {
			sev = checks.SeverityWarning
		}
	}
	r := checks.NewResult("seo.canonical_drift.mismatch", checks.CategorySEO, sev,
		"Canonical URL differs from the final fetched URL")
	r.Detail = strings.Join(issues, "; ")
	r.Recommendation = "Make the canonical match the URL crawlers will actually reach (after redirects)."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

func normalizePath(p string) string {
	if p == "" || p == "/" {
		return "/"
	}
	return strings.TrimRight(p, "/")
}
