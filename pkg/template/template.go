package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// obsidianToGoFormat converts an Obsidian/Moment.js date format string to a Go time format string.
// Supported tokens: YYYY, YY, MMMM, MMM, MM, M, DD, D, dddd, ddd, HH, hh, h, mm, m, ss, s, A, a
// Use [text] for literal text that should not be interpreted as tokens.
func obsidianToGoFormat(format string) string {
	// Token mapping ordered by length (longest first) to avoid partial matches
	tokens := []struct {
		token string
		goFmt string
	}{
		{"YYYY", "2006"},
		{"YY", "06"},
		{"MMMM", "January"},
		{"MMM", "Jan"},
		{"MM", "01"},
		{"M", "1"},
		{"DD", "02"},
		{"D", "2"},
		{"dddd", "Monday"},
		{"ddd", "Mon"},
		{"HH", "15"},
		{"hh", "03"},
		{"h", "3"},
		{"mm", "04"},
		{"m", "4"},
		{"ss", "05"},
		{"s", "5"},
		{"A", "PM"},
		{"a", "pm"},
	}

	var result strings.Builder
	i := 0
	for i < len(format) {
		// Handle escaped text in brackets [text]
		if format[i] == '[' {
			end := strings.IndexByte(format[i+1:], ']')
			if end != -1 {
				result.WriteString(format[i+1 : i+1+end])
				i += end + 2
				continue
			}
		}

		matched := false
		for _, t := range tokens {
			if strings.HasPrefix(format[i:], t.token) {
				result.WriteString(t.goFmt)
				i += len(t.token)
				matched = true
				break
			}
		}
		if !matched {
			result.WriteByte(format[i])
			i++
		}
	}
	return result.String()
}

// TemplateData holds the data available for templating.
type TemplateData struct {
	Date    time.Time
	Time    time.Time
	Summary string
	Entry   string
	// Add other fields as needed for templating
}

// Render renders a given template string with the provided data.
func Render(templateString string, data TemplateData) (string, error) {
	// Create a new template and add custom functions
	tmpl := template.New("logbook_template").Funcs(template.FuncMap{
		"formatDate": func(format string, date time.Time) string {
			return date.Format(obsidianToGoFormat(format))
		},
		"formatTime": func(format string, t time.Time) string {
			return t.Format(obsidianToGoFormat(format))
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
