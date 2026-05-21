package seo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type LinksCheck struct{}

func (LinksCheck) ID() string       { return "seo.links" }
func (LinksCheck) Category() string { return "seo" }

func (LinksCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://developers.google.com/search/docs/crawling-indexing/links-crawlable"

	base, _ := url.Parse(ctx.FinalURL)

	var internal, external, nofollow, missingText int
	ctx.Doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" || strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") ||
			strings.HasPrefix(href, "javascript:") {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		abs := u
		if !u.IsAbs() && base != nil {
			abs = base.ResolveReference(u)
		}
		if base != nil && abs.Host == base.Host {
			internal++
		} else {
			external++
		}
		rel := strings.ToLower(s.AttrOr("rel", ""))
		if strings.Contains(rel, "nofollow") {
			nofollow++
		}
		text := strings.TrimSpace(s.Text())
		if text == "" && s.Find("img[alt]").Length() == 0 {
			missingText++
		}
	})

	info := checks.NewResult("seo.links.summary", "seo", checks.SeverityInfo,
		fmt.Sprintf("Links: %d internal, %d external, %d nofollow", internal, external, nofollow))
	info.GuidelineURL = guideline

	results := []checks.Result{info}
	if missingText > 0 {
		r := checks.NewResult("seo.links.empty_text", "seo", checks.SeverityWarning,
			fmt.Sprintf("%d link(s) have no anchor text or alt-bearing image", missingText))
		r.Recommendation = "Use descriptive anchor text so crawlers and assistive tech understand the destination."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}
