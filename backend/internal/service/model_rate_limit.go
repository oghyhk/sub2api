package service

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

const (
	modelRateLimitsKey                 = "model_rate_limits"
	antigravityGeminiModelRateLimitKey = "antigravity:gemini"
	openAIImageGenerationRateLimitKey  = "openai:image_generation"
	// anthropicFableRateLimitKey 是 Anthropic 7d_oi（Fable 专属 7d 窗口）限流的
	// 家族级 scope：命中后所有 Fable 变体（含 [1m] 等后缀）都不再调度到该账号。
	anthropicFableRateLimitKey = "claude-fable-5"
)

// isRateLimitActiveForKey 检查指定 key 的限流是否生效
func (a *Account) isRateLimitActiveForKey(key string) bool {
	resetAt := a.modelRateLimitResetAt(key)
	return resetAt != nil && time.Now().Before(*resetAt)
}

// getRateLimitRemainingForKey 获取指定 key 的限流剩余时间，0 表示未限流或已过期
func (a *Account) getRateLimitRemainingForKey(key string) time.Duration {
	resetAt := a.modelRateLimitResetAt(key)
	if resetAt == nil {
		return 0
	}
	remaining := time.Until(*resetAt)
	if remaining > 0 {
		return remaining
	}
	return 0
}

func (a *Account) isModelRateLimitedWithContext(ctx context.Context, requestedModel string) bool {
	for _, key := range a.modelRateLimitKeysForRequest(ctx, requestedModel) {
		if a.isRateLimitActiveForKey(key) {
			return true
		}
	}
	return false
}

// GetModelRateLimitRemainingTime 获取模型限流剩余时间
// 返回 0 表示未限流或已过期
func (a *Account) GetModelRateLimitRemainingTime(requestedModel string) time.Duration {
	return a.GetModelRateLimitRemainingTimeWithContext(context.Background(), requestedModel)
}

func (a *Account) GetModelRateLimitRemainingTimeWithContext(ctx context.Context, requestedModel string) time.Duration {
	remaining := time.Duration(0)
	for _, key := range a.modelRateLimitKeysForRequest(ctx, requestedModel) {
		if keyRemaining := a.getRateLimitRemainingForKey(key); keyRemaining > remaining {
			remaining = keyRemaining
		}
	}
	return remaining
}

func (a *Account) modelRateLimitKeysForRequest(ctx context.Context, requestedModel string) []string {
	if a == nil {
		return nil
	}

	modelKey := a.GetMappedModel(requestedModel)
	if a.Platform == PlatformAntigravity {
		modelKey = resolveFinalAntigravityModelKey(ctx, a, requestedModel)
	}
	modelKey = strings.TrimSpace(modelKey)
	if modelKey == "" {
		return nil
	}

	keys := []string{modelKey}
	switch a.Platform {
	case PlatformAntigravity:
		if isAntigravityGeminiModel(modelKey) && modelKey != antigravityGeminiModelRateLimitKey {
			keys = append(keys, antigravityGeminiModelRateLimitKey)
		}
	case PlatformOpenAI:
		if openAIImageGenerationRateLimitApplies(ctx, requestedModel, modelKey) && modelKey != openAIImageGenerationRateLimitKey {
			keys = append(keys, openAIImageGenerationRateLimitKey)
		}
	case PlatformAnthropic:
		if isAnthropicFableModel(modelKey) && modelKey != anthropicFableRateLimitKey {
			keys = append(keys, anthropicFableRateLimitKey)
		}
	}
	return keys
}

// isAnthropicFableModel 判断是否为 Fable 模型家族（claude-fable-5、claude-fable-5[1m] 等变体）
func isAnthropicFableModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "fable")
}

func openAIImageGenerationRateLimitApplies(ctx context.Context, requestedModel, modelKey string) bool {
	if isOpenAIImageGenerationModel(requestedModel) || isOpenAIImageGenerationModel(modelKey) {
		return true
	}
	return OpenAIImageGenerationIntentFromContext(ctx)
}

