package sidebar

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type ChannelSelectedMsg struct {
	ChannelID   string
	ChannelName string
}

type ChannelsLoadedMsg struct {
	Channels []slack.Channel
	Err      error
}

type Model struct {
	client   *slack.Client
	channels []slack.Channel
	cursor   int
	width    int
	height   int
	offset   int
	loading  bool
}

func New(client *slack.Client) Model {
	return Model{
		client:  client,
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		channels, err := client.ListChannels()
		return ChannelsLoadedMsg{Channels: channels, Err: err}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ChannelsLoadedMsg:
		m.loading = false
		if msg.Err == nil {
			m.channels = msg.Channels
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.channels)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "enter":
			if m.cursor < len(m.channels) {
				ch := m.channels[m.cursor]
				return m, func() tea.Msg {
					return ChannelSelectedMsg{
						ChannelID:   ch.ID,
						ChannelName: ch.Name,
					}
				}
			}
		}
	}

	return m, nil
}

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

func (m Model) IncrementUnread(channelID string) Model {
	for i := range m.channels {
		if m.channels[i].ID == channelID {
			m.channels[i].UnreadCount++
			break
		}
	}
	return m
}

func (m Model) View(width, height int, focused bool) string {
	if width == 0 || height == 0 {
		return ""
	}

	var lines []string

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f6afef")).Bold(true)
	lines = append(lines, headerStyle.Render("SLACK-TUI"))
	lines = append(lines, strings.Repeat("─", width))

	if m.loading {
		lines = append(lines, "  Loading...")
		return strings.Join(lines, "\n")
	}

	var chans, dms []int
	for i, ch := range m.channels {
		if ch.IsDM || ch.IsGroupDM {
			dms = append(dms, i)
		} else {
			chans = append(chans, i)
		}
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5edda0")).Bold(true)
	unreadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#dfe2eb")).Bold(true)

	if len(chans) > 0 {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("CHANNELS"))
		for _, idx := range chans {
			lines = append(lines, m.renderChannel(idx, width, focused, activeStyle, unreadStyle, dimStyle))
		}
	}

	if len(dms) > 0 {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("DIRECT MESSAGES"))
		for _, idx := range dms {
			lines = append(lines, m.renderChannel(idx, width, focused, activeStyle, unreadStyle, dimStyle))
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderChannel(idx, width int, focused bool, active, unread, dim lipgloss.Style) string {
	ch := m.channels[idx]
	selected := idx == m.cursor

	name := ch.Name
	maxName := width - 6
	if maxName < 5 {
		maxName = 5
	}
	if len(name) > maxName {
		name = name[:maxName-3] + "..."
	}

	badge := ""
	if ch.UnreadCount > 0 {
		badge = fmt.Sprintf(" (%d)", ch.UnreadCount)
	}

	line := "  " + name + badge

	if selected && focused {
		return "▸ " + active.Render(name+badge)
	} else if ch.UnreadCount > 0 {
		return unread.Render(line)
	}
	return dim.Render(line)
}

func (m *Model) ensureVisible() {
	if m.height == 0 {
		return
	}
	visible := m.height - 4
	if visible < 1 {
		visible = 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}
