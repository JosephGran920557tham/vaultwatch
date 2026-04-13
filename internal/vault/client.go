package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with lease-specific helpers.
type Client struct {
	api     *vaultapi.Client
	Address string
}

// LeaseInfo holds metadata about a single Vault lease.
type LeaseInfo struct {
	LeaseID        string
	Renewable      bool
	TTL            time.Duration
	ExpiresAt      time.Time
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(token)

	return &Client{
		api:     api,
		Address: address,
	}, nil
}

// LookupLease retrieves TTL and metadata for the given lease ID.
func (c *Client) LookupLease(leaseID string) (*LeaseInfo, error) {
	secret, err := c.api.Sys().Lookup(leaseID)
	if err != nil {
		return nil, fmt.Errorf("looking up lease %q: %w", leaseID, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data returned for lease %q", leaseID)
	}

	ttlRaw, ok := secret.Data["ttl"]
	if !ok {
		return nil, fmt.Errorf("ttl missing from lease data for %q", leaseID)
	}

	ttlFloat, ok := ttlRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("unexpected ttl type for lease %q", leaseID)
	}

	ttl := time.Duration(ttlFloat) * time.Second
	renewable, _ := secret.Data["renewable"].(bool)

	return &LeaseInfo{
		LeaseID:   leaseID,
		Renewable: renewable,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// IsHealthy checks whether the Vault server is reachable and unsealed.
func (c *Client) IsHealthy() error {
	health, err := c.api.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}
	return nil
}
