package main

// The dialogs that name and delete what the rail shows.

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

func (a *App) dialogs(c *ui.Context, t *ui.Theme) {
	if !a.dialogOpen {
		a.dialog = dialogNone
		return
	}

	ui.Modal(c, &a.dialogOpen, func() {
		title, primary := dialogLabels(a.dialog)
		ui.Column(c).Gap(10).Children(func() {
			ui.Text(c, title).FontSize(13).Bold()

			switch a.dialog {
			case dialogDeleteRequest, dialogDeleteFolder:
				ui.Text(c, a.deleteMessage()).FontSize(12.5).TextColor(t.TextMuted)
			default:
				input := ui.TextInput(c, &a.dialogValue).AutoFocus().Label("Name")
				if input.Submitted() {
					a.confirm()
				}
			}

			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.dialogOpen = false
				}
				if ui.PrimaryButton(c, primary).Clicked() {
					a.confirm()
				}
			})
		})
	})

	// Escape or a click outside closed it: forget what it was about.
	if !a.dialogOpen {
		a.dialog = dialogNone
	}
}

func dialogLabels(kind dialogKind) (title string, primary string) {
	switch kind {
	case dialogRenameRequest:
		return "Rename request", "Rename"
	case dialogRenameFolder:
		return "Rename folder", "Rename"
	case dialogNewFolder:
		return "New folder", "Create"
	case dialogNewEnv:
		return "New environment", "Create"
	case dialogNewProject:
		return "New project", "Create"
	case dialogDeleteRequest:
		return "Delete request?", "Delete"
	case dialogDeleteFolder:
		return "Delete folder?", "Delete"
	}
	return "", "OK"
}

// deleteMessage names what goes, so a confirmation is never anonymous.
func (a *App) deleteMessage() string {
	project := a.doc.ActiveProject()
	if project == nil {
		return ""
	}

	switch a.dialog {
	case dialogDeleteRequest:
		if a.dialogFolder < len(project.Folders) {
			requests := project.Folders[a.dialogFolder].Requests
			if a.dialogIndex < len(requests) {
				return "“" + displayName(requests[a.dialogIndex]) + "” is removed for good. This cannot be undone."
			}
		}
	case dialogDeleteFolder:
		if a.dialogFolder < len(project.Folders) {
			folder := project.Folders[a.dialogFolder]
			return "“" + folder.Name + "” and its " + plural(len(folder.Requests), "request") +
				" are removed for good. This cannot be undone."
		}
	}
	return ""
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

// varsDialog edits the variables of the active project.
func (a *App) varsDialog(c *ui.Context, t *ui.Theme) {
	if !a.varsOpen {
		return
	}

	ui.Modal(c, &a.varsOpen, func() {
		ui.Column(c).Width(520).Gap(10).Children(func() {
			ui.Text(c, a.varsTitle()).FontSize(13).Bold()
			ui.Text(c, "Write {{name}} in a URL, a header, an auth field or a body. A name with no value here is sent as written.").
				FontSize(12).
				TextColor(t.TextMuted)

			if vars := a.EditVarsTarget(); vars != nil {
				a.pairEditor(c, t, vars, "value")
			}

			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.PrimaryButton(c, "Done").Clicked() {
					a.varsOpen = false
				}
			})
		})
	})

	// Closed with Escape or a click outside: keep what was typed anyway.
	if !a.varsOpen {
		a.save()
	}
}
