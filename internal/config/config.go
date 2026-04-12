package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for vaultwatch.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerting AlertingConfig `yaml:"alerting"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertingConfig holds alerting thresholds and intervals.
type AlertingConfig struct {
	WarnThreshold    time.Duration `yaml:"warn_threshold"`
	CriticalThreshold time.Duration `yaml:"critical_threshold"`
	PollInterval     time.Duration `yaml:"poll_interval"`
}

// Load reads and parses the config file at the given path.
// Environment variables VAULT_ADDR and VAULT_TOKEN override file values.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Alerting: AlertingConfig{
			WarnThreshold:    72 * time.Hour,
			CriticalThreshold: 24 * time.Hour,
			PollInterval:     5 * time.Minute,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		cfg.Vault.Address = addr
	}
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		cfg.Vault.Token = token
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks that required fields are present and values are sane.
func (c *Config) Validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required (or set VAULT_ADDR)")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("vault.token is required (or set VAULT_TOKEN)")
	}
	if c.Alerting.CriticalThreshold >= c.Alerting.WarnThreshold {
		return fmt.Errorf("alerting.critical_threshold must be less than warn_threshold")
	}
	if c.Alerting.PollInterval <= 0 {
		return fmt.Errorf("alerting.poll_interval must be positive")
	}
	return nil
}
