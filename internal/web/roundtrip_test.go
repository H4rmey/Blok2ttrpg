package web

import (
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/engine"
)

// TestBuilderCostMatchesStoredCost guards the round trip: the cost shown when a
// perk is opened in the builder must equal the cost computed for the same perk
// in the list. A difference means the builder form does not faithfully carry
// the stored ability, so simply opening and saving a perk would change its
// price.
func TestBuilderCostMatchesStoredCost(t *testing.T) {
	app, c := testAppWithPerk(t)
	stored := engine.AbilityCost(app.Cfg.Config, c.Abilities[0])

	rec := httptest.NewRecorder()
	app.renderBuilder(rec, c, &c.Abilities[0], false)
	form := formValuesFromHTML(rec.Body.String())
	form.Set("character_id", c.ID)

	req := httptest.NewRequest("POST", "/builder/cost", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()
	// The save path normalizes what the form posts, so the replay does too:
	// this asserts the real invariant, that opening a perk and saving it leaves
	// the cost untouched. Normalizing cannot conjure data the form dropped, so
	// genuine losses (e.g. an omitted interaction) still fail here.
	saved := engine.NormalizeAbility(app.Cfg.Config, app.buildAbilityFromForm(req, "ability-1"))
	replayed := engine.AbilityCost(app.Cfg.Config, saved)

	if replayed.Build != stored.Build || replayed.Energy != stored.Energy {
		t.Errorf("builder round trip changed the cost: stored build=%d energy=%d, builder build=%d energy=%d\nform: %v",
			stored.Build, stored.Energy, replayed.Build, replayed.Energy, form)
	}
}

var (
	inputRe  = regexp.MustCompile(`(?s)<input\b[^>]*>`)
	selectRe = regexp.MustCompile(`(?s)<select\b[^>]*>.*?</select>`)
	optionRe = regexp.MustCompile(`(?s)<option\b[^>]*>`)
)

// formValuesFromHTML scrapes the values a browser would submit from the
// rendered builder form: input values and the selected option of each select.
func formValuesFromHTML(body string) url.Values {
	v := url.Values{}
	for _, tag := range inputRe.FindAllString(body, -1) {
		name := attr(tag, "name")
		if name == "" {
			continue
		}
		if strings.Contains(tag, `type="checkbox"`) {
			if strings.Contains(tag, "checked") {
				v.Set(name, "on")
			}
			continue
		}
		v.Set(name, attr(tag, "value"))
	}
	for _, sel := range selectRe.FindAllString(body, -1) {
		name := attr(sel, "name")
		if name == "" {
			continue
		}
		opts := optionRe.FindAllString(sel, -1)
		// A select with no explicit selection submits its first option.
		chosen := ""
		if len(opts) > 0 {
			chosen = attr(opts[0], "value")
		}
		for _, o := range opts {
			if strings.Contains(o, "selected") {
				chosen = attr(o, "value")
				break
			}
		}
		v.Set(name, chosen)
	}
	return v
}

func attr(tag, name string) string {
	m := regexp.MustCompile(name + `="([^"]*)"`).FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	return m[1]
}
