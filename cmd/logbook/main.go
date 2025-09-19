package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/journal"
)

func main() {
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
		default:
			fmt.Println("Unknown command. Use 'logbook help' for more information.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Welcome to LogBook! Use 'logbook help' for more information.")
	}
}

