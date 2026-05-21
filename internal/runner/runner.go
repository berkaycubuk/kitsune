package runner

import (
	"time"

	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/fetch"
	"github.com/berkaycubuk/kitsune/internal/version"
)

type Options struct {
	UserAgent string
	Timeout   time.Duration
	Categories map[string]bool // empty = all
}

type Report struct {
	SchemaVersion string          `json:"schema_version"`
	Tool          string          `json:"tool"`
	ToolVersion   string          `json:"tool_version"`
	URL           string          `json:"url"`
	FinalURL      string          `json:"final_url"`
	StatusCode    int             `json:"status_code"`
	FetchedAt     time.Time       `json:"fetched_at"`
	Results       []checks.Result `json:"results"`
	Summary       Summary         `json:"summary"`
}

type Summary struct {
	Info    int `json:"info"`
	Warning int `json:"warning"`
	Error   int `json:"error"`
}

func Run(url string, opts Options) (*Report, error) {
	fopts := fetch.Options{UserAgent: opts.UserAgent, Timeout: opts.Timeout}
	page, err := fetch.Page(url, fopts)
	if err != nil {
		return nil, err
	}

	// Always attempt to fetch site-level resources; checks decide what to do.
	page.Robots = fetch.Resource(page.FinalURL, "/robots.txt", fopts)
	page.LLMs = fetch.Resource(page.FinalURL, "/llms.txt", fopts)
	page.LLMsFull = fetch.Resource(page.FinalURL, "/llms-full.txt", fopts)

	selected := filter(allChecks(), opts.Categories)
	var all []checks.Result
	for _, c := range selected {
		all = append(all, c.Run(page)...)
	}

	report := &Report{
		SchemaVersion: version.Schema,
		Tool:          "kitsune",
		ToolVersion:   version.Version,
		URL:           page.RequestedURL,
		FinalURL:      page.FinalURL,
		StatusCode:    page.StatusCode,
		FetchedAt:     time.Now().UTC(),
		Results:       all,
	}
	for _, r := range all {
		switch r.Severity {
		case checks.SeverityInfo:
			report.Summary.Info++
		case checks.SeverityWarning:
			report.Summary.Warning++
		case checks.SeverityError:
			report.Summary.Error++
		}
	}
	return report, nil
}
