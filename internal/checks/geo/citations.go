package geo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// CitationDensityCheck — count outbound links to authoritative TLDs (.gov, .edu, .ac.*).
type CitationDensityCheck struct{}

func (CitationDensityCheck) ID() string       { return "geo.citation_density" }
func (CitationDensityCheck) Category() string { return checks.CategoryGEO }

func (CitationDensityCheck) Run(ctx *checks.PageContext) []checks.Result {
	base, _ := url.Parse(ctx.FinalURL)

	var external, authoritative int
	hosts := map[string]int{}
	ctx.Doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" || strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
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
			return
		}
		external++
		host := strings.ToLower(abs.Host)
		if isAuthoritativeHost(host) {
			authoritative++
			hosts[host]++
		}
	})

	if external == 0 {
		return nil
	}
	r := checks.NewResult("geo.citation_density.summary", checks.CategoryGEO, checks.SeverityInfo,
		fmt.Sprintf("Authoritative citations: %d of %d external link(s)", authoritative, external))
	r.Detail = "Princeton GEO research finds citations to primary sources boost AI citation rates."
	r.GuidelineURL = "https://arxiv.org/abs/2311.09735"
	return []checks.Result{r}
}

// isAuthoritativeHost flags hosts likely to count as primary/authoritative sources
// for AI citations: .gov, .edu, .mil, academic .ac.* TLDs.
func isAuthoritativeHost(host string) bool {
	switch {
	case strings.HasSuffix(host, ".gov"), strings.HasSuffix(host, ".gov.uk"),
		strings.HasSuffix(host, ".edu"), strings.HasSuffix(host, ".mil"):
		return true
	}
	// .ac.<cc> — common academic suffix (ac.uk, ac.jp, ac.kr, ac.za, ...).
	if i := strings.LastIndex(host, ".ac."); i >= 0 && i == len(host)-7 {
		return true
	}
	return false
}
