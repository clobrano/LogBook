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
	"github.com/clobrano/LogBook/pkg/oneline"
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

	// Use hardcoded template
	templateContent := fmt.Sprintf("# %s\n<!-- add today summary below this line. If missing, the AI will generate one for you according to configuration file -->\n\n# One-line note\n\n# LOG\n\n", date.Format("Jan 02 2006 Monday"))

	_, err = file.WriteString(templateContent)
	if err != nil {
		return "", "", fmt.Errorf("failed to write daily template to file: %w", err)
	}

	return filePath, color.GreenString("Daily journal file created: %s", filePath), nil
}

// FinalizeDailyFile embeds one-line notes for a daily journal file.
// This should be called after all log entries have been added for the day.
func FinalizeDailyFile(cfg *config.Config, filePath string, date time.Time) error {
	// Embed one-line notes from past entries
	pastSummaries, err := oneline.GetPastSummaries(cfg, date)
	if err != nil {
		return fmt.Errorf("failed to get past summaries for one-line notes: %w", err)
	}

	err = oneline.EmbedOneLineNotes(filePath, pastSummaries)
	if err != nil {
		return fmt.Errorf("failed to embed one-line notes: %w", err)
	}

	return nil
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
		if strings.HasPrefix(line, "# LOG") {
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

	// Render the log entry using the configurable template
	data := template.TemplateData{
		Time:  timestamp,
		Entry: entry,
	}
	newEntryLine, err := template.Render(cfg.LogEntryTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render log entry template: %w", err)
	}

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
// Summary is inserted right after the first header line.
func GenerateSummaryIfMissing(filePath string, cfg *config.Config, summarizer ai.AISummarizer, aiPrompt string, reader io.Reader) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read journal file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Skip YAML frontmatter if present
	fmEnd := skipFrontmatter(lines)

	// Check if summary already exists:
	// First line after frontmatter: # Title
	// Next line: might be HTML comment (<!-- ... -->)
	// Summary exists if there's non-empty, non-comment, non-header content after title

	isSummaryMissing := true
	for i := fmEnd + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue // Skip empty lines
		}
		if strings.HasPrefix(trimmed, "<!--") {
			continue // Skip HTML comments
		}
		if strings.HasPrefix(trimmed, "#") {
			break // Hit a section header, no summary found
		}
		// Found non-empty, non-comment, non-header content = summary exists
		isSummaryMissing = false
		break
	}

	if !isSummaryMissing {
		return nil // Summary already exists
	}

	var finalSummary string

	if summarizer != nil {
		// Extract content to summarize (skip frontmatter and title, exclude "One-line note" section)
		contentToSummarize := strings.Join(lines[fmEnd+1:], "\n")
		oneLineNoteSection := "## One-line note"
		idx := strings.Index(contentToSummarize, oneLineNoteSection)
		if idx != -1 {
			contentToSummarize = contentToSummarize[:idx]
		}
		contentToSummarize = strings.TrimSpace(contentToSummarize)

		// Generate summary using AI agent
		generatedSummary, err := summarizer.GenerateSummary(contentToSummarize, aiPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate summary with AI: %w", err)
		}
		finalSummary = generatedSummary
	} else {
		// Prompt user for manual summary
		fmt.Print("No AI agent configured. Please enter a manual summary (or leave blank to skip): ")
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

	// Insert summary after frontmatter, title, and HTML comment (if present)
	var newContentBuilder strings.Builder

	// Write frontmatter if present
	for i := 0; i < fmEnd; i++ {
		newContentBuilder.WriteString(lines[i])
		newContentBuilder.WriteString("\n")
	}

	newContentBuilder.WriteString(lines[fmEnd]) // Title
	newContentBuilder.WriteString("\n")

	// Check if next line after title is HTML comment, if so include it
	startIdx := fmEnd + 1
	if startIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[startIdx]), "<!--") {
		newContentBuilder.WriteString(lines[startIdx])
		newContentBuilder.WriteString("\n")
		startIdx++
	}

	newContentBuilder.WriteString(strings.TrimSpace(finalSummary))
	newContentBuilder.WriteString("\n\n")

	// Skip any empty lines after comment
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		startIdx++
	}

	if startIdx < len(lines) {
		newContentBuilder.WriteString(strings.Join(lines[startIdx:], "\n"))
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

// skipFrontmatter returns the index of the first line after YAML frontmatter.
// If no frontmatter is present, returns 0.
// Frontmatter is delimited by "---" at the start and end.
func skipFrontmatter(lines []string) int {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return i + 1
		}
	}
	return 0 // Unclosed frontmatter, treat as no frontmatter
}

// isSectionHeader reports whether a trimmed line is a markdown heading
// at any level whose first word matches the given section name (e.g. "LOG"
// matches both "# LOG" and "## LOG ...").
func isSectionHeader(trimmed, name string) bool {
	if !strings.HasPrefix(trimmed, "#") {
		return false
	}
	rest := strings.TrimLeft(trimmed, "#")
	if rest == trimmed {
		return false // Not a heading (no leading #)
	}
	rest = strings.TrimLeft(rest, " \t")
	return rest == name || strings.HasPrefix(rest, name+" ") || strings.HasPrefix(rest, name+"\t")
}

// isLogOrOneLineSection reports whether a trimmed line marks the LOG or
// One-line note section, regardless of heading level.
func isLogOrOneLineSection(trimmed string) bool {
	return isSectionHeader(trimmed, "LOG") || isSectionHeader(trimmed, "One-line note")
}

// findTitleLine returns the index of the title line: the first markdown
// heading line that isn't a known section marker ("LOG" / "One-line note").
// Leading blank lines and YAML frontmatter (delimited by "---") are skipped.
// If frontmatter is unclosed or malformed, the scan continues line-by-line
// until a heading is found, so metadata can never be mistaken for the title.
// Returns -1 if no title is found.
func findTitleLine(lines []string) int {
	i := 0
	// Skip leading blank lines
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	// Skip well-formed YAML frontmatter
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "---" {
				i = j + 1
				break
			}
		}
	}
	// Find the first markdown heading line that is the title.
	// If frontmatter was unclosed, this scan will safely walk past metadata
	// (which uses YAML "key: value" lines, not markdown headings).
	for ; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Section markers are not the title.
		if isLogOrOneLineSection(trimmed) {
			return -1
		}
		// Must be an actual heading ("# Something"), not a hashtag or "#word".
		rest := strings.TrimLeft(trimmed, "#")
		if !strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "\t") {
			continue
		}
		return i
	}
	return -1
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

	// Locate the title line robustly (skips frontmatter and leading blanks).
	titleIdx := findTitleLine(lines)
	if titleIdx < 0 {
		return "", nil
	}

	// The first paragraph after the title and before the "LOG" chapter is considered the summary.
	var summaryLines []string
	readingSummary := false

	for i := titleIdx + 1; i < len(lines); i++ {
		trimmedLine := strings.TrimSpace(lines[i])

		if isLogOrOneLineSection(trimmedLine) {
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

