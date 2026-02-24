package configPkg

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Targets  []TargetConfig `yaml:"targets"`
	Database DatabaseConfig `yaml:"database"`
	// Other fields can be added here later (Server, Redis, etc.)
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
