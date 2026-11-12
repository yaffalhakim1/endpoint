package app

import (
	"strings"
	"testing"
)

// A huge answer must not be laid out whole, and Copy must still give all of
// it: the pane is a view, not the record.
func TestHugeBodyIsCutForThePaneOnly(t *testing.T) {
	app := newTestApp(t)

	big := []byte(`{"rows":[` + strings.Repeat(`{"id":1},`, 400000) + `{}]}`)
	if len(big) <= maxRendered {
		t.Fatalf("the body is only %d bytes; the cap was not reached", len(big))
	}
	app.result = Result{Code: 200, Status: "200 OK", Body: big}
	app.state = StateLoaded

	shown := app.shownText()
	if len(shown) > maxRendered+200 {
		t.Errorf("the pane was asked to show %d bytes", len(shown))
	}
	if !strings.Contains(shown, "Copy takes all of it") {
		t.Error("the cut is not explained")
	}

	copied := app.copyText()
	if len(copied) < len(big)/2 {
		t.Errorf("Copy gave %d bytes of a %d byte answer", len(copied), len(big))
	}
}
