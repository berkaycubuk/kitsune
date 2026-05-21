package a11y

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// DuplicateIDCheck — every id must be unique within the document.
type DuplicateIDCheck struct{}

func (DuplicateIDCheck) ID() string       { return "a11y.duplicate_id" }
func (DuplicateIDCheck) Category() string { return checks.CategoryA11y }

func (DuplicateIDCheck) Run(ctx *checks.PageContext) []checks.Result {
	counts := map[string]int{}
	ctx.Doc.Find("[id]").Each(func(_ int, s *goquery.Selection) {
		id := strings.TrimSpace(s.AttrOr("id", ""))
		if id == "" {
			return
		}
		counts[id]++
	})
	var dups []string
	for id, n := range counts {
		if n > 1 {
			dups = append(dups, fmt.Sprintf("%s (×%d)", id, n))
		}
	}
	if len(dups) == 0 {
		return nil
	}
	sort.Strings(dups)
	r := checks.NewResult("a11y.duplicate_id.found", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d duplicate id value(s)", len(dups)))
	if len(dups) > 5 {
		dups = append(dups[:5], fmt.Sprintf("…+%d more", len(dups)-5))
	}
	r.Detail = strings.Join(dups, ", ")
	r.Recommendation = "ids must be unique; duplicates break label[for], aria-labelledby, and other a11y wiring."
	r.GuidelineURL = "https://dequeuniversity.com/rules/axe/4.7/duplicate-id"
	return []checks.Result{r}
}

// TabindexCheck — flag tabindex > 0.
type TabindexCheck struct{}

func (TabindexCheck) ID() string       { return "a11y.tabindex" }
func (TabindexCheck) Category() string { return checks.CategoryA11y }

func (TabindexCheck) Run(ctx *checks.PageContext) []checks.Result {
	var bad int
	ctx.Doc.Find("[tabindex]").Each(func(_ int, s *goquery.Selection) {
		v := strings.TrimSpace(s.AttrOr("tabindex", ""))
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			bad++
		}
	})
	if bad == 0 {
		return nil
	}
	r := checks.NewResult("a11y.tabindex.positive", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d element(s) with tabindex > 0", bad))
	r.Recommendation = "Positive tabindex disrupts natural tab order. Use `tabindex=\"0\"` (or `-1`)."
	r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Understanding/focus-order.html"
	return []checks.Result{r}
}
