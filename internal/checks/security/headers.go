package security

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

func isHTTPS(ctx *checks.PageContext) bool {
	u, err := url.Parse(ctx.FinalURL)
	return err == nil && u.Scheme == "https"
}

// HSTSCheck — Strict-Transport-Security.
type HSTSCheck struct{}

func (HSTSCheck) ID() string       { return "security.hsts" }
func (HSTSCheck) Category() string { return checks.CategorySecurity }

func (HSTSCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security"
	if !isHTTPS(ctx) {
		return nil
	}
	v := strings.TrimSpace(ctx.Headers.Get("Strict-Transport-Security"))
	if v == "" {
		r := checks.NewResult("security.hsts.missing", checks.CategorySecurity, checks.SeverityWarning,
			"Missing Strict-Transport-Security header")
		r.Recommendation = "Add `Strict-Transport-Security: max-age=31536000; includeSubDomains` to force HTTPS for repeat visitors."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("security.hsts.present", checks.CategorySecurity, checks.SeverityInfo,
		fmt.Sprintf("HSTS: %s", v))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// NosniffCheck — X-Content-Type-Options: nosniff.
type NosniffCheck struct{}

func (NosniffCheck) ID() string       { return "security.nosniff" }
func (NosniffCheck) Category() string { return checks.CategorySecurity }

func (NosniffCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options"
	v := strings.ToLower(strings.TrimSpace(ctx.Headers.Get("X-Content-Type-Options")))
	if v == "" {
		r := checks.NewResult("security.nosniff.missing", checks.CategorySecurity, checks.SeverityWarning,
			"Missing X-Content-Type-Options header")
		r.Recommendation = "Add `X-Content-Type-Options: nosniff` to block MIME-type sniffing."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	if v != "nosniff" {
		r := checks.NewResult("security.nosniff.invalid", checks.CategorySecurity, checks.SeverityWarning,
			fmt.Sprintf("X-Content-Type-Options has unexpected value: %q", v))
		r.Recommendation = "Set the header to exactly `nosniff`."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("security.nosniff.present", checks.CategorySecurity, checks.SeverityInfo,
		"X-Content-Type-Options: nosniff")
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// ReferrerPolicyCheck.
type ReferrerPolicyCheck struct{}

func (ReferrerPolicyCheck) ID() string       { return "security.referrer_policy" }
func (ReferrerPolicyCheck) Category() string { return checks.CategorySecurity }

func (ReferrerPolicyCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy"
	v := strings.ToLower(strings.TrimSpace(ctx.Headers.Get("Referrer-Policy")))
	if v == "" {
		// Fall back to <meta name="referrer">.
		v = strings.ToLower(strings.TrimSpace(ctx.Doc.Find(`meta[name="referrer"]`).AttrOr("content", "")))
	}
	if v == "" {
		r := checks.NewResult("security.referrer_policy.missing", checks.CategorySecurity, checks.SeverityWarning,
			"Missing Referrer-Policy")
		r.Recommendation = "Set a Referrer-Policy header (e.g. `strict-origin-when-cross-origin`) to control referer leakage."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	leaky := map[string]bool{"unsafe-url": true, "no-referrer-when-downgrade": true}
	if leaky[v] && isHTTPS(ctx) {
		r := checks.NewResult("security.referrer_policy.leaky", checks.CategorySecurity, checks.SeverityWarning,
			fmt.Sprintf("Referrer-Policy %q leaks the full URL", v))
		r.Recommendation = "Use `strict-origin-when-cross-origin` or stricter."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("security.referrer_policy.present", checks.CategorySecurity, checks.SeverityInfo,
		fmt.Sprintf("Referrer-Policy: %s", v))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// PermissionsPolicyCheck — presence of Permissions-Policy (formerly Feature-Policy).
type PermissionsPolicyCheck struct{}

func (PermissionsPolicyCheck) ID() string       { return "security.permissions_policy" }
func (PermissionsPolicyCheck) Category() string { return checks.CategorySecurity }

func (PermissionsPolicyCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Permissions-Policy"
	v := strings.TrimSpace(ctx.Headers.Get("Permissions-Policy"))
	if v == "" {
		v = strings.TrimSpace(ctx.Headers.Get("Feature-Policy"))
	}
	if v == "" {
		r := checks.NewResult("security.permissions_policy.missing", checks.CategorySecurity, checks.SeverityInfo,
			"No Permissions-Policy header set")
		r.Recommendation = "Consider a `Permissions-Policy` to lock down powerful features (camera, geolocation, microphone) you don't use."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("security.permissions_policy.present", checks.CategorySecurity, checks.SeverityInfo,
		fmt.Sprintf("Permissions-Policy: %s", truncate(v, 120)))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// ServerDisclosureCheck — flag Server / X-Powered-By disclosure.
type ServerDisclosureCheck struct{}

func (ServerDisclosureCheck) ID() string       { return "security.server_disclosure" }
func (ServerDisclosureCheck) Category() string { return checks.CategorySecurity }

func (ServerDisclosureCheck) Run(ctx *checks.PageContext) []checks.Result {
	var out []checks.Result
	for _, h := range []string{"Server", "X-Powered-By", "X-AspNet-Version", "X-AspNetMvc-Version"} {
		v := strings.TrimSpace(ctx.Headers.Get(h))
		if v == "" {
			continue
		}
		sev := checks.SeverityInfo
		if hasVersion(v) {
			sev = checks.SeverityWarning
		}
		r := checks.NewResult("security.server_disclosure."+strings.ToLower(h), checks.CategorySecurity, sev,
			fmt.Sprintf("%s: %s", h, v))
		if sev == checks.SeverityWarning {
			r.Recommendation = "Strip the version from the header to reduce attack-surface enumeration."
		}
		r.GuidelineURL = "https://owasp.org/www-project-secure-headers/"
		out = append(out, r)
	}
	return out
}

func hasVersion(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '/' && s[i+1] >= '0' && s[i+1] <= '9' {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
