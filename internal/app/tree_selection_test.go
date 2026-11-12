package app

import "testing"

// Deleting one request must not move the selection, nor overwrite the one
// that slid into its place. This is why the selection follows an ID.
func TestDeletingAnotherRequestKeepsTheSelection(t *testing.T) {
	app := newTestApp(t)

	app.SelectRequest(0, 2) // Update user
	app.openDeleteRequest(0, 1)
	app.confirm()

	if app.draft.Name != "Update user" {
		t.Fatalf("the selection moved to %q", app.draft.Name)
	}
	if got := len(app.Project().Folders[0].Requests); got != 2 {
		t.Errorf("%d requests left, want 2", got)
	}
	if app.Project().Folders[0].Requests[1].Name != "Update user" {
		t.Error("the request that slid up was overwritten by the stale draft")
	}
}
