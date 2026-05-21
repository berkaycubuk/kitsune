package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// Resource fetches a site-level resource at the given path (e.g. "/robots.txt")
// relative to the host of baseURL. Returns a SiteResource with FetchErr populated
// on transport-level failure; HTTP errors are reflected via StatusCode.
func Resource(baseURL, path string, opts Options) *checks.SiteResource {
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return &checks.SiteResource{FetchErr: err.Error()}
	}
	target := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: path}

	client := &http.Client{Timeout: opts.Timeout}
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return &checks.SiteResource{URL: target.String(), FetchErr: err.Error()}
	}
	req.Header.Set("User-Agent", opts.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return &checks.SiteResource{URL: target.String(), FetchErr: err.Error()}
	}
	defer resp.Body.Close()

	const maxBytes = 2 << 20 // 2 MiB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return &checks.SiteResource{URL: target.String(), StatusCode: resp.StatusCode, FetchErr: fmt.Errorf("read: %w", err).Error()}
	}
	return &checks.SiteResource{
		URL:        target.String(),
		StatusCode: resp.StatusCode,
		Body:       body,
	}
}
