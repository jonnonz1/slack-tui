package input

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func pressKey(code rune) tea.KeyMsg {
	return tea.KeyPressMsg{Code: code}
}

func pressKeyMod(code rune, mod tea.KeyMod) tea.KeyMsg {
	return tea.KeyPressMsg{Code: code, Mod: mod}
}

func TestInput_TypeAndSend(t *testing.T) {
	m := New()

	m, _ = m.Update(pressKey('h'))
	m, _ = m.Update(pressKey('i'))

	if m.text != "hi" {
		t.Errorf("expected text 'hi', got %q", m.text)
	}
	if m.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", m.cursor)
	}

	var cmd tea.Cmd
	m, cmd = m.Update(pressKey(tea.KeyEnter))

	if m.text != "" {
		t.Errorf("expected text cleared after send, got %q", m.text)
	}
	if cmd == nil {
		t.Fatal("expected a command after send")
	}

	msg := cmd()
	sendMsg, ok := msg.(SendMessageMsg)
	if !ok {
		t.Fatalf("expected SendMessageMsg, got %T", msg)
	}
	if sendMsg.Text != "hi" {
		t.Errorf("expected sent text 'hi', got %q", sendMsg.Text)
	}
}

func TestInput_EmptyEnterDoesNotSend(t *testing.T) {
	m := New()

	m, cmd := m.Update(pressKey(tea.KeyEnter))
	if cmd != nil {
		t.Error("empty input should not produce a send command")
	}
	_ = m
}

func TestInput_Backspace(t *testing.T) {
	m := New()

	m, _ = m.Update(pressKey('a'))
	m, _ = m.Update(pressKey('b'))
	m, _ = m.Update(pressKey('c'))
	m, _ = m.Update(pressKey(tea.KeyBackspace))

	if m.text != "ab" {
		t.Errorf("expected 'ab' after backspace, got %q", m.text)
	}
	if m.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", m.cursor)
	}
}

func TestInput_BackspaceAtStart(t *testing.T) {
	m := New()
	m.text = "hello"
	m.cursor = 0

	m, _ = m.Update(pressKey(tea.KeyBackspace))

	if m.text != "hello" {
		t.Errorf("backspace at start should not change text, got %q", m.text)
	}
}

func TestInput_CursorMovement(t *testing.T) {
	m := New()
	m.text = "hello"
	m.cursor = 3

	m, _ = m.Update(pressKey(tea.KeyLeft))
	if m.cursor != 2 {
		t.Errorf("expected cursor at 2 after left, got %d", m.cursor)
	}

	m, _ = m.Update(pressKey(tea.KeyRight))
	if m.cursor != 3 {
		t.Errorf("expected cursor at 3 after right, got %d", m.cursor)
	}
}

func TestInput_CursorBounds(t *testing.T) {
	m := New()
	m.text = "hi"
	m.cursor = 0

	m, _ = m.Update(pressKey(tea.KeyLeft))
	if m.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m.cursor)
	}

	m.cursor = 2
	m, _ = m.Update(pressKey(tea.KeyRight))
	if m.cursor != 2 {
		t.Errorf("cursor should not go past len, got %d", m.cursor)
	}
}

func TestInput_SetSize(t *testing.T) {
	m := New()
	m = m.SetSize(80, 3)

	if m.width != 80 || m.height != 3 {
		t.Errorf("expected 80x3, got %dx%d", m.width, m.height)
	}
}

func TestInput_ViewEmpty(t *testing.T) {
	m := New()
	v := m.View(0, 0, false, "#test")
	if v != "" {
		t.Errorf("expected empty view for zero width, got %q", v)
	}
}

func TestInput_ViewShowsChannel(t *testing.T) {
	m := New()
	v := m.View(80, 3, true, "#general")
	if len(v) == 0 {
		t.Error("expected non-empty view")
	}
}
