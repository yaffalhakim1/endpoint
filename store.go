package main

// The document on disk: one sealed file in the user data directory, with the
// previous copy kept beside it. Nothing here may take the app down: a file
// that cannot be read is reported, and the app starts anyway.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	documentName = "workspace.json"
	backupSuffix = ".bak"
	documentVer  = 2
)

// recovery says what happened to the workspace when it was opened, so the
// window can tell the user rather than fail.
type recovery struct {
	Message string // what the user should know, if anything
	Saved   string // where the unreadable file was kept
	Fresh   bool   // started from an empty document
}

// loadDocument reads the document, falling back on the backup and then on a
// new one, and never returning an error the app cannot live with.
func loadDocument(dir string) (*Document, recovery) {
	path := filepath.Join(dir, documentName)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return seedDocument(), recovery{Fresh: true}
	}
	if err != nil {
		return seedDocument(), recovery{
			Message: "The workspace could not be read: " + err.Error(),
			Fresh:   true,
		}
	}

	doc, err := decodeDocument(dir, data)
	if err == nil {
		return doc, recovery{}
	}

	broken := err

	// The backup from the previous save is the next thing to try.
	if backup, readErr := os.ReadFile(path + backupSuffix); readErr == nil {
		if doc, backupErr := decodeDocument(dir, backup); backupErr == nil {
			return doc, recovery{Message: "The workspace was damaged and the last good copy was opened: " + broken.Error()}
		}
	}

	// Neither reads: keep the broken file so nothing is destroyed, and start
	// fresh. The user is told where it went.
	kept := path + ".damaged"
	_ = os.Rename(path, kept)
	return seedDocument(), recovery{
		Message: "The workspace could not be read, so a new one was started. The old file is at " + kept + ": " + broken.Error(),
		Saved:   kept,
		Fresh:   true,
	}
}

// decodeDocument opens either a sealed workspace or one of the readable files
// written before sealing existed.
func decodeDocument(dir string, data []byte) (*Document, error) {
	plaintext := data
	if isSealed(data) {
		key, err := workspaceKey(dir)
		if err != nil {
			return nil, err
		}
		opened, err := unseal(key, data)
		if err != nil {
			return nil, fmt.Errorf("the workspace cannot be opened with this machine's key: %w", err)
		}
		plaintext = opened
	}

	var doc Document
	if err := json.Unmarshal(plaintext, &doc); err != nil {
		return nil, err
	}
	if len(doc.Projects) == 0 {
		return nil, errors.New("the workspace holds no projects")
	}
	return &doc, nil
}

// saveDocument replaces the file in one step, keeping the copy before it as a
// backup: a crash, a full disk or a bug leaves one of the two readable.
func saveDocument(dir string, doc *Document) error {
	doc.Version = documentVer

	plaintext, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if len(plaintext) > 8<<20 {
		return errors.New("the workspace is larger than 8 MB")
	}

	key, err := workspaceKey(dir)
	if err != nil {
		return err
	}
	sealed, err := seal(key, plaintext)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(dir, documentName)
	temp := path + ".tmp"
	if err := os.WriteFile(temp, sealed, 0o600); err != nil {
		return err
	}

	// The current file becomes the backup, then the new one takes its place.
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path + backupSuffix)
		if err := os.Rename(path, path+backupSuffix); err != nil {
			return err
		}
	}
	return os.Rename(temp, path)
}
