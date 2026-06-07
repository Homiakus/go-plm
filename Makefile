.PHONY: all build run test test-race test-cover lint security clean frontend-install frontend-build frontend-dev frontend-audit

# === Go ===
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOLINT := golangci-lint

# === Frontend ===
FRONTEND_DIR := frontend
NPM := npm

# === Build ===
BINARY_NAME := plm
BUILD_DIR := build

all: lint test build

# Build frontend first, then embed into Go binary
build: frontend-build
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/plm/

# Run the built binary (starts HTTP server + opens browser)
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) .

# Start a new demo project and open it
demo: build
	./$(BUILD_DIR)/$(BINARY_NAME) init demo-project 'Demo PLM Project'
	./$(BUILD_DIR)/$(BINARY_NAME) demo-project

test: frontend-build
	$(GOTEST) -count=1 ./...

test-race: frontend-build
	$(GOTEST) -race -count=1 ./...

test-cover: frontend-build
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

lint: frontend-build
	$(GOVET) ./...
	@which $(GOLINT) >/dev/null 2>&1 && $(GOLINT) run ./... || echo "golangci-lint not installed — skipping"

security: frontend-build
	@which govulncheck >/dev/null 2>&1 && govulncheck ./... || echo "govulncheck not installed — skipping"

bench:
	$(GOTEST) -bench=. -benchmem ./...

# === Frontend ===
frontend-install:
	cd $(FRONTEND_DIR) && $(NPM) install

frontend-build:
	cd $(FRONTEND_DIR) && $(NPM) run build

frontend-dev:
	cd $(FRONTEND_DIR) && $(NPM) run dev

frontend-audit:
	cd $(FRONTEND_DIR) && $(NPM) audit --omit=dev

# === Clean ===
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	cd $(FRONTEND_DIR) && rm -rf dist/

# === CI ===
ci: lint test security frontend-audit build
