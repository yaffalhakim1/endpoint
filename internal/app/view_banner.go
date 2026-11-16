package app

// A request can name a variable the project does not define. Rather than send
// the literal braces, the window says so above the tabs, and offers the way to
// fix it.

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// unresolved lists the variables a request names that the project does not
// define. Only the URL is checked: a missing value there is the one that makes
// a request go somewhere wrong rather than nowhere.
func (a *App) unresolved() []string {
	project := a.doc.ActiveProject()
	if project == nil {
		return nil
	}
	available := project.Variables()

	var missing []string
	for _, name := range namesIn(a.draft.URL) {
		if _, ok := lookupVar(available, name); !ok {
			missing = append(missing, name)
		}
	}
	return missing
}

// namesIn reads the {{name}} placeholders out of a string.
func namesIn(s string) []string {
	var names []string
	rest := s
	for {
		start := strings.Index(rest, "{{")
		if start < 0 {
			return names
		}
		end := strings.Index(rest[start:], "}}")
		if end < 0 {
			return names
		}
		name := strings.TrimSpace(rest[start+2 : start+end])
		if name != "" {
			names = append(names, name)
		}
		rest = rest[start+end+2:]
	}
}

func lookupVar(vars []Pair, name string) (string, bool) {
	for _, pair := range vars {
		if pair.Key == name {
			return pair.Value, true
		}
	}
	return "", false
}

// unresolvedBanner is the line above the tabs, drawn only when something is
// missing.
func (a *App) unresolvedBanner(c *ui.Context, t *ui.Theme) {
	missing := a.unresolved()
	if len(missing) == 0 {
		return
	}

	ui.Row(c).
		Margin(10, 16, 0, 16).
		Height(36).
		Padding(0, 12).
		Radius(8).
		Gap(8).
		AlignItems(ui.Center).
		Background(tint(a.tok.Warning)).
		Children(func() {
			ui.Text(c, "{{"+missing[0]+"}} is unresolved. Choose an environment before sending.").
				FontSize(13).
				TextColor(a.tok.Warning).
				Grow(1).
				SingleLine().
				Ellipsis("…")
			if ui.Button(c, "Edit variables").Clicked() {
				a.varsOpen = true
			}
		})
}

// tabLabels carries the counts the design puts beside the tabs: how many
// headers there are, and whether auth is set.
func (a *App) tabLabels() []string {
	auth := "Auth"
	if a.draft.Auth.Type != "" && a.draft.Auth.Type != AuthNone {
		auth = "Auth •"
	}
	return []string{"Params", "Headers " + itoa(len(a.draft.Headers)), auth, "Body"}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
