package app

// The rail: the project switcher, the history, and the folders of requests.

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *App) rail(c *ui.Context, t *ui.Theme) {
	project := a.Project()
	bar := c.TitleBar()

	// The shell always exists, so its width change is a transition the
	// framework can follow: 260 when open, 0 once closed. The content is kept
	// while it closes, so it slides away instead of vanishing, and dropped
	// once the motion is over, so nothing of it is left at the edge.
	width := float32(260)
	if a.sidebarHidden {
		width = 0
	}
	// No content once it is closing: a column keeps a minimum width for what
	// it holds, which is what left a strip of it at the edge. The shell alone
	// narrows, and the pane takes the room.
	showContent := !a.sidebarHidden

	rail := ui.Column(c).
		Key("rail").
		Width(width).
		MinWidth(0).
		MaxWidth(width).
		FillHeight().
		ClipX()

	if !a.sidebarHidden {
		rail.Padding(6).Gap(2).BorderWidth(0, 0, 0, 1).BorderColor(t.Border)
	} else {
		rail.Padding(0)
	}

	// Over the window's material the sidebar paints nothing, as a macOS
	// sidebar is translucent; without it, it takes the window's colour.
	if !c.Vibrancy() {
		rail.Background(t.Background)
	}

	rail.Transition(ui.ElementTransition{
		Size:     true,
		Duration: motionDuration(c),
		Ease:     ui.EaseOut,
	}).Children(func() {
		if !showContent {
			return
		}
		// The window controls sit over the top of the sidebar, and dragging
		// there moves the window, as in Finder.
		if bar.Bottom > 0 {
			ui.Box(c).Height(bar.Bottom).FillWidth().DragWindow()
		}

		ui.PopoverMenuButton(c, project.Name, func(m *ui.Menu) {
			for i := range a.doc.Projects {
				p := a.doc.Projects[i]
				if m.Item(p.Name).Radio(p.ID == a.doc.Active).Chosen() {
					a.SelectProject(p.ID)
				}
			}
			m.Separator()
			if m.Item("New Project…").Chosen() {
				a.openNewProject()
			}
			if m.Item("Project Variables…").Chosen() {
				a.varsOpen = true
			}
		}).FillWidth().Tooltip("Project")

		ui.SearchField(c, &a.filter).Label("Find a request").Tooltip("Filter the requests by name, URL or method")

		ui.Scroll(c).Grow(1).Padding(6, 0, 0, 0).Gap(2).Children(func() {
			if a.filter != "" {
				a.filterResults(c, t)
				return
			}
			action := "Clear"
			if len(a.doc.Recents) == 0 {
				action = ""
			}
			a.sectionTitle(c, t, "Recent", action, a.ClearRecents)
			for _, recent := range a.doc.Recents {
				a.recentRow(c, t, recent)
			}
			if len(a.doc.Recents) == 0 {
				ui.Text(c, "No calls yet.").FontSize(12).TextColor(t.TextMuted).Padding(4, 10)
			}

			a.sectionTitle(c, t, "Collections", "+", a.NewRequest)
			for fi := range project.Folders {
				a.folderView(c, t, project, fi)
			}

			if project.CountRequests() == 0 {
				ui.Text(c, "This project has no requests yet.").
					FontSize(12).
					TextColor(t.TextMuted).
					Padding(18, 10)
			}
		})
	})
}

func (a *App) sectionTitle(c *ui.Context, t *ui.Theme, title string, action string, run func() error) {
	ui.Row(c).Padding(6, 8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontSize(11.5).TextColor(t.TextMuted).Grow(1)
		if action == "" {
			return
		}

		// A source list uses borderless symbols, not bordered buttons.
		button := ui.ButtonBase(c).Padding(2, 5).Radius(5).Tooltip(tooltipFor(action))
		button.Children(func() {
			ui.Text(c, action).FontSize(11.5).TextColor(t.TextMuted)
		})
		if button.Hovered() {
			button.Background(a.tok.Fill)
		}
		if button.Clicked() {
			_ = run()
		}
	})
}

func tooltipFor(action string) string {
	if action == "+" {
		return "New request"
	}
	return action + " the history"
}

func (a *App) recentRow(c *ui.Context, t *ui.Theme, recent Recent) {
	row := ui.Row(c).
		Gap(8).
		Padding(5, 8).
		Radius(7).
		AlignItems(ui.Center)

	row.Children(func() {
		a.methodBadge(c, t, recent.Method)
		ui.Column(c).Gap(1).Grow(1).Children(func() {
			ui.Text(c, recent.URL).Font(mono).FontSize(11.5).SingleLine().Ellipsis("…")
			ui.Text(c, recent.At).FontSize(11).TextColor(t.TextMuted)
		})
	})

	if row.Hovered() {
		row.Background(t.Surface)
	}
	if row.Clicked() {
		a.draft.URL = recent.URL
		a.draft.Method = recent.Method
		a.touch(c.Now())
	}
}

