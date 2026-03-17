VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
LOCAL_BIN_DIR ?= /Users/vabole/.local/bin
LOCAL_BIN ?= $(LOCAL_BIN_DIR)/confluence

.PHONY: build lint test clean install-local verify-local

build:
	go build $(LDFLAGS) -o bin/confluence ./cmd/confluence

lint:
	golangci-lint run ./...
	./scripts/check-file-length.sh

test:
	go test ./...

install-local:
	mkdir -p $(LOCAL_BIN_DIR)
	go build $(LDFLAGS) -o $(LOCAL_BIN) ./cmd/confluence

verify-local:
	test -x $(LOCAL_BIN)
	$(LOCAL_BIN) version >/dev/null
	$(LOCAL_BIN) pages search --help >/dev/null
	$(LOCAL_BIN) pages get --help >/dev/null
	$(LOCAL_BIN) spaces list --help >/dev/null
	$(LOCAL_BIN) --format plain version >/dev/null

clean:
	rm -rf bin/ dist/
