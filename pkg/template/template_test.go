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
	templateString := "Today is {{.Date | formatDate \"2006-01-02\"}}."
	expected := "Today is 2025-09-18."
	result, err := Render(templateString, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Test case 2: More complex date formatting
	templateString = "{{.Date | formatDate \"Jan 02 2006 Monday\"}}"
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
