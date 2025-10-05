# Clarifying Questions for LogBook PRD

Please provide answers to the following questions to help me create a detailed Product Requirements Document for LogBook:

## 1. Problem/Goal
a. What is the primary problem LogBook aims to solve for its users? Provide helpers to write journal entries daily and to perform weekly, monthly and yearly reviews according to templates the user can define in a configuration file
b. What is the main goal we want to achieve with this application? Streamline Journaling

## 2. Target User
a. Who is the primary user of LogBook? (e.g., developers, writers, students, general public): writers
b. What are their typical needs or pain points related to journaling or note-taking?: While writing daily entries is quick and painless, the process to review the notes on specific intervals is at the same time critical and tedious. The application can make even quicker to create the daily note file from a template, help with single entries, following the model of interstitial journalling, and leverage external tools (e.g. AI agents installed in the same machine) to enhance the notes

## 3. Core Functionality - Daily Journaling
a. Can you describe the process a user will follow to create a daily entry?
b. What kind of templating capabilities are expected? (e.g., predefined fields, dynamic dates, custom sections):
  I expect to be able to define:
  * files location: mandatory
  * file naming convention (default to YYYY-MM-DD.md)
  * a template of the daily note file (default to today's date, e.g. `# Sept 11 2025 Monday`)
  * AI agent extensions: if enabled, the AI agent of choice (e.g. gemini) installed locally, must be able to add a summary of the note at the top of the file as first sentence. The prompt for such summary must be configurable, but default to something simple we can define later
c. How will the configuration file for templates be structured? (e.g., YAML, TOML, JSON): toml

## 4. Core Functionality - Reviews (Weekly, Monthly, Yearly)
a. How should a user initiate a review? (e.g., a specific command, automatically triggered)
* logbook review week [week number] [year]
* logbook review month [month name] [year]
* logbook review year [year]

b. What information should be presented during a review? (e.g., summaries of entries, specific prompts, aggregated data)
* the application should report the summary of each item in the review in a given file. E.g. for a weekly review, it will report in the destination file the summary of each day of the week, according to the "AI summary" functionality described above. If the summary is already present, the app will use it, if not it will first generate the summary for the day. Finally a summary of the week will be given, either manually from the user or via AI agent
* the same goes for the monthly review, while for the yearly review we will start for the summaries of each months, to cut it short
c. How will the application identify entries for a specific review period (week, month, year)?
* each daily note has a date

## 5. Core Functionality - "One-line a day" Note Taking
a. How does this differ from a regular daily entry?
* I'd like that each daily note show the summary of the note from a week before, a month, and finally from the same date, but years before
b. How will a user add a "one-line" note? It is automated by the app
c. How should these one-line notes be displayed or reviewed? Only displayed! As they are old notes, they cannot be used for the reviews. We must put them in a section/chapter specific so that the application can ignore them when reviewing

## 6. User Stories
Could you provide a few user stories? (e.g., As a [type of user], I want to [perform an action] so that [benefit].)
a. User Story for Daily Entry:

As a user I want to quickly jot a note for today without worry to create the appropriate format
As a user I want to quickly jot a note with the hour and minute I am writing the note

b. User Story for Weekly Review:
As a user I want to start a weekly review in a new file with the appropriate template, having a summary of each day of the week
c. User Story for One-line Note:
As a user I like to open a daily note file and read summary from notes of older days

## 7. Acceptance Criteria
How will we know when this application is successfully implemented? What are the key success criteria?
a. For Daily Entry:
The daily entry has the template defined in the configuration file
The daily entry has the one-line summary from the past as defined
b. For Reviews:
The reviews have the summary from the days of the week under review
c. For One-line Notes:
The one line notes show summary from the defined days

## 8. Non-Goals (Out of Scope)
Are there any specific features or functionalities that LogBook *should not* include in its initial version? No

## 9. Technical Considerations
a. Are there any specific Go libraries or approaches you envision for file handling, templating, or configuration? Using default golang library as much as possible, or well established and largely used libraries
b. How should data be stored? (e.g., plain text files, a simple database like SQLite, specific file per entry)
plain text files. The same used for the journalling

## 10. Desired Look and Feel (CLI)
a. Are there any preferences for the command-line interface (e.g., color usage, interactive prompts, simple text output)? Colorful will be nice.
b. How should errors or confirmations be displayed to the user? If necessary some TUI can be implemented with github.com/rivo/tview


# Sun Oct  5 10:33:54 AM CEST 2025

There are some problems:
1. AI is used too much. Every time I add a new journal log, AI agent is called. Let's call the AI agent for summary ONLY when doing a review and only if the summary is missing from the daily note
2. The new logs are always added at the bottom of the journal note, regardless the format of the note. For example, the default template has "One-line note" as last paragraph, hence the new entries are added in this section, instead than in LOG section. 

To solve the issues above, we will remove the ability to decide the Journal template. We will only use a predefined and hardcoded one which is the following

```markdown
# <today's date>
<!-- add today summary below this line. If missing, the AI will generate one for you according to configuration file -->

# One-line note
* 1 week ago (dd-mm-yyyy): summary from 1 week ago, or "missing"
* 1 month ago (dd-mm-yyyy): summary from 1 month ago, or "missing"
* 6 month ago (dd-mm-yyyy): summary from 6 month ago, or "missing"
* 1 year ago (dd-mm-yyyy): summary from 1 year ago, or "missing"

# LOG

```

The summary is added ONLY IF it is not already present.
