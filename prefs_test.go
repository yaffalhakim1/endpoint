package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// What the desktop reports, which decides whether anything animates at all.
func TestPreferencesAreRead(t *testing.T) {
	var reduce bool

	tt := ui.NewTester(func(c *ui.Context) {
		prefs := c.Preferences()
		reduce = prefs.ReduceMotion
	}, 120, 120)
	_ = tt

	t.Logf("reduce motion: %v", reduce)
	if reduce {
		t.Log("animations are off in this environment, by the desktop's own setting")
	}
}
