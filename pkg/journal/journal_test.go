package journal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/clobrano/LogBook/pkg/ai"
	"github.com/clobrano/LogBook/pkg/config"

	"github.com/stretchr/testify/assert"
)

// ErrorReader is a mock io.Reader that always returns an error.
type ErrorReader struct {
	Err error
}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, r.Err
}

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

	// Test case 8: Ensure the first paragraph is the summary
	cfg.DailyTemplate = "This is the summary.\n\n## LOG\n"
	date = time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	expectedFilePath = filepath.Join(tmpDir, "2025-11-01.md")

	filePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	content, err = os.ReadFile(filePath)
	assert.NoError(t, err)
	fmt.Println("DEBUG: Content for Test case 8:", string(content))
	assert.True(t, strings.HasPrefix(string(content), "This is the summary.\n\n"))
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

	// Test GenerateSummaryIfMissing
	// Setup a temporary journal directory and file for summary tests
	summaryTmpDir := t.TempDir()
	cfg.JournalDir = summaryTmpDir
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	date = time.Date(2025, time.November, 10, 0, 0, 0, 0, time.UTC)
	summaryFilePath, err := CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	// Test case 1: No summary exists, should generate one using AI
	mockAI := &ai.MockAISummarizer{Summary: "AI generated summary.", Err: nil}
	aiPrompt := "Summarize this."

	// Use a copy of cfg for this test case to avoid modifying the original
	aiCfg := config.DefaultConfig()
	aiCfg.AISummarizer = mockAI // Set the AI summarizer in the config

	err = GenerateSummaryIfMissing(summaryFilePath, aiCfg, mockAI, aiPrompt, strings.NewReader(""))
	assert.NoError(t, err)

	content, err = os.ReadFile(summaryFilePath)
	assert.NoError(t, err)
	expectedContent := "# Daily Log\nAI generated summary.\n\n## LOG\n"
	assert.Equal(t, expectedContent, string(content))

	// Test case 2: Summary already exists, should not overwrite (AI path)
	cfg.DailyTemplate = "# Daily Log\nExisting summary.\n\n## LOG\n"
	date = time.Date(2025, time.November, 11, 0, 0, 0, 0, time.UTC)
	summaryFilePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	err = GenerateSummaryIfMissing(summaryFilePath, aiCfg, mockAI, aiPrompt, strings.NewReader(""))
	assert.NoError(t, err)

	content, err = os.ReadFile(summaryFilePath)
	assert.NoError(t, err)
	expectedContent = "# Daily Log\nExisting summary.\n\n## LOG\n"
	assert.Equal(t, expectedContent, string(content))

	// Test case 3: AI summarizer returns an error
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	date = time.Date(2025, time.November, 12, 0, 0, 0, 0, time.UTC)
	summaryFilePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	mockAIWithError := &ai.MockAISummarizer{Summary: "", Err: errors.New("AI error during summary generation")}
	aiCfgWithError := config.DefaultConfig()
	aiCfgWithError.AISummarizer = mockAIWithError

	err = GenerateSummaryIfMissing(summaryFilePath, aiCfgWithError, mockAIWithError, aiPrompt, strings.NewReader(""))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate summary with AI: AI error during summary generation")

	// Test case 4: No AI agent configured, user provides manual summary
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	date = time.Date(2025, time.November, 13, 0, 0, 0, 0, time.UTC)
	summaryFilePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	manualSummaryInput := "This is a manual summary.\n"

	// No AI summarizer in config
	noAICfg := config.DefaultConfig()
	noAICfg.AISummarizer = nil

	err = GenerateSummaryIfMissing(summaryFilePath, noAICfg, nil, aiPrompt, strings.NewReader(manualSummaryInput))
	assert.NoError(t, err)

	content, err = os.ReadFile(summaryFilePath)
	assert.NoError(t, err)
	expectedContent = "# Daily Log\nThis is a manual summary.\n\n## LOG\n"
	assert.Equal(t, expectedContent, string(content))

	// Test case 5: No AI agent configured, user skips manual summary
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	date = time.Date(2025, time.November, 14, 0, 0, 0, 0, time.UTC)
	summaryFilePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	// Empty input to simulate skipping
	err = GenerateSummaryIfMissing(summaryFilePath, noAICfg, nil, aiPrompt, strings.NewReader("\n"))
	assert.NoError(t, err)

	content, err = os.ReadFile(summaryFilePath)
	assert.NoError(t, err)
	// Content should remain unchanged (no summary added)
	expectedContent = "# Daily Log\n\n## LOG\n"
	assert.Equal(t, expectedContent, string(content))

	// Test case 6: No AI agent configured, error reading manual summary
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	date = time.Date(2025, time.November, 15, 0, 0, 0, 0, time.UTC)
	summaryFilePath, err = CreateDailyJournalFile(cfg, date)
	assert.NoError(t, err)

	// Simulate an error during read
	err = GenerateSummaryIfMissing(summaryFilePath, noAICfg, nil, aiPrompt, &ErrorReader{Err: errors.New("read error")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manual summary: read error")
}

