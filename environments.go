package main

// Environments: named sets of variables, one active per project. The active
// one wins over the project's own variables, so a project can hold what every
// environment shares and each environment what only it knows.

import "strings"

// ActiveEnvironment returns the environment in use, or nil when none is.
func (p Project) ActiveEnvironment() *Environment {
	if p.ActiveEnv == "" {
		return nil
	}
	for i := range p.Environments {
		if p.Environments[i].ID == p.ActiveEnv {
			return &p.Environments[i]
		}
	}
	return nil
}

// Variables are the values a request is resolved against: the project's, with
// the active environment's on top.
func (p Project) Variables() []Pair {
	active := p.ActiveEnvironment()
	if active == nil {
		return p.Vars
	}

	out := make([]Pair, 0, len(p.Vars)+len(active.Vars))
	out = append(out, p.Vars...)
	for _, pair := range active.Vars {
		replaced := false
		for i := range out {
			if out[i].Key == pair.Key {
				out[i].Value = pair.Value
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, pair)
		}
	}
	return out
}

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
