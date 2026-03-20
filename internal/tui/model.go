package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"dreams/internal/model"
	"dreams/internal/priming"
)

type repo interface {
	CreateDream(ctx context.Context, content string) (*model.Dream, error)
	ListDreams(ctx context.Context) ([]model.Dream, error)
	GetDream(ctx context.Context, id int64) (*model.Dream, error)
	UpdateDream(ctx context.Context, id int64, content string) (*model.Dream, error)
	DeleteDream(ctx context.Context, id int64) error
	SearchDreams(ctx context.Context, query string) ([]model.Dream, error)
	GetLatestAnalysis(ctx context.Context) (*model.Analysis, error)
	GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error)
	SaveAnalysis(ctx context.Context, analysisDate time.Time, dreamCount, nClusters int64, resultsJSON string) (*model.Analysis, error)
	SaveCluster(ctx context.Context, analysisID, clusterID, dreamCount int64, topTerms, dreamIDs string) (*model.Cluster, error)
	SaveAnalysisWithClusters(ctx context.Context, analysisDate time.Time, dreamCount, nClusters int64, resultsJSON string, clusters []model.Cluster) (*model.Analysis, error)
	GetFreshPrimingCache(ctx context.Context, source string, now time.Time, ttl time.Duration) (*model.PrimingCache, error)
	SavePrimingCache(ctx context.Context, source string, payload []string, fetchedAt time.Time) error
	SavePrimingLog(ctx context.Context, source, outcome, detail, content string, createdAt time.Time) error
	ListPrimingContent(ctx context.Context) ([]model.PrimingContent, error)
	GetPrimingContentByCategory(ctx context.Context, category string) ([]model.PrimingContent, error)
}

type viewState int

const (
	listView viewState = iota
	detailView
	createView
	updateView
	searchView
	analysisView
	exportView
	nightView
)

type primingGenerator interface {
	Next(ctx context.Context) priming.Result
}

type Model struct {
	repo               repo
	analysisRunner     analysisRunner
	analysisMinDreams  int
	state              viewState
	width              int
	height             int
	dreams             []model.Dream
	selected           int
	contentInput       textarea.Model
	error              error
	editingDream       *model.Dream
	contentInsertMode  bool
	commandMode        bool
	commandInput       string
	statusMessage      string
	pendingDeleteOp    bool
	confirmDelete      bool
	confirmDeleteYes   bool
	searchQuery        string
	isSearching        bool
	hasSearched        bool
	dreamsBeforeSearch []model.Dream
	analysis           *model.Analysis
	analysisClusters   []model.Cluster
	analysisLoading    bool
	analysisLoadErr    error
	exportDirectory    string
	exportConfirming   bool
	exportLoading      bool
	exportComplete     bool
	exportResultCount  int
	exportErr          error
	nightGenerator     primingGenerator
	nightLoading       bool
	nightContent       string
	nightSourceLabel   string
	nightStatus        string
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
	ta := textarea.New()
	ta.Placeholder = "Enter content..."
	ta.SetWidth(50)
	ta.SetHeight(10)
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	return Model{
		repo:              r,
		analysisRunner:    defaultAnalysisRunner,
		analysisMinDreams: 5,
		state:             listView,
		dreams:            []model.Dream{},
		contentInput:      ta,
		contentInsertMode: true,
		nightGenerator:    priming.NewDefaultGenerator(r),
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