func WithOpenAIImageGenerationIntent(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxkey.OpenAIImageGenerationIntent, true)
}

func OpenAIImageGenerationIntentFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	enabled, ok := ctx.Value(ctxkey.OpenAIImageGenerationIntent).(bool)
	return ok && enabled
}

func resolveFinalAntigravityModelKey(ctx context.Context, account *Account, requestedModel string) string {
	modelKey := mapAntigravityModel(account, requestedModel)
	if modelKey == "" {
		return ""
	}
	// thinking 会影响 Antigravity 最终模型名（例如 claude-sonnet-4-5 -> claude-sonnet-4-5-thinking）
	if enabled, ok := ThinkingEnabledFromContext(ctx); ok {
		modelKey = applyThinkingModelSuffix(modelKey, enabled)
	}
	return modelKey
}

func isAntigravityGeminiModel(model string) bool {
	return strings.HasPrefix(normalizeAntigravityModelName(model), "gemini-")
}

func antigravityModelRateLimitKeys(model string) []string {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil
	}
	keys := []string{model}
	if isAntigravityGeminiModel(model) && model != antigravityGeminiModelRateLimitKey {
		keys = append(keys, antigravityGeminiModelRateLimitKey)
	}
	return keys
}

func (a *Account) modelRateLimitResetAt(scope string) *time.Time {
	if a == nil || a.Extra == nil || scope == "" {
		return nil
	}
	rawLimits, ok := a.Extra[modelRateLimitsKey].(map[string]any)
	if !ok {
		return nil
	}
	rawLimit, ok := rawLimits[scope].(map[string]any)
	if !ok {
		return nil
	}
	resetAtRaw, ok := rawLimit["rate_limit_reset_at"].(string)
	if !ok || strings.TrimSpace(resetAtRaw) == "" {
		return nil
	}
	resetAt, err := time.Parse(time.RFC3339, resetAtRaw)
	if err != nil {
		return nil
	}
	return &resetAt
}

// activeModelRateLimitResetAt 返回 match 命中的限流 key 中仍然生效的最晚
// rate_limit_reset_at；没有生效中的限流时返回 nil。
func (a *Account) activeModelRateLimitResetAt(match func(scope string) bool) *time.Time {
	if a == nil || a.Extra == nil {
		return nil
	}
	rawLimits, ok := a.Extra[modelRateLimitsKey].(map[string]any)
	if !ok {
		return nil
	}
	now := time.Now()
	var latest *time.Time
	for scope := range rawLimits {
		if match != nil && !match(scope) {
			continue
		}
		resetAt := a.modelRateLimitResetAt(scope)
		if resetAt == nil || !resetAt.After(now) {
			continue
		}
		if latest == nil || resetAt.After(*latest) {
			latest = resetAt
		}
	}
	return latest
}

// antigravityGeminiScopeKey 判断限流/额度 key 是否属于 Gemini 家族
// （family key 或 gemini-* 模型 key）。
func antigravityGeminiScopeKey(scope string) bool {
	return scope == antigravityGeminiModelRateLimitKey || strings.HasPrefix(scope, "gemini")
}

// antigravityClaudeScopeKey 判断限流/额度 key 是否属于 Claude/3p 家族
// （claude-* 模型 key 或 3p-* 窗口 key）。
func antigravityClaudeScopeKey(scope string) bool {
	return scope == "3p-5h" || scope == "3p-weekly" || strings.HasPrefix(scope, "claude")
}

// isAntigravityWeeklyWindowKey 判断额度条目是否为周窗口（不适用 5h 限流标记）。
func isAntigravityWeeklyWindowKey(scope string) bool {
	switch scope {
	case "gemini-weekly", "gemini:weekly", "3p-weekly", "claude:weekly":
		return true
	}
	return false
}
