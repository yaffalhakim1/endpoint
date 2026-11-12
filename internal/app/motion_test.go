package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Closing the sidebar must move the pane that takes its room, not snap it.
// The request field is the probe: its x is where the pane starts.
func TestSidebarCollapseIsAnimated(t *testing.T) {
	app := newTestApp(t)
	tt := ui.NewTester(app.View, 1180, 780)

	open, ok := tt.Find("Request URL")
	if !ok {
		t.Fatal("the request field is not there")
	}

	// Close it the way the button does.
	app.sidebarHidden = true
	tt.Frame()
	tt.Advance(60 * time.Millisecond)

	middle, ok := tt.Find("Request URL")
	if !ok {
		t.Fatal("the request field is gone mid-way")
	}

	tt.Advance(500 * time.Millisecond)
	closed, ok := tt.Find("Request URL")
	if !ok {
		t.Fatal("the request field is gone once closed")
	}

	if closed.X >= open.X {
		t.Fatalf("the pane did not move left: %.1f open, %.1f closed", open.X, closed.X)
	}
	if middle.X == closed.X {
		t.Errorf("the pane jumped: %.1f mid-way and %.1f closed, so nothing animated", middle.X, closed.X)
	}
}
