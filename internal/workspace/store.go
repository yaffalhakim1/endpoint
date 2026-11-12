package workspace

// Package workspace is the document Endpoint keeps: the projects, the
// requests and the secrets. It reads and writes one sealed file, and never
// returns an error the app cannot live with.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DocumentName = "workspace.json"
	BackupSuffix = ".bak"
	documentVer  = 2
)

// recovery says what happened to the workspace when it was opened, so the
// window can tell the user rather than fail.
type Recovery struct {
	Message string // what the user should know, if anything
	Saved   string // where the unreadable file was kept
	Fresh   bool   // started from an empty document
}

// loadDocument reads the document, falling back on the backup and then on a
// new one, and never returning an error the app cannot live with.
func Load(dir string) (*Document, Recovery) {
	path := filepath.Join(dir, DocumentName)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Seed(), Recovery{Fresh: true}
	}
	if err != nil {
		return Seed(), Recovery{
			Message: "The workspace could not be read: " + err.Error(),
			Fresh:   true,
		}
	}

	doc, err := decodeDocument(dir, data)
	if err == nil {
		return doc, Recovery{}
	}

	broken := err

	// The backup from the previous save is the next thing to try.
	if backup, readErr := os.ReadFile(path + BackupSuffix); readErr == nil {
		if doc, backupErr := decodeDocument(dir, backup); backupErr == nil {
			return doc, Recovery{Message: "The workspace was damaged and the last good copy was opened: " + broken.Error()}
		}
	}

	// Neither reads: keep the broken file so nothing is destroyed, and start
	// fresh. The user is told where it went.
	kept := path + ".damaged"
	_ = os.Rename(path, kept)
	return Seed(), Recovery{
		Message: "The workspace could not be read, so a new one was started. The old file is at " + kept + ": " + broken.Error(),
		Saved:   kept,
		Fresh:   true,
	}
}

// decodeDocument opens either a sealed workspace or one of the readable files
// written before sealing existed.
func decodeDocument(dir string, data []byte) (*Document, error) {
	plaintext := data
	if IsSealed(data) {
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
func Save(dir string, doc *Document) error {
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

	path := filepath.Join(dir, DocumentName)
	temp := path + ".tmp"
	if err := os.WriteFile(temp, sealed, 0o600); err != nil {
		return err
	}

	// The current file becomes the backup, then the new one takes its place.
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path + BackupSuffix)
		if err := os.Rename(path, path+BackupSuffix); err != nil {
			return err
		}
	}
	return os.Rename(temp, path)
}
