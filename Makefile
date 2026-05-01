APP := passkey-sudo
BIN_DIR := bin

.PHONY: build test fmt vet tidy install clean

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/passkey-sudo

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

tidy:
	go mod tidy

install: build
	sudo install -m 0755 $(BIN_DIR)/$(APP) /usr/local/bin/$(APP)

clean:
	rm -rf $(BIN_DIR) dist coverage.out
