package review

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/ai"
	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/journal"

	"github.com/fatih/color"
)

// ReviewWeek generates a weekly review file.
func ReviewWeek(cfg *config.Config, week int, year int, summarizer ai.AISummarizer, reader io.Reader) (string, error) {
	// Calculate start and end dates for the week using ISO week definition.
	// Go's time.ISOWeek() returns the ISO year and ISO week number.
	// To get the start date of a given ISO week, we can find the Thursday of that week.
	// The Thursday of the first week of the year is always in the first week.

	// Start by finding a date in the middle of the target week to ensure we get the correct ISO week.
	// We can pick the 4th day of the year, as ISO week 1 always contains Jan 4.

	dateInTargetWeek := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)

	// Adjust to the correct year's ISO week 1
	isoYear, isoWeek := dateInTargetWeek.ISOWeek()
	for isoYear < year || (isoYear == year && isoWeek < week) {
		dateInTargetWeek = dateInTargetWeek.AddDate(0, 0, 7)
		isoYear, isoWeek = dateInTargetWeek.ISOWeek()
	}
	for isoYear > year || (isoYear == year && isoWeek > week) {
		dateInTargetWeek = dateInTargetWeek.AddDate(0, 0, -7)
		isoYear, isoWeek = dateInTargetWeek.ISOWeek()
	}

	// Now dateInTargetWeek is a date within the target ISO week.
	// Find the Monday of this week.
	startDate := dateInTargetWeek
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, -1)
	}
	endDate := startDate.AddDate(0, 0, 6)

	// List journal files for the period
	journalFiles, err := journal.ListJournalFilesByPeriod(cfg, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to list journal files for weekly review: %w", err)
	}

	var reviewContentBuilder strings.Builder
	reviewContentBuilder.WriteString(fmt.Sprintf("# Weekly Review - Week %d, %d\n\n", week, year))

	// Write to a temporary review file for now
	reviewFilePath := filepath.Join(cfg.JournalDir, fmt.Sprintf("review_week_%d_%d.md", year, week))
	if err := os.MkdirAll(filepath.Dir(reviewFilePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for weekly review file: %w", err)
	}
	err = os.WriteFile(reviewFilePath, []byte(reviewContentBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write weekly review file: %w", err)
	}

	// Generate summary for the review file if missing
	reviewSummaryPrompt := "Write a summary of the weekly review using the same Language. Use 1st person and a simple language. Use 200 characters or less."
	err = journal.GenerateSummaryIfMissing(reviewFilePath, cfg, summarizer, reviewSummaryPrompt, reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary for weekly review: %w", err)
	}

	// Read the content again after summary generation
	reviewContentBytes, err := os.ReadFile(reviewFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read weekly review file after summary generation: %w", err)
	}
	reviewContentBuilder.Reset()
	reviewContentBuilder.Write(reviewContentBytes)

	if len(journalFiles) == 0 {
		reviewContentBuilder.WriteString("No journal entries found for this week.\n\n")
	} else {
		reviewContentBuilder.WriteString("## Daily Summaries\n\n")
		for _, filePath := range journalFiles {
			summary, err := journal.ExtractSummary(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to extract summary from %s: %w", filePath, err)
			}
			fileName := filepath.Base(filePath)
			dateStr := strings.TrimSuffix(fileName, ".md") // Assuming .md extension
			reviewContentBuilder.WriteString(fmt.Sprintf("### %s\n%s\n\n", dateStr, summary))
		}
	}

	err = os.WriteFile(reviewFilePath, []byte(reviewContentBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write weekly review file: %w", err)
	}

	return color.GreenString("Weekly review generated at: %s", reviewFilePath), nil
}

// ReviewMonth generates a monthly review file.
func ReviewMonth(cfg *config.Config, month string, year int, summarizer ai.AISummarizer, reader io.Reader) (string, error) {
	// Calculate start and end dates for the month
	monthNum := map[string]time.Month{
		"January": time.January, "February": time.February, "March": time.March,
		"April": time.April, "May": time.May, "June": time.June,
		"July": time.July, "August": time.August, "September": time.September,
		"October": time.October, "November": time.November, "December": time.December,
	}[month]
	if monthNum == 0 {
		return "", fmt.Errorf("invalid month name: %s", month)
	}

	startDate := time.Date(year, monthNum, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Last day of the month

	// List journal files for the period
	journalFiles, err := journal.ListJournalFilesByPeriod(cfg, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to list journal files for monthly review: %w", err)
	}

	var reviewContentBuilder strings.Builder
	reviewContentBuilder.WriteString(fmt.Sprintf("# Monthly Review - %s %d\n\n", month, year))

	reviewFilePath := filepath.Join(cfg.JournalDir, fmt.Sprintf("review_month_%s_%d.md", month, year))
	if err := os.MkdirAll(filepath.Dir(reviewFilePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for monthly review file: %w", err)
	}
	err = os.WriteFile(reviewFilePath, []byte(reviewContentBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write monthly review file: %w", err)
	}

	reviewSummaryPrompt := "Write a summary of the monthly review. Use 1st person and a simple language. Use 200 characters or less."
	err = journal.GenerateSummaryIfMissing(reviewFilePath, cfg, summarizer, reviewSummaryPrompt, reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary for monthly review: %w", err)
	}

	reviewContentBytes, err := os.ReadFile(reviewFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read monthly review file after summary generation: %w", err)
	}
	reviewContentBuilder.Reset()
	reviewContentBuilder.Write(reviewContentBytes)

	if len(journalFiles) == 0 {
		reviewContentBuilder.WriteString("No journal entries found for this month.\n\n")
	} else {
		reviewContentBuilder.WriteString("## Daily Summaries\n\n")
		for _, filePath := range journalFiles {
			summary, err := journal.ExtractSummary(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to extract summary from %s: %w", filePath, err)
			}
			fileName := filepath.Base(filePath)
			dateStr := strings.TrimSuffix(fileName, ".md") // Assuming .md extension
			reviewContentBuilder.WriteString(fmt.Sprintf("### %s\n%s\n\n", dateStr, summary))
		}
	}

	err = os.WriteFile(reviewFilePath, []byte(reviewContentBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write monthly review file: %w", err)
	}

	return color.GreenString("Monthly review generated at: %s", reviewFilePath), nil
}

// ReviewYear generates a yearly review file with monthly summaries and daily entries organized by month.
func ReviewYear(cfg *config.Config, year int, summarizer ai.AISummarizer, reader io.Reader) (string, error) {
	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)

	journalFiles, err := journal.ListJournalFilesByPeriod(cfg, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to list journal files for yearly review: %w", err)
	}

	var reviewContentBuilder strings.Builder
	reviewContentBuilder.WriteString(fmt.Sprintf("# Yearly Review - %d\n\n", year))

	reviewFilePath := filepath.Join(cfg.JournalDir, fmt.Sprintf("review_year_%d.md", year))
	if err := os.MkdirAll(filepath.Dir(reviewFilePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for yearly review file: %w", err)
	}
	err = os.WriteFile(reviewFilePath, []byte(reviewContentBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write yearly review file: %w", err)
	}

	reviewSummaryPrompt := "Write a summary of the yearly review. Use 1st person and a simple language. Use 200 characters or less."
	err = journal.GenerateSummaryIfMissing(reviewFilePath, cfg, summarizer, reviewSummaryPrompt, reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary for yearly review: %w", err)
	}

	reviewContentBytes, err := os.ReadFile(reviewFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read yearly review file after summary generation: %w", err)
	}
	reviewContentBuilder.Reset()
	reviewContentBuilder.Write(reviewContentBytes)

	if len(journalFiles) == 0 {
		reviewContentBuilder.WriteString("No journal entries found for this year.\n\n")
	} else {
		// Group journal files by month
		filesByMonth := make(map[time.Month][]string)
		for _, filePath := range journalFiles {
			fileName := filepath.Base(filePath)
			dateStr := strings.TrimSuffix(fileName, ".md")
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue // Skip files that don't match expected format
			}
			filesByMonth[parsedDate.Month()] = append(filesByMonth[parsedDate.Month()], filePath)
		}

		reviewContentBuilder.WriteString("## Monthly Summaries\n\n")

		// Iterate through months in order
		for month := time.January; month <= time.December; month++ {
			files := filesByMonth[month]
			if len(files) == 0 {
				continue // Skip months with no entries
			}

			reviewContentBuilder.WriteString(fmt.Sprintf("### %s\n\n", month.String()))

			// Add daily summaries for this month
			for _, filePath := range files {
				summary, err := journal.ExtractSummary(filePath)
				if err != nil {
					return "", fmt.Errorf("failed to extract summary from %s: %w", filePath, err)
				}
				fileName := filepath.Base(filePath)
				dateStr := strings.TrimSuffix(fileName, ".md")
				reviewContentBuilder.WriteString(fmt.Sprintf("- **%s**: %s\n", dateStr, summary))
			}
			reviewContentBuilder.WriteString("\n")
		}
	}

	err = os.WriteFile(reviewFilePath, []byte(reviewContentBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write yearly review file: %w", err)
	}

	return color.GreenString("Yearly review generated at: %s", reviewFilePath), nil
}
