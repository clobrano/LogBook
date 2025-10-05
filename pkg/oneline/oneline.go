package oneline

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/template"
)

// GetPastSummaries retrieves summaries from past daily notes for specified periods.
// This includes: 1 week ago, 1 month ago, 6 months ago, and all past years (as far back as entries exist).
// If a file exists but has no summary and AI is enabled, it generates one.
// Returns a map with date keys in YYYY-MM-DD format.
func GetPastSummaries(cfg *config.Config, targetDate time.Time) (map[string]string, error) {
	summaries := make(map[string]string)

	// Add fixed periods: 1 week ago, 1 month ago, 6 months ago
	fixedPeriods := []time.Time{
		targetDate.AddDate(0, 0, -7),   // 1 week ago
		targetDate.AddDate(0, -1, 0),   // 1 month ago
		targetDate.AddDate(0, -6, 0),   // 6 months ago
	}

	for _, date := range fixedPeriods {
		dateKey := date.Format("2006-01-02")
		data := template.TemplateData{Date: date}
		fileName, err := template.Render(cfg.DailyFileName, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render daily file name for %s: %w", dateKey, err)
		}
		filePath := filepath.Join(cfg.JournalDir, fileName)

		summary := getSummaryWithAIFallback(filePath, cfg)
		summaries[dateKey] = summary
	}

	// Add all past years dynamically (check up to 3 years back)
	for yearsAgo := 1; yearsAgo <= 3; yearsAgo++ {
		pastDate := targetDate.AddDate(-yearsAgo, 0, 0)
		dateKey := pastDate.Format("2006-01-02")

		data := template.TemplateData{Date: pastDate}
		fileName, err := template.Render(cfg.DailyFileName, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render daily file name for %s: %w", dateKey, err)
		}
		filePath := filepath.Join(cfg.JournalDir, fileName)
		summaries[dateKey] = getSummaryWithAIFallback(filePath, cfg)
	}

	return summaries, nil
}

// getSummaryWithAIFallback gets summary from file, generates with AI if missing but file exists
// If a summary is generated, it saves it back to the file for future use
func getSummaryWithAIFallback(filePath string, cfg *config.Config) string {
	summary, err := extractSummary(filePath)
	if err != nil {
		return "missing" // File doesn't exist or can't be read
	}

	if summary == "" {
		// File exists but no summary - check if file actually has content
		content, err := os.ReadFile(filePath)
		if err != nil || len(content) == 0 {
			return "missing"
		}

		// File has content but no summary - generate with AI if available
		if cfg.AISummarizer != nil {
			// Extract LOG section content to summarize
			contentStr := string(content)
			lines := strings.Split(contentStr, "\n")

			// Find the LOG section
			logSectionStart := -1
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "# LOG") {
					logSectionStart = i + 1
					break
				}
			}

			if logSectionStart != -1 && logSectionStart < len(lines) {
				// Extract content from LOG section to end of file
				contentToSummarize := strings.Join(lines[logSectionStart:], "\n")
				contentToSummarize = strings.TrimSpace(contentToSummarize)

				if len(contentToSummarize) > 0 {
					generatedSummary, err := cfg.AISummarizer.GenerateSummary(contentToSummarize, cfg.AIPrompt)
					if err == nil && generatedSummary != "" {
						// Save the generated summary back to the file
						err = saveSummaryToFile(filePath, generatedSummary)
						if err == nil {
							return generatedSummary
						}
						// If saving failed, still return the summary but it won't be cached
						return generatedSummary
					}
				}
			}
		}
		return "missing" // Couldn't generate summary
	}

	return summary
}

