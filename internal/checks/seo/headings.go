package seo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type HeadingsCheck struct{}

func (HeadingsCheck) ID() string       { return "seo.headings" }
func (HeadingsCheck) Category() string { return "seo" }

func (HeadingsCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://www.w3.org/WAI/tutorials/page-structure/headings/"

	var levels []int
	var h1Texts []string
	ctx.Doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if len(tag) != 2 {
			return
		}
		lvl, err := strconv.Atoi(string(tag[1]))
		if err != nil {
			return
		}
		levels = append(levels, lvl)
		if lvl == 1 {
			h1Texts = append(h1Texts, strings.TrimSpace(s.Text()))
		}
	})

	var results []checks.Result

	switch len(h1Texts) {
	case 0:
		r := checks.NewResult("seo.headings.h1_missing", "seo", checks.SeverityError,
			"Missing <h1>")
		r.Recommendation = "Add exactly one <h1> describing the page's primary topic."
		r.GuidelineURL = guideline
		results = append(results, r)
	case 1:
		r := checks.NewResult("seo.headings.h1_present", "seo", checks.SeverityInfo,
			fmt.Sprintf("H1: %q", truncate(h1Texts[0], 80)))
		r.GuidelineURL = guideline
		results = append(results, r)
	default:
		r := checks.NewResult("seo.headings.h1_multiple", "seo", checks.SeverityWarning,
			fmt.Sprintf("Multiple <h1> tags found (%d)", len(h1Texts)))
		r.Detail = "First few: " + strings.Join(truncateList(h1Texts, 3, 40), " | ")
		r.Recommendation = "Use a single <h1> per page; demote the rest to <h2>."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	// Detect skipped heading levels (e.g., h2 -> h4).
	var prev int
	var skips []string
	for _, lvl := range levels {
		if prev > 0 && lvl > prev+1 {
			skips = append(skips, fmt.Sprintf("h%d→h%d", prev, lvl))
		}
		prev = lvl
	}
	if len(skips) > 0 {
		r := checks.NewResult("seo.headings.skipped_levels", "seo", checks.SeverityWarning,
			"Heading levels skip")
		r.Detail = "Skips: " + strings.Join(skips, ", ")
		r.Recommendation = "Avoid jumping heading levels; structure should be hierarchical."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

func truncateList(items []string, max, perItem int) []string {
	if len(items) > max {
		items = items[:max]
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = truncate(it, perItem)
	}
	return out
}
