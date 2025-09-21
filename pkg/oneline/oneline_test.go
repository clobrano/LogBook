package oneline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/template"
	"github.com/stretchr/testify/assert"
)

func TestGetPastSummaries(t *testing.T) {
	// Setup a temporary journal directory
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"
	cfg.DailyTemplate = "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n{{.Summary}}\n\n## LOG\n"

	// Helper to create dummy journal files
	createDummyJournalFile := func(date time.Time, summary string) string {
		data := template.TemplateData{Date: date, Summary: summary}
		fileName, _ := template.Render(cfg.DailyFileName, data)
		filePath := filepath.Join(tmpDir, fileName)
		content, _ := template.Render(cfg.DailyTemplate, data)
		os.WriteFile(filePath, []byte(content), 0644)
		return filePath
	}

	targetDate := time.Date(2025, time.September, 20, 0, 0, 0, 0, time.UTC)

	// Create dummy journal files for various past dates
	createDummyJournalFile(targetDate.AddDate(0, 0, -7), "Summary for 1 week ago.")       // 1 week ago
	createDummyJournalFile(targetDate.AddDate(0, -1, 0), "Summary for 1 month ago.")      // 1 month ago
	createDummyJournalFile(targetDate.AddDate(0, -6, 0), "Summary for 6 months ago.")     // 6 months ago
	createDummyJournalFile(targetDate.AddDate(-1, 0, 0), "Summary for 1 year ago.")     // 1 year ago
	// Do not create file for 2 years ago to test "missing" case

	// Test case 1: Retrieve summaries for past periods
	expectedSummaries := map[string]string{
		"1_week_ago":   "Summary for 1 week ago.",
		"1_month_ago":  "Summary for 1 month ago.",
		"6_months_ago": "Summary for 6 months ago.",
		"1_year_ago":   "Summary for 1 year ago.",
		"2_years_ago":  "missing",
	}

	actualSummaries, err := GetPastSummaries(cfg, targetDate)
	assert.NoError(t, err)
	assert.Equal(t, expectedSummaries, actualSummaries)
}