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
	"github.com/clobrano/LogBook/pkg/template"

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

	filePath, _, err := CreateDailyJournalFile(cfg, date, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedFilePath, filePath)
	assert.FileExists(t, expectedFilePath)

	filePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedFilePath, filePath)
	assert.FileExists(t, expectedFilePath)

	// Test case 3: Invalid configuration (empty JournalDir)
	invalidCfg := config.DefaultConfig()
	filePath, _, err = CreateDailyJournalFile(invalidCfg, date, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration: JournalDir cannot be empty")

	// Test case 4: Non-absolute JournalDir
	invalidCfg = config.DefaultConfig()
	filePath, _, err = CreateDailyJournalFile(invalidCfg, date, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JournalDir must be an absolute path")

	// Test case 5: Non-existent JournalDir - should create the directory and return no error
	invalidCfg = config.DefaultConfig()
	invalidCfg.JournalDir = filepath.Join(tmpDir, "nonexistent")
	filePath, _, err = CreateDailyJournalFile(invalidCfg, date, nil, nil)
	assert.NoError(t, err)
	assert.DirExists(t, invalidCfg.JournalDir)
	assert.FileExists(t, filePath)

	// Test case 6: Custom file naming convention
	cfg.DailyFileName = `{{.Date | formatDate "02"}}-{{.Date | formatDate "01"}}-{{.Date | formatDate "2006"}}.log`
	date = time.Date(2025, time.December, 25, 0, 0, 0, 0, time.UTC)
		filePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
	    	assert.NoError(t, err)
	assert.Equal(t, expectedFilePath, filePath)
	assert.FileExists(t, expectedFilePath)

	// Test case 7: Daily template is applied
	cfg = config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyTemplate = "# {{.Date | formatDate \"2006-01-02\"}} - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n"
	    date = time.Date(2025, time.October, 26, 0, 0, 0, 0, time.UTC)
	    		filePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
	    	    	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, "# 2025-10-26 - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n", string(content))

	// Test case 8: Ensure the first paragraph is the summary
	cfg.DailyTemplate = "This is the summary.\n\n## LOG\n"
	date = time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	expectedFilePath = filepath.Join(tmpDir, "2025-11-01.md")

	filePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
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

	filePath, _, err := CreateDailyJournalFile(cfg, date, nil, nil)
	assert.NoError(t, err)

	// Test case 1: Append a single log entry
	logEntry := "This is a new log entry."
	appendDate := time.Date(2025, time.October, 26, 14, 30, 0, 0, time.UTC)
	expectedLogContent := "# 2025-10-26 - My Daily Log\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n\n14:30 This is a new log entry.\n"

	err = AppendToLog(cfg, filePath, logEntry, appendDate)
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogContent, string(content))

	// Test case 2: Append another log entry
	logEntry2 := "Another entry."
	appendDate2 := time.Date(2025, time.October, 26, 15, 0, 0, 0, time.UTC)
	expectedLogContent2 := expectedLogContent + "15:00 Another entry.\n"

	err = AppendToLog(cfg, filePath, logEntry2, appendDate2)
	assert.NoError(t, err)

	content, err = os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogContent2, string(content))

	// Test case 3: Append to a file without LOG chapter (should return error)
	noLogFilePath := filepath.Join(tmpDir, "no_log_file.md")
	err = os.WriteFile(noLogFilePath, []byte("Just some content\n"), 0644)
	assert.NoError(t, err)

	err = AppendToLog(cfg, noLogFilePath, "Should fail", appendDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LOG chapter not found in file")

	// Test GenerateSummaryIfMissing
	// Setup a temporary journal directory and file for summary tests
	summaryTmpDir := t.TempDir()
	cfg.JournalDir = summaryTmpDir
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	date = time.Date(2025, time.November, 10, 0, 0, 0, 0, time.UTC)
	summaryFilePath, _, err := CreateDailyJournalFile(cfg, date, nil, nil)
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
	summaryFilePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
	assert.NoError(t, err)

	err = GenerateSummaryIfMissing(summaryFilePath, aiCfg, mockAI, aiPrompt, strings.NewReader(""))
	assert.NoError(t, err)

	content, err = os.ReadFile(summaryFilePath)
	assert.NoError(t, err)
	expectedContent = "# Daily Log\nExisting summary.\n\n## LOG\n"
	assert.Equal(t, expectedContent, string(content))

	// Test case 3: AI summarizer returns an error
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n"
	summaryFilePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
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
	summaryFilePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
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
	    	summaryFilePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
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
	summaryFilePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
	assert.NoError(t, err)

	// Simulate an error during read
	err = GenerateSummaryIfMissing(summaryFilePath, noAICfg, nil, aiPrompt, &ErrorReader{Err: errors.New("read error")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manual summary: read error")

	// Test case 7: AI summary generation ignores "One-line note" section
	cfg.DailyTemplate = "# Daily Log\n\n## LOG\n\n## One-line note\n- Past note: This is a past one-line note.\n"
	date = time.Date(2025, time.November, 16, 0, 0, 0, 0, time.UTC)
	summaryFilePath, _, err = CreateDailyJournalFile(cfg, date, nil, nil)
	assert.NoError(t, err)

	mockAI = &ai.MockAISummarizer{Summary: "AI generated summary without one-line note.", Err: nil}
	aiCfg = config.DefaultConfig()
	aiCfg.AISummarizer = mockAI

	err = GenerateSummaryIfMissing(summaryFilePath, aiCfg, mockAI, aiPrompt, strings.NewReader(""))
	assert.NoError(t, err)

	var contentBytesForOneLineNoteTest []byte
	contentBytesForOneLineNoteTest, err = os.ReadFile(summaryFilePath)
	assert.NoError(t, err)
	var contentForOneLineNoteTest string
	contentForOneLineNoteTest = string(contentBytesForOneLineNoteTest)
	expectedContentForOneLineNoteTest := "# Daily Log\nAI generated summary without one-line note.\n\n## LOG\n\n## One-line note\n- Past note: This is a past one-line note.\n"
	assert.Equal(t, expectedContentForOneLineNoteTest, contentForOneLineNoteTest)
}

func TestListJournalFilesByPeriod(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"

	// Create some dummy journal files
	createDummyFile := func(date time.Time) string {
		data := template.TemplateData{Date: date}
		fileName, _ := template.Render(cfg.DailyFileName, data)
		filePath := filepath.Join(tmpDir, fileName)
		os.WriteFile(filePath, []byte("dummy content"), 0644)
		return filePath
	}

	file2025_01_01 := createDummyFile(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC))
	file2025_01_02 := createDummyFile(time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC))
	file2025_01_03 := createDummyFile(time.Date(2025, time.January, 3, 0, 0, 0, 0, time.UTC))
	file2025_01_05 := createDummyFile(time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC))
	file2025_02_01 := createDummyFile(time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC))

	// Test case 1: Full range
	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC)
	expectedFiles := []string{file2025_01_01, file2025_01_02, file2025_01_03, file2025_01_05}

	files, err := ListJournalFilesByPeriod(cfg, startDate, endDate)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedFiles, files)

	// Test case 2: Partial range
	startDate = time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2025, time.January, 3, 0, 0, 0, 0, time.UTC)
	expectedFiles = []string{file2025_01_02, file2025_01_03}

	files, err = ListJournalFilesByPeriod(cfg, startDate, endDate)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedFiles, files)

	// Test case 3: Single day
	startDate = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	expectedFiles = []string{file2025_01_01}

	files, err = ListJournalFilesByPeriod(cfg, startDate, endDate)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedFiles, files)

	// Test case 4: No files in range
	startDate = time.Date(2025, time.January, 4, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2025, time.January, 4, 0, 0, 0, 0, time.UTC)
	expectedFiles = []string{}

	files, err = ListJournalFilesByPeriod(cfg, startDate, endDate)
	assert.NoError(t, err)
	assert.Empty(t, files)

	// Test case 5: Range extends beyond existing files
	startDate = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	expectedFiles = []string{file2025_01_01, file2025_01_02, file2025_01_03, file2025_01_05, file2025_02_01}

	files, err = ListJournalFilesByPeriod(cfg, startDate, endDate)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedFiles, files)

	// Test case 6: Invalid configuration (empty JournalDir)
	invalidCfg := config.DefaultConfig()
	invalidCfg.JournalDir = ""
	files, err = ListJournalFilesByPeriod(invalidCfg, startDate, endDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration: JournalDir cannot be empty")

	// Test case 7: Non-absolute JournalDir
	invalidCfg = config.DefaultConfig()
	invalidCfg.JournalDir = "./relative/path"
	files, err = ListJournalFilesByPeriod(invalidCfg, startDate, endDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JournalDir must be an absolute path")

	// Test case 8: Some files exist, some don't
	partialExistTmpDir := t.TempDir()
	partialExistCfg := config.DefaultConfig()
	partialExistCfg.JournalDir = partialExistTmpDir
	partialExistCfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"

	// Helper to create dummy journal files for partialExistTmpDir
	createPartialExistDummyFile := func(date time.Time) string {
		data := template.TemplateData{Date: date}
		fileName, _ := template.Render(partialExistCfg.DailyFileName, data)
		filePath := filepath.Join(partialExistTmpDir, fileName)
		os.WriteFile(filePath, []byte("dummy content"), 0644)
		return filePath
	}

	file2025_03_01 := createPartialExistDummyFile(time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC))
	// file2025_03_02 is intentionally not created
	file2025_03_03 := createPartialExistDummyFile(time.Date(2025, time.March, 3, 0, 0, 0, 0, time.UTC))

	startDate = time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2025, time.March, 3, 0, 0, 0, 0, time.UTC)
	expectedFiles = []string{file2025_03_01, file2025_03_03}

	files, err = ListJournalFilesByPeriod(partialExistCfg, startDate, endDate)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedFiles, files)
}

