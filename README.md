# WireGuard Gateway Portal

A lightweight web interface for managing WireGuard VPN connections.

Ideal for managing the VPN connection running on a RaspberryPi Network Gateway.

## Quick Start

1. **Configure WireGuard connections** in `/etc/wireguard/*.conf`
1. **Update configuration** in `config.yaml` (optional)
1. **Run the application**: `go run main.go`
1. **Open your browser** to `http://localhost:8080`

## Development

- **Go 1.25+** - For running the application
- **Node.js/npx** - (optional) For frontend linting (HTML/CSS/JS)
- **golangci-lint** - (optional) Go linting

### Linting

```bash
# Individual linters
make lint-js    # JavaScript (ESLint)
make lint-css   # CSS (Stylelint)
make lint-html  # HTML (html-validate)
make lint-go    # Go (go vet + golangci-lint)

# All linters
make lint

# Help
make help
```
