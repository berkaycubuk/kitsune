package seo

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/checks/a11y"
)

type HreflangCheck struct{}

func (HreflangCheck) ID() string       { return "seo.hreflang" }
func (HreflangCheck) Category() string { return checks.CategorySEO }

func (HreflangCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developers.google.com/search/docs/specialty/international/localized-versions"

	type entry struct{ lang, href string }
	var entries []entry
	ctx.Doc.Find(`link[rel="alternate"][hreflang]`).Each(func(_ int, s *goquery.Selection) {
		entries = append(entries, entry{
			lang: strings.TrimSpace(s.AttrOr("hreflang", "")),
			href: strings.TrimSpace(s.AttrOr("href", "")),
		})
	})
	if len(entries) == 0 {
		return nil
	}

	var results []checks.Result
	info := checks.NewResult("seo.hreflang.present", checks.CategorySEO, checks.SeverityInfo,
		fmt.Sprintf("hreflang: %d alternate(s)", len(entries)))
	info.GuidelineURL = guideline
	results = append(results, info)

	var bad []string
	hasSelfRef := false
	hasDefault := false
	for _, e := range entries {
		if e.lang == "x-default" {
			hasDefault = true
		} else if !a11y.IsBCP47(e.lang) {
			bad = append(bad, e.lang)
		}
		if e.href == ctx.FinalURL {
			hasSelfRef = true
		}
	}
	if len(bad) > 0 {
		r := checks.NewResult("seo.hreflang.invalid", checks.CategorySEO, checks.SeverityWarning,
			fmt.Sprintf("%d invalid hreflang code(s)", len(bad)))
		r.Detail = strings.Join(bad, ", ")
		r.Recommendation = "Use BCP-47 codes such as `en`, `en-US`, `zh-Hant`, or `x-default`."
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	if !hasSelfRef {
		r := checks.NewResult("seo.hreflang.no_self_reference", checks.CategorySEO, checks.SeverityWarning,
			"hreflang set has no self-reference for this page")
		r.Recommendation = "Each language version should include a hreflang link pointing to itself."
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	if !hasDefault {
		r := checks.NewResult("seo.hreflang.no_x_default", checks.CategorySEO, checks.SeverityInfo,
			"No `x-default` hreflang declared")
		r.Recommendation = "Add `<link rel=\"alternate\" hreflang=\"x-default\" href=\"…\">` as a fallback for unmatched locales."
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	return results
}
