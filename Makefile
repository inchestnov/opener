BINARY := bin/opener
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: build
build:
	go build -o $(BINARY) ./cmd/opener

.PHONY: install
install:
	go install ./cmd/opener
	@echo
	@echo "Installed: $(GOBIN)/opener"
	@echo
	@if echo "$(PATH)" | tr ':' '\n' | grep -qx "$(GOBIN)"; then \
		echo "$(GOBIN) is already on your PATH — run 'opener' from a new shell."; \
	else \
		echo "$(GOBIN) is NOT on your PATH. Add it, then reload your shell:"; \
		echo; \
		echo "  # bash (~/.bashrc) or zsh (~/.zshrc)"; \
		echo "  export PATH=\"$(GOBIN):\$$PATH\""; \
		echo; \
		echo "  # fish (~/.config/fish/config.fish)"; \
		echo "  fish_add_path $(GOBIN)"; \
		echo; \
		echo "Then reload: exec \$$SHELL -l   (or open a new terminal)"; \
	fi
	@echo

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: check
check: vet lint test

.PHONY: clean
clean:
	rm -rf bin
