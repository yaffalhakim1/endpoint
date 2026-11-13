package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yaffalhakim1/endpoint/internal/request"
	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// The transport options decide whether the app works on a real API: a
// session cookie, a self-signed certificate, a redirect, a hung server.

func TestCookiesSurviveBetweenCalls(t *testing.T) {
	var seen string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123", Path: "/"})
			return
		}
		if cookie, err := r.Cookie("session"); err == nil {
			seen = cookie.Value
		}
	}))
	defer server.Close()

	client, err := request.Build(workspace.HTTPOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	first := request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL + "/login"}, workspace.Auth{}, nil, client)
	if first.Err != "" {
		t.Fatalf("login failed: %s", first.Err)
	}

	// The same client, as the app keeps it, carries the cookie.
	second := request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL + "/me"}, workspace.Auth{}, nil, client)
	if second.Err != "" {
		t.Fatalf("the second call failed: %s", second.Err)
	}
	if seen != "abc123" {
		t.Errorf("the server saw cookie %q, want abc123", seen)
	}
}

func TestNoCookiesForgetsTheSession(t *testing.T) {
	var seen string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123", Path: "/"})
			return
		}
		if cookie, err := r.Cookie("session"); err == nil {
			seen = cookie.Value
		}
	}))
	defer server.Close()

	client, err := request.Build(workspace.HTTPOptions{NoCookies: true}, nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL + "/login"}, workspace.Auth{}, nil, client)
	request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL + "/me"}, workspace.Auth{}, nil, client)

	if seen != "" {
		t.Errorf("a cookie was kept although the option is off: %q", seen)
	}
}

func TestSelfSignedCertificateNeedsTheOption(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	strict, err := request.Build(workspace.HTTPOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL}, workspace.Auth{}, nil, strict); got.Err == "" {
		t.Error("a self-signed certificate was accepted without the option")
	}

	relaxed, err := request.Build(workspace.HTTPOptions{InsecureTLS: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL}, workspace.Auth{}, nil, relaxed); got.Err != "" {
		t.Errorf("the option did not accept the certificate: %s", got.Err)
	}
}

func TestRedirectsAreFollowedUnlessTurnedOff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old" {
			http.Redirect(w, r, "/new", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("arrived"))
	}))
	defer server.Close()

	following, err := request.Build(workspace.HTTPOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL + "/old"}, workspace.Auth{}, nil, following)
	if got.Code != http.StatusOK || string(got.Body) != "arrived" {
		t.Errorf("the redirect was not followed: %d %q", got.Code, got.Body)
	}

	stopping, err := request.Build(workspace.HTTPOptions{NoRedirects: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got = request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL + "/old"}, workspace.Auth{}, nil, stopping)
	if got.Code != http.StatusFound {
		t.Errorf("the redirect was followed anyway: %d", got.Code)
	}
}

func TestTimeoutEndsAHungCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client, err := request.Build(workspace.HTTPOptions{TimeoutSeconds: 1}, nil)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	got := request.Send(t.Context(), workspace.Request{Method: "GET", URL: server.URL}, workspace.Auth{}, nil, client)
	if got.Err == "" {
		t.Fatal("the call did not time out")
	}
	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Errorf("the timeout took %s", elapsed)
	}
}

func TestABadProxyIsReported(t *testing.T) {
	if _, err := request.Build(workspace.HTTPOptions{Proxy: "not a url"}, nil); err == nil {
		t.Error("a bad proxy address was accepted")
	}
}
