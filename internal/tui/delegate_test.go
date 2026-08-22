package tui

import (
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
)

func TestTruncateAndPadRunes(t *testing.T) {
	// ASCII string
	s1 := "hello world"
	trunc1 := truncateRunes(s1, 8)
	if runewidth.StringWidth(trunc1) > 8 {
		t.Errorf("expected width <= 8, got %d for %q", runewidth.StringWidth(trunc1), trunc1)
	}

	// Multibyte / Unicode string
	s2 := "こんにちは世界"
	trunc2 := truncateRunes(s2, 8)
	if runewidth.StringWidth(trunc2) > 8 {
		t.Errorf("expected width <= 8, got %d for %q", runewidth.StringWidth(trunc2), trunc2)
	}

	// Padding
	padded := padRunes("test", 10)
	if runewidth.StringWidth(padded) != 10 {
		t.Errorf("expected padded width 10, got %d (%q)", runewidth.StringWidth(padded), padded)
	}
}

func TestTimeSince(t *testing.T) {
	now := time.Now()
	if timeSince(now) != "Just now" {
		t.Errorf("expected 'Just now', got %s", timeSince(now))
	}

	tenMinsAgo := now.Add(-10 * time.Minute)
	if timeSince(tenMinsAgo) != "10m ago" {
		t.Errorf("expected '10m ago', got %s", timeSince(tenMinsAgo))
	}

	fiveHoursAgo := now.Add(-5 * time.Hour)
	if timeSince(fiveHoursAgo) != "5h ago" {
		t.Errorf("expected '5h ago', got %s", timeSince(fiveHoursAgo))
	}
}
