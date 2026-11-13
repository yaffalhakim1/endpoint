package app

// Saving what came back, so the app is not a dead end.

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// responseExtension names the file after the content type, so a saved body
// opens in the right program.
func responseExtension(result Result) string {
	for _, header := range result.Headers {
		if strings.EqualFold(header.Key, "Content-Type") {
			value := strings.ToLower(header.Value)
			switch {
			case strings.Contains(value, "json"):
				return ".json"
			case strings.Contains(value, "html"):
				return ".html"
			case strings.Contains(value, "xml"):
				return ".xml"
			case strings.Contains(value, "csv"):
				return ".csv"
			case strings.HasPrefix(value, "text/"):
				return ".txt"
			case strings.Contains(value, "image/png"):
				return ".png"
			}
		}
	}
	return ".bin"
}

// saveResponse writes whatever tab the pane shows.
func (a *App) saveResponse() {
	if a.state != StateLoaded && a.state != StateFailed {
		return
	}

	var data []byte
	var extension string
	if a.responseTab == 1 {
		data = []byte(a.headerText())
		extension = ".txt"
	} else {
		data = a.result.Body
		extension = responseExtension(a.result)
	}

	if len(data) == 0 {
		a.alert = Alert{Open: true, Title: "Nothing to save", Body: "This response has no body."}
		return
	}

	path, err := a.writeFile(fileName(a.title())+"-response", extension, data)
	if err != nil {
		a.alert = Alert{Open: true, Title: "The response could not be saved", Body: err.Error()}
		return
	}

	a.alert = Alert{Open: true, Title: "Response saved", Body: path}
}

// copyAsCurl puts the request on the clipboard as a command, with variables
// filled in.
func (a *App) copyAsCurl(c *ui.Context) {
	vars := []Pair{}
	if project := a.doc.ActiveProject(); project != nil {
		vars = project.Variables()
	}
	c.WriteClipboard(AsCurl(a.draft, vars))
	c.Toast("Copied as a curl command")
}
