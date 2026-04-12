# vaultwatch

A lightweight CLI tool that monitors HashiCorp Vault secret leases and alerts on upcoming expirations.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Set your Vault address and token, then run:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

vaultwatch --warn-before 48h
```

This will scan all active leases and print a warning for any secrets expiring within the next 48 hours.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--warn-before` | `24h` | Alert threshold before expiration |
| `--path` | `/` | Vault secret path to monitor |
| `--output` | `text` | Output format: `text` or `json` |

### Example Output

```
[WARN] secret/db/prod        expires in 6h32m  (2024-11-10 14:00 UTC)
[WARN] secret/api/stripe     expires in 23h15m (2024-11-11 07:00 UTC)
[OK]   secret/tls/cert       expires in 72h00m
```

---

## Requirements

- Go 1.21+
- HashiCorp Vault 1.10+

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)