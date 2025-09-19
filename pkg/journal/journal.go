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
)

// CreateDailyJournalFile creates a new daily journal file based on the current date and configuration.
func CreateDailyJournalFile(cfg *config.Config, date time.Time) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("invalid configuration: %w", err)
	}

	journalDir := cfg.JournalDir
	if !filepath.IsAbs(journalDir) {
		return "", fmt.Errorf("JournalDir must be an absolute path: %s", journalDir)
	}

	if _, err := os.Stat(journalDir); os.IsNotExist(err) {
		// Create the journal directory if it doesn't exist
		if err := os.MkdirAll(journalDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create journal directory: %w", err)
		}
	}

	// Render the file name using the template engine
	data := template.TemplateData{Date: date}
	fileName, err := template.Render(cfg.DailyFileName, data)
	if err != nil {
		return "", fmt.Errorf("failed to render daily file name: %w", err)
	}

	filePath := filepath.Join(journalDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil // File already exists, return its path
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create daily journal file: %w", err)
	}
	defer file.Close()

	// Render the daily template and write to file
	templateContent, err := template.Render(cfg.DailyTemplate, data)
	if err != nil {
		return "", fmt.Errorf("failed to render daily template: %w", err)
	}

	_, err = file.WriteString(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to write daily template to file: %w", err)
	}

	return filePath, nil
}

// AppendToLog appends a new entry to the "LOG" chapter of a daily journal file.
func AppendToLog(filePath, entry string, timestamp time.Time) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read journal file: %w", err)
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
	if len(lines) > 1 && strings.TrimSpace(lines[1]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[1]), "#") {
		// Summary already exists, do nothing
		return nil
	}

	var finalSummary string

	if summarizer != nil {
		fmt.Println("DEBUG: Entering AI path")
		// Extract the content to be summarized (excluding the title and potential existing summary placeholder)
		// For now, let's assume the entire content after the title needs summarizing.
		// TODO: Refine content extraction to ignore "One-line note" section later.
		contentToSummarize := strings.Join(lines, "\n")

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
			fmt.Println("Manual summary skipped.")
			return nil // User skipped manual summary
		}
	}

	// Insert the generated summary after the title
	var newContentBuilder strings.Builder
	newContentBuilder.WriteString(lines[0]) // Title
	newContentBuilder.WriteString("\n")
	newContentBuilder.WriteString(strings.TrimSpace(finalSummary))
	newContentBuilder.WriteString("\n\n") // Two newlines after the summary
	newContentBuilder.WriteString(strings.Join(lines[2:], "\n")) // Rest of the content, skipping the initial empty line

	modifiedContent := newContentBuilder.String()

	err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write generated summary to file: %w", err)
	}

	return nil
}
