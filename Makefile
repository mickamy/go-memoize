.PHONY: test lint

test:
	go test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		@echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run
