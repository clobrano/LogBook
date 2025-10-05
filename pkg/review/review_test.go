package review

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

func TestReviewWeek(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"
	cfg.DailyTemplate = "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n{{.Summary}}\n\n## LOG\n"

	// Create dummy journal files for a specific week (e.g., week 38, 2025)
	createDummyJournalFile := func(date time.Time, summary string) string {
		data := template.TemplateData{Date: date, Summary: summary}
		fileName, _ := template.Render(cfg.DailyFileName, data)
		filePath := filepath.Join(tmpDir, fileName)
		content, _ := template.Render(cfg.DailyTemplate, data)
		os.WriteFile(filePath, []byte(content), 0644)
		return filePath
	}

	// Week 38, 2025: Monday, Sep 15 to Sunday, Sep 21
	createDummyJournalFile(time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC), "Summary for Sep 15.")
	createDummyJournalFile(time.Date(2025, time.September, 16, 0, 0, 0, 0, time.UTC), "Summary for Sep 16.")
	createDummyJournalFile(time.Date(2025, time.September, 17, 0, 0, 0, 0, time.UTC), "Summary for Sep 17.")
	// Missing 18th
	createDummyJournalFile(time.Date(2025, time.September, 19, 0, 0, 0, 0, time.UTC), "Summary for Sep 19.")
	createDummyJournalFile(time.Date(2025, time.September, 20, 0, 0, 0, 0, time.UTC), "Summary for Sep 20.\n") // Added newline for testing
	createDummyJournalFile(time.Date(2025, time.September, 21, 0, 0, 0, 0, time.UTC), "Summary for Sep 21.")

	week := 38
	year := 2025

	// Test case 1: AI-generated summary for review
	aiSummarizer := &ai.MockAISummarizer{Summary: "AI generated weekly summary.", Err: nil}
	aiCfg := config.DefaultConfig()
	aiCfg.JournalDir = tmpDir
	aiCfg.DailyFileName = cfg.DailyFileName
	aiCfg.DailyTemplate = cfg.DailyTemplate
	aiCfg.AISummarizer = aiSummarizer

	result, err := ReviewWeek(aiCfg, week, year, aiSummarizer, strings.NewReader(""))
	assert.NoError(t, err)
	expectedSuccessMessage := fmt.Sprintf("Weekly review generated at: %s", filepath.Join(tmpDir, "review_week_2025_38.md"))
	assert.Equal(t, expectedSuccessMessage, result)

	reviewFilePath := filepath.Join(tmpDir, fmt.Sprintf("review_week_%d_%d.md", year, week))
	assert.FileExists(t, reviewFilePath)

	reviewContent, err := os.ReadFile(reviewFilePath)
	assert.NoError(t, err)

	expectedReviewContent := strings.Join([]string{
		"# Weekly Review - Week 38, 2025",
		"AI generated weekly summary.\n",
		"## Daily Summaries\n",
		"### 2025-09-15\nSummary for Sep 15.\n",
		"### 2025-09-16\nSummary for Sep 16.\n",
		"### 2025-09-17\nSummary for Sep 17.\n",
		"### 2025-09-19\nSummary for Sep 19.\n",
		"### 2025-09-20\nSummary for Sep 20.\n",
		"### 2025-09-21\nSummary for Sep 21.\n",
		"",
	}, "\n")
	assert.Equal(t, expectedReviewContent, string(reviewContent))

	// Test case 2: Manual summary for review
	manualSummaryInput := "This is a manual weekly summary.\n"
	manualReader := strings.NewReader(manualSummaryInput)
	manualCfg := config.DefaultConfig()
	manualCfg.JournalDir = tmpDir
	manualCfg.DailyFileName = cfg.DailyFileName
	manualCfg.DailyTemplate = cfg.DailyTemplate
	manualCfg.AISummarizer = nil // No AI summarizer

	// Re-create the review file to ensure it's clean for manual input
	os.Remove(reviewFilePath)
	result, err = ReviewWeek(manualCfg, week, year, nil, manualReader)
	assert.NoError(t, err)
	expectedSuccessMessage = fmt.Sprintf("Weekly review generated at: %s", filepath.Join(tmpDir, "review_week_2025_38.md"))
	assert.Equal(t, expectedSuccessMessage, result)

	reviewContent, err = os.ReadFile(reviewFilePath)
	assert.NoError(t, err)

	expectedManualReviewContent := strings.Join([]string{
		"# Weekly Review - Week 38, 2025",
		"This is a manual weekly summary.\n", // This line is changed
		"## Daily Summaries\n",
		"### 2025-09-15\nSummary for Sep 15.\n",
		"### 2025-09-16\nSummary for Sep 16.\n",
		"### 2025-09-17\nSummary for Sep 17.\n",
		"### 2025-09-19\nSummary for Sep 19.\n",
		"### 2025-09-20\nSummary for Sep 20.\n",
		"### 2025-09-21\nSummary for Sep 21.\n",
		"",
	}, "\n")
	assert.Equal(t, expectedManualReviewContent, string(reviewContent))

	// Test case 3: No journal entries for the week (manual summary skipped)
	noEntriesTmpDir := t.TempDir()
	noEntriesCfg := config.DefaultConfig()
	noEntriesCfg.JournalDir = noEntriesTmpDir
	noEntriesCfg.DailyFileName = cfg.DailyFileName
	noEntriesCfg.DailyTemplate = cfg.DailyTemplate
	noEntriesCfg.AISummarizer = nil

	result, err = ReviewWeek(noEntriesCfg, week, year, nil, strings.NewReader("\n")) // Simulate skipping manual summary
	assert.NoError(t, err)
	assert.Contains(t, result, fmt.Sprintf("Weekly review generated at: %s", filepath.Join(noEntriesTmpDir, "review_week_2025_38.md")))

	reviewFilePath = filepath.Join(noEntriesTmpDir, fmt.Sprintf("review_week_%d_%d.md", year, week))
	assert.FileExists(t, reviewFilePath)

	reviewContent, err = os.ReadFile(reviewFilePath)
	assert.NoError(t, err)
	expectedNoSummaryReviewContent := strings.Join([]string{
		"# Weekly Review - Week 38, 2025\n",
		"No journal entries found for this week.\n\n",
	}, "\n")
	assert.Equal(t, expectedNoSummaryReviewContent, string(reviewContent))

	// Test case 4: Error during manual summary input
	errorReader := &ErrorReader{Err: errors.New("read error during manual summary")}
	os.Remove(reviewFilePath) // Clean up previous review file
	_, err = ReviewWeek(noEntriesCfg, week, year, nil, errorReader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate summary for weekly review: failed to read manual summary: read error during manual summary")
}

