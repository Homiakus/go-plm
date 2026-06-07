.PHONY: all build run test lint clean frontend-install frontend-build frontend-dev

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

build: frontend-build
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/plm/

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	$(GOTEST) -v -race -count=1 ./...

test-cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

lint:
	$(GOVET) ./...
	@which $(GOLINT) >/dev/null 2>&1 && $(GOLINT) run ./... || echo "golangci-lint not installed — skipping"

bench:
	$(GOTEST) -bench=. -benchmem ./...

# === Frontend ===
frontend-install:
	cd $(FRONTEND_DIR) && $(NPM) install

frontend-build:
	cd $(FRONTEND_DIR) && $(NPM) run build

frontend-dev:
	cd $(FRONTEND_DIR) && $(NPM) run dev

# === Docker (future) ===
docker-build:
	docker build -t go-plm .

# === Clean ===
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	cd $(FRONTEND_DIR) && rm -rf dist/

# === CI ===
ci: lint test build
