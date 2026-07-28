//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardAsChatCompletionsAntigravityUsesGeminiTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)
	body := []byte(`{
		"model":"gemini-3.6-flash",
		"messages":[{"role":"user","content":"hello"}],
		"stream":false
	}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	upstreamBody := []byte("data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"VPS2_CHAT_OK\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":2,\"candidatesTokenCount\":3}}}\n\n")
	upstream := &queuedHTTPUpstreamStub{
		responses: []*http.Response{{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(bytes.NewReader(upstreamBody)),
		}},
	}
	cfg := &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}
	antigravityService := &AntigravityGatewayService{
		settingService: NewSettingService(&antigravitySettingRepoStub{}, cfg),
		tokenProvider:  &AntigravityTokenProvider{},
		httpUpstream:   upstream,
	}
	svc := &GeminiMessagesCompatService{
		antigravityGatewayService: antigravityService,
		cfg:                       cfg,
	}
	account := &Account{
		ID:          101,
		Name:        "google-antigravity",
		Platform:    PlatformAntigravity,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token": "token",
			"project_id":   "configured-project",
			"model_mapping": map[string]any{
				"gemini-3.6-flash": "gemini-3.6-flash-high",
			},
		},
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, http.StatusOK, writer.Code)
	require.Equal(t, "gemini-3.6-flash-high", result.UpstreamModel)
	require.Len(t, upstream.requestBodies, 1)

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
	require.Len(t, response.Choices, 1)
	require.Equal(t, "VPS2_CHAT_OK", response.Choices[0].Message.Content)
}
