APP := cview
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)
PKGS := ./...
GO ?= go
GOCACHE ?= /tmp/cview-gocache
GOMODCACHE ?= /tmp/cview-gomodcache

.PHONY: build run test fmt tidy clean

build:
	mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BIN) ./cmd/cview

run:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) run ./cmd/cview

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test $(PKGS)

fmt:
	$(GO) fmt $(PKGS)

tidy:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)
