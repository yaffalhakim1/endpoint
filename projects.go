package main

// Projects: the top level of the rail. One is active, and a new one starts
// with an open folder so it is not a dead end.

import "strings"

// AddProject creates one, makes it active and leaves the rail on it.
func (a *App) AddProject(name string) *Project {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	project := Project{
		ID:      newID("project"),
		Name:    name,
		Folders: []Folder{{ID: newID("folder"), Name: "Requests", Open: true}},
	}

	a.doc.Projects = append(a.doc.Projects, project)
	a.doc.Active = project.ID
	a.selectFirst()
	a.state = StateEmpty
	a.bodyFor = ""
	a.save()

	return &a.doc.Projects[len(a.doc.Projects)-1]
}

// ProjectNames lists what the switcher offers.
func (a *App) ProjectNames() []string {
	names := make([]string, 0, len(a.doc.Projects))
	for _, project := range a.doc.Projects {
		names = append(names, project.Name)
	}
	return names
}

// RenameProject renames the active project.
func (a *App) RenameProject(name string) {
	project := a.doc.ActiveProject()
	name = strings.TrimSpace(name)
	if project == nil || name == "" {
		return
	}
	project.Name = name
	a.save()
}

// DeleteProject removes the active project with everything in it, and leaves
// the rail on whatever project is left.
func (a *App) DeleteProject() {
	project := a.doc.ActiveProject()
	if project == nil || len(a.doc.Projects) < 2 {
		return // the last project stays: the app always has one
	}

	kept := make([]Project, 0, len(a.doc.Projects)-1)
	for _, other := range a.doc.Projects {
		if other.ID != project.ID {
			kept = append(kept, other)
		}
	}
	a.doc.Projects = kept
	a.doc.Active = kept[0].ID
	a.detach()
	a.selectFirst()
	a.state = StateEmpty
	a.save()
}
