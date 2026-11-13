package app

// Writing a request back out as a curl command, for a bug report or a
// colleague who would rather run it than click.

import (
	"fmt"
	"strings"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// AsCurl writes a request as a curl command, with the project's variables
// already filled in, so the command runs as it stands.
func AsCurl(request workspace.Request, vars []workspace.Pair) string {
	resolved := workspace.SubstituteRequest(request, vars)

	var out strings.Builder
	out.WriteString("curl")

	method := resolved.Method
	if method == "" {
		method = "GET"
	}
	if method != "GET" {
		out.WriteString(" -X " + method)
	}

	url := resolved.URL
	if len(resolved.Params) > 0 {
		separator := "?"
		if strings.Contains(url, "?") {
			separator = "&"
		}
		parts := make([]string, 0, len(resolved.Params))
		for _, pair := range resolved.Params {
			if pair.Key != "" {
				parts = append(parts, urlEncode(pair.Key)+"="+urlEncode(pair.Value))
			}
		}
		if len(parts) > 0 {
			url += separator + strings.Join(parts, "&")
		}
	}
	out.WriteString(" " + shellQuote(url))

	for _, header := range resolved.Headers {
		if header.Key == "" {
			continue
		}
		out.WriteString(" \\\n  -H " + shellQuote(header.Key+": "+header.Value))
	}

	switch resolved.Auth.Type {
	case workspace.AuthBearer:
		if resolved.Auth.Token != "" {
			out.WriteString(" \\\n  -H " + shellQuote("Authorization: Bearer "+resolved.Auth.Token))
		}
	case workspace.AuthKey:
		name := resolved.Auth.Header
		if name == "" {
			name = "X-API-Key"
		}
		if resolved.Auth.Value != "" {
			out.WriteString(" \\\n  -H " + shellQuote(name+": "+resolved.Auth.Value))
		}
	case workspace.AuthBasic:
		if resolved.Auth.Username != "" {
			out.WriteString(" \\\n  -u " + shellQuote(resolved.Auth.Username+":"+resolved.Auth.Password))
		}
	}

	switch resolved.Body {
	case workspace.BodyJSON:
		if strings.TrimSpace(resolved.JSON) != "" {
			out.WriteString(" \\\n  -H " + shellQuote("Content-Type: application/json"))
			out.WriteString(" \\\n  -d " + shellQuote(resolved.JSON))
		}
	case workspace.BodyForm:
		for _, field := range resolved.Form {
			if field.Key == "" {
				continue
			}
			out.WriteString(" \\\n  -F " + shellQuote(field.Key+"="+field.Value))
		}
	case workspace.BodyFile:
		if resolved.File != "" {
			out.WriteString(" \\\n  --data-binary " + shellQuote("@"+resolved.File))
		}
	}

	return out.String()
}

// shellQuote wraps a value so a shell reads it as one argument.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// urlEncode escapes a query key or value the way a URL needs.
func urlEncode(s string) string {
	var out strings.Builder
	for _, b := range []byte(s) {
		safe := (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') ||
			b == '-' || b == '.' || b == '_' || b == '~'
		if safe {
			out.WriteByte(b)
			continue
		}
		out.WriteString(fmt.Sprintf("%%%02X", b))
	}
	return out.String()
}
