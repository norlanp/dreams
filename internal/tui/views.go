package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.error != nil {
		return m.errorView()
	}

	switch m.state {
	case listView:
		return m.listView()
	case createView:
		return m.createView()
	case detailView:
		return m.detailView()
	case searchView:
		return m.searchView()
	case analysisView:
		return m.analysisView()
	case exportView:
		return m.exportView()
	case nightView:
		return m.nightPrimingView()
	default:
		return "Unknown state"
	}
}

func (m Model) listView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Dreams"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("A dream journal"))
	b.WriteString("\n\n")

	if len(m.dreams) == 0 {
		b.WriteString(itemStyle.MarginLeft(2).Render("No dreams recorded yet. Press 'n' to add one."))
	} else {
		for i, dream := range m.dreams {
			style := itemStyle
			if i == m.selected {
				style = selectedStyle
			}
			dateStr := dream.CreatedAt.Local().Format("Mon 02, 2006")
			preview := previewText(dream.Content, 30)
			line := fmt.Sprintf("  %s - %s", dateStr, preview)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderHelp("n: new dream • p: night priming • s: statistics • e: export • /: search • enter: view • ↑↓: navigate • q: quit", m.width))

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) nightPrimingView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Night Priming"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Prepare for recall and lucidity"))
	b.WriteString("\n\n")

	if m.nightLoading {
		b.WriteString(itemStyle.MarginLeft(2).Render("Loading priming content..."))
		b.WriteString("\n\n")
	}

	if m.nightSourceLabel != "" {
		sourceLine := fmt.Sprintf("Source: %s", m.nightSourceLabel)
		b.WriteString(itemStyle.MarginLeft(2).Render(sourceLine))
		b.WriteString("\n\n")
	}

	if m.nightContent != "" {
		b.WriteString(itemStyle.MarginLeft(2).Render(wrapText(m.nightContent, m.width-6)))
		b.WriteString("\n\n")
	}

	if m.nightStatus != "" {
		b.WriteString(statusStyle.MarginLeft(2).Render(m.nightStatus))
		b.WriteString("\n\n")
	}

	if m.nightContent == "" && !m.nightLoading {
		b.WriteString(itemStyle.MarginLeft(2).Render("Press n to load priming content."))
		b.WriteString("\n\n")
	}

	b.WriteString(renderHelp("n: next priming • esc: back to list • ctrl+c: quit", m.width))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) createView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("New Dream"))
	b.WriteString("\n\n")

	modeView := modeInsertStyle.Render("INSERT")
	if !m.contentInsertMode {
		modeView = modeNormalStyle.Render("NORMAL")
	}
	b.WriteString("  ")
	b.WriteString(modeView)
	b.WriteString("\n")
	b.WriteString(m.contentInput.View())
	b.WriteString("\n\n")

	if m.commandMode {
		b.WriteString(commandStyle.Render(":" + m.commandInput))
		b.WriteString("\n")
	} else if m.statusMessage != "" {
		b.WriteString(statusStyle.Render(m.statusMessage))
		b.WriteString("\n")
	}

	b.WriteString(renderHelp("esc: normal • o/O open line • dd delete line • :e edit in $EDITOR • :w save • :wq save+exit • :q quit", m.width))

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) detailView() string {
	var b strings.Builder

	if m.selected >= len(m.dreams) {
		return "Dream not found"
	}

	dream := m.dreams[m.selected]

	b.WriteString(subtitleStyle.Render(dream.CreatedAt.Local().Format("2006-01-02 15:04")))
	b.WriteString("\n\n")

	content := wrapText(dream.Content, m.width-4)
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(content))
	b.WriteString("\n\n")

	if m.confirmDelete {
		yesStyle := confirmChoiceStyle
		noStyle := confirmChoiceStyle
		if m.confirmDeleteYes {
			yesStyle = confirmChoiceSelectedStyle
		} else {
			noStyle = confirmChoiceSelectedStyle
		}

		b.WriteString(confirmPromptStyle.Render("Delete this dream?"))
		b.WriteString("\n")
		b.WriteString(yesStyle.Render("Yes"))
		b.WriteString(" ")
		b.WriteString(noStyle.Render("No"))
		b.WriteString("\n")
		b.WriteString(renderHelp("enter: confirm • tab: switch • n/esc: cancel", m.width))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
	}

	b.WriteString(renderHelp("e: edit • d: delete • esc: back • q: quit", m.width))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) searchView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Search Dreams"))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("Search: "))
	b.WriteString(m.searchQuery)
	if m.isSearching {
		b.WriteString(" ")
		b.WriteString(statusStyle.Render("(searching...)"))
	}
	b.WriteString("\n\n")

	if len(m.dreams) == 0 {
		if m.searchQuery != "" && !m.isSearching {
			b.WriteString(itemStyle.MarginLeft(2).Render("No dreams found."))
		}
	} else {
		for i, dream := range m.dreams {
			style := itemStyle
			if i == m.selected {
				style = selectedStyle
			}
			dateStr := dream.CreatedAt.Local().Format("Mon 02, 2006")
			preview := previewText(dream.Content, 30)
			line := fmt.Sprintf("  %s - %s", dateStr, preview)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderHelp("enter: search • esc: cancel • backspace: clear • type to filter • ↑↓: navigate", m.width))

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) analysisView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Dream Statistics"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Analysis view"))
	b.WriteString("\n\n")

	switch {
	case m.analysisLoading:
		b.WriteString(itemStyle.MarginLeft(2).Render("Running analysis..."))
	default:
		if m.analysisLoadErr != nil {
			b.WriteString(itemStyle.MarginLeft(2).Render(analysisErrorTitle(m.analysisLoadErr)))
			b.WriteString("\n")
			b.WriteString(itemStyle.MarginLeft(2).Render(m.analysisLoadErr.Error()))
			b.WriteString("\n")
		}

		if m.analysis != nil {
			if m.analysisLoadErr != nil {
				b.WriteString(itemStyle.MarginLeft(2).Render("Showing last cached analysis:"))
				b.WriteString("\n")
			}

			b.WriteString(itemStyle.MarginLeft(2).Render("Last analyzed: " + formatAnalysisTimestamp(m.analysis.AnalysisDate)))
			b.WriteString("\n")
			b.WriteString(itemStyle.MarginLeft(2).Render(fmt.Sprintf("Dreams analyzed: %d", m.analysis.DreamCount)))
			b.WriteString("\n")
			b.WriteString(itemStyle.MarginLeft(2).Render(fmt.Sprintf("Clusters: %d", m.analysis.NClusters)))

			for _, cluster := range m.analysisClusters {
				b.WriteString("\n")
				line := fmt.Sprintf("Cluster %d (%d dreams): %s", cluster.ClusterID, cluster.DreamCount, strings.Join(cluster.TopTerms, ", "))
				b.WriteString(itemStyle.MarginLeft(2).Render(line))
			}
		} else if m.analysisLoadErr == nil {
			b.WriteString(itemStyle.MarginLeft(2).Render("No cached analysis yet."))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(renderHelp("r: rerun analysis • esc/q: back • ctrl+c: quit", m.width))

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) exportView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Export Dreams"))
	b.WriteString("\n\n")

	switch {
	case m.exportLoading:
		b.WriteString(itemStyle.MarginLeft(2).Render("Exporting dreams..."))
	case m.exportComplete:
		if m.exportErr != nil {
			b.WriteString(itemStyle.MarginLeft(2).Render(fmt.Sprintf("Error: %s", m.exportErr.Error())))
			b.WriteString("\n\n")
			b.WriteString(itemStyle.MarginLeft(2).Render(fmt.Sprintf("Partially exported %d dream(s) to %s", m.exportResultCount, m.exportDirectory)))
		} else {
			b.WriteString(itemStyle.MarginLeft(2).Render(fmt.Sprintf("Exported %d dream(s) to %s", m.exportResultCount, m.exportDirectory)))
		}
		b.WriteString("\n\n")
		b.WriteString(renderHelp("enter: return to list", m.width))
	default:
		b.WriteString(itemStyle.MarginLeft(2).Render(fmt.Sprintf("Directory: %s", m.exportDirectory)))
		b.WriteString("\n\n")
		if len(m.dreams) > 0 {
			b.WriteString(confirmPromptStyle.Render(fmt.Sprintf("Export %d dream(s) to %s?", len(m.dreams), m.exportDirectory)))
		} else {
			b.WriteString(confirmPromptStyle.Render("No dreams to export."))
		}
		b.WriteString("\n\n")
		b.WriteString(renderHelp("enter: confirm • esc: cancel • type: edit directory", m.width))
	}

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func formatAnalysisTimestamp(ts time.Time) string {
	return ts.Local().Format("2006-01-02 15:04 MST")
}

func analysisErrorTitle(err error) string {
	switch analysisErrorState(err) {
	case analysisErrorTooFewDreams:
		return "Not enough dreams to run analysis."
	case analysisErrorExecution:
		return "Analysis execution failed."
	case analysisErrorParse:
		return "Failed to parse analysis results."
	default:
		return "Analysis unavailable."
	}
}

func (m Model) errorView() string {
	return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.error)
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	wrappedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		wrappedLines = append(wrappedLines, wrapSingleLine(line, width))
	}

	return strings.Join(wrappedLines, "\n")
}

func renderHelp(text string, width int) string {
	contentWidth := width - 8
	if contentWidth <= 0 {
		return helpStyle.Render(text)
	}

	return helpStyle.Render(wrapText(text, contentWidth))
}

func wrapSingleLine(line string, width int) string {
	if len(line) <= width {
		return line
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) > width {
			result.WriteString(currentLine)
			result.WriteString("\n")
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}

	result.WriteString(currentLine)
	return result.String()
}

func previewText(text string, maxLen int) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return "(empty)"
	}
	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) == 0 {
		return "(empty)"
	}
	if len(firstLine) > maxLen {
		return firstLine[:maxLen-3] + "..."
	}
	return firstLine
}

var _ tea.Model = Model{}
