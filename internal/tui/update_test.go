package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelUpdate_ShouldEnterAnalysisViewOnSFromList(t *testing.T) {
	m := NewModel(nil)

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updated := updatedModel.(Model)

	if updated.state != analysisView {
		t.Fatalf("expected state %v, got %v", analysisView, updated.state)
	}
}

func TestModelUpdate_ShouldLeaveAnalysisViewOnEscAndQ(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{
			name: "esc returns to list",
			msg:  tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name: "q returns to list",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(nil)
			m.state = analysisView

			updatedModel, _ := m.Update(tt.msg)
			updated := updatedModel.(Model)

			if updated.state != listView {
				t.Fatalf("expected state %v, got %v", listView, updated.state)
			}
		})
	}
}

func TestModelUpdate_ShouldQuitOnCtrlCInAnalysisView(t *testing.T) {
	m := NewModel(nil)
	m.state = analysisView

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}

	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected quit message, got %T", cmd())
	}
}

func TestModelView_ShouldRenderAnalysisViewWhenStateIsAnalysis(t *testing.T) {
	m := NewModel(nil)
	m.state = analysisView

	view := m.View()
	if !strings.Contains(view, "Dream Statistics") {
		t.Fatalf("expected analysis view content, got %q", view)
	}
}

func TestModelUpdate_ShouldNotEnterAnalysisViewOnSFromNonListViews(t *testing.T) {
	tests := []struct {
		name  string
		state viewState
	}{
		{name: "create", state: createView},
		{name: "detail", state: detailView},
		{name: "search", state: searchView},
		{name: "analysis", state: analysisView},
		{name: "update", state: updateView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(nil)
			m.state = tt.state

			updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
			updated := updatedModel.(Model)

			if updated.state == analysisView && tt.state != analysisView {
				t.Fatalf("expected state to remain %v, got %v", tt.state, updated.state)
			}

			if updated.state != tt.state {
				t.Fatalf("expected state %v, got %v", tt.state, updated.state)
			}
		})
	}
}
