package seo

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// TargetBlankCheck — `target="_blank"` without `rel="noopener"` is a tabnabbing vector.
type TargetBlankCheck struct{}

func (TargetBlankCheck) ID() string       { return "seo.target_blank" }
func (TargetBlankCheck) Category() string { return checks.CategorySEO }

func (TargetBlankCheck) Run(ctx *checks.PageContext) []checks.Result {
	var unsafe int
	ctx.Doc.Find(`a[target="_blank"]`).Each(func(_ int, s *goquery.Selection) {
		rel := strings.ToLower(s.AttrOr("rel", ""))
		if strings.Contains(rel, "noopener") || strings.Contains(rel, "noreferrer") {
			return
		}
		unsafe++
	})
	if unsafe == 0 {
		return nil
	}
	r := checks.NewResult("seo.target_blank.unsafe", checks.CategorySEO, checks.SeverityWarning,
		fmt.Sprintf("%d link(s) with target=\"_blank\" missing rel=\"noopener\"", unsafe))
	r.Recommendation = "Add `rel=\"noopener\"` (or `noreferrer`) to prevent the opened page from taking control of `window.opener`."
	r.GuidelineURL = "https://web.dev/articles/external-anchors-use-rel-noopener"
	return []checks.Result{r}
}
