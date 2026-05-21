package seo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type CanonicalCheck struct{}

func (CanonicalCheck) ID() string       { return "seo.canonical" }
func (CanonicalCheck) Category() string { return "seo" }

func (CanonicalCheck) Run(ctx *checks.PageContext) []checks.Result {
	href := strings.TrimSpace(ctx.Doc.Find(`head > link[rel="canonical"]`).AttrOr("href", ""))
	guideline := "https://developers.google.com/search/docs/crawling-indexing/canonicalization"

	if href == "" {
		r := checks.NewResult("seo.canonical.missing", "seo", checks.SeverityWarning,
			"Missing canonical link")
		r.Detail = `No <link rel="canonical"> found in the document head.`
		r.Recommendation = "Add a self-referential canonical to prevent duplicate-content ambiguity."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	u, err := url.Parse(href)
	if err != nil || !u.IsAbs() {
		r := checks.NewResult("seo.canonical.relative", "seo", checks.SeverityWarning,
			"Canonical URL is not absolute")
		r.Detail = fmt.Sprintf("href=%q", href)
		r.Recommendation = "Use a fully-qualified absolute URL for the canonical link."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	info := checks.NewResult("seo.canonical.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Canonical: %s", href))
	info.GuidelineURL = guideline
	return []checks.Result{info}
}
