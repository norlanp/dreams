package priming

import (
	"context"
	"fmt"
	"strings"

	"dreams/internal/model"
)

type analysisStore interface {
	GetLatestAnalysis(ctx context.Context) (*model.Analysis, error)
	GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error)
}

type PersonalizedSource struct {
	store analysisStore
}

func NewPersonalizedSource(store analysisStore) *PersonalizedSource {
	return &PersonalizedSource{store: store}
}

func (s *PersonalizedSource) Label() SourceLabel {
	return SourcePersonalized
}

func (s *PersonalizedSource) Next(ctx context.Context) (string, error) {
	terms, err := latestDreamSigns(ctx, s.store, 3)
	if err != nil {
		return "", err
	}

	if len(terms) == 0 {
		return "", errSourceUnavailable
	}

	content := fmt.Sprintf(
		"Tonight, when you notice %s, pause and ask: 'Am I dreaming?' Repeat this cue three times before sleep.",
		strings.Join(terms, ", "),
	)
	return content, nil
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
		return nil
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
