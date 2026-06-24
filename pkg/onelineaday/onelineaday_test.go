package onelineaday

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/stretchr/testify/assert"
)

func newTestConfig(dir string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.JournalDir = dir
	return cfg
}

func createJournalFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	assert.NoError(t, err)
}

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2025-01-10.md",
		"# Jan 10 2025 Friday\nStarted the new project today.\n\n# LOG\n10:00 Did stuff\n")
	createJournalFile(t, tmpDir, "2025-01-15.md",
		"# Jan 15 2025 Wednesday\nMade progress on the feature.\n\n# LOG\n11:00 More stuff\n")
	createJournalFile(t, tmpDir, "2025-03-05.md",
		"# Mar 05 2025 Wednesday\nFixed a critical bug in production.\n\n# One-line note\nold stuff\n\n# LOG\n09:00 Debugging\n")

	path, err := Generate(cfg, 2025)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "one-line-a-day-2025.md"), path)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	expected := `# Year 2025

# Jan

* 2025-01-10 Fri: Started the new project today.
* 2025-01-15 Wed: Made progress on the feature.

# Mar

* 2025-03-05 Wed: Fixed a critical bug in production.
`
	assert.Equal(t, expected, string(content))
}

func TestGenerateSkipsEmptySummaries(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2025-02-01.md",
		"# Feb 01 2025 Saturday\nHas a summary.\n\n# LOG\n")
	createJournalFile(t, tmpDir, "2025-02-02.md",
		"# Feb 02 2025 Sunday\n<!-- add today summary below this line -->\n\n# LOG\n")

	path, err := Generate(cfg, 2025)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "2025-02-01")
	assert.NotContains(t, string(content), "2025-02-02")
}

func TestGenerateSkipsNonJournalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2025-06-01.md",
		"# Jun 01 2025 Sunday\nA normal entry.\n\n# LOG\n")
	createJournalFile(t, tmpDir, "review_week_2025_23.md",
		"# Weekly Review\nSome review content.\n\n")
	createJournalFile(t, tmpDir, "notes.md",
		"# Random notes\nNot a journal.\n")

	path, err := Generate(cfg, 2025)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "2025-06-01")
	assert.NotContains(t, string(content), "review")
	assert.NotContains(t, string(content), "Random")
}

func TestGenerateNoEntriesReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	_, err := Generate(cfg, 2025)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no journal entries")
}

func TestGenerateMultipleYears(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2024-12-25.md",
		"# Dec 25 2024 Wednesday\nChristmas day notes.\n\n# LOG\n")
	createJournalFile(t, tmpDir, "2025-01-01.md",
		"# Jan 01 2025 Wednesday\nNew year notes.\n\n# LOG\n")

	// Generate for 2024 only
	path, err := Generate(cfg, 2024)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "Year 2024")
	assert.Contains(t, string(content), "2024-12-25")
	assert.NotContains(t, string(content), "2025")
}

func TestExtractFirstSectionMultipleParagraphs(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2025-04-01.md",
		"# Apr 01 2025 Tuesday\nFirst paragraph of the day.\n\nSecond paragraph with more details.\n\n# LOG\n10:00 Did stuff\n")

	path, err := Generate(cfg, 2025)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "First paragraph of the day. Second paragraph with more details.")
}

func TestExtractFirstSectionWithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2021-12-02.md", `---
date:  2021-w48
family:  green
---

[nav](link.md)

# Notes
> Winning is nice if you don't lose your integrity in the process.

First real paragraph here.

Second paragraph with more thoughts.


# Todos
- [x] test something
- [ ] do something else
`)

	path, err := Generate(cfg, 2021)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "Winning is nice")
	assert.Contains(t, s, "First real paragraph here.")
	assert.Contains(t, s, "Second paragraph with more thoughts.")
	assert.NotContains(t, s, "test something")
	assert.NotContains(t, s, "Todos")
}

func TestGenerateAll(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(tmpDir)

	createJournalFile(t, tmpDir, "2024-06-15.md",
		"# Jun 15 2024 Saturday\nMid-year entry.\n\n# LOG\n")
	createJournalFile(t, tmpDir, "2025-03-10.md",
		"# Mar 10 2025 Monday\nAnother entry.\n\n# LOG\n")

	paths, err := GenerateAll(cfg)
	assert.NoError(t, err)
	assert.Len(t, paths, 2)

	assert.Equal(t, filepath.Join(tmpDir, "one-line-a-day-2024.md"), paths[0])
	assert.Equal(t, filepath.Join(tmpDir, "one-line-a-day-2025.md"), paths[1])

	for _, p := range paths {
		assert.FileExists(t, p)
	}
}
