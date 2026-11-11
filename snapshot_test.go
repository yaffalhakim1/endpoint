package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestSnapshot renders the window without a screen, for looking at the
// layout: go test -run TestSnapshot -v
func TestSnapshot(t *testing.T) {
	loaded := Result{
		Code:     201,
		Status:   "201 Created",
		Duration: 142 * 1000 * 1000,
		Body:     []byte(`{"id":"usr_8f2a1c","name":"Ada Lovelace","email":"ada@example.com","plan":"pro","seats":12,"active":true,"last_login":null,"roles":["admin","billing"]}`),
	}
	failed := Result{
		Code:     504,
		Status:   "504 Gateway Timeout",
		Duration: 30 * 1000 * 1000 * 1000,
		Body:     []byte(`{"error":{"code":"upstream_timeout","message":"The upstream service did not answer within 30 seconds.","request_id":"req_5f19c2"}}`),
	}

	shots := []struct {
		name  string
		setup func(*App)
	}{
		{name: "response", setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
		}},
		{name: "auth", setup: func(a *App) {
			a.tab = 2
			a.authType = 1
			a.draft.Auth.Type = AuthBearer
			a.draft.Auth.Token = "sk_live_8f2a1c4d9b7e"
			a.state = StateLoaded
			a.result = loaded
		}},
		{name: "failed", setup: func(a *App) {
			a.state = StateFailed
			a.result = failed
		}},
		{name: "headers", setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
			a.result.Headers = []Pair{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Date", Value: "Wed, 11 Nov 2026 15:09:29 GMT"},
				{Key: "X-Ratelimit-Remaining", Value: "42"},
			}
			a.responseTab = 1
		}},
		{name: "sidebar-closed", setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
			a.sidebarHidden = true
		}},
		{name: "rename", setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
			a.openRenameRequest(0, 1)
		}},
	}

	for _, shot := range shots {
		for _, dark := range []bool{false, true} {
			app := newTestApp(t)
			shot.setup(app)

			tt := ui.NewTester(app.view, 1180, 780)
			tt.SetDark(dark)

			suffix := "light"
			if dark {
				suffix = "dark"
			}
			path := filepath.Join(os.TempDir(), "endpoint-"+shot.name+"-"+suffix+".png")

			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(file, tt.Image()); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			t.Logf("wrote %s", path)
		}
	}
}
