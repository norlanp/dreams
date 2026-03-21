package priming

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"dreams/internal/model"
)

type contentStore interface {
	ListPrimingContent(ctx context.Context) ([]model.PrimingContent, error)
	GetPrimingContentByCategory(ctx context.Context, category string) ([]model.PrimingContent, error)
	GetRandomDream(ctx context.Context) (*model.Dream, error)
}

// ContentSource provides community-sourced priming content with optional dream blending.
type ContentSource struct {
	store contentStore
	index atomic.Int64
	rand  *rand.Rand
}

// NewContentSource creates a new ContentSource with the given store.
// Uses a seeded random source for consistent behavior in tests.
func NewContentSource(store contentStore) *ContentSource {
	cs := &ContentSource{
		store: store,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	cs.index.Store(0)
	return cs
}

func (s *ContentSource) Label() SourceLabel {
	return SourceCommunity
}

func (s *ContentSource) Next(ctx context.Context) (string, error) {
	content, err := s.store.ListPrimingContent(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load priming content: %w", err)
	}

	if len(content) == 0 {
		return "", errSourceUnavailable
	}

	idx := int(s.index.Add(1)-1) % len(content)
	item := content[idx]

	return s.blendWithDreams(ctx, item)
}

// blendWithDreams attempts to blend community content with user's dreams.
// If dream loading fails or no dreams exist, falls back to content-only format.
func (s *ContentSource) blendWithDreams(ctx context.Context, item model.PrimingContent) (string, error) {
	dream, err := s.store.GetRandomDream(ctx)
	if err != nil {
		log.Printf(`{"event":"priming_blend","status":"dreams_fetch_failed","error":%q}`, err)
		return s.format(item, ""), nil
	}

	if dream == nil {
		log.Printf(`{"event":"priming_blend","status":"no_dreams"}`)
		return s.format(item, ""), nil
	}

	preview := dream.Content
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}

	return s.format(item, preview), nil
}

// format renders the priming content with optional dream preview.
func (s *ContentSource) format(item model.PrimingContent, dreamPreview string) string {
	var b strings.Builder
	b.WriteString(item.Title)
	b.WriteString("\n\n")
	b.WriteString(item.Content)

	if dreamPreview != "" {
		b.WriteString("\n\n")
		b.WriteString("— From your dream journal —\n")
		b.WriteString(dreamPreview)
	}

	if item.URL != "" {
		b.WriteString("\n\n")
		b.WriteString("Learn more: ")
		b.WriteString(item.URL)
	}

	b.WriteString("\n\n")
	b.WriteString("Bedtime focus: Dream journal recall, reality checks, and clear intention for lucidity.")

	return b.String()
}

// SetRand allows replacing the random source for testing.
func (s *ContentSource) SetRand(r *rand.Rand) {
	s.rand = r
}
