package modal

import (
	"testing"

	"github.com/jonnonz1/slack-tui/internal/slack"
)

func testChannels() []slack.Channel {
	return []slack.Channel{
		{ID: "C1", Name: "#general"},
		{ID: "C2", Name: "#engineering"},
		{ID: "C3", Name: "#design-ops"},
		{ID: "C4", Name: "@alice"},
		{ID: "C5", Name: "#random"},
	}
}

func TestSwitcher_OpenClose(t *testing.T) {
	s := NewSwitcher()

	if s.IsOpen() {
		t.Error("switcher should start closed")
	}

	s = s.Open(testChannels())
	if !s.IsOpen() {
		t.Error("switcher should be open after Open()")
	}
	if len(s.filtered) != 5 {
		t.Errorf("expected 5 filtered channels, got %d", len(s.filtered))
	}

	s = s.Close()
	if s.IsOpen() {
		t.Error("switcher should be closed after Close()")
	}
}

func TestSwitcher_Filter(t *testing.T) {
	s := NewSwitcher()
	s = s.Open(testChannels())

	s.query = "eng"
	s.filter()

	if len(s.filtered) != 1 {
		t.Fatalf("expected 1 result for 'eng', got %d", len(s.filtered))
	}
	if s.filtered[0].Name != "#engineering" {
		t.Errorf("expected #engineering, got %s", s.filtered[0].Name)
	}
}

func TestSwitcher_FilterCaseInsensitive(t *testing.T) {
	s := NewSwitcher()
	s = s.Open(testChannels())

	s.query = "DESIGN"
	s.filter()

	if len(s.filtered) != 1 {
		t.Fatalf("expected 1 result for 'DESIGN', got %d", len(s.filtered))
	}
}

func TestSwitcher_FilterEmpty(t *testing.T) {
	s := NewSwitcher()
	s = s.Open(testChannels())

	s.query = ""
	s.filter()

	if len(s.filtered) != 5 {
		t.Errorf("empty query should show all, got %d", len(s.filtered))
	}
}

func TestSwitcher_FilterNoMatch(t *testing.T) {
	s := NewSwitcher()
	s = s.Open(testChannels())

	s.query = "zzzzz"
	s.filter()

	if len(s.filtered) != 0 {
		t.Errorf("expected 0 results for 'zzzzz', got %d", len(s.filtered))
	}
}

func TestSwitcher_FilterResetsCursor(t *testing.T) {
	s := NewSwitcher()
	s = s.Open(testChannels())
	s.cursor = 3

	s.query = "gen"
	s.filter()

	if s.cursor != 0 {
		t.Errorf("filter should reset cursor to 0, got %d", s.cursor)
	}
}

func TestSwitcher_ViewClosed(t *testing.T) {
	s := NewSwitcher()
	v := s.View(80, 40)
	if v != "" {
		t.Errorf("closed switcher should render empty, got %q", v)
	}
}

func TestSwitcher_ViewOpen(t *testing.T) {
	s := NewSwitcher()
	s = s.Open(testChannels())
	v := s.View(80, 40)
	if len(v) == 0 {
		t.Error("open switcher should render non-empty")
	}
}
