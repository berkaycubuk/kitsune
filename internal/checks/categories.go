package checks

const (
	CategorySEO      = "seo"
	CategoryGEO      = "geo"
	CategoryPerf     = "perf"
	CategoryA11y     = "a11y"
	CategorySecurity = "security"
)

// KnownCategories lists categories in their preferred display order.
func KnownCategories() []string {
	return []string{
		CategorySEO,
		CategoryGEO,
		CategoryPerf,
		CategoryA11y,
		CategorySecurity,
	}
}

func IsKnownCategory(c string) bool {
	for _, k := range KnownCategories() {
		if k == c {
			return true
		}
	}
	return false
}
