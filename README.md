# kitsune

[![skills.sh](https://skills.sh/b/berkaycubuk/kitsune)](https://skills.sh/berkaycubuk/kitsune)

A single-binary CLI that audits a URL for **SEO** and **GEO** (Generative Engine Optimization) signals — the things that matter for getting cited by Google, Bing, and AI assistants like ChatGPT, Claude, and Perplexity.

Named after the Japanese fox spirit that sees through illusions — fitting for a tool that strips a site down to what crawlers actually see.

The v1 scope is **single-URL static analysis**: it parses the HTML, walks the head and body, and consults site-level resources (`robots.txt`, `llms.txt`, `llms-full.txt`). No headless browser, no crawling.

## Install

**Go:**

```sh
go install github.com/berkaycubuk/kitsune/cmd/kitsune@latest
```

**Pre-built binaries:** download for your OS/arch from the [Releases page](https://github.com/berkaycubuk/kitsune/releases).

**Build from source:**

```sh
go build -o bin/kitsune ./cmd/kitsune
```

Requires Go 1.21+.

## Usage

```sh
kitsune https://example.com
kitsune --json https://example.com
kitsune --checks=geo https://example.com
kitsune --fail-on=error https://example.com   # exit 2 on any error finding
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `--json` | false | Emit JSON instead of human-readable output |
| `--checks` | all | Comma-separated category filter (`seo`, `geo`) |
| `--user-agent` | `kitsune/0.1` | Override the HTTP User-Agent |
| `--timeout` | `15s` | HTTP timeout |
| `--fail-on` | — | Exit non-zero when any finding is at or above this severity (`info`/`warning`/`error`) |

## Agent usage

Kitsune is designed to be driven by AI agents and other automation. The contract:

- **`--json`** emits a machine-readable report on stdout. The top-level object includes `schema_version`, `tool`, and `tool_version` so callers can detect drift.
- **stdout** carries the report. **stderr** carries errors and diagnostics only. The two streams never mix.
- **Non-interactive.** Kitsune never prompts and never reads stdin unless the URL argument is `-`.
- **Exit codes:**
  - `0` — success
  - `1` — tool or fetch error (invalid URL, DNS failure, timeout); message on stderr
  - `2` — findings reached or exceeded the `--fail-on` severity threshold; the full report is still on stdout
- **ANSI colors** are stripped automatically when stdout is not a terminal, so piped output is clean.
- Pass `-` as the URL to read a single URL from stdin: `echo https://example.com | kitsune --json -`

Each finding in `results[]` has a stable `id` (e.g. `seo.title.short`, `geo.llms.txt.missing`), so agents can match findings without parsing titles.

## What it checks

### SEO

- HTTP status, redirect chain, HTTPS + mixed content
- `<title>` length (30–60 chars)
- `<meta name="description">` length (70–160 chars)
- Canonical link
- `robots` meta + `X-Robots-Tag` (`noindex` / `nofollow` flagged)
- `lang` attribute on `<html>`
- Viewport meta
- Open Graph (`og:title/description/image/url/type`)
- Twitter Card
- Exactly one `<h1>`, heading-level skips
- Image `alt` coverage
- Link summary (internal/external/nofollow/empty anchor text)
- `robots.txt` reachability, parse, allow status, sitemap declaration

### GEO

- `/llms.txt` and `/llms-full.txt` presence ([llmstxt.org](https://llmstxt.org/) spec)
- AI crawler directives in `robots.txt` (GPTBot, ClaudeBot, PerplexityBot, Google-Extended, OAI-SearchBot, Applebot-Extended, CCBot, …)
- JSON-LD structured data: block count, schema.org `@type` enumeration, FAQ schema check
- Author / published / modified metadata (E-E-A-T signals)
- Body word count, statistic/numeric density, question-style heading ratio (from the [Princeton GEO paper](https://arxiv.org/abs/2311.09735))

Every finding carries a severity (`info` / `warning` / `error`), a one-line title, and (where applicable) a recommendation and link to the authoritative guideline.

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

Summary
  info    15
  warn    7
  error   0
```

## Out of scope (v1)

- Multi-page crawling — single URL only
- Core Web Vitals / Lighthouse — needs a headless browser
- JS-rendered SPAs — static parser only sees server-rendered HTML
- Outbound link liveness — links are listed, not HEAD-checked

## Project layout

```
cmd/kitsune/          CLI entry (cobra)
internal/fetch/       HTTP + site-resource fetch
internal/checks/      Check interface, Result, PageContext
internal/checks/seo/  SEO checks
internal/checks/geo/  GEO checks
internal/runner/      Orchestrator + check registry
internal/report/      Terminal + JSON renderers
```

Adding a check: implement the `checks.Check` interface in `internal/checks/seo/` or `internal/checks/geo/` and register it in `internal/runner/registry.go`.
