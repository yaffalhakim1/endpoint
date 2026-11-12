package app

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestSiteShots writes the frames the landing page shows, at twice the size,
// so they stay sharp on a high-density screen. It is a build step for the
// site, not an assertion: go test -run TestSiteShots ./internal/app
func TestSiteShots(t *testing.T) {
	body := `{"id":"usr_8f2a1c","name":"Ada Lovelace","email":"ada@example.com",` +
		`"plan":"pro","seats":12,"active":true,"last_login":null,"roles":["admin","billing"],` +
		`"limits":{"requests_per_minute":10000,"used":4213}}`

	loaded := Result{
		Code:     201,
		Status:   "201 Created",
		Duration: 142 * 1000 * 1000,
		Body:     []byte(body),
		Headers: []Pair{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "Date", Value: "Wed, 11 Nov 2026 15:09:29 GMT"},
			{Key: "X-Ratelimit-Remaining", Value: "42"},
			{Key: "X-Request-Id", Value: "req_5f19c2"},
		},
	}

	shots := []struct {
		name  string
		dark  bool
		setup func(*App)
	}{
		{name: "window-dark", dark: true, setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
		}},
		{name: "window-light", setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
		}},
		{name: "headers-dark", dark: true, setup: func(a *App) {
			a.state = StateLoaded
			a.result = loaded
			a.responseTab = 1
		}},
		{name: "auth-light", setup: func(a *App) {
			a.tab = 2
			a.authType = 1
			a.draft.Auth.Type = AuthBearer
			a.draft.Auth.Token = "sk_live_8f2a1c4d9b7e"
			a.state = StateLoaded
			a.result = loaded
		}},
	}

	dir := filepath.Join("..", "..", "site", "images")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, shot := range shots {
		app := newTestApp(t)
		shot.setup(app)

		tt := ui.NewTester(app.View, 1180, 780)
		tt.SetDark(shot.dark)
		tt.SetScale(2)

		path := filepath.Join(dir, shot.name+".png")
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
