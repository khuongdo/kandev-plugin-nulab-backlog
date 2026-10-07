# Standard targets (contract C4). CI calls exactly these targets.
.DEFAULT_GOAL := help
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

PLUGIN_ID   := nulab-backlog
VERSION     := $(shell sed -n 's/^version: *"\{0,1\}\([^"]*\)"\{0,1\} *$$/\1/p' manifest.yaml)
# The minimum Kandev version comes from the manifest, never from code (AC7.4.1).
MIN_KANDEV_VERSION := $(shell sed -n 's/^min_kandev_version: *"\{0,1\}\([^"]*\)"\{0,1\} *$$/\1/p' manifest.yaml)
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
ACTIONLINT_VERSION    := v1.7.12

# Packaged-host contract test (US7.4): a throwaway Kandev built from a
# checkout at exactly v$(MIN_KANDEV_VERSION).
KANDEV_MIN_DIR ?= ../kandev-min
CONTRACT_PORT  ?= 38529
# Kandev's backend will not start without its agentctl sidecar; give it its own
# port so the test never touches a Kandev already running on this machine.
CONTRACT_AGENTCTL_PORT ?= 39529

MARKETPLACE_REPO := khuongdo/kandev-plugin-nulab-backlog
REGISTRY         ?=

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64
BUILD     := build
STAGE     := $(BUILD)/stage
DIST      := dist
PACKAGE   := $(DIST)/$(PLUGIN_ID)-$(VERSION).tar.gz
LDFLAGS   := -s -w -X $(MODULE)/internal/plugin.Version=$(VERSION) -X $(MODULE)/internal/plugin.SDKRef=$(SDK_REF)

.PHONY: help check-sdk check-format vet lint test coverage check-secrets ui-build build package verify-package \
	contract-test release-preflight marketplace-entry clean

help:
	@echo "Targets: check-format vet lint test coverage check-secrets build package verify-package clean"
	@echo "Release: contract-test [KANDEV_MIN_DIR=...] release-preflight TAG=vX.Y.Z marketplace-entry [REGISTRY=plugins.yaml]"

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
	go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)
	go run ./cmd/ci workflows -dir .github/workflows

test: check-sdk
	go test -race $(GO_PKGS)
	cd ui && npx vitest run

# The profile stays under build/, never at the repository root.
coverage: check-sdk
	@mkdir -p $(BUILD)
	go test -race -coverprofile=$(BUILD)/coverage.out $(GO_PKGS)
	@grep -v -F $(foreach f,$(COVERAGE_EXCLUDE),-e '$(MODULE)/$(f):') $(BUILD)/coverage.out > $(BUILD)/coverage.filtered.out
	@total=$$(go tool cover -func=$(BUILD)/coverage.filtered.out | awk '/^total:/ {sub("%","",$$3); print $$3}'); \
	echo "coverage: $$total% (floor $(COVERAGE_MIN)%, excluded: $(COVERAGE_EXCLUDE))"; \
	awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN { exit (t + 0 >= m + 0) ? 0 : 1 }' || { echo "coverage: below the $(COVERAGE_MIN)% floor"; exit 1; }

## check-secrets: no credential-shaped string in test data or test artifacts
## (AC7.3.4). Run after coverage so coverage.out is scanned too.
check-secrets: check-sdk
	go run ./cmd/ci secrets -root .

## ui-build: one self-contained ES module; React comes from the host, never the bundle.
ui-build:
	cd ui && npx esbuild src/index.ts --bundle --format=esm --platform=browser --target=es2022 \
		--jsx=transform --jsx-factory=h --jsx-fragment=Fragment --outfile=../$(BUILD)/ui/bundle.js
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

## contract-test: build Kandev v$(MIN_KANDEV_VERSION) from $(KANDEV_MIN_DIR), start it
## (with its agentctl sidecar) and a temporary home, install $(PACKAGE) and
## run it (US7.4). Needs gcc.
contract-test:
	@head=$$(git -C $(KANDEV_MIN_DIR) rev-parse HEAD 2>/dev/null) || { echo "contract-test: $(KANDEV_MIN_DIR) is missing; check out kandev at v$(MIN_KANDEV_VERSION)"; exit 1; }; \
	want=$$(git -C $(KANDEV_MIN_DIR) rev-parse "v$(MIN_KANDEV_VERSION)^{commit}" 2>/dev/null) || want=""; \
	if [ "$$head" != "$$want" ]; then echo "contract-test: $(KANDEV_MIN_DIR) is at $$head, expected the tag v$(MIN_KANDEV_VERSION)"; exit 1; fi
	$(MAKE) -C $(KANDEV_MIN_DIR)/apps/backend build-kandev build-agentctl VERSION=v$(MIN_KANDEV_VERSION)
	@home=$$(mktemp -d); \
	KANDEV_HOME_DIR="$$home" KANDEV_SERVER_HOST=127.0.0.1 KANDEV_SERVER_PORT=$(CONTRACT_PORT) AGENTCTL_PORT=$(CONTRACT_AGENTCTL_PORT) \
		$(KANDEV_MIN_DIR)/apps/backend/bin/kandev __backend > "$$home/kandev.log" 2>&1 & pid=$$!; \
	trap 'kill $$pid 2>/dev/null || true; wait $$pid 2>/dev/null || true; rm -rf "$$home"' EXIT; \
	go run ./cmd/ci contract -base-url http://127.0.0.1:$(CONTRACT_PORT) -package $(PACKAGE) \
		-plugin-id $(PLUGIN_ID) -host-version v$(MIN_KANDEV_VERSION) \
		|| { echo "contract-test: failed; last Kandev log lines:"; tail -n 40 "$$home/kandev.log"; exit 1; }

## release-preflight: refuse a malformed tag, a tag that differs from the
## manifest version, a tag not on main, a tag that already has a Release, and a
## first release without a passing first-release record (AC7.5.2, AC7.5.3).
## "First release" means no published stable GitHub Release yet, never "no other
## git tag" (R-01). A failed gh call passes no list and the preflight refuses (R-03).
## ponytail: --limit 1000 Releases; page through the API if the repo ever has more.
release-preflight: check-sdk
	@test -n "$(TAG)" || { echo "release-preflight: TAG is required, for example TAG=v0.1.0"; exit 1; }
	@on_main=false; git merge-base --is-ancestor "$(TAG)^{commit}" origin/main && on_main=true; \
	releases=$$(gh release list --limit 1000 --json tagName,isDraft,isPrerelease) || releases=""; \
	go run ./cmd/ci preflight -tag "$(TAG)" -version "$(VERSION)" -on-main=$$on_main \
		-releases "$$releases" -manual-checks docs/manual-checks

## marketplace-entry: print the plugin-registry/plugins.yaml entry; with
## REGISTRY=<plugins.yaml>, also check it against the catalogue (AC7.6.1, AC7.6.2).
marketplace-entry: check-sdk
	go run ./cmd/ci marketplace -repo $(MARKETPLACE_REPO) -categories integrations $(if $(REGISTRY),-registry $(REGISTRY))

clean:
	rm -rf $(BUILD) $(DIST) coverage.out
