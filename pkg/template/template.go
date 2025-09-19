package template

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

// TemplateData holds the data available for templating.
type TemplateData struct {
	Date time.Time
	// Add other fields as needed for templating
}

// Render renders a given template string with the provided data.
func Render(templateString string, data TemplateData) (string, error) {
	// Create a new template and add custom functions
	tmpl := template.New("logbook_template").Funcs(template.FuncMap{
		"formatDate": func(format string, date time.Time) string {
			return date.Format(format)
		},
	})

	// Parse the template string
	parsedTmpl, err := tmpl.Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with the provided data
	var buf bytes.Buffer
	if err := parsedTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
