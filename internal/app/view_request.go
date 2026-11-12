package app

// The request pane: the method and URL bar, the tabs, and the panels.

import (
	"github.com/egoist/mygo/ui"
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

func (a *App) requestBar(c *ui.Context, t *ui.Theme) {
	// Send, New Request and the rest are menu commands, with their
	// shortcuts shown there, so nothing doubles up here.
	ui.Row(c).
		Height(52).
		PaddingX(14).
		Gap(8).
		AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).
		BorderColor(t.Border).
		Children(func() {
			method := ui.Select(c, &a.draft.Method, methods)
			method.Width(96)
			if method.Changed() {
				a.touch(c.Now())
			}

			url := ui.TextInput(c, &a.draft.URL).
				Grow(1).
				Font(mono).
				FontSize(12.5).
				Placeholder("https://api.example.com/v1/users")
			url.Label("Request URL")
			if url.Changed() {
				a.touch(c.Now())
			}

			if a.state == StateLoading {
				if ui.Button(c, "Cancel").Clicked() {
					a.Cancel()
				}
				return
			}
			if ui.PrimaryButton(c, "Send").Clicked() {
				a.startSend()
			}
		})
}

func (a *App) panel(c *ui.Context, t *ui.Theme) {
	ui.Column(c).
		BorderWidth(0, 0, 1, 0).
		BorderColor(t.Border).
		Children(func() {
			switch a.tab {
			case 0:
				a.pairEditor(c, t, &a.draft.Params, "value")
			case 1:
				a.pairEditor(c, t, &a.draft.Headers, "value")
			case 2:
				a.authPanel(c, t)
			default:
				a.bodyPanel(c, t)
			}
		})
}

// pairEditor edits keys and values: query params, headers, form fields. A
// row asked for removal is dropped after the pass that drew it.
func (a *App) pairEditor(c *ui.Context, t *ui.Theme, pairs *[]Pair, valuePlaceholder string) {
	remove := -1

	ui.Scroll(c).Grow(1).Padding(8, 14).Gap(8).Children(func() {
		for i := range *pairs {
			i := i
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				key := ui.TextInput(c, &(*pairs)[i].Key).
					Grow(1).
					Font(mono).
					FontSize(12).
					Placeholder("key")
				value := ui.TextInput(c, &(*pairs)[i].Value).
					Grow(1).
					Font(mono).
					FontSize(12).
					Placeholder(valuePlaceholder)
				if key.Changed() || value.Changed() {
					a.touch(c.Now())
				}
				if ui.Button(c, "×").Tooltip("Remove").Clicked() {
					remove = i
				}
			})
		}

		ui.Row(c).Children(func() {
			if ui.Button(c, "Add").Clicked() {
				*pairs = append(*pairs, Pair{})
			}
		})
	})

	if remove >= 0 {
		*pairs = append((*pairs)[:remove], (*pairs)[remove+1:]...)
		a.touch(c.Now())
	}
}
