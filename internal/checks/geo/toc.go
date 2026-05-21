package geo

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// TOCCheck — look for a <nav> containing in-page anchor links (table of contents).
type TOCCheck struct{}

func (TOCCheck) ID() string       { return "geo.toc" }
func (TOCCheck) Category() string { return checks.CategoryGEO }

func (TOCCheck) Run(ctx *checks.PageContext) []checks.Result {
	hasTOC := false
	var entries int
	ctx.Doc.Find("nav, [role=\"navigation\"]").EachWithBreak(func(_ int, n *goquery.Selection) bool {
		var internal int
		n.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
			href := strings.TrimSpace(a.AttrOr("href", ""))
			if strings.HasPrefix(href, "#") && len(href) > 1 {
				internal++
			}
		})
		if internal >= 2 {
			hasTOC = true
			entries = internal
			return false
		}
		return true
	})
	if !hasTOC {
		return nil
	}
	r := checks.NewResult("geo.toc.present", checks.CategoryGEO, checks.SeverityInfo,
		fmt.Sprintf("Table-of-contents detected: %d in-page anchor link(s)", entries))
	r.Detail = "TOC structure helps AI assistants extract and cite specific sections."
	r.GuidelineURL = "https://arxiv.org/abs/2311.09735"
	return []checks.Result{r}
}
