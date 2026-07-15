package android

import (
	"strings"
	"testing"
)

func TestLogFilterApply(t *testing.T) {
	raw := `07-15 10:00:00.000  1234  1234 E MyTag   : boom
07-15 10:00:01.000  1234  1234 D MyTag   : quiet debug line
07-15 10:00:02.000  1234  1234 W OtherTag: warning here
07-15 10:00:03.000  1234  1234 I 9999    : chatty: uid=10123 expire 3 lines
07-15 10:00:04.000  1234  1234 I MyTag   : mentions chatty in the message text`

	cases := []struct {
		name string
		f    LogFilter
		want []string
	}{
		{
			name: "no filter strips only chatty noise",
			f:    LogFilter{},
			want: []string{
				"07-15 10:00:00.000  1234  1234 E MyTag   : boom",
				"07-15 10:00:01.000  1234  1234 D MyTag   : quiet debug line",
				"07-15 10:00:02.000  1234  1234 W OtherTag: warning here",
				"07-15 10:00:04.000  1234  1234 I MyTag   : mentions chatty in the message text",
			},
		},
		{
			name: "substring",
			f:    LogFilter{Substring: "warning"},
			want: []string{"07-15 10:00:02.000  1234  1234 W OtherTag: warning here"},
		},
		{
			name: "priority E keeps E and above only",
			f:    LogFilter{Priority: "e"},
			want: []string{"07-15 10:00:00.000  1234  1234 E MyTag   : boom"},
		},
		{
			name: "priority W keeps W and above",
			f:    LogFilter{Priority: "W"},
			want: []string{
				"07-15 10:00:00.000  1234  1234 E MyTag   : boom",
				"07-15 10:00:02.000  1234  1234 W OtherTag: warning here",
			},
		},
		{
			name: "tags OR match, case-insensitive",
			f:    LogFilter{Tags: []string{"othertag"}},
			want: []string{"07-15 10:00:02.000  1234  1234 W OtherTag: warning here"},
		},
		{
			name: "priority and tags combine as AND",
			f:    LogFilter{Priority: "I", Tags: []string{"MyTag"}},
			want: []string{
				// Both lines carry the MyTag tag; priority=I keeps I and
				// above, which includes E (E outranks I) as well as I itself.
				"07-15 10:00:00.000  1234  1234 E MyTag   : boom",
				"07-15 10:00:04.000  1234  1234 I MyTag   : mentions chatty in the message text",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			applied := c.f.apply(raw)
			var got []string
			if applied != "" {
				got = strings.Split(applied, "\n")
			}
			if len(got) != len(c.want) {
				t.Fatalf("apply() = %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("line %d = %q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestLogFilterValidate(t *testing.T) {
	for _, p := range []string{"", "V", "d", "I", "W", "E", "F"} {
		if err := (LogFilter{Priority: p}).validate(); err != nil {
			t.Errorf("validate() with priority %q: unexpected error %v", p, err)
		}
	}
	for _, p := range []string{"X", "error", "ee"} {
		if err := (LogFilter{Priority: p}).validate(); err == nil {
			t.Errorf("validate() with priority %q: expected error, got nil", p)
		}
	}
}

func TestParseLogLine(t *testing.T) {
	prio, tag, ok := parseLogLine("07-15 10:00:00.000  1234  1234 E MyTag   : boom")
	if !ok || prio != "E" || tag != "MyTag" {
		t.Errorf("parseLogLine() = (%q, %q, %v), want (E, MyTag, true)", prio, tag, ok)
	}
	if _, _, ok := parseLogLine("not a logcat line"); ok {
		t.Error("parseLogLine() on a malformed line: expected ok=false")
	}
}
