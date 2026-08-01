package antigravity

import "testing"

func TestDefaultModels_ContainsOnlyProductModels(t *testing.T) {
	t.Parallel()

	models := DefaultModels()
	byID := make(map[string]ClaudeModel, len(models))
	for _, m := range models {
		byID[m.ID] = m
	}

	requiredIDs := []string{
		"claude-opus-4-6",
		"claude-opus-4-6-thinking",
		"gemini-3.1-flash-lite",
		"gemini-3.6-flash",
	}

	for _, id := range requiredIDs {
		if _, ok := byID[id]; !ok {
			t.Fatalf("expected model %q to be exposed in DefaultModels", id)
		}
	}
	if len(byID) != len(requiredIDs) {
		t.Fatalf("expected exactly %d product models, got %d", len(requiredIDs), len(byID))
	}
}
