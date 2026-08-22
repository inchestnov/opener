BINARY := bin/opener

.PHONY: build
build:
	go build -o $(BINARY) ./cmd/opener

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf bin
