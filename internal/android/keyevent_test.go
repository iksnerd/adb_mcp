package android

import "testing"

func TestResolveKey(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"enter", 66},
		{"ENTER", 66},
		{" Back ", 4},
		{"home", 3},
		{"esc", 111},
		{"escape", 111},
		{"del", 67},
		{"app_switch", 187},
		{"66", 66},   // raw numeric passthrough
		{"999", 999}, // arbitrary keycode
	}
	for _, c := range cases {
		got, err := ResolveKey(c.in)
		if err != nil {
			t.Errorf("ResolveKey(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveKey(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestResolveKeyErrors(t *testing.T) {
	for _, in := range []string{"", "  ", "boguskey"} {
		if _, err := ResolveKey(in); err == nil {
			t.Errorf("ResolveKey(%q) expected error, got nil", in)
		}
	}
}

func TestEscapeInputText(t *testing.T) {
	if got := escapeInputText("hello world"); got != "hello%sworld" {
		t.Errorf("escapeInputText spaces = %q", got)
	}
	if got := escapeInputText("a&b"); got != "a\\&b" {
		t.Errorf("escapeInputText ampersand = %q", got)
	}
}
