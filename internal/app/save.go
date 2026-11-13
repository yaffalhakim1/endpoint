package app

// Files the app writes for the user: an exported workspace, a saved response.
// There is no file dialog in the toolkit, so they land in a folder the user
// can find and the app says exactly where.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// saveDir is where an exported file goes: Downloads, then Desktop, then the
// user's home.
func saveDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	for _, name := range []string{"Downloads", "Desktop"} {
		dir := filepath.Join(home, name)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return home, nil
}

// writeFile puts data in the save folder under a name built from the stem and
// the moment, so nothing is overwritten. It returns the path it wrote.
func (a *App) writeFile(stem, extension string, data []byte) (string, error) {
	dir, err := saveDir()
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("%s-%s%s", stem, time.Now().Format("2006-01-02-1504"), extension)
	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// fileName makes a request name safe for a file system.
func fileName(name string) string {
	if name == "" {
		return "endpoint"
	}

	var out strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out.WriteRune(r)
		case r == ' ', r == '-', r == '_':
			out.WriteRune('-')
		}
	}

	cleaned := strings.Trim(out.String(), "-")
	if cleaned == "" {
		return "endpoint"
	}
	if len(cleaned) > 40 {
		cleaned = cleaned[:40]
	}
	return cleaned
}
