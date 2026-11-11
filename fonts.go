package main

// The app's typefaces. Inter is an open font built to work where SF Pro does,
// which cannot be shipped outside Apple's platforms; Geist Mono gives the code
// the same voice as the rest of the interface.

import (
	"embed"
	"io/fs"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
)

//go:embed all:resources/fonts
var fontFiles embed.FS

const (
	uiFont   = "Inter"
	monoFont = "Geist Mono"
)

var registerFontsOnce sync.Once

// registerFonts brings both families into the renderer, once.
func registerFonts() error {
	var err error
	registerFontsOnce.Do(func() {
		names, globErr := fs.Glob(fontFiles, "resources/fonts/*.otf")
		if globErr != nil {
			err = globErr
			return
		}
		for _, name := range names {
			data, readErr := fontFiles.ReadFile(name)
			if readErr != nil {
				err = readErr
				return
			}
			if regErr := ui.RegisterFont(data, familyOf(name)); regErr != nil {
				err = regErr
				return
			}
		}
	})
	return err
}

// familyOf reads which family a file belongs to from its name.
func familyOf(path string) string {
	if strings.Contains(strings.ToLower(path), "geistmono") {
		return monoFont
	}
	return uiFont
}
