// Package config handles loading and parsing of MasterDnsVPN configuration files.
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// ServerConfig holds the DNS/VPN server configuration.
type ServerConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Protocol string `toml:"protocol"`
}

// TunnelConfig holds tunnel-specific settings.
type TunnelConfig struct {
	MTU        int    `toml:"mtu"`
	Interface  string `toml:"interface"`
	Subnet     string `toml:"subnet"`
	DNSServer  string `toml:"dns_server"`
}

// AuthConfig holds authentication credentials.
type AuthConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Token    string `toml:"token"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `toml:"level"`
	File   string `toml:"file"`
	Format string `toml:"format"`
}

// ClientConfig is the top-level configuration structure for the client.
type ClientConfig struct {
	Server ServerConfig `toml:"server"`
	Tunnel TunnelConfig `toml:"tunnel"`
	Auth   AuthConfig   `toml:"auth"`
	Log    LogConfig    `toml:"log"`
}

// DefaultClientConfig returns a ClientConfig populated with sensible defaults.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Server: ServerConfig{
			Host:     "127.0.0.1",
			Port:     5300,
			Protocol: "udp",
		},
		Tunnel: TunnelConfig{
			MTU:       1500,
			Interface: "tun0",
			Subnet:    "10.0.0.0/24",
			DNSServer: "8.8.8.8",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// LoadClientConfig reads and parses a TOML configuration file from the given path.
// If the file does not exist, it returns the default configuration.
func LoadClientConfig(path string) (*ClientConfig, error) {
	cfg := DefaultClientConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, fmt.Errorf("config file not found at %s, using defaults", path)
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that the configuration values are within acceptable ranges.
func (c *ClientConfig) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host must not be empty")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port %d is out of valid range (1-65535)", c.Server.Port)
	}
	if c.Server.Protocol != "udp" && c.Server.Protocol != "tcp" {
		return fmt.Errorf("unsupported protocol %q, must be 'udp' or 'tcp'", c.Server.Protocol)
	}
	if c.Tunnel.MTU < 576 || c.Tunnel.MTU > 9000 {
		return fmt.Errorf("tunnel MTU %d is out of valid range (576-9000)", c.Tunnel.MTU)
	}
	return nil
}

// ServerAddr returns the formatted server address as host:port.
func (c *ClientConfig) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
