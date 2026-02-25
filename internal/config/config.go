package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Targets  []TargetConfig `yaml:"targets"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	// Other fields can be added here later (Server, Redis, etc.)
}

type AuthConfig struct {
	PassportURL string `yaml:"passport_url"`
}

type TargetConfig struct {
	Path  string `yaml:"path"`
	Table string `yaml:"table"`
}

type DatabaseConfig struct {
	ClearOnStartup bool `yaml:"clear_on_startup"`
}

// LoadConfig reads and parses the YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
