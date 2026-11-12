package app

// The response pane: the status line, the body or the headers of the answer.

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"github.com/yaffalhakim1/endpoint/internal/request"
)

func (a *App) response(c *ui.Context, t *ui.Theme) {
	switch a.state {
	case StateEmpty:
		a.emptyResponse(c, t)
	case StateLoading:
		a.loadingResponse(c, t)
	default:
		a.answeredResponse(c, t)
	}
}

func (a *App) emptyResponse(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Grow(1).AlignItems(ui.Center).Justify(ui.Center).Gap(4).Children(func() {
		ui.Text(c, "Nothing sent yet").FontSize(13.5).TextColor(t.TextMuted)
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Paste a URL, then press").FontSize(13).TextColor(t.TextMuted)
			kbd(c, t, "Ctrl")
			kbd(c, t, "Enter")
		})
	})
}

func (a *App) loadingResponse(c *ui.Context, t *ui.Theme) {
	ui.Row(c).
		Height(38).
		PaddingX(14).
		Gap(10).
		AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).
		BorderColor(t.Border).
		Children(func() {
			ui.Spinner(c).Width(14)
			ui.Text(c, "Waiting").FontSize(12).TextColor(t.TextMuted)
		})

	ui.Column(c).Grow(1).Padding(14).Gap(10).Children(func() {
		widths := []float32{38, 62, 45, 70, 30}
		for _, width := range widths {
			ui.Box(c).Height(10).WidthPercent(width).Radius(4).Background(t.Border)
		}
	})
}

func (a *App) answeredResponse(c *ui.Context, t *ui.Theme) {
	result := a.result
	failed := a.state == StateFailed

	ui.Row(c).
		Height(38).
		PaddingX(14).
		Gap(10).
		AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).
		BorderColor(t.Border).
		Children(func() {
			a.statusBadge(c, t, result, failed)

			if result.Duration > 0 {
				ui.Text(c, request.FormatDuration(result.Duration)).FontSize(12).TextColor(t.TextMuted)
			}
			if !failed {
				ui.Text(c, request.FormatSize(len(result.Body))).FontSize(12).TextColor(t.TextMuted)
			}

			ui.Segmented(c, &a.responseTab, "Body", "Headers").Label("Response")

			ui.Spacer(c)

			if ui.Button(c, "Copy").Clicked() {
				c.WriteClipboard(a.copyText())
				c.Toast("Copied")
			}
			if a.responseTab == 0 && ui.Button(c, "Wrap").Clicked() {
				a.wrap = !a.wrap
			}
			if failed && ui.Button(c, "Retry").Clicked() {
				a.startSend()
			}
		})

	if a.responseTab == 1 {
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Text(c, a.headerText()).
				Font(mono).
				FontSize(12.5).
				LineHeight(1.6).
				Selectable().
				NoWrap().
				Padding(12, 14)
		})
		return
	}

	a.responseEditor(c)
}

func (a *App) statusBadge(c *ui.Context, t *ui.Theme, result Result, failed bool) {
	label := request.StatusLabel(result.Code, result.Status)
	if failed {
		label = "Request failed"
	}

	color := a.tok.Success
	switch {
	case failed:
		color = a.tok.Danger
	case result.Code >= 500:
		color = a.tok.Danger
	case result.Code >= 400:
		color = a.tok.Warning
	case result.Code >= 300:
		color = a.tok.Info
	}

	ui.Box(c).
		Padding(2, 8).
		Radius(6).
		Background(tint(color)).
		Children(func() {
			ui.Text(c, label).Font(mono).FontSize(11.5).Bold().TextColor(color)
		})
}

// copyText is the whole answer, however much of it the pane shows.
func (a *App) copyText() string {
	if a.responseTab == 1 {
		return a.headerText()
	}
	if a.result.Err != "" {
		return a.result.Err
	}
	if len(a.result.Body) == 0 {
		return "This response has no body."
	}
	return request.BodyPreview(a.result.Body)
}

// shownText is what the pane shows: the body, cut short when it is huge, or
// the headers.
func (a *App) shownText() string {
	if a.responseTab == 1 {
		return a.headerText()
	}
	return a.bodyText()
}

// maxRendered caps what the editor is asked to lay out: a body far larger
// than this would freeze the window while it styles every line.
const maxRendered = 2 << 20

func (a *App) bodyText() string {
	if a.result.Err != "" {
		return a.result.Err
	}
	if len(a.result.Body) == 0 {
		return "This response has no body."
	}

	if len(a.result.Body) > maxRendered {
		note := "\n\n… the body is " + request.FormatSize(len(a.result.Body)) + "; the first " +
			request.FormatSize(maxRendered) + " are shown. Copy takes all of it."
		return string(a.result.Body[:maxRendered]) + note
	}
	return request.BodyPreview(a.result.Body)
}

// headerText lists the headers the server sent, one per line.
func (a *App) headerText() string {
	if len(a.result.Headers) == 0 {
		return "This response has no headers."
	}

	var out strings.Builder
	for _, header := range a.result.Headers {
		out.WriteString(header.Key)
		out.WriteString(": ")
		out.WriteString(header.Value)
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

// responseEditor shows the answer in the toolkit's code editor, read only:
// JSON folds by its braces, and the tokens keep their colours. A new answer
// or a new wrapping builds a new buffer.
func (a *App) responseEditor(c *ui.Context) {
	if a.answer == nil {
		buffer := editor.NewBuffer(a.bodyText())
		if language := editor.LanguageFor("body.json"); language != nil {
			buffer.SetLanguage(language)
		}
		a.answer = &editor.State{
			Buffer:      buffer,
			ReadOnly:    true,
			LineNumbers: editor.LineNumbersOn,
			SoftWrap:    a.wrap,
		}
	}

	editor.View(c, a.answer).Grow(1).FillWidth()
}
