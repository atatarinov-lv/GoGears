PWD := $(shell pwd)

GOLANG_LINTER_VERSION := 1.63.4
GOLANG_LINTER_DOCKER = docker run --rm -t -w /app \
	-v $(PWD):/app \
	-v ~/.cache/golangci-lint/v$(GOLANG_LINTER_VERSION):/root/.cache \
	golangci/golangci-lint:v$(GOLANG_LINTER_VERSION)

GO_FILES = $(shell find -name '*.go')

.PHONY: install-tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/daixiang0/gci@latest
	go install mvdan.cc/gofumpt@latest

.PHONY: tidy
tidy:
	goimports -w $(GO_FILES)
	gci write $(GO_FILES)
	gofumpt -l -w $(GO_FILES)
	go mod tidy

.PHONY: tests
tests:
	go test ./... -v -coverprofile coverage.out
	go tool cover -func coverage.out
	go tool cover -html coverage.out -o coverage.html
	rm coverage.out

.PHONY: lint
lint:
	$(GOLANG_LINTER_DOCKER) golangci-lint run

