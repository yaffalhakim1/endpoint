package main

// Live checks against dummyjson.com, an open API with products, carts, users
// and a login. They run only when ENDPOINT_LIVE is set, so the normal suite
// stays offline:
//
//	ENDPOINT_LIVE=1 go test -run TestLive -v

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const liveBase = "https://dummyjson.com"

func liveClient() *http.Client {
	return &http.Client{Timeout: 20 * time.Second}
}

func requireLive(t *testing.T) {
	t.Helper()
	if os.Getenv("ENDPOINT_LIVE") == "" {
		t.Skip("set ENDPOINT_LIVE=1 to run the live checks")
	}
}

// liveSend runs a draft the way the app does, with the project's variables.
func liveSend(t *testing.T, request Request) Result {
	t.Helper()
	vars := []Pair{{Key: "base", Value: liveBase}}
	result := Send(context.Background(), request, Auth{}, vars, liveClient())
	if result.Err != "" {
		t.Fatalf("%s %s failed: %s", request.Method, request.URL, result.Err)
	}
	return result
}

func TestLiveListProductsWithParams(t *testing.T) {
	requireLive(t)

	result := liveSend(t, Request{
		Method: "GET",
		URL:    "{{base}}/products",
		Params: []Pair{{Key: "limit", Value: "3"}, {Key: "select", Value: "title,price"}},
	})

	if result.Code != http.StatusOK {
		t.Fatalf("status %d", result.Code)
	}

	var body struct {
		Products []struct {
			Title string  `json:"title"`
			Price float64 `json:"price"`
		} `json:"products"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal([]byte(BodyPreview(result.Body)), &body); err != nil {
		t.Fatalf("the body is not JSON: %v", err)
	}

	if len(body.Products) != 3 {
		t.Errorf("limit=3 returned %d products", len(body.Products))
	}
	if body.Total == 0 {
		t.Error("the answer has no total")
	}
	if body.Products[0].Title == "" || body.Products[0].Price == 0 {
		t.Errorf("select=title,price gave %+v", body.Products[0])
	}
}

func TestLivePathParameter(t *testing.T) {
	requireLive(t)

	result := liveSend(t, Request{Method: "GET", URL: "{{base}}/products/1"})
	if result.Code != http.StatusOK {
		t.Fatalf("status %d", result.Code)
	}
	if !strings.Contains(BodyPreview(result.Body), `"id": 1`) {
		t.Errorf("product 1 came back as:\n%s", BodyPreview(result.Body))
	}
}

func TestLiveCreateUpdateDelete(t *testing.T) {
	requireLive(t)

	created := liveSend(t, Request{
		Method: "POST",
		URL:    "{{base}}/products/add",
		Headers: []Pair{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "Accept", Value: "application/json"},
		},
		Body: BodyJSON,
		JSON: `{"title":"Endpoint test","price":42,"brand":"MyGo"}`,
	})
	if created.Code != http.StatusCreated && created.Code != http.StatusOK {
		t.Fatalf("add returned %d: %s", created.Code, BodyPreview(created.Body))
	}
	if !strings.Contains(BodyPreview(created.Body), "Endpoint test") {
		t.Errorf("the created product is not in the answer:\n%s", BodyPreview(created.Body))
	}

	updated := liveSend(t, Request{
		Method: "PUT",
		URL:    "{{base}}/products/1",
		Body:   BodyJSON,
		JSON:   `{"price":99.5}`,
	})
	if updated.Code != http.StatusOK {
		t.Fatalf("update returned %d", updated.Code)
	}
	if !strings.Contains(BodyPreview(updated.Body), "99.5") {
		t.Errorf("the update did not take:\n%s", BodyPreview(updated.Body))
	}

	deleted := liveSend(t, Request{Method: "DELETE", URL: "{{base}}/products/1"})
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete returned %d", deleted.Code)
	}
}

func TestLiveBearerTokenFromAVariable(t *testing.T) {
	requireLive(t)

	login := liveSend(t, Request{
		Method: "POST",
		URL:    "{{base}}/auth/login",
		Body:   BodyJSON,
		JSON:   `{"username":"emilys","password":"emilyspass","expiresInMins":5}`,
	})
	if login.Code != http.StatusOK {
		t.Fatalf("login returned %d: %s", login.Code, BodyPreview(login.Body))
	}

	token := tokenFrom(t, login.Body)
	if token == "" {
		t.Fatalf("the login answer has no token:\n%s", BodyPreview(login.Body))
	}

	// The token reaches the next request through a variable, as the app does.
	result := Send(
		context.Background(),
		Request{
			Method: "GET",
			URL:    "{{base}}/auth/me",
			Auth:   Auth{Type: AuthBearer, Token: "{{session}}"},
		},
		Auth{},
		[]Pair{{Key: "base", Value: liveBase}, {Key: "session", Value: token}},
		liveClient(),
	)
	if result.Err != "" {
		t.Fatalf("/auth/me failed: %s", result.Err)
	}
	if result.Code != http.StatusOK {
		t.Fatalf("/auth/me returned %d: %s", result.Code, BodyPreview(result.Body))
	}
	if !strings.Contains(BodyPreview(result.Body), "emilys") {
		t.Errorf("/auth/me did not answer as the user:\n%s", BodyPreview(result.Body))
	}
}

func TestLiveNotFoundIsAStatusNotAnError(t *testing.T) {
	requireLive(t)

	result := liveSend(t, Request{Method: "GET", URL: "{{base}}/products/0"})
	if result.Err != "" {
		t.Fatalf("a 404 came back as an error: %s", result.Err)
	}
	if result.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", result.Code)
	}
	if len(result.Headers) == 0 {
		t.Error("the answer has no headers to show")
	}
}

// The pane must be able to render a real answer: status, size, headers, body.
func TestLiveRenderedAnswer(t *testing.T) {
	requireLive(t)

	app := newTestApp(t)
	app.result = liveSend(t, Request{Method: "GET", URL: "{{base}}/products/1"})
	app.state = StateLoaded

	body := app.shownText()
	if !strings.Contains(body, `"title"`) {
		t.Errorf("the body pane shows:\n%s", body)
	}

	app.responseTab = 1
	headers := app.shownText()
	if !strings.Contains(strings.ToLower(headers), "content-type:") {
		t.Errorf("the header pane shows:\n%s", headers)
	}

	if len(app.result.Headers) < 2 {
		t.Errorf("only %d headers parsed", len(app.result.Headers))
	}
}

// tokenFrom reads the access token of a login answer, whichever name it uses.
func tokenFrom(t *testing.T, body []byte) string {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"accessToken", "token"} {
		if value, ok := payload[key].(string); ok {
			return value
		}
	}
	return ""
}
