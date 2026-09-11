# Technology Stack: tellme

> System truth (`specs/truth/techstack.md`) — the current, complete technology stack for tellme.
> This file always reflects the whole system, not just the latest round's delta (`techstack-complete`).
> It records decisions that are **both settled and technological**; behaviour contracts are
> spec/interface truth, and unratified proposals stay plan-side in `research.md`.

## Technology Stack Overview

### CLI Application

| Category | Technology | Purpose |
| --- | --- | --- |
| Language | `Go 1.26` | Implementation language for the `tellme` binary |
| Module path | `github.com/gosharplite/tellme` | Go module identity |
| Project layout | `cmd/tellme/` + `internal/{config,home,cli}` | Entrypoint plus single-responsibility packages, each owning one FR cluster and one E2E-observable behaviour (separation of concerns; the acceptance strategy is E2E) |
| CLI flag parsing | `spf13/pflag` | GNU-style flags (`-c/--config`, `-d`, `--version`) and usage-error classification. `--json` was **removed** (round 002) — the system **intentionally diverges** from the reference's documented `-d --json` machine-readable capability (`tell-me-go/README.md:157`); the divergence was **accepted to resolve review finding F4** (the flag silently did nothing without `-d`), trading a reference capability for a single honest diagnostic path (full rationale: `specs/plans/002-followup-cleanups/research.md`) |
| Version injection | `go build -ldflags "-X main.version=…"` | Bake the build version read by `--version`; the `version` var in `cmd/tellme/main.go` is the **only** version symbol and the single injection target |

### Configuration

| Category | Technology | Purpose |
| --- | --- | --- |
| Configuration format | YAML | Configuration input loaded from `$TELL_ME_HOME/configs/<mode>.yaml` or explicit `-c` path |
| YAML parsing | `gopkg.in/yaml.v3` | Load YAML configuration into typed structs (`Config` with tolerant root decoding; typed `Provider` entries) |
| Provider entry schema | Typed Go struct (`internal/config.Provider`) | Models complete request attributes: `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL` (expanded in round 003 from round-001 boot subset) |
| Variable expansion | Hand-crafted stdlib regex (`internal/config/expand.go`) | Deterministic `${VAR}` and `${VAR:-default}` substitution in `API_KEY`, `URL`, and `HEADERS` values using `os.LookupEnv` with zero third-party dependencies; fails on unset variables |
| Effective-value resolution & validation | Hand-written Go (no framework) | Apply `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER` precedence, expand active provider variables, and validate required fields/bounds with frozen class phrase `tellme: the provider configuration is invalid` |

### Testing & Verification

| Category | Technology | Purpose |
| --- | --- | --- |
| CLI BDD techstack | `godog` (Cucumber for Go) | Run the executable **interface** Gherkin (`specs/truth/features/**`) via step definitions — E2E against the built binary. Instantiated in round 001 by the `tests/e2e/` suite |
| E2E runner / step definitions | Go test package at `tests/e2e/` (`godog.TestSuite`) | Loads `specs/truth/features/**`; builds the binary once per suite; per-scenario fresh `TELL_ME_HOME`; exit-code + stdout + stderr + filesystem assertions |
| Test strategy | E2E (black-box) for the acceptance path; fast unit tests for pure helpers | The acceptance path invokes the built `tellme` binary as a subprocess under a controlled environment and asserts exit code, stdout, stderr, and workspace effects; the pure resolution helpers are covered by fast, isolated unit tests |
| No-network verification | No-egress sandbox (privileged netns **or** unprivileged hostile-env fallback) + build-graph capability guard | Prove SC-004 (zero network on `-d`, incl. the failure path): the sandbox is a differential witness (assert identical output with egress blocked — `unshare -n` where permitted, else a hostile DNS/proxy env; the privileged netns is **not** available on the local dev host); the capability guard asserts **network-capability absence**, not bare package presence — no `net/http` in the binary's dependency closure and no dialing/listening symbol in the linked binary (`go tool nm`). Bare `net` / `net/netip` are linked transitively by `spf13/pflag` IP-flag parsing without performing I/O and are **not** indicators |
| Version assertion | `VERSION=0.0.0-harness` sentinel | The suite builds with a sentinel and asserts the exact `--version` string (single target `main.version`), making FR-010 falsifiable instead of passing against the `dev` default |
| Host test harness | Go stdlib `testing` | Runs the godog suites and any supporting assertions; determinism — no `time.Sleep` for synchronization (ADR-036 parity) |
| Pure-helper unit tests | Go stdlib `testing` (table-driven) | Fast, isolated, offline tests for pure helpers — effective mode, effective provider, workspace idempotency (round 001/002 F9), `${VAR}` expansion, and provider validation (round 003) — complementing the E2E acceptance path |

### Build & Tooling

| Category | Technology | Purpose |
| --- | --- | --- |
| Build | `go build` (Makefile `build`) | Produce the `tellme` binary with the injected version |
| Formatting | `gofmt` / `go fmt` | Canonical source formatting |
| Static check | `go vet` | Minimum correctness gate (toolchain-native) |
| Static analysis | `staticcheck` | Secondary static gate (`unused` + SA/S classes) — direct analyzer, no policy artifact; the Makefile resolves it explicitly (`command -v` + `$GOPATH/bin` fallback) |
| Module hygiene | `go mod tidy` | Keep `go.mod` / `go.sum` consistent |
| Task runner | `make` | `fmt`, `tidy`, `build`, `test`, `lint`, `vulncheck`, `verify` |
| Lint aggregator | `golangci-lint` (with `errcheck`) | Multi-linter gate; `errcheck` closes the round-001 unchecked-error residual. Policy artifact: a committed `.golangci.yml` |
| Dependency vulnerability scan | `govulncheck` | Fails the verification on a known vulnerability in dependencies |

## Adopted, Not Yet Instantiated

*(none — `godog` was instantiated in round 001: the `tests/e2e/` suite executes the interface Gherkin
under `specs/truth/features/**`.)*

## Not Introduced Yet

- `spf13/cobra` (subcommand framework) — deferred until subcommands (`browse`, `retry`) exist
- `spf13/viper` (config framework with env binding / watchers)
- `testify` (assertion / mock library)
- Provider SDKs and HTTP clients (Google/Vertex, OpenAI, DeepSeek, Anthropic, Moonshot, Z.AI)
- TUI libraries (Bubble Tea, Lipgloss, Glamour)
- MCP client SDK
- SQLite / history persistence
- Web frontend and HTTP server — tellme has a single CLI end
