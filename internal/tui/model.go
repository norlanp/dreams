package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"dreams/internal/model"
)

type repo interface {
	CreateDream(ctx context.Context, title, content string) (*model.Dream, error)
	ListDreams(ctx context.Context) ([]model.Dream, error)
	GetDream(ctx context.Context, id int64) (*model.Dream, error)
	UpdateDream(ctx context.Context, id int64, title, content string) (*model.Dream, error)
	DeleteDream(ctx context.Context, id int64) error
}

type viewState int

const (
	listView viewState = iota
	detailView
	createView
	updateView
)

type Model struct {
	repo              repo
	state             viewState
	width             int
	height            int
	dreams            []model.Dream
	selected          int
	titleInput        textinput.Model
	contentInput      textarea.Model
	error             error
	editingDream      *model.Dream
	focusContent      bool
	contentInsertMode bool
	commandMode       bool
	commandInput      string
	statusMessage     string
	pendingDeleteOp   bool
	confirmDelete     bool
	confirmDeleteYes  bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginLeft(2)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(lipgloss.Color("#F8F8F2"))

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1).
			Padding(0, 2)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	inputLabelFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF79C6")).
				Bold(true)

	modeInsertStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	modeNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#44475A")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true)

	confirmPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")).
				MarginTop(1)

	confirmChoiceStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")).
				Padding(0, 1)

	confirmChoiceSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#F8F8F2")).
					Background(lipgloss.Color("#FF79C6")).
					Padding(0, 1).
					Bold(true)
)

func NewModel(r repo) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter title..."
	ti.Focus()
	ti.Width = 50

	ta := textarea.New()
	ta.Placeholder = "Enter content..."
	ta.SetWidth(50)
	ta.SetHeight(10)
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	return Model{
		repo:              r,
		state:             listView,
		dreams:            []model.Dream{},
		titleInput:        ti,
		contentInput:      ta,
		contentInsertMode: true,
	}
}

func (m Model) Init() tea.Cmd {
	return loadDreams(m.repo)
}

func Run(repo repo) error {
	p := tea.NewProgram(NewModel(repo), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
