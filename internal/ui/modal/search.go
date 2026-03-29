package modal

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// SearchResultMsg is returned after a workspace search.
type SearchResultMsg struct {
	Query    string
	Messages []slack.Message
	Err      error
}

// JumpToMessageMsg is emitted when the user picks a search result.
type JumpToMessageMsg struct {
	ChannelID string
	Timestamp string
}

// Search is the Ctrl+F search modal.
type Search struct {
	client   *slack.Client
	query    string
	results  []slack.Message
	cursor   int
	open     bool
	loading  bool
}

func NewSearch(client *slack.Client) Search {
	return Search{client: client}
}

func (s Search) Open() Search {
	s.open = true
	s.query = ""
	s.results = nil
	s.cursor = 0
	s.loading = false
	return s
}

func (s Search) Close() Search {
	s.open = false
	return s
}

func (s Search) IsOpen() bool {
	return s.open
}

func (s Search) Update(msg tea.Msg) (Search, tea.Cmd) {
	if !s.open {
		return s, nil
	}

	switch msg := msg.(type) {
	case SearchResultMsg:
		s.loading = false
		if msg.Err == nil {
			s.results = msg.Messages
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			s.open = false
		case "enter":
			if s.loading {
				return s, nil
			}
			if len(s.results) > 0 && s.cursor < len(s.results) {
				result := s.results[s.cursor]
				s.open = false
				return s, func() tea.Msg {
					return JumpToMessageMsg{
						ChannelID: result.ChannelID,
						Timestamp: result.Timestamp.String(),
					}
				}
			}
			// If no results yet, execute search
			if s.query != "" {
				s.loading = true
				client := s.client
				query := s.query
				return s, func() tea.Msg {
					msgs, err := client.SearchMessages(query)
					return SearchResultMsg{Query: query, Messages: msgs, Err: err}
				}
			}
		case "up", "ctrl+p":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "ctrl+n":
			if s.cursor < len(s.results)-1 {
				s.cursor++
			}
		case "backspace":
			if len(s.query) > 0 {
				s.query = s.query[:len(s.query)-1]
			}
		default:
			if len(msg.String()) == 1 {
				s.query += msg.String()
			}
		}
	}

	return s, nil
}

func (s Search) View(screenWidth, screenHeight int) string {
	if !s.open {
		return ""
	}

	width := 70
	if width > screenWidth-4 {
		width = screenWidth - 4
	}

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f6afef")).
		Bold(true).
		Render("[ SEARCH // CTRL+F ]"))
	b.WriteString("\n\n")

	// Search input
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5edda0")).
		Bold(true).
		Render(">>> "))
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#dfe2eb")).
		Render(s.query))
	b.WriteString(lipgloss.NewStyle().
		Background(lipgloss.Color("#f6afef")).
		Render(" "))
	b.WriteString("\n")

	if s.loading {
		b.WriteString("\n  Searching...")
	} else if len(s.results) > 0 {
		b.WriteString(fmt.Sprintf("\n  %d results:\n\n", len(s.results)))

		maxResults := 8
		if len(s.results) < maxResults {
			maxResults = len(s.results)
		}

		for i := 0; i < maxResults; i++ {
			msg := s.results[i]
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
			if i == s.cursor {
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#dfe2eb")).
					Background(lipgloss.Color("#181c22"))
			}

			text := msg.Text
			if len(text) > width-10 {
				text = text[:width-13] + "..."
			}

			line := fmt.Sprintf("  <%s> %s", msg.Username, text)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	} else if s.query != "" {
		b.WriteString("\n  Press ENTER to search")
	}

	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#4a154b")).
		Background(lipgloss.Color("#10141a")).
		Render(b.String())
}
