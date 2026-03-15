package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"

	"dreams/internal/model"
)

type dreamsLoadedMsg struct {
	dreams []model.Dream
	err    error
}

type dreamSavedMsg struct {
	dream         *model.Dream
	err           error
	exitAfterSave bool
}

type dreamDeletedMsg struct {
	err error
}

type editorClosedMsg struct {
	content string
	changed bool
	err     error
}

func loadDreams(r repo) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dreams, err := r.ListDreams(ctx)
		return dreamsLoadedMsg{dreams: dreams, err: err}
	}
}

func saveDream(r repo, existing *model.Dream, content string, exitAfterSave bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var (
			dream *model.Dream
			err   error
		)

		if existing == nil {
			dream, err = r.CreateDream(ctx, content)
		} else {
			dream, err = r.UpdateDream(ctx, existing.ID, content)
		}

		return dreamSavedMsg{dream: dream, err: err, exitAfterSave: exitAfterSave}
	}
}

func openExternalEditorCmd(content string) (tea.Cmd, error) {
	if err := os.MkdirAll("./tmp", 0o755); err != nil {
		return nil, fmt.Errorf("failed to create tmp directory: %w", err)
	}

	file, err := os.CreateTemp("./tmp", "dream-editor-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create editor file: %w", err)
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close editor file: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("failed to seed editor file: %w", err)
	}

	bin, args := configuredEditorCommand()
	args = append(args, path)
	cmd := exec.Command(bin, args...)

	return tea.ExecProcess(cmd, func(runErr error) tea.Msg {
		defer func() { _ = os.Remove(path) }()

		if runErr != nil {
			return editorClosedMsg{err: fmt.Errorf("editor exited with error: %w", runErr)}
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return editorClosedMsg{err: fmt.Errorf("failed to read editor output: %w", readErr)}
		}

		next := string(data)
		return editorClosedMsg{content: next, changed: next != content}
	}), nil
}

