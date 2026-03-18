package priming

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAISource_ShouldFailFastOnMissingConfig(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		apiKey          string
		model           string
		expectedMessage string
	}{
		{
			name:            "missing base url",
			baseURL:         "",
			apiKey:          "key",
			model:           "model",
			expectedMessage: "AI_BASE_URL is required",
		},
		{
			name:            "invalid base url",
			baseURL:         "ht!tp://invalid",
			apiKey:          "key",
			model:           "model",
			expectedMessage: "AI_BASE_URL must be a valid URL",
		},
		{
			name:            "missing api key",
			baseURL:         "https://example.test/v1",
			apiKey:          "",
			model:           "model",
			expectedMessage: "AI_API_KEY is required",
		},
		{
			name:            "missing model",
			baseURL:         "https://example.test/v1",
			apiKey:          "key",
			model:           "",
			expectedMessage: "AI_MODEL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AI_BASE_URL", tt.baseURL)
			t.Setenv("AI_API_KEY", tt.apiKey)
			t.Setenv("AI_MODEL", tt.model)
			t.Setenv("AI_MODEL_FALLBACK", "")

			source := NewAISource(&http.Client{}, nil)
			_, err := source.Next(context.Background())
			if err == nil {
				t.Fatal("expected config validation error")
			}
			if !strings.Contains(err.Error(), tt.expectedMessage) {
				t.Fatalf("expected %q in error, got %v", tt.expectedMessage, err)
			}
		})
	}
}

func TestAISource_ShouldRetryWithFallbackModel(t *testing.T) {
	t.Setenv("AI_BASE_URL", "https://example.test/v1")
	t.Setenv("AI_API_KEY", "key")
	t.Setenv("AI_MODEL", "primary")
	t.Setenv("AI_MODEL_FALLBACK", "fallback")

	requests := 0
	models := []string{}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		models = append(models, string(data))
		status := http.StatusOK
		body := `{"choices":[{"message":{"content":"fallback content"}}]}`
		if requests == 1 {
			status = http.StatusInternalServerError
			body = `{"error":"primary failed"}`
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	source := NewAISource(client, nil)
	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected fallback retry to succeed, got %v", err)
	}
	if text != "fallback content" {
		t.Fatalf("expected fallback content, got %q", text)
	}
	if requests != 2 {
		t.Fatalf("expected primary + fallback attempts, got %d", requests)
	}
	if !strings.Contains(models[0], `"model":"primary"`) || !strings.Contains(models[1], `"model":"fallback"`) {
		t.Fatalf("expected retry with fallback model, got requests %#v", models)
	}
}
