package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Targets  []TargetConfig `yaml:"targets"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type AuthConfig struct {
	PassportURL   string `yaml:"passport_url"`
	PermissionURL string `yaml:"permission_url"`
	SysCode       string `yaml:"sys_code"`
}

type TargetConfig struct {
	Path  string `yaml:"path"`
	Table string `yaml:"table"`
}

type DatabaseConfig struct {
	ClearOnStartup bool `yaml:"clear_on_startup"`
}

// LoadConfig reads and parses the YAML configuration file.
// Values in the form ${VAR_NAME} are expanded from environment variables,
// which allows CI/CD pipelines (e.g. GitHub Actions) to inject secrets at runtime.
func LoadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand ${VAR} / $VAR placeholders using the current process environment.
	expanded := os.ExpandEnv(string(raw))

	var cfg Config
	if err = yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
