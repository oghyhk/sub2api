package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCachePrefixFingerprintIgnoresNewestConversationTurn(t *testing.T) {
	first := []byte(`{"systemInstruction":{"parts":[{"text":"be concise"}]},"contents":[{"role":"user","parts":[{"text":"hello"}]},{"role":"model","parts":[{"text":"hi"}]},{"role":"user","parts":[{"text":"write tests"}]}]}`)
	second := []byte(`{"systemInstruction":{"parts":[{"text":"be concise"}]},"contents":[{"role":"user","parts":[{"text":"hello"}]},{"role":"model","parts":[{"text":"hi"}]},{"role":"user","parts":[{"text":"explain the failure"}]}]}`)
	require.Equal(t, CachePrefixFingerprint(first), CachePrefixFingerprint(second))
}

func TestEnsureGeminiUpstreamSessionIDIsStableAndPreservesExplicitValue(t *testing.T) {
	body := []byte(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`)
	first, firstID := ensureGeminiUpstreamSessionID(body, "cache:digest:abc:uuid", 9, "gemini-3.6-flash")
	second, secondID := ensureGeminiUpstreamSessionID(body, "cache:digest:abc:uuid", 9, "gemini-3.6-flash")
	require.NotEmpty(t, firstID)
	require.Equal(t, firstID, secondID)
	require.Equal(t, first, second)

	explicit := []byte(`{"sessionId":"client-session","contents":[{"role":"user","parts":[{"text":"hello"}]}]}`)
	unchanged, id := ensureGeminiUpstreamSessionID(explicit, "cache:digest:abc:uuid", 9, "gemini-3.6-flash")
	require.Equal(t, "client-session", id)
	require.Equal(t, explicit, unchanged)
}

func TestScopedClientSessionIDIsTenantIsolated(t *testing.T) {
	require.NotEqual(t, ScopedClientSessionID(1, "same-session"), ScopedClientSessionID(2, "same-session"))
	require.Equal(t, ScopedClientSessionID(1, "same-session"), ScopedClientSessionID(1, "same-session"))
}
