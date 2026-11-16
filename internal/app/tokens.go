package app

// The app's colors, named by the job they do rather than by what they look
// like. A view asks for "hairline" or "textSubtle"; the value behind the role
// changes with the appearance, so no view holds a colour of its own.
//
// The palette is the one from the design reference: indigo on white in light,
// neutral near-black in dark, with a code surface that sits one step away from
// the pane so data reads as data.

import "github.com/egoist/mygo/ui"

type tokens struct {
	Canvas ui.Color // the rail, behind the content
	Base   ui.Color // the pane content sits on
	Code   ui.Color // the surface of a code editor
	Field  ui.Color // text fields and inputs
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

	Focus ui.Color
}

func tokensFor(dark bool) tokens {
	if dark {
		return tokens{
			Canvas:        ui.Hex("#1d1d21"),
			Base:          ui.Hex("#18181b"),
			Code:          ui.Hex("#131316"),
			Field:         ui.Hex("#202025"),
			Fill:          ui.Hex("#2a2a31"),
			Hairline:      ui.Hex("#2c2c33"),
			Line:          ui.Hex("#232329"),
			TextDefault:   ui.Hex("#ececf0"),
			TextSubtle:    ui.Hex("#a1a1ad"),
			TextInactive:  ui.Hex("#8a8a96"),
			Accent:        ui.Hex("#4f46e5"),
			AccentHover:   ui.Hex("#6366f1"),
			AccentPressed: ui.Hex("#4338ca"),
			AccentText:    ui.Hex("#ffffff"),
			Selection:     ui.RGBA(129, 140, 248, 0.16),
			Success:       ui.Hex("#5be0b0"),
			Warning:       ui.Hex("#fbbf24"),
			Info:          ui.Hex("#7cb4ff"),
			Danger:        ui.Hex("#f87171"),
			Focus:         ui.Hex("#aab4ff"),
		}
	}

	return tokens{
		Canvas:        ui.Hex("#f6f6f8"),
		Base:          ui.Hex("#ffffff"),
		Code:          ui.Hex("#f7f7fa"),
		Field:         ui.Hex("#ffffff"),
		Fill:          ui.Hex("#ececf1"),
		Hairline:      ui.Hex("#e3e3ea"),
		Line:          ui.Hex("#ededf2"),
		TextDefault:   ui.Hex("#18181b"),
		TextSubtle:    ui.Hex("#52525b"),
		TextInactive:  ui.Hex("#6b6b76"),
		Accent:        ui.Hex("#4f46e5"),
		AccentHover:   ui.Hex("#4338ca"),
		AccentPressed: ui.Hex("#3730a3"),
		AccentText:    ui.Hex("#ffffff"),
		Selection:     ui.RGBA(79, 70, 229, 0.10),
		Success:       ui.Hex("#166534"),
		Warning:       ui.Hex("#92400e"),
		Info:          ui.Hex("#1d4ed8"),
		Danger:        ui.Hex("#b91c1c"),
		Focus:         ui.Hex("#4f46e5"),
	}
}

// tint is a role's colour at the strength of a badge behind it.
func tint(c ui.Color) ui.Color { return c.Alpha(0.12) }
