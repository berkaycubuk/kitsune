package geo

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/berkaycubuk/kitsune/internal/checks"
)

var sentenceSplit = regexp.MustCompile(`[.!?]+\s+`)

// ReadabilityCheck — mean sentence length + Flesch-Kincaid grade level.
type ReadabilityCheck struct{}

func (ReadabilityCheck) ID() string       { return "geo.readability" }
func (ReadabilityCheck) Category() string { return checks.CategoryGEO }

func (ReadabilityCheck) Run(ctx *checks.PageContext) []checks.Result {
	clone := ctx.Doc.Clone()
	clone.Find("script, style, nav, header, footer, aside").Remove()
	text := strings.TrimSpace(clone.Find("body").Text())
	if text == "" {
		return nil
	}

	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil
	}
	words := strings.Fields(text)
	if len(words) < 30 {
		// Too short for meaningful readability scoring.
		return nil
	}

	var syllables int
	for _, w := range words {
		syllables += countSyllables(w)
	}

	avgSentLen := float64(len(words)) / float64(len(sentences))
	avgSyllPerWord := float64(syllables) / float64(len(words))
	// Flesch-Kincaid Grade Level.
	fk := 0.39*avgSentLen + 11.8*avgSyllPerWord - 15.59

	sev := checks.SeverityInfo
	rec := ""
	if avgSentLen > 25 || fk > 14 {
		sev = checks.SeverityInfo // surface as info, not warning — readability is preference-laden
		rec = "Long sentences (>25 words) and grade level >14 reduce comprehension and AI extraction quality. Shorter prose helps both."
	}
	r := checks.NewResult("geo.readability.metrics", checks.CategoryGEO, sev,
		fmt.Sprintf("Readability: %.1f avg words/sentence, F-K grade %.1f", avgSentLen, fk))
	r.Recommendation = rec
	r.GuidelineURL = "https://arxiv.org/abs/2311.09735"
	return []checks.Result{r}
}

func splitSentences(text string) []string {
	parts := sentenceSplit.Split(text, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

// countSyllables is a heuristic: count vowel-group transitions per word, with a
// minimum of 1 and a small subtraction for trailing silent "e". Good enough for
// the F-K formula without shipping a pronunciation dictionary.
func countSyllables(w string) int {
	w = strings.ToLower(strings.TrimFunc(w, func(r rune) bool { return !unicode.IsLetter(r) }))
	if w == "" {
		return 0
	}
	vowels := "aeiouy"
	count := 0
	prevVowel := false
	for _, r := range w {
		isV := strings.ContainsRune(vowels, r)
		if isV && !prevVowel {
			count++
		}
		prevVowel = isV
	}
	if strings.HasSuffix(w, "e") && count > 1 {
		count--
	}
	if count < 1 {
		count = 1
	}
	return count
}
