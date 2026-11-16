package app

// The code editors draw their own background, and by default that is the
// window's canvas, which leaves them grey inside a pane that is white. The
// editors take a theme, so they can be given the same surface the pane has.

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

// editorTheme is the code theme for a window theme, with the code surface
// behind the text so a body reads as data rather than as chrome.
func editorTheme(t *ui.Theme, tok tokens) *editor.Theme {
	theme := editor.DefaultTheme(t)
	theme.Background = tok.Code
	return theme
}
