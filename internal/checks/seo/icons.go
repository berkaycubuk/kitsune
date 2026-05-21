package seo

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// IconsCheck — favicon, apple-touch-icon, theme-color presence.
type IconsCheck struct{}

func (IconsCheck) ID() string       { return "seo.icons" }
func (IconsCheck) Category() string { return checks.CategorySEO }

func (IconsCheck) Run(ctx *checks.PageContext) []checks.Result {
	hasFavicon := false
	hasAppleIcon := false
	ctx.Doc.Find(`head > link[rel]`).Each(func(_ int, s *goquery.Selection) {
		rel := strings.ToLower(s.AttrOr("rel", ""))
		for _, t := range strings.Fields(rel) {
			switch t {
			case "icon", "shortcut":
				hasFavicon = true
			case "apple-touch-icon", "apple-touch-icon-precomposed":
				hasAppleIcon = true
			}
		}
	})
	hasThemeColor := ctx.Doc.Find(`head > meta[name="theme-color" i]`).Length() > 0

	var out []checks.Result
	if !hasFavicon {
		r := checks.NewResult("seo.icons.no_favicon", checks.CategorySEO, checks.SeverityWarning,
			"No <link rel=\"icon\"> declared")
		r.Recommendation = "Add `<link rel=\"icon\" href=\"/favicon.ico\">` (or a PNG/SVG variant)."
		r.GuidelineURL = "https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel/icon"
		out = append(out, r)
	}
	if !hasAppleIcon {
		r := checks.NewResult("seo.icons.no_apple_touch", checks.CategorySEO, checks.SeverityInfo,
			"No apple-touch-icon declared")
		r.Recommendation = "Add `<link rel=\"apple-touch-icon\" href=\"/apple-touch-icon.png\">` for iOS home-screen shortcuts."
		out = append(out, r)
	}
	if !hasThemeColor {
		r := checks.NewResult("seo.icons.no_theme_color", checks.CategorySEO, checks.SeverityInfo,
			"No <meta name=\"theme-color\"> declared")
		r.Recommendation = "Add `<meta name=\"theme-color\" content=\"#…\">` to colour the mobile browser UI chrome."
		out = append(out, r)
	}
	if len(out) == 0 {
		r := checks.NewResult("seo.icons.ok", checks.CategorySEO, checks.SeverityInfo,
			"favicon + apple-touch-icon + theme-color all present")
		out = append(out, r)
	}
	return out
}

// ManifestCheck — presence of <link rel="manifest">.
type ManifestCheck struct{}

func (ManifestCheck) ID() string       { return "seo.manifest" }
func (ManifestCheck) Category() string { return checks.CategorySEO }

func (ManifestCheck) Run(ctx *checks.PageContext) []checks.Result {
	sel := ctx.Doc.Find(`head > link[rel="manifest" i]`)
	if sel.Length() == 0 {
		return nil
	}
	href := strings.TrimSpace(sel.AttrOr("href", ""))
	r := checks.NewResult("seo.manifest.present", checks.CategorySEO, checks.SeverityInfo,
		"Web App Manifest linked: "+href)
	r.GuidelineURL = "https://developer.mozilla.org/en-US/docs/Web/Manifest"
	return []checks.Result{r}
}
