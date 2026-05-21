package fetch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

const DefaultUserAgent = "kitsune/0.1 (+https://github.com/berkaycubuk/kitsune)"

type Options struct {
	UserAgent string
	Timeout   time.Duration
}

func Page(rawURL string, opts Options) (*checks.PageContext, error) {
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	if opts.Timeout == 0 {
		opts.Timeout = 15 * time.Second
	}

	var redirects []string
	client := &http.Client{
		Timeout: opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			redirects = append(redirects, req.URL.String())
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	const maxBytes = 10 << 20 // 10 MiB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(strings.ToLower(ct), "html") {
		// Not fatal — checks can still report; but warn caller via header inspection.
	}

	return &checks.PageContext{
		RequestedURL: rawURL,
		FinalURL:     resp.Request.URL.String(),
		StatusCode:   resp.StatusCode,
		Headers:      resp.Header,
		HTML:         body,
		Doc:          doc,
		Redirects:    redirects,
	}, nil
}
