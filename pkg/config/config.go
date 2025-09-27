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
	AIBinary        string `toml:"ai_binary"`
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
		AIBinary:        "", // Default to empty string, meaning no specific AI binary is configured
		AIPrompt:        "Write a summary of the note at the given file. Use 1st person and a simple language. Use 200 characters or less",
		OneLineTemplate: "{{.Date | formatDate \"2006-01-02\"}}: {{.Summary}}",
	}
}

// LoadConfig loads configuration from a TOML file.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file %s: %w", path, err)
	}

	if cfg.AIEnabled {
		cfg.AISummarizer = ai.NewAISummarizer(cfg.AIBinary)
	}

	return cfg, nil
}

// SaveConfig saves configuration to a TOML file.
func SaveConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file %s: %w", path, err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config to file %s: %w", path, err)
	}
	return nil
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
	if cfg.AIEnabled && cfg.AIBinary == "" {
		return fmt.Errorf("AIBinary cannot be empty if AI is enabled")
	}
	return nil
}
