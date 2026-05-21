package geo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

type AuthorDateCheck struct{}

func (AuthorDateCheck) ID() string       { return "geo.author_date" }
func (AuthorDateCheck) Category() string { return "geo" }

func (AuthorDateCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://developers.google.com/search/docs/appearance/publication-dates"

	author := strings.TrimSpace(ctx.Doc.Find(`meta[name="author"]`).AttrOr("content", ""))
	published := strings.TrimSpace(ctx.Doc.Find(`meta[property="article:published_time"]`).AttrOr("content", ""))
	modified := strings.TrimSpace(ctx.Doc.Find(`meta[property="article:modified_time"]`).AttrOr("content", ""))

	var results []checks.Result
	if author == "" {
		r := checks.NewResult("geo.author.missing", "geo", checks.SeverityInfo,
			"No author signal in metadata")
		r.Recommendation = `Add <meta name="author"> or schema.org "author" — strengthens E-E-A-T for AI citations.`
		r.GuidelineURL = "https://developers.google.com/search/docs/fundamentals/creating-helpful-content"
		results = append(results, r)
	} else {
		r := checks.NewResult("geo.author.present", "geo", checks.SeverityInfo,
			fmt.Sprintf("Author: %s", author))
		results = append(results, r)
	}

	if published == "" && modified == "" {
		r := checks.NewResult("geo.date.missing", "geo", checks.SeverityInfo,
			"No published/modified date metadata")
		r.Recommendation = "Add article:published_time and article:modified_time meta tags; freshness affects AI citation rates."
		r.GuidelineURL = guideline
		results = append(results, r)
	} else {
		summary := []string{}
		if published != "" {
			summary = append(summary, "published="+published)
		}
		if modified != "" {
			summary = append(summary, "modified="+modified)
		}
		r := checks.NewResult("geo.date.present", "geo", checks.SeverityInfo,
			"Date metadata: "+strings.Join(summary, ", "))
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}

type ContentDepthCheck struct{}

func (ContentDepthCheck) ID() string       { return "geo.content_depth" }
func (ContentDepthCheck) Category() string { return "geo" }

var wordSplit = regexp.MustCompile(`\s+`)
var sentenceWithNumber = regexp.MustCompile(`\b\d[\d,.]*\s*(%|percent|million|billion|thousand|users|customers|years?|months?|days?|hours?|x)?\b`)
var questionHeading = regexp.MustCompile(`(?i)^(how|what|why|when|where|which|who|can|does|do|is|are|should)\b.*\?$`)

func (ContentDepthCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://arxiv.org/abs/2311.09735" // Princeton GEO paper

	// Strip script/style before counting.
	clone := ctx.Doc.Clone()
	clone.Find("script, style, nav, header, footer, aside").Remove()
	text := strings.TrimSpace(clone.Find("body").Text())
	text = wordSplit.ReplaceAllString(text, " ")
	words := 0
	if text != "" {
		words = len(strings.Fields(text))
	}

	stats := len(sentenceWithNumber.FindAllString(text, -1))

	var qHeadings, totalHeadings int
	clone.Find("h2, h3, h4").Each(func(_ int, s *goquery.Selection) {
		totalHeadings++
		t := strings.TrimSpace(s.Text())
		if questionHeading.MatchString(t) {
			qHeadings++
		}
	})

	var results []checks.Result
	wInfo := checks.NewResult("geo.content_depth.words", "geo", checks.SeverityInfo,
		fmt.Sprintf("Body word count: %d", words))
	wInfo.GuidelineURL = guideline
	results = append(results, wInfo)

	if words < 300 {
		r := checks.NewResult("geo.content_depth.thin", "geo", checks.SeverityWarning,
			"Page content looks thin")
		r.Detail = "AI search engines tend to cite substantive, in-depth pages."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	r := checks.NewResult("geo.content_depth.stats", "geo", checks.SeverityInfo,
		fmt.Sprintf("Numeric/statistical mentions: %d", stats))
	r.Detail = "GEO research finds statistics + citations significantly boost AI citation rates."
	r.GuidelineURL = guideline
	results = append(results, r)

	if totalHeadings > 0 {
		r := checks.NewResult("geo.content_depth.qa_headings", "geo", checks.SeverityInfo,
			fmt.Sprintf("Question-style headings: %d of %d", qHeadings, totalHeadings))
		r.Detail = "Question-shaped headings improve match against natural-language AI queries."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}