func TestExtractSummary(t *testing.T) {
	// Setup a temporary directory
	tmpDir := t.TempDir()

	// Test case 1: File with a summary
	filePath1 := filepath.Join(tmpDir, "file1.md")
	content1 := "# Title\nSummary of the file.\n\n## LOG\nEntry 1"
	err := os.WriteFile(filePath1, []byte(content1), 0644)
	assert.NoError(t, err)

	summary, err := ExtractSummary(filePath1)
	assert.NoError(t, err)
	assert.Equal(t, "Summary of the file.", summary)

	// Test case 2: File with multiple empty lines after title before summary
	filePath2 := filepath.Join(tmpDir, "file2.md")
	content2 := "# Title\n\n\nSummary of the file 2.\n\n## LOG\nEntry 1"
	err = os.WriteFile(filePath2, []byte(content2), 0644)
	assert.NoError(t, err)

	summary, err = ExtractSummary(filePath2)
	assert.NoError(t, err)
	assert.Equal(t, "Summary of the file 2.", summary)

	// Test case 3: File without a summary
	filePath3 := filepath.Join(tmpDir, "file3.md")
	content3 := "# Title\n\n## LOG\nEntry 1"
	err = os.WriteFile(filePath3, []byte(content3), 0644)
	assert.NoError(t, err)

	summary, err = ExtractSummary(filePath3)
	assert.NoError(t, err)
	assert.Empty(t, summary)

	// Test case 4: Empty file
	filePath4 := filepath.Join(tmpDir, "file4.md")
	content4 := ""
	err = os.WriteFile(filePath4, []byte(content4), 0644)
	assert.NoError(t, err)

	summary, err = ExtractSummary(filePath4)
	assert.NoError(t, err)
	assert.Empty(t, summary)

	// Test case 5: File does not exist
	filePath5 := filepath.Join(tmpDir, "nonexistent.md")
	summary, err = ExtractSummary(filePath5)
	assert.NoError(t, err) // Should not return error for non-existent file
	assert.Empty(t, summary)

	// Test case 6: Summary is a title (should be skipped)
	filePath6 := filepath.Join(tmpDir, "file6.md")
	content6 := "# Title\n## Another Title\nSummary after title.\n\n## LOG\nEntry 1"
	err = os.WriteFile(filePath6, []byte(content6), 0644)
	assert.NoError(t, err)

	summary, err = ExtractSummary(filePath6)
	assert.NoError(t, err)
	assert.Equal(t, "Summary after title.", summary)
}

