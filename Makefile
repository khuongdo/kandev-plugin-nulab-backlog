# Standard targets (contract C4). CI calls exactly these targets.
.DEFAULT_GOAL := help
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

PLUGIN_ID   := nulab-backlog
VERSION     := $(shell sed -n 's/^version: *"\{0,1\}\([^"]*\)"\{0,1\} *$$/\1/p' manifest.yaml)
MODULE      := github.com/khuongdo/kandev-plugin-nulab-backlog

# The Kandev SDK is a sibling checkout pinned by commit (BR5.5).
SDK_DIR     := ../kandev
SDK_REF     := $(shell cat .kandev-sdk-ref)
SDK_BACKEND := $(abspath $(SDK_DIR))/apps/backend

GO_PKGS     := ./internal/... ./server/...
# Coverage floor and the ONLY exclusions (team Testing Posture). Never lower
# the floor or add exclusions to make the target pass.
COVERAGE_MIN     := 80
COVERAGE_EXCLUDE := server/main.go

GOLANGCI_LINT_VERSION := v2.14.0

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64
BUILD     := build
STAGE     := $(BUILD)/stage
DIST      := dist
PACKAGE   := $(DIST)/$(PLUGIN_ID)-$(VERSION).tar.gz
LDFLAGS   := -s -w -X $(MODULE)/internal/plugin.Version=$(VERSION) -X $(MODULE)/internal/plugin.SDKRef=$(SDK_REF)

.PHONY: help check-sdk check-format vet lint test coverage ui-build build package verify-package clean

help:
	@echo "Targets: check-format vet lint test coverage build package verify-package clean"

## check-sdk: fail unless ../kandev exists at the commit in .kandev-sdk-ref.
check-sdk:
	@head=$$(git -C $(SDK_DIR) rev-parse HEAD 2>/dev/null) || { echo "check-sdk: $(SDK_DIR) is missing; check out kandev at $(SDK_REF)"; exit 1; }; \
	if [ "$$head" != "$(SDK_REF)" ]; then echo "check-sdk: $(SDK_DIR) is at $$head, expected $(SDK_REF) (.kandev-sdk-ref)"; exit 1; fi

check-format:
	@out=$$(gofmt -l server internal); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	cd ui && npx prettier --check .

vet: check-sdk
	go vet ./...

lint: check-sdk
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...
	cd ui && npx tsc --noEmit
	cd ui && npx eslint .

test: check-sdk
	go test -race $(GO_PKGS)
	cd ui && npx vitest run

coverage: check-sdk
	@mkdir -p $(BUILD)
	go test -race -coverprofile=coverage.out $(GO_PKGS)
	@grep -v -F $(foreach f,$(COVERAGE_EXCLUDE),-e '$(MODULE)/$(f):') coverage.out > $(BUILD)/coverage.filtered.out
	@total=$$(go tool cover -func=$(BUILD)/coverage.filtered.out | awk '/^total:/ {sub("%","",$$3); print $$3}'); \
	echo "coverage: $$total% (floor $(COVERAGE_MIN)%, excluded: $(COVERAGE_EXCLUDE))"; \
	awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN { exit (t + 0 >= m + 0) ? 0 : 1 }' || { echo "coverage: below the $(COVERAGE_MIN)% floor"; exit 1; }

## ui-build: one self-contained ES module; React comes from the host, never the bundle.
ui-build:
	cd ui && npx esbuild src/index.ts --bundle --format=esm --platform=browser --target=es2022 \
		--jsx=transform --jsx-factory=h --outfile=../$(BUILD)/ui/bundle.js
	@if grep -q 'react-dom\|__SECRET_INTERNALS\|react.production' $(BUILD)/ui/bundle.js; then echo "ui-build: React was bundled"; exit 1; fi

build: check-sdk
	@mkdir -p $(BUILD)/server
	@for p in $(PLATFORMS); do \
		os=$${p%-*}; arch=$${p#*-}; ext=""; [ "$$os" = windows ] && ext=".exe"; \
		echo "build: $$p"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD)/server/plugin-$$p$$ext ./server; \
	done

package: build ui-build
	rm -rf $(STAGE) && mkdir -p $(STAGE)/server $(STAGE)/ui $(DIST)
	cp manifest.yaml $(STAGE)/
	cp $(BUILD)/server/plugin-* $(STAGE)/server/
	cp $(BUILD)/ui/bundle.js $(STAGE)/ui/
	go -C $(SDK_BACKEND) run ./cmd/plugin-pack -dir $(abspath $(STAGE)) -out $(abspath $(PACKAGE))
	cd $(DIST) && sha256sum $(notdir $(PACKAGE)) > checksums.txt

## verify-package: package checksum, per-file checksums, contents and manifest
## (BR5.2), checked by this repo's verifier (Kandev v0.96.0 has no verify CLI).
verify-package: check-sdk
	go run ./cmd/verifypkg -archive $(PACKAGE) -checksums $(DIST)/checksums.txt -expected-id $(PLUGIN_ID) -expected-version $(VERSION)

clean:
	rm -rf $(BUILD) $(DIST) coverage.out
