# tellme — development tasks
#
# Build / test conventions for the tellme CLI (round 001 — narrow foundation).
# See specs/truth/techstack.md and the plan package research.md (Decisions 4 & 7).

VERSION ?= dev

STATICCHECK := $(shell command -v staticcheck 2>/dev/null)

.PHONY: help build fmt vet staticcheck tidy test verify verify-no-test-sleep verify-no-network

help:
	@echo "tellme development tasks:"
	@echo "  make build                - Build ./cmd/tellme with -X main.version=\$$(VERSION)"
	@echo "  make fmt                  - go fmt ./..."
	@echo "  make vet                  - go vet ./..."
	@echo "  make staticcheck          - run staticcheck ./... (resolved from PATH)"
	@echo "  make tidy                 - go mod tidy"
	@echo "  make test                 - go test ./..."
	@echo "  make verify-no-test-sleep - forbid time.Sleep for synchronization in *_test.go (ADR-036 parity)"
	@echo "  make verify-no-network    - build-graph capability guard: no net/net/http in ./cmd/tellme closure"
	@echo "  make verify               - aggregate: verify-no-test-sleep + verify-no-network + vet"

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

# Build-graph capability guard (research.md Decision 5, backstop witness):
# the diagnostic binary must have no network capability — no net/http package in
# its dependency closure, and no dialing/listening symbol in the linked binary.
# Bare `net` is intentionally NOT flagged: spf13/pflag (the mandated flag library)
# links `net` for IP flag parsing but never performs I/O; the symbol check covers
# real capability.
verify-no-network:
	@echo "verify-no-network: build-graph capability guard ..."
	@if ! go list ./cmd/tellme >/dev/null 2>&1; then \
		echo "  (skip) ./cmd/tellme package not present yet"; exit 0; \
	fi; \
	if go list -deps ./cmd/tellme | grep -qx 'net/http'; then \
		echo "❌ net/http is linked into ./cmd/tellme"; exit 1; \
	fi; \
	tmpdir="$$(mktemp -d)"; \
	if ! go build -o "$$tmpdir/tellme" ./cmd/tellme; then \
		echo "❌ could not build ./cmd/tellme for the capability check"; exit 1; \
	fi; \
	if go tool nm "$$tmpdir/tellme" 2>/dev/null | grep -Eq 'net\.Dial|net\.\(\*Dialer\)\.Dial|net\.Listen|net\.ListenPacket|net\.LookupHost|net\.LookupIP|net\.LookupAddr|net\.Resolve|net/http\.|crypto/tls\.\(\*Conn\)\.Handshake'; then \
		echo "❌ network dialing/listening symbols present in ./cmd/tellme:"; \
		go tool nm "$$tmpdir/tellme" | grep -E 'net\.Dial|net/http\.' | head; \
		exit 1; \
	fi; \
	echo "  ✓ no network capability in ./cmd/tellme"

verify: verify-no-test-sleep verify-no-network vet
	@echo "verify: OK"