func TestReviewMonth(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}" + ".md"
	cfg.DailyTemplate = "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n{{.Summary}}\n\n## LOG\n"

	// Create dummy journal files for a specific month (e.g., September 2025)
	createDummyJournalFile := func(date time.Time, summary string) string {
		data := template.TemplateData{Date: date, Summary: summary}
		fileName, _ := template.Render(cfg.DailyFileName, data)
		filePath := filepath.Join(tmpDir, fileName)
		content, _ := template.Render(cfg.DailyTemplate, data)
		os.WriteFile(filePath, []byte(content), 0644)
		return filePath
	}

	// September 2025
	createDummyJournalFile(time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC), "Summary for Sep 01.")
	createDummyJournalFile(time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC), "Summary for Sep 15.")
	createDummyJournalFile(time.Date(2025, time.September, 30, 0, 0, 0, 0, time.UTC), "Summary for Sep 30.")

	month := "September"
	year := 2025

	// Test case 1: AI-generated summary for review
	aiSummarizer := &ai.MockAISummarizer{Summary: "AI generated monthly summary.", Err: nil}
	aiCfg := config.DefaultConfig()
	aiCfg.JournalDir = tmpDir
	aiCfg.DailyFileName = cfg.DailyFileName
	aiCfg.DailyTemplate = cfg.DailyTemplate
	aiCfg.AISummarizer = aiSummarizer

	result, err := ReviewMonth(aiCfg, month, year, aiSummarizer, strings.NewReader(""))
	assert.NoError(t, err)
	expectedSuccessMessage := fmt.Sprintf("Monthly review generated at: %s", filepath.Join(tmpDir, "review_month_September_2025.md"))
	assert.Equal(t, expectedSuccessMessage, result)

	reviewFilePath := filepath.Join(tmpDir, fmt.Sprintf("review_month_%s_%d.md", month, year))
	assert.FileExists(t, reviewFilePath)

	reviewContent, err := os.ReadFile(reviewFilePath)
	assert.NoError(t, err)

	expectedReviewContent := strings.Join([]string{
		"# Monthly Review - September 2025",
		"AI generated monthly summary.\n",
		"## Daily Summaries\n",
		"### 2025-09-01\nSummary for Sep 01.\n",
		"### 2025-09-15\nSummary for Sep 15.\n",
		"### 2025-09-30\nSummary for Sep 30.\n",
		"",
	}, "\n")
	assert.Equal(t, expectedReviewContent, string(reviewContent))

	// Test case 2: Manual summary for review
	manualSummaryInput := "This is a manual monthly summary.\n"
	manualReader := strings.NewReader(manualSummaryInput)
	manualCfg := config.DefaultConfig()
	manualCfg.JournalDir = tmpDir
	manualCfg.DailyFileName = cfg.DailyFileName
	manualCfg.DailyTemplate = cfg.DailyTemplate
	manualCfg.AISummarizer = nil // No AI summarizer

	// Re-create the review file to ensure it's clean for manual input
	os.Remove(reviewFilePath)
	result, err = ReviewMonth(manualCfg, month, year, nil, manualReader)
	assert.NoError(t, err)
	expectedSuccessMessage = fmt.Sprintf("Monthly review generated at: %s", filepath.Join(tmpDir, "review_month_September_2025.md"))
	assert.Equal(t, expectedSuccessMessage, result)

	reviewContent, err = os.ReadFile(reviewFilePath)
	assert.NoError(t, err)

	expectedManualReviewContent := strings.Join([]string{
		"# Monthly Review - September 2025",
		"This is a manual monthly summary.\n",
		"## Daily Summaries\n",
		"### 2025-09-01\nSummary for Sep 01.\n",
		"### 2025-09-15\nSummary for Sep 15.\n",
		"### 2025-09-30\nSummary for Sep 30.\n",
		"",
	}, "\n")
	assert.Equal(t, expectedManualReviewContent, string(reviewContent))

	// Test case 3: No journal entries for the month (manual summary skipped)
	noEntriesTmpDir := t.TempDir()
	noEntriesCfg := config.DefaultConfig()
	noEntriesCfg.JournalDir = noEntriesTmpDir
	noEntriesCfg.DailyFileName = cfg.DailyFileName
	noEntriesCfg.DailyTemplate = cfg.DailyTemplate
	noEntriesCfg.AISummarizer = nil

	os.Remove(reviewFilePath) // Clean up previous review file
	result, err = ReviewMonth(noEntriesCfg, month, year, nil, strings.NewReader("\n")) // Simulate skipping manual summary
	assert.NoError(t, err)
	assert.Contains(t, result, fmt.Sprintf("Monthly review generated at: %s", filepath.Join(noEntriesTmpDir, "review_month_September_2025.md")))

	reviewFilePath = filepath.Join(noEntriesTmpDir, fmt.Sprintf("review_month_%s_%d.md", month, year))
	assert.FileExists(t, reviewFilePath)

	reviewContent, err = os.ReadFile(reviewFilePath)
	assert.NoError(t, err)
	expectedNoSummaryReviewContent := strings.Join([]string{
		"# Monthly Review - September 2025\n",
		"No journal entries found for this month.\n\n",
	}, "\n")
	assert.Equal(t, expectedNoSummaryReviewContent, string(reviewContent))

	// Test case 4: Error during manual summary input
	errorReader := &ErrorReader{Err: errors.New("read error during manual summary")}
	os.Remove(reviewFilePath) // Clean up previous review file
	_, err = ReviewMonth(noEntriesCfg, month, year, nil, errorReader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate summary for monthly review: failed to read manual summary: read error during manual summary")
}


