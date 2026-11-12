package main

// The theme the app draws with: the system's appearance, or the one the user
// picked, filled from the app's tokens.

import (
	"time"

	"github.com/egoist/mygo/ui"
)

const (
	themeAuto = iota
	themeLight
	themeDark
)

// mono is the family of the URL, the bodies and the methods.
const mono = monoFont

// appTheme is the theme of the frame. Themed applies it to the whole window
// in one pass, where SetTheme would only reach the next frame.
func (a *App) appTheme(c *ui.Context) *ui.Theme {
	dark := c.Theme().Dark
	switch a.themeMode {
	case themeLight:
		dark = false
	case themeDark:
		dark = true
	}

	a.tok = tokensFor(dark)

	t := ui.LightTheme()
	if dark {
		t = ui.DarkTheme()
	}
	applyTokens(t, a.tok, textScale(c))
	t.TooltipDelay = 400 * time.Millisecond
	return t
}

// textScale is how much larger the desktop asks text to be: 1 when it asks
// for nothing, which is what a window with no such setting reports.
func textScale(c *ui.Context) float32 {
	scale := c.Preferences().TextScale
	if scale <= 0 {
		return 1
	}
	return scale
}

// applyTokens hands the roles to the widgets, which know nothing of them.
func applyTokens(t *ui.Theme, tok tokens, scale float32) {
	t.Background = tok.Canvas
	t.Surface = tok.Base
	t.SurfaceHover = tok.Fill
	t.SurfacePressed = tok.Fill
	t.Border = tok.Hairline
	t.Text = tok.TextDefault
	t.TextMuted = tok.TextSubtle

	t.Accent = tok.Accent
	t.AccentHover = tok.AccentHover
	t.AccentPressed = tok.AccentPressed
	t.AccentText = tok.AccentText
	t.Selection = tok.Selection

	t.Success = tok.Success
	t.Warning = tok.Warning
	t.Danger = tok.Danger
	t.Focus = tok.Focus

	t.Inverse = tok.TextDefault
	if t.Dark {
		t.Inverse = ui.Hex("#f5f5f7")
		t.InverseText = ui.Hex("#1c1c1e")
	} else {
		t.Inverse = ui.Hex("#1d1d1f")
		t.InverseText = ui.Hex("#ffffff")
	}

	t.FocusRingWidth = 2
	t.FocusRingOffset = 1
	t.Radius = 7
	t.FontSize = 13 * scale
	t.Font = uiFont
	t.Spacing = 4
}
