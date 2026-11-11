package main

// The app's colors, named by the job they do rather than by what they look
// like. A view asks for "hairline" or "textSubtle"; the value behind the role
// changes with the appearance, so no view holds a colour of its own.
//
// The vocabulary is Cloudflare's Kumo: canvas and base for the two surfaces,
// hairline for the line between regions, fill for a wash over a control, and
// the intent colours with their tints. The values are Apple's, kept.

import "github.com/egoist/mygo/ui"

type tokens struct {
	Canvas ui.Color // the window behind the panes
	Base   ui.Color // the pane content sits on
	Fill   ui.Color // the wash under a hovered or pressed control

	Hairline ui.Color // the line between regions
	Line     ui.Color // a quieter line, inside a region

	TextDefault  ui.Color
	TextSubtle   ui.Color
	TextInactive ui.Color

	Accent        ui.Color
	AccentHover   ui.Color
	AccentPressed ui.Color
	AccentText    ui.Color
	Selection     ui.Color // the tint behind a chosen row

	Success ui.Color
	Warning ui.Color
	Info    ui.Color
	Danger  ui.Color

	Focus ui.Color // a neutral ring, as Kumo draws it
}

func tokensFor(dark bool) tokens {
	if dark {
		return tokens{
			Canvas:        ui.Hex("#1c1c1e"),
			Base:          ui.Hex("#232325"),
			Fill:          ui.Hex("#2c2c2e"),
			Hairline:      ui.Hex("#38383a"),
			Line:          ui.Hex("#2c2c2e"),
			TextDefault:   ui.Hex("#f5f5f7"),
			TextSubtle:    ui.Hex("#98989d"),
			TextInactive:  ui.Hex("#6e6e73"),
			Accent:        ui.Hex("#e8172c"),
			AccentHover:   ui.Hex("#ff3b4f"),
			AccentPressed: ui.Hex("#d91f33"),
			AccentText:    ui.Hex("#ffffff"),
			Selection:     ui.Hex("#e8172c").Alpha(0.20),
			Success:       ui.Hex("#30d158"),
			Warning:       ui.Hex("#ff9f0a"),
			Info:          ui.Hex("#409cff"),
			Danger:        ui.Hex("#ff453a"),
			Focus:         ui.Hex("#f5f5f7"),
		}
	}

	return tokens{
		Canvas:        ui.Hex("#f5f5f7"),
		Base:          ui.Hex("#ffffff"),
		Fill:          ui.Hex("#ebebed"),
		Hairline:      ui.Hex("#d2d2d7"),
		Line:          ui.Hex("#e6e6ea"),
		TextDefault:   ui.Hex("#1d1d1f"),
		TextSubtle:    ui.Hex("#6e6e73"),
		TextInactive:  ui.Hex("#aeaeb2"),
		Accent:        ui.Hex("#e8172c"),
		AccentHover:   ui.Hex("#ff3b4f"),
		AccentPressed: ui.Hex("#d91f33"),
		AccentText:    ui.Hex("#ffffff"),
		Selection:     ui.Hex("#e8172c").Alpha(0.20),
		Success:       ui.Hex("#17784a"),
		Warning:       ui.Hex("#c93400"),
		Info:          ui.Hex("#0071e3"),
		Danger:        ui.Hex("#d70015"),
		Focus:         ui.Hex("#1d1d1f"),
	}
}

// tint is a role's colour at the strength of a badge behind it.
func tint(c ui.Color) ui.Color { return c.Alpha(0.12) }
