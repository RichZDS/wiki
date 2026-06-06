.PHONY: build run test clean

APP_NAME = aisearch
BUILD_DIR = build

build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME).exe .

build-linux:
	@echo "Building $(APP_NAME) for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) .

run:
	go run main.go

run-prod:
	go run main.go prod

test:
	go test ./tests/...

test-verbose:
	go test -v ./tests/...

clean:
	rm -rf $(BUILD_DIR)
