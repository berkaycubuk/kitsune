package runner

import (
	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/checks/geo"
	"github.com/berkaycubuk/kitsune/internal/checks/seo"
)

func allChecks() []checks.Check {
	return []checks.Check{
		// SEO — request-level
		seo.StatusCheck{},
		seo.HTTPSCheck{},
		// SEO — head/meta
		seo.TitleCheck{},
		seo.DescriptionCheck{},
		seo.CanonicalCheck{},
		seo.RobotsMetaCheck{},
		seo.LangCheck{},
		seo.ViewportCheck{},
		seo.OpenGraphCheck{},
		seo.TwitterCardCheck{},
		// SEO — structure
		seo.HeadingsCheck{},
		seo.ImageAltCheck{},
		seo.LinksCheck{},
		// SEO — site
		seo.RobotsTxtCheck{},

		// GEO
		geo.LLMsTxtCheck{},
		geo.AIBotsCheck{},
		geo.JSONLDCheck{},
		geo.AuthorDateCheck{},
		geo.ContentDepthCheck{},
	}
}

func filter(all []checks.Check, allowed map[string]bool) []checks.Check {
	if len(allowed) == 0 {
		return all
	}
	var out []checks.Check
	for _, c := range all {
		if allowed[c.Category()] {
			out = append(out, c)
		}
	}
	return out
}
