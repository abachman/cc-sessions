package ui

import (
	"fmt"
	"strings"
	"time"

	"cview/internal/claude"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
	detailStyle = listStyle.Copy()
)

type Model struct {
	search        textinput.Model
	list          viewport.Model
	details       viewport.Model
	sessions      []claude.Session
	filtered      []claude.Session
	selected      int
	width         int
	height        int
	err           error
	projectFolder string
}

func NewModel() (Model, error) {
	search := textinput.New()
	search.Placeholder = "Search session history"
	search.Prompt = "Search: "
	search.Focus()

	sessions, err := claude.DiscoverForCurrentDir()
	if err != nil {
		return Model{}, err
	}

	model := Model{
		search:   search,
		sessions: sessions,
		filtered: sessions,
	}
	model.projectFolder = currentProjectDir(sessions)
	model.list = viewport.New(0, 0)
	model.details = viewport.New(0, 0)
	model.syncList()
	model.syncDetails()
	return model, nil
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.move(-1)
			return m, nil
		case "down", "j":
			m.move(1)
			return m, nil
		}
	}

	var cmd tea.Cmd
	prev := m.search.Value()
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != prev {
		m.applyFilter()
	}

	var vpCmd tea.Cmd
	m.details, vpCmd = m.details.Update(msg)
	return m, tea.Batch(cmd, vpCmd)
}

func (m Model) View() string {
	if m.err != nil {
		return appStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	header := []string{
		titleStyle.Render("Claude Session Viewer"),
		mutedStyle.Render(fmt.Sprintf("%d sessions", len(m.filtered))),
	}
	if m.projectFolder != "" {
		header = append(header, mutedStyle.Render(m.projectFolder))
	}

	leftWidth := max(32, m.width/3)
	rightWidth := max(40, m.width-leftWidth-8)
	list := listStyle.Width(leftWidth).Height(max(8, m.height-8)).Render(m.list.View())
	detail := detailStyle.Width(rightWidth).Height(max(8, m.height-8)).Render(m.details.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, detail)
	return appStyle.Render(strings.Join([]string{
		strings.Join(header, "  "),
		m.search.View(),
		body,
		mutedStyle.Render("Controls: type to search, j/k or arrows to move, q to quit"),
	}, "\n\n"))
}

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	if query == "" {
		m.filtered = m.sessions
		m.selected = 0
		m.syncList()
		m.syncDetails()
		return
	}

	filtered := make([]claude.Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		if strings.Contains(strings.ToLower(session.Summary), query) || strings.Contains(session.SearchText, query) {
			filtered = append(filtered, session)
		}
	}
	m.filtered = filtered
	m.selected = 0
	m.syncList()
	m.syncDetails()
}

func (m *Model) move(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	m.syncList()
	m.syncDetails()
}

func (m *Model) syncList() {
	m.list.SetContent(m.renderListContent(max(1, m.list.Width)))
	m.ensureListSelectionVisible()
}

func (m *Model) syncDetails() {
	if len(m.filtered) == 0 {
		m.details.SetContent("No sessions matched the current filter.")
		m.details.GotoTop()
		return
	}

	selected := m.filtered[m.selected]
	lines := []string{
		titleStyle.Render(selected.Summary),
		fmt.Sprintf("Session: %s", selected.ID),
		fmt.Sprintf("Updated: %s", formatTime(selected.UpdatedAt)),
	}
	if !selected.StartedAt.IsZero() {
		lines = append(lines, fmt.Sprintf("Started: %s", formatTime(selected.StartedAt)))
	}
	if selected.Branch != "" {
		lines = append(lines, fmt.Sprintf("Branch: %s", selected.Branch))
	}
	if selected.CWD != "" {
		lines = append(lines, fmt.Sprintf("CWD: %s", selected.CWD))
	}
	lines = append(lines,
		fmt.Sprintf("Messages: %d total, %d user, %d assistant", selected.MessageCount, selected.UserPrompts, selected.AssistantMsgs),
		fmt.Sprintf("File: %s", selected.Path),
		"",
		titleStyle.Render("Transcript"),
	)

	for _, entry := range selected.Transcript {
		label := entry.Role
		if label == "" {
			label = entry.Type
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", strings.ToUpper(label), oneLine(entry.Content)))
	}

	m.details.SetContent(strings.Join(lines, "\n"))
	m.details.GotoTop()
}

func (m *Model) resize() {
	leftWidth := max(32, m.width/3)
	rightWidth := max(40, m.width-leftWidth-10)
	height := max(8, m.height-12)
	m.list.Width = leftWidth - 4
	m.list.Height = height - 2
	m.details.Width = rightWidth - 4
	m.details.Height = height - 2
	m.syncList()
}

func (m Model) renderListContent(width int) string {
	if len(m.filtered) == 0 {
		return mutedStyle.Width(width).Render("No sessions found.")
	}

	lines := make([]string, 0, len(m.filtered))
	for i, session := range m.filtered {
		item := fmt.Sprintf("%s\n%s", truncate(session.Summary, width), mutedStyle.Render(sessionMeta(session, width)))
		if i == m.selected {
			lines = append(lines, selectedStyle.Width(width).Render(item))
			continue
		}
		lines = append(lines, lipgloss.NewStyle().Width(width).Render(item))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) ensureListSelectionVisible() {
	if len(m.filtered) == 0 || m.list.Height <= 0 {
		m.list.GotoTop()
		return
	}

	itemHeight := 2
	top := m.selected * itemHeight
	bottom := top + itemHeight
	visibleTop := m.list.YOffset
	visibleBottom := visibleTop + m.list.Height

	if top < visibleTop {
		m.list.SetYOffset(top)
		return
	}
	if bottom > visibleBottom {
		m.list.SetYOffset(bottom - m.list.Height)
	}
}

func sessionMeta(session claude.Session, width int) string {
	bits := []string{formatTime(session.UpdatedAt)}
	if session.Branch != "" {
		bits = append(bits, session.Branch)
	}
	return truncate(strings.Join(bits, "  "), width)
}

func currentProjectDir(sessions []claude.Session) string {
	if len(sessions) == 0 {
		return ""
	}
	return sessions[0].ProjectPath
}

func formatTime(ts time.Time) string {
	if ts.IsZero() {
		return "unknown"
	}
	return ts.Local().Format("2006-01-02 15:04")
}

func truncate(value string, width int) string {
	if width <= 3 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-3]) + "..."
}

func oneLine(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return truncate(value, 120)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
