lint:
	golangci-lint cache clean
	golangci-lint run -v

release:
	goreleaser release --skip=publish --clean

.PHONY: lint release