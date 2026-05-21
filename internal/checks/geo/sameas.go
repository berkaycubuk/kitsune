package geo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// SameAsCheck — surface schema.org `sameAs` URLs from Person/Organization blocks
// as an E-E-A-T signal.
type SameAsCheck struct{}

func (SameAsCheck) ID() string       { return "geo.sameas" }
func (SameAsCheck) Category() string { return checks.CategoryGEO }

func (SameAsCheck) Run(ctx *checks.PageContext) []checks.Result {
	urls := map[string]bool{}
	ctx.Doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		var v any
		if err := json.Unmarshal([]byte(strings.TrimSpace(s.Text())), &v); err != nil {
			return
		}
		collectSameAs(v, urls)
	})
	if len(urls) == 0 {
		return nil
	}
	domains := map[string]bool{}
	for u := range urls {
		if parsed, err := url.Parse(u); err == nil && parsed.Host != "" {
			domains[parsed.Host] = true
		}
	}
	list := make([]string, 0, len(domains))
	for d := range domains {
		list = append(list, d)
	}
	sort.Strings(list)
	r := checks.NewResult("geo.sameas.present", checks.CategoryGEO, checks.SeverityInfo,
		fmt.Sprintf("schema.org sameAs: %d URL(s) across %d domain(s)", len(urls), len(domains)))
	r.Detail = strings.Join(list, ", ")
	r.GuidelineURL = "https://schema.org/sameAs"
	return []checks.Result{r}
}

func collectSameAs(v any, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		if sa, ok := x["sameAs"]; ok {
			switch sv := sa.(type) {
			case string:
				out[sv] = true
			case []any:
				for _, it := range sv {
					if s, ok := it.(string); ok {
						out[s] = true
					}
				}
			}
		}
		for _, val := range x {
			collectSameAs(val, out)
		}
	case []any:
		for _, it := range x {
			collectSameAs(it, out)
		}
	}
}
