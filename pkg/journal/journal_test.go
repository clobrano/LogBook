package journal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clobrano/LogBook/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestCreateDailyJournalFile(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"

	// Test case 1: File does not exist, should create successfully
	date := time.Date(2025, time.September, 18, 0, 0, 0, 0, time.UTC)
	expectedFilePath := filepath.Join(tmpDir, "2025-09-18.md")

	filePath, err := CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)
	assert.Equal(t, expectedFilePath, filePath)
	assert.FileExists(t, expectedFilePath)

	// Test case 2: File already exists, should return existing path without error
	filePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)
	assert.Equal(t, expectedFilePath, filePath)
	assert.FileExists(t, expectedFilePath)

	// Test case 3: Invalid configuration (empty JournalDir)
	invalidCfg := config.DefaultConfig()
	invalidCfg.JournalDir = ""
	filePath, err = CreateDailyJournalFile(invalidCfg, date)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration: JournalDir cannot be empty")

	// Test case 4: Non-absolute JournalDir
	invalidCfg = config.DefaultConfig()
	invalidCfg.JournalDir = "./relative/path"
	filePath, err = CreateDailyJournalFile(invalidCfg, date)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JournalDir must be an absolute path")

	// Test case 5: Non-existent JournalDir - should create the directory and return no error
	invalidCfg = config.DefaultConfig()
	invalidCfg.JournalDir = filepath.Join(tmpDir, "nonexistent")
	filePath, err = CreateDailyJournalFile(invalidCfg, date)
	assert.NoError(t, err)
	assert.DirExists(t, invalidCfg.JournalDir)
	assert.FileExists(t, filePath)

	// Test case 6: Custom file naming convention
	cfg.DailyFileName = `{{.Date | formatDate "02"}}-{{.Date | formatDate "01"}}-{{.Date | formatDate "2006"}}.log`
	date = time.Date(2025, time.December, 25, 0, 0, 0, 0, time.UTC)
	expectedFilePath = filepath.Join(tmpDir, "25-12-2025.log")

	filePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)
	assert.Equal(t, expectedFilePath, filePath)
	assert.FileExists(t, expectedFilePath)

	// Test case 7: Daily template is applied
	cfg = config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyTemplate = "# {{.Date | formatDate \"2006-01-02\"}} - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n"
	date = time.Date(2025, time.October, 26, 0, 0, 0, 0, time.UTC)
	expectedFilePath = filepath.Join(tmpDir, "2025-10-26.md")

	filePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, "# 2025-10-26 - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n", string(content))
}

func TestAppendToLog(t *testing.T) {
	// Setup a temporary journal directory and file
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyTemplate = "# {{.Date | formatDate \"2006-01-02\"}} - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n"
	date := time.Date(2025, time.October, 26, 0, 0, 0, 0, time.UTC)

	filePath, err := CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	// Test case 1: Append a single log entry
	logEntry := "This is a new log entry."
	appendDate := time.Date(2025, time.October, 26, 14, 30, 0, 0, time.UTC)
	expectedLogContent := "# 2025-10-26 - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n\n14:30 This is a new log entry.\n"

	err = AppendToLog(filePath, logEntry, appendDate)
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogContent, string(content))

	// Test case 2: Append another log entry
	logEntry2 := "Another entry."
	appendDate2 := time.Date(2025, time.October, 26, 15, 0, 0, 0, time.UTC)
	expectedLogContent2 := expectedLogContent + "15:00 Another entry.\n"

	err = AppendToLog(filePath, logEntry2, appendDate2)
	assert.NoError(t, err)

	content, err = os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogContent2, string(content))

	// Test case 3: Append to a file without LOG chapter (should return error)
	noLogFilePath := filepath.Join(tmpDir, "no_log_file.md")
	err = os.WriteFile(noLogFilePath, []byte("Just some content\n"), 0644)
	assert.NoError(t, err)

	err = AppendToLog(noLogFilePath, "Should fail", appendDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LOG chapter not found in file")
}
