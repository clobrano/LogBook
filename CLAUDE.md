# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LogBook is a command-line journaling application written in Go that streamlines daily note-taking and periodic reviews (weekly, monthly, yearly). It uses template-based daily entries, optional AI-powered summarization, and a "one-line a day" feature that displays summaries from past entries.

## Commands

### Building
```bash
# Build the binary
go build -o logbook cmd/logbook/main.go

# The binary will be created as `logbook` in the current directory
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/journal
go test ./pkg/config
go test ./pkg/review

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestFunctionName ./pkg/packagename
```

### Running the Application
```bash
# Show help
./logbook help

# Create default configuration file at ~/.config/logbook/config.toml
./logbook config

# Add an entry to today's journal
./logbook log "Your journal entry text"

# Perform reviews
./logbook review week [week_number] [year]
./logbook review month [month_name] [year]
./logbook review year [year]
```

## Architecture

### Project Structure
```
cmd/logbook/         # Main application entry point
pkg/
  ├── ai/           # AI summarization interface and implementations
  ├── config/       # Configuration loading/saving (TOML)
  ├── journal/      # Core journal file operations
  ├── oneline/      # One-line note feature
  ├── review/       # Weekly/monthly/yearly review generation
  └── template/     # Template rendering engine
```

### Key Components

**Configuration (`pkg/config/`)**
- Config is loaded from `~/.config/logbook/config.toml` (TOML format)
- Key settings: `journal_dir`, `daily_file_name`, `daily_template`, `ai_enabled`, `ai_binary`, `ai_prompt`, `one_line_template`
- The `Config` struct includes an `AISummarizer` interface (not serialized to TOML)
- AI summarizer is initialized in `LoadConfig()` if `ai_enabled` is true

**Journal Management (`pkg/journal/`)**
- `CreateDailyJournalFile()`: Creates daily journal files using templates, handles AI/manual summary generation
- `AppendToLog()`: Adds timestamped entries to the "## LOG" section of daily notes
- `GenerateSummaryIfMissing()`: Generates or prompts for summaries if `[SUMMARY_PLACEHOLDER]` exists
- `ExtractSummary()`: Extracts the first paragraph after title as summary
- Daily journal structure:
  ```
  # [Date Title]

  [Summary paragraph]

  ## LOG
  HH:MM Entry text

  ## One-line note
  [Historical summaries]
  ```

**AI Integration (`pkg/ai/`)**
- `AISummarizer` interface with `GenerateSummary(text, prompt)` method
- `ExternalAISummarizer`: Executes external AI command using configurable template with placeholders
- `PlaceholderAISummarizer`: Fallback when no AI command configured
- `MockAISummarizer`: For testing
- AI command template supports placeholders: `{PROMPT}` and `{TEXT}`
- Examples:
  - Gemini: `gemini --prompt '{PROMPT} {TEXT}'`
  - Claude: `claude --text '{TEXT}' --instructions '{PROMPT}'`
  - Custom: `curl -X POST api.ai.com -d '{"prompt":"{PROMPT}","text":"{TEXT}"}'`

**Review System (`pkg/review/`)**
- `ReviewWeek()`: Generates weekly review with daily summaries (ISO week calculation)
- `ReviewMonth()`: Generates monthly review with daily summaries
- `ReviewYear()`: Generates yearly review with **monthly** summaries (groups daily entries by month as per PRD req #15)
- Review files are created in journal_dir as `review_{period}_{identifier}.md`
- Reviews extract summaries from existing journal files or generate them if missing

**One-Line Notes (`pkg/oneline/`)**
- `GetPastSummaries()`: Retrieves summaries from 1 week ago, 1 month ago, 6 months ago, and all past years (dynamically checks up to 3 years back)
- `EmbedOneLineNotes()`: Embeds summaries into "## One-line note" section
- `extractSummary()`: Private helper to extract summary from journal files
- Automatically integrated into `CreateDailyJournalFile()` - runs every time a new daily file is created

**Template Engine (`pkg/template/`)**
- Uses Go's `text/template` package
- Custom functions: `formatDate` for date formatting, `formatTime` for time formatting
- Template data includes: `Date`, `Time`, `Summary`, `Entry` fields
- Used for rendering file names, daily templates, and log entries

### Important Patterns

1. **Summary Workflow**:
   - First paragraph after title is always the summary
   - Template includes `[SUMMARY_PLACEHOLDER]` for new files
   - AI generates summary OR user enters manual summary OR skipped (shows as "missing")
   - Reviews rely on summaries from daily files

2. **Date Handling**:
   - Daily file names use template: `{{.Date | formatDate "2006-01-02"}}.md`
   - Reviews use ISO week calculation for week boundaries
   - One-line notes show summaries from: 1 week ago, 1 month ago, 6 months ago, and all past years (dynamically)

3. **Error Handling**:
   - All operations validate config first via `cfg.Validate()`
   - Journal directory must be absolute path
   - File existence checks use `os.Stat()` with `os.IsNotExist(err)` pattern

4. **Testing**:
   - Uses testify/assert for assertions
   - Mock implementations available for AI summarizer
   - Tests use temporary directories for file operations

### Recent Improvements

All PRD requirements have been implemented:

1. **Yearly Review** (Req #15): ✅ Now shows monthly summaries with daily entries grouped by month
2. **One-Line Notes Integration** (Req #17-20): ✅ Fully integrated into `CreateDailyJournalFile` - automatically embeds historical summaries
3. **One-Line Past Years** (Req #18): ✅ Dynamically checks all past years (up to 3 years back) until no entries found
4. **One-Line Periods**: ✅ Includes 1 week ago, 1 month ago, 6 months ago, and all past years
5. **LOG Entry Template** (Req #5): ✅ Now configurable via `log_entry_template` config setting (default: `{{.Time | formatTime "15:04"}} {{.Entry}}`)

## Configuration File Location

The config file is always at `~/.config/logbook/config.toml` (XDG config directory). The default config includes:
- `journal_dir`: `~/.logbook/journal`
- `daily_template`: Includes "## One-line note" section at the end
- `log_entry_template`: `{{.Time | formatTime "15:04"}} {{.Entry}}` (configurable timestamp and entry format)
- `ai_enabled`: `false` (must be explicitly enabled)
- `ai_command`: Command template with `{PROMPT}` and `{TEXT}` placeholders
  - Example: `gemini --prompt '{PROMPT} {TEXT}'`
  - Example: `claude --text '{TEXT}' --instructions '{PROMPT}'`
  - Example: `ollama run llama2 "{PROMPT}\n\n{TEXT}"`
- `ai_prompt`: Default summary prompt (200 char limit, 1st person)
