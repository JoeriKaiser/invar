package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/invar/internal/date"
	"github.com/user/invar/internal/storage"
	"github.com/user/invar/internal/task"
	"github.com/user/invar/internal/ui"
)

type viewState int

const (
	viewList viewState = iota
	viewInput
	viewDeadline
	viewArchive
	viewPriority
	viewDeadlineMenu
)

type inputMode int

const (
	modeNew inputMode = iota
	modeEdit
)

type menuItem struct {
	label string
	style lipgloss.Style
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	New      key.Binding
	Edit     key.Binding
	Complete key.Binding
	Archive  key.Binding
	Delete   key.Binding
	Priority key.Binding
	Deadline key.Binding
	Switch   key.Binding
	Quit     key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		New:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Complete: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "complete")),
		Archive:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
		Delete:   key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete")),
		Priority: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "priority")),
		Deadline: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "deadline")),
		Switch:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch view")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

type Model struct {
	keys      keyMap
	store     *storage.Store
	view      viewState
	inputMode inputMode
	textarea  textarea.Model
	textinput textinput.Model
	tasks     []*task.Task
	editTask  *task.Task
	cursor     int
	scroll     int
	menuCursor int
	width      int
	height    int
}

func New() (*Model, error) {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".local", "share", "invar", "tasks")

	store, err := storage.New(dataDir)
	if err != nil {
		return nil, err
	}

	ta := textarea.New()
	ta.SetHeight(5)
	ta.FocusedStyle.Base = ta.FocusedStyle.Base.
		BorderForeground(ui.ColorPrimary)
	ta.BlurredStyle.Base = ta.BlurredStyle.Base.
		BorderForeground(ui.ColorBorder)

	ti := textinput.New()
	ti.Placeholder = "today, tomorrow, next week, or YYYY-MM-DD"
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ui.ColorFg)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.ColorMuted)

	m := &Model{
		keys:      defaultKeyMap(),
		store:     store,
		view:      viewList,
		inputMode: modeNew,
		textarea:  ta,
		textinput: ti,
	}

	m.loadTasks()
	return m, nil
}

