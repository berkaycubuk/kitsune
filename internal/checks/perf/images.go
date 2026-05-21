package perf

import (
	"fmt"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// ResponsiveImagesCheck — flag <img> without srcset/sizes (and not in <picture>).
type ResponsiveImagesCheck struct{}

func (ResponsiveImagesCheck) ID() string       { return "perf.responsive_images" }
func (ResponsiveImagesCheck) Category() string { return checks.CategoryPerf }

func (ResponsiveImagesCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/serve-responsive-images"
	var missing int
	ctx.Doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		if s.ParentFiltered("picture").Length() > 0 {
			return
		}
		if strings.TrimSpace(s.AttrOr("srcset", "")) != "" {
			return
		}
		missing++
	})
	if missing == 0 {
		return nil
	}
	r := checks.NewResult("perf.responsive_images.missing_srcset", checks.CategoryPerf, checks.SeverityInfo,
		fmt.Sprintf("%d <img> without srcset/<picture>", missing))
	r.Recommendation = "Use `srcset`/`sizes` (or a `<picture>` element) so browsers can pick an appropriately-sized variant."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// LegacyImageFormatsCheck — flag .jpg/.png without a modern-format alternative.
type LegacyImageFormatsCheck struct{}

func (LegacyImageFormatsCheck) ID() string       { return "perf.legacy_image_formats" }
func (LegacyImageFormatsCheck) Category() string { return checks.CategoryPerf }

func (LegacyImageFormatsCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/uses-webp-images"
	var legacy int
	ctx.Doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		if s.ParentFiltered("picture").Length() > 0 {
			// Assume <picture> has a modern source — could deepen later.
			return
		}
		ext := strings.ToLower(path.Ext(strings.SplitN(s.AttrOr("src", ""), "?", 2)[0]))
		switch ext {
		case ".jpg", ".jpeg", ".png":
			legacy++
		}
	})
	if legacy == 0 {
		return nil
	}
	r := checks.NewResult("perf.legacy_image_formats.count", checks.CategoryPerf, checks.SeverityInfo,
		fmt.Sprintf("%d <img> using legacy format (.jpg/.png) without <picture> fallback", legacy))
	r.Recommendation = "Serve WebP or AVIF via `<picture>` with a `<source type=\"image/webp\">` and the legacy as fallback."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// LazyLoadingCheck — first <img> should NOT be lazy; later images SHOULD be.
type LazyLoadingCheck struct{}

func (LazyLoadingCheck) ID() string       { return "perf.lazy_loading" }
func (LazyLoadingCheck) Category() string { return checks.CategoryPerf }

func (LazyLoadingCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/browser-level-image-lazy-loading"

	imgs := ctx.Doc.Find("body img")
	if imgs.Length() == 0 {
		return nil
	}

	var out []checks.Result
	first := imgs.First()
	if strings.EqualFold(first.AttrOr("loading", ""), "lazy") {
		r := checks.NewResult("perf.lazy_loading.first_image_lazy", checks.CategoryPerf, checks.SeverityWarning,
			"First body image has loading=\"lazy\" (hurts LCP)")
		r.Detail = "src: " + first.AttrOr("src", "")
		r.Recommendation = "Remove `loading=\"lazy\"` from above-the-fold/LCP images."
		r.GuidelineURL = guideline
		out = append(out, r)
	}

	var nonLazyBelow int
	imgs.Each(func(i int, s *goquery.Selection) {
		if i < 3 {
			return
		}
		if strings.EqualFold(s.AttrOr("loading", ""), "lazy") {
			return
		}
		nonLazyBelow++
	})
	if nonLazyBelow > 0 {
		r := checks.NewResult("perf.lazy_loading.missing_lazy", checks.CategoryPerf, checks.SeverityInfo,
			fmt.Sprintf("%d likely-offscreen <img> without loading=\"lazy\"", nonLazyBelow))
		r.Recommendation = "Add `loading=\"lazy\"` to images below the fold to defer offscreen network work."
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	return out
}

// FetchPriorityCheck — first body <img> should have fetchpriority="high".
type FetchPriorityCheck struct{}

func (FetchPriorityCheck) ID() string       { return "perf.fetchpriority" }
func (FetchPriorityCheck) Category() string { return checks.CategoryPerf }

func (FetchPriorityCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/fetch-priority"
	first := ctx.Doc.Find("body img").First()
	if first.Length() == 0 {
		return nil
	}
	fp := strings.ToLower(strings.TrimSpace(first.AttrOr("fetchpriority", "")))
	if fp == "high" {
		r := checks.NewResult("perf.fetchpriority.lcp_set", checks.CategoryPerf, checks.SeverityInfo,
			"Likely LCP image has fetchpriority=\"high\"")
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("perf.fetchpriority.lcp_missing", checks.CategoryPerf, checks.SeverityInfo,
		"Likely LCP image lacks fetchpriority=\"high\"")
	r.Detail = "src: " + first.AttrOr("src", "")
	r.Recommendation = "Add `fetchpriority=\"high\"` to the LCP image so the browser prioritizes its download."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// MediaDimensionsCheck — <img>/<iframe>/<video> missing width+height (CLS).
type MediaDimensionsCheck struct{}

func (MediaDimensionsCheck) ID() string       { return "perf.media_dimensions" }
func (MediaDimensionsCheck) Category() string { return checks.CategoryPerf }

func (MediaDimensionsCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/optimize-cls"
	missing := map[string]int{}
	for _, tag := range []string{"img", "iframe", "video"} {
		ctx.Doc.Find(tag).Each(func(_ int, s *goquery.Selection) {
			w := strings.TrimSpace(s.AttrOr("width", ""))
			h := strings.TrimSpace(s.AttrOr("height", ""))
			if w == "" || h == "" {
				missing[tag]++
			}
		})
	}
	var out []checks.Result
	for _, tag := range []string{"img", "iframe", "video"} {
		if n := missing[tag]; n > 0 {
			r := checks.NewResult("perf.media_dimensions."+tag, checks.CategoryPerf, checks.SeverityWarning,
				fmt.Sprintf("%d <%s> missing width/height (CLS risk)", n, tag))
			r.Recommendation = "Set explicit `width` and `height` so the browser reserves layout space (prevents shift)."
			r.GuidelineURL = guideline
			out = append(out, r)
		}
	}
	return out
}

// AnimatedGIFCheck — flag <img src="*.gif">.
type AnimatedGIFCheck struct{}

func (AnimatedGIFCheck) ID() string       { return "perf.animated_gif" }
func (AnimatedGIFCheck) Category() string { return checks.CategoryPerf }

func (AnimatedGIFCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/replace-gifs-with-videos"
	var n int
	ctx.Doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		src := strings.ToLower(strings.SplitN(s.AttrOr("src", ""), "?", 2)[0])
		if strings.HasSuffix(src, ".gif") {
			n++
		}
	})
	if n == 0 {
		return nil
	}
	r := checks.NewResult("perf.animated_gif.count", checks.CategoryPerf, checks.SeverityInfo,
		fmt.Sprintf("%d <img> using .gif", n))
	r.Recommendation = "For animated content, replace with `<video autoplay muted loop playsinline>` — typically 5-10× smaller."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
