package priming

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type AISource struct {
	httpClient *http.Client
	store      analysisStore
}

type aiConfig struct {
	BaseURL       string
	APIKey        string
	Model         string
	FallbackModel string
}

func NewAISource(client *http.Client, store analysisStore) *AISource {
	if client == nil {
		client = http.DefaultClient
	}
	return &AISource{httpClient: client, store: store}
}

func (s *AISource) Label() SourceLabel {
	return SourceAI
}

func (s *AISource) Next(ctx context.Context) (string, error) {
	config, err := loadAIConfig()
	if err != nil {
		return "", err
	}

	terms, _ := latestDreamSigns(ctx, s.store, 3)
	prompt := buildAIPrompt(terms)

	content, err := s.chatCompletion(ctx, config.BaseURL, config.APIKey, config.Model, prompt)
	if err == nil {
		return content, nil
	}

	if config.FallbackModel == "" || config.FallbackModel == config.Model {
		return "", err
	}

	return s.chatCompletion(ctx, config.BaseURL, config.APIKey, config.FallbackModel, prompt)
}

func loadAIConfig() (*aiConfig, error) {
	baseURL := strings.TrimSpace(os.Getenv("AI_BASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))
	model := strings.TrimSpace(os.Getenv("AI_MODEL"))
	fallback := strings.TrimSpace(os.Getenv("AI_MODEL_FALLBACK"))

	if baseURL == "" {
		return nil, fmt.Errorf("AI_BASE_URL is required (example: https://api.openai.com/v1)")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("AI_BASE_URL must be a valid URL: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("AI_API_KEY is required; export AI_API_KEY before using AI priming")
	}
	if model == "" {
		return nil, fmt.Errorf("AI_MODEL is required (example: gpt-4o-mini)")
	}

	return &aiConfig{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: apiKey, Model: model, FallbackModel: fallback}, nil
}

func buildAIPrompt(terms []string) string {
	if len(terms) == 0 {
		return "Write a short, calming dream-lucidity priming paragraph for bedtime."
	}

	return fmt.Sprintf("Write a short bedtime priming paragraph that references these dream signs: %s.", strings.Join(terms, ", "))
}

func (s *AISource) chatCompletion(ctx context.Context, baseURL, apiKey, model, prompt string) (string, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type request struct {
		Model    string    `json:"model"`
		Messages []message `json:"messages"`
	}
	type response struct {
		Choices []struct {
			Message message `json:"message"`
		} `json:"choices"`
	}

	body, err := json.Marshal(request{Model: model, Messages: []message{{Role: "user", Content: prompt}}})
	if err != nil {
		return "", fmt.Errorf("failed to encode AI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create AI request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call AI provider: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read AI response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var parsed response
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("AI provider returned no choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("AI provider returned empty content")
	}

	return content, nil
}
