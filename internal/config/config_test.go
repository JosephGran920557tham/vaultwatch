package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTemp(t, `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
alerting:
  warn_threshold: 48h
  critical_threshold: 12h
  poll_interval: 1m
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("expected address %q, got %q", "http://127.0.0.1:8200", cfg.Vault.Address)
	}
	if cfg.Alerting.WarnThreshold != 48*time.Hour {
		t.Errorf("expected warn_threshold 48h, got %v", cfg.Alerting.WarnThreshold)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	path := writeTemp(t, `vault:
  address: "http://original:8200"
  token: "old-token"
`)
	t.Setenv("VAULT_ADDR", "http://env-override:8200")
	t.Setenv("VAULT_TOKEN", "env-token")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://env-override:8200" {
		t.Errorf("env override failed, got %q", cfg.Vault.Address)
	}
	if cfg.Vault.Token != "env-token" {
		t.Errorf("token env override failed, got %q", cfg.Vault.Token)
	}
}

func TestLoad_MissingFile_UsesDefaults(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "root")

	cfg, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Alerting.PollInterval != 5*time.Minute {
		t.Errorf("expected default poll_interval 5m, got %v", cfg.Alerting.PollInterval)
	}
}

func TestValidate_InvalidThresholds(t *testing.T) {
	cfg := &Config{
		Vault: VaultConfig{Address: "http://127.0.0.1:8200", Token: "root"},
		Alerting: AlertingConfig{
			WarnThreshold:    12 * time.Hour,
			CriticalThreshold: 24 * time.Hour,
			PollInterval:     1 * time.Minute,
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for critical >= warn threshold")
	}
}
