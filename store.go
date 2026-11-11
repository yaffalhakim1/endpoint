package main

// The document on disk: one JSON file in the user data directory, so the
// app and `mygo dev` read the same projects.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const documentName = "workspace.json"
const documentVersion = 1

// loadDocument reads the document, and seeds a new one when there is none.
func loadDocument(dir string) (*Document, error) {
	data, err := os.ReadFile(filepath.Join(dir, documentName))
	if errors.Is(err, os.ErrNotExist) {
		return seedDocument(), nil
	}
	if err != nil {
		return nil, err
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Version == 0 || len(doc.Projects) == 0 {
		return seedDocument(), nil
	}
	return &doc, nil
}

// saveDocument replaces the file in one step: a crash mid-write leaves the
// previous document intact.
func saveDocument(dir string, doc *Document) error {
	doc.Version = documentVersion

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(dir, documentName)
	temp := path + ".tmp"
	if err := os.WriteFile(temp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

// seedDocument is the first run: the shapes from the design, so the window
// is never empty.
func seedDocument() *Document {
	users := Folder{
		ID:   "folder-users",
		Name: "Users",
		Open: true,
		Requests: []Request{
			{
				ID:     "req-create-user",
				Name:   "Create user",
				Method: "POST",
				URL:    "https://api.example.com/v1/users",
				Headers: []Pair{
					{Key: "Content-Type", Value: "application/json"},
					{Key: "Accept", Value: "application/json"},
				},
				Body: BodyJSON,
				JSON: "{\n  \"name\": \"Ada Lovelace\",\n  \"email\": \"ada@example.com\",\n  \"plan\": \"pro\",\n  \"seats\": 12\n}",
			},
			{ID: "req-get-user", Name: "Get user", Method: "GET", URL: "{{baseUrl}}/v1/users/ada"},
			{ID: "req-update-user", Name: "Update user", Method: "PUT", URL: "https://api.example.com/v1/users/ada"},
		},
	}
	billing := Folder{
		ID:   "folder-billing",
		Name: "Billing",
		Requests: []Request{
			{ID: "req-list-invoices", Name: "List invoices", Method: "GET", URL: "https://api.example.com/v1/invoices"},
			{ID: "req-create-invoice", Name: "Create invoice", Method: "POST", URL: "https://api.example.com/v1/invoices"},
		},
	}
	sessions := Folder{
		ID:   "folder-sessions",
		Name: "Sessions",
		Requests: []Request{
			{ID: "req-revoke", Name: "Revoke session", Method: "DELETE", URL: "https://api.example.com/v1/sessions/9f2"},
		},
	}

	return &Document{
		Version: documentVersion,
		Active:  "project-personal",
		Projects: []Project{
			{ID: "project-personal", Name: "Personal", Vars: []Pair{
				{Key: "baseUrl", Value: "https://api.example.com"},
				{Key: "token", Value: ""},
			}, Folders: []Folder{users, billing, sessions}},
			{ID: "project-work", Name: "Work", Folders: nil},
			{ID: "project-sandbox", Name: "Sandbox", Folders: nil},
		},
	}
}
