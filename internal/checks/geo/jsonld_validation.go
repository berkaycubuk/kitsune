package geo

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// Required field map per schema.org @type. Keeps to the highest-leverage types
// for AI search; a longer list adds noise without proportional value.
var jsonldRequired = map[string][]string{
	"Article":      {"headline", "author", "datePublished"},
	"BlogPosting":  {"headline", "author", "datePublished"},
	"NewsArticle":  {"headline", "author", "datePublished"},
	"Product":      {"name", "image"},
	"Person":       {"name"},
	"Organization": {"name", "url"},
	"WebSite":      {"name", "url"},
	"BreadcrumbList": {"itemListElement"},
	"HowTo":        {"name", "step"},
	"Recipe":       {"name", "recipeIngredient", "recipeInstructions"},
	"FAQPage":      {"mainEntity"},
	"Event":        {"name", "startDate", "location"},
}

// JSONLDValidationCheck — for each parsed @type, verify expected fields exist.
type JSONLDValidationCheck struct{}

func (JSONLDValidationCheck) ID() string       { return "geo.jsonld_validation" }
func (JSONLDValidationCheck) Category() string { return checks.CategoryGEO }

func (JSONLDValidationCheck) Run(ctx *checks.PageContext) []checks.Result {
	missing := map[string]map[string]int{} // type → field → count

	ctx.Doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return
		}
		validateNode(v, missing)
	})

	if len(missing) == 0 {
		return nil
	}
	types := make([]string, 0, len(missing))
	for t := range missing {
		types = append(types, t)
	}
	sort.Strings(types)

	var out []checks.Result
	for _, t := range types {
		fields := make([]string, 0, len(missing[t]))
		for f := range missing[t] {
			fields = append(fields, f)
		}
		sort.Strings(fields)
		r := checks.NewResult("geo.jsonld_validation."+strings.ToLower(t), checks.CategoryGEO, checks.SeverityWarning,
			fmt.Sprintf("JSON-LD %s missing required field(s): %s", t, strings.Join(fields, ", ")))
		r.Recommendation = "Fill in the missing schema.org fields to maximize crawler/AI understanding."
		r.GuidelineURL = "https://schema.org/" + t
		out = append(out, r)
	}
	return out
}

func validateNode(v any, missing map[string]map[string]int) {
	switch x := v.(type) {
	case map[string]any:
		var typeNames []string
		if t, ok := x["@type"]; ok {
			switch tv := t.(type) {
			case string:
				typeNames = []string{tv}
			case []any:
				for _, it := range tv {
					if s, ok := it.(string); ok {
						typeNames = append(typeNames, s)
					}
				}
			}
		}
		for _, t := range typeNames {
			for _, field := range jsonldRequired[t] {
				if _, ok := x[field]; !ok {
					if missing[t] == nil {
						missing[t] = map[string]int{}
					}
					missing[t][field]++
				}
			}
		}
		for _, val := range x {
			validateNode(val, missing)
		}
	case []any:
		for _, it := range x {
			validateNode(it, missing)
		}
	}
}
