package app

// Editing the tree: renaming, deleting and moving what the rail shows, and
// the dialogs that ask for a name or a confirmation.

type dialogKind string

const (
	dialogNone          dialogKind = ""
	dialogRenameRequest dialogKind = "rename-request"
	dialogRenameFolder  dialogKind = "rename-folder"
	dialogNewFolder     dialogKind = "new-folder"
	dialogNewEnv        dialogKind = "new-environment"
	dialogNewProject    dialogKind = "new-project"
	dialogRenameProject dialogKind = "rename-project"
	dialogDeleteProject dialogKind = "delete-project"
	dialogRenameEnv     dialogKind = "rename-environment"
	dialogDeleteEnv     dialogKind = "delete-environment"
	dialogDeleteRequest dialogKind = "delete-request"
	dialogDeleteFolder  dialogKind = "delete-folder"
)

func (a *App) openRenameRequest(folder, index int) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) {
		return
	}
	requests := project.Folders[folder].Requests
	if index >= len(requests) {
		return
	}
	a.dialog, a.dialogFolder, a.dialogIndex = dialogRenameRequest, folder, index
	a.dialogValue = requests[index].Name
	a.dialogOpen = true
}

func (a *App) openRenameFolder(folder int) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) {
		return
	}
	a.dialog, a.dialogFolder, a.dialogIndex = dialogRenameFolder, folder, 0
	a.dialogValue = project.Folders[folder].Name
	a.dialogOpen = true
}

func (a *App) openNewFolder() {
	a.dialog, a.dialogValue = dialogNewFolder, "Folder"
	a.dialogOpen = true
}

func (a *App) openNewEnvironment() {
	a.dialog, a.dialogValue = dialogNewEnv, "Staging"
	a.dialogOpen = true
}

func (a *App) openNewProject() {
	a.dialog, a.dialogValue = dialogNewProject, "Project"
	a.dialogOpen = true
}

func (a *App) openRenameProject() {
	a.dialog, a.dialogValue = dialogRenameProject, a.Project().Name
	a.dialogOpen = true
}

func (a *App) openDeleteProject() {
	a.dialog = dialogDeleteProject
	a.dialogOpen = true
}

func (a *App) openRenameEnvironment() {
	environment := a.Project().ActiveEnvironment()
	if environment == nil {
		return
	}
	a.dialog, a.dialogValue = dialogRenameEnv, environment.Name
	a.dialogOpen = true
}

func (a *App) openDeleteEnvironment() {
	if a.Project().ActiveEnvironment() == nil {
		return
	}
	a.dialog = dialogDeleteEnv
	a.dialogOpen = true
}

func (a *App) openDeleteRequest(folder, index int) {
	a.dialog, a.dialogFolder, a.dialogIndex = dialogDeleteRequest, folder, index
	a.dialogOpen = true
}

func (a *App) openDeleteFolder(folder int) {
	a.dialog, a.dialogFolder, a.dialogIndex = dialogDeleteFolder, folder, 0
	a.dialogOpen = true
}

// confirm applies whatever the open dialog asked about. It is the only place
// the tree changes shape, so every path saves the document once.
func (a *App) confirm() {
	name := trimSpace(a.dialogValue)
	switch a.dialog {
	case dialogRenameRequest:
		a.renameRequest(a.dialogFolder, a.dialogIndex, name)
	case dialogRenameFolder:
		a.renameFolder(a.dialogFolder, name)
	case dialogNewFolder:
		a.addFolder(name)
	case dialogNewEnv:
		a.AddEnvironment(name)
	case dialogNewProject:
		a.AddProject(name)
	case dialogRenameProject:
		a.RenameProject(name)
	case dialogDeleteProject:
		a.DeleteProject()
	case dialogRenameEnv:
		a.RenameEnvironment(name)
	case dialogDeleteEnv:
		a.DeleteEnvironment()
	case dialogDeleteRequest:
		a.deleteRequest(a.dialogFolder, a.dialogIndex)
	case dialogDeleteFolder:
		a.deleteFolder(a.dialogFolder)
	}
	a.dialog = dialogNone
	a.dialogOpen = false
}

func (a *App) renameRequest(folder, index int, name string) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) || name == "" {
		return
	}
	requests := project.Folders[folder].Requests
	if index >= len(requests) {
		return
	}
	requests[index].Name = name
	if a.folder == folder && a.index == index {
		a.draft.Name = name
	}
	a.save()
}

func (a *App) renameFolder(folder int, name string) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) || name == "" {
		return
	}
	project.Folders[folder].Name = name
	a.save()
}

func (a *App) addFolder(name string) {
	project := a.doc.ActiveProject()
	if project == nil || name == "" {
		return
	}
	project.Folders = append(project.Folders, Folder{ID: newID("folder"), Name: name, Open: true})
	a.save()
}

// deleteRequest removes a request and keeps the selection on something real.
func (a *App) deleteRequest(folder, index int) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) {
		return
	}
	requests := project.Folders[folder].Requests
	if index >= len(requests) {
		return
	}

	id := a.draft.ID
	a.detach()
	project.Folders[folder].Requests = append(requests[:index], requests[index+1:]...)
	a.save()
	a.relocate(id)
	a.loadSelection()
}

// deleteFolder removes a folder with everything in it.
func (a *App) deleteFolder(folder int) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) {
		return
	}

	id := a.draft.ID
	a.detach()
	project.Folders = append(project.Folders[:folder], project.Folders[folder+1:]...)
	a.save()
	a.relocate(id)
	a.loadSelection()
}

// moveRequest puts a request in another folder and selects it there.
func (a *App) moveRequest(folder, index, to int) {
	project := a.doc.ActiveProject()
	if project == nil || folder >= len(project.Folders) || to >= len(project.Folders) || folder == to {
		return
	}
	requests := project.Folders[folder].Requests
	if index >= len(requests) {
		return
	}

	moved := requests[index]
	a.detach()
	project.Folders[folder].Requests = append(requests[:index], requests[index+1:]...)

	target := &project.Folders[to]
	target.Requests = append(target.Requests, moved)
	target.Open = true
	a.save()
	a.relocate(moved.ID)
	a.loadSelection()
}

// trimSpace keeps a name tidy without letting it be blank.
func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