func configuredEditorCommand() (string, []string) {
	editor := strings.TrimSpace(os.Getenv("DREAMS_EDITOR"))
	if editor == "" {
		editor = strings.TrimSpace(os.Getenv("VISUAL"))
	}
	if editor == "" {
		editor = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if editor == "" {
		editor = "nvim"
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return "nvim", nil
	}

	return parts[0], parts[1:]
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case listView:
			return m.handleListKeys(msg)
		case createView:
			return m.handleCreateKeys(msg)
		case detailView:
			return m.handleDetailKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dreamsLoadedMsg:
		if msg.err != nil {
			m.error = msg.err
		} else {
			m.dreams = msg.dreams
		}
		return m, nil

	case dreamSavedMsg:
		if msg.err != nil {
			m.statusMessage = "Save failed: " + msg.err.Error()
			return m, nil
		}

		m.editingDream = msg.dream
		m.statusMessage = "Saved."

		if msg.exitAfterSave {
			m = m.resetCreateForm()
			m.state = listView
			return m, loadDreams(m.repo)
		}

		return m, nil

	case editorClosedMsg:
		if msg.err != nil {
			m.statusMessage = msg.err.Error()
			return m, nil
		}

		if msg.changed {
			m.contentInput.SetValue(msg.content)
			m.statusMessage = "Imported editor changes."
		} else {
			m.statusMessage = "Editor closed without changes."
		}

		focusCmd := m.contentInput.Focus()
		mode := cursor.CursorStatic
		if m.contentInsertMode {
			mode = cursor.CursorBlink
		}
		modeCmd := m.contentInput.Cursor.SetMode(mode)
		return m, tea.Batch(focusCmd, modeCmd)

	case dreamDeletedMsg:
		if msg.err != nil {
			m.error = msg.err
		} else {
			m.confirmDelete = false
			m.confirmDeleteYes = false
			m.state = listView
			return m, loadDreams(m.repo)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "n":
		m = m.resetCreateForm()
		m.state = createView
		return m, m.contentInput.Focus()
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil
	case "down", "j":
		if m.selected < len(m.dreams)-1 {
			m.selected++
		}
		return m, nil
	case "enter":
		if len(m.dreams) > 0 {
			m.confirmDelete = false
			m.confirmDeleteYes = false
			m.state = detailView
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.commandMode {
		switch msg.Type {
		case tea.KeyEsc:
			m.commandMode = false
			m.commandInput = ""
			return m, nil
		case tea.KeyBackspace, tea.KeyDelete:
			runes := []rune(m.commandInput)
			if len(runes) > 0 {
				m.commandInput = string(runes[:len(runes)-1])
			}
			return m, nil
		case tea.KeyEnter:
			return m.executeCreateCommand()
		case tea.KeyRunes:
			m.commandInput += string(msg.Runes)
			return m, nil
		default:
			return m, nil
		}
	}

	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	if msg.Type == tea.KeyEsc {
		if m.contentInsertMode {
			m.contentInsertMode = false
			m.statusMessage = ""
			m.pendingDeleteOp = false
			focusCmd := m.contentInput.Focus()
			modeCmd := m.contentInput.Cursor.SetMode(cursor.CursorStatic)
			return m, tea.Batch(focusCmd, modeCmd)
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ':' {
		if !m.contentInsertMode {
			m.commandMode = true
			m.commandInput = ""
			m.statusMessage = ""
			m.pendingDeleteOp = false
			return m, nil
		}
	}

	if !m.contentInsertMode {
		return m.handleContentNormalModeKeys(msg)
	}

	var cmd tea.Cmd
	m.contentInput, cmd = m.contentInput.Update(msg)
	return m, cmd
}

func (m Model) executeCreateCommand() (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(m.commandInput)
	m.commandMode = false
	m.commandInput = ""

	if cmd == "" {
		return m, nil
	}

	content := m.contentInput.Value()

	switch cmd {
	case "w":
		return m, saveDream(m.repo, m.editingDream, content, false)
	case "wq":
		return m, saveDream(m.repo, m.editingDream, content, true)
	case "q":
		m = m.resetCreateForm()
		m.state = listView
		return m, loadDreams(m.repo)
	case "e":
		editorCmd, err := openExternalEditorCmd(content)
		if err != nil {
			m.statusMessage = err.Error()
			return m, nil
		}
		m.statusMessage = ""
		return m, editorCmd
	default:
		m.statusMessage = "Not an editor command: :" + cmd
		return m, nil
	}
}

func (m Model) handleContentNormalModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()

	if key != "d" {
		m.pendingDeleteOp = false
		if m.statusMessage == "d" {
			m.statusMessage = ""
		}
	}

	switch key {
	case "d":
		if m.pendingDeleteOp {
			m.pendingDeleteOp = false
			m.statusMessage = ""
			return m.deleteCurrentContentLine(), nil
		}
		m.pendingDeleteOp = true
		m.statusMessage = "d"
		return m, nil
	case "i":
		m.contentInsertMode = true
		m.statusMessage = ""
		m.pendingDeleteOp = false
		focusCmd := m.contentInput.Focus()
		modeCmd := m.contentInput.Cursor.SetMode(cursor.CursorBlink)
		return m, tea.Batch(focusCmd, modeCmd)
	case "a":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyRight})
		m.contentInsertMode = true
		m.statusMessage = ""
		m.pendingDeleteOp = false
		focusCmd := m.contentInput.Focus()
		modeCmd := m.contentInput.Cursor.SetMode(cursor.CursorBlink)
		return m, tea.Batch(cmd, focusCmd, modeCmd)
	case "o":
		m.contentInput.CursorEnd()
		m.contentInput.InsertString("\n")
		m.contentInsertMode = true
		m.statusMessage = ""
		m.pendingDeleteOp = false
		focusCmd := m.contentInput.Focus()
		modeCmd := m.contentInput.Cursor.SetMode(cursor.CursorBlink)
		return m, tea.Batch(focusCmd, modeCmd)
	case "O":
		m.contentInput.CursorStart()
		m.contentInput.InsertString("\n")
		m.contentInput.CursorUp()
		m.contentInsertMode = true
		m.statusMessage = ""
		m.pendingDeleteOp = false
		focusCmd := m.contentInput.Focus()
		modeCmd := m.contentInput.Cursor.SetMode(cursor.CursorBlink)
		return m, tea.Batch(focusCmd, modeCmd)
	case "h":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyLeft})
		return m, cmd
	case "j":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyDown})
		return m, cmd
	case "k":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyUp})
		return m, cmd
	case "l":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyRight})
		return m, cmd
	case "0":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyHome})
		return m, cmd
	case "$":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyEnd})
		return m, cmd
	case "w":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}, Alt: true})
		return m, cmd
	case "b":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}, Alt: true})
		return m, cmd
	case "x":
		m.contentInput, cmd = m.contentInput.Update(tea.KeyMsg{Type: tea.KeyDelete})
		return m, cmd
	}

	return m, nil
}

