package app

// Auto-updates: the app asks GitHub for a newer signed build when it starts,
// and offers it once, without interrupting anything.

import (
	"context"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// checkForUpdate looks for a newer version in the background. It is quiet
// about failure: no network, no release, or a build without updates
// configured all mean the same thing to the user, which is nothing.
func (a *App) checkForUpdate() {
	if !mygo.Updater.Enabled() {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		update, err := mygo.Updater.Check(ctx)
		if err != nil || update == nil {
			return
		}

		if a.win != nil {
			a.win.Update(func() {
				a.update = update
				a.updateOpen = true
			})
		}
	}()
}

// updateDialog offers the version it found.
func (a *App) updateDialog(c *ui.Context, t *ui.Theme) {
	if !a.updateOpen || a.update == nil {
		return
	}

	ui.Modal(c, &a.updateOpen, func() {
		ui.Column(c).Width(520).Gap(10).Children(func() {
			ui.Text(c, "Endpoint "+a.update.Version+" is available").FontSize(13).Bold()
			ui.Text(c, "You are on "+mygo.App.Version()+". The update is signed, and the app restarts when it is installed.").
				FontSize(12).
				TextColor(t.TextMuted)

			if a.updateNotes != "" {
				ui.Text(c, a.updateNotes).FontSize(12).TextColor(t.TextMuted).MaxLines(6)
			}

			if a.updateError != "" {
				ui.Text(c, a.updateError).FontSize(12).TextColor(t.Danger)
			}

			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Later").Clicked() {
					a.updateOpen = false
				}
				if ui.PrimaryButton(c, "Install and restart").Clicked() {
					a.installUpdate()
				}
			})
		})
	})
}

// installUpdate downloads and installs, then relaunches. The download runs on
// its own goroutine; the window stays usable while it goes.
func (a *App) installUpdate() {
	update := a.update
	if update == nil {
		return
	}

	a.updateError = ""
	a.updateNotes = "Downloading…"

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		err := update.Install(ctx, nil)
		if a.win == nil {
			return
		}

		a.win.Update(func() {
			if err != nil {
				a.updateError = "The update could not be installed: " + err.Error()
				a.updateNotes = ""
				return
			}
			a.updateOpen = false
			mygo.App.Relaunch()
		})
	}()
}
