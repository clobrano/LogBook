# Product Requirements Document: LogBook

## 1. Introduction/Overview

LogBook is a command-line application written in Go, designed to streamline the daily journaling process and facilitate periodic reviews (weekly, monthly, yearly). It aims to help users, particularly writers, maintain a consistent journaling habit by providing templating capabilities for daily entries and automating the tedious process of reviewing past notes. Additionally, it supports a "one-line a day" note-taking model, displaying summaries of past entries within current daily notes.

**Goal:** Streamline Journaling by providing helpers to write journal entries daily and to perform weekly, monthly, and yearly reviews according to user-defined templates.

## 2. Goals

*   Enable users to quickly create daily journal entries based on configurable templates.
*   Automate the generation of summaries for daily entries, optionally leveraging local AI agents.
*   Provide a structured way to perform weekly, monthly, and yearly reviews, aggregating summaries of past entries.
*   Integrate a "one-line a day" feature that displays summaries of older notes within current daily entries.
*   Ensure a user-friendly command-line interface with clear feedback and optional colorful output.

## 3. User Stories

*   **Daily Entry:**
    *   As a user, I want to quickly jot a note for today without worrying about creating the appropriate format.
    *   As a user, I want to quickly jot a note with the hour and minute I am writing the note.
    *   As a user, I want my daily note to include a summary from a week ago, a month ago, and the same date in previous years.
*   **Weekly Review:**
    *   As a user, I want to start a weekly review in a new file with the appropriate template, having a summary of each day of the week.
*   **One-line Note:**
    *   As a user, I like to open a daily note file and read summaries from notes of older days.

## 4. Functional Requirements

1.  The application must allow users to define the location for storing journal entry files.
2.  The application must support a configurable file naming convention for daily entries (defaulting to `YYYY-MM-DD.md`).
3.  The application must support templating for daily note files, including dynamic dates (e.g., `# Sept 11 2025 Monday`).
4.  After the first paragraph (summary), each daily note must contain a chapter titled "LOG" where daily entries are added.
5.  Daily entries within the "LOG" chapter can be added manually or via the `logbook log <some text to add>` command, following a configurable template.
6.  The application must support AI agent extensions to add a summary of the note at the top of the file.
7.  The prompt for the AI agent summary must be configurable (default: "Write a summary of the note at the given file. Use 1st person and a simple language. Use 200 characters or less").
8.  The configuration file for templates and other settings must be in TOML format and located in the XDG default configuration directory.
9.  The application must provide commands to initiate reviews:
    *   `logbook review week [week number] [year]`
    *   `logbook review month [month name] [year]`
    *   `logbook review year [year]`
8.  During a review, the application must report the summary of each item in the review period into a given file.
9.  For each file (daily note, weekly review, etc.), the first paragraph must be considered its summary.
10. If a summary paragraph exists, it must be used by the review functions.
11. If a summary paragraph does not exist and an AI agent is configured, the application can generate the summary using the AI agent.
12. If a summary paragraph does not exist and no AI agent is configured, the application must prompt the user to write a summary manually or skip it. Skipped summaries will not be included in reviews.
13. For weekly reviews, the application must report summaries of each day of the week and provide a summary of the week (manual or AI-generated).
14. For monthly reviews, the application must report summaries of each day of the month and provide a summary of the month (manual or AI-generated).
15. For yearly reviews, the application must report summaries of each month and provide a summary of the year (manual or AI-generated).
16. The application must identify entries for review periods based on the date in each daily note.
17. The last paragraph of each daily note must be the "One-line note" section.
18. The "One-line note" section must contain summaries of daily notes from:
    *   One week before
    *   One month before
    *   Past years (1 year before, 2 years before, 3 years before, and so on, as long as notes exist for those dates).
19. The "One-line note" section must be formatted as `<date>: <note>` for each line. The AI agent must receive a prompt that specifically instructs it to ignore the content in this "One-line note" section during reviews.
20. When no past entries exist for a specific date in the "One-line note" feature (e.g., a week ago, a month ago, years ago), the `<note>` for that line will be displayed as "missing".
21. The application must use plain text files for storing all journal data.

## 5. Non-Goals (Out of Scope)

*   No specific non-goals were identified for the initial version.

## 6. Design Considerations (Optional)

*   The command-line interface should be user-friendly, potentially incorporating color for better readability.
*   Error messages and confirmations should be clear and concise.
*   Consider using `github.com/rivo/tview` for more complex TUI elements if necessary for displaying errors or confirmations.

## 7. Technical Considerations (Optional)

*   Prioritize using default Go libraries as much as possible.
*   For external libraries, prefer well-established and widely used ones.
*   Data storage will be exclusively in plain text files, consistent with the journaling format.

## 8. Success Metrics

*   Daily entries are consistently created with the defined template.
*   Review files are generated accurately with summaries of the respective periods.
*   One-line summaries from past entries are correctly displayed in daily notes.
*   User feedback indicates a streamlined and efficient journaling and review process.


