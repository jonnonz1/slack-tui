package slack

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache[string](5 * time.Minute)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	v, ok := c.Get("key1")
	if !ok || v != "value1" {
		t.Errorf("expected value1, got %q (ok=%v)", v, ok)
	}

	v, ok = c.Get("key2")
	if !ok || v != "value2" {
		t.Errorf("expected value2, got %q (ok=%v)", v, ok)
	}
}

func TestCache_Miss(t *testing.T) {
	c := NewCache[string](5 * time.Minute)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected cache miss for nonexistent key")
	}
}

func TestCache_Expiry(t *testing.T) {
	c := NewCache[string](10 * time.Millisecond)

	c.Set("ephemeral", "data")

	v, ok := c.Get("ephemeral")
	if !ok || v != "data" {
		t.Errorf("expected data before expiry, got %q (ok=%v)", v, ok)
	}

	time.Sleep(20 * time.Millisecond)

	_, ok = c.Get("ephemeral")
	if ok {
		t.Error("expected cache miss after expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := NewCache[string](5 * time.Minute)

	c.Set("key", "value")
	c.Delete("key")

	_, ok := c.Get("key")
	if ok {
		t.Error("expected cache miss after delete")
	}
}

func TestCache_Clear(t *testing.T) {
	c := NewCache[int](5 * time.Minute)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Clear()

	for _, key := range []string{"a", "b", "c"} {
		if _, ok := c.Get(key); ok {
			t.Errorf("expected cache miss for %q after clear", key)
		}
	}
}

func TestCache_Prune(t *testing.T) {
	c := NewCache[string](10 * time.Millisecond)

	c.Set("old", "stale")
	time.Sleep(20 * time.Millisecond)
	c.Set("new", "fresh")

	c.Prune()

	if _, ok := c.Get("old"); ok {
		t.Error("expected pruned entry to be gone")
	}
	if v, ok := c.Get("new"); !ok || v != "fresh" {
		t.Errorf("expected fresh entry to survive prune, got %q (ok=%v)", v, ok)
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := NewCache[string](5 * time.Minute)

	c.Set("key", "v1")
	c.Set("key", "v2")

	v, ok := c.Get("key")
	if !ok || v != "v2" {
		t.Errorf("expected v2 after overwrite, got %q", v)
	}
}

func TestCache_StructValues(t *testing.T) {
	type item struct {
		Name  string
		Count int
	}

	c := NewCache[item](5 * time.Minute)

	c.Set("x", item{Name: "test", Count: 42})

	v, ok := c.Get("x")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v.Name != "test" || v.Count != 42 {
		t.Errorf("unexpected struct value: %+v", v)
	}
}
