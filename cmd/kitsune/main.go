package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/report"
	"github.com/berkaycubuk/kitsune/internal/runner"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	var (
		jsonOut    bool
		categories []string
		userAgent  string
		timeout    time.Duration
		failOn     string
	)

	cmd := &cobra.Command{
		Use:   "kitsune <url>",
		Short: "Analyze a single URL for SEO and GEO signals",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := normalizeURL(args[0])
			if err != nil {
				return err
			}

			catSet := map[string]bool{}
			for _, c := range categories {
				c = strings.ToLower(strings.TrimSpace(c))
				if c != "" {
					catSet[c] = true
				}
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
	cmd.Flags().StringSliceVar(&categories, "checks", nil, "Categories to run (seo, geo). Default: all")
	cmd.Flags().StringVar(&userAgent, "user-agent", "", "Override the HTTP User-Agent")
	cmd.Flags().DurationVar(&timeout, "timeout", 15*time.Second, "HTTP timeout")
	cmd.Flags().StringVar(&failOn, "fail-on", "", "Exit non-zero if any finding has severity >= this (info|warning|error)")

	return cmd
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
