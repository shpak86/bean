package configuration

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// AppConfig представляет полную конфигурацию приложения.
// Содержит настройки логгера, сервера и анализа поведенческих данных.
type AppConfig struct {
	// Logger — конфигурация компонента логгирования
	Logger LoggerConfig `mapstructure:"logger"`
	// Server — конфигурация HTTP-сервера
	Server ServerConfig `mapstructure:"server"`
	// Analysis — конфигурация модуля анализа поведения
	Analysis AnalysisConfig `mapstructure:"analysis"`
}

// LoggerConfig определяет настройки логгирования.
type LoggerConfig struct {
	// Level — уровень логирования: debug, info, warn, warning, error.
	// Значение чувствительно к регистру, но проверяется в нижнем регистре.
	Level string `mapstructure:"level"`
}

// ServerConfig содержит параметры HTTP-сервера.
type ServerConfig struct {
	// Address — адрес и порт, на котором будет запущен сервер (например, ":8080").
	Address string `mapstructure:"address"`
	// Static — путь к директории со статическими файлами, которые будут раздаваться сервером.
	// Может быть пустым, если статика не требуется.
	Static string `mapstructure:"static"`
}

// AnalysisConfig определяет параметры поведенческого анализа.
type AnalysisConfig struct {
	// Token — токен для аутентификации запросов к системе анализа.
	// Должен быть задан, иначе конфигурация будет недействительной.
	Token string `mapstructure:"token"`
	// Rules — путь к файлу с правилами анализа в формате YAML.
	// Должен быть указан, чтобы система могла загрузить логику оценки.
	Rules string `mapstructure:"rules"`
	// TracesLength — максимальное количество хранимых трейсов на один идентификатор.
	// Используется в TracesRepository для ограничения размера буфера.
	TracesLength int `mapstructure:"traces_length"`
	// TracesTtl — время жизни трейсов (time.Duration), после которого неактивные записи удаляются.
	// Например: "5m", "1h", "24h".
	TracesTtl time.Duration `mapstructure:"traces_ttl"`
}

// Validate проверяет корректность всей конфигурации приложения.
// Вызывает валидацию каждой вложенной структуры и возвращает первую обнаруженную ошибку.
// Возвращает nil, если конфигурация валидна.
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

// Validate проверяет корректность конфигурации логгера.
// Проверяет, что уровень логирования задан и является одним из поддерживаемых значений.
// Поддерживаемые значения: debug, info, warn, warning, error (без учёта регистра).
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

// Validate проверяет корректность конфигурации сервера.
// Проверяет, что адрес сервера задан.
func (n *ServerConfig) Validate() error {
	if n.Address == "" {
		return errors.New("server.address: must be specified")
	}
	return nil
}

// Validate проверяет корректность конфигурации анализа.
// Проверяет, что указаны обязательные поля: Token и Rules.
func (a *AnalysisConfig) Validate() error {
	if a.Rules == "" {
		return errors.New("analysis.rules: must be specified")
	}
	if a.Token == "" {
		return errors.New("analysis.token: must be specified")
	}
	return nil
}

// LoadConfig загружает конфигурацию из указанного файла с использованием Viper.
// Поддерживает YAML-формат. Также включает загрузку переменных окружения (AutomaticEnv),
// которые могут переопределять значения из файла.
//
// Параметр configPath — путь к файлу конфигурации.
//
// Возвращает указатель на AppConfig или ошибку, если:
//   - файл не найден или недоступен
//   - конфигурация имеет неверный формат
//   - одна из секций не проходит валидацию
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
