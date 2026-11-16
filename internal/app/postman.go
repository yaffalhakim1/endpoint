package app

// Importing a Postman collection. Version 2 of the format is a JSON document
// of items, each a folder or a request, with the collection's own variables
// and auth at the top. Our document has the same shape two levels deep, so
// deeper folders flatten into their parent with a joined name.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// postmanCollection is the part of v2.1 we read.
type postmanCollection struct {
	Info struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	} `json:"info"`
	Item     []postmanItem `json:"item"`
	Variable []postmanPair `json:"variable"`
	Auth     *postmanAuth  `json:"auth"`
}

type postmanItem struct {
	Name    string        `json:"name"`
	Item    []postmanItem `json:"item"`
	Request *postmanReq   `json:"request"`
}

type postmanReq struct {
	Method string        `json:"method"`
	Header []postmanPair `json:"header"`
	Body   *postmanBody  `json:"body"`
	URL    *postmanURL   `json:"url"`
	Auth   *postmanAuth  `json:"auth"`
}

type postmanURL struct {
	Raw   string        `json:"raw"`
	Host  []string      `json:"host"`
	Path  []string      `json:"path"`
	Query []postmanPair `json:"query"`
}

type postmanBody struct {
	Mode       string        `json:"mode"`
	Raw        string        `json:"raw"`
	FormData   []postmanPair `json:"formdata"`
	URLEncoded []postmanPair `json:"urlencoded"`
	File       struct {
		Src string `json:"src"`
	} `json:"file"`
	Options struct {
		Raw struct {
			Language string `json:"language"`
		} `json:"raw"`
	} `json:"options"`
}

type postmanAuth struct {
	Type   string        `json:"type"`
	Bearer []postmanPair `json:"bearer"`
	Basic  []postmanPair `json:"basic"`
	APIKey []postmanPair `json:"apikey"`
}

