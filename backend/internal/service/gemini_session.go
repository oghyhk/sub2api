package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/cespare/xxhash/v2"
)

// shortHash 使用 XXHash64 + Base36 生成短 hash（16 字符）
// XXHash64 比 SHA256 快约 10 倍，Base36 比 Hex 短约 20%
func shortHash(data []byte) string {
	h := xxhash.Sum64(data)
	return strconv.FormatUint(h, 36)
}

// BuildGeminiDigestChain builds a stable, privacy-preserving conversation identity
// for native Gemini sticky-session recovery.
//
// STICKY-SESSION INVARIANT:
//   - This digest identifies the logical conversation, not a particular upstream
//     account. Therefore it MUST NOT include account-scoped transport fields such
//     as thoughtSignature. Those signatures are regenerated whenever a request is
//     safely moved to another Google account.
//   - Message text, tool calls, tool responses, images, and thought markers remain
//     part of the digest. Removing any of those would let different conversations
//     collide and incorrectly inherit one another's account binding.
//
// Format: s:<hash>-u:<hash>-m:<hash>-u:<hash>-...
// s = systemInstruction, u = user, m = model
func BuildGeminiDigestChain(req *antigravity.GeminiRequest) string {
	if req == nil {
		return ""
	}

	var parts []string

	// 1. system instruction
	if req.SystemInstruction != nil && len(req.SystemInstruction.Parts) > 0 {
		partsData, _ := json.Marshal(canonicalGeminiPartsForDigest(req.SystemInstruction.Parts))
		parts = append(parts, "s:"+shortHash(partsData))
	}

	// 2. contents
	for _, c := range req.Contents {
		prefix := "u" // user
		if c.Role == "model" {
			prefix = "m"
		}
		partsData, _ := json.Marshal(canonicalGeminiPartsForDigest(c.Parts))
		parts = append(parts, prefix+":"+shortHash(partsData))
	}

	return strings.Join(parts, "-")
}

// canonicalGeminiPartsForDigest removes only fields that are tied to an upstream
// account instead of the logical conversation. Keep this deliberately narrow: it
// is the last-resort binding used when the primary sticky cache misses.
func canonicalGeminiPartsForDigest(parts []antigravity.GeminiPart) []antigravity.GeminiPart {
	canonical := make([]antigravity.GeminiPart, len(parts))
	copy(canonical, parts)
	for i := range canonical {
		canonical[i].ThoughtSignature = ""
	}
	return canonical
}

// GenerateCachePrefixHash generates a stable digest-store namespace shared by
// native Gemini, OpenAI-compatible chat clients, and agent clients.
//
// Do not reintroduce the observed client IP here. Public clients commonly arrive
// through Cloudflare or another reverse proxy whose egress IP changes between
// turns. Including it makes the digest fallback miss and can break an otherwise
// sticky conversation. API-key identity, normalized user agent, platform, and
// model provide the required isolation without depending on network topology.
//
// Composition: userID + apiKeyID + userAgent + platform + model.
// It deliberately excludes observed client IP because proxy egress addresses
// routinely change within one logical conversation.
func GenerateCachePrefixHash(userID, apiKeyID int64, userAgent, platform, model string) string {
	// 组合所有标识符
	normalizedUserAgent := NormalizeSessionUserAgent(userAgent)
	combined := strconv.FormatInt(userID, 10) + ":" +
		strconv.FormatInt(apiKeyID, 10) + ":" +
		normalizedUserAgent + ":" +
		platform + ":" +
		model

	hash := sha256.Sum256([]byte(combined))
	// 取前 12 字节，Base64 编码后正好 16 字符
	return base64.RawURLEncoding.EncodeToString(hash[:12])
}

// GenerateGeminiPrefixHash remains as a compatibility wrapper for existing
// native Gemini call sites. clientIP is intentionally ignored.
func GenerateGeminiPrefixHash(userID, apiKeyID int64, _ string, userAgent, platform, model string) string {
	return GenerateCachePrefixHash(userID, apiKeyID, userAgent, platform, model)
}

// ParseGeminiSessionValue 解析 Gemini 会话缓存值
// 格式: {uuid}:{accountID}
func ParseGeminiSessionValue(value string) (uuid string, accountID int64, ok bool) {
	if value == "" {
		return "", 0, false
	}

	// 找到最后一个 ":" 的位置（因为 uuid 可能包含 ":"）
	i := strings.LastIndex(value, ":")
	if i <= 0 || i >= len(value)-1 {
		return "", 0, false
	}

	uuid = value[:i]
	accountID, err := strconv.ParseInt(value[i+1:], 10, 64)
	if err != nil {
		return "", 0, false
	}

	return uuid, accountID, true
}

// FormatGeminiSessionValue 格式化 Gemini 会话缓存值
// 格式: {uuid}:{accountID}
func FormatGeminiSessionValue(uuid string, accountID int64) string {
	return uuid + ":" + strconv.FormatInt(accountID, 10)
}

// geminiDigestSessionKeyPrefix Gemini 摘要 fallback 会话 key 前缀
const geminiDigestSessionKeyPrefix = "gemini:digest:"

const cacheDigestSessionKeyPrefix = "cache:digest:"

// GenerateCacheDigestSessionKey produces the protocol-neutral sticky key for a
// digest lineage. It contains only shortened hashes/UUIDs and is safe to store
// in Redis logs without exposing request content.
func GenerateCacheDigestSessionKey(prefixHash, uuid string) string {
	prefix := prefixHash
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	uuidPart := uuid
	if len(uuidPart) > 8 {
		uuidPart = uuidPart[:8]
	}
	return cacheDigestSessionKeyPrefix + prefix + ":" + uuidPart
}

// GenerateGeminiDigestSessionKey 生成 Gemini 摘要 fallback 的 sessionKey
// 组合 prefixHash 前 8 位 + uuid 前 8 位，确保不同会话产生不同的 sessionKey
// 用于在 SelectAccountWithLoadAwareness 中保持粘性会话
func GenerateGeminiDigestSessionKey(prefixHash, uuid string) string {
	prefix := prefixHash
	if len(prefixHash) >= 8 {
		prefix = prefixHash[:8]
	}
	uuidPart := uuid
	if len(uuid) >= 8 {
		uuidPart = uuid[:8]
	}
	return geminiDigestSessionKeyPrefix + prefix + ":" + uuidPart
}
