package configuration

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Logger   LoggerConfig   `mapstructure:"logger"`
	Server   ServerConfig   `mapstructure:"server"`
	Analysis AnalysisConfig `mapstructure:"analysis"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
	Static  string `mapstructure:"static"`
}

type AnalysisConfig struct {
	Token        string        `mapstructure:"token"`
	Rules        string        `mapstructure:"rules"`
	TracesLength int           `mapstructure:"traces_length"`
	TracesTtl    time.Duration `mapstructure:"traces_ttl"`
}

// Validate валидирует всю конфигурацию
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

func (n *ServerConfig) Validate() error {
	if n.Address == "" {
		return errors.New("server.address: must be specified")
	}
	return nil
}

func (a *AnalysisConfig) Validate() error {
	if a.Rules == "" {
		return errors.New("analysis.rules: must be specified")
	}
	if a.Token == "" {
		return errors.New("analysis.token: must be specified")
	}
	return nil
}

// LoadConfig загружает конфиг с помощью Viper и возвращает AppConfig
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
