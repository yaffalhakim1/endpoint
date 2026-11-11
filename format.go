package main

// Formatting of what the response pane shows.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PrettyJSON indents a JSON body; ok is false when it is not JSON.
func PrettyJSON(body []byte) (string, bool) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return "", false
	}
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return "", false
	}

	var out bytes.Buffer
	if err := json.Indent(&out, trimmed, "", "  "); err != nil {
		return "", false
	}
	return out.String(), true
}

// FormatSize writes a byte count the way the status bar wants it.
func FormatSize(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}

// FormatDuration writes a time the way the status bar wants it.
func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%d µs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%d ms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.1f s", d.Seconds())
	}
}

// StatusLabel reads as "201 Created", and "Request failed" for no answer.
func StatusLabel(code int, status string) string {
	if status != "" {
		return status
	}
	if code == 0 {
		return "Request failed"
	}
	return fmt.Sprintf("HTTP %d", code)
}

// BodyPreview decodes the body for display, falling back to raw text.
func BodyPreview(body []byte) string {
	if pretty, ok := PrettyJSON(body); ok {
		return pretty
	}
	text := string(body)
	if strings.ContainsRune(text, 0) {
		return fmt.Sprintf("<%s>", FormatSize(len(body)))
	}
	return text
}
