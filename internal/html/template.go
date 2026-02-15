package html

import (
	"embed"
	"html/template"
)

//go:embed components/*.gohtml pages/*.gohtml
var templateFS embed.FS

func NewTemplate() (*template.Template, error) {
	return template.ParseFS(templateFS, "components/*.gohtml", "pages/*.gohtml")
}
