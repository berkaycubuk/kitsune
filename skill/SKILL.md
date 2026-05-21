---
name: kitsune
description: Audit a URL for SEO, GEO (Generative Engine Optimization), accessibility, performance, and security signals using the kitsune CLI. Use when a user wants to check a website, optimize a page for search engines or AI assistants (ChatGPT, Claude, Perplexity), validate llms.txt or robots.txt, fix indexability or metadata issues, audit a site before launch, or gate CI on site-health checks. Triggers include "audit this URL", "SEO check", "is my site rankable", "optimize for AI search", "GEO optimization", "llms.txt", "structured data check", "what's wrong with my page".
---

# kitsune

`kitsune` is a single-binary CLI that fetches one URL, parses the HTML, consults site-level resources (`robots.txt`, `llms.txt`, `llms-full.txt`), and emits findings across five categories: **seo**, **geo**, **perf**, **a11y**, **security**.

It is designed for agentic use: JSON output with a stable schema, stable finding IDs, clean stdout/stderr separation, and documented exit codes.

## Install

```sh
go install github.com/berkaycubuk/kitsune/cmd/kitsune@latest
```

Or download a pre-built binary from <https://github.com/berkaycubuk/kitsune/releases>.

## Quick start

**Always pass `--json` when calling from an agent.** The default output is for humans; the JSON output is for you.

```sh
kitsune --json https://example.com
```

The JSON shape (stdout):

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

## Workflow

1. Run `kitsune --json <url>`. Capture stdout; treat stderr as diagnostics only.
2. Check the exit code: `0` = ok, `1` = fetch/tool error, `2` = findings exceeded `--fail-on` threshold.
3. Filter `results[]` to entries with `severity` of `warning` or `error` — these are actionable. `info` entries are observations, not problems.
4. Match findings by **`id`**, never by `title`. IDs are stable across releases; titles are not.
5. When producing fixes, lean on the `recommendation` and `guideline_url` fields. Tie each fix back to specific code or content the user is editing.
6. Group output by `category` when presenting to the user so the report stays readable.

## Useful flags

- `--checks=seo,geo` — restrict to categories. Valid: `seo`, `geo`, `perf`, `a11y`, `security`. Default runs all.
- `--fail-on=error` — exit 2 when any finding meets/exceeds the severity. Severity order: `info` < `warning` < `error`. Use this for CI gates.
- `--timeout=30s` — default is 15s; bump for slow targets.
- `--user-agent="MyBot/1.0"` — override UA if the target blocks the default.
- Pass `-` as the URL to read one URL from stdin: `echo https://example.com | kitsune --json -`.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. Report on stdout. |
| `1` | Tool/fetch error (invalid URL, DNS, timeout). Message on stderr. No report. |
| `2` | Findings reached/exceeded `--fail-on` threshold. Full report still on stdout. |

## Out of scope — do not promise these

- **No multi-page crawling.** Single URL only.
- **No JavaScript rendering.** Server-rendered HTML only; SPAs will look empty.
- **No Core Web Vitals or Lighthouse.** No headless browser is run. `perf` checks are static signals only.
- **No outbound link liveness.** Links are summarized, not HEAD-checked.

If the user asks for any of the above, say so explicitly rather than running kitsune and glossing over the gap.

## Finding ID catalog

Every finding has a stable, dotted `id` like `seo.title.short` or `geo.llms_txt.absent`. The namespace and common patterns are documented in [REFERENCE.md](REFERENCE.md). When you need the exhaustive list for a category, just run kitsune against any representative URL and read the `id`s out of the JSON.
