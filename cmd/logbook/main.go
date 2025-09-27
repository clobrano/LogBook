package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/clobrano/LogBook/pkg/config"
	"github.com/clobrano/LogBook/pkg/journal"
	"github.com/clobrano/LogBook/pkg/review"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(usr.HomeDir, ".config", "logbook")
	configFilePath := filepath.Join(configDir, "config.toml")

	var cfg *config.Config // Declare cfg here, initialize later if needed

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Println(`LogBook is a command-line application for daily journaling and periodic reviews.

Usage:

  logbook <command> [arguments]

Available Commands:
  config  Create a default configuration file.
  help    Display help information for LogBook.
  log     Add an entry to today's journal.
          Usage: logbook log <your entry text>
  review  Perform a review of journal entries for a specific period.
          Usage:
            logbook review week <week number> <year>
            logbook review month <month name> <year>
            logbook review year <year>

Examples:
  logbook config
  logbook log "Started working on the LogBook help command."
  logbook review week 38 2025
  logbook review month September 2025
  logbook review year 2025`)
		case "config":
			usr, err := user.Current()
			if err != nil {
				fmt.Printf("Error getting current user: %v\n", err)
				os.Exit(1)
			}

			configDir := filepath.Join(usr.HomeDir, ".config", "logbook")
			configFilePath := filepath.Join(configDir, "config.toml")

			_, err = os.Stat(configFilePath)
			if err == nil {
				fmt.Printf("Configuration file already exists at: %s\n", configFilePath)
				os.Exit(0)
			} else if !os.IsNotExist(err) {
				fmt.Printf("Error checking config file: %v\n", err)
				os.Exit(1)
			}

			// If we reach here, the file does not exist, so create it.

			// Create config directory if it doesn't exist
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Printf("Error creating config directory %s: %v\n", configDir, err)
				os.Exit(1)
			}

			defaultCfg := config.DefaultConfig()
			err = config.SaveConfig(configFilePath, defaultCfg)
			if err != nil {
				fmt.Printf("Error saving default config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Default configuration file created at: %s\n", configFilePath)
			os.Exit(0)
		case "log":
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				fmt.Printf("Error loading configuration: %v\n", err)
				os.Exit(1)
			}
			if len(os.Args) < 3 {
				fmt.Println("Usage: logbook log <entry>")
				os.Exit(1)
			}
			entry := strings.Join(os.Args[2:], " ")

			journalFilePath, message, err := journal.CreateDailyJournalFile(cfg, time.Now())
			if err != nil {
				fmt.Printf("Error creating/getting daily journal file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(message)

			err = journal.AppendToLog(cfg, journalFilePath, entry, time.Now())
			if err != nil {
				fmt.Printf("Error appending to log: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Entry added to log.")
		case "review":
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				fmt.Printf("Error loading configuration: %v\n", err)
				os.Exit(1)
			}
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