package configuration

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	ScorerTypeML    = "ml"
	ScorerTypeRules = "rules"
)

// AppConfig represents the complete application configuration.
type AppConfig struct {
	// Logger — logger component configuration
	Logger LoggerConfig `mapstructure:"logger"`
	// Server — HTTP server configuration
	Server ServerConfig `mapstructure:"server"`
	// Analysis — behavioral analysis module configuration
	Analysis AnalysisConfig `mapstructure:"analysis"`
	// Dataset — behavioral dataset configuration
	Dataset DatasetConfig `mapstructure:"dataset"`
}

// LoggerConfig defines logging settings.
type LoggerConfig struct {
	// Level — log level: debug, info, warn, warning, error.
	// Value is case-insensitive but checked in lowercase.
	Level string `mapstructure:"level"`
}

// ServerConfig contains HTTP server parameters.
type ServerConfig struct {
	// Address — address and port where the server will listen (e.g., ":8080").
	Address string `mapstructure:"address"`
	// Static — path to directory with static files served by the server.
	// Can be empty if static serving is not required.
	Static string `mapstructure:"static"`
}

type ScorerConfig struct {
	// Type — scorer type
	Type string `mapstructure:"type"`
	// Model — path to the model file
	Model string `mapstructure:"model"`
	// URL — URL to the scorer service
	Url string `mapstructure:"model"`
	// Rules — path to the file with analysis rules in YAML format.
	Rules string `mapstructure:"rules"`
}

// AnalysisConfig defines behavioral analysis parameters.
type AnalysisConfig struct {
	// Token — token for authenticating requests to the analysis system.
	// Must be set, otherwise the configuration will be invalid.
	Token string `mapstructure:"token"`
	// Scorers — list of scorers
	Scorers []ScorerConfig `mapstructure:"scorers"`
	// TracesLength — maximum number of stored traces per identifier.
	// Used in TracesRepository to limit buffer size.
	TracesLength int `mapstructure:"traces_length"`
	// TracesTtl — lifetime of traces (time.Duration), after which inactive records are deleted.
	// Example: "5m", "1h", "24h".
	TracesTtl time.Duration `mapstructure:"traces_ttl"`
}

// DatasetConfig defines behavioral dataset parameters
type DatasetConfig struct {
	// Dataset file path (optional)
	File string `mapstructure:"file"`
	// Maximal dataset file size (default 100M)
	Size int `mapstructure:"size"`
	// Number of dataset files (default 20)
	Amount int `mapstructure:"amount"`
}

// Validate checks the correctness of the entire application configuration.
// Calls validation for each nested structure and returns the first detected error.
// Returns nil if the configuration is valid.
func (c *AppConfig) Validate() error {
	if err := c.Logger.Validate(); err != nil {
		return err
	}

	if err := c.Server.Validate(); err != nil {
		return err
	}

	if err := c.Analysis.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate checks the correctness of the logger configuration.
// Verifies that the log level is set and is one of the supported values.
// Supported values: debug, info, warn, warning, error (case-insensitive).
func (l *LoggerConfig) Validate() error {
	if l.Level == "" {
		return errors.New("logger.level: must be specified")
	}

	valid := map[string]bool{"debug": true, "info": true, "warn": true, "warning": true, "error": true}
	if !valid[strings.ToLower(l.Level)] {
		return fmt.Errorf("logger.level: unsupported level '%s'", l.Level)
	}

	return nil
}

// Validate dataset parameters
func (d *DatasetConfig) Validate() error {
	if d.Amount == 0 {
		d.Amount = 20
	}

	if d.Size == 0 {
		d.Size = 100
	}

	return nil
}

// Validate checks the correctness of the scorer configuration.
func (c *ScorerConfig) Validate() error {
	switch c.Type {
	case ScorerTypeML:
		if len(c.Model) == 0 {
			return errors.New("ML scorer: model name must be specified")
		}
		if _, err := url.Parse(c.Url); err != nil {
			return errors.New("ML scorer: URL is incorrect")
		}
	case ScorerTypeRules:
		if len(c.Rules) == 0 {
			return errors.New("scorer rules: path must be specified")
		}
	default:
		return errors.New("Scorer type must be specified")
	}

	return nil
}

// Validate checks the correctness of the server configuration.
// Verifies that the server address is set.
func (n *ServerConfig) Validate() error {
	if n.Address == "" {
		return errors.New("server.address: must be specified")
	}

	return nil
}

// Validate checks the correctness of the analysis configuration.
// Verifies that required fields are set: Token and Rules.
func (a *AnalysisConfig) Validate() error {
	if len(a.Scorers) == 0 {
		return errors.New("analysis.scorers: must be specified")
	}

	for i := range a.Scorers {
		if err := a.Scorers[i].Validate(); err != nil {
			return err
		}
	}

	if a.Token == "" {
		return errors.New("analysis.token: must be specified")
	}

	return nil
}

// LoadConfig loads configuration from the specified file using Viper.
// Supports YAML format. Also includes environment variable loading (AutomaticEnv),
// which can override values from the file.
//
// Parameter configPath — path to the configuration file.
//
// Returns a pointer to AppConfig or an error if:
// - the file is not found or inaccessible
// - the configuration has invalid format
// - one of the sections fails validation
func LoadConfig(configPath string) (*AppConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}
