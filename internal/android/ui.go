package android

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// dumpOnce performs a single uiautomator dump + parse, retrying once on the
// transient "could not get idle state" error that occurs mid-animation.
func dumpOnce(ctx context.Context, serial string) ([]Element, error) {
	const remote = "/sdcard/window_dump.xml"
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(1500 * time.Millisecond)
		}
		if _, err := runAdb(ctx, serial, "shell", "uiautomator", "dump", remote); err != nil {
			lastErr = err
			continue
		}
		out, err := runAdbBytes(ctx, serial, "shell", "cat", remote)
		if err != nil {
			lastErr = err
			continue
		}
		xmlStr := string(out)
		if i := strings.Index(xmlStr, "<?xml"); i > 0 {
			xmlStr = xmlStr[i:]
		}
		elems, err := ParseHierarchy(xmlStr)
		if err != nil {
			lastErr = err
			continue
		}
		return elems, nil
	}
	return nil, fmt.Errorf("uiautomator dump failed (device busy?): %w", lastErr)
}

// DescribeUI reads the UI hierarchy and returns a *settled* snapshot with
// true-pixel bounds and precomputed centers. uiautomator can hand back a STALE
// tree mid-refresh/animation (e.g. a list's empty-state while it is really
// populating). To guard against that, DescribeUI dumps twice ~500ms apart; if
// the snapshots differ the UI is still changing, so it takes one more after a
// longer wait and returns that freshest snapshot.
//
// Note: content drawn on a canvas (React Native / Flutter / Skia PIN pads) does
// not appear in the hierarchy at all — no amount of settling surfaces it. For
// those, use a screenshot to read state and tap by coordinate (see enter_pin's
// grid/coords options for custom pads).
func DescribeUI(ctx context.Context, serial string) ([]Element, error) {
	first, err := dumpOnce(ctx, serial)
	if err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
	second, err := dumpOnce(ctx, serial)
	if err != nil {
		return first, nil // a failed re-dump is non-fatal; return what we have
	}
	if signature(first) == signature(second) {
		return second, nil
	}
	time.Sleep(1000 * time.Millisecond)
	third, err := dumpOnce(ctx, serial)
	if err != nil {
		return second, nil
	}
	return third, nil
}

// signature is a cheap fingerprint of a hierarchy used to detect whether two
// consecutive dumps describe the same (settled) screen.
func signature(elems []Element) string {
	var b strings.Builder
	for i := range elems {
		e := &elems[i]
		fmt.Fprintf(&b, "%s|%s|%s|%d,%d;", e.Text, e.Desc, e.ResourceID, e.Bounds.X1, e.Bounds.Y1)
	}
	return b.String()
}

// WaitForText polls the UI hierarchy until an element matching query (by text or
// content-description) appears, or timeout elapses. It is the reliable
// alternative to a manual sleep-then-screenshot loop after an async action.
func WaitForText(ctx context.Context, serial, query string, partial bool, timeout time.Duration) (Element, error) {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return Element{}, err
		}
		elems, err := dumpOnce(ctx, serial)
		if err == nil {
			if e, ok := FindByText(elems, query, partial); ok {
				return e, nil
			}
		}
		if time.Now().After(deadline) {
			return Element{}, fmt.Errorf("text %q did not appear within %s", query, timeout)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
