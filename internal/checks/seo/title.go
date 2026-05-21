package seo

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type TitleCheck struct{}

func (TitleCheck) ID() string       { return "seo.title" }
func (TitleCheck) Category() string { return "seo" }

func (TitleCheck) Run(ctx *checks.PageContext) []checks.Result {
	title := strings.TrimSpace(ctx.Doc.Find("head > title").First().Text())
	guideline := "https://developers.google.com/search/docs/appearance/title-link"

	if title == "" {
		r := checks.NewResult(
			"seo.title.missing", "seo", checks.SeverityError,
			"Missing <title> tag",
		)
		r.Detail = "No <title> tag found in the document head."
		r.Recommendation = "Add a unique, descriptive <title> between 30 and 60 characters."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	var results []checks.Result
	length := len([]rune(title))

	info := checks.NewResult("seo.title.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Title: %q (%d chars)", title, length))
	info.GuidelineURL = guideline
	results = append(results, info)

	switch {
	case length < 30:
		r := checks.NewResult("seo.title.short", "seo", checks.SeverityWarning,
			"Title is shorter than recommended")
		r.Detail = fmt.Sprintf("Title is %d characters; aim for 30–60.", length)
		r.Recommendation = "Expand the title with descriptive, keyword-relevant terms."
		r.GuidelineURL = guideline
		results = append(results, r)
	case length > 60:
		r := checks.NewResult("seo.title.long", "seo", checks.SeverityWarning,
			"Title is longer than recommended")
		r.Detail = fmt.Sprintf("Title is %d characters; SERP truncates around 60.", length)
		r.Recommendation = "Shorten the title so it doesn't get truncated in search results."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}
