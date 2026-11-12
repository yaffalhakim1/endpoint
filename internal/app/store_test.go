package app

import (
	"encoding/json"
	"github.com/yaffalhakim1/endpoint/internal/workspace"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The workspace must never be the reason the app will not open, and it must
// not leave tokens readable on disk.

func TestWorkspaceIsSealedOnDisk(t *testing.T) {
	dir := t.TempDir()

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.draft.Auth = Auth{Type: AuthBearer, Token: "sk_live_super_secret"}
	app.save()

	data, err := os.ReadFile(filepath.Join(dir, workspace.DocumentName))
	if err != nil {
		t.Fatal(err)
	}
	if !workspace.IsSealed(data) {
		t.Fatal("the workspace was written as readable JSON")
	}
	if strings.Contains(string(data), "sk_live_super_secret") {
		t.Fatal("the token is readable in the file")
	}

	// And it opens again with the same key.
	reloaded, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.draft.Auth.Token; got != "sk_live_super_secret" {
		t.Errorf("the token came back as %q", got)
	}
}

func TestWorkspaceKeepsABackup(t *testing.T) {
	dir := t.TempDir()

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.save()
	first, err := os.ReadFile(filepath.Join(dir, workspace.DocumentName))
	if err != nil {
		t.Fatal(err)
	}

	app.draft.URL = "https://api.example.com/v1/health"
	app.save()

	backup, err := os.ReadFile(filepath.Join(dir, workspace.DocumentName+workspace.BackupSuffix))
	if err != nil {
		t.Fatalf("there is no backup: %v", err)
	}
	if string(backup) != string(first) {
		t.Error("the backup is not the file the save replaced")
	}
	if !workspace.IsSealed(backup) {
		t.Error("the backup is not sealed")
	}
	current, err := os.ReadFile(filepath.Join(dir, workspace.DocumentName))
	if err != nil {
		t.Fatal(err)
	}
	if string(current) == string(first) {
		t.Error("the second save changed nothing, so the backup was not tested")
	}
}

func TestDamagedWorkspaceFallsBackToTheBackup(t *testing.T) {
	dir := t.TempDir()

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.draft.URL = "https://api.example.com/keep-me" // written to the backup on the next save
	app.save()
	app.draft.URL = "https://api.example.com/latest"
	app.save()

	// Corrupt the live file, keeping the backup intact.
	path := filepath.Join(dir, workspace.DocumentName)
	if err := os.WriteFile(path, []byte("this is not a workspace"), 0o600); err != nil {
		t.Fatal(err)
	}

	recovered, err := New(dir)
	if err != nil {
		t.Fatalf("a damaged workspace stopped the app: %v", err)
	}
	if recovered.recovered.Message == "" {
		t.Error("the user was not told what happened")
	}
	if recovered.draft.URL != "https://api.example.com/keep-me" {
		t.Errorf("the backup did not come back: %q", recovered.draft.URL)
	}
}

func TestUnreadableWorkspaceKeepsTheBrokenFile(t *testing.T) {
	dir := t.TempDir()

	// A file and its backup that both fail.
	for _, name := range []string{workspace.DocumentName, workspace.DocumentName + workspace.BackupSuffix} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{ not json"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	app, err := New(dir)
	if err != nil {
		t.Fatalf("a broken workspace stopped the app: %v", err)
	}
	if !app.recovered.Fresh {
		t.Error("the app did not start from a new workspace")
	}
	if app.recovered.Saved == "" {
		t.Fatal("the user was not told where the broken file went")
	}
	if _, err := os.Stat(app.recovered.Saved); err != nil {
		t.Errorf("the broken file was not kept: %v", err)
	}
	if len(app.doc.Projects) == 0 {
		t.Error("the new workspace has no projects")
	}
}

func TestReadableWorkspaceFromTheOlderVersionOpens(t *testing.T) {
	dir := t.TempDir()

	// What version 1 wrote: plain JSON, no sealing.
	doc := workspace.Seed()
	doc.Version = 1
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, workspace.DocumentName), data, 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := New(dir)
	if err != nil {
		t.Fatalf("an older workspace stopped the app: %v", err)
	}
	if app.recovered.Message != "" {
		t.Errorf("an older workspace was reported as damaged: %q", app.recovered.Message)
	}
	if app.Project().Name != "Personal" {
		t.Errorf("the project is %q", app.Project().Name)
	}
}
