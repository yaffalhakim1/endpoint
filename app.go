package main

// The app's state and actions. Everything here runs on the main thread, so
// there is no locking: the goroutine that sends a request carries a copy and
// posts its result back with Window.Update.

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
)

// State is what the response pane shows.
type State string

const (
	StateEmpty   State = "empty"
	StateLoading State = "loading"
	StateLoaded  State = "loaded"
	StateFailed  State = "failed"
)

// The choices of the two segmented controls, in the order they are drawn.
var bodyTypes = []BodyType{BodyJSON, BodyForm, BodyFile, BodyNone}
var bodyLabels = []string{"JSON", "Form", "File", "None"}
var authTypes = []AuthType{AuthNone, AuthBearer, AuthBasic, AuthKey, AuthInherit}
var authLabels = []string{"None", "Bearer", "Basic", "API key", "Inherit"}

// App holds the document, the draft the window edits and the last answer.
type App struct {
	dir    string
	win    *mygo.Window
	client *http.Client

	doc           *Document
	draft         Request
	folder, index int

	state  State
	result Result

	// requestHeight is the height of the request pane, which the divider sets.
	requestHeight float32

	// tok is this frame's colours, by role.
	tok tokens

	// The dialogs that edit the tree, and what they are about.
	dialog       dialogKind
	dialogOpen   bool
	dialogValue  string
	dialogFolder int
	dialogIndex  int

	// responseTab is 0 for the body and 1 for the headers.
	responseTab int

	sidebarHidden bool
	copyRequested bool

	// recovered says what happened to the workspace when it opened, and
	// alert is the message the window shows about it.
	recovered recovery
	alert     Alert

	// answer is the read-only editor of the last response, with its folding.
	answer *editor.State

	// body is the JSON editor of the request the window shows.
	body    *editor.State
	bodyFor string

	// cancel stops the call in flight. varsOpen shows the project's
	// variables, and filter narrows the rail.
	cancel   context.CancelFunc
	varsOpen bool
	filter   string

	tab      int
	bodyType int
	authType int

	themeMode int // 0 follows the system, 1 light, 2 dark

	wrap bool

	dirty    bool
	editedAt time.Time
}

// NewApp loads the document and starts on its first request. Only a broken
// install, not a broken workspace, stops it: the workspace is recovered from
// its backup or started fresh, and the user is told in the window.
func NewApp(dir string) (*App, error) {
	if err := registerFonts(); err != nil {
		return nil, err
	}

	doc, recovered := loadDocument(dir)

	app := &App{dir: dir, doc: doc, client: &http.Client{}, state: StateEmpty, requestHeight: 200, recovered: recovered}
	app.selectFirst()
	return app, nil
}

// Document is the whole document, for the rail.
func (a *App) Document() *Document { return a.doc }

// Project is the project the rail shows.
func (a *App) Project() Project {
	if p := a.doc.ActiveProject(); p != nil {
		return *p
	}
	return Project{Name: "Personal"}
}

// State and Result are what the response pane reads.
func (a *App) State() State   { return a.state }
func (a *App) Result() Result { return a.result }

// SelectProject switches the project and picks its first request.
func (a *App) SelectProject(id string) {
	a.sync()
	a.doc.Active = id
	a.selectFirst()
	a.state = StateEmpty
	a.save()
}

// SelectRequest loads a saved request into the draft.
func (a *App) SelectRequest(folder, request int) {
	project := a.doc.ActiveProject()
	if project == nil || folder < 0 || folder >= len(project.Folders) {
		return
	}
	requests := project.Folders[folder].Requests
	if request < 0 || request >= len(requests) {
		return
	}

	a.sync()
	a.folder, a.index = folder, request
	a.draft = cloneRequest(requests[request])
	a.state = StateEmpty
	a.bodyType = bodyIndex(a.draft.Body)
	a.authType = authIndex(a.draft.Auth.Type)
}

// NewRequest adds a request to the active project and selects it.
func (a *App) NewRequest() error {
	project := a.doc.ActiveProject()
	if project == nil {
		return fmt.Errorf("there is no project")
	}
	a.sync()
	if len(project.Folders) == 0 {
		project.Folders = []Folder{{ID: newID("folder"), Name: "Requests", Open: true}}
	}

	folder := &project.Folders[0]
	folder.Requests = append(folder.Requests, Request{
		ID:     newID("req"),
		Name:   "New request",
		Method: "GET",
		Body:   BodyNone,
	})

	a.folder, a.index = 0, len(folder.Requests)-1
	a.draft = cloneRequest(folder.Requests[a.index])
	a.state = StateEmpty
	a.bodyType, a.authType = 3, 0
	a.save()
	return nil
}

// ClearRecents empties the history.
func (a *App) ClearRecents() error {
	a.doc.Recents = nil
	a.save()
	return nil
}

// SetFolderOpen remembers whether a folder shows its requests.
func (a *App) SetFolderOpen(id string, open bool) {
	for p := range a.doc.Projects {
		for f := range a.doc.Projects[p].Folders {
			if a.doc.Projects[p].Folders[f].ID == id {
				a.doc.Projects[p].Folders[f].Open = open
				a.save()
				return
			}
		}
	}
}

