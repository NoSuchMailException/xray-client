// Package config provides configuration loading from a YAML file.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full application configuration.
type Config struct {
	Inbound  InboundConfig  `yaml:"inbound"`
	Outbound OutboundConfig `yaml:"outbound"`
}

type InboundConfig struct {
	Listen string `yaml:"listen"`
	Port   uint16 `yaml:"port"`
}

type OutboundConfig struct {
	Address     string `yaml:"address"`
	Port        uint16 `yaml:"port"`
	UUID        string `yaml:"uuid"`
	PublicKey   string `yaml:"public_key"`
	ShortID     string `yaml:"short_id"`
	ServerName  string `yaml:"server_name"`
	Fingerprint string `yaml:"fingerprint"`
}

// Load reads and parses a YAML configuration file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	return &cfg, nil
}
