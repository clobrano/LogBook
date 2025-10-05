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
	JournalDir       string `toml:"journal_dir"`
	DailyFileName    string `toml:"daily_file_name"`
	DailyTemplate    string `toml:"daily_template"`
	LogEntryTemplate string `toml:"log_entry_template"`
	AIEnabled        bool   `toml:"ai_enabled"`
	AICommand        string `toml:"ai_command"`
	AIPrompt         string `toml:"ai_prompt"`
	OneLineTemplate  string `toml:"one_line_template"`
	AISummarizer     ai.AISummarizer `toml:"-"` // Not serialized to TOML
}

// DefaultConfig returns a new Config with default values.
func DefaultConfig() *Config {
	return &Config{
		JournalDir:       filepath.Join(os.Getenv("HOME"), ".logbook", "journal"),
		DailyFileName:    "{{.Date | formatDate \"2006-01-02\"}}.md",
		DailyTemplate:    "# {{.Date | formatDate \"Jan 02 2006 Monday\"}}\n<!-- add today summary below this line. If missing, the AI will generate one for you according to configuration file -->\n\n# One-line note\n\n# LOG\n\n",
		LogEntryTemplate: "{{.Time | formatTime \"15:04\"}} {{.Entry}}",
		AIEnabled:        false,
		AICommand:        "", // Example: "gemini --prompt '{PROMPT} {TEXT}'" or "claude --text '{TEXT}' --instructions '{PROMPT}'"
		AIPrompt:         "Write a summary of the note at the given file. Use 1st person and a simple language. Use 200 characters or less",
		OneLineTemplate:  "{{.Date | formatDate \"2006-01-02\"}}: {{.Summary}}",
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
		cfg.AISummarizer = ai.NewAISummarizer(cfg.AICommand)
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
	if cfg.LogEntryTemplate == "" {
		return fmt.Errorf("LogEntryTemplate cannot be empty")
	}
	if cfg.AIEnabled && cfg.AIPrompt == "" {
		return fmt.Errorf("AIPrompt cannot be empty if AI is enabled")
	}
	if cfg.AIEnabled && cfg.AICommand == "" {
		return fmt.Errorf("AICommand cannot be empty if AI is enabled")
	}
	return nil
}
