package perf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// CacheControlCheck — Cache-Control on the document.
type CacheControlCheck struct{}

func (CacheControlCheck) ID() string       { return "perf.cache_control" }
func (CacheControlCheck) Category() string { return checks.CategoryPerf }

func (CacheControlCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control"
	v := strings.TrimSpace(ctx.Headers.Get("Cache-Control"))
	if v == "" {
		r := checks.NewResult("perf.cache_control.missing", checks.CategoryPerf, checks.SeverityWarning,
			"No Cache-Control header on document")
		r.Recommendation = "Set Cache-Control to control browser caching (e.g. `public, max-age=300, must-revalidate` for HTML)."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	lc := strings.ToLower(v)
	if strings.Contains(lc, "no-store") {
		r := checks.NewResult("perf.cache_control.no_store", checks.CategoryPerf, checks.SeverityInfo,
			fmt.Sprintf("Cache-Control: %s (no caching at all)", v))
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	maxAge := parseMaxAge(lc)
	if maxAge >= 0 && maxAge < 60 {
		r := checks.NewResult("perf.cache_control.short_ttl", checks.CategoryPerf, checks.SeverityWarning,
			fmt.Sprintf("Cache-Control max-age is very short: %d s", maxAge))
		r.Detail = fmt.Sprintf("Cache-Control: %s", v)
		r.Recommendation = "Consider a longer TTL (≥ 60s) for HTML or use `stale-while-revalidate`."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("perf.cache_control.present", checks.CategoryPerf, checks.SeverityInfo,
		fmt.Sprintf("Cache-Control: %s", v))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

func parseMaxAge(lc string) int {
	for _, p := range strings.Split(lc, ",") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "max-age=") {
			n, err := strconv.Atoi(strings.TrimPrefix(p, "max-age="))
			if err == nil {
				return n
			}
		}
	}
	return -1
}

// CompressionCheck — Content-Encoding on the document.
type CompressionCheck struct{}

func (CompressionCheck) ID() string       { return "perf.compression" }
func (CompressionCheck) Category() string { return checks.CategoryPerf }

func (CompressionCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding"
	enc := strings.ToLower(strings.TrimSpace(ctx.Headers.Get("Content-Encoding")))
	if enc == "" {
		r := checks.NewResult("perf.compression.missing", checks.CategoryPerf, checks.SeverityWarning,
			"Document is served uncompressed")
		r.Recommendation = "Enable gzip, br (Brotli), or zstd at the origin/CDN for text responses."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	sev := checks.SeverityInfo
	rec := ""
	if enc == "gzip" {
		rec = "Brotli (`br`) typically saves ~15-25% over gzip; consider enabling it if your CDN supports it."
	}
	r := checks.NewResult("perf.compression.present", checks.CategoryPerf, sev,
		fmt.Sprintf("Content-Encoding: %s", enc))
	r.Recommendation = rec
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// HTTPProtocolCheck — flag HTTP/1.1 vs HTTP/2/3.
type HTTPProtocolCheck struct{}

func (HTTPProtocolCheck) ID() string       { return "perf.http_protocol" }
func (HTTPProtocolCheck) Category() string { return checks.CategoryPerf }

func (HTTPProtocolCheck) Run(ctx *checks.PageContext) []checks.Result {
	if ctx.Proto == "" {
		return nil
	}
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism"
	sev := checks.SeverityInfo
	rec := ""
	if strings.HasPrefix(ctx.Proto, "HTTP/1") {
		sev = checks.SeverityWarning
		rec = "Upgrade to HTTP/2 or HTTP/3 (most CDNs offer this for free) to reduce request overhead."
	}
	r := checks.NewResult("perf.http_protocol", checks.CategoryPerf, sev,
		fmt.Sprintf("Protocol: %s", ctx.Proto))
	r.Recommendation = rec
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// RedirectChainCheck — score redirect chain length. The SEO StatusCheck
// surfaces a count; this gives a perf-category band.
type RedirectChainCheck struct{}

func (RedirectChainCheck) ID() string       { return "perf.redirects" }
func (RedirectChainCheck) Category() string { return checks.CategoryPerf }

func (RedirectChainCheck) Run(ctx *checks.PageContext) []checks.Result {
	n := len(ctx.Redirects)
	if n == 0 {
		return nil
	}
	sev := checks.SeverityInfo
	rec := ""
	if n == 2 {
		sev = checks.SeverityWarning
		rec = "Each redirect adds a round trip; collapse where possible."
	}
	if n >= 3 {
		sev = checks.SeverityError
		rec = "Long redirect chain — collapse to a single hop to recover latency."
	}
	r := checks.NewResult("perf.redirects.chain", checks.CategoryPerf, sev,
		fmt.Sprintf("Redirect chain length: %d", n))
	r.Detail = "Hops: " + strings.Join(ctx.Redirects, " → ")
	r.Recommendation = rec
	r.GuidelineURL = "https://web.dev/articles/redirects"
	return []checks.Result{r}
}
