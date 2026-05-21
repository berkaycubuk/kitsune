package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/report"
	"github.com/berkaycubuk/kitsune/internal/runner"
	"github.com/berkaycubuk/kitsune/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

const longDescription = `Analyze a single URL for SEO and GEO (Generative Engine Optimization) signals.

Kitsune fetches the URL, parses the HTML, consults site-level resources
(robots.txt, llms.txt, llms-full.txt), and emits a list of findings.

Output:
  Default            Human-readable terminal output on stdout.
  --json             Machine-readable JSON on stdout. Stable schema; the
                     top-level object carries schema_version, tool, and
                     tool_version fields so callers can detect drift.

I/O contract:
  - All report data is written to stdout.
  - All errors and diagnostics are written to stderr.
  - The tool is fully non-interactive; it never prompts.
  - Pass "-" as the URL to read a single URL from stdin.

Exit codes:
  0   Success.
  1   Tool or fetch error (invalid URL, DNS failure, timeout, etc.).
      The error message is printed to stderr.
  2   Findings reached or exceeded the --fail-on severity threshold.
      A full report was still printed to stdout.

Examples:
  kitsune https://example.com
  kitsune --json https://example.com
  kitsune --checks=geo https://example.com
  kitsune --fail-on=error --json https://example.com
  echo https://example.com | kitsune --json -`

func newRoot() *cobra.Command {
	var (
		jsonOut    bool
		categories []string
		userAgent  string
		timeout    time.Duration
		failOn     string
	)

	cmd := &cobra.Command{
		Use:           "kitsune <url>",
		Short:         "Analyze a single URL for SEO and GEO signals",
		Long:          longDescription,
		Version:       version.Version,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			rawURL := args[0]
			if rawURL == "-" {
				s, err := readURLFromStdin()
				if err != nil {
					return err
				}
				rawURL = s
			}
			target, err := normalizeURL(rawURL)
			if err != nil {
				return err
			}

			catSet := map[string]bool{}
			for _, c := range categories {
				c = strings.ToLower(strings.TrimSpace(c))
				if c == "" {
					continue
				}
				if !checks.IsKnownCategory(c) {
					return fmt.Errorf("unknown --checks value %q (known: %s)", c, strings.Join(checks.KnownCategories(), ", "))
				}
				catSet[c] = true
			}

			r, err := runner.Run(target, runner.Options{
				UserAgent:  userAgent,
				Timeout:    timeout,
				Categories: catSet,
			})
			if err != nil {
				return fmt.Errorf("analyze: %w", err)
			}

			if jsonOut {
				if err := report.WriteJSON(os.Stdout, r); err != nil {
					return err
				}
			} else {
				report.WriteTerminal(os.Stdout, r)
			}

			// Exit code based on --fail-on threshold.
			if failOn != "" {
				thr, ok := checks.ParseSeverity(strings.ToLower(failOn))
				if !ok {
					return fmt.Errorf("invalid --fail-on value: %s (expected info, warning, or error)", failOn)
				}
				worst := worstSeverity(r)
				if worst >= thr {
					os.Exit(2)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON instead of human-readable output")
	cmd.Flags().StringSliceVar(&categories, "checks", nil, "Categories to run ("+strings.Join(checks.KnownCategories(), ", ")+"). Default: all")
	cmd.Flags().StringVar(&userAgent, "user-agent", "", "Override the HTTP User-Agent")
	cmd.Flags().DurationVar(&timeout, "timeout", 15*time.Second, "HTTP timeout")
	cmd.Flags().StringVar(&failOn, "fail-on", "", "Exit non-zero if any finding has severity >= this (info|warning|error)")

	return cmd
}

func readURLFromStdin() (string, error) {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			return line, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return "", fmt.Errorf("no URL provided on stdin")
}

func normalizeURL(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty URL")
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("URL has no host: %s", s)
	}
	return u.String(), nil
}

func worstSeverity(r *runner.Report) checks.Severity {
	worst := checks.SeverityInfo
	for _, res := range r.Results {
		if res.Severity > worst {
			worst = res.Severity
		}
	}
	return worst
}
