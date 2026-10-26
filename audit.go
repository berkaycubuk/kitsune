// Package kitsune audits a URL from Go, for programs that embed kitsune
// instead of running its binary. It gives the same findings as
// `kitsune --json`, in types of its own, so callers don't depend on
// kitsune's internals.
package kitsune

import (
	"time"

	"github.com/berkaycubuk/kitsune/internal/runner"
)

// Finding is one result of a check. ID is stable across versions, so
// two audits of a site can be compared by it. Severity is "info",
// "warning" or "error".
type Finding struct {
	ID             string `json:"id"`
	Category       string `json:"category"`
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Detail         string `json:"detail,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}

type Report struct {
	URL        string    `json:"url"`
	FinalURL   string    `json:"final_url"`
	StatusCode int       `json:"status_code"`
	Findings   []Finding `json:"findings"`
}

// Audit fetches url, a full URL with its scheme, and runs every check on
// it. An error means the page couldn't be fetched at all.
func Audit(url string, timeout time.Duration) (*Report, error) {
	r, err := runner.Run(url, runner.Options{Timeout: timeout})
	if err != nil {
		return nil, err
	}
	out := &Report{URL: r.URL, FinalURL: r.FinalURL, StatusCode: r.StatusCode}
	for _, res := range r.Results {
		out.Findings = append(out.Findings, Finding{
			ID:             res.ID,
			Category:       res.Category,
			Severity:       res.Severity.String(),
			Title:          res.Title,
			Detail:         res.Detail,
			Recommendation: res.Recommendation,
		})
	}
	return out, nil
}
