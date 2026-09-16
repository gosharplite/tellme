# tellme — development tasks
#
# Build / test conventions for the tellme CLI (round 001 — narrow foundation).
# See specs/truth/techstack.md and the plan package research.md (Decisions 4 & 7).

VERSION ?= dev

STATICCHECK := $(shell command -v staticcheck 2>/dev/null)
GOLANGCI := $(shell command -v golangci-lint 2>/dev/null)
GOVULNCHECK := $(shell command -v govulncheck 2>/dev/null)

.PHONY: help build fmt vet staticcheck tidy lint vulncheck test verify verify-no-test-sleep verify-no-network verify-cross-compile verify-mcp-sdk-confinement

help:
	@echo "tellme development tasks:"
	@echo "  make build                - Build ./cmd/tellme with -X main.version=\$$(VERSION)"
	@echo "  make fmt                  - go fmt ./..."
	@echo "  make vet                  - go vet ./..."
	@echo "  make staticcheck          - run staticcheck ./... (resolved from PATH)"
	@echo "  make lint                 - run golangci-lint ./... (resolved from PATH; errcheck via .golangci.yml)"
	@echo "  make vulncheck            - run govulncheck ./... (resolved from PATH)"
	@echo "  make tidy                 - go mod tidy"
	@echo "  make test                 - go test ./..."
	@echo "  make verify-no-test-sleep - forbid time.Sleep for synchronization in *_test.go (ADR-036 parity)"
	@echo "  make verify-no-network    - build-graph capability guard: no net/net/http in ./cmd/tellme closure"
	@echo "  make verify-cross-compile - build + vet the module for every supported POSIX target (linux/darwin, amd64/arm64)"
	@echo "  make verify-mcp-sdk-confinement - verify the MCP Go SDK is imported only under internal/infrastructure/mcp/"
	@echo "  make verify               - aggregate: verify-no-test-sleep + verify-no-network + vet + verify-cross-compile + verify-mcp-sdk-confinement + lint + vulncheck"

# NOTE: `VERSION ?= dev` is the local/release default ONLY.
# The E2E harness must build explicitly with the sentinel
# (go build -ldflags "-X main.version=0.0.0-harness"), never via `make build`,
# so a missed injection cannot pass against the `dev` default.
build:
	go build -ldflags "-X main.version=$(VERSION)" -o tellme ./cmd/tellme

fmt:
	go fmt ./...

vet:
	@if find . -name '*.go' -not -path './.git/*' -not -path './vendor/*' | grep -q .; then \
		go vet ./...; \
	else \
		echo "  (skip) no Go packages to vet yet"; \
	fi

staticcheck:
ifeq ($(STATICCHECK),)
	@echo "staticcheck not found; install: go install honnef.co/go/tools/cmd/staticcheck@latest" >&2
	@exit 1
else
	$(STATICCHECK) ./...
endif

tidy:
	go mod tidy

lint:
ifeq ($(GOLANGCI),)
	@echo "golangci-lint not found; install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest" >&2
	@exit 1
else
	$(GOLANGCI) run ./...
endif

vulncheck:
ifeq ($(GOVULNCHECK),)
	@echo "govulncheck not found; install: go install golang.org/x/vuln/cmd/govulncheck@latest" >&2
	@exit 1
else
	$(GOVULNCHECK) ./...
endif

test: verify-mcp-sdk-confinement
	go test ./...

# Determinism gate (ADR-036 parity): no time.Sleep for synchronization in tests.
verify-no-test-sleep:
	@echo "verify-no-test-sleep: scanning *_test.go for time.Sleep synchronization ..."
	@if grep -rn 'time\.Sleep(' --include='*_test.go' . ; then \
		echo "❌ time.Sleep found in a test file; use deterministic primitives (ready channels / poll loops)."; \
		exit 1; \
	fi
	@echo "  ✓ no time.Sleep in test files"

# Build-graph capability guard (research.md Decision 5, backstop witness).
# Single definition: delegates to the Go guard in tests/e2e (harness), so the
# Makefile and the scenario step never drift (previously two diverging copies).
verify-no-network:
	@echo "verify-no-network: offline-path witness (recording sink + differential) ..."
	@if ! go list ./cmd/tellme >/dev/null 2>&1; then \
		echo "  (skip) ./cmd/tellme package not present yet"; exit 0; \
	fi
	@go test -count=1 -run TestOfflinePathsDoNotContactProvider ./tests/e2e/
	@echo "  ✓ offline paths (--version, -d, prompt-less boot) make no provider request"

# Cross-compile gate (round 020): compile + vet the module for every supported
# target, so build-tagged, OS-specific code (e.g.
# internal/infrastructure/telemetry/system_metrics_{linux,darwin}.go) can never
# silently fail to compile for a platform we ship. Host-independent: it verifies
# the whole matrix regardless of the host GOOS/GOARCH — round 019's darwin
# sampler shipped broken because `make verify` only builds the host target.
# CGO_ENABLED=0 pins pure-Go cross-compilation so an ambient `CGO_ENABLED=1`
# (a host shell/CI export) can never make a cross build fail by invoking the
# host C compiler against target assembly/headers (PR #46 review — TD1).
CROSS_TARGETS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
verify-cross-compile:
	@echo "verify-cross-compile: build + vet for $(CROSS_TARGETS) ..."
	@for target in $(CROSS_TARGETS); do os=$${target%/*}; arch=$${target#*/}; echo "  cross-build $$os/$$arch"; CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build ./... || { echo "❌ go build failed for $$os/$$arch"; exit 1; }; CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go vet ./... || { echo "❌ go vet failed for $$os/$$arch"; exit 1; }; done
	@echo "  ✓ cross-compiles and vets for all supported targets"

# MCP SDK confinement gate (round 032, ADR-067 parity): the official MCP Go SDK
# may be imported ONLY under internal/infrastructure/mcp/ (production AND
# _test.go files, including the SDK-built fake in mcptest/). Every other layer
# consumes the tools.MCPClient domain port. Mirrors the reference's
# verify-mcp-sdk-confinement.
verify-mcp-sdk-confinement:
	@echo "verify-mcp-sdk-confinement: MCP Go SDK imports confined to internal/infrastructure/mcp/ ..."
	@VIOLATIONS="$$( grep -rn 'github.com/modelcontextprotocol/go-sdk' --include='*.go' --exclude-dir=vendor --exclude-dir=.git . | grep -v '^\./internal/infrastructure/mcp/' )"; \
	if [ -n "$$VIOLATIONS" ]; then \
		echo ""; \
		echo "❌ verification violation: MCP Go SDK imported outside internal/infrastructure/mcp/."; \
		echo "   The SDK is confined to the MCP adapter; consume tools.MCPClient instead."; \
		echo ""; \
		echo "$$VIOLATIONS"; \
		exit 1; \
	fi
	@echo "  ✓ MCP Go SDK imports confined to internal/infrastructure/mcp/"

# `vet` runs before `verify-cross-compile` for fail-fast on host-local errors;
# `verify-cross-compile` then re-covers the host target as part of the matrix.
verify: verify-no-test-sleep verify-no-network vet verify-cross-compile verify-mcp-sdk-confinement lint vulncheck
	@echo "verify: OK"
