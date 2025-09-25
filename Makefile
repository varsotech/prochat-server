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

.PHONY: generate
generate: install-buf
	go generate ./...
