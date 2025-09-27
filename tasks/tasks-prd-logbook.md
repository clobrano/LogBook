## Relevant Files

- `main.go` - Main application entry point and CLI command parsing.
- `config.go` - Handles TOML configuration parsing, loading, and saving.
- `journal.go` - Core logic for daily entry creation, reading, writing, and parsing.
- `template.go` - Handles dynamic templating for daily notes and review files.
- `ai.go` - Interface and implementation for AI agent integration and summary generation.
- `review.go` - Logic for weekly, monthly, and yearly review commands and summary aggregation.
- `oneline.go` - Logic for generating and embedding one-line summaries from past notes.
- `utils.go` - General utility functions (e.g., date handling, file system operations).
- `cmd/logbook/main.go` - (If using Cobra/Viper for CLI) Main entry point for the `logbook` command.
- `cmd/logbook/daily.go` - (If using Cobra/Viper for CLI) Handles `logbook daily` command for creating/editing daily notes.
- `cmd/logbook/log.go` - (If using Cobra/Viper for CLI) Handles `logbook log` command for adding entries to the "LOG" chapter.
- `cmd/logbook/review.go` - (If using Cobra/Viper for CLI) Handles `logbook review` commands.
- `go.mod` - Go module definition file.
- `go.sum` - Go module checksums.
- `README.md` - Application documentation.
- `test/config_test.go` - Unit tests for `config.go`.
- `test/journal_test.go` - Unit tests for `journal.go`.
- `test/template_test.go` - Unit tests for `template.go`.
- `test/ai_test.go` - Unit tests for `ai.go`.
- `test/review_test.go` - Unit tests for `review.go`.
- `test/oneline_test.go` - Unit tests for `oneline.go`.
- `test/utils_test.go` - Unit tests for `utils.go`.

### Notes

- Unit tests should typically be placed alongside the code files they are testing (e.g., `config.go` and `config_test.go` in the same directory or a `test/` subdirectory).
- Use `go test ./...` to run all tests.
- **TDD Approach:** For each task, write the unit tests first, ensure they fail, then write the minimum amount of code to make them pass, and finally refactor.

## Tasks

- [x] 1.0 Set up Project Structure and Basic CLI
  - [x] 1.1 Initialize Go module and create basic directory structure.
  - [x] 1.2 Implement basic CLI command parsing (e.g., using `flag` or `cobra`) using TDD.
  - [x] 1.3 Implement `logbook help` command using TDD.
- [x] 2.0 Configuration Management
  - [x] 2.1 Define configuration struct for TOML parsing (file location, naming convention, AI prompt, etc.) using TDD.
  - [x] 2.2 Implement function to load configuration from a TOML file using TDD.
  - [x] 2.3 Implement function to save configuration to a TOML file using TDD.
  - [x] 2.4 Implement validation for configuration parameters using TDD.
- [x] 3.0 Daily Entry Creation and Management
  - [x] 3.1 Implement function to create a new daily journal file based on the current date and naming convention using TDD.
  - [x] 3.2 Implement templating engine for daily notes, including dynamic date insertion using TDD.
  - [x] 3.3 Implement logic to ensure the first paragraph is the summary using TDD.
  - [x] 3.4 Implement logic to create the "LOG" chapter after the summary using TDD.
  - [x] 3.5 Implement `logbook log <text>` command to append entries to the "LOG" chapter with timestamps using TDD.
- [x] 4.0 AI Agent Integration for Summaries
  - [x] 4.1 Define an interface for AI agent interaction using TDD.
  - [x] 4.2 Implement a concrete AI agent (e.g., a placeholder or a simple local model integration) using TDD.
  - [x] 4.3 Implement logic to generate a summary using the AI agent if no summary exists using TDD.
  - [x] 4.4 Implement logic to prompt the user for a manual summary if no AI agent is configured and no summary exists using TDD.
- [ ] 5.0 Review Functionality (Weekly, Monthly, Yearly)
  - [x] 5.1 Implement `logbook review week [week number] [year]` command using TDD.
  - [x] 5.2 Implement `logbook review month [month name] [year]` command using TDD.
  - [x] 5.3 Implement `logbook review year [year]` command using TDD.
  - [x] 5.4 Implement logic to identify relevant daily entries for each review period using TDD.
  - [x] 5.5 Implement logic to aggregate summaries of daily entries into a review file using TDD.
  - [x] 5.6 Implement logic to generate a summary for the review period (manual or AI-generated) using TDD.
- [x] 6.0 One-Line Note Feature
  - [x] 6.1 Implement logic to retrieve summaries from past daily notes (1 week, 1 month, 6 months, past years) using TDD.
  - [x] 6.2 Implement logic to format and embed these summaries into the "One-line note" section of the current daily note using TDD.
  - [x] 6.3 Implement logic to display "missing" for notes that don't exist for specific past dates using TDD.
  - [x] 6.4 Ensure the "One-line note" section is ignored by the AI agent during summary generation for the main daily note using TDD.
- [ ] 7.0 Error Handling and User Feedback
  - [x] 7.1 Implement robust error handling for file operations, configuration, and AI interactions (Config, Journal, AI, Review packages) using TDD.
  - [x] 7.2 Implement clear and concise user feedback messages (success, error, warnings) using TDD.
  - [ ] 7.3 Explore optional colorful output for the CLI using TDD.
  - [ ] 7.4 Consider `github.com/rivo/tview` for interactive prompts if needed using TDD.
- [ ] 8.0 Documentation and Testing
  - [ ] 8.1 Update `README.md` with installation instructions, usage examples, and configuration details.
  - [ ] 8.2 Ensure all code adheres to Go best practices and style guides.