func (m *Model) loadTasks() {
	tasks, _ := m.store.List(m.view == viewArchive)

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CompletedAt != nil && tasks[j].CompletedAt == nil {
			return false
		}
		if tasks[i].CompletedAt == nil && tasks[j].CompletedAt != nil {
			return true
		}

		priorityOrder := map[task.Priority]int{
			task.PriorityHigh:   0,
			task.PriorityMedium: 1,
			task.PriorityLow:    2,
		}
		if priorityOrder[tasks[i].Priority] != priorityOrder[tasks[j].Priority] {
			return priorityOrder[tasks[i].Priority] < priorityOrder[tasks[j].Priority]
		}

		if tasks[i].Deadline != nil && tasks[j].Deadline != nil {
			return tasks[i].Deadline.Before(*tasks[j].Deadline)
		}
		if tasks[i].Deadline != nil {
			return true
		}
		if tasks[j].Deadline != nil {
			return false
		}

		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	m.tasks = tasks

	// Clamp cursor.
	if m.cursor >= len(m.tasks) {
		m.cursor = len(m.tasks) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) selectedTask() *task.Task {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}
	return m.tasks[m.cursor]
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(56)

	case tea.KeyMsg:
		switch m.view {
		case viewInput:
			return m.handleInputKey(msg)
		case viewDeadline:
			return m.handleDeadlineKey(msg)
		case viewPriority:
			return m.handlePriorityKey(msg)
		case viewDeadlineMenu:
			return m.handleDeadlineMenuKey(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scroll {
					m.scroll = m.cursor
				}
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
				vis := m.visibleRowCount()
				if m.cursor >= m.scroll+vis {
					m.scroll = m.cursor - vis + 1
				}
			}
		case key.Matches(msg, m.keys.New):
			m.view = viewInput
			m.inputMode = modeNew
			m.editTask = nil
			m.textarea.SetValue("")
			m.textarea.Focus()
			return m, textarea.Blink
		case key.Matches(msg, m.keys.Switch):
			if m.view == viewList {
				m.view = viewArchive
			} else {
				m.view = viewList
			}
			m.cursor = 0
			m.scroll = 0
			m.loadTasks()
			return m, nil
		case key.Matches(msg, m.keys.Complete):
			if t := m.selectedTask(); t != nil {
				if t.CompletedAt != nil {
					t.Uncomplete()
				} else {
					t.Complete()
				}
				m.store.Save(t)
				m.loadTasks()
			}
		case key.Matches(msg, m.keys.Archive):
			if t := m.selectedTask(); t != nil {
				if m.view == viewArchive {
					t.Unarchive()
				} else {
					t.Archive()
				}
				m.store.Save(t)
				m.loadTasks()
			}
		case key.Matches(msg, m.keys.Delete):
			if t := m.selectedTask(); t != nil {
				m.store.Delete(t.ID)
				m.loadTasks()
			}
		case key.Matches(msg, m.keys.Priority):
			if t := m.selectedTask(); t != nil {
				m.view = viewPriority
				m.editTask = t
				switch t.Priority {
				case task.PriorityHigh:
					m.menuCursor = 0
				case task.PriorityMedium:
					m.menuCursor = 1
				case task.PriorityLow:
					m.menuCursor = 2
				}
			}
		case key.Matches(msg, m.keys.Edit):
			if t := m.selectedTask(); t != nil {
				m.view = viewInput
				m.inputMode = modeEdit
				m.editTask = t
				m.textarea.SetValue(t.Content)
				m.textarea.Focus()
				return m, textarea.Blink
			}
		case key.Matches(msg, m.keys.Deadline):
			if t := m.selectedTask(); t != nil {
				m.view = viewDeadlineMenu
				m.editTask = t
				m.menuCursor = 0
			}
		}
	}

	return m, nil
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewList
		m.editTask = nil
		return m, nil
	case "enter":
		content := m.textarea.Value()
		if content != "" {
			if m.inputMode == modeEdit && m.editTask != nil {
				m.editTask.Content = content
				m.store.Save(m.editTask)
			} else {
				t := task.New(content)
				m.store.Save(t)
			}
			m.loadTasks()
		}
		m.view = viewList
		m.editTask = nil
		m.textarea.SetValue("")
		return m, nil
	case "shift+enter":
		m.textarea, _ = m.textarea.Update(tea.KeyMsg{Type: tea.KeyEnter})
		return m, nil
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) handleDeadlineKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewList
		m.editTask = nil
		return m, nil
	case "enter":
		if m.editTask != nil {
			input := m.textinput.Value()
			deadline, _ := date.Parse(input)
			m.editTask.SetDeadline(deadline)
			m.store.Save(m.editTask)
			m.loadTasks()
		}
		m.view = viewList
		m.editTask = nil
		m.textinput.SetValue("")
		return m, nil
	}
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m Model) handlePriorityKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewList
		m.editTask = nil
		return m, nil
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < 2 {
			m.menuCursor++
		}
	case "enter":
		if m.editTask != nil {
			priorities := []task.Priority{task.PriorityHigh, task.PriorityMedium, task.PriorityLow}
			m.editTask.SetPriority(priorities[m.menuCursor])
			m.store.Save(m.editTask)
			m.loadTasks()
		}
		m.view = viewList
		m.editTask = nil
		return m, nil
	}
	return m, nil
}

func (m Model) deadlineMenuItemCount() int {
	if m.editTask != nil && m.editTask.Deadline != nil {
		return 5
	}
	return 4
}