func TestReviewYear(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}" + ".md"
	cfg.DailyTemplate = "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n{{.Summary}}\n\n## LOG\n"

	// Create dummy journal files for a specific year (e.g., 2025)
	createDummyJournalFile := func(date time.Time, summary string) string {
		data := template.TemplateData{Date: date, Summary: summary}
		fileName, _ := template.Render(cfg.DailyFileName, data)
		filePath := filepath.Join(tmpDir, fileName)
		content, _ := template.Render(cfg.DailyTemplate, data)
		os.WriteFile(filePath, []byte(content), 0644)
		return filePath
	}

	// 2025
	createDummyJournalFile(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), "Summary for Jan 01.")
	createDummyJournalFile(time.Date(2025, time.June, 15, 0, 0, 0, 0, time.UTC), "Summary for Jun 15.")
	createDummyJournalFile(time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC), "Summary for Dec 31.")

	year := 2025

	// Test case 1: AI-generated summary for review
	aiSummarizer := &ai.MockAISummarizer{Summary: "AI generated yearly summary.", Err: nil}
	aiCfg := config.DefaultConfig()
	aiCfg.JournalDir = tmpDir
	aiCfg.DailyFileName = cfg.DailyFileName
	aiCfg.DailyTemplate = cfg.DailyTemplate
	aiCfg.AISummarizer = aiSummarizer

	result, err := ReviewYear(aiCfg, year, aiSummarizer, strings.NewReader(""))
	assert.NoError(t, err)
	expectedSuccessMessage := fmt.Sprintf("Yearly review generated at: %s", filepath.Join(tmpDir, "review_year_2025.md"))
	assert.Equal(t, expectedSuccessMessage, result)

	reviewFilePath := filepath.Join(tmpDir, fmt.Sprintf("review_year_%d.md", year))
	assert.FileExists(t, reviewFilePath)

	reviewContent, err := os.ReadFile(reviewFilePath)
	assert.NoError(t, err)

	expectedReviewContent := strings.Join([]string{
		"# Yearly Review - 2025",
		"AI generated yearly summary.\n",
		"## Monthly Summaries\n",
		"### January\n",
		"- **2025-01-01**: Summary for Jan 01.\n",
		"### June\n",
		"- **2025-06-15**: Summary for Jun 15.\n",
		"### December\n",
		"- **2025-12-31**: Summary for Dec 31.\n",
		"",
	}, "\n")
	assert.Equal(t, expectedReviewContent, string(reviewContent))

	// Test case 2: Manual summary for review
	manualSummaryInput := "This is a manual yearly summary.\n"
	manualReader := strings.NewReader(manualSummaryInput)
	manualCfg := config.DefaultConfig()
	manualCfg.JournalDir = tmpDir
	manualCfg.DailyFileName = cfg.DailyFileName
	manualCfg.DailyTemplate = cfg.DailyTemplate
	manualCfg.AISummarizer = nil // No AI summarizer

	// Re-create the review file to ensure it's clean for manual input
	os.Remove(reviewFilePath)
	result, err = ReviewYear(manualCfg, year, nil, manualReader)
	assert.NoError(t, err)
	expectedSuccessMessage = fmt.Sprintf("Yearly review generated at: %s", filepath.Join(tmpDir, "review_year_2025.md"))
	assert.Equal(t, expectedSuccessMessage, result)

	reviewContent, err = os.ReadFile(reviewFilePath)
	assert.NoError(t, err)

	expectedManualReviewContent := strings.Join([]string{
		"# Yearly Review - 2025",
		"This is a manual yearly summary.\n",
		"## Monthly Summaries\n",
		"### January\n",
		"- **2025-01-01**: Summary for Jan 01.\n",
		"### June\n",
		"- **2025-06-15**: Summary for Jun 15.\n",
		"### December\n",
		"- **2025-12-31**: Summary for Dec 31.\n",
		"",
	}, "\n")
	assert.Equal(t, expectedManualReviewContent, string(reviewContent))

	// Test case 3: No journal entries for the year (manual summary skipped)
	noEntriesTmpDir := t.TempDir()
	noEntriesCfg := config.DefaultConfig()
	noEntriesCfg.JournalDir = noEntriesTmpDir
	noEntriesCfg.DailyFileName = cfg.DailyFileName
	noEntriesCfg.DailyTemplate = cfg.DailyTemplate
	noEntriesCfg.AISummarizer = nil

	os.Remove(reviewFilePath) // Clean up previous review file
	result, err = ReviewYear(noEntriesCfg, year, nil, strings.NewReader("\n")) // Simulate skipping manual summary
	assert.NoError(t, err)
	assert.Contains(t, result, fmt.Sprintf("Yearly review generated at: %s", filepath.Join(noEntriesTmpDir, "review_year_2025.md")))

	reviewFilePath = filepath.Join(noEntriesTmpDir, fmt.Sprintf("review_year_%d.md", year))
	assert.FileExists(t, reviewFilePath)

	reviewContent, err = os.ReadFile(reviewFilePath)
	assert.NoError(t, err)
	expectedNoSummaryReviewContent := strings.Join([]string{
		"# Yearly Review - 2025\n",
		"No journal entries found for this year.\n\n",
	}, "\n")
	assert.Equal(t, expectedNoSummaryReviewContent, string(reviewContent))

	// Test case 4: Error during manual summary input
	errorReader := &ErrorReader{Err: errors.New("read error during manual summary")}
	os.Remove(reviewFilePath) // Clean up previous review file
	_, err = ReviewYear(noEntriesCfg, year, nil, errorReader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate summary for yearly review: failed to read manual summary: read error during manual summary")
}

