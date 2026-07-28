package domain

import "testing"

func TestDefaultAntigravityModelMapping_ContainsOpusAndGemini36Flash(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"claude-opus-4-6":          "claude-opus-4-6-thinking",
		"claude-opus-4-6-thinking": "claude-opus-4-6-thinking",
		"gemini-3.6-flash":         "gemini-3.6-flash-high",
		"gemini-3.6-flash-high":    "gemini-3.6-flash-high",
	}

	for from, want := range cases {
		got, ok := DefaultAntigravityModelMapping[from]
		if !ok {
			t.Fatalf("expected mapping for %q to exist", from)
		}
		if got != want {
			t.Fatalf("unexpected mapping for %q: got %q want %q", from, got, want)
		}
	}
}

func TestDefaultBedrockModelMapping_ContainsNewClaudeModels(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"claude-fable-5":  "anthropic.claude-fable-5",
		"claude-opus-4-8": "us.anthropic.claude-opus-4-8-v1",
	}
	for from, want := range cases {
		got, ok := DefaultBedrockModelMapping[from]
		if !ok {
			t.Fatalf("expected Bedrock mapping for %q to exist", from)
		}
		if got != want {
			t.Fatalf("unexpected Bedrock mapping for %q: got %q want %q", from, got, want)
		}
	}
}
