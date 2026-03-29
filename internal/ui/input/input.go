package input

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// SendMessageMsg is emitted when the user submits a message.
type SendMessageMsg struct {
	Text string
}

// SendErrorMsg is returned if sending fails.
type SendErrorMsg struct {
	Err error
}

type Model struct {
	text   string
	cursor int
	width  int
	height int
}

func New() Model {
	return Model{}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if strings.TrimSpace(m.text) != "" {
				text := m.text
				m.text = ""
				m.cursor = 0
				return m, func() tea.Msg {
					return SendMessageMsg{Text: text}
				}
			}
		case "backspace":
			if m.cursor > 0 {
				m.text = m.text[:m.cursor-1] + m.text[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.text) {
				m.cursor++
			}
		case "ctrl+a":
			m.cursor = 0
		case "ctrl+e":
			m.cursor = len(m.text)
		case "ctrl+u":
			m.text = m.text[m.cursor:]
			m.cursor = 0
		case "ctrl+k":
			m.text = m.text[:m.cursor]
		default:
			if msg.String() != "" && len(msg.String()) == 1 {
				ch := msg.String()
				m.text = m.text[:m.cursor] + ch + m.text[m.cursor:]
				m.cursor++
			} else if r := msg.Key().Code; r > 31 && r < 127 {
				ch := string(rune(r))
				m.text = m.text[:m.cursor] + ch + m.text[m.cursor:]
				m.cursor++
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

func (m Model) View(width, height int, focused bool, channelName string) string {
	if width == 0 {
		return ""
	}

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5edda0")).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#dfe2eb"))

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#f6afef")).
		Foreground(lipgloss.Color("#10141a"))

	prompt := promptStyle.Render("[" + channelName + "] >")

	var displayText string
	if focused {
		// Show cursor
		before := m.text[:m.cursor]
		after := m.text[m.cursor:]
		cursor := cursorStyle.Render(" ")
		if m.cursor < len(m.text) {
			cursor = cursorStyle.Render(string(m.text[m.cursor]))
			after = m.text[m.cursor+1:]
		}
		displayText = textStyle.Render(before) + cursor + textStyle.Render(after)
	} else {
		if m.text == "" {
			displayText = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render("Type a message or /command...")
		} else {
			displayText = textStyle.Render(m.text)
		}
	}

	line := prompt + " " + displayText

	borderColor := lipgloss.Color("#4f434c")
	if focused {
		borderColor = lipgloss.Color("#f6afef")
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(line)
}
