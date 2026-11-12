package main

// The workspace holds tokens and passwords, so it is not written as readable
// JSON: it is sealed with AES-GCM under a key the system's credential store
// keeps, and a key file beside it on a machine without one.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
)

var sealedMagic = []byte("ENDPOINT\x01")

const (
	keyName = "workspace-key"
	keySize = 32
)

// workspaceKey returns the key that seals the workspace, making one on first
// run. The credential store is asked first; without it, as on a bare runner,
// a key file of 0600 beside the workspace is used instead.
func workspaceKey(dir string) ([]byte, error) {
	if key, err := mygo.Secrets.Get(keyName); err == nil && len(key) == keySize {
		return key, nil
	}

	path := filepath.Join(dir, "workspace.key")
	if data, err := os.ReadFile(path); err == nil && len(data) == keySize {
		return data, nil
	}

	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, err
	}
	// Best effort: a machine whose store refuses the key still works from
	// the file, and one that keeps it never needs the file again.
	_ = mygo.Secrets.Set(keyName, key)
	return key, nil
}

// seal encrypts the document and marks it, so an old readable file is still
// recognised, and sealed once more.
func seal(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	out := append([]byte{}, sealedMagic...)
	out = append(out, nonce...)
	return gcm.Seal(out, nonce, plaintext, nil), nil
}

func isSealed(data []byte) bool {
	return len(data) >= len(sealedMagic) && string(data[:len(sealedMagic)]) == string(sealedMagic)
}

// unseal opens what seal wrote.
func unseal(key, data []byte) ([]byte, error) {
	if !isSealed(data) {
		return nil, errors.New("the file is not a sealed workspace")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	body := data[len(sealedMagic):]
	if len(body) < gcm.NonceSize() {
		return nil, fmt.Errorf("the workspace is cut short")
	}
	nonce, ciphertext := body[:gcm.NonceSize()], body[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
