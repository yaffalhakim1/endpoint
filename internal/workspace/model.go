package workspace

// The document the app keeps, and the values it is built from.

// Pair is one editable row: a key and a value, in params, headers or a form.
type Pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AuthType names how a request proves who it is.
type AuthType string

const (
	AuthNone    AuthType = "none"
	AuthBearer  AuthType = "bearer"
	AuthBasic   AuthType = "basic"
	AuthKey     AuthType = "key"
	AuthInherit AuthType = "inherit"
)

// Auth is a request's credentials.
type Auth struct {
	Type     AuthType `json:"type"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Token    string   `json:"token,omitempty"`
	Header   string   `json:"header,omitempty"`
	Value    string   `json:"value,omitempty"`
}

// BodyType names what a request sends, if anything.
type BodyType string

const (
	BodyNone BodyType = "none"
	BodyJSON BodyType = "json"
	BodyForm BodyType = "form"
	BodyFile BodyType = "file"
)

// Request is one call: where it goes and what it carries.
type Request struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Method  string   `json:"method"`
	URL     string   `json:"url"`
	Params  []Pair   `json:"params"`
	Headers []Pair   `json:"headers"`
	Auth    Auth     `json:"auth"`
	Body    BodyType `json:"body"`
	JSON    string   `json:"json,omitempty"`
	Form    []Pair   `json:"form,omitempty"`
	File    string   `json:"file,omitempty"`
}

// Folder groups requests inside a project.
type Folder struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Open     bool      `json:"open"`
	Requests []Request `json:"requests"`
}

// HTTPOptions is how a project reaches the network. Every field is written
// so that its zero value is the safe default.
type HTTPOptions struct {
	// TimeoutSeconds ends a call that hangs: 0 waits as long as the server
	// takes.
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
	// NoRedirects stops at the first 3xx, so its Location shows through.
	NoRedirects bool `json:"noRedirects,omitempty"`
	// InsecureTLS accepts a certificate nothing signed, as a dev host has.
	InsecureTLS bool `json:"insecureTLS,omitempty"`
	// NoCookies forgets what the server sets between calls.
	NoCookies bool `json:"noCookies,omitempty"`
	// Proxy overrides the environment, as http://127.0.0.1:8080.
	Proxy string `json:"proxy,omitempty"`
}

// Environment is a named set of variables, as dev, staging and production
// are, one of which is active at a time.
type Environment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Vars []Pair `json:"vars,omitempty"`
}

// Project is a set of requests sharing credentials, variables and a base
// URL. Variables are written {{name}} in a request and replaced when it is
// sent.
type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Auth         Auth          `json:"auth"`
	HTTP         HTTPOptions   `json:"http,omitempty"`
	Vars         []Pair        `json:"vars,omitempty"`
	Environments []Environment `json:"environments,omitempty"`
	ActiveEnv    string        `json:"activeEnv,omitempty"`
	Folders      []Folder      `json:"folders"`
}

// Recent is one line of the history: enough to run the call again.
type Recent struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	At     string `json:"at"`
}

// Document is everything the app keeps on disk.
type Document struct {
	Version  int       `json:"version"`
	Projects []Project `json:"projects"`
	Active   string    `json:"active,omitempty"`
	Recents  []Recent  `json:"recents,omitempty"`
}

// ActiveProject returns the project the app shows, or nil.
func (d *Document) ActiveProject() *Project {
	for i := range d.Projects {
		if d.Projects[i].ID == d.Active {
			return &d.Projects[i]
		}
	}
	if len(d.Projects) > 0 {
		return &d.Projects[0]
	}
	return nil
}

// CountRequests counts the requests of a project, in its folders.
func (p Project) CountRequests() int {
	n := 0
	for _, folder := range p.Folders {
		n += len(folder.Requests)
	}
	return n
}

// Clone copies a request, so a draft can be edited without touching the
// document until it is saved.
func Clone(req Request) Request {
	out := req
	out.Params = append([]Pair(nil), req.Params...)
	out.Headers = append([]Pair(nil), req.Headers...)
	out.Form = append([]Pair(nil), req.Form...)
	return out
}
