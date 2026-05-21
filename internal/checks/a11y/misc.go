package a11y

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// MetaRefreshCheck — flag <meta http-equiv="refresh">.
type MetaRefreshCheck struct{}

func (MetaRefreshCheck) ID() string       { return "a11y.meta_refresh" }
func (MetaRefreshCheck) Category() string { return checks.CategoryA11y }

func (MetaRefreshCheck) Run(ctx *checks.PageContext) []checks.Result {
	sel := ctx.Doc.Find(`meta[http-equiv="refresh" i]`)
	if sel.Length() == 0 {
		return nil
	}
	r := checks.NewResult("a11y.meta_refresh.present", checks.CategoryA11y, checks.SeverityWarning,
		"Page uses <meta http-equiv=\"refresh\">")
	r.Detail = "content=" + sel.AttrOr("content", "")
	r.Recommendation = "Avoid client-side refresh — it disorients screen-reader users and signals a soft redirect to crawlers."
	r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Understanding/timing-adjustable.html"
	return []checks.Result{r}
}

// VideoCaptionsCheck — every <video> should have a <track kind="captions"|"subtitles">.
type VideoCaptionsCheck struct{}

func (VideoCaptionsCheck) ID() string       { return "a11y.video_captions" }
func (VideoCaptionsCheck) Category() string { return checks.CategoryA11y }

func (VideoCaptionsCheck) Run(ctx *checks.PageContext) []checks.Result {
	var missing int
	ctx.Doc.Find("video").Each(func(_ int, s *goquery.Selection) {
		has := false
		s.Find("track").Each(func(_ int, t *goquery.Selection) {
			k := strings.ToLower(strings.TrimSpace(t.AttrOr("kind", "")))
			if k == "captions" || k == "subtitles" {
				has = true
			}
		})
		if !has {
			missing++
		}
	})
	if missing == 0 {
		return nil
	}
	r := checks.NewResult("a11y.video_captions.missing", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d <video> without a captions/subtitles track", missing))
	r.Recommendation = "Add `<track kind=\"captions\" src=\"…\">` so deaf/HoH users can access spoken content."
	r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Understanding/captions-prerecorded.html"
	return []checks.Result{r}
}

// SkipLinkCheck — first focusable link in body should target an in-page anchor.
type SkipLinkCheck struct{}

func (SkipLinkCheck) ID() string       { return "a11y.skip_link" }
func (SkipLinkCheck) Category() string { return checks.CategoryA11y }

func (SkipLinkCheck) Run(ctx *checks.PageContext) []checks.Result {
	first := ctx.Doc.Find("body a[href]").First()
	if first.Length() == 0 {
		return nil
	}
	href := strings.TrimSpace(first.AttrOr("href", ""))
	if strings.HasPrefix(href, "#") && len(href) > 1 {
		r := checks.NewResult("a11y.skip_link.present", checks.CategoryA11y, checks.SeverityInfo,
			fmt.Sprintf("Skip link present: %s", href))
		r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Techniques/general/G1"
		return []checks.Result{r}
	}
	r := checks.NewResult("a11y.skip_link.missing", checks.CategoryA11y, checks.SeverityInfo,
		"No skip-link as the first focusable link in body")
	r.Recommendation = "Add a `<a href=\"#main\">Skip to content</a>` as the first focusable element so keyboard users can bypass nav."
	r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Techniques/general/G1"
	return []checks.Result{r}
}

// TableHeadersCheck — data tables should use <th> with scope, or headers/id wiring.
type TableHeadersCheck struct{}

func (TableHeadersCheck) ID() string       { return "a11y.table_headers" }
func (TableHeadersCheck) Category() string { return checks.CategoryA11y }

func (TableHeadersCheck) Run(ctx *checks.PageContext) []checks.Result {
	var noTH, scopeMissing, brokenHeaders int

	// Build id set for headers="" validation.
	idSet := map[string]bool{}
	ctx.Doc.Find("[id]").Each(func(_ int, s *goquery.Selection) {
		idSet[strings.TrimSpace(s.AttrOr("id", ""))] = true
	})

	ctx.Doc.Find("table").Each(func(_ int, t *goquery.Selection) {
		// Skip presentation tables (role="presentation"|"none").
		role := strings.ToLower(strings.TrimSpace(t.AttrOr("role", "")))
		if role == "presentation" || role == "none" {
			return
		}
		// Require at least one row with data cells to flag.
		if t.Find("td").Length() == 0 {
			return
		}
		ths := t.Find("th")
		if ths.Length() == 0 {
			noTH++
			return
		}
		ths.Each(func(_ int, th *goquery.Selection) {
			if strings.TrimSpace(th.AttrOr("scope", "")) == "" && strings.TrimSpace(th.AttrOr("id", "")) == "" {
				scopeMissing++
			}
		})
		t.Find("[headers]").Each(func(_ int, cell *goquery.Selection) {
			for _, id := range strings.Fields(cell.AttrOr("headers", "")) {
				if !idSet[id] {
					brokenHeaders++
				}
			}
		})
	})

	var out []checks.Result
	if noTH > 0 {
		r := checks.NewResult("a11y.table_headers.no_th", checks.CategoryA11y, checks.SeverityWarning,
			fmt.Sprintf("%d data table(s) without any <th>", noTH))
		r.Recommendation = "Use `<th>` for column/row headers, or set `role=\"presentation\"` for layout tables."
		r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Techniques/html/H51"
		out = append(out, r)
	}
	if scopeMissing > 0 {
		r := checks.NewResult("a11y.table_headers.no_scope", checks.CategoryA11y, checks.SeverityInfo,
			fmt.Sprintf("%d <th> without scope/id", scopeMissing))
		r.Recommendation = "Add `scope=\"col\"`/`scope=\"row\"` (or `id` + cell `headers`) so screen readers can associate headers with cells."
		r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Techniques/html/H63"
		out = append(out, r)
	}
	if brokenHeaders > 0 {
		r := checks.NewResult("a11y.table_headers.broken_ref", checks.CategoryA11y, checks.SeverityWarning,
			fmt.Sprintf("%d headers=\"\" reference(s) point to missing ids", brokenHeaders))
		r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Techniques/html/H43"
		out = append(out, r)
	}
	return out
}
