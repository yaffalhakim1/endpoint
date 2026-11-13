package app

// The HTTP options dialog: how the active project reaches the network.

import (
	"github.com/egoist/mygo/ui"
)

func (a *App) httpDialog(c *ui.Context, t *ui.Theme) {
	if !a.optionsOpen {
		return
	}

	project := a.doc.ActiveProject()
	if project == nil {
		a.optionsOpen = false
		return
	}

	ui.Modal(c, &a.optionsOpen, func() {
		ui.Column(c).Width(520).Gap(12).Children(func() {
			ui.Text(c, "HTTP options").FontSize(13).Bold()
			ui.Text(c, "These apply to every request sent from "+project.Name+".").
				FontSize(12).
				TextColor(t.TextMuted)

			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				timeout := float64(project.HTTP.TimeoutSeconds)
				ui.NumberInput(c, &timeout, 0, 3600, 5).Width(110)
				if int(timeout) != project.HTTP.TimeoutSeconds {
					project.HTTP.TimeoutSeconds = int(timeout)
					a.save()
				}
				ui.Column(c).Gap(1).Children(func() {
					ui.Text(c, "Timeout").FontSize(12.5)
					ui.Text(c, "Seconds. 0 waits as long as the server takes.").
						FontSize(11.5).
						TextColor(t.TextMuted)
				})
			})

			follow := !project.HTTP.NoRedirects
			if ui.Checkbox(c, &follow, "Follow redirects").Changed() {
				project.HTTP.NoRedirects = !follow
				a.save()
			}

			insecure := project.HTTP.InsecureTLS
			if ui.Checkbox(c, &insecure, "Accept a certificate nothing signed").Changed() {
				project.HTTP.InsecureTLS = insecure
				a.save()
			}

			cookies := !project.HTTP.NoCookies
			if ui.Checkbox(c, &cookies, "Keep cookies between calls").Changed() {
				project.HTTP.NoCookies = !cookies
				a.client = nil // the next call builds a client with the new jar
				a.save()
			}

			ui.Column(c).Gap(5).Children(func() {
				ui.Text(c, "Proxy").FontSize(11.5).TextColor(t.TextMuted)
				proxy := ui.TextInput(c, &project.HTTP.Proxy).
					Font(mono).
					FontSize(12.5).
					Placeholder("http://127.0.0.1:8080, empty for the system's")
				proxy.Label("Proxy")
				if proxy.Changed() {
					a.save()
				}
			})

			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if ui.Button(c, "Clear cookies").Clicked() {
					a.client = nil
					c.Toast("Cookies cleared")
				}
				ui.Spacer(c)
				if ui.PrimaryButton(c, "Done").Clicked() {
					a.optionsOpen = false
				}
			})
		})
	})
}
