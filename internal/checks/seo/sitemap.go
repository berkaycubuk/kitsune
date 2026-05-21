package seo

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// SitemapCheck — fetch and validate the sitemap declared in robots.txt
// (falling back to /sitemap.xml).
type SitemapCheck struct{}

func (SitemapCheck) ID() string       { return "seo.sitemap" }
func (SitemapCheck) Category() string { return checks.CategorySEO }

func (SitemapCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developers.google.com/search/docs/crawling-indexing/sitemaps/overview"

	urls := sitemapURLsFromRobots(ctx)
	if len(urls) == 0 {
		// Fall back to the conventional location.
		if base, err := url.Parse(ctx.FinalURL); err == nil {
			urls = []string{(&url.URL{Scheme: base.Scheme, Host: base.Host, Path: "/sitemap.xml"}).String()}
		}
	}
	if len(urls) == 0 {
		return nil
	}

	var out []checks.Result
	for _, u := range urls {
		body, status, err := httpGet(u)
		if err != nil {
			r := checks.NewResult("seo.sitemap.unreachable", checks.CategorySEO, checks.SeverityWarning,
				fmt.Sprintf("Sitemap unreachable: %s", u))
			r.Detail = err.Error()
			r.GuidelineURL = guideline
			out = append(out, r)
			continue
		}
		if status < 200 || status >= 300 {
			r := checks.NewResult("seo.sitemap.bad_status", checks.CategorySEO, checks.SeverityWarning,
				fmt.Sprintf("Sitemap returned HTTP %d: %s", status, u))
			r.GuidelineURL = guideline
			out = append(out, r)
			continue
		}
		root, count, err := parseSitemap(body)
		if err != nil {
			r := checks.NewResult("seo.sitemap.invalid", checks.CategorySEO, checks.SeverityWarning,
				fmt.Sprintf("Sitemap failed to parse: %s", u))
			r.Detail = err.Error()
			r.GuidelineURL = guideline
			out = append(out, r)
			continue
		}
		r := checks.NewResult("seo.sitemap.ok", checks.CategorySEO, checks.SeverityInfo,
			fmt.Sprintf("Sitemap %s: %s with %d entry(ies)", u, root, count))
		r.GuidelineURL = guideline
		out = append(out, r)
	}
	return out
}

func sitemapURLsFromRobots(ctx *checks.PageContext) []string {
	if ctx.Robots == nil || !ctx.Robots.Found() {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(ctx.Robots.Body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "sitemap:") {
			out = append(out, strings.TrimSpace(line[len("sitemap:"):]))
		}
	}
	return out
}

func httpGet(rawURL string) ([]byte, int, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "kitsune/0.1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

type sitemapURLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URL     []struct{} `xml:"url"`
}

type sitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Sitemap []struct{} `xml:"sitemap"`
}

func parseSitemap(body []byte) (root string, count int, err error) {
	dec := xml.NewDecoder(strings.NewReader(string(body)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return "", 0, fmt.Errorf("no root element found")
		}
		if err != nil {
			return "", 0, err
		}
		if se, ok := tok.(xml.StartElement); ok {
			switch se.Name.Local {
			case "urlset":
				var us sitemapURLSet
				if err := dec.DecodeElement(&us, &se); err != nil {
					return "", 0, err
				}
				return "<urlset>", len(us.URL), nil
			case "sitemapindex":
				var si sitemapIndex
				if err := dec.DecodeElement(&si, &se); err != nil {
					return "", 0, err
				}
				return "<sitemapindex>", len(si.Sitemap), nil
			default:
				return "", 0, fmt.Errorf("unexpected root element <%s>", se.Name.Local)
			}
		}
	}
}
