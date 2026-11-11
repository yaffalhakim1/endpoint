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
