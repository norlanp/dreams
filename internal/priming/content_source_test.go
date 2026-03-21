package priming

import (
	"context"
	"strings"
	"testing"

	"dreams/internal/model"
)

type mockContentStore struct {
	content     []model.PrimingContent
	dreams      []model.Dream
	randomDream *model.Dream
	listErr     error
	getByCat    error
	randomErr   error
}

func (m *mockContentStore) ListPrimingContent(ctx context.Context) ([]model.PrimingContent, error) {
	return m.content, m.listErr
}

func (m *mockContentStore) GetPrimingContentByCategory(ctx context.Context, category string) ([]model.PrimingContent, error) {
	return m.content, m.getByCat
}

func (m *mockContentStore) GetRandomDream(ctx context.Context) (*model.Dream, error) {
	if m.randomDream != nil {
		return m.randomDream, m.randomErr
	}
	if len(m.dreams) > 0 {
		return &m.dreams[0], m.randomErr
	}
	return nil, m.randomErr
}

func TestContentSource_Next_ShouldReturnContent(t *testing.T) {
	store := &mockContentStore{
		content: []model.PrimingContent{
			{
				ID:       1,
				Title:    "Test Title",
				Content:  "Test content here",
				Category: "test",
				URL:      "http://example.com",
			},
		},
	}

	source := NewContentSource(store)

	result, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(result, "Test Title") {
		t.Errorf("expected result to contain title, got %q", result)
	}

	if !strings.Contains(result, "Test content here") {
		t.Errorf("expected result to contain content, got %q", result)
	}

	if !strings.Contains(result, "Bedtime focus") {
		t.Errorf("expected result to contain bedtime focus, got %q", result)
	}
}

func TestContentSource_Next_ShouldBlendWithDreams(t *testing.T) {
	store := &mockContentStore{
		content: []model.PrimingContent{
			{
				ID:       1,
				Title:    "Test Title",
				Content:  "Test content here",
				Category: "test",
				URL:      "http://example.com",
			},
		},
		dreams: []model.Dream{
			{ID: 1, Content: "I was flying over mountains"},
		},
	}

	source := NewContentSource(store)

	result, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(result, "Test Title") {
		t.Errorf("expected result to contain title, got %q", result)
	}

	if !strings.Contains(result, "From your dream journal") {
		t.Errorf("expected result to contain dream journal section, got %q", result)
	}
}

func TestContentSource_Next_ShouldReturnErrorWhenNoContent(t *testing.T) {
	store := &mockContentStore{
		content: []model.PrimingContent{},
	}

	source := NewContentSource(store)

	_, err := source.Next(context.Background())
	if err == nil {
		t.Fatal("expected error when no content available")
	}
}

func TestContentSource_Label_ShouldReturnCommunity(t *testing.T) {
	source := NewContentSource(&mockContentStore{})
	if source.Label() != SourceCommunity {
		t.Errorf("expected label to be Community, got %q", source.Label())
	}
}

func TestContentSource_Next_ShouldCycleThroughContent(t *testing.T) {
	store := &mockContentStore{
		content: []model.PrimingContent{
			{ID: 1, Title: "First", Content: "First content"},
			{ID: 2, Title: "Second", Content: "Second content"},
			{ID: 3, Title: "Third", Content: "Third content"},
		},
	}

	source := NewContentSource(store)

	result1, _ := source.Next(context.Background())
	result2, _ := source.Next(context.Background())
	result3, _ := source.Next(context.Background())
	result4, _ := source.Next(context.Background())

	if !strings.Contains(result1, "First") {
		t.Errorf("expected first result to contain 'First', got %q", result1)
	}

	if !strings.Contains(result2, "Second") {
		t.Errorf("expected second result to contain 'Second', got %q", result2)
	}

	if !strings.Contains(result3, "Third") {
		t.Errorf("expected third result to contain 'Third', got %q", result3)
	}

	if !strings.Contains(result4, "First") {
		t.Errorf("expected fourth result to cycle back to 'First', got %q", result4)
	}
}
