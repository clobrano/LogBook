package oneline

import (
	"os"
	"path/filepath"
	"strings"
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
	createDummyJournalFile(targetDate.AddDate(0, 0, -7), "Summary for 1 week ago.")   // 1 week ago
	createDummyJournalFile(targetDate.AddDate(0, -1, 0), "Summary for 1 month ago.")  // 1 month ago
	createDummyJournalFile(targetDate.AddDate(0, -6, 0), "Summary for 6 months ago.") // 6 months ago
	createDummyJournalFile(targetDate.AddDate(-1, 0, 0), "Summary for 1 year ago.")   // 1 year ago
	createDummyJournalFile(targetDate.AddDate(-2, 0, 0), "Summary for 2 years ago.")  // 2 years ago
	// Do not create file for 3 years ago to test breaking the loop

	// Test case 1: Retrieve summaries for past periods
	// Keys are now date strings in YYYY-MM-DD format
	expectedSummaries := map[string]string{
		"2025-09-13": "Summary for 1 week ago.",
		"2025-08-20": "Summary for 1 month ago.",
		"2025-03-20": "Summary for 6 months ago.",
		"2024-09-20": "Summary for 1 year ago.",
		"2023-09-20": "Summary for 2 years ago.",
	}

	actualSummaries, err := GetPastSummaries(cfg, targetDate)
	assert.NoError(t, err)
	assert.Equal(t, expectedSummaries, actualSummaries)
}

// TestExtractSummaryRobustEdgeCases verifies that the oneline summary
// extractor never returns YAML metadata as the summary, even when the
// frontmatter is malformed or the layout has unexpected whitespace.
func TestExtractSummaryRobustEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	type tc struct {
		name    string
		content string
		want    string
	}
	cases := []tc{
		{
			// User-reported bug: blank line between closing "---" and title
			name:    "blank line between frontmatter and title",
			content: "---\nsome metadata\n---\n\n# 2024-04-12\n<summary>\n",
			want:    "<summary>",
		},
		{
			name:    "leading blank line before frontmatter",
			content: "\n---\nkey: val\n---\n# 2024\nMy summary\n\n# LOG\n",
			want:    "My summary",
		},
		{
			name:    "unclosed frontmatter must not leak metadata",
			content: "---\nfoo: bar\nbaz: qux\n# 2024\nMy summary\n",
			want:    "My summary",
		},
		{
			name:    "obsidian multiline tags",
			content: "---\ntags:\n  - daily\n  - work\n---\n# 2024\nMy summary\n\n# LOG\n",
			want:    "My summary",
		},
		{
			name:    "frontmatter only, no title",
			content: "---\nfoo: bar\n---\n",
			want:    "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, strings.ReplaceAll(c.name, " ", "_")+".md")
			err := os.WriteFile(path, []byte(c.content), 0644)
			assert.NoError(t, err)
			got, err := extractSummary(path)
			assert.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestGetPastSummariesWithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.JournalDir = tmpDir
	cfg.DailyFileName = "{{.Date | formatDate \"2006-01-02\"}}.md"

	// Helper to create journal files with YAML frontmatter
	createFrontmatterFile := func(date time.Time, summary string) {
		data := template.TemplateData{Date: date}
		fileName, _ := template.Render(cfg.DailyFileName, data)
		filePath := filepath.Join(tmpDir, fileName)
		content := "---\ncreated: " + date.Format("2006-01-02") + "\n---\n# " + date.Format("Jan 02 2006 Monday") + "\n" + summary + "\n\n# LOG\n"
		os.WriteFile(filePath, []byte(content), 0644)
	}

	targetDate := time.Date(2025, time.September, 20, 0, 0, 0, 0, time.UTC)

	// Create files with frontmatter
	createFrontmatterFile(targetDate.AddDate(0, 0, -7), "Summary for 1 week ago.")
	createFrontmatterFile(targetDate.AddDate(0, -1, 0), "Summary for 1 month ago.")
	createFrontmatterFile(targetDate.AddDate(0, -6, 0), "Summary for 6 months ago.")
	createFrontmatterFile(targetDate.AddDate(-1, 0, 0), "Summary for 1 year ago.")
	createFrontmatterFile(targetDate.AddDate(-2, 0, 0), "Summary for 2 years ago.")

	expectedSummaries := map[string]string{
		"2025-09-13": "Summary for 1 week ago.",
		"2025-08-20": "Summary for 1 month ago.",
		"2025-03-20": "Summary for 6 months ago.",
		"2024-09-20": "Summary for 1 year ago.",
		"2023-09-20": "Summary for 2 years ago.",
		"2022-09-20": "missing", // 3 years ago, no file exists
	}

	actualSummaries, err := GetPastSummaries(cfg, targetDate)
	assert.NoError(t, err)
	assert.Equal(t, expectedSummaries, actualSummaries)
}