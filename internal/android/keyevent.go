package android

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// keyCodes maps human-friendly key names to Android keyevent codes. See
// https://developer.android.com/reference/android/view/KeyEvent for the full
// list; these are the ones useful when driving a UI by hand.
var keyCodes = map[string]int{
	"back":         4,
	"home":         3,
	"menu":         82,
	"enter":        66,
	"tab":          61,
	"space":        62,
	"del":          67, // backspace
	"backspace":    67,
	"forward_del":  112,
	"escape":       111,
	"esc":          111,
	"up":           19,
	"down":         20,
	"left":         21,
	"right":        22,
	"dpad_center":  23,
	"power":        26,
	"volume_up":    24,
	"volume_down":  25,
	"camera":       27,
	"search":       84,
	"app_switch":   187, // recent apps
	"notification": 83,
}

// ResolveKey converts a named key (case-insensitive) or a raw integer keycode
// string into an Android keyevent code.
func ResolveKey(key string) (int, error) {
	k := strings.TrimSpace(strings.ToLower(key))
	if k == "" {
		return 0, fmt.Errorf("empty key")
	}
	if code, ok := keyCodes[k]; ok {
		return code, nil
	}
	if n, err := strconv.Atoi(k); err == nil {
		return n, nil
	}
	return 0, fmt.Errorf("unknown key %q; use a keycode number or one of: %s", key, strings.Join(KeyNames(), ", "))
}

// KeyNames returns the sorted list of supported named keys.
func KeyNames() []string {
	names := make([]string, 0, len(keyCodes))
	for name := range keyCodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