func (m Model) handleDeadlineMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := m.deadlineMenuItemCount() - 1
	switch msg.String() {
	case "esc":
		m.view = viewList
		m.editTask = nil
		return m, nil
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < maxIdx {
			m.menuCursor++
		}
	case "enter":
		if m.editTask == nil {
			m.view = viewList
			return m, nil
		}
		switch m.menuCursor {
		case 0: // Today
			d, _ := date.Parse("today")
			m.editTask.SetDeadline(d)
			m.store.Save(m.editTask)
			m.loadTasks()
			m.view = viewList
			m.editTask = nil
		case 1: // Tomorrow
			d, _ := date.Parse("tomorrow")
			m.editTask.SetDeadline(d)
			m.store.Save(m.editTask)
			m.loadTasks()
			m.view = viewList
			m.editTask = nil
		case 2: // Next week
			d, _ := date.Parse("next week")
			m.editTask.SetDeadline(d)
			m.store.Save(m.editTask)
			m.loadTasks()
			m.view = viewList
			m.editTask = nil
		case 3: // Custom
			m.view = viewDeadline
			m.textinput.SetValue("")
			m.textinput.Focus()
			return m, textinput.Blink
		case 4: // Clear deadline
			m.editTask.SetDeadline(nil)
			m.store.Save(m.editTask)
			m.loadTasks()
			m.view = viewList
			m.editTask = nil
		}
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.view {
	case viewInput:
		return m.viewOverlay("input")
	case viewDeadline:
		return m.viewOverlay("deadline")
	case viewPriority:
		return m.viewMenuOverlay("Priority", []menuItem{
			{label: " HIGH ", style: ui.PriorityPillHigh},
			{label: " MED ", style: ui.PriorityPillMed},
			{label: " LOW ", style: ui.PriorityPillLow},
		})
	case viewDeadlineMenu:
		return m.viewDeadlineMenuOverlay()
	}
	return m.viewDashboard()
}

// viewDashboard renders the main task list dashboard.
func (m Model) viewDashboard() string {
	w := max(m.width, 40)

	// Inner width accounts for the outer border (2 chars).
	inner := w - 2

	// Header.
	appName := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true).
		Render("◆ invar")

	activeLabel := "Active"
	archiveLabel := "Archive"
	if m.view == viewArchive {
		activeLabel = ui.TabInactive.Render(activeLabel)
		archiveLabel = ui.TabActive.Render(archiveLabel)
	} else {
		activeLabel = ui.TabActive.Render(activeLabel)
		archiveLabel = ui.TabInactive.Render(archiveLabel)
	}
	tabs := activeLabel + " " + archiveLabel

	headerLeft := lipgloss.NewStyle().Padding(0, 2).Render(appName)
	headerRight := lipgloss.NewStyle().Padding(0, 2).Render(tabs)

	leftW := lipgloss.Width(headerLeft)
	rightW := lipgloss.Width(headerRight)
	gap := max(inner-leftW-rightW, 1)

	header := lipgloss.NewStyle().
		Width(inner).
		Render(headerLeft + strings.Repeat(" ", gap) + headerRight)

	// Task rows.
	vis := m.visibleRowCount()
	var rows []string
	end := min(m.scroll+vis, len(m.tasks))

	for i := m.scroll; i < end; i++ {
		rows = append(rows, m.renderTaskRow(m.tasks[i], i == m.cursor, inner))
	}

	// Pad empty space if fewer tasks than visible slots.
	for len(rows) < vis {
		rows = append(rows, strings.Repeat("\n", 3))
	}

	taskArea := strings.Join(rows, "\n")

	// Footer.
	total, pending, overdue := m.taskCounts()
	statsText := fmt.Sprintf("%d tasks · %d pending · %d overdue", total, pending, overdue)
	stats := ui.FooterStats.Width(inner).Render(statsText)

	helpText := "n new  e edit  space complete  p priority  d deadline  a archive  D delete  tab switch  q quit"
	helpLine := ui.FooterHelp.Width(inner).Render(helpText)

	// Assemble the card.
	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		taskArea,
		stats,
		helpLine,
	)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBorder).
		Width(inner).
		Render(body)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		card,
	)
}

// viewOverlay renders a centered input overlay on top of the dashboard.
func (m Model) viewOverlay(mode string) string {
	var title, hint, content string

	if mode == "input" {
		if m.inputMode == modeEdit {
			title = "Edit Task"
		} else {
			title = "New Task"
		}
		hint = "Enter to save · Shift+Enter for new line · Esc to cancel"
		content = m.textarea.View()
	} else {
		title = "Set Deadline"
		hint = "Enter to save · Esc to cancel"
		content = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
			"Examples: today, tomorrow, next week, 2026-02-01",
		) + "\n\n" + m.textinput.View()
	}

	titleRendered := ui.OverlayTitle.Render(title)
	hintRendered := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(hint)

	card := ui.OverlayCard.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleRendered,
			"",
			content,
			"",
			hintRendered,
		),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		card,
	)
}

