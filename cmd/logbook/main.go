package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/journal"
	"github.com/clobrano/LogBook/pkg/review"
)

func main() {
	cfg := config.DefaultConfig() // For now, use default config

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Println("This is the help command for LogBook.")
			// TODO: Implement more comprehensive help
		case "log":
			if len(os.Args) < 3 {
				fmt.Println("Usage: logbook log <entry>")
				os.Exit(1)
			}
			entry := strings.Join(os.Args[2:], " ")
			// For now, assume a daily file exists for today. In future, create if not exists.
			// TODO: Implement logic to get/create today's journal file path
			// For testing, let's use a dummy file path for now.
			// This will be replaced with actual file creation logic later.
			journalFilePath := "/tmp/daily_journal.md" // Placeholder

			// Create a dummy file for testing purposes if it doesn't exist
			if _, err := os.Stat(journalFilePath); os.IsNotExist(err) {
				// Create a basic file with a LOG chapter for testing
				err := os.WriteFile(journalFilePath, []byte("Summary\n\n## LOG\n"), 0644)
				if err != nil {
					fmt.Printf("Error creating dummy journal file: %v\n", err)
					os.Exit(1)
				}
			}

			err := journal.AppendToLog(journalFilePath, entry, time.Now())
			if err != nil {
				fmt.Printf("Error appending to log: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Entry added to log.")
		case "review":
			if len(os.Args) < 3 {
				fmt.Println("Usage: logbook review <week|month|year> [args]")
				os.Exit(1)
			}
			subCommand := os.Args[2]
			switch subCommand {
			case "week":
				if len(os.Args) < 5 {
					fmt.Println("Usage: logbook review week <week number> <year>")
					os.Exit(1)
				}
				week, err := strconv.Atoi(os.Args[3])
				if err != nil {
					fmt.Println("Invalid week number:", os.Args[3])
					os.Exit(1)
				}
				year, err := strconv.Atoi(os.Args[4])
				if err != nil {
					fmt.Println("Invalid year:", os.Args[4])
					os.Exit(1)
				}
				result, err := review.ReviewWeek(cfg, week, year, cfg.AISummarizer, os.Stdin)
				if err != nil {
					fmt.Printf("Error generating weekly review: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(result)
			case "month":
				if len(os.Args) < 5 {
					fmt.Println("Usage: logbook review month <month name> <year>")
					os.Exit(1)
				}
				month := os.Args[3]
				year, err := strconv.Atoi(os.Args[4])
				if err != nil {
					fmt.Println("Invalid year:", os.Args[4])
					os.Exit(1)
				}
				result, err := review.ReviewMonth(cfg, month, year, cfg.AISummarizer, os.Stdin)
				if err != nil {
					fmt.Printf("Error generating monthly review: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(result)
			case "year":
				if len(os.Args) < 4 {
					fmt.Println("Usage: logbook review year <year>")
					os.Exit(1)
				}
				year, err := strconv.Atoi(os.Args[3])
				if err != nil {
					fmt.Println("Invalid year:", os.Args[3])
					os.Exit(1)
				}
				result, err := review.ReviewYear(cfg, year, cfg.AISummarizer, os.Stdin)
				if err != nil {
					fmt.Printf("Error generating yearly review: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(result)
			default:
				fmt.Println("Unknown review subcommand. Use 'logbook review help' for more information.")
				os.Exit(1)
			}
		default:
			fmt.Println("Unknown command. Use 'logbook help' for more information.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Welcome to LogBook! Use 'logbook help' for more information.")
	}
}
