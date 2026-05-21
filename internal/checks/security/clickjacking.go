package security

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// ClickjackingCheck — X-Frame-Options OR CSP frame-ancestors must be present.
type ClickjackingCheck struct{}

func (ClickjackingCheck) ID() string       { return "security.clickjacking" }
func (ClickjackingCheck) Category() string { return checks.CategorySecurity }

func (ClickjackingCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options"

	xfo := strings.TrimSpace(ctx.Headers.Get("X-Frame-Options"))
	csp := strings.TrimSpace(ctx.Headers.Get("Content-Security-Policy"))
	if csp == "" {
		csp = strings.TrimSpace(ctx.Doc.Find(`meta[http-equiv="Content-Security-Policy" i]`).AttrOr("content", ""))
	}
	hasFrameAncestors := false
	if csp != "" {
		_, hasFrameAncestors = parseCSP(csp)["frame-ancestors"]
	}

	if xfo == "" && !hasFrameAncestors {
		r := checks.NewResult("security.clickjacking.missing", checks.CategorySecurity, checks.SeverityWarning,
			"No clickjacking protection (X-Frame-Options or CSP frame-ancestors)")
		r.Recommendation = "Set `X-Frame-Options: DENY` or a CSP `frame-ancestors` directive."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	parts := []string{}
	if xfo != "" {
		parts = append(parts, fmt.Sprintf("X-Frame-Options: %s", xfo))
	}
	if hasFrameAncestors {
		parts = append(parts, "CSP frame-ancestors")
	}
	r := checks.NewResult("security.clickjacking.present", checks.CategorySecurity, checks.SeverityInfo,
		"Clickjacking protection: "+strings.Join(parts, ", "))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
