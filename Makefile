# tellme — development tasks
#
# Build / test conventions for the tellme CLI (round 001 — narrow foundation).
# See specs/truth/techstack.md and the plan package research.md (Decisions 4 & 7).

VERSION ?= dev

STATICCHECK := $(shell command -v staticcheck 2>/dev/null)
GOLANGCI := $(shell command -v golangci-lint 2>/dev/null)
GOVULNCHECK := $(shell command -v govulncheck 2>/dev/null)

.PHONY: help build fmt vet staticcheck tidy lint vulncheck test verify verify-no-test-sleep verify-no-network

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
	@echo "  make verify               - aggregate: verify-no-test-sleep + verify-no-network + vet + lint + vulncheck"

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

test:
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

verify: verify-no-test-sleep verify-no-network vet lint vulncheck
	@echo "verify: OK"
