package workspace

// The document a new install starts from: the shapes the window shows, so it
// is never blank, with a variable in a URL so the feature is visible.

func Seed() *Document {
	users := Folder{
		ID:   "folder-users",
		Name: "Users",
		Open: true,
		Requests: []Request{
			{
				ID:     "req-create-user",
				Name:   "Create user",
				Method: "POST",
				URL:    "{{baseUrl}}/v1/users",
				Headers: []Pair{
					{Key: "Content-Type", Value: "application/json"},
					{Key: "Accept", Value: "application/json"},
				},
				Body: BodyJSON,
				JSON: "{\n  \"name\": \"Ada Lovelace\",\n  \"email\": \"ada@example.com\",\n  \"plan\": \"pro\",\n  \"seats\": 12\n}",
			},
			{ID: "req-get-user", Name: "Get user", Method: "GET", URL: "{{baseUrl}}/v1/users/ada"},
			{ID: "req-update-user", Name: "Update user", Method: "PUT", URL: "{{baseUrl}}/v1/users/ada", Body: BodyNone},
		},
	}

	billing := Folder{
		ID:   "folder-billing",
		Name: "Billing",
		Requests: []Request{
			{ID: "req-list-invoices", Name: "List invoices", Method: "GET", URL: "{{baseUrl}}/v1/invoices", Body: BodyNone},
			{ID: "req-create-invoice", Name: "Create invoice", Method: "POST", URL: "{{baseUrl}}/v1/invoices", Body: BodyNone},
		},
	}

	sessions := Folder{
		ID:   "folder-sessions",
		Name: "Sessions",
		Requests: []Request{
			{ID: "req-revoke", Name: "Revoke session", Method: "DELETE", URL: "{{baseUrl}}/v1/sessions/9f2", Body: BodyNone},
		},
	}

	return &Document{
		Version: documentVer,
		Active:  "project-personal",
		Projects: []Project{
			{
				ID:   "project-personal",
				Name: "Personal",
				Vars: []Pair{
					{Key: "baseUrl", Value: "https://api.example.com"},
					{Key: "token", Value: ""},
				},
				Folders: []Folder{users, billing, sessions},
			},
			{ID: "project-work", Name: "Work"},
			{ID: "project-sandbox", Name: "Sandbox"},
		},
	}
}
