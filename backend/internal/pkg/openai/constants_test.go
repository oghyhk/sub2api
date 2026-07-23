package openai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultModelsIncludeBareGPT56Alias(t *testing.T) {
	require.Contains(t, DefaultModelIDs(), "gpt-5.6")
}

func TestGPT56ModelsAdvertiseAccurateContextWindow(t *testing.T) {
	expected := map[string]struct{}{
		"gpt-5.6":       {},
		"gpt-5.6-sol":   {},
		"gpt-5.6-terra": {},
		"gpt-5.6-luna":  {},
	}

	for _, model := range DefaultModels {
		if _, ok := expected[model.ID]; !ok {
			continue
		}
		require.Equal(t, 358000, model.ContextWindow, model.ID)
		delete(expected, model.ID)
	}
	require.Empty(t, expected)
}
