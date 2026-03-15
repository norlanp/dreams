package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"dreams/internal/model"
)

func TestModelUpdate_ShouldValidateStatisticsNavigationCacheAndRerunFlow(t *testing.T) {
	oldLocal := time.Local
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}
	time.Local = loc
	t.Cleanup(func() { time.Local = oldLocal })

	repo := &analysisTestRepo{
		latestAnalysis: &model.Analysis{
			ID:           19,
			AnalysisDate: time.Date(2025, 1, 15, 18, 30, 0, 0, time.UTC),
			DreamCount:   12,
			NClusters:    1,
		},
		analysisClusters: []model.Cluster{{
			ClusterID:  1,
			DreamCount: 12,
			TopTerms:   []string{"flight", "mirror"},
		}},
		listDreamsResult: makeDreams(6),
	}

	runnerCalls := 0
	m := NewModel(repo)
	m.analysisRunner = func(minDreams int) ([]byte, error) {
		runnerCalls++
		return []byte(`{"dream_count":6,"n_clusters":1,"clusters":[{"cluster_id":0,"dream_count":6,"top_terms":["ocean","stairs"],"dream_ids":[1,2,3,4,5,6]}]}`), nil
	}

	enteredModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	entered := enteredModel.(Model)

	if entered.state != analysisView {
		t.Fatalf("expected to enter analysis view, got %v", entered.state)
	}

	cacheView := entered.View()
	if !strings.Contains(cacheView, "Last analyzed: 2025-01-15 13:30 EST") {
		t.Fatalf("expected cached timestamp converted to local timezone, got %q", cacheView)
	}

	if !strings.Contains(cacheView, "Dreams analyzed: 12") {
		t.Fatalf("expected cached dream count on initial render, got %q", cacheView)
	}

	loadingModel, cmd := entered.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	loading := loadingModel.(Model)

	if cmd == nil {
		t.Fatal("expected rerun command")
	}

	if !loading.analysisLoading {
		t.Fatal("expected loading state while rerun is in progress")
	}

	loadingView := loading.View()
	if !strings.Contains(loadingView, "Running analysis...") {
		t.Fatalf("expected loading indicator after rerun trigger, got %q", loadingView)
	}

	rerunMsg := cmd()
	afterRerunModel, _ := loading.Update(rerunMsg)
	afterRerun := afterRerunModel.(Model)

	if runnerCalls != 1 {
		t.Fatalf("expected runner to execute once, got %d", runnerCalls)
	}

	if repo.saveAtomicCalls != 1 {
		t.Fatalf("expected one atomic save call, got %d", repo.saveAtomicCalls)
	}

	if afterRerun.analysisLoading {
		t.Fatal("expected loading state to clear after rerun")
	}

	rerunView := afterRerun.View()
	if !strings.Contains(rerunView, "Dreams analyzed: 6") {
		t.Fatalf("expected refreshed analysis data after rerun, got %q", rerunView)
	}

	if !strings.Contains(rerunView, "Cluster 0") {
		t.Fatalf("expected refreshed cluster data after rerun, got %q", rerunView)
	}

	listModel, _ := afterRerun.Update(tea.KeyMsg{Type: tea.KeyEsc})
	list := listModel.(Model)
	if list.state != listView {
		t.Fatalf("expected list view after escape, got %v", list.state)
	}

	reopenedModel, _ := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	reopened := reopenedModel.(Model)

	if reopened.state != analysisView {
		t.Fatalf("expected to reopen analysis view, got %v", reopened.state)
	}

	reopenedView := reopened.View()
	if !strings.Contains(reopenedView, "Dreams analyzed: 6") {
		t.Fatalf("expected rerun results loaded from cache on reopen, got %q", reopenedView)
	}

	if strings.Contains(reopenedView, "Running analysis...") {
		t.Fatalf("expected cached rerun results, got loading state: %q", reopenedView)
	}
}

func TestModelUpdate_ShouldShowLocalTimezoneTimestampWhenOpeningStatistics(t *testing.T) {
	oldLocal := time.Local
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}
	time.Local = loc
	t.Cleanup(func() { time.Local = oldLocal })

	repo := &analysisTestRepo{
		latestAnalysis: &model.Analysis{
			ID:           41,
			AnalysisDate: time.Date(2025, 1, 1, 23, 30, 0, 0, time.UTC),
			DreamCount:   5,
			NClusters:    1,
		},
	}

	m := NewModel(repo)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	view := updatedModel.(Model).View()

	if !strings.Contains(view, "Last analyzed: 2025-01-02 08:30 JST") {
		t.Fatalf("expected UTC timestamp converted to local timezone in view, got %q", view)
	}
}
