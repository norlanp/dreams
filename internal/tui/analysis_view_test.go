package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"dreams/internal/model"
)

type analysisTestRepo struct {
	latestAnalysis   *model.Analysis
	latestErr        error
	analysisClusters []model.Cluster
	clustersErr      error
	listDreamsResult []model.Dream
	listDreamsErr    error
	saveAnalysisErr  error
	saveClusterErr   error
	saveAtomicErr    error
	saveAnalysisCall int
	saveClusterCalls int
	saveAtomicCalls  int
	saveCtxRemaining time.Duration
}

func (r *analysisTestRepo) CreateDream(ctx context.Context, content string) (*model.Dream, error) {
	return nil, nil
}

func (r *analysisTestRepo) ListDreams(ctx context.Context) ([]model.Dream, error) {
	return r.listDreamsResult, r.listDreamsErr
}

func (r *analysisTestRepo) GetDream(ctx context.Context, id int64) (*model.Dream, error) {
	return nil, nil
}

func (r *analysisTestRepo) UpdateDream(ctx context.Context, id int64, content string) (*model.Dream, error) {
	return nil, nil
}

func (r *analysisTestRepo) DeleteDream(ctx context.Context, id int64) error {
	return nil
}

func (r *analysisTestRepo) SearchDreams(ctx context.Context, query string) ([]model.Dream, error) {
	return nil, nil
}

func (r *analysisTestRepo) GetLatestAnalysis(ctx context.Context) (*model.Analysis, error) {
	return r.latestAnalysis, r.latestErr
}

func (r *analysisTestRepo) GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error) {
	return r.analysisClusters, r.clustersErr
}

func (r *analysisTestRepo) SaveAnalysis(ctx context.Context, analysisDate time.Time, dreamCount, nClusters int64, resultsJSON string) (*model.Analysis, error) {
	r.saveAnalysisCall++
	if r.saveAnalysisErr != nil {
		return nil, r.saveAnalysisErr
	}

	r.latestAnalysis = &model.Analysis{
		ID:           77,
		AnalysisDate: analysisDate,
		DreamCount:   dreamCount,
		NClusters:    nClusters,
		ResultsJSON:  resultsJSON,
	}
	return r.latestAnalysis, nil
}

func (r *analysisTestRepo) SaveCluster(ctx context.Context, analysisID, clusterID, dreamCount int64, topTerms, dreamIDs string) (*model.Cluster, error) {
	r.saveClusterCalls++
	if r.saveClusterErr != nil {
		return nil, r.saveClusterErr
	}

	cluster := model.Cluster{
		AnalysisID: analysisID,
		ClusterID:  clusterID,
		DreamCount: dreamCount,
	}
	if err := cluster.SetTopTermsFromJSON(topTerms); err != nil {
		return nil, err
	}
	if err := cluster.SetDreamIDsFromJSON(dreamIDs); err != nil {
		return nil, err
	}
	r.analysisClusters = append(r.analysisClusters, cluster)
	return &cluster, nil
}

func (r *analysisTestRepo) SaveAnalysisWithClusters(ctx context.Context, analysisDate time.Time, dreamCount, nClusters int64, resultsJSON string, clusters []model.Cluster) (*model.Analysis, error) {
	r.saveAtomicCalls++
	deadline, ok := ctx.Deadline()
	if ok {
		r.saveCtxRemaining = time.Until(deadline)
	}

	if r.saveAtomicErr != nil {
		return nil, r.saveAtomicErr
	}

	analysis, err := r.SaveAnalysis(ctx, analysisDate, dreamCount, nClusters, resultsJSON)
	if err != nil {
		return nil, err
	}

	r.analysisClusters = nil
	for _, cluster := range clusters {
		_, err := r.SaveCluster(ctx, analysis.ID, cluster.ClusterID, cluster.DreamCount, cluster.TopTermsJSON(), cluster.DreamIDsJSON())
		if err != nil {
			return nil, err
		}
	}

	return analysis, nil
}

func TestModelUpdate_ShouldRenderRepoCachedAnalysisOnFirstAnalysisViewRender(t *testing.T) {
	repo := &analysisTestRepo{
		latestAnalysis: &model.Analysis{
			ID:           9,
			AnalysisDate: time.Date(2025, 2, 8, 15, 30, 0, 0, time.UTC),
			DreamCount:   12,
			NClusters:    2,
		},
		analysisClusters: []model.Cluster{{
			ClusterID:  1,
			DreamCount: 5,
			TopTerms:   []string{"flight", "teeth"},
		}},
	}

	m := NewModel(repo)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updated := updatedModel.(Model)
	view := updated.View()

	if !strings.Contains(view, "Last analyzed:") {
		t.Fatalf("expected cached analysis timestamp in first render, got %q", view)
	}

	if !strings.Contains(view, "Dreams analyzed: 12") {
		t.Fatalf("expected dreams analyzed count in view, got %q", view)
	}

	if !strings.Contains(view, "Cluster 1") {
		t.Fatalf("expected cluster details in view, got %q", view)
	}

	if strings.Contains(view, "Loading cached analysis") {
		t.Fatalf("expected cached analysis content instead of loading state, got %q", view)
	}
}

