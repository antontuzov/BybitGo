# Makefile for Bybit Trading Bot

# Variables
BINARY_NAME=bot
MAIN_FILE=cmd/bot/main.go
STATIC_DIR=web/static

# Default target
.PHONY: all
all: build

# Build the bot
.PHONY: build
build:
	go build -o ${BINARY_NAME} ${MAIN_FILE}

# Run the bot
.PHONY: run
run: build
	./${BINARY_NAME}

# Install dependencies
.PHONY: deps
deps:
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	rm -f ${BINARY_NAME}

# Run tests
.PHONY: test
test:
	go test ./...

# Run the bot in the background
.PHONY: background
background:
	./run-bot.sh

# Deploy to Vercel (requires vercel CLI)
.PHONY: deploy-vercel
deploy-vercel:
	vercel --prod

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Build the bot (default)"
	@echo "  build        - Build the bot"
	@echo "  run          - Build and run the bot"
	@echo "  deps         - Install dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  background   - Run the bot in the background"
	@echo "  deploy-vercel- Deploy to Vercel (requires vercel CLI)"
	@echo "  help         - Show this help message"