// startSend copies the draft, marks the pane busy and sends on its own
// goroutine. The result comes back on the main thread.
func (a *App) startSend() {
	if a.state == StateLoading || a.win == nil {
		return
	}

	draft := cloneRequest(a.draft)
	inherit, vars := Auth{}, []Pair{}
	if project := a.doc.ActiveProject(); project != nil {
		inherit, vars = project.Auth, project.Variables()
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.state = StateLoading
	a.result = Result{}
	a.save()

	go func() {
		result := Send(ctx, draft, inherit, vars, a.client)
		a.win.Update(func() {
			a.cancel = nil
			a.answer = nil // a new answer gets a new buffer, and its folding
			a.result = result
			if result.Err != "" {
				a.state = StateFailed
			} else {
				a.state = StateLoaded
			}
			a.remember(draft)
			a.save()
		})
	}()
}

// Cancel stops the call that is running.
func (a *App) Cancel() {
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
}

// relocate puts the selection back on the request with this ID, wherever it
// sits after the tree changed shape, or on the first request when that one is
// gone.
func (a *App) relocate(id string) {
	project := a.doc.ActiveProject()
	if project != nil && id != "" {
		for f := range project.Folders {
			for r := range project.Folders[f].Requests {
				if project.Folders[f].Requests[r].ID == id {
					a.folder, a.index = f, r
					return
				}
			}
		}
	}
	a.selectFirst()
	a.state = StateEmpty
}

// loadSelection reads the request at the current place into the draft.
func (a *App) loadSelection() {
	project := a.doc.ActiveProject()
	if project == nil || a.folder < 0 || a.folder >= len(project.Folders) {
		return
	}
	requests := project.Folders[a.folder].Requests
	if a.index < 0 || a.index >= len(requests) {
		return
	}
	a.draft = cloneRequest(requests[a.index])
	a.bodyType, a.authType = bodyIndex(a.draft.Body), authIndex(a.draft.Auth.Type)
}

// detach stops the draft from being written back: it is called before the
// tree changes shape, so a stale index cannot overwrite another request.
func (a *App) detach() { a.folder, a.index = -1, -1 }

// sync writes the draft back into the document.
func (a *App) sync() {
	project := a.doc.ActiveProject()
	if project == nil || a.folder < 0 || a.folder >= len(project.Folders) {
		return
	}
	requests := project.Folders[a.folder].Requests
	if a.index < len(requests) {
		requests[a.index] = cloneRequest(a.draft)
	}
}

// save writes the document, at most once per edit burst.
func (a *App) save() {
	a.sync()
	if err := saveDocument(a.dir, a.doc); err != nil {
		a.state = StateFailed
		a.result = Result{Err: "cannot save the workspace: " + err.Error()}
	}
	a.dirty = false
}

// touch marks the draft as edited, and asks for a frame to save it later.
func (a *App) touch(now time.Time) {
	a.dirty = true
	a.editedAt = now
}

// saveIfDue writes the draft once the user stopped typing for a moment.
func (a *App) saveIfDue(now time.Time) bool {
	if !a.dirty {
		return false
	}
	if now.Sub(a.editedAt) < 700*time.Millisecond {
		return false
	}
	a.save()
	return true
}

func (a *App) selectFirst() {
	project := a.doc.ActiveProject()
	if project == nil {
		a.draft = Request{Method: "GET", Body: BodyNone}
		return
	}
	for f := range project.Folders {
		if len(project.Folders[f].Requests) > 0 {
			a.folder, a.index = f, 0
			a.draft = cloneRequest(project.Folders[f].Requests[0])
			a.bodyType = bodyIndex(a.draft.Body)
			a.authType = authIndex(a.draft.Auth.Type)
			return
		}
	}
	a.folder, a.index = 0, 0
	a.draft = Request{Method: "GET", Body: BodyNone}
	a.bodyType, a.authType = 3, 0
}

// remember keeps the last calls, newest first, without duplicates.
func (a *App) remember(req Request) {
	entry := Recent{Method: req.Method, URL: req.URL, At: time.Now().Format("15:04")}
	recents := []Recent{entry}
	for _, old := range a.doc.Recents {
		if old.Method == entry.Method && old.URL == entry.URL {
			continue
		}
		recents = append(recents, old)
		if len(recents) == 5 {
			break
		}
	}
	a.doc.Recents = recents
}

func bodyIndex(t BodyType) int {
	for i, v := range bodyTypes {
		if v == t {
			return i
		}
	}
	return 0
}

func authIndex(t AuthType) int {
	for i, v := range authTypes {
		if v == t {
			return i
		}
	}
	return 0
}

func cloneRequest(req Request) Request {
	out := req
	out.Params = append([]Pair(nil), req.Params...)
	out.Headers = append([]Pair(nil), req.Headers...)
	out.Form = append([]Pair(nil), req.Form...)
	return out
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
