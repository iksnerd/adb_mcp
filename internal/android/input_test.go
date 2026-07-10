package android

import "testing"

func TestEscapeInputText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello", `'hello'`},
		{"spaces", "hello world", `'hello world'`},
		{"shell metachars", "a&b|c;d>e<f?(g)", `'a&b|c;d>e<f?(g)'`},
		{"dollar and backtick", "$PATH `id`", "'$PATH `id`'"},
		{"double quote", `say "hi"`, `'say "hi"'`},
		{"single quote", "it's", `'it'\''s'`},
		{"empty", "", `''`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := escapeInputText(c.in); got != c.want {
				t.Errorf("escapeInputText(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
