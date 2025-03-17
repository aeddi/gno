GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)

rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: build
build:
	go build -o build/gasstat .

.PHONY: install
install:
	go install .

.PHONY: lint
lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

.PHONY: test
test:
	go test $(GOTEST_FLAGS) -v ./...

.PHONY: clean
clean:
	rm -rf build
