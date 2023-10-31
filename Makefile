.PHONY: build clean run

BINARY_NAME := dddns
BUILD_DIR := ./bin

GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GORUN := $(GOCMD) run

all: test build

build:
	@echo "  >  Building binary..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/dddns/main.go

clean:
	@echo "  >  Cleaning build cache"
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

run:
	@echo "  >  Running application..."
	$(GORUN) cmd/dddns/main.go