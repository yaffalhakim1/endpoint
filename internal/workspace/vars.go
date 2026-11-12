package workspace

// Variables: a request writes {{name}}, and the project's values replace it
// when the call is made. An unknown name stays as written, so a typo shows up
// in the request rather than vanishing.

import (
	"strings"
)

// substitute replaces every {{name}} in s, allowing spaces inside the braces.
func Substitute(s string, vars []Pair) string {
	if !strings.Contains(s, "{{") || len(vars) == 0 {
		return s
	}

	var out strings.Builder
	rest := s
	for {
		start := strings.Index(rest, "{{")
		if start < 0 {
			out.WriteString(rest)
			return out.String()
		}
		end := strings.Index(rest[start:], "}}")
		if end < 0 {
			out.WriteString(rest)
			return out.String()
		}
		end += start

		name := strings.TrimSpace(rest[start+2 : end])
		out.WriteString(rest[:start])

		if value, ok := Lookup(vars, name); ok {
			out.WriteString(value)
		} else {
			out.WriteString(rest[start : end+2])
		}
		rest = rest[end+2:]
	}
}

func Lookup(vars []Pair, name string) (string, bool) {
	for _, pair := range vars {
		if pair.Key == name {
			return pair.Value, true
		}
	}
	return "", false
}

// substituteRequest resolves a draft against the project's variables, so what
// is sent is what the pane says it will send.
func SubstituteRequest(req Request, vars []Pair) Request {
	out := Clone(req)
	out.URL = Substitute(out.URL, vars)

	for i := range out.Params {
		out.Params[i].Key = Substitute(out.Params[i].Key, vars)
		out.Params[i].Value = Substitute(out.Params[i].Value, vars)
	}
	for i := range out.Headers {
		out.Headers[i].Key = Substitute(out.Headers[i].Key, vars)
		out.Headers[i].Value = Substitute(out.Headers[i].Value, vars)
	}
	for i := range out.Form {
		out.Form[i].Key = Substitute(out.Form[i].Key, vars)
		out.Form[i].Value = Substitute(out.Form[i].Value, vars)
	}

	out.Auth.Username = Substitute(out.Auth.Username, vars)
	out.Auth.Password = Substitute(out.Auth.Password, vars)
	out.Auth.Token = Substitute(out.Auth.Token, vars)
	out.Auth.Header = Substitute(out.Auth.Header, vars)
	out.Auth.Value = Substitute(out.Auth.Value, vars)

	out.File = Substitute(out.File, vars)
	out.JSON = Substitute(out.JSON, vars)
	return out
}
