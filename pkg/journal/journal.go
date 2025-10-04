package journal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/ai"
	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/template"

	"github.com/fatih/color"
)

// CreateDailyJournalFile creates a new daily journal file based on the current date and configuration.
func CreateDailyJournalFile(cfg *config.Config, date time.Time, summarizer ai.AISummarizer, reader io.Reader) (string, string, error) {
	if err := cfg.Validate(); err != nil {
		return "", "", fmt.Errorf("invalid configuration: %w", err)
	}

	journalDir := cfg.JournalDir
	if !filepath.IsAbs(journalDir) {
		return "", "", fmt.Errorf("JournalDir must be an absolute path: %s", journalDir)
	}

	if _, err := os.Stat(journalDir); os.IsNotExist(err) {
		// Create the journal directory if it doesn't exist
		if err := os.MkdirAll(journalDir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create journal directory: %w", err)
		}
	}

	// Render the file name using the template engine

	data := template.TemplateData{Date: date}
	fileName, err := template.Render(cfg.DailyFileName, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to render daily file name: %w", err)
	}

	filePath := filepath.Join(journalDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return filePath, color.GreenString("Daily journal file already exists: %s", filePath), nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create daily journal file: %w", err)
	}
	defer file.Close()

	// Render the daily template and write to file
	templateContent, err := template.Render(cfg.DailyTemplate, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to render daily template: %w", err)
	}

	_, err = file.WriteString(templateContent)
	if err != nil {
		return "", "", fmt.Errorf("failed to write daily template to file: %w", err)
	}

	// Generate summary for the daily file if missing
	dailySummaryPrompt := cfg.AIPrompt
	err = GenerateSummaryIfMissing(filePath, cfg, summarizer, dailySummaryPrompt, reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate summary for daily journal: %w", err)
	}

	return filePath, color.GreenString("Daily journal file created: %s", filePath), nil
}

// AppendToLog appends a new entry to the "LOG" chapter of a daily journal file.
func AppendToLog(cfg *config.Config, filePath, entry string, timestamp time.Time) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read journal file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	logChapterIndex := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "## LOG") {
			logChapterIndex = i
			break
		}
	}

	if logChapterIndex == -1 {
		return fmt.Errorf("LOG chapter not found in file: %s", filePath)
	}

	// Find the insertion point: after the "## LOG" line, skip any subsequent empty lines, ...
	insertIndex := logChapterIndex + 1
	for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
		insertIndex++
	}
	// ... then find where the last already existing entry lies
	for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) != "" {
		insertIndex++
	}

	newEntryLine := fmt.Sprintf("%02d:%02d %s", timestamp.Hour(), timestamp.Minute(), entry)

	// Insert the new entry
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, newEntryLine)
	newLines = append(newLines, lines[insertIndex:]...)

	modifiedContent := strings.Join(newLines, "\n")

	// Ensure the file ends with a single newline
	if !strings.HasSuffix(modifiedContent, "\n") {
		modifiedContent += "\n"
	}

	err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to journal file: %w", err)
	}

	fmt.Println(color.GreenString("Log entry appended to %s", filePath))
	return nil
}