func TestModelUpdate_ShouldShowEmptyFallbackWhenNoCachedAnalysis(t *testing.T) {
	m := NewModel(&analysisTestRepo{})
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	view := updatedModel.(Model).View()

	if !strings.Contains(view, "No cached analysis yet") {
		t.Fatalf("expected empty-state fallback, got %q", view)
	}
}

func TestModelUpdate_ShouldRenderCachedAnalysisImmediatelyOnAnalysisViewEntry(t *testing.T) {
	m := NewModel(&analysisTestRepo{})
	m.analysis = &model.Analysis{
		ID:           2,
		AnalysisDate: time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC),
		DreamCount:   7,
		NClusters:    2,
	}
	m.analysisClusters = []model.Cluster{{
		ClusterID:  1,
		DreamCount: 4,
		TopTerms:   []string{"water", "storm"},
	}}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	view := updatedModel.(Model).View()

	if !strings.Contains(view, "Dreams analyzed: 7") {
		t.Fatalf("expected cached analysis details on initial render, got %q", view)
	}

	if strings.Contains(view, "Loading cached analysis") {
		t.Fatalf("expected cached content instead of loading-only state, got %q", view)
	}
}

func TestModelUpdate_ShouldPreserveExistingCacheWhenEntryFetchFails(t *testing.T) {
	repo := &analysisTestRepo{latestErr: fmt.Errorf("sqlite busy")}

	m := NewModel(repo)
	m.analysis = &model.Analysis{
		ID:           88,
		AnalysisDate: time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
		DreamCount:   11,
		NClusters:    2,
	}
	m.analysisClusters = []model.Cluster{{
		ClusterID:  1,
		DreamCount: 6,
		TopTerms:   []string{"mountain", "snow"},
	}}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updated := updatedModel.(Model)

	if updated.analysisLoadErr == nil {
		t.Fatal("expected fetch failure to be surfaced")
	}

	if updated.analysis == nil || updated.analysis.ID != 88 {
		t.Fatalf("expected previous analysis to remain cached, got %#v", updated.analysis)
	}

	if len(updated.analysisClusters) != 1 {
		t.Fatalf("expected previous clusters to remain cached, got %d", len(updated.analysisClusters))
	}

	view := updated.View()
	if !strings.Contains(view, "Analysis unavailable.") {
		t.Fatalf("expected entry-load error message, got %q", view)
	}

	if !strings.Contains(view, "Showing last cached analysis:") {
		t.Fatalf("expected cached analysis notice on fetch error, got %q", view)
	}

	if !strings.Contains(view, "Dreams analyzed: 11") {
		t.Fatalf("expected cached analysis content after fetch error, got %q", view)
	}
}

func TestFormatAnalysisTimestamp_ShouldConvertUTCToLocalTime(t *testing.T) {
	oldLocal := time.Local
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}
	time.Local = loc
	t.Cleanup(func() { time.Local = oldLocal })

	utc := time.Date(2025, 1, 15, 18, 30, 0, 0, time.UTC)
	formatted := formatAnalysisTimestamp(utc)

	if formatted != "2025-01-15 13:30 EST" {
		t.Fatalf("expected converted local time, got %q", formatted)
	}
}

func TestModelUpdate_ShouldStartAsyncRerunOnRFromAnalysisView(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(5)}
	runnerCalls := 0

	m := NewModel(repo)
	m.state = analysisView
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		runnerCalls++
		return []byte(`{"dream_count":5,"n_clusters":1,"clusters":[]}`), nil
	}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedModel.(Model)

	if !updated.analysisLoading {
		t.Fatal("expected analysis loading state after pressing r")
	}

	if cmd == nil {
		t.Fatal("expected async rerun command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(analysisRerunMsg); !ok {
		t.Fatalf("expected rerun message, got %T", msg)
	}

	if runnerCalls != 1 {
		t.Fatalf("expected runner to execute once, got %d", runnerCalls)
	}
}

