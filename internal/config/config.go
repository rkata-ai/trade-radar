package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AI AIConfig `yaml:"ai"`
}

type AIConfig struct {
	OllamaBaseURL string `yaml:"ollama_base_url"`
	OllamaModel   string `yaml:"ollama_model"`
}

// Load загружает конфигурацию из указанного файла.
func Load(configPath string) (*Config, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
