# kitsune

[![skills.sh](https://skills.sh/b/berkaycubuk/kitsune)](https://skills.sh/berkaycubuk/kitsune)
[![GitHub release](https://img.shields.io/github/v/release/berkaycubuk/kitsune)](https://github.com/berkaycubuk/kitsune/releases/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Audit any URL for SEO, GEO (AI-search optimization), performance, accessibility, and security signals — in one binary, with stable JSON output built for AI agents.**

`kitsune` fetches a URL, parses the HTML, consults `robots.txt` / `llms.txt` / `llms-full.txt`, and emits 100+ findings with stable IDs, severities, and concrete recommendations. No headless browser, no crawling — fast, deterministic, and pipeline-friendly.

*Named after the Japanese fox spirit that sees through illusions — fitting for a tool that strips a site down to what crawlers actually see.*

---

## TL;DR

```sh
go install github.com/berkaycubuk/kitsune/cmd/kitsune@latest
kitsune https://example.com
```

For agents and CI:

```sh
kitsune --json --fail-on=error https://example.com
```

## Install

**Go:**

```sh
go install github.com/berkaycubuk/kitsune/cmd/kitsune@latest
```

**Pre-built binaries:** download for your OS/arch from the [Releases page](https://github.com/berkaycubuk/kitsune/releases) (macOS + Linux arm64/amd64, Windows amd64).

**Build from source:**

```sh
go build -o bin/kitsune ./cmd/kitsune
```

Requires Go 1.21+.

## Usage

```sh
kitsune https://example.com                           # full audit, terminal output
kitsune --json https://example.com                    # machine-readable JSON
kitsune --checks=seo,geo https://example.com          # restrict to categories
kitsune --fail-on=error https://example.com           # exit 2 on any error finding (CI gate)
echo https://example.com | kitsune --json -           # read URL from stdin
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Emit JSON instead of human-readable output |
| `--checks` | all | Comma-separated category filter (`seo`, `geo`, `perf`, `a11y`, `security`) |
| `--user-agent` | — | Override the HTTP User-Agent |
| `--timeout` | `15s` | HTTP timeout |
| `--fail-on` | — | Exit non-zero when any finding is at or above this severity (`info` / `warning` / `error`) |
| `--version` | — | Print version and exit |

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. Report on stdout. |
| `1` | Tool / fetch error (invalid URL, DNS, timeout). Message on stderr. |
| `2` | Findings reached or exceeded the `--fail-on` threshold. Full report still on stdout. |

## Use with AI agents

Kitsune is built for agentic use first. A published [skills.sh](https://skills.sh/berkaycubuk/kitsune) skill ships with the tool — drop it into your agent and it'll know when to run kitsune, how to invoke it, and how to interpret the output.

**Install the skill:**

```sh
npx skills add berkaycubuk/kitsune
```

**The contract agents rely on:**

- **`--json`** emits a machine-readable report on stdout. The top-level object includes `schema_version`, `tool`, and `tool_version` so callers can detect drift.
- **stdout** carries the report. **stderr** carries errors only. The two streams never mix.
- **Stable IDs.** Every finding has a dotted `id` like `seo.title.short` or `geo.llms_txt.absent`. Match on `id`, never on `title`. IDs do not change between releases.
- **Non-interactive.** Kitsune never prompts. Pass `-` as the URL to read it from stdin.
- **ANSI colors** are stripped automatically when stdout is not a terminal — piped output is clean.

The full agent reference lives in [`skills/kitsune/`](skills/kitsune/).

### JSON output shape

```json
{
  "schema_version": "1",
  "tool": "kitsune",
  "tool_version": "0.1.0",
  "url": "https://example.com",
  "final_url": "https://example.com",
  "status_code": 200,
  "fetched_at": "2026-05-21T11:20:49Z",
  "results": [
    {
      "id": "seo.title.short",
      "category": "seo",
      "severity": "warning",
      "title": "Title is shorter than recommended",
      "detail": "Title is 14 characters; aim for 30–60.",
      "recommendation": "Expand the title with descriptive, keyword-relevant terms.",
      "guideline_url": "https://developers.google.com/search/docs/appearance/title-link"
    }
  ],
  "summary": { "info": 26, "warning": 16, "error": 0 }
}
```

## What it checks

### SEO — crawlability and search appearance

HTTP status, redirect chain, HTTPS + mixed content; `<title>` length (30–60); meta description length (70–160); canonical link + canonical drift; `robots` meta + `X-Robots-Tag`; `<html lang>`; viewport; Open Graph (`og:title/description/image/url/type` + image dims + locale); Twitter Card; exactly one `<h1>` and heading-level skips; image `alt` coverage; link summary (internal/external/nofollow/empty/`target=_blank` safety); `robots.txt` reachability + parse + allow status + sitemap declaration; sitemap reachability; charset; doctype; hreflang validity (self-reference, x-default); favicon / apple-touch-icon / theme color / manifest; base href; duplicate meta tags.

### GEO — Generative Engine Optimization (AI search)

`/llms.txt` and `/llms-full.txt` presence per the [llmstxt.org](https://llmstxt.org/) spec; AI crawler directives in `robots.txt` (GPTBot, ClaudeBot, PerplexityBot, Google-Extended, OAI-SearchBot, Applebot-Extended, CCBot, …); JSON-LD structured data (block count, schema.org `@type` enumeration, FAQ schema, parse + validation errors); microdata + RDFa; `sameAs` (E-E-A-T); author / published / modified metadata; body word count, statistic / numeric density, question-style heading ratio (from the [Princeton GEO paper](https://arxiv.org/abs/2311.09735)); citation density; readability metrics; table-of-contents detection.

### Performance — static signals (no headless browser)

HTTP protocol version; `Content-Encoding` (gzip/brotli/zstd); `Cache-Control` (presence, `no-store`, short TTLs); redirect chain length; render-blocking scripts / styles; DOM size; inline payload size; external head requests; image lazy-loading + first-image-lazy anti-pattern; responsive `srcset`; image dimensions; legacy image formats and animated GIFs; `fetchpriority` on LCP image; resource hints (`preconnect`, `preload`).

### Accessibility (a11y)

`<html lang>` validity; form input labels + `autocomplete`; button / link accessible name; ARIA `aria-hidden` on focusable elements; duplicate `id`s; iframe `title`; `<object>` / `<embed>` accessible name; skip-to-content link; positive `tabindex` (anti-pattern); table headers (`<th>`, `scope`, broken `headers` refs); video `<track>` captions; `<meta http-equiv="refresh">`.

### Security — response-header hygiene

`Content-Security-Policy` (presence + unsafe directives); HSTS; `X-Frame-Options` / `frame-ancestors` (clickjacking); `X-Content-Type-Options` (`nosniff`); `Referrer-Policy`; `Permissions-Policy`; cookie attributes (`Secure` / `HttpOnly` / `SameSite`); `Server` / `X-Powered-By` disclosure.

Every finding carries a severity (`info` / `warning` / `error`), a stable `id`, a one-line title, and — where applicable — a `recommendation` and a link to the authoritative `guideline_url`.

## Example output

```
▌SEO
  • Title: "Example Domain" (14 chars)
  ! Title is shorter than recommended
    Title is 14 characters; aim for 30–60.
    → Expand the title with descriptive, keyword-relevant terms.
  ! Missing meta description
  ! Missing canonical link
  • H1: "Example Domain"

▌GEO
  • /llms.txt not found
  ! No JSON-LD structured data found

▌Performance
  • HTTP/2
  ! No Cache-Control header set

▌Accessibility
  • <html lang="en">
  ! Image missing alt text (3 of 7)

▌Security
  ! No Content-Security-Policy header
  ! No Strict-Transport-Security header

Summary
  info    26
  warn    16
  error   0
```

## Out of scope (v1)

- **Multi-page crawling** — single URL only.
- **JavaScript rendering** — server-rendered HTML only; SPAs will look empty.
- **Core Web Vitals / Lighthouse** — no headless browser involved.
- **Outbound link liveness** — links are summarized, not HEAD-checked.

If you need any of these, kitsune is the wrong tool — and an honest agent should say so rather than pretend the gap doesn't exist.

## Project layout

```
cmd/kitsune/             CLI entry (cobra)
internal/fetch/          HTTP + site-resource fetch
internal/checks/         Check interface, Result, PageContext
internal/checks/seo/     SEO checks
internal/checks/geo/     GEO checks
internal/checks/perf/    Performance checks
internal/checks/a11y/    Accessibility checks
internal/checks/security/ Security-header checks
internal/runner/         Orchestrator + check registry
internal/report/         Terminal + JSON renderers
internal/version/        Build-time version + JSON schema version
skills/kitsune/          Agent skill (SKILL.md + REFERENCE.md)
```

Adding a check: implement the `checks.Check` interface in the appropriate category package and register it in `internal/runner/registry.go`.

## License

MIT © Berkay Çubuk — see [LICENSE](LICENSE).
