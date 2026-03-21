package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"dreams/internal/model"
	"dreams/internal/priming"
)

type primingGeneratorStub struct {
	results []priming.Result
	calls   int
}

func (s *primingGeneratorStub) Next(ctx context.Context) priming.Result {
	_ = ctx
	if s.calls >= len(s.results) {
		return priming.Result{Err: context.DeadlineExceeded}
	}
	result := s.results[s.calls]
	s.calls++
	return result
}

func TestModelUpdate_ShouldOpenNightViewAndReturnToListWithSelectionPreserved(t *testing.T) {
	m := NewModel(nil, "test.db")
	m.dreams = []model.Dream{{ID: 1, CreatedAt: time.Now()}, {ID: 2, CreatedAt: time.Now()}}
	m.selected = 1
	m.nightGenerator = &primingGeneratorStub{results: []priming.Result{{
		Source: priming.SourceTemplate,
		Text:   "Priming text",
	}}}

	openedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	opened := openedModel.(Model)
	if opened.state != nightView {
		t.Fatalf("expected night view after p, got %v", opened.state)
	}
	if cmd == nil {
		t.Fatal("expected priming load command")
	}

	msg := cmd()
	afterLoadModel, _ := opened.Update(msg)
	afterLoad := afterLoadModel.(Model)
	if afterLoad.nightSourceLabel != string(priming.SourceTemplate) {
		t.Fatalf("expected source label to be set, got %q", afterLoad.nightSourceLabel)
	}

	listModel, _ := afterLoad.Update(tea.KeyMsg{Type: tea.KeyEsc})
	list := listModel.(Model)
	if list.state != listView {
		t.Fatalf("expected list view on esc, got %v", list.state)
	}
	if list.selected != 1 {
		t.Fatalf("expected list selection preserved, got %d", list.selected)
	}
}

func TestModelUpdate_ShouldRenderSourceLabelAndSupportNextKey(t *testing.T) {
	m := NewModel(nil, "test.db")
	m.state = nightView
	m.nightGenerator = &primingGeneratorStub{results: []priming.Result{
		{Source: priming.SourcePersonalized, Text: "first"},
		{Source: priming.SourceCommunity, Text: "second"},
	}}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	view := updated.View()
	if !strings.Contains(view, "Source: Personalized") {
		t.Fatalf("expected source label in night view, got %q", view)
	}

	updatedModel, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	msg = cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated = updatedModel.(Model)

	if updated.nightSourceLabel != string(priming.SourceCommunity) {
		t.Fatalf("expected second content source label, got %q", updated.nightSourceLabel)
	}
}

func TestModelView_ShouldRenderActionableFallbackStatusForAIConfigErrors(t *testing.T) {
	m := NewModel(nil, "test.db")
	m.state = nightView
	m.nightGenerator = &primingGeneratorStub{results: []priming.Result{{
		Source: priming.SourceTemplate,
		Text:   "template priming",
		Status: "Using Template fallback. Fix AI config: set AI_BASE_URL, AI_API_KEY, and AI_MODEL.",
	}}}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)

	view := updated.View()
	if !strings.Contains(view, "Source: Template") {
		t.Fatalf("expected template source label, got %q", view)
	}
	if !strings.Contains(view, "Fix AI config") {
		t.Fatalf("expected actionable AI guidance in degraded status, got %q", view)
	}
}

func TestModelUpdate_ShouldClearStaleNightContentAfterRefreshFailure(t *testing.T) {
	m := NewModel(nil, "test.db")
	m.state = nightView
	m.nightGenerator = &primingGeneratorStub{results: []priming.Result{
		{Source: priming.SourcePersonalized, Text: "fresh content"},
		{Source: priming.SourceCommunity, Err: context.DeadlineExceeded, Status: "Load failed. Press n to retry."},
	}}

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	msg := cmd()
	updatedModel, _ = updatedModel.(Model).Update(msg)
	updated := updatedModel.(Model)
	if updated.nightContent != "fresh content" {
		t.Fatalf("expected initial load content, got %q", updated.nightContent)
	}

	updatedModel, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated = updatedModel.(Model)
	if updated.nightContent != "" {
		t.Fatalf("expected refresh to clear prior content while loading, got %q", updated.nightContent)
	}

	msg = cmd()
	updatedModel, _ = updated.Update(msg)
	updated = updatedModel.(Model)
	if updated.nightContent != "" {
		t.Fatalf("expected failed refresh to keep content cleared, got %q", updated.nightContent)
	}
	if !strings.Contains(updated.View(), "Load failed. Press n to retry.") {
		t.Fatalf("expected failure status in night view, got %q", updated.View())
	}
}
