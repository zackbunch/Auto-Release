package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for SYAC.
type Config struct {
	ProtectedBranches []string `yaml:"protected_branches"`
}

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
