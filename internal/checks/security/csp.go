package security

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type CSPCheck struct{}

func (CSPCheck) ID() string       { return "security.csp" }
func (CSPCheck) Category() string { return checks.CategorySecurity }

func (CSPCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy"

	v := strings.TrimSpace(ctx.Headers.Get("Content-Security-Policy"))
	source := "header"
	if v == "" {
		// CSP may also be set via <meta http-equiv="Content-Security-Policy">.
		v = strings.TrimSpace(ctx.Doc.Find(`meta[http-equiv="Content-Security-Policy" i]`).AttrOr("content", ""))
		if v != "" {
			source = "meta"
		}
	}
	if v == "" {
		r := checks.NewResult("security.csp.missing", checks.CategorySecurity, checks.SeverityWarning,
			"No Content-Security-Policy set")
		r.Recommendation = "Add a CSP header to mitigate XSS and data-injection attacks."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	directives := parseCSP(v)
	var out []checks.Result
	r := checks.NewResult("security.csp.present", checks.CategorySecurity, checks.SeverityInfo,
		fmt.Sprintf("CSP (%s): %d directive(s)", source, len(directives)))
	r.Detail = truncate(v, 240)
	r.GuidelineURL = guideline
	out = append(out, r)

	for _, dir := range []string{"script-src", "style-src", "default-src"} {
		vals := directives[dir]
		if len(vals) == 0 {
			continue
		}
		joined := strings.Join(vals, " ")
		flags := []string{}
		for _, bad := range []string{"'unsafe-inline'", "'unsafe-eval'"} {
			if strings.Contains(joined, bad) {
				flags = append(flags, bad)
			}
		}
		if len(flags) > 0 {
			r := checks.NewResult("security.csp.unsafe."+dir, checks.CategorySecurity, checks.SeverityWarning,
				fmt.Sprintf("CSP %s allows %s", dir, strings.Join(flags, ", ")))
			r.Recommendation = "Replace unsafe-inline/unsafe-eval with nonces or hashes where possible."
			r.GuidelineURL = guideline
			out = append(out, r)
		}
	}
	return out
}

func parseCSP(v string) map[string][]string {
	out := map[string][]string{}
	for _, part := range strings.Split(v, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		name := strings.ToLower(fields[0])
		out[name] = fields[1:]
	}
	return out
}
