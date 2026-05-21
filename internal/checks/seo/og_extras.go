package seo

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// OGImageDimsCheck — recommend og:image:width/height + og:locale.
type OGImageDimsCheck struct{}

func (OGImageDimsCheck) ID() string       { return "seo.og_image_dims" }
func (OGImageDimsCheck) Category() string { return checks.CategorySEO }

func (OGImageDimsCheck) Run(ctx *checks.PageContext) []checks.Result {
	props := map[string]string{}
	ctx.Doc.Find(`head > meta[property^="og:"]`).Each(func(_ int, s *goquery.Selection) {
		k := strings.TrimSpace(s.AttrOr("property", ""))
		v := strings.TrimSpace(s.AttrOr("content", ""))
		if k != "" && v != "" {
			props[k] = v
		}
	})
	if props["og:image"] == "" {
		return nil // OpenGraphCheck already flags missing og:image
	}

	const guideline = "https://ogp.me/#structured"
	var out []checks.Result
	if props["og:image:width"] == "" || props["og:image:height"] == "" {
		r := checks.NewResult("seo.og_image_dims.missing", checks.CategorySEO, checks.SeverityInfo,
			"og:image has no declared width/height")
		r.Recommendation = "Add `og:image:width` and `og:image:height` so platforms can reserve layout without re-fetching the image."
		r.GuidelineURL = guideline
		out = append(out, r)
	} else {
		r := checks.NewResult("seo.og_image_dims.present", checks.CategorySEO, checks.SeverityInfo,
			fmt.Sprintf("og:image dimensions: %s × %s", props["og:image:width"], props["og:image:height"]))
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	if props["og:locale"] == "" {
		r := checks.NewResult("seo.og_locale.missing", checks.CategorySEO, checks.SeverityInfo,
			"No og:locale declared")
		r.Recommendation = "Add `og:locale` (e.g. `en_US`) to help platforms localize the preview."
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	return out
}
