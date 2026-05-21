package seo

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type DescriptionCheck struct{}

func (DescriptionCheck) ID() string       { return "seo.description" }
func (DescriptionCheck) Category() string { return "seo" }

func (DescriptionCheck) Run(ctx *checks.PageContext) []checks.Result {
	desc := strings.TrimSpace(ctx.Doc.Find(`head > meta[name="description"]`).AttrOr("content", ""))
	guideline := "https://developers.google.com/search/docs/appearance/snippet"

	if desc == "" {
		r := checks.NewResult("seo.description.missing", "seo", checks.SeverityWarning,
			"Missing meta description")
		r.Detail = `No <meta name="description"> tag found.`
		r.Recommendation = "Add a 70–160 character meta description summarizing the page."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	length := len([]rune(desc))
	var results []checks.Result

	info := checks.NewResult("seo.description.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Meta description present (%d chars)", length))
	info.GuidelineURL = guideline
	results = append(results, info)

	switch {
	case length < 70:
		r := checks.NewResult("seo.description.short", "seo", checks.SeverityWarning,
			"Meta description is short")
		r.Detail = fmt.Sprintf("Description is %d characters; aim for 70–160.", length)
		r.GuidelineURL = guideline
		results = append(results, r)
	case length > 160:
		r := checks.NewResult("seo.description.long", "seo", checks.SeverityWarning,
			"Meta description is long")
		r.Detail = fmt.Sprintf("Description is %d characters; Google truncates around 160.", length)
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}
