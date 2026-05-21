package seo

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type RobotsMetaCheck struct{}

func (RobotsMetaCheck) ID() string       { return "seo.robots_meta" }
func (RobotsMetaCheck) Category() string { return "seo" }

func (RobotsMetaCheck) Run(ctx *checks.PageContext) []checks.Result {
	content := strings.ToLower(strings.TrimSpace(
		ctx.Doc.Find(`head > meta[name="robots"]`).AttrOr("content", ""),
	))
	xRobots := strings.ToLower(strings.TrimSpace(ctx.Headers.Get("X-Robots-Tag")))
	guideline := "https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag"

	var results []checks.Result

	if content == "" && xRobots == "" {
		r := checks.NewResult("seo.robots_meta.absent", "seo", checks.SeverityInfo,
			"No robots meta or X-Robots-Tag (defaults to index, follow)")
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	combined := content
	if xRobots != "" {
		combined = combined + "," + xRobots
	}

	if strings.Contains(combined, "noindex") {
		r := checks.NewResult("seo.robots_meta.noindex", "seo", checks.SeverityError,
			"Page is set to noindex")
		r.Detail = fmt.Sprintf("robots meta=%q, X-Robots-Tag=%q", content, xRobots)
		r.Recommendation = "Remove noindex if you want this page in search results."
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	if strings.Contains(combined, "nofollow") {
		r := checks.NewResult("seo.robots_meta.nofollow", "seo", checks.SeverityWarning,
			"Page is set to nofollow")
		r.Detail = "All outbound links will not pass link equity."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	if len(results) == 0 {
		r := checks.NewResult("seo.robots_meta.present", "seo", checks.SeverityInfo,
			fmt.Sprintf("Robots directives: %s", strings.TrimPrefix(combined, ",")))
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}
