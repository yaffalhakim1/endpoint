package request

// Package request sends a call the way the app means it: itself, so there is
// no CORS, any header is allowed, and every platform behaves the same.

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// maxBody caps what the app reads of a response, in both directions.
const maxBody = 64 << 20

// Result is what one call produced.
type Result struct {
	Code     int
	Status   string
	Duration time.Duration
	Body     []byte
	Headers  []workspace.Pair
	Err      string
}

// Send runs a request. A failing HTTP status is a result, not an error; Err
// is set only when the call itself failed.
func Send(ctx context.Context, req workspace.Request, inherit workspace.Auth, vars []workspace.Pair, client *http.Client) Result {
	out, err := buildRequest(req, inherit, vars)
	if err != nil {
		return Result{Err: err.Error()}
	}

	if ctx != nil {
		out = out.WithContext(ctx)
	}

	start := time.Now()
	res, err := client.Do(out)
	if err != nil {
		return Result{Duration: time.Since(start), Err: describeCallError(ctx, err)}
	}
	defer res.Body.Close()

	elapsed := time.Since(start)
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBody))
	if err != nil {
		return Result{Duration: elapsed, Err: err.Error()}
	}

	return Result{
		Code:     res.StatusCode,
		Status:   res.Status,
		Duration: elapsed,
		Body:     body,
		Headers:  headerPairs(res.Header),
	}
}

func buildRequest(req workspace.Request, inherit workspace.Auth, vars []workspace.Pair) (*http.Request, error) {
	resolved := workspace.SubstituteRequest(req, vars)
	raw := strings.TrimSpace(resolved.URL)
	if raw == "" {
		return nil, fmt.Errorf("the URL is empty")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	target, err := url.Parse(raw)
	if err != nil || target.Host == "" {
		return nil, fmt.Errorf("%q is not a URL", resolved.URL)
	}

	if len(resolved.Params) > 0 {
		query := target.Query()
		for _, p := range resolved.Params {
			if p.Key != "" {
				query.Set(p.Key, p.Value)
			}
		}
		target.RawQuery = query.Encode()
	}

	body, contentType, err := buildBody(resolved)
	if err != nil {
		return nil, err
	}

	method := resolved.Method
	if method == "" {
		method = http.MethodGet
	}

	out, err := http.NewRequest(method, target.String(), body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		out.Header.Set("Content-Type", contentType)
	}
	for _, h := range resolved.Headers {
		if h.Key != "" {
			out.Header.Set(h.Key, h.Value)
		}
	}

	applyAuth(out, resolveAuth(resolved.Auth, inherit))
	return out, nil
}

// resolveAuth picks the credentials a request sends: its own, or the
// project's when it inherits.
func resolveAuth(own workspace.Auth, inherit workspace.Auth) workspace.Auth {
	if own.Type == "" || own.Type == workspace.AuthInherit {
		return inherit
	}
	return own
}

func applyAuth(req *http.Request, auth workspace.Auth) {
	switch auth.Type {
	case workspace.AuthBearer:
		if auth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		}
	case workspace.AuthBasic:
		if auth.Username != "" || auth.Password != "" {
			raw := auth.Username + ":" + auth.Password
			req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(raw)))
		}
	case workspace.AuthKey:
		name := auth.Header
		if name == "" {
			name = "X-API-Key"
		}
		if auth.Value != "" {
			req.Header.Set(name, auth.Value)
		}
	}
}

// buildBody returns the body, its content type, and any problem reading a
// file it names.
func buildBody(req workspace.Request) (io.Reader, string, error) {
	switch req.Body {
	case workspace.BodyJSON:
		if strings.TrimSpace(req.JSON) == "" {
			return nil, "", nil
		}
		return strings.NewReader(req.JSON), "application/json", nil
	case workspace.BodyForm:
		return buildMultipart(req)
	case workspace.BodyFile:
		return buildFileBody(req.File)
	default:
		return nil, "", nil
	}
}

func buildMultipart(req workspace.Request) (io.Reader, string, error) {
	var buffer bytes.Buffer
	form := multipart.NewWriter(&buffer)

	for _, field := range req.Form {
		if field.Key == "" {
			continue
		}
		if err := form.WriteField(field.Key, field.Value); err != nil {
			return nil, "", err
		}
	}

	if req.File != "" {
		if err := addFilePart(form, "file", req.File); err != nil {
			return nil, "", err
		}
	}

	if err := form.Close(); err != nil {
		return nil, "", err
	}
	return &buffer, form.FormDataContentType(), nil
}

func addFilePart(form *multipart.Writer, field string, path string) error {
	data, err := readFile(path)
	if err != nil {
		return err
	}
	part, err := form.CreateFormFile(field, pathBase(path))
	if err != nil {
		return err
	}
	_, err = part.Write(data)
	return err
}

func buildFileBody(path string) (io.Reader, string, error) {
	if path == "" {
		return nil, "", nil
	}
	data, err := readFile(path)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(data), "application/octet-stream", nil
}

func readFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", err)
	}
	if info.Size() > maxBody {
		return nil, fmt.Errorf("the file is larger than %s", FormatSize(maxBody))
	}
	return os.ReadFile(path)
}

func pathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// headerPairs flattens a response's headers, sorted by name, as the pane shows
// them and Copy puts them on the clipboard.
func headerPairs(header http.Header) []workspace.Pair {
	names := make([]string, 0, len(header))
	for name := range header {
		names = append(names, name)
	}
	sort.Strings(names)

	pairs := make([]workspace.Pair, 0, len(names))
	for _, name := range names {
		for _, value := range header[name] {
			pairs = append(pairs, workspace.Pair{Key: name, Value: value})
		}
	}
	return pairs
}

// describeCallError tells a canceled call apart from a failed one.
func describeCallError(ctx context.Context, err error) string {
	if ctx != nil && ctx.Err() != nil {
		return "the request was canceled"
	}
	return err.Error()
}
