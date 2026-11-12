package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// The leading title-bar button is only about the sidebar: it hides and shows
// it, and it is named for what it does. The rail leaves the tree when closed,
// so nothing of it is left at the edge and none of its controls can be
// focused.
func TestSidebarToggleHidesTheRail(t *testing.T) {
	app := newTestApp(t)
	tt := ui.NewTester(app.View, 1180, 780)

	if !tt.HasText("Collections") {
		t.Fatal("the rail is not there to begin with")
	}

	if err := tt.Click("Hide Sidebar"); err != nil {
		t.Fatalf("the sidebar button: %v", err)
	}
	if !app.sidebarHidden {
		t.Fatal("the sidebar did not close")
	}
	// The rail stays drawn while it slides away, then it is gone for good.
	tt.Advance(400 * time.Millisecond)
	if tt.HasText("Collections") {
		t.Error("the rail is still built once the motion is over")
	}
	if !tt.HasText("Show Sidebar") {
		t.Error("the button does not offer to show it again")
	}

	if err := tt.Click("Show Sidebar"); err != nil {
		t.Fatalf("the button did not rename itself: %v", err)
	}
	if app.sidebarHidden {
		t.Fatal("the sidebar did not open")
	}
	if !tt.HasText("Collections") {
		t.Error("the rail did not come back")
	}
}
