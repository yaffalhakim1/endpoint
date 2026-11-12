package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The two families ship inside the binary, so a rename or a bad file must
// fail here rather than quietly fall back to a system font.
func TestFontsAreRegistered(t *testing.T) {
	if err := registerFonts(); err != nil {
		t.Fatalf("registering the fonts: %v", err)
	}

	families := ui.FontFamilies()
	for _, want := range []string{uiFont, monoFont} {
		found := false
		for _, family := range families {
			if family.Name == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%q is not among the %d families the renderer has", want, len(families))
		}
	}
}

// Inter must really resolve, not silently fall back to the system face.
func TestInterResolves(t *testing.T) {
	if err := registerFonts(); err != nil {
		t.Fatal(err)
	}
	var system, inter, monoWidth float32
	view := func(c *ui.Context) {
		system, _ = c.MeasureText(2000, ui.Span{Text: "Create user"})
		inter, _ = c.MeasureText(2000, ui.Span{Text: "Create user", Font: uiFont})
		monoWidth, _ = c.MeasureText(2000, ui.Span{Text: "Create user", Font: monoFont})
	}
	ui.NewTester(view, 400, 200)

	if system == inter {
		t.Fatal("Inter measures like the system font, so it is not being used")
	}
	if monoWidth == 0 {
		t.Fatal("the mono family measured nothing")
	}
}
