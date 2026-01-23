package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Master MasterConfig `yaml:"master" envPrefix:"MASTER_"`
	Worker WorkerConfig `yaml:"worker" envPrefix:"WORKER_"`
	Claude ClaudeConfig `yaml:"claude" envPrefix:"CLAUDE_"`
}

// MasterConfig contains Master server configuration
type MasterConfig struct {
	Host      string `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port      int    `yaml:"port" env:"PORT" envDefault:"8848"`
	DBPath    string `yaml:"db_path" env:"DB_PATH" envDefault:".claude/cpds-manager/agentflow.db"`
	AutoStart bool   `yaml:"auto_start" env:"AUTO_START" envDefault:"false"`
}

// WorkerConfig contains Worker configuration
type WorkerConfig struct {
	ID        string `yaml:"id" env:"ID"`
	MasterURL string `yaml:"master_url" env:"MASTER_URL" envDefault:"http://localhost:8848"`
	DBPath    string `yaml:"db_path" env:"DB_PATH" envDefault:".claude/cpds-manager/agentflow.db"`
	GroupName string `yaml:"group_name" env:"GROUP_NAME" envDefault:"default"`
	Mode      string `yaml:"mode" env:"MODE" envDefault:"auto"` // auto, oneshot, manual
}

// ClaudeConfig contains Claude executor configuration
type ClaudeConfig struct {
	ServerURL   string  `yaml:"server_url" env:"SERVER_URL" envDefault:"http://localhost:8849"`
	Model       string  `yaml:"model" env:"MODEL"`
	MaxTokens   int     `yaml:"max_tokens" env:"MAX_TOKENS" envDefault:"4096"`
	Temperature float64 `yaml:"temperature" env:"TEMPERATURE" envDefault:"0.7"`
	Timeout     int     `yaml:"timeout" env:"TIMEOUT" envDefault:"120"` // seconds
}

// Load loads configuration from file and environment variables
// Priority: Environment variables > Config file > Defaults
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Load from file if provided
	if configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Override with environment variables
	if err := loadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from env: %w", err)
	}

	return cfg, nil
}

// LoadFromFile loads configuration only from file (no env override)
func LoadFromFile(configPath string) (*Config, error) {
	cfg := &Config{}
	if err := loadFromFile(cfg, configPath); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFromEnv loads configuration only from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	if err := loadFromEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// loadFromFile reads configuration from YAML file
func loadFromFile(cfg *Config, configPath string) error {
	// Expand ~ in path
	if len(configPath) > 0 && configPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, configPath[1:])
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// loadFromEnv reads configuration from environment variables
func loadFromEnv(cfg *Config) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Master: MasterConfig{
			Host:      "0.0.0.0",
			Port:      8848,
			DBPath:    ".claude/cpds-manager/agentflow.db",
			AutoStart: false,
		},
		Worker: WorkerConfig{
			MasterURL: "http://localhost:8848",
			DBPath:    ".claude/cpds-manager/agentflow.db",
			GroupName: "default",
			Mode:      "auto",
		},
		Claude: ClaudeConfig{
			ServerURL:   "http://localhost:8849",
			MaxTokens:   4096,
			Temperature: 0.7,
			Timeout:     120,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Master config
	if c.Master.Port <= 0 || c.Master.Port > 65535 {
		return fmt.Errorf("invalid master port: %d", c.Master.Port)
	}

	// Validate Worker config
	if c.Worker.MasterURL == "" {
		return fmt.Errorf("worker master_url cannot be empty")
	}

	// Validate Claude config
	if c.Claude.MaxTokens <= 0 {
		return fmt.Errorf("invalid claude max_tokens: %d", c.Claude.MaxTokens)
	}

	if c.Claude.Temperature < 0 || c.Claude.Temperature > 2 {
		return fmt.Errorf("invalid claude temperature: %f (must be 0-2)", c.Claude.Temperature)
	}

	if c.Claude.Timeout <= 0 {
		return fmt.Errorf("invalid claude timeout: %d", c.Claude.Timeout)
	}

	return nil
}

// String returns a string representation of the config (for debugging)
func (c *Config) String() string {
	return fmt.Sprintf("Config{Master: %s:%d, Worker: %s, Claude: %s}",
		c.Master.Host, c.Master.Port, c.Worker.MasterURL, c.Claude.ServerURL)
}
