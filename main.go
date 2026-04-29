package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

// Version information
const (
	AppName    = "MasterDnsVPN"
	AppVersion = "1.0.0"
)

// Config holds the top-level configuration loaded from client_config.toml
type Config struct {
	General  GeneralConfig  `toml:"general"`
	DNS      DNSConfig      `toml:"dns"`
	Tunnel   TunnelConfig   `toml:"tunnel"`
	Logging  LoggingConfig  `toml:"logging"`
}

// GeneralConfig contains general application settings
type GeneralConfig struct {
	Mode       string `toml:"mode"`       // "client" or "server"
	ServerAddr string `toml:"server_addr"`
	ServerPort int    `toml:"server_port"`
	Secret     string `toml:"secret"`
}

// DNSConfig contains DNS-related settings
type DNSConfig struct {
	ListenAddr  string   `toml:"listen_addr"`
	ListenPort  int      `toml:"listen_port"`
	Upstream    []string `toml:"upstream"`
	FakeDomain  string   `toml:"fake_domain"`
	FakeIP      string   `toml:"fake_ip"`
}

// TunnelConfig contains VPN tunnel settings
type TunnelConfig struct {
	Interface  string `toml:"interface"`
	LocalIP    string `toml:"local_ip"`
	RemoteIP   string `toml:"remote_ip"`
	MTU        int    `toml:"mtu"`
	Protocol   string `toml:"protocol"` // "udp" or "tcp"
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `toml:"level"`
	File   string `toml:"file"`
}

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "client_config.toml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// Load configuration
	var cfg Config
	if _, err := toml.DecodeFile(*configFile, &cfg); err != nil {
		log.Fatalf("[ERROR] Failed to load config file '%s': %v", *configFile, err)
	}

	log.Printf("[INFO] Starting %s v%s in %s mode", AppName, AppVersion, cfg.General.Mode)

	// Initialize logging
	if err := initLogging(cfg.Logging); err != nil {
		log.Fatalf("[ERROR] Failed to initialize logging: %v", err)
	}

	// Start the appropriate mode
	switch cfg.General.Mode {
	case "client":
		log.Printf("[INFO] Connecting to server %s:%d", cfg.General.ServerAddr, cfg.General.ServerPort)
	case "server":
		log.Printf("[INFO] Starting server on port %d", cfg.General.ServerPort)
	default:
		log.Fatalf("[ERROR] Unknown mode '%s'. Must be 'client' or 'server'", cfg.General.Mode)
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("[INFO] Shutting down...")
}

// initLogging configures the application logger based on LoggingConfig
func initLogging(cfg LoggingConfig) error {
	if cfg.File != "" {
		f, err := os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("could not open log file: %w", err)
		}
		log.SetOutput(f)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return nil
}