func (m Model) viewMenuOverlay(title string, items []menuItem) string {
	titleRendered := ui.OverlayTitle.Render(title)
	hintRendered := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("↑/↓ navigate · Enter select · Esc cancel")

	var rows []string
	for i, item := range items {
		rendered := item.style.Render(item.label)
		if i == m.menuCursor {
			rendered = "▸ " + rendered
		} else {
			rendered = "  " + rendered
		}
		rows = append(rows, rendered)
	}

	content := strings.Join(rows, "\n")

	card := ui.OverlayCard.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleRendered,
			"",
			content,
			"",
			hintRendered,
		),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		card,
	)
}

func (m Model) viewDeadlineMenuOverlay() string {
	titleRendered := ui.OverlayTitle.Render("Deadline")
	hintRendered := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("↑/↓ navigate · Enter select · Esc cancel")

	options := []string{"Today", "Tomorrow", "Next week", "Custom..."}
	if m.editTask != nil && m.editTask.Deadline != nil {
		options = append(options, "Clear deadline")
	}

	optStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	var rows []string
	for i, opt := range options {
		if i == m.menuCursor {
			rows = append(rows, lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▸ "+opt))
		} else {
			rows = append(rows, "  "+optStyle.Render(opt))
		}
	}

	content := strings.Join(rows, "\n")

	card := ui.OverlayCard.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleRendered,
			"",
			content,
			"",
			hintRendered,
		),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		card,
	)
}

// visibleRowCount returns how many task rows fit in the viewport.
// Each task is 3 lines. Chrome = header(1) + stats(1) + help(1) + border(2) = 5.
func (m Model) visibleRowCount() int {
	available := max(m.height-2-5, 4)
	return available / 4
}

// taskCounts returns total, pending, and overdue task counts.
func (m Model) taskCounts() (total, pending, overdue int) {
	total = len(m.tasks)
	for _, t := range m.tasks {
		if t.CompletedAt == nil {
			pending++
		}
		if t.IsOverdue() {
			overdue++
		}
	}
	return
}

// renderTaskRow renders a single task as a card with a rounded border.
func (m Model) renderTaskRow(t *task.Task, selected bool, width int) string {
	cardStyle := ui.TaskCardNormal.Width(width - 2)
	if selected {
		cardStyle = ui.TaskCardSelected.Width(width - 2)
	}

	// Line 1: bullet + content + deadline
	var bullet string
	if t.CompletedAt != nil {
		bullet = lipgloss.NewStyle().Foreground(ui.ColorLow).Render("✓")
	} else if selected {
		bullet = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Render("●")
	} else {
		bullet = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("○")
	}

	content := t.Content
	lines := splitLines(content)
	if len(lines) > 0 {
		content = lines[0]
	}
	if t.CompletedAt != nil {
		content = lipgloss.NewStyle().Foreground(ui.ColorMuted).Strikethrough(true).Render(content)
	}

	deadline := ""
	if t.Deadline != nil {
		dl := t.Deadline.Format("Jan 02")
		if t.IsOverdue() {
			deadline = ui.DeadlineOverdue.Render("! " + dl)
		} else {
			deadline = ui.DeadlineNormal.Render(dl)
		}
	}

	leftPart := bullet + " " + content
	leftW := lipgloss.Width(leftPart)
	rightW := lipgloss.Width(deadline)
	innerW := width - 2 - 2
	gap := max(innerW-leftW-rightW, 1)
	line1 := leftPart + strings.Repeat(" ", gap) + deadline

	// Line 2: priority pill + overdue
	pill := ui.PriorityPill(string(t.Priority))
	var line2Extra string
	if t.IsOverdue() {
		line2Extra = "  " + ui.DeadlineOverdue.Render("overdue")
	}
	line2 := pill + line2Extra

	return cardStyle.Render(line1 + "\n" + line2)
}

func splitLines(s string) []string {
	var lines []string
	var current string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" || len(lines) > 0 {
		lines = append(lines, current)
	}
	return lines
}
