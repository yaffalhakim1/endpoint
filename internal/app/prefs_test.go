package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The theme is a choice about this install, so it has to survive a restart.

func TestThemeChoiceSurvivesARestart(t *testing.T) {
	dir := t.TempDir()

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if app.themeMode != themeAuto {
		t.Fatalf("a new install starts on %d, want auto", app.themeMode)
	}

	app.setThemeMode(themeDark)

	reopened, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.themeMode != themeDark {
		t.Errorf("after reopening the theme is %d, want dark", reopened.themeMode)
	}
}

func TestEveryThemeRoundTrips(t *testing.T) {
	for _, mode := range []int{themeAuto, themeLight, themeDark} {
		dir := t.TempDir()
		app, err := New(dir)
		if err != nil {
			t.Fatal(err)
		}
		app.setThemeMode(mode)

		reopened, err := New(dir)
		if err != nil {
			t.Fatal(err)
		}
		if reopened.themeMode != mode {
			t.Errorf("mode %d came back as %d", mode, reopened.themeMode)
		}
	}
}

func TestBrokenPreferencesFallBackToAuto(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, prefsName), []byte("{ not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	app, err := New(dir)
	if err != nil {
		t.Fatalf("broken preferences stopped the app: %v", err)
	}
	if app.themeMode != themeAuto {
		t.Errorf("the theme is %d, want auto", app.themeMode)
	}
}

func TestPreferencesAreNotInTheWorkspace(t *testing.T) {
	dir := t.TempDir()
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.setThemeMode(themeLight)
	app.save()

	// The workspace is sealed, so the theme cannot be read out of it, and the
	// export must not carry it either.
	data, err := exportDocument(app.doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "theme") {
		t.Error("the exported workspace carries the theme")
	}
}
