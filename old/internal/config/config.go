package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// DeploymentMode represents the deployment mode
type DeploymentMode int

const (
	ModeStandalone DeploymentMode = iota
	ModeCloud
)

func (m DeploymentMode) String() string {
	switch m {
	case ModeCloud:
		return "cloud"
	case ModeStandalone:
		return "standalone"
	default:
		return "unknown"
	}
}

// ParseDeploymentMode parses a string to DeploymentMode
func ParseDeploymentMode(s string) DeploymentMode {
	switch s {
	case "cloud":
		return ModeCloud
	case "standalone":
		return ModeStandalone
	default:
		return ModeStandalone
	}
}

// Config is the main configuration structure
type Config struct {
	Mode    DeploymentMode
	Master  MasterConfig   `mapstructure:"master"`
	Worker  WorkerConfig   `mapstructure:"worker"`
	Claude  ClaudeConfig   `mapstructure:"claude"`
	Log     LogConfig      `mapstructure:"log"`
}

// MasterConfig contains Master server configuration
type MasterConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	AutoShutdown   bool          `mapstructure:"auto_shutdown"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	DBPath         string        `mapstructure:"db_path"`
}

// WorkerConfig contains Worker client configuration
type WorkerConfig struct {
	MasterURL      string        `mapstructure:"master_url"`
	Name           string        `mapstructure:"name"`
	Auto           bool          `mapstructure:"auto"`
	OneShot        bool          `mapstructure:"oneshot"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	PollInterval   time.Duration `mapstructure:"poll_interval"`
}

// ClaudeConfig contains Claude API configuration
type ClaudeConfig struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json or text
	File   string `mapstructure:"file"`
}

// rawConfig is used for unmarshaling (Mode is string)
type rawConfig struct {
	Mode    string
	Master  MasterConfig   `mapstructure:"master"`
	Worker  WorkerConfig   `mapstructure:"worker"`
	Claude  ClaudeConfig   `mapstructure:"claude"`
	Log     LogConfig      `mapstructure:"log"`
}

// Load loads configuration from file, flags, and environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	setDefaults(cfg)

	// Read from config file if provided
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Unmarshal to raw config (Mode as string)
	var raw rawConfig
	if err := viper.Unmarshal(&raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Copy fields to cfg
	cfg.Master = raw.Master
	cfg.Worker = raw.Worker
	cfg.Claude = raw.Claude
	cfg.Log = raw.Log

	// Parse deployment mode
	cfg.Mode = ParseDeploymentMode(raw.Mode)

	// Re-apply defaults to ensure empty values are set
	if cfg.Master.DBPath == "" {
		cfg.Master.DBPath = ".claude/cpds-manager/master.db"
	}
	if cfg.Master.Host == "" {
		cfg.Master.Host = "0.0.0.0"
	}
	if cfg.Master.Port == 0 {
		cfg.Master.Port = 8848
	}

	// Override with command-line flags for Master
	if viper.IsSet("master.host") {
		cfg.Master.Host = viper.GetString("master.host")
	}
	if viper.IsSet("master.port") {
		cfg.Master.Port = viper.GetInt("master.port")
	}
	if viper.IsSet("master.auto_shutdown") {
		cfg.Master.AutoShutdown = viper.GetBool("master.auto_shutdown")
	}

	// Override with command-line flags for Worker
	if viper.IsSet("worker.master") {
		cfg.Worker.MasterURL = viper.GetString("worker.master")
	}
	if viper.IsSet("worker.name") {
		cfg.Worker.Name = viper.GetString("worker.name")
	}
	if viper.IsSet("worker.auto") {
		cfg.Worker.Auto = viper.GetBool("worker.auto")
	}
	if viper.IsSet("worker.oneshot") {
		cfg.Worker.OneShot = viper.GetBool("worker.oneshot")
	}

	return cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(cfg *Config) {
	// Master defaults
	cfg.Master.Host = "0.0.0.0"
	cfg.Master.Port = 8848
	cfg.Master.AutoShutdown = false
	cfg.Master.ShutdownTimeout = 60 * time.Second
	cfg.Master.DBPath = ".claude/cpds-manager/master.db"

	// Worker defaults
	cfg.Worker.MasterURL = "http://localhost:8848"
	cfg.Worker.Auto = false
	cfg.Worker.OneShot = false
	cfg.Worker.IdleTimeout = 5 * time.Minute
	cfg.Worker.PollInterval = 30 * time.Second

	// Claude defaults
	cfg.Claude.Model = "claude-3-5-sonnet-20241022"
	cfg.Claude.MaxTokens = 4096

	// Log defaults
	cfg.Log.Level = "info"
	cfg.Log.Format = "text"
}

// GetMasterAddr returns the master server address
func (c *Config) GetMasterAddr() string {
	return fmt.Sprintf("%s:%d", c.Master.Host, c.Master.Port)
}

// IsCloudMode returns true if running in cloud mode
func (c *Config) IsCloudMode() bool {
	return c.Mode == ModeCloud
}

// IsStandaloneMode returns true if running in standalone mode
func (c *Config) IsStandaloneMode() bool {
	return c.Mode == ModeStandalone
}
