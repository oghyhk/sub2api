package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

// ScopedClientSessionID namespaces an optional caller-provided session ID by
// API key without retaining the raw identifier in cache keys or logs.
func ScopedClientSessionID(apiKeyID int64, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(strconv.FormatInt(apiKeyID, 10) + "|" + raw))
	return "client:" + base64.RawURLEncoding.EncodeToString(digest[:12])
}

// CachePrefixFingerprint returns a privacy-safe identifier for the cacheable
// historical request prefix. It deliberately omits the newest conversation turn
// and transport-only sessionId so it remains comparable as a conversation grows.
// No request content is logged or persisted by this helper.
func CachePrefixFingerprint(body []byte) string {
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		digest := sha256.Sum256(body)
		return base64.RawURLEncoding.EncodeToString(digest[:12])
	}
	delete(request, "sessionId")
	for _, field := range []string{"contents", "messages", "input"} {
		items, ok := request[field].([]any)
		if !ok || len(items) == 0 {
			continue
		}
		request[field] = items[:len(items)-1]
	}
	canonical, err := json.Marshal(request)
	if err != nil {
		canonical = body
	}
	digest := sha256.Sum256(canonical)
	return base64.RawURLEncoding.EncodeToString(digest[:12])
}
