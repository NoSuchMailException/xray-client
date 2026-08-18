// Package config provides configuration loading, parsing, and saving from a YAML file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full application configuration, including both inbound
// listening settings and outbound proxy connection parameters.
type Config struct {
	Inbound  InboundConfig  `yaml:"inbound"`
	Outbound OutboundConfig `yaml:"outbound"`
}

// InboundConfig defines the local listening address and port for the proxy.
type InboundConfig struct {
	Listen string `yaml:"listen"`
	Port   uint16 `yaml:"port"`
}

// OutboundConfig stores the remote VLESS/REALITY server connection details.
type OutboundConfig struct {
	Address         string `yaml:"address"`
	Port            uint16 `yaml:"port"`
	UUID            string `yaml:"uuid"`
	PublicKey       string `yaml:"public_key"`
	ShortID         string `yaml:"short_id"`
	ServerName      string `yaml:"server_name"`
	GRPCServiceName string `yaml:"grpc_service_name"`
	Fingerprint     string `yaml:"fingerprint"`
	Remark          string `yaml:"remark"`
}

var (
	defaultListen = "127.0.0.1"
	defaultPort   = uint16(1080)
)

// InitConfig checks if the configuration file exists, creates a default one if it doesn't,
// and ensures all required fields (like valid ports and VLESS links) are properly set.
func InitConfig(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Config not found, creating a default one...")
			cfg = &Config{}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	inboundChanged := ensureInbound(&cfg.Inbound)
	outboundChanged := ensureOutbound(&cfg.Outbound)

	if inboundChanged || outboundChanged {
		if err := Save(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to update config: %w", err)
		}
		fmt.Println("Config saved to disk.")
	}

	return cfg, nil
}

// ensureInbound validates the inbound configuration, requesting user input
// if the default port is already in use. Returns true if the struct was modified.
func ensureInbound(in *InboundConfig) bool {
	changed := false
	if in.Listen == "" {
		in.Listen = defaultListen
		changed = true
	}

	if in.Port == 0 {
		in.Port = defaultPort
		changed = true
	}

	actualPort := CheckOrSetPort(in.Listen, in.Port)
	if actualPort != in.Port {
		in.Port = actualPort
		changed = true
	}

	return changed
}

// ensureOutbound validates the outbound configuration, requesting a VLESS link
// from the user if the UUID is missing. Returns true if the struct was modified.
func ensureOutbound(out *OutboundConfig) bool {
	if out.UUID != "" {
		return false
	}

	return CheckOrSetVLESS(out)
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

// Save marshals the configuration state and writes it to the specified file path
// with restricted 0600 permissions.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	// 0600 is perfect here (read/write for owner only) since it contains sensitive keys
	return os.WriteFile(path, data, 0600)
}
