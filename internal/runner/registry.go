package runner

import (
	"github.com/berkaycubuk/kitsune/internal/checks"
	"github.com/berkaycubuk/kitsune/internal/checks/a11y"
	"github.com/berkaycubuk/kitsune/internal/checks/geo"
	"github.com/berkaycubuk/kitsune/internal/checks/perf"
	"github.com/berkaycubuk/kitsune/internal/checks/security"
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
		// SEO — document structure
		seo.DoctypeCheck{},
		seo.CharsetCheck{},
		seo.BaseHrefCheck{},
		seo.DuplicateMetaCheck{},
		seo.CanonicalDriftCheck{},
		seo.HreflangCheck{},
		seo.OGImageDimsCheck{},
		seo.IconsCheck{},
		seo.ManifestCheck{},
		seo.TargetBlankCheck{},
		// SEO — site
		seo.RobotsTxtCheck{},
		seo.SitemapCheck{},

		// GEO
		geo.LLMsTxtCheck{},
		geo.AIBotsCheck{},
		geo.JSONLDCheck{},
		geo.JSONLDValidationCheck{},
		geo.MicrodataCheck{},
		geo.RDFaCheck{},
		geo.SameAsCheck{},
		geo.AuthorDateCheck{},
		geo.ContentDepthCheck{},
		geo.ReadabilityCheck{},
		geo.CitationDensityCheck{},
		geo.TOCCheck{},

		// Performance
		perf.CacheControlCheck{},
		perf.CompressionCheck{},
		perf.HTTPProtocolCheck{},
		perf.RedirectChainCheck{},
		perf.RenderBlockingCheck{},
		perf.ResourceHintsCheck{},
		perf.InlineSizeCheck{},
		perf.ExternalHeadRequestsCheck{},
		perf.ResponsiveImagesCheck{},
		perf.LegacyImageFormatsCheck{},
		perf.LazyLoadingCheck{},
		perf.FetchPriorityCheck{},
		perf.MediaDimensionsCheck{},
		perf.AnimatedGIFCheck{},
		perf.DOMSizeCheck{},

		// Accessibility
		a11y.LangValidityCheck{},
		a11y.ButtonLinkNameCheck{},
		a11y.IframeTitleCheck{},
		a11y.ObjectEmbedNameCheck{},
		a11y.FormLabelCheck{},
		a11y.AutocompleteCheck{},
		a11y.DuplicateIDCheck{},
		a11y.TabindexCheck{},
		a11y.MetaRefreshCheck{},
		a11y.VideoCaptionsCheck{},
		a11y.SkipLinkCheck{},
		a11y.TableHeadersCheck{},
		a11y.ARIACheck{},

		// Security
		security.HSTSCheck{},
		security.CSPCheck{},
		security.NosniffCheck{},
		security.ReferrerPolicyCheck{},
		security.PermissionsPolicyCheck{},
		security.ClickjackingCheck{},
		security.ServerDisclosureCheck{},
		security.CookiesCheck{},
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
