// Package tls provides utilities for constructing TLS configurations used
// when connecting to HashiCorp Vault over HTTPS.
//
// It supports:
//   - Custom CA certificate pools for private PKI environments
//   - Mutual TLS (mTLS) via client certificate and key pairs
//   - Insecure skip-verify mode for development/testing
//
// Example usage:
//
//	cfg := tls.Config{
//		CACertFile:     "/etc/vault/ca.pem",
//		ClientCertFile: "/etc/vault/client.pem",
//		ClientKeyFile:  "/etc/vault/client-key.pem",
//	}
//	tlsCfg, err := tls.Build(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
package tls
