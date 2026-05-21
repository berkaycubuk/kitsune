package seo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

// DoctypeCheck — first significant bytes should be `<!DOCTYPE html>` (case-insensitive).
type DoctypeCheck struct{}

func (DoctypeCheck) ID() string       { return "seo.doctype" }
func (DoctypeCheck) Category() string { return checks.CategorySEO }

func (DoctypeCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://developer.mozilla.org/en-US/docs/Glossary/Doctype"
	// Strip leading whitespace and an optional UTF-8 BOM (EF BB BF).
	head := bytes.TrimLeft(ctx.HTML, " \t\r\n")
	head = bytes.TrimPrefix(head, []byte{0xEF, 0xBB, 0xBF})
	if !bytes.HasPrefix(bytes.ToLower(head), []byte("<!doctype")) {
		r := checks.NewResult("seo.doctype.missing", checks.CategorySEO, checks.SeverityWarning,
			"Document is missing DOCTYPE declaration")
		r.Recommendation = "Begin the document with `<!DOCTYPE html>` to ensure standards mode."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	// Extract first 80 bytes of the DOCTYPE for the detail.
	end := bytes.IndexByte(head, '>')
	if end < 0 || end > 200 {
		end = 200
	}
	dt := string(head[:end+1])
	if !strings.EqualFold(strings.TrimSpace(dt), "<!doctype html>") {
		r := checks.NewResult("seo.doctype.legacy", checks.CategorySEO, checks.SeverityWarning,
			"Document uses a legacy DOCTYPE")
		r.Detail = dt
		r.Recommendation = "Use the HTML5 doctype: `<!DOCTYPE html>`."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}
	r := checks.NewResult("seo.doctype.html5", checks.CategorySEO, checks.SeverityInfo,
		"HTML5 DOCTYPE present")
	r.GuidelineURL = guideline
	return []checks.Result{r}
}

// CharsetCheck — <meta charset="utf-8"> within first 1024 bytes.
type CharsetCheck struct{}

func (CharsetCheck) ID() string       { return "seo.charset" }
func (CharsetCheck) Category() string { return checks.CategorySEO }

func (CharsetCheck) Run(ctx *checks.PageContext) []checks.Result {
	const guideline = "https://html.spec.whatwg.org/multipage/semantics.html#charset"

	meta := ctx.Doc.Find(`meta[charset], meta[http-equiv="Content-Type" i]`).First()
	if meta.Length() == 0 {
		r := checks.NewResult("seo.charset.missing", checks.CategorySEO, checks.SeverityWarning,
			"No <meta charset> declared")
		r.Recommendation = "Add `<meta charset=\"utf-8\">` as the first child of <head>."
		r.GuidelineURL = guideline
		return []checks.Result{r}
	}

	charset := strings.ToLower(strings.TrimSpace(meta.AttrOr("charset", "")))
	if charset == "" {
		ct := strings.ToLower(meta.AttrOr("content", ""))
		if i := strings.Index(ct, "charset="); i >= 0 {
			charset = strings.TrimSpace(ct[i+len("charset="):])
		}
	}

	var results []checks.Result
	if charset != "utf-8" && charset != "" {
		r := checks.NewResult("seo.charset.non_utf8", checks.CategorySEO, checks.SeverityWarning,
			fmt.Sprintf("Charset is %q, not utf-8", charset))
		r.Recommendation = "Use UTF-8; it's the only encoding HTML5 mandates browsers support universally."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	// Verify the charset declaration is within the first 1024 bytes of the document.
	scan := ctx.HTML
	if len(scan) > 1024 {
		scan = scan[:1024]
	}
	lower := bytes.ToLower(scan)
	if !bytes.Contains(lower, []byte("charset")) {
		r := checks.NewResult("seo.charset.late", checks.CategorySEO, checks.SeverityWarning,
			"<meta charset> appears after the first 1024 bytes")
		r.Recommendation = "Move the charset declaration to the very top of <head> to avoid a re-parse."
		r.GuidelineURL = guideline
		results = append(results, r)
	}

	if len(results) == 0 {
		r := checks.NewResult("seo.charset.ok", checks.CategorySEO, checks.SeverityInfo,
			"Charset declared (utf-8)")
		r.GuidelineURL = guideline
		results = append(results, r)
	}
	return results
}

// BaseHrefCheck — warn when <base href> is present and changes origin.
type BaseHrefCheck struct{}

func (BaseHrefCheck) ID() string       { return "seo.base_href" }
func (BaseHrefCheck) Category() string { return checks.CategorySEO }

func (BaseHrefCheck) Run(ctx *checks.PageContext) []checks.Result {
	sel := ctx.Doc.Find("head > base[href]")
	if sel.Length() == 0 {
		return nil
	}
	href := strings.TrimSpace(sel.AttrOr("href", ""))
	r := checks.NewResult("seo.base_href.present", checks.CategorySEO, checks.SeverityInfo,
		fmt.Sprintf("<base href=\"%s\"> present", href))
	r.Recommendation = "<base> changes the resolution origin for every relative URL — easy to misuse; double-check the intent."
	r.GuidelineURL = "https://developer.mozilla.org/en-US/docs/Web/HTML/Element/base"
	return []checks.Result{r}
}

// DuplicateMetaCheck — multiple <title>, description, or canonical = warning.
type DuplicateMetaCheck struct{}

func (DuplicateMetaCheck) ID() string       { return "seo.duplicate_meta" }
func (DuplicateMetaCheck) Category() string { return checks.CategorySEO }

func (DuplicateMetaCheck) Run(ctx *checks.PageContext) []checks.Result {
	var out []checks.Result
	pairs := []struct {
		sel, label, id string
	}{
		{"title", "<title>", "title"},
		{`meta[name="description" i]`, `meta description`, "description"},
		{`link[rel="canonical" i]`, `link rel=canonical`, "canonical"},
	}
	for _, p := range pairs {
		n := ctx.Doc.Find(p.sel).Length()
		if n <= 1 {
			continue
		}
		r := checks.NewResult("seo.duplicate_meta."+p.id, checks.CategorySEO, checks.SeverityWarning,
			fmt.Sprintf("%d %s elements (expected 1)", n, p.label))
		r.Recommendation = "Keep exactly one — crawlers and social cards pick unpredictably when there are multiple."
		out = append(out, r)
	}
	return out
}
