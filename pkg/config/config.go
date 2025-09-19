package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/clobrano/LogBook/pkg/ai"
)

// Config represents the application's configuration.
type Config struct {
	JournalDir      string `toml:"journal_dir"`
	DailyFileName   string `toml:"daily_file_name"`
	DailyTemplate   string `toml:"daily_template"`
	AIEnabled       bool   `toml:"ai_enabled"`
	AIPrompt        string `toml:"ai_prompt"`
	OneLineTemplate string `toml:"one_line_template"`
	AISummarizer    ai.AISummarizer `toml:"-"` // Not serialized to TOML
}

// DefaultConfig returns a new Config with default values.
func DefaultConfig() *Config {
	return &Config{
		JournalDir:      filepath.Join(os.Getenv("HOME"), ".logbook", "journal"),
		DailyFileName:   "{{.Date | formatDate \"2006-01-02\"}}.md",
		DailyTemplate:   "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n\n[SUMMARY_PLACEHOLDER]\n\n## LOG\n",
		AIEnabled:       false,
		AIPrompt:        "Write a summary of the note at the given file. Use 1st person and a simple language. Use 200 characters or less",
		OneLineTemplate: "{{.Date | formatDate \"2006-01-02\"}}: {{.Summary}}",
	}
}

// LoadConfig loads configuration from a TOML file.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveConfig saves configuration to a TOML file.
func SaveConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

// Validate checks if the configuration is valid.
func (cfg *Config) Validate() error {
	if cfg.JournalDir == "" {
		return fmt.Errorf("JournalDir cannot be empty")
	}
	if cfg.DailyFileName == "" {
		return fmt.Errorf("DailyFileName cannot be empty")
	}
	if cfg.DailyTemplate == "" {
		return fmt.Errorf("DailyTemplate cannot be empty")
	}
	if cfg.AIEnabled && cfg.AIPrompt == "" {
		return fmt.Errorf("AIPrompt cannot be empty if AI is enabled")
	}
	return nil
}
