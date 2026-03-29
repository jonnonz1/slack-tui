package sidebar

import (
	"testing"

	"github.com/jonnonz1/slack-tui/internal/slack"
)

func TestSidebar_IncrementUnread(t *testing.T) {
	m := Model{
		channels: []slack.Channel{
			{ID: "C1", Name: "#general", UnreadCount: 0},
			{ID: "C2", Name: "#random", UnreadCount: 5},
		},
	}

	m = m.IncrementUnread("C1")

	if m.channels[0].UnreadCount != 1 {
		t.Errorf("expected unread count 1 for C1, got %d", m.channels[0].UnreadCount)
	}
	if m.channels[1].UnreadCount != 5 {
		t.Errorf("C2 unread should be unchanged at 5, got %d", m.channels[1].UnreadCount)
	}
}

func TestSidebar_IncrementUnread_Unknown(t *testing.T) {
	m := Model{
		channels: []slack.Channel{
			{ID: "C1", Name: "#general", UnreadCount: 0},
		},
	}

	m = m.IncrementUnread("UNKNOWN")

	if m.channels[0].UnreadCount != 0 {
		t.Errorf("unknown channel should not affect existing, got %d", m.channels[0].UnreadCount)
	}
}

func TestSidebar_SetSize(t *testing.T) {
	m := Model{}
	m = m.SetSize(30, 50)

	if m.width != 30 || m.height != 50 {
		t.Errorf("expected 30x50, got %dx%d", m.width, m.height)
	}
}

func TestSidebar_CursorBounds(t *testing.T) {
	m := Model{
		channels: []slack.Channel{
			{ID: "C1", Name: "#a"},
			{ID: "C2", Name: "#b"},
			{ID: "C3", Name: "#c"},
		},
		cursor: 0,
	}

	// Move down
	m, _ = m.Update(testKeyMsg("j"))
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.cursor)
	}

	m, _ = m.Update(testKeyMsg("j"))
	if m.cursor != 2 {
		t.Errorf("expected cursor 2, got %d", m.cursor)
	}

	// Should not go past last
	m, _ = m.Update(testKeyMsg("j"))
	if m.cursor != 2 {
		t.Errorf("cursor should stay at 2 (last), got %d", m.cursor)
	}

	// Move up
	m, _ = m.Update(testKeyMsg("k"))
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after k, got %d", m.cursor)
	}

	// Move to top
	m, _ = m.Update(testKeyMsg("k"))
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// Should not go below 0
	m, _ = m.Update(testKeyMsg("k"))
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestSidebar_ViewEmpty(t *testing.T) {
	m := Model{}
	v := m.View(0, 0, false)
	if v != "" {
		t.Errorf("expected empty view for zero dimensions, got %q", v)
	}
}

func TestSidebar_ViewLoading(t *testing.T) {
	m := Model{loading: true}
	v := m.View(30, 20, false)
	if len(v) == 0 {
		t.Error("expected non-empty view while loading")
	}
}
