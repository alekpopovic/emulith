.PHONY: build test run clean

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILT ?= unknown
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.built=$(BUILT)

build:
	go build -ldflags "$(LDFLAGS)" -o emulith ./cmd/emulith

test:
	go test ./...

run:
	go run ./cmd/emulith serve

clean:
	rm -f emulith
