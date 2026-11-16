package app

// The Auth panel's field, with the reveal the design asks for: a secret is
// masked until you ask to see it, and the line under it says how it will be
// sent.

import (
	"github.com/egoist/mygo/ui"
)

func (a *App) authField(c *ui.Context, t *ui.Theme, label string, value *string, secret bool) {
	ui.Column(c).Gap(5).Children(func() {
		ui.Text(c, label).FontSize(11.5).TextColor(t.TextMuted)

		ui.Row(c).Gap(8).AlignItems(ui.Center).MaxWidth(440).Children(func() {
			input := ui.TextInput(c, value).Grow(1).Font(mono).FontSize(12.5)
			input.Label(label)
			if secret && !a.showSecret {
				input.Password()
			}
			if input.Changed() {
				a.touch(c.Now())
			}

			if !secret {
				return
			}

			label := "Show " + label
			if a.showSecret {
				label = "Hide " + label
			}
			button := ui.Button(c, "Show")
			button.Label(label).Tooltip(label)
			if a.showSecret {
				button.Label("Hide")
			}
			if button.Clicked() {
				a.showSecret = !a.showSecret
			}
		})
	})
}