func TestModelUpdate_ShouldGuardMinimumDreamThresholdBeforeRunningPython(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(4)}
	runnerCalls := 0

	m := NewModel(repo)
	m.state = analysisView
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		runnerCalls++
		return nil, fmt.Errorf("should not run")
	}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedModel.(Model)

	msg := cmd()
	updatedModel, _ = updated.Update(msg)
	updated = updatedModel.(Model)

	if runnerCalls != 0 {
		t.Fatalf("expected runner to be skipped below threshold, got %d calls", runnerCalls)
	}

	if repo.saveAnalysisCall != 0 {
		t.Fatalf("expected no analysis persistence below threshold, got %d", repo.saveAnalysisCall)
	}

	if updated.analysisLoadErr == nil {
		t.Fatal("expected threshold error after rerun attempt")
	}

	view := updated.View()
	if !strings.Contains(view, "need at least 5 dreams") {
		t.Fatalf("expected actionable threshold message in analysis view, got %q", view)
	}

	if !strings.Contains(view, "Not enough dreams to run analysis.") {
		t.Fatalf("expected too-few-dreams state label, got %q", view)
	}
}

func TestModelUpdate_ShouldPersistAndRefreshAnalysisAfterSuccessfulRerun(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(6)}

	m := NewModel(repo)
	m.state = analysisView
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		return []byte(`{
			"dream_count": 6,
			"n_clusters": 2,
			"clusters": [
				{"cluster_id": 0, "dream_count": 4, "top_terms": ["water", "storm"], "dream_ids": [1,2,3,4]},
				{"cluster_id": 1, "dream_count": 2, "top_terms": ["school", "hall"], "dream_ids": [5,6]}
			]
		}`), nil
	}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected rerun command")
	}

	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	if repo.saveAnalysisCall != 1 {
		t.Fatalf("expected analysis persisted once, got %d", repo.saveAnalysisCall)
	}

	if repo.saveClusterCalls != 2 {
		t.Fatalf("expected two cluster persists, got %d", repo.saveClusterCalls)
	}

	if updated.analysis == nil {
		t.Fatal("expected refreshed analysis in model")
	}

	if updated.analysis.DreamCount != 6 {
		t.Fatalf("expected dream count 6, got %d", updated.analysis.DreamCount)
	}

	if len(updated.analysisClusters) != 2 {
		t.Fatalf("expected 2 refreshed clusters, got %d", len(updated.analysisClusters))
	}

	if updated.analysisLoading {
		t.Fatal("expected loading to finish after rerun")
	}
}

func TestModelUpdate_ShouldExposeRunnerFailureInAnalysisState(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(6)}

	m := NewModel(repo)
	m.state = analysisView
	m.analysis = &model.Analysis{
		ID:           41,
		AnalysisDate: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC),
		DreamCount:   7,
		NClusters:    2,
	}
	m.analysisClusters = []model.Cluster{{
		ClusterID:  1,
		DreamCount: 4,
		TopTerms:   []string{"forest", "river"},
	}}
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		return nil, fmt.Errorf("python runtime unavailable")
	}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	if updated.analysisLoadErr == nil {
		t.Fatal("expected runner failure to be surfaced")
	}

	if !strings.Contains(updated.analysisLoadErr.Error(), "python runtime unavailable") {
		t.Fatalf("expected runner error details, got %v", updated.analysisLoadErr)
	}

	view := updated.View()
	if !strings.Contains(view, "Analysis execution failed.") {
		t.Fatalf("expected execution-failure state label, got %q", view)
	}

	if !strings.Contains(view, "Showing last cached analysis:") {
		t.Fatalf("expected cached analysis notice after execution failure, got %q", view)
	}

	if !strings.Contains(view, "Dreams analyzed: 7") {
		t.Fatalf("expected cached analysis data after execution failure, got %q", view)
	}
}

func TestModelUpdate_ShouldExposeRunnerTimeoutInAnalysisState(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(6)}

	tmpDir := t.TempDir()
	uvPath := filepath.Join(tmpDir, "uv")
	uvScript := "#!/bin/sh\nsleep 1\n"
	if err := os.WriteFile(uvPath, []byte(uvScript), 0o755); err != nil {
		t.Fatalf("failed to create fake uv binary: %v", err)
	}

	oldPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", tmpDir+":"+oldPath); err != nil {
		t.Fatalf("failed to set PATH: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("PATH", oldPath)
	})

	oldRunnerTimeout := analysisRunnerTimeout
	analysisRunnerTimeout = 50 * time.Millisecond
	t.Cleanup(func() {
		analysisRunnerTimeout = oldRunnerTimeout
	})

	m := NewModel(repo)
	m.state = analysisView

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	if updated.analysisLoadErr == nil {
		t.Fatal("expected runner timeout to be surfaced")
	}

	if !strings.Contains(updated.analysisLoadErr.Error(), "timed out") {
		t.Fatalf("expected timeout details, got %v", updated.analysisLoadErr)
	}
}

