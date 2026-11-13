package app

// Undo for the destructive edits. Rather than a trash list with a restore
// path per kind, the app keeps the last few documents whole: a delete, a
// rename or a move is one snapshot, and undo puts the document back as it
// was. The document is a few kilobytes, so this costs nothing.

import (
	"encoding/json"

	"github.com/egoist/mygo/ui"
)

// undoDepth is how many edits back the app can go.
const undoDepth = 20

// snapshot remembers the document as it is now, before an edit changes it.
func (a *App) snapshot() {
	data, err := json.Marshal(a.doc)
	if err != nil {
		return
	}

	var copy Document
	if err := json.Unmarshal(data, &copy); err != nil {
		return
	}

	a.undo = append(a.undo, &copy)
	if len(a.undo) > undoDepth {
		a.undo = a.undo[len(a.undo)-undoDepth:]
	}
}

// undoLast puts the document back as it was before the last edit, and says
// whether there was anything to undo.
func (a *App) undoLast() bool {
	if len(a.undo) == 0 {
		return false
	}

	last := a.undo[len(a.undo)-1]
	a.undo = a.undo[:len(a.undo)-1]

	a.doc = last
	a.detach()
	a.selectFirst()
	a.state = StateEmpty
	a.answer = nil
	a.save()
	return true
}

// canUndo says whether the menu item should be enabled.
func (a *App) canUndo() bool { return len(a.undo) > 0 }

// offerUndo shows the toast that makes a delete reversible in the moment.
func (a *App) offerUndo(c *ui.Context, what string) {
	if !a.canUndo() {
		return
	}
	c.ToastAction(what+" deleted", "Undo", func() {
		if a.undoLast() && a.win != nil {
			a.win.Invalidate()
		}
	})
}
