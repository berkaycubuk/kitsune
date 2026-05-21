package a11y

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
)

var nonLabelableInputTypes = map[string]bool{
	"hidden": true, "submit": true, "reset": true, "button": true, "image": true,
}

// FormLabelCheck — every labelable input must have an associated label.
type FormLabelCheck struct{}

func (FormLabelCheck) ID() string       { return "a11y.form_label" }
func (FormLabelCheck) Category() string { return checks.CategoryA11y }

func (FormLabelCheck) Run(ctx *checks.PageContext) []checks.Result {
	// Build id → label[for] map once.
	labelFor := map[string]bool{}
	ctx.Doc.Find("label[for]").Each(func(_ int, s *goquery.Selection) {
		if id := strings.TrimSpace(s.AttrOr("for", "")); id != "" {
			labelFor[id] = true
		}
	})

	var unlabeled int
	ctx.Doc.Find("input, select, textarea").Each(func(_ int, s *goquery.Selection) {
		if s.Is("input") {
			t := strings.ToLower(strings.TrimSpace(s.AttrOr("type", "text")))
			if nonLabelableInputTypes[t] {
				return
			}
		}
		// Wrapped by a <label>.
		if s.ParentFiltered("label").Length() > 0 {
			return
		}
		// label[for=id]
		if id := strings.TrimSpace(s.AttrOr("id", "")); id != "" && labelFor[id] {
			return
		}
		// aria-label / aria-labelledby / title.
		for _, a := range []string{"aria-label", "aria-labelledby", "title"} {
			if strings.TrimSpace(s.AttrOr(a, "")) != "" {
				return
			}
		}
		unlabeled++
	})
	if unlabeled == 0 {
		return nil
	}
	r := checks.NewResult("a11y.form_label.unlabeled", checks.CategoryA11y, checks.SeverityWarning,
		fmt.Sprintf("%d form control(s) without a label", unlabeled))
	r.Recommendation = "Wire `<label for>` to the input id, wrap the input in a `<label>`, or set `aria-label`."
	r.GuidelineURL = "https://www.w3.org/WAI/tutorials/forms/labels/"
	return []checks.Result{r}
}

// Common input types whose autocomplete is well-defined and high-leverage.
var autocompleteFor = map[string]string{
	"email":    "email",
	"tel":      "tel",
	"url":      "url",
	"password": "current-password (or new-password)",
}

// AutocompleteCheck — flag inputs that should declare an autocomplete token.
type AutocompleteCheck struct{}

func (AutocompleteCheck) ID() string       { return "a11y.autocomplete" }
func (AutocompleteCheck) Category() string { return checks.CategoryA11y }

func (AutocompleteCheck) Run(ctx *checks.PageContext) []checks.Result {
	var missing []string
	ctx.Doc.Find("input").Each(func(_ int, s *goquery.Selection) {
		t := strings.ToLower(strings.TrimSpace(s.AttrOr("type", "text")))
		if _, ok := autocompleteFor[t]; !ok {
			return
		}
		if strings.TrimSpace(s.AttrOr("autocomplete", "")) != "" {
			return
		}
		name := s.AttrOr("name", s.AttrOr("id", t))
		missing = append(missing, fmt.Sprintf("%s (expected %s)", name, autocompleteFor[t]))
	})
	if len(missing) == 0 {
		return nil
	}
	r := checks.NewResult("a11y.autocomplete.missing", checks.CategoryA11y, checks.SeverityInfo,
		fmt.Sprintf("%d input(s) missing recommended autocomplete attribute", len(missing)))
	if len(missing) > 5 {
		missing = append(missing[:5], fmt.Sprintf("…+%d more", len(missing)-5))
	}
	r.Detail = strings.Join(missing, ", ")
	r.Recommendation = "Setting `autocomplete` improves both a11y (assistive tech) and conversion (browser autofill)."
	r.GuidelineURL = "https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/autocomplete"
	return []checks.Result{r}
}
