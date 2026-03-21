package priming

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"dreams/internal/model"
)

type analysisStore interface {
	GetLatestAnalysis(ctx context.Context) (*model.Analysis, error)
	GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error)
}

type PersonalizedSource struct {
	store analysisStore
	index atomic.Int64
}

func NewPersonalizedSource(store analysisStore) *PersonalizedSource {
	ps := &PersonalizedSource{
		store: store,
	}
	ps.index.Store(0)
	return ps
}

func (s *PersonalizedSource) Label() SourceLabel {
	return SourcePersonalized
}

func (s *PersonalizedSource) Next(ctx context.Context) (string, error) {
	terms, err := latestDreamSigns(ctx, s.store, 5)
	if err != nil {
		return "", err
	}

	if len(terms) == 0 {
		return "", errSourceUnavailable
	}

	content := s.formatWithRotation(terms)
	s.index.Add(1)
	return content, nil
}

func (s *PersonalizedSource) formatWithRotation(terms []string) string {
	templates := []string{
		"Tonight, when you notice %s, pause and ask: 'Am I dreaming?' Repeat this cue three times before sleep.",
		"As you drift off, keep %s in mind. When you encounter them in dreams, recognize the signal and become lucid.",
		"Your dream signs: %s. Set an intention to notice these tonight and question your reality when they appear.",
		"Tonight's focus: %s. When these appear in your dreams, use them as triggers to remember you're dreaming.",
		"Before sleep, visualize %s. Practice the habit of questioning reality whenever you encounter them in waking life.",
	}

	templateIdx := int(s.index.Load()) % len(templates)
	selectedTerms := s.selectTermsForRotation(terms)

	return fmt.Sprintf(templates[templateIdx], strings.Join(selectedTerms, ", "))
}

// selectTermsForRotation implements a sliding window rotation across available terms.
// When there are more than 3 terms, it cycles through different 3-term combinations
// to provide variety across successive calls.
func (s *PersonalizedSource) selectTermsForRotation(terms []string) []string {
	if len(terms) <= 3 {
		return terms
	}

	// Use modulo to cycle through starting positions, ensuring we always have
	// 3 consecutive terms. The window size (len - 2) ensures we don't overflow.
	startIdx := int(s.index.Load()) % (len(terms) - 2)
	endIdx := startIdx + 3
	if endIdx > len(terms) {
		endIdx = len(terms)
	}

	return terms[startIdx:endIdx]
}

func latestDreamSigns(ctx context.Context, store analysisStore, limit int) ([]string, error) {
	if store == nil {
		return nil, errSourceUnavailable
	}

	analysis, err := store.GetLatestAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load latest analysis: %w", err)
	}
	if analysis == nil {
		return nil, errSourceUnavailable
	}

	clusters, err := store.GetAnalysisClusters(ctx, analysis.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load analysis clusters: %w", err)
	}

	return collectTopTerms(clusters, limit), nil
}

func collectTopTerms(clusters []model.Cluster, limit int) []string {
	if limit <= 0 {
		return []string{}
	}

	seen := map[string]struct{}{}
	terms := make([]string, 0, limit)
	for _, cluster := range clusters {
		for _, term := range cluster.TopTerms {
			normalized := strings.TrimSpace(term)
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			terms = append(terms, normalized)
			if len(terms) == limit {
				return terms
			}
		}
	}

	return terms
}
