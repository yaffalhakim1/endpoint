package app

// A plain message the app has to say: what went wrong, and what to do about
// it. It is not for questions; those are dialogs.

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Alert is a message shown over the window.
type Alert struct {
	Open  bool
	Title string
	Body  string
}

func (a *App) alertDialog(c *ui.Context, t *ui.Theme) {
	if !a.alert.Open {
		return
	}

	ui.Modal(c, &a.alert.Open, func() {
		ui.Column(c).Width(460).Gap(10).Children(func() {
			ui.Text(c, a.alert.Title).FontSize(13).Bold()
			ui.Text(c, a.alert.Body).FontSize(12.5).TextColor(t.TextMuted).Selectable()
			ui.Row(c).Justify(ui.End).Children(func() {
				if ui.PrimaryButton(c, "OK").Clicked() {
					a.alert.Open = false
				}
			})
		})
	})
}

// showAbout names the build and where its workspace lives, which is what a
// bug report needs.
func (a *App) showAbout() {
	a.alert = Alert{
		Open:  true,
		Title: "Endpoint " + mygo.App.Version(),
		Body: "A desktop HTTP client in Go.\n\n" +
			"Workspace: " + a.dir + "\n" +
			"MIT licensed. Type: Inter and Geist Mono, under the SIL Open Font License.",
	}
}
