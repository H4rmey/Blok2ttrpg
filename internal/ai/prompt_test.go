package ai

import (
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// TestSystemPromptListsTraitKeys guards the regression where the schema showed
// only `name`, so generated characters came back with no traits at all.
func TestSystemPromptListsTraitKeys(t *testing.T) {
	loaded, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	prompt := SystemPrompt(loaded, "")
	for _, key := range []string{"name:", "age:", "description:", "motivation:", "problem:", "trait_1:", "trait_2:", "trait_3:"} {
		if !strings.Contains(prompt, key) {
			t.Errorf("system prompt is missing trait key %q", key)
		}
	}
}
