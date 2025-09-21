package oneline

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/journal"
	"github.com/clobrano/LogBook/pkg/template"
)

// GetPastSummaries retrieves summaries from past daily notes for specified periods.
func GetPastSummaries(cfg *config.Config, targetDate time.Time) (map[string]string, error) {
	summaries := make(map[string]string)

	periods := map[string]time.Time{
		"1_week_ago":   targetDate.AddDate(0, 0, -7),
		"1_month_ago":  targetDate.AddDate(0, -1, 0),
		"6_months_ago": targetDate.AddDate(0, -6, 0),
		"1_year_ago":   targetDate.AddDate(-1, 0, 0),
		"2_years_ago":  targetDate.AddDate(-2, 0, 0),
	}

	for key, date := range periods {
		data := template.TemplateData{Date: date}
		fileName, err := template.Render(cfg.DailyFileName, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render daily file name for %s: %w", key, err)
		}
		filePath := filepath.Join(cfg.JournalDir, fileName)

		summary, err := journal.ExtractSummary(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract summary from %s: %w", filePath, err)
		}

		if summary == "" {
			summaries[key] = "missing"
		} else {
			summaries[key] = summary
		}
	}

	return summaries, nil
}