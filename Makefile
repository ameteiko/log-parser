check: test lint

test: ## Runs the unit tests.
	@go test $(PKG_NAME)

lint: ## Runs the linter.
	@golangci-lint run