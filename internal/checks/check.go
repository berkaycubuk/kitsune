package checks

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

func ParseSeverity(s string) (Severity, bool) {
	switch s {
	case "info":
		return SeverityInfo, true
	case "warning":
		return SeverityWarning, true
	case "error":
		return SeverityError, true
	}
	return 0, false
}

type Result struct {
	ID             string   `json:"id"`
	Category       string   `json:"category"`
	Severity       Severity `json:"-"`
	SeverityLabel  string   `json:"severity"`
	Title          string   `json:"title"`
	Detail         string   `json:"detail,omitempty"`
	Recommendation string   `json:"recommendation,omitempty"`
	GuidelineURL   string   `json:"guideline_url,omitempty"`
}

func NewResult(id, category string, sev Severity, title string) Result {
	return Result{
		ID:            id,
		Category:      category,
		Severity:      sev,
		SeverityLabel: sev.String(),
		Title:         title,
	}
}

type PageContext struct {
	RequestedURL string
	FinalURL     string
	StatusCode   int
	Proto        string // e.g. "HTTP/1.1", "HTTP/2.0"
	Headers      http.Header
	HTML         []byte
	Doc          *goquery.Document
	Redirects    []string

	// Site-level resources, populated by the runner. Nil if not fetched / not found.
	Robots *SiteResource
	LLMs   *SiteResource
	LLMsFull *SiteResource
}

// SiteResource represents a root-level file like robots.txt or llms.txt.
type SiteResource struct {
	URL        string
	StatusCode int
	Body       []byte
	FetchErr   string
}

func (r *SiteResource) Found() bool {
	return r != nil && r.FetchErr == "" && r.StatusCode >= 200 && r.StatusCode < 300
}

type Check interface {
	ID() string
	Category() string
	Run(ctx *PageContext) []Result
}
