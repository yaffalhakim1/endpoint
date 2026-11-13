package app

import "testing"

// A delete must be reversible: the app keeps the document as it was, and one
// step puts it back.

func TestUndoRestoresADeletedRequest(t *testing.T) {
	app := newTestApp(t)
	before := app.Project().CountRequests()

	app.openDeleteRequest(0, 0)
	app.confirm()

	if got := app.Project().CountRequests(); got != before-1 {
		t.Fatalf("%d requests after the delete, want %d", got, before-1)
	}
	if !app.canUndo() {
		t.Fatal("there is nothing to undo")
	}

	if !app.undoLast() {
		t.Fatal("undo reported nothing to do")
	}
	if got := app.Project().CountRequests(); got != before {
		t.Errorf("%d requests after undo, want %d", got, before)
	}
	if app.draft.Name != "Create user" {
		t.Errorf("the selection is %q after undo", app.draft.Name)
	}
}

func TestUndoRestoresADeletedFolder(t *testing.T) {
	app := newTestApp(t)
	before := len(app.Project().Folders)

	app.openDeleteFolder(1)
	app.confirm()
	if got := len(app.Project().Folders); got != before-1 {
		t.Fatalf("%d folders after the delete", got)
	}

	app.undoLast()
	if got := len(app.Project().Folders); got != before {
		t.Errorf("%d folders after undo, want %d", got, before)
	}
	if app.Project().Folders[1].Name != "Billing" {
		t.Errorf("the folder came back as %q", app.Project().Folders[1].Name)
	}
}

func TestUndoRestoresADeletedProject(t *testing.T) {
	app := newTestApp(t)
	before := len(app.doc.Projects)

	app.openDeleteProject()
	app.confirm()
	if got := len(app.doc.Projects); got != before-1 {
		t.Fatalf("%d projects after the delete", got)
	}

	app.undoLast()
	if got := len(app.doc.Projects); got != before {
		t.Errorf("%d projects after undo, want %d", got, before)
	}
	if app.Project().Name != "Personal" {
		t.Errorf("the active project is %q after undo", app.Project().Name)
	}
}

func TestUndoIsBounded(t *testing.T) {
	app := newTestApp(t)

	for i := 0; i < undoDepth+5; i++ {
		app.snapshot()
	}
	if got := len(app.undo); got != undoDepth {
		t.Errorf("the stack holds %d, want %d", got, undoDepth)
	}
}

func TestUndoWithNothingToUndo(t *testing.T) {
	app := newTestApp(t)
	if app.canUndo() {
		t.Fatal("a fresh app has something to undo")
	}
	if app.undoLast() {
		t.Error("undo reported success with an empty stack")
	}
}

func TestUndoSurvivesAReload(t *testing.T) {
	app := newTestApp(t)
	app.openDeleteRequest(0, 0)
	app.confirm()
	app.undoLast()

	reloaded, err := New(app.dir)
	if err != nil {
		t.Fatal(err)
	}
	// The seed holds six requests across its three folders.
	if got := reloaded.Project().CountRequests(); got != 6 {
		t.Errorf("after reloading there are %d requests, want 6", got)
	}
}
