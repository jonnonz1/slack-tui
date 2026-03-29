package statusbar

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	connected   bool
	channelName string
	username    string
}

func New() Model {
	return Model{
		username: "root_user",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) SetConnected(connected bool) Model {
	m.connected = connected
	return m
}

func (m Model) SetChannel(name string) Model {
	m.channelName = name
	return m
}

func (m Model) SetUsername(name string) Model {
	m.username = name
	return m
}

func (m Model) View(width int) string {
	if width == 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("#181c22")).
		Foreground(lipgloss.Color("#dfe2eb"))

	var status string
	if m.connected {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5edda0")).
			Render("● CONNECTED")
	} else {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffb4ab")).
			Render("● DISCONNECTED")
	}

	user := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f6afef")).
		Render("@" + m.username)

	channel := ""
	if m.channelName != "" {
		channel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5edda0")).
			Render(m.channelName)
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("? help")

	left := fmt.Sprintf(" %s | %s | %s", status, user, channel)
	right := help + " "

	gap := width - len(left) - len(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + fmt.Sprintf("%*s", gap, "") + right

	return style.Render(bar)
}
