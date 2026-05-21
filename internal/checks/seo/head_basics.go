package seo

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type ViewportCheck struct{}

func (ViewportCheck) ID() string       { return "seo.viewport" }
func (ViewportCheck) Category() string { return "seo" }

func (ViewportCheck) Run(ctx *checks.PageContext) []checks.Result {
	content := strings.TrimSpace(ctx.Doc.Find(`head > meta[name="viewport"]`).AttrOr("content", ""))
	guideline := "https://developers.google.com/search/docs/crawling-indexing/mobile/mobile-sites-mobile-first-indexing"

	if content == "" {
		r := checks.NewResult("seo.viewport.missing", "seo", checks.SeverityWarning,
			"Missing viewport meta tag")
		r.Recommendation = `Add <meta name="viewport" content="width=device-width, initial-scale=1"> for mobile usability.`
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("seo.viewport.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Viewport: %s", content))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

type LangCheck struct{}

func (LangCheck) ID() string       { return "seo.lang" }
func (LangCheck) Category() string { return "seo" }

func (LangCheck) Run(ctx *checks.PageContext) []checks.Result {
	lang := strings.TrimSpace(ctx.Doc.Find("html").AttrOr("lang", ""))
	guideline := "https://www.w3.org/International/questions/qa-html-language-declarations"

	if lang == "" {
		r := checks.NewResult("seo.lang.missing", "seo", checks.SeverityWarning,
			"Missing lang attribute on <html>")
		r.Recommendation = `Add lang="en" (or appropriate BCP-47 code) to the <html> element.`
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("seo.lang.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Document language: %s", lang))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
