package modal

import (
	"strings"

	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// SwitchChannelMsg is emitted when the user selects a channel from the switcher.
type SwitchChannelMsg struct {
	ChannelID   string
	ChannelName string
}

// Switcher is the Ctrl+K quick channel picker modal.
type Switcher struct {
	channels []slack.Channel
	filtered []slack.Channel
	query    string
	cursor   int
	width    int
	height   int
	open     bool
}

func NewSwitcher() Switcher {
	return Switcher{}
}

func (s Switcher) Open(channels []slack.Channel) Switcher {
	s.open = true
	s.channels = channels
	s.filtered = channels
	s.query = ""
	s.cursor = 0
	return s
}

func (s Switcher) Close() Switcher {
	s.open = false
	return s
}

func (s Switcher) IsOpen() bool {
	return s.open
}

func (s Switcher) Update(msg tea.Msg) (Switcher, tea.Cmd) {
	if !s.open {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			s.open = false
		case "enter":
			if s.cursor < len(s.filtered) {
				ch := s.filtered[s.cursor]
				s.open = false
				return s, func() tea.Msg {
					return SwitchChannelMsg{
						ChannelID:   ch.ID,
						ChannelName: ch.Name,
					}
				}
			}
		case "up", "ctrl+p":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "ctrl+n":
			if s.cursor < len(s.filtered)-1 {
				s.cursor++
			}
		case "backspace":
			if len(s.query) > 0 {
				s.query = s.query[:len(s.query)-1]
				s.filter()
			}
		default:
			if len(msg.String()) == 1 {
				s.query += msg.String()
				s.filter()
			}
		}
	}

	return s, nil
}

func (s *Switcher) filter() {
	if s.query == "" {
		s.filtered = s.channels
		s.cursor = 0
		return
	}

	q := strings.ToLower(s.query)
	s.filtered = nil
	for _, ch := range s.channels {
		if strings.Contains(strings.ToLower(ch.Name), q) {
			s.filtered = append(s.filtered, ch)
		}
	}
	s.cursor = 0
}

func (s Switcher) View(screenWidth, screenHeight int) string {
	if !s.open {
		return ""
	}

	width := 50
	if width > screenWidth-4 {
		width = screenWidth - 4
	}

	var b strings.Builder

	// Header
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f6afef")).
		Bold(true).
		Render("[ QUICK_SWITCH // CTRL+K ]"))
	b.WriteString("\n\n")

	// Search input
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5edda0")).
		Bold(true).
		Render("> "))
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#dfe2eb")).
		Render(s.query))
	b.WriteString(lipgloss.NewStyle().
		Background(lipgloss.Color("#f6afef")).
		Render(" "))
	b.WriteString("\n\n")

	// Results
	maxResults := 10
	if len(s.filtered) < maxResults {
		maxResults = len(s.filtered)
	}

	for i := 0; i < maxResults; i++ {
		ch := s.filtered[i]
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		if i == s.cursor {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5edda0")).
				Bold(true).
				Background(lipgloss.Color("#181c22"))
		}
		b.WriteString(style.Render("  " + ch.Name))
		b.WriteString("\n")
	}

	if len(s.filtered) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true).
			Render("  No matches"))
	}

	// Modal frame
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#4a154b")).
		Background(lipgloss.Color("#10141a")).
		Render(b.String())
}
