package seo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type HTTPSCheck struct{}

func (HTTPSCheck) ID() string       { return "seo.https" }
func (HTTPSCheck) Category() string { return "seo" }

func (HTTPSCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://developers.google.com/search/docs/crawling-indexing/site-move-with-url-changes"

	u, err := url.Parse(ctx.FinalURL)
	if err != nil {
		return nil
	}

	var results []checks.Result
	if u.Scheme != "https" {
		r := checks.NewResult("seo.https.not_https", "seo", checks.SeverityError,
			"Page is served over HTTP, not HTTPS")
		r.Recommendation = "Serve the page over HTTPS; modern browsers and search engines penalize HTTP."
		r.GuidelineURL = guideline
		results = append(results, r)
		return results
	}

	// Mixed content: page is HTTPS but references http:// resources.
	mixed := 0
	selectors := []struct{ sel, attr string }{
		{"script[src]", "src"},
		{"img[src]", "src"},
		{"link[href]", "href"},
		{"iframe[src]", "src"},
	}
	for _, s := range selectors {
		ctx.Doc.Find(s.sel).Each(func(_ int, sel *goquery.Selection) {
			v := strings.TrimSpace(sel.AttrOr(s.attr, ""))
			if strings.HasPrefix(strings.ToLower(v), "http://") {
				mixed++
			}
		})
	}
	if mixed > 0 {
		r := checks.NewResult("seo.https.mixed_content", "seo", checks.SeverityError,
			fmt.Sprintf("%d resource(s) loaded over HTTP from an HTTPS page", mixed))
		r.Recommendation = "Update resource URLs to https:// to avoid mixed-content blocking."
		r.GuidelineURL = "https://developer.mozilla.org/en-US/docs/Web/Security/Mixed_content"
		results = append(results, r)
	}

	if len(results) == 0 {
		r := checks.NewResult("seo.https.ok", "seo", checks.SeverityInfo,
			"Page served over HTTPS with no mixed content detected")
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	return results
}

type StatusCheck struct{}

func (StatusCheck) ID() string       { return "seo.status" }
func (StatusCheck) Category() string { return "seo" }

func (StatusCheck) Run(ctx *checks.PageContext) []checks.Result {
	var results []checks.Result
	sev := checks.SeverityInfo
	if ctx.StatusCode >= 400 {
		sev = checks.SeverityError
	} else if ctx.StatusCode >= 300 {
		sev = checks.SeverityWarning
	}
	r := checks.NewResult("seo.status.code", "seo", sev,
		fmt.Sprintf("HTTP %d (final URL: %s)", ctx.StatusCode, ctx.FinalURL))
	results = append(results, r)

	if n := len(ctx.Redirects); n > 0 {
		sev := checks.SeverityInfo
		if n >= 3 {
			sev = checks.SeverityWarning
		}
		r := checks.NewResult("seo.status.redirects", "seo", sev,
			fmt.Sprintf("%d redirect(s) followed", n))
		if n >= 3 {
			r.Recommendation = "Long redirect chains slow crawlers; collapse where possible."
		}
		results = append(results, r)
	}
	return results
}
