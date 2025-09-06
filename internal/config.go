package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	Title string `yaml:"title"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
}

// Default configuration values
func DefaultConfig() *Config {
	config := &Config{}
	config.Server.Host = "0.0.0.0"
	config.Server.Port = "8080"
	config.Server.Title = "WireGuard Gateway Portal"
	return config
}

// LoadConfig loads configuration from file, falls back to defaults if file doesn't exist
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}
