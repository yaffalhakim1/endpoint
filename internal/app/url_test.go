package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A pasted URL keeps its query, but the pairs belong in the Params tab.

func TestPastingAQueryURLFillsTheParams(t *testing.T) {
	app := newTestApp(t)

	app.draft.URL = "/products?sortBy=title&order=asc&limit=3"
	app.absorbQuery()

	if app.draft.URL != "/products" {
		t.Errorf("the address is %q, want the query taken off", app.draft.URL)
	}
	if len(app.draft.Params) != 3 {
		t.Fatalf("%d params, want 3: %+v", len(app.draft.Params), app.draft.Params)
	}
	if app.draft.Params[0].Key != "sortBy" || app.draft.Params[0].Value != "title" {
		t.Errorf("the first row is %+v", app.draft.Params[0])
	}
	if app.draft.Params[2].Key != "limit" || app.draft.Params[2].Value != "3" {
		t.Errorf("the third row is %+v", app.draft.Params[2])
	}
}

func TestAbsorbingTwiceDoesNotDoubleTheRows(t *testing.T) {
	app := newTestApp(t)

	app.draft.URL = "/products?limit=3"
	app.absorbQuery()
	app.draft.URL = "/products?limit=3"
	app.absorbQuery()

	if len(app.draft.Params) != 1 {
		t.Errorf("%d params after pasting twice, want 1", len(app.draft.Params))
	}
}

func TestSplitQueryKeepsOrderAndDecodes(t *testing.T) {
	base, pairs, ok := splitQuery("https://api.example.com/search?q=ada+lovelace&tag=a%2Fb")
	if !ok {
		t.Fatal("the query was not found")
	}
	if base != "https://api.example.com/search" {
		t.Errorf("the address is %q", base)
	}
	if pairs[0].Value != "ada lovelace" {
		t.Errorf("the plus was not decoded: %q", pairs[0].Value)
	}
	if pairs[1].Value != "a/b" {
		t.Errorf("the escape was not decoded: %q", pairs[1].Value)
	}
}

func TestSplitQueryIgnoresAURLWithoutOne(t *testing.T) {
	base, pairs, ok := splitQuery("https://api.example.com/products")
	if ok || len(pairs) != 0 {
		t.Error("a URL without a query was split")
	}
	if base != "https://api.example.com/products" {
		t.Errorf("the address changed to %q", base)
	}
}

func TestSplitQueryKeepsTheFragment(t *testing.T) {
	base, pairs, ok := splitQuery("https://api.example.com/x?a=1#top")
	if !ok {
		t.Fatal("the query was not found")
	}
	if base != "https://api.example.com/x#top" {
		t.Errorf("the address is %q, want the fragment kept", base)
	}
	if len(pairs) != 1 || pairs[0].Key != "a" {
		t.Errorf("the pairs are %+v", pairs)
	}
}

func TestMergingParamsReplacesRatherThanDoubles(t *testing.T) {
	app := newTestApp(t)
	app.draft.Params = []Pair{{Key: "limit", Value: "10"}, {Key: "active", Value: "true"}}

	app.mergeParams([]Pair{{Key: "limit", Value: "3"}, {Key: "sortBy", Value: "title"}})

	if len(app.draft.Params) != 3 {
		t.Fatalf("%d params after merging, want 3", len(app.draft.Params))
	}
	if app.draft.Params[0].Value != "3" {
		t.Errorf("limit is %q, want the pasted value", app.draft.Params[0].Value)
	}
	if app.draft.Params[1].Key != "active" {
		t.Errorf("the existing row was disturbed: %+v", app.draft.Params[1])
	}
	if app.draft.Params[2].Key != "sortBy" {
		t.Errorf("the new row was not appended: %+v", app.draft.Params[2])
	}
}

// Splitting must not change what is sent: the params rebuild the same query.
func TestSplitThenSendStillCarriesTheQuery(t *testing.T) {
	var seen string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
	}))
	defer server.Close()

	request := Request{
		Method: "GET",
		URL:    server.URL + "/products",
		Params: []Pair{
			{Key: "sortBy", Value: "title"},
			{Key: "order", Value: "asc"},
			{Key: "limit", Value: "3"},
		},
	}

	if result := sendForTest(request, Auth{}, server.Client()); result.Err != "" {
		t.Fatalf("the call failed: %s", result.Err)
	}

	for _, want := range []string{"sortBy=title", "order=asc", "limit=3"} {
		if !strings.Contains(seen, want) {
			t.Errorf("the query %q lacks %q", seen, want)
		}
	}
}
