package main

import "testing"

// Renaming and deleting the top level: projects and environments now have
// what folders always had.

func TestRenameProject(t *testing.T) {
	app := newTestApp(t)

	app.openRenameProject()
	if app.dialogValue != "Personal" {
		t.Fatalf("the dialog shows %q", app.dialogValue)
	}
	app.dialogValue = "  Acme  "
	app.confirm()

	if got := app.Project().Name; got != "Acme" {
		t.Errorf("the project is %q", got)
	}
}

func TestDeleteProjectLeavesOneBehind(t *testing.T) {
	app := newTestApp(t)

	app.openDeleteProject()
	if msg := app.deleteMessage(); msg == "" {
		t.Error("the confirmation says nothing")
	}
	app.confirm()

	if got := len(app.doc.Projects); got != 2 {
		t.Fatalf("%d projects left, want 2", got)
	}
	if app.Project().Name == "Personal" {
		t.Error("the deleted project is still active")
	}
	if len(app.doc.Projects) == 1 {
		app.openDeleteProject()
		app.confirm()
		if len(app.doc.Projects) != 1 {
			t.Error("the last project was deleted, leaving none")
		}
	}
}

func TestRenameAndDeleteEnvironment(t *testing.T) {
	app := newTestApp(t)
	app.AddEnvironment("Staging")

	app.openRenameEnvironment()
	app.dialogValue = "Pre-production"
	app.confirm()
	if got := app.environmentName(); got != "Pre-production" {
		t.Errorf("the environment is %q", got)
	}

	app.openDeleteEnvironment()
	if msg := app.deleteMessage(); msg == "" {
		t.Error("the confirmation says nothing")
	}
	app.confirm()

	if got := app.environmentName(); got != "No environment" {
		t.Errorf("after deleting it the environment is %q", got)
	}
	if len(app.Project().Environments) != 0 {
		t.Error("the environment survived")
	}
}
