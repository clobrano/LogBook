package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, filepath.Join(os.Getenv("HOME"), ".logbook", "journal"), cfg.JournalDir)
	assert.Equal(t, "{{.Date | formatDate \"2006-01-02\"}}.md", cfg.DailyFileName)
	assert.Equal(t, "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n", cfg.DailyTemplate)
	assert.False(t, cfg.AIEnabled)
	assert.Equal(t, "Write a summary of the note at the given file. Use 1st person and a simple language. Use 200 characters or less", cfg.AIPrompt)
	assert.Equal(t, "{{.Date | formatDate \"2006-01-02\"}}: {{.Summary}}", cfg.OneLineTemplate)
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile := filepath.Join(t.TempDir(), "config.toml")

	expectedConfig := &Config{
		JournalDir:      "/tmp/myjournal",
		DailyFileName:   "DD-MM-YYYY.md",
		DailyTemplate:   "## {{.Date | formatDate \"Monday, January 2, 2006\"}}\n",
		AIEnabled:       true,
		AIPrompt:        "Summarize this entry.",
		OneLineTemplate: "{{.Date | formatDate \"01/02\"}} - {{.Summary}}",
	}

	err := SaveConfig(tmpfile, expectedConfig)
	assert.NoError(t, err)

	// Load the config
	loadedConfig, err := LoadConfig(tmpfile)
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, loadedConfig)
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile := filepath.Join(t.TempDir(), "config.toml")

	cfg := DefaultConfig()
	cfg.JournalDir = "/path/to/journal"
	cfg.AIEnabled = true

	err := SaveConfig(tmpfile, cfg)
	assert.NoError(t, err)

	// Read the file content and verify
	content, err := os.ReadFile(tmpfile)
	assert.NoError(t, err)

	expectedContent := `journal_dir = "/path/to/journal"
daily_file_name = "{{.Date | formatDate \"2006-01-02\"}}.md"
daily_template = "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n"
ai_enabled = true
ai_prompt = "Write a summary of the note at the given file. Use 1st person and a simple language. Use 200 characters or less"
one_line_template = "{{.Date | formatDate \"2006-01-02\"}}: {{.Summary}}"
`
	assert.Equal(t, expectedContent, string(content))
}

func TestConfigValidate(t *testing.T) {
	// Test valid config
	cfg := DefaultConfig()
	assert.NoError(t, cfg.Validate())

	// Test empty JournalDir
	cfg.JournalDir = ""
	assert.ErrorContains(t, cfg.Validate(), "JournalDir cannot be empty")
	cfg = DefaultConfig() // Reset

	// Test empty DailyFileName
	cfg.DailyFileName = ""
	assert.ErrorContains(t, cfg.Validate(), "DailyFileName cannot be empty")
	cfg = DefaultConfig() // Reset

	// Test empty DailyTemplate
	cfg.DailyTemplate = ""
	assert.ErrorContains(t, cfg.Validate(), "DailyTemplate cannot be empty")
	cfg = DefaultConfig() // Reset

	// Test AI enabled with empty AIPrompt
	cfg.AIEnabled = true
	cfg.AIPrompt = ""
	assert.ErrorContains(t, cfg.Validate(), "AIPrompt cannot be empty if AI is enabled")
	cfg = DefaultConfig() // Reset
}
