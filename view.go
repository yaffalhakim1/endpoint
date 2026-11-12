package main

// The window: a full-height sidebar with the window controls over it, and the
// content column, whose title bar names what the window shows.

import (
	"time"

	"github.com/egoist/mygo/ui"
)

func (a *App) view(c *ui.Context) {
	t := a.appTheme(c)

	if a.dirty && !a.saveIfDue(c.Now()) {
		c.After(400 * time.Millisecond)
	}

	// A workspace that had to be recovered says so, once, in the window.
	if a.recovered.Message != "" {
		message := a.recovered.Message
		if a.recovered.Saved != "" {
			message += "\nCopy it somewhere safe, then delete it."
		}
		a.recovered.Message = "" // said once, not on every frame
		go func() {
			if a.win != nil {
				a.win.Update(func() {
					a.alert = Alert{Open: true, Title: "The workspace was recovered", Body: message}
				})
			}
		}()
	}

	// A command from the menu bar that needs the view runs here.
	if a.copyRequested {
		a.copyRequested = false
		c.WriteClipboard(a.copyText())
		c.Toast("Copied")
	}

	// No background here: the window's material shows through the sidebar,
	// as a vibrant sidebar does on macOS.
	ui.Themed(c, t).Fill().Children(func() {
		ui.Row(c).Grow(1).Children(func() {
			// The rail is always built: its shell animates open and closed,
			// and drops its own content once the motion is over.
			a.rail(c, t)
			a.content(c, t)
		})
		a.dialogs(c, t)
		a.varsDialog(c, t)
		a.alertDialog(c, t)
	})
}

func (a *App) content(c *ui.Context, t *ui.Theme) {
	// The rail takes no room the moment it goes, so the pane animates its own
	// way into the space it leaves; without this it jumps at once.
	content := ui.Column(c).Grow(1).FillHeight().Background(t.Surface)
	content.Transition(ui.ElementTransition{
		Position: true,
		Size:     true,
		Duration: motionDuration(c),
		Ease:     ui.EaseOut,
	})
	content.Children(func() {
		a.titleBar(c, t)
		a.requestBar(c, t)
		ui.Row(c).
			Height(38).
			PaddingX(8).
			AlignItems(ui.Center).
			BorderWidth(0, 0, 1, 0).
			BorderColor(t.Border).
			Children(func() {
				ui.Tabs(c, &a.tab, "Params", "Headers", "Auth", "Body")
			})
		// A divider the user can move: the request pane takes the room the
		// answer does not need.
		ui.SplitVertical(c, &a.requestHeight, func() {
			a.panel(c, t)
		}, func() {
			a.response(c, t)
		}).Grow(1).SplitReset(200)
	})
}

// titleBar names the request the window shows. A window title says where you
// are, so it is never the app's own name.
func (a *App) titleBar(c *ui.Context, t *ui.Theme) {
	bar := c.TitleBar()
	height := bar.Height
	if height < 44 {
		height = 44
	}

	ui.Row(c).
		Height(height).
		PaddingStart(10).
		PaddingEnd(8).
		Gap(8).
		AlignItems(ui.Center).
		DragWindow().
		Children(func() {
			a.sidebarToggle(c, t)
			ui.Text(c, a.title()).FontSize(13).Bold().SingleLine().Ellipsis("…")
			a.environmentButton(c, t)
			ui.Spacer(c)
			a.appMenu(c, t)
			a.windowButtons(c, t)
		})
}

// environmentButton names the active environment and switches it.
func (a *App) environmentButton(c *ui.Context, t *ui.Theme) {
	ui.PopoverMenuButton(c, a.environmentName(), func(m *ui.Menu) {
		project := a.Project()
		for i := range project.Environments {
			environment := project.Environments[i]
			if m.Item(environment.Name).Radio(project.ActiveEnv == environment.ID).Chosen() {
				a.UseEnvironment(environment.ID)
			}
		}
		if len(project.Environments) > 0 {
			m.Separator()
		}
		if m.Item("New Environment…").Chosen() {
			a.openNewEnvironment()
		}
		if m.Item("Edit Variables…").Chosen() {
			a.varsOpen = true
		}
	}).Tooltip("Environment")
}

// sidebarToggle shows and hides the sidebar, as the leading button of a macOS
// window does.
func (a *App) sidebarToggle(c *ui.Context, t *ui.Theme) {
	label := "Hide Sidebar"
	if a.sidebarHidden {
		label = "Show Sidebar"
	}

	button := ui.ButtonBase(c).Width(28).Height(26).Radius(6).Label(label).Tooltip(label)
	hovered := button.Hovered()
	if hovered {
		button.Background(a.tok.Fill)
	}
	button.Children(func() {
		ui.Icon(c, sidebarIcon).FontSize(12).TextColor(t.TextMuted)
	})
	if button.Clicked() {
		a.sidebarHidden = !a.sidebarHidden
	}
}