func (a *App) folderView(c *ui.Context, t *ui.Theme, project Project, fi int) {
	folder := project.Folders[fi]
	open := folder.Open

	el := ui.Collapsible(c, folder.Name, &open, func() {
		for ri := range folder.Requests {
			a.requestRow(c, t, fi, ri, folder.Requests[ri])
		}
	})

	el.ContextMenu(func(m *ui.Menu) {
		if m.Item("Rename…").Chosen() {
			a.openRenameFolder(fi)
		}
		m.Separator()
		if m.Item("Delete…").Chosen() {
			a.openDeleteFolder(fi)
		}
	})

	if el.Changed() {
		a.SetFolderOpen(folder.ID, open)
	}
}

func (a *App) requestRow(c *ui.Context, t *ui.Theme, fi int, ri int, req Request) {
	selected := a.folder == fi && a.index == ri

	row := ui.Row(c).
		Gap(8).
		Padding(4, 8).
		PaddingStart(16).
		Radius(7).
		AlignItems(ui.Center)

	row.Children(func() {
		a.methodBadge(c, t, req.Method)
		ui.Text(c, displayName(req)).
			FontSize(12).
			SingleLine().
			Ellipsis("…").
			Grow(1)
	})

	if selected {
		row.Background(a.tok.Selection)
	} else if row.Hovered() {
		row.Background(a.tok.Fill)
	}

	row.ContextMenu(func(m *ui.Menu) {
		if m.Item("Rename…").Chosen() {
			a.openRenameRequest(fi, ri)
		}
		folders := a.Project().Folders
		if len(folders) > 1 {
			m.Submenu("Move to", func(sub *ui.Menu) {
				for i := range folders {
					if i == fi {
						continue
					}
					if sub.Item(folders[i].Name).Chosen() {
						a.moveRequest(fi, ri, i)
					}
				}
			})
		}
		m.Separator()
		if m.Item("Delete…").Chosen() {
			a.openDeleteRequest(fi, ri)
		}
	})

	if row.Clicked() {
		a.SelectRequest(fi, ri)
	}
}

// displayName is the request's name, or its path when it has none.
func displayName(req Request) string {
	if req.Name != "" {
		return req.Name
	}
	if req.URL != "" {
		return req.URL
	}
	return "Untitled"
}

// methodBadge is the method of a request: its colour is information, so the
// fill stays neutral and the text carries it.
func (a *App) methodBadge(c *ui.Context, t *ui.Theme, method string) {
	fg := a.methodColor(method)
	ui.Box(c).
		Width(42).
		Height(18).
		AlignItems(ui.Center).
		Justify(ui.Center).
		Children(func() {
			ui.Text(c, shortMethod(method)).Font(mono).FontSize(9.5).Bold().TextColor(fg)
		})
}

// methodColor is the colour of a method: success, warning and info carry
// the information, and a method with none of them stays muted.
func (a *App) methodColor(method string) ui.Color {
	switch method {
	case "GET":
		return a.tok.Success
	case "POST":
		return a.tok.Warning
	case "PUT", "PATCH":
		return a.tok.Info
	case "DELETE":
		return a.tok.Danger
	default:
		return a.tok.TextSubtle
	}
}

func shortMethod(method string) string {
	if method == "DELETE" {
		return "DEL"
	}
	return method
}

// collectionsHeader holds the one control that adds to the tree.
func (a *App) collectionsHeader(c *ui.Context, t *ui.Theme) {
	ui.Row(c).Padding(6, 8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "Collections").FontSize(11.5).TextColor(t.TextMuted).Grow(1)
		ui.PopoverMenuButton(c, "+", func(m *ui.Menu) {
			if m.Item("New Request").Chosen() {
				_ = a.NewRequest()
			}
			if m.Item("New Folder…").Chosen() {
				a.openNewFolder()
			}
		}).Tooltip("Add")
	})
}

// filterResults lists every request whose name, URL or method matches, across
// folders, keeping the indices the row needs to open it.
func (a *App) filterResults(c *ui.Context, t *ui.Theme) {
	project := a.Project()
	found := 0

	for fi := range project.Folders {
		for ri := range project.Folders[fi].Requests {
			request := project.Folders[fi].Requests[ri]
			if !matchesFilter(request, a.filter) {
				continue
			}
			found++
			a.requestRow(c, t, fi, ri, request)
		}
	}

	if found == 0 {
		ui.Text(c, "No request matches.").FontSize(12).TextColor(t.TextMuted).Padding(4, 10)
	}
}

func matchesFilter(request Request, filter string) bool {
	needle := strings.ToLower(strings.TrimSpace(filter))
	if needle == "" {
		return true
	}
	haystack := strings.ToLower(request.Name + " " + request.URL + " " + request.Method)
	return strings.Contains(haystack, needle)
}