// GenerateSummaryIfMissing reads a journal file, and if no summary exists, generates one using the provided AI summarizer.
func GenerateSummaryIfMissing(filePath string, cfg *config.Config, summarizer ai.AISummarizer, aiPrompt string, reader io.Reader) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read journal file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Check if a summary already exists (first non-empty paragraph after the title)
	// A summary is considered missing if:
	// 1. The file has less than 3 lines (title, empty line, placeholder/summary)
	// 2. The second line (index 1) is empty AND the third line (index 2) is either empty or contains the placeholder.
	// 3. The second line (index 1) is not empty, not a heading, and not the placeholder.

	isSummaryMissing := false
	if len(lines) < 3 {
		isSummaryMissing = true
	} else {
		trimmedLine1 := strings.TrimSpace(lines[1])
		trimmedLine2 := strings.TrimSpace(lines[2])

		if trimmedLine1 == "" && (trimmedLine2 == "" || trimmedLine2 == "[SUMMARY_PLACEHOLDER]") {
			isSummaryMissing = true
		}
	}

	if !isSummaryMissing {
		fmt.Println("DEBUG: Summary already exists, do nothing")
		return nil
	}

	var finalSummary string

	if summarizer != nil {
		fmt.Println("DEBUG: Entering AI path")
		// Extract the content to be summarized (excluding the title, summary/placeholder, and "One-line note" section)
		contentLines := make([]string, 0, len(lines))
		// Skip title (lines[0])
		// Skip potential empty line (lines[1]) and placeholder/summary (lines[2])
		// Start from lines[3] if placeholder was present, otherwise from lines[1] or lines[2] depending on actual content
		startIndex := 1 // Default to start after title
		if len(lines) > 2 && strings.TrimSpace(lines[2]) == "[SUMMARY_PLACEHOLDER]" {
			startIndex = 3 // Skip title, empty line, and placeholder
		} else if len(lines) > 1 && strings.TrimSpace(lines[1]) == "" {
			startIndex = 2 // Skip title and empty line
		}

		for i := startIndex; i < len(lines); i++ {
			contentLines = append(contentLines, lines[i])
		}
		contentToSummarize := strings.Join(contentLines, "\n")
		oneLineNoteSection := "## One-line note"

		fmt.Println("DEBUG: content to summarize: ", contentToSummarize)
		idx := strings.Index(contentToSummarize, oneLineNoteSection)
		if idx != -1 {
			contentToSummarize = contentToSummarize[:idx]
		}

		// Generate summary using AI agent
		generatedSummary, err := summarizer.GenerateSummary(contentToSummarize, aiPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate summary with AI: %w", err)
		}
		finalSummary = generatedSummary
	} else {
		fmt.Println("DEBUG: Entering Manual path")
		// Prompt user for manual summary
		fmt.Print("No AI agent configured or provided. Please enter a manual summary (or leave blank to skip): ")
		scanner := bufio.NewScanner(reader)
		if scanner.Scan() {
			finalSummary = scanner.Text()
		} else {
			return fmt.Errorf("failed to read manual summary: %w", scanner.Err())
		}

		if strings.TrimSpace(finalSummary) == "" {
			fmt.Println(color.YellowString("Manual summary skipped."))
			return nil // User skipped manual summary
		}
	}

	// Insert the generated summary after the title, replacing the placeholder if it exists
	var newContentBuilder strings.Builder
	newContentBuilder.WriteString(lines[0]) // Title
	newContentBuilder.WriteString("\n")

	// If there was a placeholder, replace it. Otherwise, insert after the empty line.
	if len(lines) > 2 && strings.TrimSpace(lines[2]) == "[SUMMARY_PLACEHOLDER]" {
		newContentBuilder.WriteString(strings.TrimSpace(finalSummary))
		newContentBuilder.WriteString("\n\n")                        // Two newlines after the summary
		newContentBuilder.WriteString(strings.Join(lines[3:], "\n")) // Rest of the content, skipping title, empty line, and placeholder
	} else {
		newContentBuilder.WriteString("\n") // Keep the empty line after title
		newContentBuilder.WriteString(strings.TrimSpace(finalSummary))
		newContentBuilder.WriteString("\n\n")                        // Two newlines after the summary
		newContentBuilder.WriteString(strings.Join(lines[2:], "\n")) // Rest of the content, skipping title and empty line
	}

	modifiedContent := newContentBuilder.String()

	err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write generated summary to file: %w", err)
	}

	return nil
}

// ListJournalFilesByPeriod returns a list of absolute paths to journal files within the specified date range.
func ListJournalFilesByPeriod(cfg *config.Config, startDate, endDate time.Time) ([]string, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	journalDir := cfg.JournalDir
	if !filepath.IsAbs(journalDir) {
		return nil, fmt.Errorf("JournalDir must be an absolute path: %s", journalDir)
	}

	var files []string

	// Iterate through the date range
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Render the file name for the current date
		data := template.TemplateData{Date: d}
		fileName, err := template.Render(cfg.DailyFileName, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render daily file name for date %s: %w", d.Format("2006-01-02"), err)
		}
		filePath := filepath.Join(journalDir, fileName)

		// Check if the file exists
		if _, err := os.Stat(filePath); err == nil {
			files = append(files, filePath)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to check file %s: %w", filePath, err)
		}
	}
	return files, nil
}

// ExtractSummary reads a journal file and returns its first paragraph as the summary.
func ExtractSummary(filePath string) (string, error) {
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

		if strings.HasPrefix(trimmedLine, "## LOG") {
			break // Reached the LOG chapter, stop reading summary
		}

		if trimmedLine == "" {
			if readingSummary { // If we were reading summary and hit an empty line, the paragraph ends
				break
			}
			continue // Skip empty lines before the summary starts
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
func EmbedOneLineNotes(filePath string, summaries map[string]string) error {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := string(contentBytes)
	oneLineNoteSection := "## One-line note\n\n"

	// Find the "One-line note" section
	idx := strings.Index(content, oneLineNoteSection)
	if idx == -1 {
		return fmt.Errorf("\"One-line note\" section not found in file %s", filePath)
	}

	// Build the one-line notes content
	var oneLineNotesBuilder strings.Builder
	for key, summary := range summaries {
		oneLineNotesBuilder.WriteString(fmt.Sprintf("- %s: %s\n", strings.ReplaceAll(key, "_", " "), summary))
	}

	// Insert the one-line notes after the "## One-line note" line and its immediate newline
	insertionPoint := idx + len(oneLineNoteSection)
	updatedContent := content[:insertionPoint] + oneLineNotesBuilder.String() + content[insertionPoint:]

	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated content to %s: %w", filePath, err)
	}

	return nil
}