// appMenu holds the commands that are not about the window itself. It sits
// apart from the sidebar toggle, next to the window controls.
func (a *App) appMenu(c *ui.Context, t *ui.Theme) {
	button := ui.ButtonBase(c).Width(28).Height(26).Radius(6).Label("Menu").Tooltip("Menu")
	hovered := button.Hovered()
	if hovered {
		button.Background(a.tok.Fill)
	}
	button.Children(func() {
		ui.Icon(c, ellipsisIcon).FontSize(12).TextColor(t.TextMuted)
	})

	button.PopoverMenuButton(func(m *ui.Menu) {
		m.Submenu("File", func(m *ui.Menu) {
			if m.Item("New Request").Chosen() {
				_ = a.NewRequest()
			}
			if m.Item("New Project…").Chosen() {
				a.openNewProject()
			}
			if m.Item("Rename Project…").Chosen() {
				a.openRenameProject()
			}
			if m.Item("Delete Project…").Chosen() {
				a.openDeleteProject()
			}
			if m.Item("New Folder…").Chosen() {
				a.openNewFolder()
			}
			if m.Item("Project Variables…").Chosen() {
				a.varsOpen = true
			}
			m.Separator()
			if m.Item("About Endpoint").Chosen() {
				a.showAbout()
			}
		})
		m.Submenu("Request", func(m *ui.Menu) {
			if m.Item("Send").Chosen() {
				a.startSend()
			}
			if m.Item("Cancel Request").Disabled(a.state != StateLoading).Chosen() {
				a.Cancel()
			}
			if m.Item("Copy Response").Chosen() {
				a.copyRequested = true
			}
		})
		m.Submenu("View", func(m *ui.Menu) {
			if m.Item("Show Sidebar").Checked(!a.sidebarHidden).Chosen() {
				a.sidebarHidden = !a.sidebarHidden
			}
			m.Separator()
			for _, choice := range []struct {
				label string
				mode  int
			}{{"Follow the System", themeAuto}, {"Light", themeLight}, {"Dark", themeDark}} {
				choice := choice
				if m.Item(choice.label).Radio(a.themeMode == choice.mode).Chosen() {
					a.themeMode = choice.mode
				}
			}
		})
	})
}

// windowButtons draws minimize, maximize and close: the window controls a
// title bar drawn by the view has to provide itself.
func (a *App) windowButtons(c *ui.Context, t *ui.Theme) {
	minimize := ui.ButtonBase(c).Width(34).Height(26).Radius(6).Label("Minimize").Tooltip("Minimize")
	minimize.Children(func() {
		ui.Icon(c, minimizeIcon).FontSize(10).TextColor(t.TextMuted)
	})
	if minimize.Hovered() {
		minimize.Background(a.tok.Fill)
	}
	if minimize.Clicked() && a.win != nil {
		a.win.Minimize()
	}

	maximize := ui.ButtonBase(c).Width(34).Height(26).Radius(6).Label("Maximize").Tooltip("Maximize")
	maximize.Children(func() {
		icon := maximizeIcon
		if a.win != nil && a.win.IsMaximized() {
			icon = restoreIcon
		}
		ui.Icon(c, icon).FontSize(10).TextColor(t.TextMuted)
	})
	if maximize.Hovered() {
		maximize.Background(a.tok.Fill)
	}
	if maximize.Clicked() && a.win != nil {
		if a.win.IsMaximized() {
			a.win.Unmaximize()
		} else {
			a.win.Maximize()
		}
	}

	closing := ui.ButtonBase(c).Width(34).Height(26).Radius(6).Label("Close").Tooltip("Close")
	closeHovered := closing.Hovered()
	closeColor := t.TextMuted
	if closeHovered {
		closing.Background(ui.Hex("#e81123"))
		closeColor = ui.Hex("#ffffff")
	}
	closing.Children(func() {
		ui.Icon(c, closeIcon).FontSize(10).TextColor(closeColor)
	})
	if closing.Clicked() && a.win != nil {
		a.win.Close()
	}
}

// title is the request the window edits, or its project when none is loaded.
func (a *App) title() string {
	if a.draft.Name != "" {
		return a.draft.Name
	}
	if a.draft.URL != "" {
		return a.draft.URL
	}
	return a.Project().Name
}

// kbd draws one key cap, for the panels that show a shortcut.
func kbd(c *ui.Context, t *ui.Theme, label string) {
	ui.Box(c).
		Padding(3, 5).
		Radius(5).
		Background(t.Surface).
		Border(1, t.Border).
		Children(func() {
			ui.Text(c, label).Font(mono).FontSize(10.5).TextColor(t.TextMuted)
		})
}
