.PHONY: build test race vet fmt smoke clean

BINARY ?= vmlg
GOCACHE ?= /tmp/vmlg-gocache
GO_FILES := $(shell find . -name '*.go' -not -path './.git/*')

build:
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) go build -buildvcs=false -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/vmlg

test:
	GOCACHE=$(GOCACHE) go test ./...

race:
	GOCACHE=$(GOCACHE) go test -race ./...

vet:
	GOCACHE=$(GOCACHE) go vet ./...

fmt:
	gofmt -w $(GO_FILES)

smoke:
	GOCACHE=$(GOCACHE) GOFLAGS=-buildvcs=false go run ./cmd/vmlg version
	GOCACHE=$(GOCACHE) GOFLAGS=-buildvcs=false go run ./cmd/vmlg version -json
	GOCACHE=$(GOCACHE) GOFLAGS=-buildvcs=false go run ./cmd/vmlg check -config ./examples/standalone.toml
	GOCACHE=$(GOCACHE) GOFLAGS=-buildvcs=false go run ./cmd/vmlg check -config ./examples/hub.toml
	GOCACHE=$(GOCACHE) GOFLAGS=-buildvcs=false go run ./cmd/vmlg check -config ./examples/agent.toml

clean:
	rm -f $(BINARY)
	rm -rf bin dist
