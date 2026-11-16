package app

// Bringing work in and taking it out: a pasted curl command, and the whole
// workspace as JSON.

import (
	"encoding/json"

	"github.com/egoist/mygo/ui"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// addRequest puts a request at the end of the first folder and selects it.
func (a *App) addRequest(request workspace.Request) {
	project := a.doc.ActiveProject()
	if project == nil {
		return
	}
	a.sync()
	if len(project.Folders) == 0 {
		project.Folders = []Folder{{ID: workspace.NewID("folder"), Name: "Requests", Open: true}}
	}

	request.ID = workspace.NewID("req")
	folder := &project.Folders[0]
	folder.Requests = append(folder.Requests, request)
	folder.Open = true

	a.folder, a.index = 0, len(folder.Requests)-1
	a.draft = workspace.Clone(request)
	a.bodyType, a.authType = bodyIndex(a.draft.Body), authIndex(a.draft.Auth.Type)
	a.state = StateEmpty
	a.save()
}

// exportDocument is the workspace as readable JSON, for a backup or another
// machine. It is not the sealed file: this one is meant to be read.
func exportDocument(doc *Document) ([]byte, error) {
	return json.MarshalIndent(doc, "", "  ")
}

// importDialog takes a curl command.
func (a *App) importDialog(c *ui.Context, t *ui.Theme) {
	if !a.importOpen {
		return
	}

	ui.Modal(c, &a.importOpen, func() {
		ui.Column(c).Width(620).Gap(10).Children(func() {
			ui.Text(c, a.importTitle()).FontSize(13).Bold()
			ui.Text(c, a.importHint()).FontSize(12).TextColor(t.TextMuted)

			ui.Segmented(c, &a.importKind, "curl command", "Postman JSON").Label("Import from")

			area := ui.TextArea(c, &a.importText).
				Height(150).
				Font(mono).
				FontSize(12).
				Placeholder("curl 'https://api.example.com/v1/users' -H 'Accept: application/json'")
			area.Label("curl command")
			area.AutoFocus()

			if a.importError != "" {
				ui.Text(c, a.importError).FontSize(12).TextColor(t.Danger)
			}

			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.importOpen = false
				}
				if ui.PrimaryButton(c, "Import").Clicked() {
					a.runImport()
				}
			})
		})
	})
}

// runImport parses what the dialog holds, and keeps the dialog open with the
// reason when it cannot.
func (a *App) runImport() {
	if a.importKind == 1 {
		a.runPostmanImport()
		return
	}

	request, err := ParseCurl(a.importText)
	if err != nil {
		a.importError = err.Error()
		return
	}

	a.addRequest(request)
	a.closeImport()
}

// runPostmanImport reads a collection into a project of its own, since a
// collection is a project: its folders, its variables and its auth.
func (a *App) runPostmanImport() {
	project, err := ParsePostman([]byte(a.importText))
	if err != nil {
		a.importError = err.Error()
		return
	}

	a.snapshot()
	a.doc.Projects = append(a.doc.Projects, project)
	a.doc.Active = project.ID
	a.selectFirst()
	a.state = StateEmpty
	a.save()
	a.closeImport()
}

func (a *App) closeImport() {
	a.importOpen = false
	a.importText = ""
	a.importError = ""
}

func (a *App) importTitle() string {
	if a.importKind == 1 {
		return "Import a Postman collection"
	}
	return "Import from curl"
}

func (a *App) importHint() string {
	if a.importKind == 1 {
		return "Paste the collection JSON. Its folders, requests, variables and auth become a project."
	}
	return "Paste a command. Method, URL, headers, body and basic auth come across."
}

// exportWorkspace writes the document where the user can find it.
func (a *App) exportWorkspace() {
	data, err := exportDocument(a.doc)
	if err != nil {
		a.alert = Alert{Open: true, Title: "The workspace could not be exported", Body: err.Error()}
		return
	}

	path, err := a.writeFile("endpoint-workspace", ".json", data)
	if err != nil {
		a.alert = Alert{Open: true, Title: "The workspace could not be written", Body: err.Error()}
		return
	}

	a.alert = Alert{
		Open:  true,
		Title: "Workspace exported",
		Body:  "Saved to\n" + path + "\n\nIt is readable JSON, so keep it somewhere safe: it holds your tokens.",
	}
}
