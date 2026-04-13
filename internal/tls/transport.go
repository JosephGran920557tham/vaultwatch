package tls

import (
	"fmt"
	"net/http"
	"time"
)

// TransportConfig extends Config with HTTP transport tuning parameters.
type TransportConfig struct {
	Config
	DialTimeout         time.Duration
	TLSHandshakeTimeout time.Duration
	MaxIdleConns        int
}

// DefaultTransportConfig returns a TransportConfig with sensible defaults.
func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		DialTimeout:         10 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        100,
	}
}

// NewHTTPClient constructs an *http.Client configured with TLS settings
// derived from the given TransportConfig.
func NewHTTPClient(cfg TransportConfig) (*http.Client, error) {
	if cfg.DialTimeout <= 0 {
		return nil, fmt.Errorf("tls: DialTimeout must be positive")
	}

	tlsCfg, err := Build(cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("tls: build config: %w", err)
	}

	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 100
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsCfg,
		TLSHandshakeTimeout: cfg.TLSHandshakeTimeout,
		MaxIdleConns:        maxIdle,
		IdleConnTimeout:     90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.DialTimeout,
	}, nil
}
