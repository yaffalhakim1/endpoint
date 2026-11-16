package app

// A URL pasted with its query string keeps working, but the parameters belong
// in the Params tab where they can be edited. So a URL that arrives with a
// query is split: the address stays, and the pairs move to the editor.

import (
	"net/url"
	"strings"
)

// splitQuery takes the query off a URL and returns the address, the pairs it
// held in the order they were written, and whether there were any.
func splitQuery(raw string) (string, []Pair, bool) {
	question := strings.Index(raw, "?")
	if question < 0 {
		return raw, nil, false
	}

	base := raw[:question]
	rest := raw[question+1:]

	// A fragment belongs to the address, not to the query.
	fragment := ""
	if hash := strings.Index(rest, "#"); hash >= 0 {
		fragment = rest[hash:]
		rest = rest[:hash]
	}

	pairs := parseQuery(rest)
	if len(pairs) == 0 {
		return raw, nil, false
	}
	return base + fragment, pairs, true
}

// parseQuery reads a query string, keeping the order the pairs were written
// in, unlike url.Values which is a map.
func parseQuery(query string) []Pair {
	var pairs []Pair

	for _, field := range strings.Split(query, "&") {
		if field == "" {
			continue
		}

		key, value, _ := strings.Cut(field, "=")
		name, err := url.QueryUnescape(key)
		if err != nil {
			name = key
		}
		decoded, err := url.QueryUnescape(value)
		if err != nil {
			decoded = value
		}
		pairs = append(pairs, Pair{Key: name, Value: decoded})
	}
	return pairs
}

// mergeParams puts the pairs into the draft, replacing a key that is already
// there and appending the rest, so pasting the same URL twice does not double
// every row.
func (a *App) mergeParams(pairs []Pair) {
	for _, incoming := range pairs {
		replaced := false
		for i := range a.draft.Params {
			if a.draft.Params[i].Key == incoming.Key {
				a.draft.Params[i].Value = incoming.Value
				replaced = true
				break
			}
		}
		if !replaced {
			a.draft.Params = append(a.draft.Params, incoming)
		}
	}
}

// absorbQuery moves a query string out of the draft URL into the Params tab.
// It is called when the URL field reports a change, which is when a paste
// lands.
func (a *App) absorbQuery() {
	base, pairs, ok := splitQuery(a.draft.URL)
	if !ok {
		return
	}
	a.draft.URL = base
	a.mergeParams(pairs)
}
