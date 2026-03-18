package priming

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestNightPrimingIntegration_ShouldShowConfigGuidanceOnTemplateFallback(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		apiKey  string
		model   string
	}{
		{
			name:    "missing api key",
			baseURL: "https://example.test/v1",
			apiKey:  "",
			model:   "gpt-4o-mini",
		},
		{
			name:    "invalid base url",
			baseURL: "ht!tp://bad-url",
			apiKey:  "key",
			model:   "gpt-4o-mini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AI_BASE_URL", tt.baseURL)
			t.Setenv("AI_API_KEY", tt.apiKey)
			t.Setenv("AI_MODEL", tt.model)
			t.Setenv("AI_MODEL_FALLBACK", "")

			generator := NewGenerator(
				nil,
				&fakeSource{label: SourcePersonalized, err: errors.New("no clusters")},
				&fakeSource{label: SourceCommunity, err: errors.New("network unavailable")},
				NewAISource(&http.Client{}, nil),
				NewTemplateSource(),
			)

			result := generator.Next(context.Background())
			if result.Err != nil {
				t.Fatalf("expected template fallback success, got %v", result.Err)
			}
			if result.Source != SourceTemplate {
				t.Fatalf("expected template source, got %s", result.Source)
			}
			if strings.TrimSpace(result.Text) == "" {
				t.Fatal("expected template content")
			}
			if !strings.Contains(result.Status, "Using Template fallback") {
				t.Fatalf("expected fallback source status, got %q", result.Status)
			}
			if !strings.Contains(result.Status, "Fix AI config") {
				t.Fatalf("expected actionable config guidance, got %q", result.Status)
			}
		})
	}
}
