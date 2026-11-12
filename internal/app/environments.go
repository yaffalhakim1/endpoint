package app

// Environments: named sets of variables, one active per project.

import (
	"strings"
)

// AddEnvironment creates one, makes it active and returns it.
func (a *App) AddEnvironment(name string) *Environment {
	project := a.doc.ActiveProject()
	if project == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	project.Environments = append(project.Environments, Environment{
		ID:   newID("env"),
		Name: name,
	})
	environment := &project.Environments[len(project.Environments)-1]
	project.ActiveEnv = environment.ID
	a.save()
	return environment
}

// UseEnvironment makes one active, and saves the choice.
func (a *App) UseEnvironment(id string) {
	project := a.doc.ActiveProject()
	if project == nil {
		return
	}
	project.ActiveEnv = id
	a.save()
}

// EditVarsTarget is the list the variables dialog shows: the active
// environment's, or the project's when no environment is active.
func (a *App) EditVarsTarget() *[]Pair {
	project := a.doc.ActiveProject()
	if project == nil {
		return nil
	}
	if active := project.ActiveEnvironment(); active != nil {
		if active.Vars == nil {
			active.Vars = []Pair{}
		}
		return &active.Vars
	}
	if project.Vars == nil {
		project.Vars = []Pair{}
	}
	return &project.Vars
}

// environmentName is what the title bar shows.
func (a *App) environmentName() string {
	project := a.Project()
	if active := project.ActiveEnvironment(); active != nil {
		return active.Name
	}
	return "No environment"
}

// varsTitle names what the dialog edits, so it is never ambiguous.
func (a *App) varsTitle() string {
	project := a.Project()
	if active := project.ActiveEnvironment(); active != nil {
		return "Variables of " + active.Name
	}
	return "Project variables"
}

// RenameEnvironment renames the active one.
func (a *App) RenameEnvironment(name string) {
	project := a.doc.ActiveProject()
	active := project.ActiveEnvironment()
	name = strings.TrimSpace(name)
	if project == nil || active == nil || name == "" {
		return
	}
	active.Name = name
	a.save()
}

// DeleteEnvironment removes the active one and leaves the project with none.
func (a *App) DeleteEnvironment() {
	project := a.doc.ActiveProject()
	active := project.ActiveEnvironment()
	if project == nil || active == nil {
		return
	}

	kept := make([]Environment, 0, len(project.Environments)-1)
	for _, environment := range project.Environments {
		if environment.ID != active.ID {
			kept = append(kept, environment)
		}
	}
	project.Environments = kept
	project.ActiveEnv = ""
	if len(kept) > 0 {
		project.ActiveEnv = kept[0].ID
	}
	a.save()
}
