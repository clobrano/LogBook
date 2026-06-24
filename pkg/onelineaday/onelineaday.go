package onelineaday

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/journal"
)

// Generate scans all daily journal files and produces one markdown file per year
// with entries grouped by month. Each entry contains the summary (first paragraph
// after the title) from the corresponding daily journal file.
func Generate(cfg *config.Config, year int) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("invalid configuration: %w", err)
	}

	journalDir := cfg.JournalDir
	if !filepath.IsAbs(journalDir) {
		return "", fmt.Errorf("JournalDir must be an absolute path: %s", journalDir)
	}

	entries, err := os.ReadDir(journalDir)
	if err != nil {
		return "", fmt.Errorf("failed to read journal directory: %w", err)
	}

	type dailyEntry struct {
		date    time.Time
		summary string
	}

	byMonth := make(map[time.Month][]dailyEntry)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		dateStr := strings.TrimSuffix(name, ".md")
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if parsed.Year() != year {
			continue
		}

		filePath := filepath.Join(journalDir, name)
		summary, err := journal.ExtractSummary(filePath)
		if err != nil {
			continue
		}
		if summary == "" {
			continue
		}

		byMonth[parsed.Month()] = append(byMonth[parsed.Month()], dailyEntry{
			date:    parsed,
			summary: summary,
		})
	}

	if len(byMonth) == 0 {
		return "", fmt.Errorf("no journal entries with summaries found for year %d", year)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Year %d\n", year))

	for month := time.January; month <= time.December; month++ {
		entries := byMonth[month]
		if len(entries) == 0 {
			continue
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].date.Before(entries[j].date)
		})

		b.WriteString(fmt.Sprintf("\n# %s\n\n", month.String()[:3]))

		for _, e := range entries {
			b.WriteString(fmt.Sprintf("* %s: %s\n", e.date.Format("2006-01-02"), e.summary))
		}
	}

	outputPath := filepath.Join(journalDir, fmt.Sprintf("one-line-a-day-%d.md", year))
	if err := os.WriteFile(outputPath, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write one-line-a-day file: %w", err)
	}

	return outputPath, nil
}

// GenerateAll scans the journal directory and generates one-line-a-day files
// for every year that has journal entries. Returns the list of generated file paths.
func GenerateAll(cfg *config.Config) ([]string, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	journalDir := cfg.JournalDir
	if !filepath.IsAbs(journalDir) {
		return nil, fmt.Errorf("JournalDir must be an absolute path: %s", journalDir)
	}

	entries, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read journal directory: %w", err)
	}

	years := make(map[int]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		dateStr := strings.TrimSuffix(name, ".md")
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		years[parsed.Year()] = true
	}

	sortedYears := make([]int, 0, len(years))
	for y := range years {
		sortedYears = append(sortedYears, y)
	}
	sort.Ints(sortedYears)

	var paths []string
	for _, y := range sortedYears {
		path, err := Generate(cfg, y)
		if err != nil {
			continue
		}
		paths = append(paths, path)
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no journal entries with summaries found")
	}

	return paths, nil
}
