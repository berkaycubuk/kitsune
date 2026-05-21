package security

import (
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// CookiesCheck — audit Set-Cookie attributes for Secure / HttpOnly / SameSite.
type CookiesCheck struct{}

func (CookiesCheck) ID() string       { return "security.cookies" }
func (CookiesCheck) Category() string { return checks.CategorySecurity }

func (CookiesCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie"

	raws := ctx.Headers.Values("Set-Cookie")
	if len(raws) == 0 {
		return nil
	}

	var out []checks.Result
	https := isHTTPS(ctx)
	for _, raw := range raws {
		name := cookieName(raw)
		flags, samesite := cookieAttrs(raw)
		var issues []string
		if https && !flags["secure"] {
			issues = append(issues, "missing Secure")
		}
		if !flags["httponly"] {
			issues = append(issues, "missing HttpOnly")
		}
		if samesite == "" {
			issues = append(issues, "missing SameSite")
		} else if strings.EqualFold(samesite, "none") && !flags["secure"] {
			issues = append(issues, "SameSite=None without Secure")
		}

		if len(issues) == 0 {
			r := checks.NewResult("security.cookies.ok", checks.CategorySecurity, checks.SeverityInfo,
				fmt.Sprintf("Cookie %q has Secure + HttpOnly + SameSite", name))
			r.GuidelineURL = guideline
			out = append(out, r)
			continue
		}
		r := checks.NewResult("security.cookies.flags", checks.CategorySecurity, checks.SeverityWarning,
			fmt.Sprintf("Cookie %q: %s", name, strings.Join(issues, ", ")))
		r.Recommendation = "Set Secure (on HTTPS), HttpOnly (for session cookies), and an explicit SameSite policy."
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	return out
}

func cookieName(raw string) string {
	if i := strings.Index(raw, "="); i > 0 {
		return strings.TrimSpace(raw[:i])
	}
	return raw
}

// cookieAttrs returns presence flags for boolean attributes and the SameSite value.
func cookieAttrs(raw string) (flags map[string]bool, samesite string) {
	flags = map[string]bool{}
	parts := strings.Split(raw, ";")
	for i, p := range parts {
		if i == 0 {
			continue // name=value
		}
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.IndexByte(p, '=')
		var key, val string
		if eq < 0 {
			key = strings.ToLower(p)
		} else {
			key = strings.ToLower(strings.TrimSpace(p[:eq]))
			val = strings.TrimSpace(p[eq+1:])
		}
		switch key {
		case "secure", "httponly":
			flags[key] = true
		case "samesite":
			samesite = val
		}
	}
	return flags, samesite
}
