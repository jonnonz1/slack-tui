package sidebar

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// ChannelSelectedMsg is sent when the user picks a channel.
type ChannelSelectedMsg struct {
	ChannelID   string
	ChannelName string
}

// ChannelsLoadedMsg is the result of loading channels.
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
	offset   int // scroll offset
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

	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Foreground(lipgloss.Color("#f6afef")).
		Render("MONOSPACE_CMD")
	b.WriteString(header)
	b.WriteString("\n")

	// Divider
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	if m.loading {
		b.WriteString("  Loading channels...")
		return m.containerStyle(width, height, focused).Render(b.String())
	}

	// Group channels
	var publicChans, privateChans, dms []indexedChannel
	for i, ch := range m.channels {
		ic := indexedChannel{index: i, channel: ch}
		switch {
		case ch.IsDM || ch.IsGroupDM:
			dms = append(dms, ic)
		case ch.IsPrivate:
			privateChans = append(privateChans, ic)
		default:
			publicChans = append(publicChans, ic)
		}
	}

	usedLines := 2 // header + divider

	usedLines += m.renderGroup(&b, "CHANNELS", publicChans, width, focused)
	usedLines += m.renderGroup(&b, "PRIVATE", privateChans, width, focused)
	usedLines += m.renderGroup(&b, "DIRECT_MESSAGES", dms, width, focused)

	// Pad remaining space
	for i := usedLines; i < height; i++ {
		b.WriteString("\n")
	}

	return m.containerStyle(width, height, focused).Render(b.String())
}

type indexedChannel struct {
	index   int
	channel slack.Channel
}

func (m Model) renderGroup(b *strings.Builder, label string, items []indexedChannel, width int, focused bool) int {
	if len(items) == 0 {
		return 0
	}

	lines := 0

	// Section label
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Bold(true)
	b.WriteString("\n")
	b.WriteString(labelStyle.Render(label))
	b.WriteString("\n")
	lines += 2

	for _, ic := range items {
		ch := ic.channel
		selected := ic.index == m.cursor

		name := ch.Name
		if len(name) > width-4 {
			name = name[:width-7] + "..."
		}

		unread := ""
		if ch.UnreadCount > 0 {
			unread = fmt.Sprintf(" (%d)", ch.UnreadCount)
		}

		line := fmt.Sprintf("  %s%s", name, unread)

		style := lipgloss.NewStyle().Width(width)

		if selected && focused {
			style = style.
				Foreground(lipgloss.Color("#5edda0")).
				Bold(true).
				Background(lipgloss.Color("#181c22")).
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#5edda0"))
		} else if ch.UnreadCount > 0 {
			style = style.
				Foreground(lipgloss.Color("#dfe2eb")).
				Bold(true)
		} else {
			style = style.
				Foreground(lipgloss.Color("#666666"))
		}

		b.WriteString(style.Render(line))
		b.WriteString("\n")
		lines++
	}

	return lines
}

func (m Model) containerStyle(width, height int, focused bool) lipgloss.Style {
	borderColor := lipgloss.Color("#4f434c")
	if focused {
		borderColor = lipgloss.Color("#f6afef")
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(lipgloss.Color("#10141a")).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)
}

func (m *Model) ensureVisible() {
	if m.height == 0 {
		return
	}
	visible := m.height - 4 // account for headers
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}
