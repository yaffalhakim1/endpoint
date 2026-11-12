package app

// The icons the title bar draws. They are inline SVG so the app needs no
// image files, and ui.Icon takes their colour from the text around them.

import "github.com/egoist/mygo/ui"

var (
	minimizeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><path d="M2.2 5h5.6" stroke="#000" stroke-width="1.1" stroke-linecap="round"/></svg>`))
	maximizeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><rect x="2.4" y="2.4" width="5.2" height="5.2" rx="1" fill="none" stroke="#000" stroke-width="1.1"/></svg>`))
	restoreIcon  = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><rect x="1.8" y="3.2" width="4.6" height="4.6" rx="1" fill="none" stroke="#000" stroke-width="1.1"/><path d="M3.6 2.6V2.2c0-.3.2-.5.5-.5h3.7c.3 0 .5.2.5.5v3.7c0 .3-.2.5-.5.5h-.4" fill="none" stroke="#000" stroke-width="1.1" stroke-linecap="round"/></svg>`))
	closeIcon    = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><path d="M2.6 2.6l4.8 4.8M7.4 2.6l-4.8 4.8" stroke="#000" stroke-width="1.1" stroke-linecap="round"/></svg>`))
	sidebarIcon  = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" width="12" height="12"><rect x="1.6" y="2.4" width="8.8" height="7.2" rx="1.8" fill="none" stroke="#000" stroke-width="1.1"/><path d="M4.9 2.4v7.2" stroke="#000" stroke-width="1.1"/></svg>`))
	ellipsisIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><circle cx="2.2" cy="5" r="0.9" fill="#000"/><circle cx="5" cy="5" r="0.9" fill="#000"/><circle cx="7.8" cy="5" r="0.9" fill="#000"/></svg>`))
)
