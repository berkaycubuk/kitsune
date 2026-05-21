package seo

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/temoto/robotstxt"
)

type RobotsTxtCheck struct{}

func (RobotsTxtCheck) ID() string       { return "seo.robots_txt" }
func (RobotsTxtCheck) Category() string { return "seo" }

func (RobotsTxtCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://developers.google.com/search/docs/crawling-indexing/robots/intro"
	r := ctx.Robots

	if r == nil || r.FetchErr != "" {
		res := checks.NewResult("seo.robots_txt.unreachable", "seo", checks.SeverityWarning,
			"robots.txt could not be fetched")
		if r != nil {
			res.Detail = r.FetchErr
		}
		res.GuidelineURL = guideline
		return []checks.Result{res}
	}
	if !r.Found() {
		res := checks.NewResult("seo.robots_txt.missing", "seo", checks.SeverityWarning,
			fmt.Sprintf("robots.txt returned HTTP %d", r.StatusCode))
		res.Recommendation = "Add a robots.txt at the site root to control crawling."
		res.GuidelineURL = guideline
		return []checks.Result{res}
	}

	results := []checks.Result{}

	parsed, perr := robotstxt.FromBytes(r.Body)
	if perr != nil {
		res := checks.NewResult("seo.robots_txt.parse_error", "seo", checks.SeverityWarning,
			"robots.txt failed to parse")
		res.Detail = perr.Error()
		res.GuidelineURL = guideline
		results = append(results, res)
		return results
	}

	allowed := parsed.TestAgent(ctx.FinalURL, "Googlebot")
	if !allowed {
		res := checks.NewResult("seo.robots_txt.blocked", "seo", checks.SeverityError,
			"URL is disallowed for Googlebot by robots.txt")
		res.GuidelineURL = guideline
		results = append(results, res)
	} else {
		res := checks.NewResult("seo.robots_txt.allowed", "seo", checks.SeverityInfo,
			"robots.txt allows Googlebot for this URL")
		res.GuidelineURL = guideline
		results = append(results, res)
	}

	// Look for sitemap references.
	var sitemaps []string
	for _, line := range strings.Split(string(r.Body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "sitemap:") {
			sitemaps = append(sitemaps, strings.TrimSpace(line[len("sitemap:"):]))
		}
	}
	if len(sitemaps) == 0 {
		res := checks.NewResult("seo.sitemap.unreferenced", "seo", checks.SeverityWarning,
			"No Sitemap directive in robots.txt")
		res.Recommendation = "Add `Sitemap: https://example.com/sitemap.xml` to robots.txt."
		res.GuidelineURL = "https://developers.google.com/search/docs/crawling-indexing/sitemaps/overview"
		results = append(results, res)
	} else {
		res := checks.NewResult("seo.sitemap.referenced", "seo", checks.SeverityInfo,
			fmt.Sprintf("Sitemap(s) declared: %s", strings.Join(sitemaps, ", ")))
		res.GuidelineURL = "https://developers.google.com/search/docs/crawling-indexing/sitemaps/overview"
		results = append(results, res)
	}

	return results
}
