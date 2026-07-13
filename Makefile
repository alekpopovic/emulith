.PHONY: build test run clean docker-build docker-run compatibility

IMAGE ?= emulith/emulith
TAG ?= dev

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILT ?= unknown
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.built=$(BUILT)

build:
	go build -ldflags "$(LDFLAGS)" -o emulith ./cmd/emulith

test:
	go test ./...

compatibility:
	AWS_EC2_METADATA_DISABLED=true go test ./test/compatibility/aws/...

run:
	go run ./cmd/emulith serve

clean:
	rm -f emulith

docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg BUILT=$(BUILT) -t $(IMAGE):$(TAG) .

docker-run: docker-build
	docker run --rm -p 4566:4566 -v emulith-data:/var/lib/emulith $(IMAGE):$(TAG)
