// Package tls provides TLS configuration helpers for secure Vault communication.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config holds TLS configuration parameters.
type Config struct {
	CACertFile     string
	ClientCertFile string
	ClientKeyFile  string
	InsecureSkipVerify bool
}

// Build constructs a *tls.Config from the given Config.
// If CACertFile is empty and InsecureSkipVerify is false, the system
// certificate pool is used.
func Build(cfg Config) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec
	}

	if cfg.CACertFile != "" {
		pool, err := loadCACert(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("tls: load CA cert: %w", err)
		}
		tlsCfg.RootCAs = pool
	}

	if cfg.ClientCertFile != "" || cfg.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCertFile, cfg.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("tls: load client key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}

// loadCACert reads a PEM-encoded CA certificate file and returns a cert pool.
func loadCACert(path string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("no valid PEM certificates found in %s", path)
	}
	return pool, nil
}
