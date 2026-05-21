package a11y

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

func hasAccessibleName(s *goquery.Selection) bool {
	if strings.TrimSpace(s.Text()) != "" {
		return true
	}
	for _, attr := range []string{"aria-label", "aria-labelledby", "title"} {
		if strings.TrimSpace(s.AttrOr(attr, "")) != "" {
			return true
		}
	}
	// <button> with an <img alt> or <svg><title> child counts.
	if s.Find("img[alt]").FilterFunction(func(_ int, img *goquery.Selection) bool {
		return strings.TrimSpace(img.AttrOr("alt", "")) != ""
	}).Length() > 0 {
		return true
	}
	if s.Find("svg > title").Length() > 0 {
		return true
	}
	return false
}

// ButtonLinkNameCheck — <button> and <a> must have an accessible name.
type ButtonLinkNameCheck struct{}

func (ButtonLinkNameCheck) ID() string       { return "a11y.button_link_name" }
func (ButtonLinkNameCheck) Category() string { return checks.CategoryA11y }

func (ButtonLinkNameCheck) Run(ctx *checks.PageContext) []checks.Result {
	var noNameBtn, noNameLink int
	ctx.Doc.Find("button").Each(func(_ int, s *goquery.Selection) {
		if !hasAccessibleName(s) {
			noNameBtn++
		}
	})
	ctx.Doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		if !hasAccessibleName(s) {
			noNameLink++
		}
	})
	var out []checks.Result
	if noNameBtn > 0 {
		r := checks.NewResult("a11y.button_link_name.button", checks.CategoryA11y, checks.SeverityWarning,
			fmt.Sprintf("%d <button> without an accessible name", noNameBtn))
		r.Recommendation = "Add visible text, `aria-label`, or `aria-labelledby` — common gap for icon-only buttons."
		r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Understanding/name-role-value.html"
		out = append(out, r)
	}
	if noNameLink > 0 {
		r := checks.NewResult("a11y.button_link_name.link", checks.CategoryA11y, checks.SeverityWarning,
			fmt.Sprintf("%d <a href> without accessible link text", noNameLink))
		r.Recommendation = "Add descriptive link text or an `aria-label`."
		r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Understanding/link-purpose-in-context.html"
		out = append(out, r)
	}
	return out
}

// IframeTitleCheck — every <iframe> needs a non-empty title.
type IframeTitleCheck struct{}

func (IframeTitleCheck) ID() string       { return "a11y.iframe_title" }
func (IframeTitleCheck) Category() string { return checks.CategoryA11y }

func (IframeTitleCheck) Run(ctx *checks.PageContext) []checks.Result {
	var missing int
	ctx.Doc.Find("iframe").Each(func(_ int, s *goquery.Selection) {
		if strings.TrimSpace(s.AttrOr("title", "")) == "" &&
			strings.TrimSpace(s.AttrOr("aria-label", "")) == "" {
			missing++
		}
	})
	if missing == 0 {
		return nil
	}
	r := checks.NewResult("a11y.iframe_title.missing", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d <iframe> missing title/aria-label", missing))
	r.Recommendation = "Add a `title` attribute describing the embedded content."
	r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Techniques/html/H64"
	return []checks.Result{r}
}

// ObjectEmbedNameCheck — <object>/<embed> need an accessible name.
type ObjectEmbedNameCheck struct{}

func (ObjectEmbedNameCheck) ID() string       { return "a11y.object_embed_name" }
func (ObjectEmbedNameCheck) Category() string { return checks.CategoryA11y }

func (ObjectEmbedNameCheck) Run(ctx *checks.PageContext) []checks.Result {
	var missing int
	ctx.Doc.Find("object, embed").Each(func(_ int, s *goquery.Selection) {
		if !hasAccessibleName(s) {
			missing++
		}
	})
	if missing == 0 {
		return nil
	}
	r := checks.NewResult("a11y.object_embed_name.missing", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d <object>/<embed> without accessible name", missing))
	r.Recommendation = "Provide `title`, `aria-label`, or fallback content."
	r.GuidelineURL = "https://www.w3.org/WAI/WCAG21/Understanding/non-text-content.html"
	return []checks.Result{r}
}
