package template

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	date := time.Date(2025, time.September, 18, 10, 30, 0, 0, time.UTC)
	data := TemplateData{Date: date}

	// Test case 1: Basic date formatting
	templateString := "Today is {{.Date | formatDate \"YYYY-MM-DD\"}}."
	expected := "Today is 2025-09-18."
	result, err := Render(templateString, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Test case 2: More complex date formatting
	templateString = "{{.Date | formatDate \"MMM DD YYYY dddd\"}}"
	expected = "Sep 18 2025 Thursday"
	result, err = Render(templateString, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Test case 3: Template with literal text and no data
	templateString = "Hello, World!"
	expected = "Hello, World!"
	result, err = Render(templateString, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Test case 4: Invalid template string
	templateString = "{{.Date | invalidFunc \"format\"}}"
	result, err = Render(templateString, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "function \"invalidFunc\" not defined")
}

func TestObsidianToGoFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ISO date", "YYYY-MM-DD", "2006-01-02"},
		{"date with month name", "MMM DD YYYY dddd", "Jan 02 2006 Monday"},
		{"time 24h", "HH:mm", "15:04"},
		{"time 12h with AM/PM", "hh:mm A", "03:04 PM"},
		{"time with seconds", "HH:mm:ss", "15:04:05"},
		{"full date verbose", "dddd, MMMM D, YYYY", "Monday, January 2, 2006"},
		{"2-digit year", "YY-MM-DD", "06-01-02"},
		{"month/day without leading zero", "M/D/YYYY", "1/2/2006"},
		{"escaped text", "YYYY [year] MM [month]", "2006 year 01 month"},
		{"no tokens", "--- / :", "--- / :"},
		{"separators only", "-/: ", "-/: "},
		{"full month name", "MMMM", "January"},
		{"abbreviated weekday", "ddd", "Mon"},
		{"lowercase am/pm", "hh:mm a", "03:04 pm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := obsidianToGoFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
