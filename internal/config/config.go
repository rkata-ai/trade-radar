package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AI       AIConfig       `mapstructure:"ai"`
	Database DatabaseConfig `mapstructure:"database"`
}

type AIConfig struct {
	OllamaBaseURL string `mapstructure:"ollama_base_url"`
	OllamaModel   string `mapstructure:"ollama_model"`
	Debug         bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	User             string `mapstructure:"user"`
	Password         string `mapstructure:"password"`
	DBName           string `mapstructure:"dbname"`
	SSLMode          string `mapstructure:"sslmode"`
	ConnectionString string `mapstructure:"connection_string"`
}

func Load(configPath string) (*Config, error) {
	// Устанавливаем значения по умолчанию
	viper.SetDefault("ai.ollama_base_url", "http://localhost:11434")
	viper.SetDefault("ai.ollama_model", "llama2")
	viper.SetDefault("ai.debug", false)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "tg_reader") // Используем tg_reader как базу по умолчанию
	viper.SetDefault("database.sslmode", "disable")

	// Читаем конфигурацию из указанного файла
	viper.SetConfigFile(configPath)

	// Читаем переменные окружения
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TRADING")

	// Привязываем переменные окружения к конфигурации
	bindEnvs()

	// Пытаемся прочитать файл конфигурации
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Валидация обязательных полей
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func bindEnvs() {
	viper.BindEnv("ai.ollama_base_url", "TRADING_AI_OLLAMA_BASE_URL")
	viper.BindEnv("ai.ollama_model", "TRADING_AI_OLLAMA_MODEL")
	viper.BindEnv("ai.debug", "TRADING_AI_DEBUG")

	viper.BindEnv("database.host", "TRADING_DATABASE_HOST")
	viper.BindEnv("database.port", "TRADING_DATABASE_PORT")
	viper.BindEnv("database.user", "TRADING_DATABASE_USER")
	viper.BindEnv("database.password", "TRADING_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "TRADING_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "TRADING_DATABASE_SSLMODE")
}

func validateConfig(config *Config) error {
	// Проверяем конфигурацию базы данных
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Port == 0 {
		return fmt.Errorf("database port is required")
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}
