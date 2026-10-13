package engine

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// refundCfg is a four-rung ladder with the baseline at "untrained": one rung
// below it (inept) and two above (novice, proficient), each costing 1 point.
func refundCfg() *config.Config {
	return &config.Config{
		DefaultProficiency: "untrained",
		Proficiencies: []config.Proficiency{
			{ID: "inept", Cost: 1},
			{ID: "untrained", Cost: 1},
			{ID: "novice", Cost: 1},
			{ID: "proficient", Cost: 1},
		},
		Skills: config.SkillMap{
			Order: []string{"general"},
			Items: map[string][]string{"general": {"Stealth"}},
		},
	}
}

// A skill pushed below the baseline must refund a point, not sit at zero: Inept
// is an ordinary rung and its cost is symmetric with the rungs above the
// default.
func TestSkillPointsUsedRefundsBelowBaseline(t *testing.T) {
	cfg := refundCfg()
	for _, tt := range []struct {
		prof string
		want int
	}{
		{"inept", -1},
		{"untrained", 0},
		{"novice", 1},
		{"proficient", 2},
	} {
		c := model.Character{Skills: map[string]string{model.SkillKey("general", "Stealth"): tt.prof}}
		if got := SkillPointsUsed(cfg, c); got != tt.want {
			t.Errorf("SkillPointsUsed(%s) = %d, want %d", tt.prof, got, tt.want)
		}
	}
}
