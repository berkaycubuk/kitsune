package seo

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type ImageAltCheck struct{}

func (ImageAltCheck) ID() string       { return "seo.image_alt" }
func (ImageAltCheck) Category() string { return "seo" }

func (ImageAltCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://www.w3.org/WAI/tutorials/images/"

	var total, missing int
	ctx.Doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		total++
		_, has := s.Attr("alt")
		if !has {
			missing++
		}
	})

	if total == 0 {
		return nil
	}

	if missing == 0 {
		r := checks.NewResult("seo.image_alt.ok", "seo", checks.SeverityInfo,
			fmt.Sprintf("All %d image(s) have an alt attribute", total))
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	r := checks.NewResult("seo.image_alt.missing", "seo", checks.SeverityWarning,
		fmt.Sprintf("%d of %d image(s) missing alt attribute", missing, total))
	r.Recommendation = "Add descriptive alt text; use alt=\"\" for purely decorative images."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
