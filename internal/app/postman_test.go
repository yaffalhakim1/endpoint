package app

import (
	"strings"
	"testing"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// A Postman collection is a project: folders, requests, variables and auth.

const sampleCollection = `{
  "info": {
    "name": "Payments API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    { "key": "baseUrl", "value": "https://api.example.com" },
    { "key": "token", "value": "sk_live_123" }
  ],
  "auth": {
    "type": "bearer",
    "bearer": [ { "key": "token", "value": "{{token}}" } ]
  },
  "item": [
    {
      "name": "Users",
      "item": [
        {
          "name": "Create user",
          "request": {
            "method": "POST",
            "header": [
              { "key": "Content-Type", "value": "application/json" },
              { "key": "X-Skip", "value": "no", "disabled": true }
            ],
            "url": {
              "raw": "{{baseUrl}}/v1/users?notify=true",
              "query": [ { "key": "limit", "value": "3" } ]
            },
            "body": {
              "mode": "raw",
              "raw": "{\"name\":\"Ada\"}",
              "options": { "raw": { "language": "json" } }
            }
          }
        },
        {
          "name": "Get user",
          "request": {
            "method": "GET",
            "url": { "raw": "{{baseUrl}}/v1/users/ada" }
          }
        }
      ]
    },
    {
      "name": "Billing",
      "item": [
        {
          "name": "Invoices",
          "item": [
            {
              "name": "List invoices",
              "request": {
                "method": "GET",
                "url": { "raw": "{{baseUrl}}/v1/invoices" },
                "auth": {
                  "type": "basic",
                  "basic": [
                    { "key": "username", "value": "ada" },
                    { "key": "password", "value": "secret" }
                  ]
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`

func TestParsePostmanCollection(t *testing.T) {
	project, err := ParsePostman([]byte(sampleCollection))
	if err != nil {
		t.Fatal(err)
	}

	if project.Name != "Payments API" {
		t.Errorf("the project is %q", project.Name)
	}
	if len(project.Vars) != 2 {
		t.Fatalf("%d variables, want 2", len(project.Vars))
	}
	if project.Auth.Type != workspace.AuthBearer || project.Auth.Token != "{{token}}" {
		t.Errorf("the collection auth is %+v", project.Auth)
	}
	if len(project.Folders) != 2 {
		t.Fatalf("%d folders, want 2: %+v", len(project.Folders), project.Folders)
	}

	users := project.Folders[0]
	if users.Name != "Users" || len(users.Requests) != 2 {
		t.Fatalf("the first folder is %q with %d requests", users.Name, len(users.Requests))
	}

	create := users.Requests[0]
	if create.Method != "POST" || create.Name != "Create user" {
		t.Errorf("the request is %s %q", create.Method, create.Name)
	}
	if create.URL != "{{baseUrl}}/v1/users" {
		t.Errorf("the address is %q, want the query taken off", create.URL)
	}
	if len(create.Params) != 2 {
		t.Errorf("%d params, want the one in the address and the one in the list", len(create.Params))
	}
	if len(create.Headers) != 1 {
		t.Errorf("%d headers, want the disabled one dropped", len(create.Headers))
	}
	if create.Body != workspace.BodyJSON || !strings.Contains(create.JSON, "Ada") {
		t.Errorf("the body is %v %q", create.Body, create.JSON)
	}
}

func TestNestedFoldersFlatten(t *testing.T) {
	project, err := ParsePostman([]byte(sampleCollection))
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, folder := range project.Folders {
		if folder.Name == "Billing / Invoices" {
			found = true
			if len(folder.Requests) != 1 {
				t.Errorf("the nested folder holds %d requests", len(folder.Requests))
			}
			if folder.Requests[0].Auth.Type != workspace.AuthBasic {
				t.Errorf("the request auth is %+v", folder.Requests[0].Auth)
			}
		}
	}
	if !found {
		t.Errorf("the nested folder did not flatten: %+v", project.Folders)
	}
}

func TestPostmanFormBody(t *testing.T) {
	collection := `{
	  "info": { "name": "Forms", "schema": "collection" },
	  "item": [ {
	    "name": "Upload",
	    "request": {
	      "method": "POST",
	      "url": { "raw": "https://api.example.com/upload" },
	      "body": {
	        "mode": "formdata",
	        "formdata": [
	          { "key": "name", "value": "Ada" },
	          { "key": "skip", "value": "x", "disabled": true }
	        ]
	      }
	    }
	  } ]
	}`

	project, err := ParsePostman([]byte(collection))
	if err != nil {
		t.Fatal(err)
	}
	request := project.Folders[0].Requests[0]
	if request.Body != workspace.BodyForm {
		t.Fatalf("the body is %v", request.Body)
	}
	if len(request.Form) != 1 || request.Form[0].Key != "name" {
		t.Errorf("the form is %+v", request.Form)
	}
}

func TestPostmanRejectsNonsense(t *testing.T) {
	for _, input := range []string{"", "not json", `{"hello":"world"}`, `[]`} {
		if _, err := ParsePostman([]byte(input)); err == nil {
			t.Errorf("%q was accepted", input)
		}
	}
}

func TestImportingACollectionAddsAProject(t *testing.T) {
	app := newTestApp(t)
	before := len(app.doc.Projects)

	app.importKind = 1
	app.importText = sampleCollection
	app.runImport()

	if got := len(app.doc.Projects); got != before+1 {
		t.Fatalf("%d projects after the import, want %d", got, before+1)
	}
	if app.Project().Name != "Payments API" {
		t.Errorf("the active project is %q", app.Project().Name)
	}
	if app.importOpen {
		t.Error("the dialog stayed open")
	}
	if app.draft.Name != "Create user" {
		t.Errorf("the selection is %q", app.draft.Name)
	}
}

func TestABadCollectionKeepsTheDialogOpen(t *testing.T) {
	app := newTestApp(t)
	app.importKind = 1
	app.importText = "not a collection"
	app.importOpen = true

	app.runImport()

	if !app.importOpen {
		t.Error("the dialog closed on a bad collection")
	}
	if app.importError == "" {
		t.Error("no reason was given")
	}
}
