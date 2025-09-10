# WireGuard Gateway Portal (wg-portal)

A lightweight web interface for managing WireGuard VPN connections.

Ideal for managing the VPN connection running on a RaspberryPi Network Gateway.

## Installation

```bash
curl -fsSL https://git.sr.ht/~a14m/wg-portal/blob/main/deployment/install.sh | sudo bash
```

## Development

Dependencies:

- **Go 1.25+** - For running the application
- **Node.js/npx** - (optional) For frontend linting (HTML/CSS/JS)
- **golangci-lint** - (optional) Go linting

Steps:

1. Clone the repo
1. Configure WireGuard connections in `/etc/wireguard/*.conf`
1. Update configuration in `<repo>/config.yaml` (optional)
1. Run the application: `go run main.go`
1. Open your browser to `http://localhost:8080`

## Uninstall

```bash
curl -fsSL https://git.sr.ht/~a14m/wg-portal/blob/main/deployment/uninstall.sh | sudo bash
```

## Linting

```bash
# Individual linters
make lint-js    # JavaScript (ESLint)
make lint-css   # CSS (Stylelint)
make lint-html  # HTML (html-validate)
make lint-go    # Go (go vet + golangci-lint)
make lint-shell # Shell scripts (shellcheck)

# All linters
make lint

# Help
make help
```

## Release

```bash
git tag -a v[major].[minor].[patch]
git push --tags
```
