package app

import (
	"context"
	"github.com/yaffalhakim1/endpoint/internal/request"
	"github.com/yaffalhakim1/endpoint/internal/workspace"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Variables, cancel and the filter: the three behaviours the rail and the
// request bar now depend on.

func TestSubstitute(t *testing.T) {
	vars := []Pair{
		{Key: "baseUrl", Value: "https://api.example.com"},
		{Key: "token", Value: "abc"},
	}

	cases := map[string]string{
		"{{baseUrl}}/v1/users": "https://api.example.com/v1/users",
		"Bearer {{ token }}":   "Bearer abc",
		"{{missing}}/x":        "{{missing}}/x",
		"no variables here":    "no variables here",
		"{{baseUrl}}{{token}}": "https://api.example.comabc",
		"{{baseUrl":            "{{baseUrl",
	}

	for in, want := range cases {
		if got := workspace.Substitute(in, vars); got != want {
			t.Errorf("workspace.Substitute(%q) = %q, want %q", in, got, want)
		}
	}

	// With no variables at all, nothing is touched.
	if got := workspace.Substitute("{{baseUrl}}/x", nil); got != "{{baseUrl}}/x" {
		t.Errorf("substitute with no variables = %q", got)
	}
}

func TestSendResolvesProjectVariables(t *testing.T) {
	var gotPath, gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
	}))
	defer server.Close()

	result := request.Send(
		context.Background(),
		Request{
			Method: "GET",
			URL:    "{{base}}/v1/users",
			Auth:   Auth{Type: AuthBearer, Token: "{{token}}"},
		},
		Auth{},
		[]Pair{
			{Key: "base", Value: server.URL},
			{Key: "token", Value: "secret"},
		},
		server.Client(),
	)

	if result.Err != "" {
		t.Fatalf("the call failed: %s", result.Err)
	}
	if gotPath != "/v1/users" {
		t.Errorf("reached %q, want /v1/users", gotPath)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization is %q, want the resolved token", gotAuth)
	}
}

func TestCancelingStopsTheCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := request.Send(ctx, Request{Method: "GET", URL: server.URL}, Auth{}, nil, server.Client())
	if !strings.Contains(result.Err, "canceled") {
		t.Fatalf("the call reported %q, want it canceled", result.Err)
	}
}

func TestFilterMatchesNameURLAndMethod(t *testing.T) {
	request := Request{Name: "Create user", Method: "POST", URL: "https://api.example.com/v1/users"}

	for _, match := range []string{"create", "POST", "/v1/", "  users  "} {
		if !matchesFilter(request, match) {
			t.Errorf("%q did not match", match)
		}
	}
	for _, miss := range []string{"invoice", "GET"} {
		if matchesFilter(request, miss) {
			t.Errorf("%q matched but should not", miss)
		}
	}
	if !matchesFilter(request, "") {
		t.Error("an empty filter hid a request")
	}
}