type postmanPair struct {
	Key      string `json:"key"`
	Value    any    `json:"value"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
}

// ParsePostman reads a collection into a project.
func ParsePostman(data []byte) (workspace.Project, error) {
	var collection postmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return workspace.Project{}, errors.New("this is not a Postman collection: " + err.Error())
	}
	if !strings.Contains(collection.Info.Schema, "collection") && len(collection.Item) == 0 {
		return workspace.Project{}, errors.New("this JSON holds no Postman collection")
	}

	name := strings.TrimSpace(collection.Info.Name)
	if name == "" {
		name = "Imported collection"
	}

	project := workspace.Project{
		ID:   workspace.NewID("project"),
		Name: name,
	}

	for _, variable := range collection.Variable {
		if variable.Key == "" {
			continue
		}
		project.Vars = append(project.Vars, workspace.Pair{Key: variable.Key, Value: pairValue(variable)})
	}

	if collection.Auth != nil {
		project.Auth = authOf(*collection.Auth)
	}

	for _, item := range collection.Item {
		project.Folders = append(project.Folders, foldersOf(item, "")...)
	}

	if len(project.Folders) == 0 {
		project.Folders = []workspace.Folder{{ID: workspace.NewID("folder"), Name: "Requests", Open: true}}
	}
	return project, nil
}

// foldersOf turns an item into folders. A folder keeps its requests; folders
// nested inside it flatten with the parent's name in front, since our rail
// shows two levels.
func foldersOf(item postmanItem, prefix string) []workspace.Folder {
	name := strings.TrimSpace(item.Name)
	if prefix != "" && name != "" {
		name = prefix + " / " + name
	}
	if name == "" {
		name = prefix
	}

	if item.Request != nil {
		request := requestOf(*item.Request, item.Name)
		folder := workspace.Folder{ID: workspace.NewID("folder"), Name: name, Open: true}
		folder.Requests = append(folder.Requests, request)
		return []workspace.Folder{folder}
	}

	var folders []workspace.Folder
	var requests []workspace.Request
	for _, child := range item.Item {
		if child.Request != nil {
			requests = append(requests, requestOf(*child.Request, child.Name))
			continue
		}
		folders = append(folders, foldersOf(child, name)...)
	}

	if len(requests) > 0 {
		folder := workspace.Folder{ID: workspace.NewID("folder"), Name: name, Open: true}
		folder.Requests = requests
		folders = append([]workspace.Folder{folder}, folders...)
	}
	return folders
}

// requestOf reads one request: method, headers, query, body and auth.
func requestOf(req postmanReq, name string) workspace.Request {
	request := workspace.Request{
		ID:     workspace.NewID("req"),
		Name:   strings.TrimSpace(name),
		Method: strings.ToUpper(strings.TrimSpace(req.Method)),
		Body:   workspace.BodyNone,
	}
	if request.Method == "" {
		request.Method = "GET"
	}

	for _, header := range req.Header {
		if header.Key == "" || header.Disabled {
			continue
		}
		request.Headers = append(request.Headers, workspace.Pair{Key: header.Key, Value: pairValue(header)})
	}

	if req.URL != nil {
		raw := strings.TrimSpace(req.URL.Raw)
		if raw == "" {
			raw = composeURL(req.URL)
		}

		// The query may be in the address or in its own list; both become
		// params, which is how this app holds them.
		if base, pairs, ok := splitQuery(raw); ok {
			raw = base
			request.Params = append(request.Params, pairs...)
		}
		for _, query := range req.URL.Query {
			if query.Key == "" || query.Disabled {
				continue
			}
			request.Params = append(request.Params, workspace.Pair{Key: query.Key, Value: pairValue(query)})
		}
		request.URL = raw
	}

	if req.Body != nil {
		switch req.Body.Mode {
		case "raw":
			request.Body = workspace.BodyJSON
			request.JSON = req.Body.Raw
		case "formdata", "urlencoded":
			request.Body = workspace.BodyForm
			fields := req.Body.FormData
			if req.Body.Mode == "urlencoded" {
				fields = req.Body.URLEncoded
			}
			for _, field := range fields {
				if field.Key == "" || field.Disabled {
					continue
				}
				request.Form = append(request.Form, workspace.Pair{Key: field.Key, Value: pairValue(field)})
			}
		case "file":
			request.Body = workspace.BodyFile
			request.File = req.Body.File.Src
		}
	}

	if req.Auth != nil {
		request.Auth = authOf(*req.Auth)
	}
	return request
}

// authOf maps Postman's auth block onto the four kinds this app has.
func authOf(auth postmanAuth) workspace.Auth {
	switch auth.Type {
	case "bearer":
		return workspace.Auth{Type: workspace.AuthBearer, Token: valueOf(auth.Bearer, "token")}
	case "basic":
		return workspace.Auth{
			Type:     workspace.AuthBasic,
			Username: valueOf(auth.Basic, "username"),
			Password: valueOf(auth.Basic, "password"),
		}
	case "apikey":
		return workspace.Auth{
			Type:   workspace.AuthKey,
			Header: valueOf(auth.APIKey, "key"),
			Value:  valueOf(auth.APIKey, "value"),
		}
	default:
		return workspace.Auth{Type: workspace.AuthNone}
	}
}

func valueOf(pairs []postmanPair, key string) string {
	for _, pair := range pairs {
		if pair.Key == key {
			return pairValue(pair)
		}
	}
	return ""
}

// pairValue reads a value that Postman sometimes writes as a string and
// sometimes as a number or a bool.
func pairValue(pair postmanPair) string {
	switch value := pair.Value.(type) {
	case string:
		return value
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

// composeURL builds an address from the parts, for collections that wrote the
// URL as host and path instead of a string.
func composeURL(u *postmanURL) string {
	if len(u.Host) == 0 {
		return strings.Join(u.Path, "/")
	}
	return strings.Join(u.Host, ".") + "/" + strings.Join(u.Path, "/")
}
