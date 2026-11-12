package app

// The Body and Auth panels.

import (
	"fmt"
	"path/filepath"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

// jsonEditor is the request body in the toolkit's code editor, the one place a
// text area was not enough: JSON wants its tokens coloured and its lines
// counted. The buffer is rebuilt when another request is loaded.
func (a *App) jsonEditor(c *ui.Context) {
	if a.body == nil || a.bodyFor != a.draft.ID {
		buffer := editor.NewBuffer(a.draft.JSON)
		if language := editor.LanguageFor("body.json"); language != nil {
			buffer.SetLanguage(language)
		}
		a.body = &editor.State{Buffer: buffer, LineNumbers: editor.LineNumbersOn, SoftWrap: true}
		a.bodyFor = a.draft.ID
	}

	before := a.body.Buffer.Version()
	editor.View(c, a.body).FillWidth().Height(148)

	if a.body.Buffer.Version() != before {
		a.draft.JSON = a.body.Buffer.String()
		a.touch(c.Now())
	}
}

func (a *App) bodyPanel(c *ui.Context, t *ui.Theme) {
	ui.Row(c).
		Height(38).
		PaddingX(14).
		AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).
		BorderColor(t.Border).
		Children(func() {
			seg := ui.Segmented(c, &a.bodyType, bodyLabels...)
			if seg.Changed() {
				a.draft.Body = bodyTypes[a.bodyType]
				a.touch(c.Now())
			}
			ui.Spacer(c)
			ui.Text(c, contentType(a.bodyType)).FontSize(12).TextColor(t.TextMuted)
		})

	switch a.bodyType {
	case 0:
		a.jsonEditor(c)
	case 1:
		a.pairEditor(c, t, &a.draft.Form, "value")
		a.fileChip(c, t)
	case 2:
		a.fileDrop(c, t)
	case 3:
		ui.Text(c, "This request sends no body.").FontSize(12.5).TextColor(t.TextMuted).Padding(16)
	}
}

// fileDrop is the target of a dragged file. The pointer's answer is the
// border, as the design's comp does in HTML.
func (a *App) fileDrop(c *ui.Context, t *ui.Theme) {
	drop := ui.Column(c).
		Margin(10, 14).
		Padding(14).
		Radius(10).
		Gap(6).
		AlignItems(ui.Center).
		Border(1, t.Border)

	drop.Children(func() {
		ui.Text(c, "Drop a file here").FontSize(12.5).Bold()
		ui.Text(c, "Sent as the whole body, up to 64 MB. A path typed below is sent too.").FontSize(11.5).TextColor(t.TextMuted)
		if a.draft.File != "" {
			ui.Text(c, filepath.Base(a.draft.File)).Font(mono).FontSize(11.5).TextColor(t.TextMuted)
		}
	})

	// The file comes from a drop or from a typed path: this toolkit has no
	// file dialog of its own.
	path := ui.TextInput(c, &a.draft.File).
		Font(mono).
		FontSize(12).
		Placeholder("…\\file.png, or drop one above")
	path.Label("File path")
	if path.Changed() {
		a.touch(c.Now())
	}

	if drop.FileDragOver() {
		drop.BorderColor(a.tok.Accent).Background(a.tok.Accent.Alpha(0.08))
	}
	if files := drop.DroppedFiles(); len(files) > 0 {
		a.draft.File = files[0]
		a.touch(c.Now())
	}
}

func (a *App) fileChip(c *ui.Context, t *ui.Theme) {
	if a.draft.File == "" {
		return
	}
	ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Box(c).
			Padding(2, 8).
			Radius(5).
			Background(t.Background).
			Children(func() {
				ui.Text(c, filepath.Base(a.draft.File)).FontSize(11.5).TextColor(t.TextMuted)
			})
		ui.Text(c, "file part").FontSize(11.5).TextColor(t.TextMuted).Grow(1)
		if ui.Button(c, "×").Tooltip("Remove the file").Clicked() {
			a.draft.File = ""
			a.touch(c.Now())
		}
	})
}

func (a *App) authPanel(c *ui.Context, t *ui.Theme) {
	ui.Row(c).
		Height(38).
		PaddingX(14).
		AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).
		BorderColor(t.Border).
		Children(func() {
			seg := ui.Segmented(c, &a.authType, authLabels...)
			if seg.Changed() {
				a.draft.Auth.Type = authTypes[a.authType]
				a.touch(c.Now())
			}
		})

	ui.Scroll(c).Grow(1).Padding(12, 14).Gap(10).Children(func() {
		switch a.draft.Auth.Type {
		case AuthBearer:
			a.authField(c, t, "Token", &a.draft.Auth.Token, true)
			ui.Text(c, "Sent as Authorization: Bearer <token>").Font(mono).FontSize(11.5).TextColor(t.TextMuted)
		case AuthBasic:
			a.authField(c, t, "Username", &a.draft.Auth.Username, false)
			a.authField(c, t, "Password", &a.draft.Auth.Password, true)
			ui.Text(c, "Sent as Authorization: Basic <base64>").Font(mono).FontSize(11.5).TextColor(t.TextMuted)
		case AuthKey:
			a.authField(c, t, "Header", &a.draft.Auth.Header, false)
			a.authField(c, t, "Value", &a.draft.Auth.Value, true)
			ui.Text(c, "Added to every request that uses it").Font(mono).FontSize(11.5).TextColor(t.TextMuted)
		case AuthInherit:
			ui.Text(c, a.inheritNote()).FontSize(12.5).TextColor(t.TextMuted)
		default:
			ui.Text(c, "This request carries no credentials.").FontSize(12.5).TextColor(t.TextMuted)
		}
	})
}

func (a *App) authField(c *ui.Context, t *ui.Theme, label string, value *string, secret bool) {
	ui.Column(c).Gap(5).Children(func() {
		ui.Text(c, label).FontSize(11.5).TextColor(t.TextMuted)
		input := ui.TextInput(c, value).Font(mono).FontSize(12.5).MaxWidth(440)
		input.Label(label)
		if secret {
			input.Password()
		}
		if input.Changed() {
			a.touch(c.Now())
		}
	})
}

// inheritNote says what the project's own credentials are.
func (a *App) inheritNote() string {
	project := a.Project()
	switch project.Auth.Type {
	case AuthBearer:
		return "Uses this project's Bearer token."
	case AuthBasic:
		return "Uses this project's Basic credentials."
	case AuthKey:
		return fmt.Sprintf("Uses this project's API key in %s.", a.keyName(project.Auth))
	default:
		return "This project has no credentials of its own."
	}
}

func (a *App) keyName(auth Auth) string {
	if auth.Header != "" {
		return auth.Header
	}
	return "X-API-Key"
}

func contentType(index int) string {
	switch index {
	case 0:
		return "application/json"
	case 1:
		return "multipart/form-data"
	case 2:
		return "raw file"
	default:
		return "no body"
	}
}
