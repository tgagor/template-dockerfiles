package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EventMsg struct {
	ImageName string
	Status    string
	IsDone    bool
	Err       error
}

type Model struct {
	TotalImages     int
	CompletedImages int

	activeTasks map[string]string // ImageName -> Status
	orderedKeys []string          // to keep rendering deterministic

	spinner  spinner.Model
	progress progress.Model

	err error
}

var (
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true)
	taskStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	doneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)

func NewModel(total int) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		TotalImages: total,
		activeTasks: make(map[string]string),
		orderedKeys: make([]string, 0),
		spinner:     s,
		progress:    progress.New(progress.WithDefaultGradient()),
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case EventMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, tea.Quit
		}

		if msg.IsDone {
			m.CompletedImages++
			delete(m.activeTasks, msg.ImageName)
			// Remove from ordered keys
			for i, v := range m.orderedKeys {
				if v == msg.ImageName {
					m.orderedKeys = append(m.orderedKeys[:i], m.orderedKeys[i+1:]...)
					break
				}
			}

			if m.CompletedImages >= m.TotalImages {
				// Wait a moment so the user sees 100% completion before quitting
				return m, tea.Sequence(
					tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
						return tea.Quit()
					}),
					m.spinner.Tick,
				)
			}
			return m, nil
		}

		// Update or add task
		if _, exists := m.activeTasks[msg.ImageName]; !exists {
			m.orderedKeys = append(m.orderedKeys, msg.ImageName)
		}
		m.activeTasks[msg.ImageName] = msg.Status
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n❌ Build failed: %v", m.err)) + "\n"
	}

	if m.CompletedImages >= m.TotalImages && m.TotalImages > 0 {
		return doneStyle.Render(fmt.Sprintf("\n✅ All %d images completed successfully!\n", m.TotalImages))
	}

	var b strings.Builder

	header := fmt.Sprintf("🚀 Executing Plan: %d / %d Images Completed", m.CompletedImages, m.TotalImages)
	b.WriteString(headerStyle.Render(header) + "\n")

	// Progress bar
	percent := float64(m.CompletedImages) / float64(m.TotalImages)
	if m.TotalImages == 0 {
		percent = 0
	}
	b.WriteString(m.progress.ViewAs(percent) + "\n\n")

	// Active Tasks
	if len(m.orderedKeys) > 0 {
		b.WriteString("Active Tasks:\n")
		// Sort keys to prevent jumping
		sort.Strings(m.orderedKeys)
		for _, imgName := range m.orderedKeys {
			status := m.activeTasks[imgName]
			line := fmt.Sprintf(" %s [%s] %s", m.spinner.View(), taskStyle.Render(imgName), status)
			b.WriteString(line + "\n")
		}
	} else {
		b.WriteString(m.spinner.View() + " Waiting for tasks to begin...\n")
	}

	return b.String()
}
