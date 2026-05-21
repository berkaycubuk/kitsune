package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/runner"
	"github.com/charmbracelet/lipgloss"
)

// orderedCategories returns the categories present in `grouped`, sorted by the
// canonical KnownCategories() order with any unknown values appended alphabetically.
func orderedCategories(grouped map[string][]checks.Result) []string {
	var out []string
	seen := map[string]bool{}
	for _, k := range checks.KnownCategories() {
		if _, ok := grouped[k]; ok {
			out = append(out, k)
			seen[k] = true
		}
	}
	var extras []string
	for c := range grouped {
		if !seen[c] {
			extras = append(extras, c)
		}
	}
	sort.Strings(extras)
	return append(out, extras...)
}

var (
	styleHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleCat     = lipgloss.NewStyle().Bold(true)
)

func WriteTerminal(w io.Writer, r *runner.Report) {
	fmt.Fprintln(w, styleHeader.Render("kitsune"))
	fmt.Fprintf(w, "  URL:        %s\n", r.URL)
	if r.FinalURL != r.URL {
		fmt.Fprintf(w, "  Final URL:  %s\n", r.FinalURL)
	}
	fmt.Fprintf(w, "  HTTP:       %d\n", r.StatusCode)
	fmt.Fprintf(w, "  Fetched:    %s\n\n", r.FetchedAt.Format("2006-01-02 15:04:05 MST"))

	grouped := map[string][]checks.Result{}
	for _, res := range r.Results {
		grouped[res.Category] = append(grouped[res.Category], res)
	}
	cats := orderedCategories(grouped)

	for _, cat := range cats {
		fmt.Fprintln(w, styleCat.Render(catLabel(cat)))
		for _, res := range grouped[cat] {
			fmt.Fprintln(w, renderResult(res))
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, styleHeader.Render("Summary"))
	fmt.Fprintf(w, "  %s   %d\n", styleInfo.Render("info "), r.Summary.Info)
	fmt.Fprintf(w, "  %s   %d\n", styleWarning.Render("warn "), r.Summary.Warning)
	fmt.Fprintf(w, "  %s   %d\n", styleError.Render("error"), r.Summary.Error)
}

func catLabel(c string) string {
	switch c {
	case checks.CategorySEO:
		return "▌SEO"
	case checks.CategoryGEO:
		return "▌GEO"
	case checks.CategoryPerf:
		return "▌Performance"
	case checks.CategoryA11y:
		return "▌Accessibility"
	case checks.CategorySecurity:
		return "▌Security"
	default:
		return "▌" + c
	}
}

func renderResult(res checks.Result) string {
	icon, style := iconAndStyle(res.Severity)
	line := fmt.Sprintf("  %s %s", style.Render(icon), res.Title)
	if res.Detail != "" {
		line += "\n    " + styleDim.Render(res.Detail)
	}
	if res.Recommendation != "" {
		line += "\n    " + styleDim.Render("→ "+res.Recommendation)
	}
	if res.GuidelineURL != "" {
		line += "\n    " + styleDim.Render(res.GuidelineURL)
	}
	return line
}

func iconAndStyle(s checks.Severity) (string, lipgloss.Style) {
	switch s {
	case checks.SeverityInfo:
		return "•", styleInfo
	case checks.SeverityWarning:
		return "!", styleWarning
	case checks.SeverityError:
		return "✗", styleError
	}
	return "?", styleDim
}
