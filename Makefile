.PHONY: build install clean test run

BINARY_NAME=efx-skills
BUILD_DIR=bin
VERSION=0.1.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/efx-skills

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

install-local: build
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)
	go clean

test:
	go test -v ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

run-status: build
	./$(BUILD_DIR)/$(BINARY_NAME) status

run-search: build
	./$(BUILD_DIR)/$(BINARY_NAME) search

run-list: build
	./$(BUILD_DIR)/$(BINARY_NAME) list

# Cross-compilation
build-all:
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/efx-skills
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/efx-skills
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/efx-skills
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/efx-skills

# Development
dev:
	go run ./cmd/efx-skills

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
