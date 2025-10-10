BUF_VERSION ?= v1.57.0
BIN_DIR := $(HOME)/.local/bin

.PHONY: install-buf
install-buf:
	@if [ ! -x "$(BIN_DIR)/buf" ]; then \
		echo "Installing buf $(BUF_VERSION)..."; \
		mkdir -p $(BIN_DIR); \
		curl -sSL -o $(BIN_DIR)/buf https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-$(shell uname -s)-$(shell uname -m); \
		chmod +x $(BIN_DIR)/buf; \
	fi

.PHONY: install-gotestsum
install-gotestsum:
	@if [ ! -x "$(shell go env GOPATH)/bin/gotestsum" ]; then \
		go install gotest.tools/gotestsum@v1.13.0; \
  	fi

.PHONY: install-goimports
install-goimports:
	@if [ ! -x "$(shell go env GOPATH)/bin/gotestsum" ]; then \
		go install golang.org/x/tools/cmd/goimports@latest; \
  	fi

.PHONY: generate
generate: install-buf
	go generate ./...

.PHONY: test
test: install-gotestsum
	@gotestsum --format testname
	
.PHONY: test-watch
test-watch: install-gotestsum
	@gotestsum --format testname --watch --watch-clear
	
.PHONY: lint
lint: install-goimports
	@goimports -w .
