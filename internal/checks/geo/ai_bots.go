package geo

import (
	"fmt"
	"sort"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// Known AI training / search crawlers. Sources: each vendor's published docs.
var knownAIBots = []string{
	"GPTBot",            // OpenAI training crawler
	"ChatGPT-User",      // OpenAI live browsing
	"OAI-SearchBot",     // OpenAI search crawler
	"ClaudeBot",         // Anthropic training crawler
	"Claude-Web",        // Anthropic live browsing
	"PerplexityBot",     // Perplexity crawler
	"Perplexity-User",   // Perplexity live browsing
	"Google-Extended",   // Google Gemini training opt-out
	"Applebot-Extended", // Apple Intelligence training
	"CCBot",             // Common Crawl
	"Bytespider",        // ByteDance / Doubao
	"Amazonbot",         // Amazon
	"Meta-ExternalAgent",
}

type AIBotsCheck struct{}

func (AIBotsCheck) ID() string       { return "geo.ai_bots" }
func (AIBotsCheck) Category() string { return "geo" }

func (AIBotsCheck) Run(ctx *checks.PageContext) []checks.Result {
	guideline := "https://platform.openai.com/docs/bots"
	if ctx.Robots == nil || !ctx.Robots.Found() {
		r := checks.NewResult("geo.ai_bots.no_robots", "geo", checks.SeverityInfo,
			"No robots.txt — AI crawlers will operate under default rules")
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	body := strings.ToLower(string(ctx.Robots.Body))
	allowed, blocked := []string{}, []string{}
	for _, bot := range knownAIBots {
		lowerBot := strings.ToLower(bot)
		if !strings.Contains(body, "user-agent: "+lowerBot) {
			continue
		}
		if isBotBlocked(string(ctx.Robots.Body), bot) {
			blocked = append(blocked, bot)
		} else {
			allowed = append(allowed, bot)
		}
	}

	sort.Strings(allowed)
	sort.Strings(blocked)

	var results []checks.Result
	if len(allowed) == 0 && len(blocked) == 0 {
		r := checks.NewResult("geo.ai_bots.unspecified", "geo", checks.SeverityInfo,
			"No explicit rules for known AI crawlers in robots.txt")
		r.Detail = "Without rules, most AI crawlers will fall back to the wildcard user-agent block (or none)."
		r.Recommendation = "Decide explicitly whether you want to be cited by AI search; add rules for GPTBot, ClaudeBot, PerplexityBot, Google-Extended, OAI-SearchBot."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	if len(blocked) > 0 {
		r := checks.NewResult("geo.ai_bots.blocked", "geo", checks.SeverityWarning,
			fmt.Sprintf("%d AI crawler(s) blocked", len(blocked)))
		r.Detail = "Blocked: " + strings.Join(blocked, ", ")
		r.Recommendation = "These crawlers won't index your content for AI answers; verify this is intentional."
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	if len(allowed) > 0 {
		r := checks.NewResult("geo.ai_bots.allowed", "geo", checks.SeverityInfo,
			fmt.Sprintf("%d AI crawler(s) explicitly allowed", len(allowed)))
		r.Detail = "Allowed: " + strings.Join(allowed, ", ")
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	return results
}

// isBotBlocked walks groups in robots.txt and returns true if the named bot
// has a top-level "Disallow: /" applied to it. Simple heuristic, not a full parser.
func isBotBlocked(body, bot string) bool {
	lines := strings.Split(body, "\n")
	matched := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if matched {
				return false
			}
			matched = false
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "user-agent:") {
			ua := strings.TrimSpace(line[len("user-agent:"):])
			matched = strings.EqualFold(ua, bot)
			continue
		}
		if matched && strings.HasPrefix(lower, "disallow:") {
			path := strings.TrimSpace(line[len("disallow:"):])
			if path == "/" {
				return true
			}
		}
	}
	return false
}
