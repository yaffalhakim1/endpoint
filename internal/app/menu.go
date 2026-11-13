package app

// The menu bar. Every command lives here, with its shortcut shown, so nothing
// is reachable only through a gesture the user has to discover.

import "github.com/egoist/mygo"

func (a *App) installMenu() {
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "File", Submenu: []*mygo.MenuItem{
			{Label: "New Request", Accelerator: "CmdOrCtrl+N", Click: a.onMenu(func() {
				_ = a.NewRequest()
			})},
			{Label: "New Folder…", Accelerator: "CmdOrCtrl+Shift+N", Click: a.onMenu(a.openNewFolder)},
			{Label: "New Project…", Accelerator: "CmdOrCtrl+Shift+P", Click: a.onMenu(a.openNewProject)},
			{Label: "Project Variables…", Accelerator: "CmdOrCtrl+E", Click: a.onMenu(func() {
				a.varsOpen = true
			})},
			{Label: "HTTP Options…", Accelerator: "CmdOrCtrl+,", Click: a.onMenu(func() {
				a.optionsOpen = true
			})},
			{Type: mygo.MenuItemSeparator},
			{Label: "Import from curl…", Accelerator: "CmdOrCtrl+I", Click: a.onMenu(func() {
				a.importOpen = true
			})},
			{Label: "Export workspace…", Click: a.onMenu(a.exportWorkspace)},
			{Type: mygo.MenuItemSeparator},
			{Label: "Rename…", Accelerator: "CmdOrCtrl+R", Click: a.onMenu(func() {
				a.openRenameRequest(a.folder, a.index)
			})},
			{Label: "Delete Request…", Accelerator: "CmdOrCtrl+Backspace", Click: a.onMenu(func() {
				a.openDeleteRequest(a.folder, a.index)
			})},
			{Label: "Undo Delete", Accelerator: "CmdOrCtrl+Shift+Z", Click: a.onMenu(func() {
				a.undoLast()
			})},
			{Type: mygo.MenuItemSeparator},
			{Role: mygo.RoleClose},
		}},
		{Label: "Edit", Submenu: []*mygo.MenuItem{
			{Role: mygo.RoleUndo},
			{Role: mygo.RoleRedo},
			{Type: mygo.MenuItemSeparator},
			{Role: mygo.RoleCut},
			{Role: mygo.RoleCopy},
			{Role: mygo.RolePaste},
			{Role: mygo.RoleSelectAll},
		}},
		{Label: "Request", Submenu: []*mygo.MenuItem{
			{Label: "Send", Accelerator: "CmdOrCtrl+Enter", Click: a.onMenu(a.startSend)},
			{Label: "Cancel Request", Accelerator: "CmdOrCtrl+.", Click: a.onMenu(a.Cancel)},
			{Label: "Copy Response", Accelerator: "CmdOrCtrl+Shift+C", Click: a.onMenu(func() {
				a.copyRequested = true
			})},
			{Label: "Save Response…", Accelerator: "CmdOrCtrl+S", Click: a.onMenu(a.saveResponse)},
			{Label: "Copy as cURL", Click: a.onMenu(func() {
				a.curlRequested = true
			})},
		}},
		{Label: "View", Submenu: []*mygo.MenuItem{
			{Label: "Show Sidebar", Type: mygo.MenuItemCheckbox, Checked: !a.sidebarHidden, Click: a.onMenu(func() {
				a.sidebarHidden = !a.sidebarHidden
			})},
			{Type: mygo.MenuItemSeparator},
			{Label: "Appearance", Submenu: []*mygo.MenuItem{
				{Label: "Follow the System", Type: mygo.MenuItemRadio, Checked: a.themeMode == themeAuto, Click: a.onMenu(func() {
					a.themeMode = themeAuto
				})},
				{Label: "Light", Type: mygo.MenuItemRadio, Checked: a.themeMode == themeLight, Click: a.onMenu(func() {
					a.themeMode = themeLight
				})},
				{Label: "Dark", Type: mygo.MenuItemRadio, Checked: a.themeMode == themeDark, Click: a.onMenu(func() {
					a.themeMode = themeDark
				})},
			}},
			{Type: mygo.MenuItemSeparator},
			{Role: mygo.RoleToggleFullScreen},
		}},
	})

	mygo.App.SetMenu(menu)
}

// onMenu wraps a command so the window redraws once it ran on the main thread.
func (a *App) onMenu(run func()) func(*mygo.MenuItem, *mygo.Window) {
	return func(*mygo.MenuItem, *mygo.Window) {
		run()
		if a.win != nil {
			a.win.Invalidate()
		}
	}
}
