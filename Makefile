VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build lint test clean

build:
	go build $(LDFLAGS) -o bin/confluence ./cmd/confluence

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -rf bin/ dist/