func TestEmbedOneLineNotes(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"
	cfg.DailyTemplate = "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n{{.Summary}}\n\n## LOG\n"

	// Create a dummy daily journal file
	date := time.Date(2025, time.September, 20, 0, 0, 0, 0, time.UTC)
	data := template.TemplateData{Date: date, Summary: "Initial summary."}
	fileName, _ := template.Render(cfg.DailyFileName, data)
	filePath := filepath.Join(tmpDir, fileName)
	content, _ := template.Render(cfg.DailyTemplate, data)
	initialContent := content + "\n## One-line note\n\n"
	os.WriteFile(filePath, []byte(initialContent), 0644)

	// Sample summaries to embed
	summaries := map[string]string{
		"1_week_ago":   "Summary from 1 week ago.",
		"1_month_ago":  "Summary from 1 month ago.",
		"6_months_ago": "Summary from 6 months ago.",
		"1_year_ago":   "Summary from 1 year ago.",
		"2_years_ago":  "Summary from 2 years ago.",
	}

	err := EmbedOneLineNotes(filePath, summaries)
	assert.NoError(t, err)

	// Read the updated file content
	updatedContentBytes, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	updatedContent := string(updatedContentBytes)

	// Assert that each summary line is present in the updated content
	assert.Contains(t, updatedContent, "- 1 week ago: Summary from 1 week ago.\n")
	assert.Contains(t, updatedContent, "- 1 month ago: Summary from 1 month ago.\n")
	assert.Contains(t, updatedContent, "- 6 months ago: Summary from 6 months ago.\n")
	assert.Contains(t, updatedContent, "- 1 year ago: Summary from 1 year ago.\n")
	assert.Contains(t, updatedContent, "- 2 years ago: Summary from 2 years ago.\n")

	// Also assert the overall structure around the one-line notes section
	assert.Contains(t, updatedContent, "## LOG\n\n## One-line note\n")
	assert.Contains(t, updatedContent, "# Sep 20 2025 Saturday\n\nInitial summary.\n\n")
}

