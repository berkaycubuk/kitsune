package a11y

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/berkaycubuk/kitsune/internal/checks"
	"golang.org/x/net/html"
)

// ARIA 1.2 abstract + concrete role names (subset of the role taxonomy that
// authors actually use). Abstract roles are intentionally excluded — they
// should not appear on real elements.
var knownRoles = map[string]bool{
	"alert": true, "alertdialog": true, "application": true, "article": true,
	"banner": true, "blockquote": true, "button": true, "caption": true,
	"cell": true, "checkbox": true, "code": true, "columnheader": true,
	"combobox": true, "complementary": true, "contentinfo": true, "definition": true,
	"deletion": true, "dialog": true, "directory": true, "document": true,
	"emphasis": true, "feed": true, "figure": true, "form": true,
	"generic": true, "grid": true, "gridcell": true, "group": true,
	"heading": true, "img": true, "insertion": true, "link": true,
	"list": true, "listbox": true, "listitem": true, "log": true,
	"main": true, "marquee": true, "math": true, "menu": true,
	"menubar": true, "menuitem": true, "menuitemcheckbox": true, "menuitemradio": true,
	"meter": true, "navigation": true, "none": true, "note": true,
	"option": true, "paragraph": true, "presentation": true, "progressbar": true,
	"radio": true, "radiogroup": true, "region": true, "row": true,
	"rowgroup": true, "rowheader": true, "scrollbar": true, "search": true,
	"searchbox": true, "separator": true, "slider": true, "spinbutton": true,
	"status": true, "strong": true, "subscript": true, "superscript": true,
	"switch": true, "tab": true, "table": true, "tablist": true,
	"tabpanel": true, "term": true, "textbox": true, "time": true,
	"timer": true, "toolbar": true, "tooltip": true, "tree": true,
	"treegrid": true, "treeitem": true,
}

// ARIA 1.2 attributes (states + properties). The role= attribute is checked separately.
var knownAriaAttrs = map[string]bool{
	"aria-activedescendant": true, "aria-atomic": true, "aria-autocomplete": true,
	"aria-braillelabel": true, "aria-brailleroledescription": true, "aria-busy": true,
	"aria-checked": true, "aria-colcount": true, "aria-colindex": true,
	"aria-colindextext": true, "aria-colspan": true, "aria-controls": true,
	"aria-current": true, "aria-describedby": true, "aria-description": true,
	"aria-details": true, "aria-disabled": true, "aria-dropeffect": true,
	"aria-errormessage": true, "aria-expanded": true, "aria-flowto": true,
	"aria-grabbed": true, "aria-haspopup": true, "aria-hidden": true,
	"aria-invalid": true, "aria-keyshortcuts": true, "aria-label": true,
	"aria-labelledby": true, "aria-level": true, "aria-live": true,
	"aria-modal": true, "aria-multiline": true, "aria-multiselectable": true,
	"aria-orientation": true, "aria-owns": true, "aria-placeholder": true,
	"aria-posinset": true, "aria-pressed": true, "aria-readonly": true,
	"aria-relevant": true, "aria-required": true, "aria-roledescription": true,
	"aria-rowcount": true, "aria-rowindex": true, "aria-rowindextext": true,
	"aria-rowspan": true, "aria-selected": true, "aria-setsize": true,
	"aria-sort": true, "aria-valuemax": true, "aria-valuemin": true,
	"aria-valuenow": true, "aria-valuetext": true,
}

var focusableTags = map[string]bool{
	"a": true, "button": true, "input": true, "select": true, "textarea": true,
	"summary": true, "details": true, "iframe": true,
}

// ARIACheck — validate roles, attribute names, and a common misuse.
type ARIACheck struct{}

func (ARIACheck) ID() string       { return "a11y.aria" }
func (ARIACheck) Category() string { return checks.CategoryA11y }

func (ARIACheck) Run(ctx *checks.PageContext) []checks.Result {
	badRoles := map[string]int{}
	badAttrs := map[string]int{}
	hiddenFocusable := 0

	ctx.Doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		n := s.Get(0)
		if n == nil || n.Type != html.ElementNode {
			return
		}
		for _, a := range n.Attr {
			k := strings.ToLower(a.Key)
			if k == "role" {
				for _, role := range strings.Fields(strings.ToLower(a.Val)) {
					if !knownRoles[role] {
						badRoles[role]++
					}
				}
				continue
			}
			if strings.HasPrefix(k, "aria-") && !knownAriaAttrs[k] {
				badAttrs[k]++
			}
		}
		// aria-hidden="true" on a focusable element.
		if strings.EqualFold(strings.TrimSpace(s.AttrOr("aria-hidden", "")), "true") {
			if focusableTags[strings.ToLower(n.Data)] && !isExplicitlyNonFocusable(s) {
				hiddenFocusable++
			}
		}
	})

	var out []checks.Result
	if len(badRoles) > 0 {
		out = append(out, mapToResult(
			"a11y.aria.unknown_role", checks.SeverityWarning,
			"unknown role(s)", badRoles,
			"Use a role from the WAI-ARIA 1.2 taxonomy or remove the attribute.",
			"https://www.w3.org/TR/wai-aria-1.2/#role_definitions",
		))
	}
	if len(badAttrs) > 0 {
		out = append(out, mapToResult(
			"a11y.aria.unknown_attr", checks.SeverityWarning,
			"unknown aria-* attribute(s)", badAttrs,
			"Check the spelling against the WAI-ARIA attribute list.",
			"https://www.w3.org/TR/wai-aria-1.2/#state_prop_def",
		))
	}
	if hiddenFocusable > 0 {
		r := checks.NewResult("a11y.aria.hidden_focusable", checks.CategoryA11y, checks.SeverityWarning,
			fmt.Sprintf("%d focusable element(s) with aria-hidden=\"true\"", hiddenFocusable))
		r.Recommendation = "Either remove `aria-hidden`, or add `tabindex=\"-1\"` and disable interaction."
		r.GuidelineURL = "https://dequeuniversity.com/rules/axe/4.7/aria-hidden-focus"
		out = append(out, r)
	}
	return out
}

func isExplicitlyNonFocusable(s *goquery.Selection) bool {
	if _, ok := s.Attr("disabled"); ok {
		return true
	}
	if v, ok := s.Attr("tabindex"); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n < 0 {
			return true
		}
	}
	return false
}

func mapToResult(id string, sev checks.Severity, label string, counts map[string]int, rec, url string) checks.Result {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s (×%d)", k, counts[k])
	}
	r := checks.NewResult(id, checks.CategoryA11y, sev,
		fmt.Sprintf("%d %s", len(counts), label))
	if len(parts) > 5 {
		parts = append(parts[:5], fmt.Sprintf("…+%d more", len(parts)-5))
	}
	r.Detail = strings.Join(parts, ", ")
	r.Recommendation = rec
	r.GuidelineURL = url
	return r
}
