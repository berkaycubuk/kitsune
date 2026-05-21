package geo

import (
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

// MicrodataCheck — detect Microdata (itemscope/itemtype) blocks.
type MicrodataCheck struct{}

func (MicrodataCheck) ID() string       { return "geo.microdata" }
func (MicrodataCheck) Category() string { return checks.CategoryGEO }

func (MicrodataCheck) Run(ctx *checks.PageContext) []checks.Result {
	types := map[string]int{}
	count := 0
	ctx.Doc.Find("[itemscope]").Each(func(_ int, s *goquery.Selection) {
		count++
		t := strings.TrimSpace(s.AttrOr("itemtype", ""))
		if t != "" {
			types[t]++
		}
	})
	if count == 0 {
		return nil
	}
	list := make([]string, 0, len(types))
	for t := range types {
		list = append(list, t)
	}
	sort.Strings(list)
	r := checks.NewResult("geo.microdata.present", checks.CategoryGEO, checks.SeverityInfo,
		fmt.Sprintf("Microdata: %d itemscope block(s)", count))
	if len(list) > 0 {
		r.Detail = "types: " + strings.Join(list, ", ")
	}
	r.GuidelineURL = "https://schema.org/docs/gs.html#microdata_how"
	return []checks.Result{r}
}

// RDFaCheck — detect RDFa attributes (typeof/vocab).
type RDFaCheck struct{}

func (RDFaCheck) ID() string       { return "geo.rdfa" }
func (RDFaCheck) Category() string { return checks.CategoryGEO }

func (RDFaCheck) Run(ctx *checks.PageContext) []checks.Result {
	types := map[string]int{}
	ctx.Doc.Find("[typeof]").Each(func(_ int, s *goquery.Selection) {
		for _, t := range strings.Fields(s.AttrOr("typeof", "")) {
			types[t]++
		}
	})
	// Exclude og:* / fb:* / twitter:* meta tags — those are OGP not page-level RDFa.
	properties := 0
	ctx.Doc.Find("[property]").Each(func(_ int, s *goquery.Selection) {
		p := strings.ToLower(strings.TrimSpace(s.AttrOr("property", "")))
		if strings.HasPrefix(p, "og:") || strings.HasPrefix(p, "fb:") || strings.HasPrefix(p, "twitter:") {
			return
		}
		properties++
	})
	hasVocab := ctx.Doc.Find("[vocab]").Length() > 0

	if len(types) == 0 && properties == 0 && !hasVocab {
		return nil
	}
	list := make([]string, 0, len(types))
	for t := range types {
		list = append(list, t)
	}
	sort.Strings(list)
	r := checks.NewResult("geo.rdfa.present", checks.CategoryGEO, checks.SeverityInfo,
		fmt.Sprintf("RDFa: %d typeof block(s), %d property attribute(s)", len(types), properties))
	if len(list) > 0 {
		r.Detail = "types: " + strings.Join(list, ", ")
	}
	r.GuidelineURL = "https://www.w3.org/TR/rdfa-primer/"
	return []checks.Result{r}
}
