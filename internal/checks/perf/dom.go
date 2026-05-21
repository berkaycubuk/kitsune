package perf

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// DOMSizeCheck — total nodes, max depth, max sibling count.
type DOMSizeCheck struct{}

func (DOMSizeCheck) ID() string       { return "perf.dom_size" }
func (DOMSizeCheck) Category() string { return checks.CategoryPerf }

func (DOMSizeCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/dom-size-and-interactivity"
	const (
		warnNodes    = 1500
		errNodes     = 3000
		warnDepth    = 32
		warnChildren = 60
	)

	total := 0
	maxDepth := 0
	maxChildren := 0

	var walk func(s *goquery.Selection, depth int)
	walk = func(s *goquery.Selection, depth int) {
		total++
		if depth > maxDepth {
			maxDepth = depth
		}
		children := s.Children()
		if n := children.Length(); n > maxChildren {
			maxChildren = n
		}
		children.Each(func(_ int, c *goquery.Selection) {
			walk(c, depth+1)
		})
	}
	ctx.Doc.Find("body").Each(func(_ int, s *goquery.Selection) {
		walk(s, 1)
	})

	sev := checks.SeverityInfo
	if total >= errNodes {
		sev = checks.SeverityError
	} else if total >= warnNodes || maxDepth >= warnDepth || maxChildren >= warnChildren {
		sev = checks.SeverityWarning
	}

	r := checks.NewResult("perf.dom_size.metrics", checks.CategoryPerf, sev,
		fmt.Sprintf("DOM: %d nodes, depth %d, max %d children", total, maxDepth, maxChildren))
	if sev != checks.SeverityInfo {
		r.Recommendation = "Lighthouse flags >1,500 nodes, depth >32, or >60 children. Split large lists, virtualize, or simplify markup."
	}
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