// saveSummaryToFile inserts a summary into a journal file right after the title and HTML comment
func saveSummaryToFile(filePath string, summary string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("file %s is empty", filePath)
	}

	// Build new content with summary inserted after title and optional HTML comment
	var newContentBuilder strings.Builder
	newContentBuilder.WriteString(lines[0]) // Title
	newContentBuilder.WriteString("\n")

	// Check if line 1 is HTML comment, if so include it
	startIdx := 1
	if len(lines) > 1 && strings.HasPrefix(strings.TrimSpace(lines[1]), "<!--") {
		newContentBuilder.WriteString(lines[1])
		newContentBuilder.WriteString("\n")
		startIdx = 2
	}

	// Insert the summary
	newContentBuilder.WriteString(strings.TrimSpace(summary))
	newContentBuilder.WriteString("\n\n")

	// Skip any empty lines after comment
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		startIdx++
	}

	// Append the rest of the file
	if startIdx < len(lines) {
		newContentBuilder.WriteString(strings.Join(lines[startIdx:], "\n"))
	}

	modifiedContent := newContentBuilder.String()

	err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write summary to file %s: %w", filePath, err)
	}

	return nil
}

// extractSummary reads a journal file and returns its first paragraph as the summary.
func extractSummary(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File does not exist, return empty summary and no error
		}
		return "", fmt.Errorf("failed to read journal file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")

	// The first paragraph after the title and before the "LOG" chapter is considered the summary.
	var summaryLines []string
	readingSummary := false

	for i := 1; i < len(lines); i++ {
		trimmedLine := strings.TrimSpace(lines[i])

		if strings.HasPrefix(trimmedLine, "# LOG") || strings.HasPrefix(trimmedLine, "# One-line note") {
			break // Reached the LOG or One-line note section, stop reading summary
		}

		if trimmedLine == "" {
			if readingSummary { // If we were reading summary and hit an empty line, the paragraph ends
				break
			}
			continue // Skip empty lines before the summary starts
		}

		// Skip HTML comments
		if strings.HasPrefix(trimmedLine, "<!--") {
			continue
		}

		if !readingSummary && strings.HasPrefix(trimmedLine, "#") {
			continue // Skip any sub-headings before the actual summary paragraph
		}

		readingSummary = true
		summaryLines = append(summaryLines, trimmedLine)
	}

	if len(summaryLines) > 0 {
		return strings.Join(summaryLines, " "), nil
	}

	return "", nil // No summary found
}

// EmbedOneLineNotes embeds one-line summaries into the "One-line note" section of a daily note.
// If one-line notes already exist, it skips embedding to avoid duplicates.
func EmbedOneLineNotes(filePath string, summaries map[string]string) error {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := string(contentBytes)
	oneLineNoteSection := "# One-line note\n"

	// Find the "One-line note" section
	idx := strings.Index(content, oneLineNoteSection)
	if idx == -1 {
		return fmt.Errorf("\"One-line note\" section not found in file %s", filePath)
	}

	// Find where to insert/replace one-line notes
	afterSection := idx + len(oneLineNoteSection)

	// Find the end of the one-line note section (next # header or end of file)
	endOfSection := afterSection
	lines := strings.Split(content[afterSection:], "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Found next section
			endOfSection = afterSection + strings.Index(content[afterSection:], trimmed)
			break
		}
		if i == len(lines)-1 {
			// End of file
			endOfSection = len(content)
		}
	}

	// Build the one-line notes content
	var oneLineNotesBuilder strings.Builder

	// Extract and sort dates in reverse chronological order (most recent first)
	var dates []string
	for dateKey := range summaries {
		dates = append(dates, dateKey)
	}
	// Sort in reverse chronological order
	sort.Strings(dates)
	// Reverse the slice to get most recent first
	for i, j := 0, len(dates)-1; i < j; i, j = i+1, j-1 {
		dates[i], dates[j] = dates[j], dates[i]
	}

	// Format each entry with wikilink
	for _, dateKey := range dates {
		summary := summaries[dateKey]
		oneLineNotesBuilder.WriteString(fmt.Sprintf("* [[%s]]: %s\n", dateKey, summary))
	}
	oneLineNotesBuilder.WriteString("\n")

	// Replace the one-line notes section content
	updatedContent := content[:afterSection] + oneLineNotesBuilder.String() + content[endOfSection:]

	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated content to %s: %w", filePath, err)
	}

	return nil
}