func TestModelUpdate_ShouldExposeParseFailureInAnalysisState(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(6)}

	m := NewModel(repo)
	m.state = analysisView
	m.analysis = &model.Analysis{
		ID:           55,
		AnalysisDate: time.Date(2025, 1, 11, 9, 0, 0, 0, time.UTC),
		DreamCount:   8,
		NClusters:    3,
	}
	m.analysisClusters = []model.Cluster{{
		ClusterID:  2,
		DreamCount: 3,
		TopTerms:   []string{"school", "stairs"},
	}}
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		return []byte("{not-json"), nil
	}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	if updated.analysisLoadErr == nil {
		t.Fatal("expected parse failure to be surfaced")
	}

	if !strings.Contains(updated.analysisLoadErr.Error(), "failed to parse analysis output") {
		t.Fatalf("expected parse error context, got %v", updated.analysisLoadErr)
	}

	view := updated.View()
	if !strings.Contains(view, "Failed to parse analysis results.") {
		t.Fatalf("expected parse-failure state label, got %q", view)
	}

	if !strings.Contains(view, "Showing last cached analysis:") {
		t.Fatalf("expected cached analysis notice after parse failure, got %q", view)
	}

	if !strings.Contains(view, "Dreams analyzed: 8") {
		t.Fatalf("expected cached analysis data after parse failure, got %q", view)
	}
}

func TestModelView_ShouldRenderLoadingStateInAnalysisView(t *testing.T) {
	m := NewModel(&analysisTestRepo{})
	m.state = analysisView
	m.analysisLoading = true

	view := m.View()

	if !strings.Contains(view, "Running analysis...") {
		t.Fatalf("expected loading state text, got %q", view)
	}
}

func TestModelUpdate_ShouldExposeSaveFailureAndKeepPreviousCache(t *testing.T) {
	repo := &analysisTestRepo{
		listDreamsResult: makeDreams(6),
		saveAtomicErr:    fmt.Errorf("disk is read-only"),
		latestAnalysis: &model.Analysis{
			ID:           12,
			AnalysisDate: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC),
			DreamCount:   4,
			NClusters:    1,
		},
	}

	m := NewModel(repo)
	m.state = analysisView
	m.analysis = repo.latestAnalysis
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		return []byte(`{"dream_count": 6, "n_clusters": 1, "clusters": [{"cluster_id": 0, "dream_count": 6, "top_terms": ["forest"], "dream_ids": [1,2,3,4,5,6]}]}`), nil
	}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	if updated.analysisLoadErr == nil {
		t.Fatal("expected save failure to be surfaced")
	}

	if !strings.Contains(updated.analysisLoadErr.Error(), "failed to persist analysis") {
		t.Fatalf("expected persistence context in error, got %v", updated.analysisLoadErr)
	}

	if updated.analysis == nil || updated.analysis.ID != 12 {
		t.Fatalf("expected previous cached analysis to remain, got %#v", updated.analysis)
	}
}

func TestRerunAnalysis_ShouldUseFreshSaveTimeoutAfterSlowRunner(t *testing.T) {
	repo := &analysisTestRepo{listDreamsResult: makeDreams(6)}
	runnerDelay := 300 * time.Millisecond

	cmd := rerunAnalysis(repo, 5, func(minDreams int) ([]byte, error) {
		time.Sleep(runnerDelay)
		return []byte(`{"dream_count":6,"n_clusters":1,"clusters":[{"cluster_id":1,"dream_count":6,"top_terms":["sky"],"dream_ids":[1,2,3,4,5,6]}]}`), nil
	})

	msg := cmd().(analysisRerunMsg)
	if msg.err != nil {
		t.Fatalf("expected rerun to succeed, got %v", msg.err)
	}

	if repo.saveAtomicCalls != 1 {
		t.Fatalf("expected one atomic save call, got %d", repo.saveAtomicCalls)
	}

	minimumExpected := analysisSaveTimeout - (runnerDelay / 2)
	if repo.saveCtxRemaining < minimumExpected {
		t.Fatalf("expected fresh save timeout after runner delay, remaining=%v minimum=%v", repo.saveCtxRemaining, minimumExpected)
	}
}

func makeDreams(n int) []model.Dream {
	dreams := make([]model.Dream, n)
	for i := range n {
		dreams[i] = model.Dream{ID: int64(i + 1), Content: "dream"}
	}
	return dreams
}
