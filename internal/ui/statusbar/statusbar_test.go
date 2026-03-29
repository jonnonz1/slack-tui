package statusbar

import (
	"strings"
	"testing"
)

func TestStatusbar_SetConnected(t *testing.T) {
	m := New()
	m = m.SetConnected(true)

	v := m.View(80)
	if !strings.Contains(v, "CONNECTED") {
		t.Error("expected CONNECTED in view when connected")
	}
}

func TestStatusbar_SetDisconnected(t *testing.T) {
	m := New()
	m = m.SetConnected(false)

	v := m.View(80)
	if !strings.Contains(v, "DISCONNECTED") {
		t.Errorf("expected DISCONNECTED in view, got: %s", v)
	}
}

func TestStatusbar_SetChannel(t *testing.T) {
	m := New()
	m = m.SetChannel("#engineering")

	v := m.View(80)
	if !strings.Contains(v, "#engineering") {
		t.Errorf("expected channel name in view, got: %s", v)
	}
}

func TestStatusbar_SetUsername(t *testing.T) {
	m := New()
	m = m.SetUsername("alice")

	v := m.View(80)
	if !strings.Contains(v, "@alice") {
		t.Errorf("expected @alice in view, got: %s", v)
	}
}

func TestStatusbar_ViewZeroWidth(t *testing.T) {
	m := New()
	v := m.View(0)
	if v != "" {
		t.Errorf("expected empty view for zero width, got %q", v)
	}
}

func TestStatusbar_DefaultUsername(t *testing.T) {
	m := New()
	v := m.View(80)
	if !strings.Contains(v, "@root_user") {
		t.Errorf("expected default @root_user, got: %s", v)
	}
}
