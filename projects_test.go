package main

import "testing"

// A new project is reachable from the switcher menu, starts with one open
// folder, becomes active, and stays chosen after a reload.
func TestAddProject(t *testing.T) {
	app := newTestApp(t)
	before := len(app.doc.Projects)

	project := app.AddProject("  Payments  ")
	if project == nil {
		t.Fatal("the project was not created")
	}
	if project.Name != "Payments" {
		t.Errorf("the name is %q, want it trimmed", project.Name)
	}
	if got := len(app.doc.Projects); got != before+1 {
		t.Errorf("%d projects, want %d", got, before+1)
	}
	if app.doc.Active != project.ID {
		t.Error("the new project is not active")
	}
	if len(app.Project().Folders) != 1 {
		t.Fatalf("a new project has %d folders, want 1", len(app.Project().Folders))
	}

	reloaded, err := NewApp(app.dir)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Project().Name != "Payments" {
		t.Errorf("after reloading the project is %q", reloaded.Project().Name)
	}
}

func TestBlankProjectNameIsIgnored(t *testing.T) {
	app := newTestApp(t)
	before := len(app.doc.Projects)

	if app.AddProject("   ") != nil {
		t.Error("a blank name created a project")
	}
	if got := len(app.doc.Projects); got != before {
		t.Errorf("%d projects after a blank name, want %d", got, before)
	}
}
