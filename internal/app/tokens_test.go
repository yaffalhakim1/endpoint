package app

// The colour roles must stay readable. This is the check the design comp had
// in JavaScript; here it runs on the same values the app draws with.

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func luminance(c ui.Color) float64 {
	channel := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.04045 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
}

func contrast(fg, bg ui.Color) float64 {
	a := luminance(fg.Over(bg))
	b := luminance(bg)
	return (math.Max(a, b) + 0.05) / (math.Min(a, b) + 0.05)
}

func TestTokensMeetAA(t *testing.T) {
	for _, dark := range []bool{false, true} {
		tok := tokensFor(dark)
		appearance := "light"
		if dark {
			appearance = "dark"
		}

		pairs := []struct {
			name string
			fg   ui.Color
			bg   ui.Color
			min  float64
		}{
			{"text on the canvas", tok.TextDefault, tok.Canvas, 4.5},
			{"text on the base", tok.TextDefault, tok.Base, 4.5},
			{"subtle text on the canvas", tok.TextSubtle, tok.Canvas, 4.5},
			{"subtle text on the base", tok.TextSubtle, tok.Base, 4.5},
			{"text on a selected row", tok.TextDefault, tok.Selection.Over(tok.Canvas), 4.5},
			{"button text on the accent", tok.AccentText, tok.Accent, 4.5},
			{"danger text on the base", tok.Danger, tok.Base, 4.5},
			{"success text on the base", tok.Success, tok.Base, 4.5},
			{"warning text on the base", tok.Warning, tok.Base, 4.5},
			{"info text on the base", tok.Info, tok.Base, 4.5},
			{"method text in the rail", tok.Danger, tok.Base, 4.5},
			{"the focus ring on the base", tok.Focus, tok.Base, 3},
		}

		for _, p := range pairs {
			got := contrast(p.fg, p.bg)
			if got < p.min {
				t.Errorf("%s, %s: %.2f:1, below %.1f:1", appearance, p.name, got, p.min)
			}
		}
	}
}
