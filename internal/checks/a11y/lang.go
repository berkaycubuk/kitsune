package a11y

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// LangValidityCheck — validates lang / xml:lang / hreflang values against a
// permissive BCP-47 subset (primary language + optional region).
type LangValidityCheck struct{}

func (LangValidityCheck) ID() string       { return "a11y.lang_validity" }
func (LangValidityCheck) Category() string { return checks.CategoryA11y }

// Permissive BCP-47: 2-3 alpha primary, optional script/region subtags. Good
// enough to catch typos without shipping a full IANA registry.
var bcp47 = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z]{4})?(-([A-Za-z]{2}|[0-9]{3}))?(-[A-Za-z0-9]{5,8})*$`)

// IsBCP47 is exported so the SEO hreflang check can reuse the same validator.
func IsBCP47(s string) bool { return bcp47.MatchString(s) }

func (LangValidityCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://www.w3.org/International/articles/language-tags/"

	var bad []string
	check := func(label, v string) {
		v = strings.TrimSpace(v)
		if v == "" || IsBCP47(v) {
			return
		}
		bad = append(bad, fmt.Sprintf("%s=%q", label, v))
	}

	check("html lang", ctx.Doc.Find("html").AttrOr("lang", ""))
	check("html xml:lang", ctx.Doc.Find("html").AttrOr("xml:lang", ""))
	ctx.Doc.Find(`link[rel="alternate"][hreflang]`).Each(func(_ int, s *goquery.Selection) {
		v := s.AttrOr("hreflang", "")
		if v == "x-default" {
			return
		}
		check("hreflang", v)
	})

	if len(bad) == 0 {
		return nil
	}
	r := checks.NewResult("a11y.lang_validity.invalid", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d invalid language tag(s)", len(bad)))
	r.Detail = strings.Join(bad, ", ")
	r.Recommendation = "Use BCP-47 tags (e.g. `en`, `en-US`, `zh-Hant`)."
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
