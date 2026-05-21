package perf

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// RenderBlockingCheck — sync <script> and <link rel="stylesheet"> in <head>.
type RenderBlockingCheck struct{}

func (RenderBlockingCheck) ID() string       { return "perf.render_blocking" }
func (RenderBlockingCheck) Category() string { return checks.CategoryPerf }

func (RenderBlockingCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://web.dev/articles/render-blocking-resources"

	var scripts, styles []string
	ctx.Doc.Find("head > script[src]").Each(func(_ int, s *goquery.Selection) {
		_, hasAsync := s.Attr("async")
		_, hasDefer := s.Attr("defer")
		typ := strings.ToLower(strings.TrimSpace(s.AttrOr("type", "")))
		if hasAsync || hasDefer || typ == "module" {
			return
		}
		scripts = append(scripts, s.AttrOr("src", ""))
	})
	ctx.Doc.Find(`head > link[rel~="stylesheet"]`).Each(func(_ int, s *goquery.Selection) {
		media := strings.ToLower(strings.TrimSpace(s.AttrOr("media", "")))
		if media != "" && media != "all" && media != "screen" {
			return
		}
		styles = append(styles, s.AttrOr("href", ""))
	})

	if len(scripts) == 0 && len(styles) == 0 {
		r := checks.NewResult("perf.render_blocking.none", checks.CategoryPerf, checks.SeverityInfo,
			"No render-blocking resources in <head>")
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	var out []checks.Result
	if len(scripts) > 0 {
		r := checks.NewResult("perf.render_blocking.scripts", checks.CategoryPerf, checks.SeverityWarning,
			fmt.Sprintf("%d render-blocking script(s) in <head>", len(scripts)))
		r.Detail = strings.Join(truncList(scripts, 5), ", ")
		r.Recommendation = "Add `async` or `defer`, or move below the fold; mark ES modules with `type=\"module\"`."
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	if len(styles) > 0 {
		r := checks.NewResult("perf.render_blocking.styles", checks.CategoryPerf, checks.SeverityWarning,
			fmt.Sprintf("%d render-blocking stylesheet(s) in <head>", len(styles)))
		r.Detail = strings.Join(truncList(styles, 5), ", ")
		r.Recommendation = "Inline critical CSS or load non-critical sheets with `media=\"print\" onload=\"this.media='all'\"`."
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	return out
}

func truncList(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	out := make([]string, n+1)
	copy(out, in[:n])
	out[n] = fmt.Sprintf("…+%d more", len(in)-n)
	return out
}

// ResourceHintsCheck — preconnect/preload coverage.
type ResourceHintsCheck struct{}

func (ResourceHintsCheck) ID() string       { return "perf.resource_hints" }
func (ResourceHintsCheck) Category() string { return checks.CategoryPerf }

func (ResourceHintsCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel/preconnect"

	docOrigin := originOf(ctx.FinalURL)

	// Cross-origin domains referenced from <head> (scripts, stylesheets, fonts).
	external := map[string]bool{}
	addExternal := func(href string) {
		o := originOf(href)
		if o == "" || o == docOrigin {
			return
		}
		external[o] = true
	}
	ctx.Doc.Find("head > script[src], head > link[rel~=\"stylesheet\"]").Each(func(_ int, s *goquery.Selection) {
		addExternal(s.AttrOr("src", "") + s.AttrOr("href", ""))
	})
	ctx.Doc.Find(`head > link[rel="preconnect"], head > link[rel="dns-prefetch"]`).Each(func(_ int, s *goquery.Selection) {
		// Remove already-hinted origins from the "missing" set.
		delete(external, originOf(s.AttrOr("href", "")))
	})

	var out []checks.Result
	if len(external) > 0 {
		domains := make([]string, 0, len(external))
		for d := range external {
			domains = append(domains, d)
		}
		r := checks.NewResult("perf.resource_hints.missing_preconnect", checks.CategoryPerf, checks.SeverityInfo,
			fmt.Sprintf("%d cross-origin domain(s) in <head> without preconnect/dns-prefetch", len(domains)))
		r.Detail = strings.Join(truncList(domains, 5), ", ")
		r.Recommendation = "Add `<link rel=\"preconnect\" href=\"…\" crossorigin>` for fonts and critical third-party origins."
		r.GuidelineURL = guideline
		out = append(out, r)
	}

	// Suggest preload for likely-LCP image if none is preloaded.
	firstImg := ctx.Doc.Find("body img").First()
	if firstImg.Length() > 0 {
		src := strings.TrimSpace(firstImg.AttrOr("src", ""))
		preloaded := false
		ctx.Doc.Find(`head > link[rel="preload"][as="image"]`).Each(func(_ int, s *goquery.Selection) {
			if s.AttrOr("href", "") == src {
				preloaded = true
			}
		})
		if !preloaded && src != "" {
			r := checks.NewResult("perf.resource_hints.lcp_preload", checks.CategoryPerf, checks.SeverityInfo,
				"Likely LCP image not preloaded")
			r.Detail = "First body <img>: " + src
			r.Recommendation = "Consider `<link rel=\"preload\" as=\"image\" href=\"…\" fetchpriority=\"high\">` for above-the-fold hero images."
			r.GuidelineURL = "https://web.dev/articles/preload-critical-assets"
			out = append(out, r)
		}
	}
	return out
}

func originOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// InlineSizeCheck — flag oversized inline <script>/<style> blocks.
type InlineSizeCheck struct{}

func (InlineSizeCheck) ID() string       { return "perf.inline_size" }
func (InlineSizeCheck) Category() string { return checks.CategoryPerf }

func (InlineSizeCheck) Run(ctx *checks.PageContext) []checks.Result {
	const (
		perBlock = 10 * 1024
		total    = 50 * 1024
	)
	const guideline = "https://web.dev/articles/extract-critical-css"

	var out []checks.Result
	for _, tag := range []string{"script", "style"} {
		var sum, oversized int
		ctx.Doc.Find(tag).Each(func(_ int, s *goquery.Selection) {
			if tag == "script" && strings.TrimSpace(s.AttrOr("src", "")) != "" {
				return
			}
			n := len(s.Text())
			sum += n
			if n >= perBlock {
				oversized++
			}
		})
		if oversized > 0 {
			r := checks.NewResult("perf.inline_size."+tag+"_block", checks.CategoryPerf, checks.SeverityWarning,
				fmt.Sprintf("%d inline <%s> block(s) > %d KiB", oversized, tag, perBlock/1024))
			r.Recommendation = "Extract to an external file (gets compressed + cached) and inline only what's render-critical."
			r.GuidelineURL = guideline
			out = append(out, r)
		}
		if sum >= total {
			r := checks.NewResult("perf.inline_size."+tag+"_total", checks.CategoryPerf, checks.SeverityWarning,
				fmt.Sprintf("Total inline <%s> bytes: %d KiB (>%d KiB)", tag, sum/1024, total/1024))
			r.GuidelineURL = guideline
			out = append(out, r)
		}
	}
	return out
}

// ExternalHeadRequestsCheck — third-party origins in <head>.
type ExternalHeadRequestsCheck struct{}

func (ExternalHeadRequestsCheck) ID() string       { return "perf.external_head" }
func (ExternalHeadRequestsCheck) Category() string { return checks.CategoryPerf }

func (ExternalHeadRequestsCheck) Run(ctx *checks.PageContext) []checks.Result {
	docOrigin := originOf(ctx.FinalURL)
	byOrigin := map[string]int{}
	ctx.Doc.Find("head script[src], head link[href]").Each(func(_ int, s *goquery.Selection) {
		ref := s.AttrOr("src", "") + s.AttrOr("href", "")
		o := originOf(ref)
		if o == "" || o == docOrigin {
			return
		}
		byOrigin[o]++
	})
	if len(byOrigin) == 0 {
		return nil
	}
	total := 0
	for _, n := range byOrigin {
		total += n
	}
	sev := checks.SeverityInfo
	if len(byOrigin) > 4 {
		sev = checks.SeverityWarning
	}
	r := checks.NewResult("perf.external_head.count", checks.CategoryPerf, sev,
		fmt.Sprintf("%d third-party request(s) in <head> across %d origin(s)", total, len(byOrigin)))
	parts := make([]string, 0, len(byOrigin))
	for o, n := range byOrigin {
		parts = append(parts, fmt.Sprintf("%s (%d)", o, n))
	}
	r.Detail = strings.Join(truncList(parts, 6), ", ")
	if sev == checks.SeverityWarning {
		r.Recommendation = "Each third-party origin needs a DNS lookup + TLS handshake. Consolidate or self-host where possible."
	}
	r.GuidelineURL = "https://web.dev/articles/third-party-summary"
	return []checks.Result{r}
}
