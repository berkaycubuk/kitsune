package geo

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type JSONLDCheck struct{}

func (JSONLDCheck) ID() string       { return "geo.jsonld" }
func (JSONLDCheck) Category() string { return "geo" }

func (JSONLDCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://schema.org/docs/gs.html"

	types := map[string]int{}
	parseErrors := 0
	blocks := 0
	ctx.Doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		blocks++
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			parseErrors++
			return
		}
		collectTypes(v, types)
	})

	if blocks == 0 {
		r := checks.NewResult("geo.jsonld.absent", "geo", checks.SeverityWarning,
			"No JSON-LD structured data found")
		r.Recommendation = "Add schema.org JSON-LD (Article, FAQPage, HowTo, Organization, etc.). AI search engines heavily favor structured content."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	var typeList []string
	for t := range types {
		typeList = append(typeList, t)
	}
	sort.Strings(typeList)

	var results []checks.Result
	r := checks.NewResult("geo.jsonld.present", "geo", checks.SeverityInfo,
		fmt.Sprintf("JSON-LD: %d block(s), types: %s", blocks, strings.Join(typeList, ", ")))
	r.GuidelineURL = guideline
	results = append(results, r)

	if parseErrors > 0 {
		pe := checks.NewResult("geo.jsonld.parse_error", "geo", checks.SeverityWarning,
			fmt.Sprintf("%d JSON-LD block(s) failed to parse", parseErrors))
		pe.GuidelineURL = guideline
		results = append(results, pe)
	}

	// Recommend an FAQ block if absent — it's a strong GEO signal.
	if _, has := types["FAQPage"]; !has {
		r := checks.NewResult("geo.jsonld.faq_missing", "geo", checks.SeverityInfo,
			"No FAQPage schema present")
		r.Recommendation = "FAQ schema is a high-leverage signal for AI search; consider adding one if the page has Q&A content."
		r.GuidelineURL = "https://developers.google.com/search/docs/appearance/structured-data/faqpage"
		results = append(results, r)
	}

	return results
}

func collectTypes(v any, out map[string]int) {
	switch x := v.(type) {
	case map[string]any:
		if t, ok := x["@type"]; ok {
			switch tv := t.(type) {
			case string:
				out[tv]++
			case []any:
				for _, item := range tv {
					if s, ok := item.(string); ok {
						out[s]++
					}
				}
			}
		}
		if graph, ok := x["@graph"]; ok {
			collectTypes(graph, out)
		}
		for _, val := range x {
			collectTypes(val, out)
		}
	case []any:
		for _, item := range x {
			collectTypes(item, out)
		}
	}
}
