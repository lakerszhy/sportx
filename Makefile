lint:
	golangci-lint cache clean
	golangci-lint run -v

.PHONY: lint