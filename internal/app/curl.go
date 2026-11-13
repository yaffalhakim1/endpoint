package app

// Importing a curl command: the fastest way to bring work in from a terminal,
// a bug report or a colleague's message.

import (
	"fmt"
	"strings"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// ParseCurl turns a curl command into a request. It understands the flags
// people actually paste: -X, -H, -d and its variants, -F, -u, --url, and the
// line continuations a shell uses.
func ParseCurl(command string) (workspace.Request, error) {
	request := workspace.Request{Method: "GET", Body: workspace.BodyNone}

	args, err := splitCommand(command)
	if err != nil {
		return request, err
	}
	if len(args) == 0 || !strings.HasPrefix(args[0], "curl") {
		return request, fmt.Errorf("this does not start with curl")
	}

	var data []string
	var form []string
	var method string

	for i := 1; i < len(args); i++ {
		arg := args[i]
		next := func() (string, bool) {
			if i+1 >= len(args) {
				return "", false
			}
			i++
			return args[i], true
		}

		switch {
		case arg == "-X" || arg == "--request":
			if value, ok := next(); ok {
				method = strings.ToUpper(value)
			}
		case arg == "-H" || arg == "--header":
			if value, ok := next(); ok {
				if key, val, found := strings.Cut(value, ":"); found {
					request.Headers = append(request.Headers, workspace.Pair{
						Key:   strings.TrimSpace(key),
						Value: strings.TrimSpace(val),
					})
				}
			}
		case arg == "-d" || arg == "--data" || arg == "--data-raw" || arg == "--data-binary" || arg == "--data-ascii":
			if value, ok := next(); ok {
				data = append(data, value)
			}
		case arg == "-F" || arg == "--form":
			if value, ok := next(); ok {
				form = append(form, value)
			}
		case arg == "-u" || arg == "--user":
			if value, ok := next(); ok {
				user, pass, _ := strings.Cut(value, ":")
				request.Auth = workspace.Auth{Type: workspace.AuthBasic, Username: user, Password: pass}
			}
		case arg == "--url":
			if value, ok := next(); ok {
				request.URL = value
			}
		case arg == "-k" || arg == "--insecure":
			// The project's HTTP options carry this, not the request.
		case strings.HasPrefix(arg, "-"):
			// A flag this parser does not model, with or without a value.
			if !strings.HasPrefix(arg, "--") && len(arg) == 2 && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
		default:
			if request.URL == "" {
				request.URL = arg
			}
		}
	}

	switch {
	case len(form) > 0:
		request.Body = workspace.BodyForm
		for _, field := range form {
			key, value, _ := strings.Cut(field, "=")
			request.Form = append(request.Form, workspace.Pair{Key: key, Value: value})
		}
	case len(data) > 0:
		body := strings.Join(data, "&")
		request.Body = workspace.BodyJSON
		request.JSON = body
		if !strings.Contains(body, "{") && !strings.Contains(body, "[") {
			// A form-encoded body, which the JSON editor would mangle.
			request.Body = workspace.BodyForm
			request.JSON = ""
			for _, pair := range strings.Split(body, "&") {
				key, value, _ := strings.Cut(pair, "=")
				request.Form = append(request.Form, workspace.Pair{Key: key, Value: value})
			}
		}
	}

	if method != "" {
		request.Method = method
	} else if request.Body != workspace.BodyNone {
		request.Method = "POST"
	}

	if request.URL == "" {
		return request, fmt.Errorf("the command has no URL")
	}

	request.Name = nameFromURL(request.URL)
	return request, nil
}

// splitCommand breaks a command into arguments, honouring quotes and the
// backslash continuations a pasted command carries.
func splitCommand(command string) ([]string, error) {
	command = strings.ReplaceAll(command, "\\\r\n", " ")
	command = strings.ReplaceAll(command, "\\\n", " ")
	command = strings.ReplaceAll(command, "\r\n", " ")
	command = strings.ReplaceAll(command, "\n", " ")

	var args []string
	var current strings.Builder
	var quote rune
	started := false

	for _, r := range command {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
			started = true
		case r == ' ' || r == '\t':
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
		default:
			current.WriteRune(r)
			started = true
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("a quote is not closed")
	}
	if started {
		args = append(args, current.String())
	}
	return args, nil
}

// nameFromURL gives the imported request something to be called in the rail.
func nameFromURL(raw string) string {
	trimmed := strings.TrimSuffix(raw, "/")
	if i := strings.LastIndex(trimmed, "/"); i >= 0 && i+1 < len(trimmed) {
		last := trimmed[i+1:]
		if !strings.Contains(last, ".") {
			return last
		}
	}
	return "Imported request"
}
