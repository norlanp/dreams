package priming

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dreams/internal/model"
)

type fakeSource struct {
	label SourceLabel
	text  string
	err   error
	calls int
}

func (s *fakeSource) Label() SourceLabel {
	return s.label
}

func (s *fakeSource) Next(ctx context.Context) (string, error) {
	_ = ctx
	s.calls++
	return s.text, s.err
}

type fakeLogStore struct {
	logs []model.PrimingLog
}

func (s *fakeLogStore) SavePrimingLog(ctx context.Context, source, outcome, detail, content string, createdAt time.Time) error {
	_ = ctx
	s.logs = append(s.logs, model.PrimingLog{
		CreatedAt: createdAt,
		Source:    source,
		Outcome:   outcome,
		Detail:    detail,
		Content:   content,
	})
	return nil
}

func TestGenerator_ShouldRespectStrictFallbackOrder(t *testing.T) {
	tests := []struct {
		name           string
		sources        []Source
		expectedSource SourceLabel
		expectedStatus string
		expectedCalls  map[SourceLabel]int
	}{
		{
			name: "first source success",
			sources: []Source{
				&fakeSource{label: SourcePersonalized, text: "personalized"},
				&fakeSource{label: SourceCommunity, text: "community"},
			},
			expectedSource: SourcePersonalized,
			expectedStatus: "",
			expectedCalls: map[SourceLabel]int{
				SourcePersonalized: 1,
				SourceCommunity:    0,
			},
		},
		{
			name: "falls back to community",
			sources: []Source{
				&fakeSource{label: SourcePersonalized, err: errors.New("no clusters")},
				&fakeSource{label: SourceCommunity, text: "community"},
				&fakeSource{label: SourceAI, text: "ai"},
			},
			expectedSource: SourceCommunity,
			expectedStatus: "Using Community fallback after upstream source errors.",
			expectedCalls: map[SourceLabel]int{
				SourcePersonalized: 1,
				SourceCommunity:    1,
				SourceAI:           0,
			},
		},
		{
			name: "falls back to ai",
			sources: []Source{
				&fakeSource{label: SourcePersonalized, err: errors.New("no clusters")},
				&fakeSource{label: SourceCommunity, err: errors.New("network down")},
				&fakeSource{label: SourceAI, text: "AI content"},
				&fakeSource{label: SourceTemplate, text: "template"},
			},
			expectedSource: SourceAI,
			expectedStatus: "Using AI Generated fallback after upstream source errors.",
			expectedCalls: map[SourceLabel]int{
				SourcePersonalized: 1,
				SourceCommunity:    1,
				SourceAI:           1,
				SourceTemplate:     0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLogStore{}
			generator := NewGenerator(store, tt.sources...)
			result := generator.Next(context.Background())

			if result.Err != nil {
				t.Fatalf("expected fallback success, got %v", result.Err)
			}
			if result.Source != tt.expectedSource {
				t.Fatalf("expected source %s, got %s", tt.expectedSource, result.Source)
			}
			if result.Status != tt.expectedStatus {
				t.Fatalf("expected status %q, got %q", tt.expectedStatus, result.Status)
			}
			if len(store.logs) != 1 || store.logs[0].Source != string(tt.expectedSource) {
				t.Fatalf("expected one persisted success log for final source, got %#v", store.logs)
			}

			for _, source := range tt.sources {
				fake := source.(*fakeSource)
				expectedCalls := tt.expectedCalls[fake.label]
				if fake.calls != expectedCalls {
					t.Fatalf("expected %s calls %d, got %d", fake.label, expectedCalls, fake.calls)
				}
			}
		})
	}
}

func TestPersonalizedSource_ShouldUseLatestClusterTerms(t *testing.T) {
	store := &analysisStub{
		analysis: &model.Analysis{ID: 7},
		clusters: []model.Cluster{{TopTerms: []string{"mirror", "stairs"}}},
	}

	source := NewPersonalizedSource(store)
	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected personalized source to succeed: %v", err)
	}
	if !strings.Contains(text, "mirror") {
		t.Fatalf("expected personalized text to contain top term, got %q", text)
	}
}

func TestPersonalizedSource_ShouldRotateContentOnSuccessiveCalls(t *testing.T) {
	store := &analysisStub{
		analysis: &model.Analysis{ID: 7},
		clusters: []model.Cluster{{TopTerms: []string{"mirror", "stairs", "flight", "ocean", "teeth"}}},
	}

	source := NewPersonalizedSource(store)

	text1, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected first call to succeed: %v", err)
	}

	text2, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected second call to succeed: %v", err)
	}

	if text1 == text2 {
		t.Fatalf("expected different content on successive calls, got same: %q", text1)
	}

	text3, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected third call to succeed: %v", err)
	}

	if text1 == text3 || text2 == text3 {
		t.Fatalf("expected different content on third call")
	}
}

func TestGenerator_ShouldReturnTerminalErrorWhenAllSourcesFail(t *testing.T) {
	generator := NewGenerator(
		nil,
		&fakeSource{label: SourcePersonalized, err: errors.New("missing")},
		&fakeSource{label: SourceCommunity, err: errors.New("missing")},
	)

	result := generator.Next(context.Background())
	if result.Err == nil {
		t.Fatal("expected terminal failure")
	}
	if result.Status == "" {
		t.Fatal("expected actionable terminal status")
	}
}

func TestGenerator_ShouldSurfaceActionableAIConfigGuidanceWhenFallbackSucceeds(t *testing.T) {
	generator := NewGenerator(
		nil,
		&fakeSource{label: SourcePersonalized, err: errors.New("missing")},
		&fakeSource{label: SourceCommunity, err: errors.New("missing")},
		&fakeSource{label: SourceAI, err: errors.New("AI_API_KEY is required; export AI_API_KEY before using AI priming")},
		&fakeSource{label: SourceTemplate, text: "template"},
	)

	result := generator.Next(context.Background())
	if result.Err != nil {
		t.Fatalf("expected fallback success, got %v", result.Err)
	}
	if result.Source != SourceTemplate {
		t.Fatalf("expected template fallback source, got %s", result.Source)
	}
	if !strings.Contains(result.Status, "Using Template fallback") {
		t.Fatalf("expected fallback source in status, got %q", result.Status)
	}
	if !strings.Contains(result.Status, "Fix AI config") {
		t.Fatalf("expected actionable config guidance, got %q", result.Status)
	}
}

type analysisStub struct {
	analysis *model.Analysis
	clusters []model.Cluster
	err      error
}

func (s *analysisStub) GetLatestAnalysis(ctx context.Context) (*model.Analysis, error) {
	_ = ctx
	return s.analysis, s.err
}

func (s *analysisStub) GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error) {
	_ = ctx
	_ = analysisID
	return s.clusters, s.err
}
