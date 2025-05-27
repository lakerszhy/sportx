lint:
	golangci-lint cache clean
	golangci-lint run -v

release:
	goreleaser release  --snapshot --clean

.PHONY: lint release