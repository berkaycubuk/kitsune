# kitsune finding ID reference

Every finding kitsune emits carries an `id` field. IDs are **stable across releases** — match on them in code instead of parsing the human-readable `title`.

## Naming convention

```
<category>.<area>.<verdict>
```

- **category** — one of `seo`, `geo`, `perf`, `a11y`, `security`
- **area** — what subsystem the check probes (e.g. `title`, `canonical`, `llms_txt`, `csp`)
- **verdict** — what the check concluded

### Common verdict suffixes

| Suffix | Meaning | Severity it usually carries |
|---|---|---|
| `.present` / `.ok` / `.allowed` | The signal is healthy. | `info` |
| `.missing` / `.absent` | The signal isn't there at all. | `warning` or `error` |
| `.short` / `.long` | The signal exists but is out of recommended bounds. | `warning` |
| `.invalid` / `.parse_error` / `.broken_ref` | The signal exists but is malformed. | `warning` or `error` |
| `.blocked` / `.unreachable` | A resource was denied or could not be fetched. | `warning` or `error` |
| `.summary` / `.metrics` / `.count` | Numeric/structural rollup, no judgment attached. | `info` |
| `.unsafe` / `.leaky` | Present but with a known security/SEO foot-gun. | `warning` |

This means an agent can often act without recognizing the specific check — `*.missing` + a `recommendation` field is enough to surface as "add this thing".

## What each category covers

### `seo.*` — crawlability and search appearance

`<title>`, meta description, canonical link, `robots` meta + `X-Robots-Tag`, `<html lang>`, viewport, Open Graph, Twitter Card, `<h1>` count and heading skips, image `alt` coverage, link summary, `robots.txt` reachability/parse/allow status, sitemap declaration, charset, doctype, hreflang, favicon/manifest/apple-touch-icon, base href, HTTPS + mixed content, redirect chain, target=\_blank safety, canonical drift, duplicate meta tags.

### `geo.*` — Generative Engine Optimization (AI search)

`/llms.txt` and `/llms-full.txt` presence (per [llmstxt.org](https://llmstxt.org/)), AI crawler directives in `robots.txt` (GPTBot, ClaudeBot, PerplexityBot, Google-Extended, OAI-SearchBot, Applebot-Extended, CCBot, …), JSON-LD structured data (block count, schema.org `@type` enumeration, FAQ schema, validation), microdata, RDFa, `sameAs`, author / published / modified metadata (E-E-A-T), content depth (word count, statistic/numeric density, Q-style heading ratio per the [Princeton GEO paper](https://arxiv.org/abs/2311.09735)), citation density, readability metrics, table of contents.

### `perf.*` — static performance signals

Note: no headless browser involved. These are signals visible from HTML + response headers — not Lighthouse scores.

HTTP protocol version, compression (`Content-Encoding`), `Cache-Control`, redirect chain length, render-blocking scripts/styles, DOM size, inline payload size, external head requests, image lazy-loading, first-image-lazy anti-pattern, responsive image `srcset`, image dimensions, legacy image formats (gif/bmp/etc), animated GIF count, `fetchpriority` on LCP image, resource hints (`preconnect`/`preload`).

### `a11y.*` — accessibility checks

`<html lang>` validity, form input labels and `autocomplete`, button/link accessible name, ARIA `aria-hidden` on focusable elements, duplicate `id`s, iframe `title`, `<object>`/`<embed>` accessible name, skip-to-content link, positive `tabindex` (anti-pattern), table headers (`<th>`, `scope`, broken `headers` refs), video captions/tracks, `<meta http-equiv="refresh">`.

### `security.*` — response-header hygiene

`Content-Security-Policy` (presence and unsafe directives), `Strict-Transport-Security` (HSTS), `X-Frame-Options` / `frame-ancestors` (clickjacking), `X-Content-Type-Options` (`nosniff`), `Referrer-Policy`, `Permissions-Policy`, cookie attributes (Secure/HttpOnly/SameSite), `Server`/`X-Powered-By` disclosure.

## Reading a finding

```json
{
  "id": "geo.llms_txt.absent",
  "category": "geo",
  "severity": "warning",
  "title": "/llms.txt not found",
  "detail": "Site does not publish an llms.txt manifest.",
  "recommendation": "Publish /llms.txt summarizing the site for LLMs (see https://llmstxt.org).",
  "guideline_url": "https://llmstxt.org/"
}
```

Use `id` for branching logic. Use `title` / `detail` / `recommendation` when generating human-facing explanations or code fixes. Use `guideline_url` when the user wants to read the spec.

## Severity ladder

`info` < `warning` < `error`

- **info** — observation; nothing wrong. Skip when summarizing problems.
- **warning** — recommended fix; the page works but is missing best practice.
- **error** — likely indexability/security failure; fix before publishing.

`--fail-on=error` is the common CI gate. `--fail-on=warning` is stricter, useful for content/SEO polish.

## Discovering IDs at runtime

The set above is illustrative, not authoritative. Future releases may add IDs. To enumerate every ID the installed binary emits, run `kitsune --json <url>` against any representative page and read `results[].id` out of the JSON. New IDs follow the same naming convention.
