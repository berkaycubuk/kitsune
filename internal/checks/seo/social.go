package seo

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type OpenGraphCheck struct{}

func (OpenGraphCheck) ID() string       { return "seo.open_graph" }
func (OpenGraphCheck) Category() string { return "seo" }

func (OpenGraphCheck) Run(ctx *checks.PageContext) []checks.Result {
	required := []string{"og:title", "og:description", "og:image", "og:url", "og:type"}
	found := map[string]string{}
	ctx.Doc.Find(`head > meta[property^="og:"]`).Each(func(_ int, s *goquery.Selection) {
		prop := strings.TrimSpace(s.AttrOr("property", ""))
		content := strings.TrimSpace(s.AttrOr("content", ""))
		if prop != "" && content != "" {
			found[prop] = content
		}
	})

	guideline := "https://ogp.me/"
	var missing []string
	for _, k := range required {
		if found[k] == "" {
			missing = append(missing, k)
		}
	}

	if len(found) == 0 {
		r := checks.NewResult("seo.open_graph.absent", "seo", checks.SeverityWarning,
			"No Open Graph tags found")
		r.Recommendation = "Add og:title, og:description, og:image, og:url, og:type for richer social previews."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	var results []checks.Result
	info := checks.NewResult("seo.open_graph.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Open Graph: %d tag(s) found", len(found)))
	info.GuidelineURL = guideline
	results = append(results, info)

	if len(missing) > 0 {
		r := checks.NewResult("seo.open_graph.incomplete", "seo", checks.SeverityWarning,
			"Open Graph tag set is incomplete")
		r.Detail = "Missing: " + strings.Join(missing, ", ")
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}

type TwitterCardCheck struct{}

func (TwitterCardCheck) ID() string       { return "seo.twitter_card" }
func (TwitterCardCheck) Category() string { return "seo" }

func (TwitterCardCheck) Run(ctx *checks.PageContext) []checks.Result {
	card := strings.TrimSpace(ctx.Doc.Find(`head > meta[name="twitter:card"]`).AttrOr("content", ""))
	guideline := "https://developer.x.com/en/docs/twitter-for-websites/cards/overview/abouts-cards"

	if card == "" {
		r := checks.NewResult("seo.twitter_card.absent", "seo", checks.SeverityInfo,
			"No Twitter Card tags found")
		r.Recommendation = `Add <meta name="twitter:card" content="summary_large_image"> plus title/description/image.`
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	r := checks.NewResult("seo.twitter_card.present", "seo", checks.SeverityInfo,
		fmt.Sprintf("Twitter Card: %s", card))
	r.GuidelineURL = guideline
	return []checks.Result{r}
}