func (m Model) deleteCurrentContentLine() Model {
	value := m.contentInput.Value()
	lines := strings.Split(value, "\n")
	if len(lines) == 0 {
		return m
	}

	row := m.contentInput.Line()
	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}

	lines = append(lines[:row], lines[row+1:]...)
	newValue := ""
	if len(lines) > 0 {
		newValue = strings.Join(lines, "\n")
	}

	m.contentInput.SetValue(newValue)

	targetRow := row
	if targetRow >= len(lines) {
		targetRow = len(lines) - 1
	}
	if targetRow < 0 {
		targetRow = 0
	}

	m = m.setContentCursor(targetRow, 0)
	return m
}

func (m Model) setContentCursor(row, col int) Model {
	if row < 0 {
		row = 0
	}

	for i := 0; i < row; i++ {
		m.contentInput.CursorDown()
	}

	if col < 0 {
		col = 0
	}
	m.contentInput.SetCursor(col)

	return m
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmDelete {
		switch msg.String() {
		case "tab", "left", "right", "h", "l":
			m.confirmDeleteYes = !m.confirmDeleteYes
			return m, nil
		case "y":
			if len(m.dreams) > 0 && m.selected < len(m.dreams) {
				id := m.dreams[m.selected].ID
				m.confirmDelete = false
				m.confirmDeleteYes = false
				return m, deleteDream(m.repo, id)
			}
			m.confirmDelete = false
			m.confirmDeleteYes = false
			return m, nil
		case "n", "esc":
			m.confirmDelete = false
			m.confirmDeleteYes = false
			return m, nil
		case "enter":
			if m.confirmDeleteYes {
				if len(m.dreams) > 0 && m.selected < len(m.dreams) {
					id := m.dreams[m.selected].ID
					m.confirmDelete = false
					m.confirmDeleteYes = false
					return m, deleteDream(m.repo, id)
				}
			}
			m.confirmDelete = false
			m.confirmDeleteYes = false
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}

		return m, nil
	}

	switch msg.String() {
	case "esc", "q":
		m.state = listView
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "e":
		if len(m.dreams) > 0 && m.selected < len(m.dreams) {
			dream := m.dreams[m.selected]
			m = m.resetCreateForm()
			m.editingDream = &dream
			m.contentInput.SetValue(dream.Content)
			m.state = createView
			return m, m.contentInput.Focus()
		}
		return m, nil
	case "d":
		if len(m.dreams) > 0 && m.selected < len(m.dreams) {
			m.confirmDelete = true
			m.confirmDeleteYes = false
		}
		return m, nil
	}

	return m, nil
}

func deleteDream(r repo, id int64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := r.DeleteDream(ctx, id)
		return dreamDeletedMsg{err: err}
	}
}

func (m Model) resetCreateForm() Model {
	m.contentInput.SetValue("")
	m.contentInsertMode = true
	m.commandMode = false
	m.commandInput = ""
	m.statusMessage = ""
	m.pendingDeleteOp = false
	m.editingDream = nil
	_ = m.contentInput.Cursor.SetMode(cursor.CursorBlink)
	m.contentInput.Blur()
	return m
}
