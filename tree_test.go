package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Editing the tree must change the document, the disk, and the selection.

func TestNewFolderAndMoveRequest(t *testing.T) {
	app := newTestApp(t)

	app.addFolder("Archive")
	project := app.Project()
	if got := len(project.Folders); got != 4 {
		t.Fatalf("%d folders after adding one, want 4", got)
	}

	// Get user is the second request of the first folder.
	if err := app.doc.ActiveProject(); err != nil {
		_ = err
	}
	app.moveRequest(0, 1, 3)

	moved := app.Project().Folders[3]
	if len(moved.Requests) != 1 || moved.Requests[0].Name != "Get user" {
		t.Fatalf("the archive holds %+v", moved.Requests)
	}
	if app.draft.Name != "Get user" {
		t.Errorf("the selection is %q after the move", app.draft.Name)
	}

	// And it is still there after a reload.
	reloaded, err := NewApp(app.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(reloaded.doc.Projects[0].Folders); got != 4 {
		t.Errorf("%d folders after reloading, want 4", got)
	}
}

func TestRenameRequestThroughTheDialog(t *testing.T) {
	app := newTestApp(t)

	app.openRenameRequest(0, 1)
	if !app.dialogOpen || app.dialogValue != "Get user" {
		t.Fatalf("the dialog is %v with %q", app.dialogOpen, app.dialogValue)
	}

	app.dialogValue = "  Fetch user  "
	app.confirm()

	if app.dialogOpen {
		t.Error("the dialog stayed open")
	}
	if got := app.Project().Folders[0].Requests[1].Name; got != "Fetch user" {
		t.Errorf("the request is %q, want Fetch user", got)
	}
}

func TestDeleteRequestKeepsARealSelection(t *testing.T) {
	app := newTestApp(t)

	// The first request is selected at start.
	app.openDeleteRequest(0, 0)
	app.confirm()

	project := app.Project()
	if got := len(project.Folders[0].Requests); got != 2 {
		t.Fatalf("%d requests left, want 2", got)
	}
	if app.draft.Name != "Get user" {
		t.Errorf("the selection is %q, want Get user", app.draft.Name)
	}
}

func TestDeleteFolderTakesItsRequests(t *testing.T) {
	app := newTestApp(t)

	app.openDeleteFolder(1)
	if msg := app.deleteMessage(); !strings.Contains(msg, "Billing") {
		t.Errorf("the confirmation reads %q", msg)
	}
	app.confirm()

	if got := len(app.Project().Folders); got != 2 {
		t.Fatalf("%d folders left, want 2", got)
	}
	for _, folder := range app.Project().Folders {
		if folder.Name == "Billing" {
			t.Fatal("the Billing folder survived")
		}
	}
}

func TestABlankNameIsIgnored(t *testing.T) {
	app := newTestApp(t)

	app.openRenameFolder(0)
	app.dialogValue = "   "
	app.confirm()

	if got := app.Project().Folders[0].Name; got != "Users" {
		t.Errorf("the folder is now %q, want it unchanged", got)
	}
}

// The response pane must be able to show what the server said.

func TestSendReturnsHeadersAndRendersThem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	app := newTestApp(t)
	app.result = sendForTest(Request{Method: "GET", URL: server.URL}, Auth{}, server.Client())
	if app.result.Err != "" {
		t.Fatalf("the call failed: %s", app.result.Err)
	}

	names := []string{}
	for _, header := range app.result.Headers {
		names = append(names, header.Key)
	}
	if len(names) < 2 {
		t.Fatalf("got %d headers: %v", len(names), names)
	}

	app.responseTab = 1
	text := app.shownText()
	for _, want := range []string{"X-Ratelimit-Remaining: 42", "Content-Type: application/json"} {
		if !strings.Contains(text, want) {
			t.Errorf("the header pane lacks %q:\n%s", want, text)
		}
	}

	// Sorted, so the list reads the same way every time.
	if names[0] > names[len(names)-1] {
		t.Errorf("headers are not sorted: %v", names)
	}

	app.responseTab = 0
	if !strings.Contains(app.shownText(), `"ok": true`) {
		t.Errorf("the body pane shows:\n%s", app.shownText())
	}
}
