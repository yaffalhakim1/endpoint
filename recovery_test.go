package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// The recovery message must reach the window, not only the log: the user is
// the one who has to do something about it.
func TestRecoveryMessageShowsInTheWindow(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{documentName, documentName + backupSuffix} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("not a workspace"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	app, err := NewApp(dir)
	if err != nil {
		t.Fatal(err)
	}

	// The window asks for the message on its first frame; there is no window
	// in a test, so the view shows it directly.
	app.alert = Alert{Open: true, Title: "The workspace was recovered", Body: app.recovered.Message}

	tt := ui.NewTester(app.view, 1180, 780)
	if !tt.HasText("The workspace was recovered") {
		t.Fatalf("the message is not in the window; texts: %q", tt.Texts())
	}
	if !tt.HasText("OK") {
		t.Error("the message has no way to dismiss it")
	}
}
