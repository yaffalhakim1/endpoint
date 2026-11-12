package app

// How long a layout change takes. One place decides, so the sidebar and the
// pane that takes its room move together, and the desktop's wish for less
// motion turns both off: a negative duration is how a transition asks for
// none at all.

import (
	"time"

	"github.com/egoist/mygo/ui"
)

const collapseMotion = 200 * time.Millisecond

func motionDuration(c *ui.Context) time.Duration {
	if c.Preferences().ReduceMotion {
		return -1
	}
	return collapseMotion
}
