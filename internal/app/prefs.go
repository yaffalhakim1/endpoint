package app

// Preferences that belong to this install rather than to the workspace: how
// the window is drawn. They live beside the workspace in a small file, in the
// clear, because there is nothing secret in them.

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const prefsName = "prefs.json"

type prefs struct {
	// Theme is "auto", "light" or "dark". Empty means auto.
	Theme string `json:"theme"`
}

func loadPrefs(dir string) prefs {
	data, err := os.ReadFile(filepath.Join(dir, prefsName))
	if err != nil {
		return prefs{}
	}

	var p prefs
	if err := json.Unmarshal(data, &p); err != nil {
		return prefs{}
	}
	return p
}

// save writes the preferences, best effort: a window that cannot remember its
// theme is still a working window.
func savePrefs(dir string, p prefs) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, prefsName), data, 0o644)
}

// themeMode maps a stored name to the constant the view uses.
func themeModeOf(name string) int {
	switch name {
	case "light":
		return themeLight
	case "dark":
		return themeDark
	default:
		return themeAuto
	}
}

// themeNameOf is the inverse, for writing.
func themeNameOf(mode int) string {
	switch mode {
	case themeLight:
		return "light"
	case themeDark:
		return "dark"
	default:
		return "auto"
	}
}

// setThemeMode changes how the window is drawn and remembers the choice.
func (a *App) setThemeMode(mode int) {
	a.themeMode = mode
	savePrefs(a.dir, prefs{Theme: themeNameOf(mode)})
}
