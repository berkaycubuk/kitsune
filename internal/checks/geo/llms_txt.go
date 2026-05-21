package geo

import (
	"fmt"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

type LLMsTxtCheck struct{}

func (LLMsTxtCheck) ID() string       { return "geo.llms_txt" }
func (LLMsTxtCheck) Category() string { return "geo" }

func (LLMsTxtCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://llmstxt.org/"
	var results []checks.Result

	if ctx.LLMs.Found() {
		r := checks.NewResult("geo.llms_txt.present", "geo", checks.SeverityInfo,
			fmt.Sprintf("/llms.txt present (%d bytes)", len(ctx.LLMs.Body)))
		r.GuidelineURL = guideline
		results = append(results, r)
	} else {
		r := checks.NewResult("geo.llms_txt.absent", "geo", checks.SeverityInfo,
			"/llms.txt not found")
		r.Detail = "An /llms.txt file gives LLMs a curated entry point to your site's most important content."
		r.Recommendation = "Consider adding /llms.txt per the llmstxt.org spec to improve discoverability for AI assistants."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	if ctx.LLMsFull.Found() {
		r := checks.NewResult("geo.llms_full_txt.present", "geo", checks.SeverityInfo,
			fmt.Sprintf("/llms-full.txt present (%d bytes)", len(ctx.LLMsFull.Body)))
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	return results
}
