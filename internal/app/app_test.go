package app

import (
	"context"
	"encoding/json"
	"github.com/yaffalhakim1/endpoint/internal/request"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// newTestApp is an app on a document of its own, so tests never touch the
// user's workspace.
func newTestApp(t *testing.T) *App {
	t.Helper()
	app, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app.tab = 3 // Body, as the design opens
	return app
}

func TestViewShowsTheWorkspace(t *testing.T) {
	app := newTestApp(t)
	tt := ui.NewTester(app.View, 1180, 780)

	// The title bar names what the window shows, not the app.
	for _, want := range []string{"Create user", "No environment", "Personal", "Users", "Send", "Body"} {
		if !tt.HasText(want) {
			t.Errorf("the view does not show %q; texts: %q", want, tt.Texts())
		}
	}
	if tt.HasText("Endpoint") {
		t.Error("the title bar names the app instead of the request")
	}
}

func TestChoosingARequestLoadsIt(t *testing.T) {
	app := newTestApp(t)
	tt := ui.NewTester(app.View, 1180, 780)

	if err := tt.Click("Get user"); err != nil {
		t.Fatal(err)
	}

	if app.draft.URL != "{{baseUrl}}/v1/users/ada" {
		t.Fatalf("the URL is %q after choosing Get user", app.draft.URL)
	}
	if app.draft.Method != "GET" {
		t.Fatalf("the method is %q, want GET", app.draft.Method)
	}
}

func TestAddingARequestFromTheRail(t *testing.T) {
	app := newTestApp(t)
	tt := ui.NewTester(app.View, 1180, 780)

	before := app.Project().CountRequests()
	if err := tt.Click("+"); err != nil {
		t.Fatal(err)
	}

	if got := app.Project().CountRequests(); got != before+1 {
		t.Fatalf("%d requests after adding one, want %d", got, before+1)
	}
	if app.draft.Name != "New request" {
		t.Fatalf("the draft is %q, want the new request", app.draft.Name)
	}
}

func TestSendBuildsTheRequest(t *testing.T) {
	var gotPath, gotQuery, gotAuth, gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"usr_1","ok":true}`))
	}))
	defer server.Close()

	result := sendForTest(Request{
		Method:  "POST",
		URL:     server.URL + "/v1/users",
		Params:  []Pair{{Key: "limit", Value: "20"}},
		Headers: []Pair{{Key: "Accept", Value: "application/json"}},
		Auth:    Auth{Type: AuthBearer, Token: "abc123"},
		Body:    BodyJSON,
		JSON:    `{"name":"Ada"}`,
	}, Auth{}, server.Client())

	if result.Err != "" {
		t.Fatalf("the call failed: %s", result.Err)
	}
	if result.Code != http.StatusCreated {
		t.Errorf("status %d, want 201", result.Code)
	}
	if gotPath != "/v1/users" || gotQuery != "limit=20" {
		t.Errorf("reached %s?%s, want /v1/users?limit=20", gotPath, gotQuery)
	}
	if gotAuth != "Bearer abc123" {
		t.Errorf("Authorization is %q", gotAuth)
	}
	if gotBody != `{"name":"Ada"}` {
		t.Errorf("body is %q", gotBody)
	}
	if !strings.Contains(string(result.Body), "usr_1") {
		t.Errorf("body of the answer is %q", result.Body)
	}
}

func TestInheritedAuthFillsTheHeader(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-API-Key")
	}))
	defer server.Close()

	result := sendForTest(
		Request{Method: "GET", URL: server.URL, Auth: Auth{Type: AuthInherit}},
		Auth{Type: AuthKey, Header: "X-API-Key", Value: "project-key"},
		server.Client(),
	)
	if result.Err != "" {
		t.Fatalf("the call failed: %s", result.Err)
	}
	if got != "project-key" {
		t.Errorf("X-API-Key is %q, want the project's key", got)
	}
}

func TestEmptyURLIsAnError(t *testing.T) {
	result := sendForTest(Request{Method: "GET"}, Auth{}, http.DefaultClient)
	if result.Err == "" {
		t.Fatal("an empty URL did not fail")
	}
}

func TestDocumentRoundTrips(t *testing.T) {
	dir := t.TempDir()

	first, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	first.draft.URL = "https://api.example.com/v1/health"
	first.save()

	second, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if second.draft.URL != "https://api.example.com/v1/health" {
		t.Fatalf("reloaded URL is %q", second.draft.URL)
	}
}

func TestPrettyJSON(t *testing.T) {
	pretty, ok := request.PrettyJSON([]byte(`{"a":1}`))
	if !ok {
		t.Fatal("a JSON object was not recognised")
	}
	if !json.Valid([]byte(pretty)) {
		t.Fatalf("the pretty body is not JSON: %q", pretty)
	}
	if _, ok := request.PrettyJSON([]byte("nope")); ok {
		t.Fatal("plain text was taken for JSON")
	}
}

// sendForTest calls the client without a context or variables, which is what
// the plain cases need.
func sendForTest(req Request, inherit Auth, client *http.Client) Result {
	return request.Send(context.Background(), req, inherit, nil, client)
}
