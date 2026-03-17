VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
LOCAL_BIN_DIR ?= /Users/vabole/.local/bin
LOCAL_BIN ?= $(LOCAL_BIN_DIR)/confluence

.PHONY: build fmt fmt-check lint test test-live test-live-update dev-refresh clean install-local verify-local

build:
	go build $(LDFLAGS) -o bin/confluence ./cmd/confluence

fmt:
	./scripts/fmt-go.sh

fmt-check:
	./scripts/check-gofmt.sh

lint: fmt-check
	golangci-lint run ./...
	./scripts/check-file-length.sh

test:
	go test ./...

test-live:
	CONFLUENCE_LIVE_E2E=1 go test -run TestLiveAPIContractGolden_Integration ./...

test-live-update:
	CONFLUENCE_LIVE_E2E=1 CONFLUENCE_LIVE_E2E_UPDATE=1 go test -run TestLiveAPIContractGolden_Integration ./...

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

dev-refresh:
	$(MAKE) fmt
	$(MAKE) test
	$(MAKE) install-local
	$(MAKE) verify-local

clean:
	rm -rf bin/ dist/
