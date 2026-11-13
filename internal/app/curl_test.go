package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// A pasted curl command is the fastest way to bring work in, so the parser
// has to survive what people actually paste.

func TestParseCurlBasic(t *testing.T) {
	request, err := ParseCurl(`curl https://api.example.com/v1/users`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "GET" || request.URL != "https://api.example.com/v1/users" {
		t.Errorf("got %s %s", request.Method, request.URL)
	}
	if request.Name != "users" {
		t.Errorf("the name is %q", request.Name)
	}
}

func TestParseCurlPostWithHeadersAndBody(t *testing.T) {
	command := `curl 'https://api.example.com/v1/users' \
  -X POST \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer abc123' \
  -d '{"name":"Ada","plan":"pro"}'`

	request, err := ParseCurl(command)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "POST" {
		t.Errorf("method %q", request.Method)
	}
	if len(request.Headers) != 2 {
		t.Fatalf("%d headers", len(request.Headers))
	}
	if request.Headers[0].Key != "Content-Type" || request.Headers[0].Value != "application/json" {
		t.Errorf("the first header is %+v", request.Headers[0])
	}
	if request.Body != workspace.BodyJSON || !strings.Contains(request.JSON, "Ada") {
		t.Errorf("the body is %v %q", request.Body, request.JSON)
	}
}

func TestParseCurlInfersPostFromData(t *testing.T) {
	request, err := ParseCurl(`curl -d 'a=1' https://api.example.com/x`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "POST" {
		t.Errorf("method %q, want POST", request.Method)
	}
	if request.Body != workspace.BodyForm || len(request.Form) != 1 {
		t.Errorf("the body is %v with %d fields", request.Body, len(request.Form))
	}
}

func TestParseCurlBasicAuth(t *testing.T) {
	request, err := ParseCurl(`curl -u ada:secret https://api.example.com/x`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Auth.Type != workspace.AuthBasic || request.Auth.Username != "ada" || request.Auth.Password != "secret" {
		t.Errorf("auth is %+v", request.Auth)
	}
}

func TestParseCurlForm(t *testing.T) {
	request, err := ParseCurl(`curl -F 'name=Ada' -F 'plan=pro' https://api.example.com/x`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Body != workspace.BodyForm || len(request.Form) != 2 {
		t.Errorf("the body is %v with %d fields", request.Body, len(request.Form))
	}
}

func TestParseCurlRejectsNonsense(t *testing.T) {
	for _, command := range []string{"", "wget https://x.dev", "curl", `curl 'unclosed`} {
		if _, err := ParseCurl(command); err == nil {
			t.Errorf("%q was accepted", command)
		}
	}
}

func TestImportedRequestJoinsTheProject(t *testing.T) {
	app := newTestApp(t)
	before := app.Project().CountRequests()

	request, err := ParseCurl(`curl -X PUT https://api.example.com/v1/widgets/7 -d '{"on":true}'`)
	if err != nil {
		t.Fatal(err)
	}
	app.addRequest(request)

	if got := app.Project().CountRequests(); got != before+1 {
		t.Fatalf("%d requests after the import, want %d", got, before+1)
	}
	if app.draft.Method != "PUT" || app.draft.URL != "https://api.example.com/v1/widgets/7" {
		t.Errorf("the imported draft is %s %s", app.draft.Method, app.draft.URL)
	}
}

func TestExportWritesReadableJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workspace.json")

	data, err := exportDocument(workspace.Seed())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"Personal"`) {
		t.Error("the export does not name the project")
	}
	if !strings.Contains(string(data), `"baseUrl"`) {
		t.Error("the export does not carry the variables")
	}
}
