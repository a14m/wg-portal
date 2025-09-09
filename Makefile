.PHONY: lint lint-html lint-css lint-go lint-js build build-all clean deps verify

help:
	@echo "Available commands:"
	@echo "  make deps      - Download and verify Go dependencies"
	@echo "  make build     - Build for multiple architectures (amd64, arm64)"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make verify    - Verify Go module checksums"
	@echo "  make lint      - Run all linters"
	@echo "  make lint-html - Lint HTML files (html-validate)"
	@echo "  make lint-css  - Lint CSS files (stylelint)"
	@echo "  make lint-go   - Lint Go files (go vet + golangci-lint)"
	@echo "  make lint-js   - Lint JavaScript files (eslint)"

lint-html:
	@echo "Linting HTML files..."
	npx html-validate templates/*.html

lint-css:
	@echo "Linting CSS files..."
	npx stylelint static/css/*.css

lint-go:
	@echo "Linting Go files..."
	go vet ./...
	golangci-lint run

lint-js:
	@echo "Linting JavaScript files..."
	npx eslint static/js/*.js

lint-shell:
	@echo "Linting shell files..."
	shellcheck deployment/install.sh
	shellcheck deployment/uninstall.sh

deps:
	@echo "Downloading Go dependencies..."
	go mod download
	@echo "Verifying Go module checksums..."
	go mod verify
	@echo "Cleaning up unused dependencies..."
	go mod tidy
	@echo "Dependencies ready!"

verify:
	@echo "Verifying Go module checksums..."
	go mod verify

build: deps
	@echo "Building for multiple Linux architectures..."
	@mkdir -p dist
	@echo "Building for Linux amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/wg-portal-linux-amd64 main.go
	@echo "Building for Linux arm64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/wg-portal-linux-arm64 main.go
	@echo "Building for Linux arm..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o dist/wg-portal-linux-arm main.go
	@echo "Generating checksums..."
	@cd dist && sha256sum * > checksums.txt
	@echo "All builds complete! Check the dist/ directory."

clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/
	@echo "Clean complete!"

lint: lint-go lint-js lint-html lint-css lint-shell
	@echo "All linting complete!"
