.PHONY: lint lint-html lint-css lint-go lint-js

help:
	@echo "Available lint commands:"
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

lint: lint-go lint-js lint-html lint-css
	@echo "All linting complete!"
