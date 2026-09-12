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
| Project layout | `cmd/tellme/` + `internal/{config,home,cli}` + `internal/domain` (port) & `internal/infrastructure` (adapter) | Entrypoint plus single-responsibility packages, each owning one FR cluster and one E2E-observable behaviour (separation of concerns; the acceptance strategy is E2E). Round 004 adds the layered seam: a network-free domain port and an infrastructure adapter |
| CLI flag parsing | `spf13/pflag` | GNU-style flags (`-c/--config`, `-d`, `--version`) and usage-error classification. `--json` was **removed** (round 002) — the system **intentionally diverges** from the reference's documented `-d --json` machine-readable capability (`tell-me-go/README.md:157`); the divergence was **accepted to resolve review finding F4** (the flag silently did nothing without `-d`), trading a reference capability for a single honest diagnostic path (full rationale: `specs/plans/002-followup-cleanups/research.md`) |
| Version injection | `go build -ldflags "-X main.version=…"` | Bake the build version read by `--version`; the `version` var in `cmd/tellme/main.go` is the **only** version symbol and the single injection target |
| Prompt input | `spf13/pflag` positional argument(s) + standard input | The operator's prompt is the positional argument(s) and/or piped standard input. Round 005 combines them — `args` (joined by single spaces) + `"\n"` + piped stdin — then trims, matching the reference's main chat path; a positional-only prompt is unchanged from round 004. Dispatch precedence stays `--version` → `-d` → prompt turn → boot; standard input is read only on the prompt-turn path |
| Terminal detection | Go stdlib `os.File` + `os.ModeCharDevice` | Detect, separately for stdin and stdout, whether a stream is a terminal, behind an injected seam (dependency-free — no `golang.org/x/term`). Gates whether stdin is read (round 005) and future presentation suppression (FR-007) |
| Piped stdin read | Go stdlib `io.LimitReader` | Read piped/redirected standard input bounded by a fixed 1 MiB cap (matching the reference's `maxPromptSize`); content beyond the cap is not read |

### Configuration

| Category | Technology | Purpose |
| --- | --- | --- |
| Configuration format | YAML | Configuration input loaded from `$TELL_ME_HOME/configs/<mode>.yaml` or explicit `-c` path |
| YAML parsing | `gopkg.in/yaml.v3` | Load YAML configuration into typed structs (`Config` with tolerant root decoding; typed `Provider` entries) |
| Provider entry schema | Typed Go struct (`internal/config.Provider`) | Models complete request attributes: `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL` (expanded in round 003 from round-001 boot subset) |
| Variable expansion | Hand-crafted stdlib regex (`internal/config/expand.go`) | Deterministic `${VAR}` and `${VAR:-default}` substitution in `API_KEY`, `URL`, and `HEADERS` values with zero third-party dependencies; fails on unset variables. The environment lookup is injected via an `EnvLookupFunc` port defaulting to `os.LookupEnv` (dependency inversion, review finding #1, so pure-helper unit tests are in-memory and parallel-safe) |
| Effective-value resolution & validation | Hand-written Go (no framework) | Apply `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER` precedence, **expand the active provider's variables before validating** the resolved entry (so a mandatory field whose placeholder resolves to empty is rejected — review finding #2), validate required fields/bounds with frozen class phrase `tellme: the provider configuration is invalid`, and carry the resolved provider on the resolution outcome (now consumed by the round-004 reasoning path) |

### Reasoning & Provider Transport

| Category | Technology | Purpose |
| --- | --- | --- |
| Provider gateway port | `internal/domain/llm` (`Gateway` interface + `Request`/`Response`/`ProviderError` value types) | Domain-facing, network-free provider abstraction; single method `Complete(ctx, Request) (Response, error)` so the transport is swappable and the turn is fake-testable |
| OpenAI-compatible adapter | `internal/infrastructure/llm/openai` | The first concrete provider transport — sends the OpenAI-compatible Chat Completions request and normalizes the response |
| HTTP transport | Go stdlib `net/http` (+ `encoding/json`) | Provider request/response over JSON; no provider SDK (round-004 Decision 2) |
| Request assembly | Hand-written mapping from the resolved `config.Provider` | Builds endpoint (`<URL>/chat/completions`), auth (`API_KEY` → `Authorization: Bearer`), `MODEL`, `MAX_TOKENS`, merged `HEADERS`, and `reasoning_effort` from the resolved provider entry |
| Response normalization | Hand-written extraction of `choices[0].message.content` | Reduces the provider response to the minimal answer text printed to stdout (full provider-agnostic `Thought` model deferred) |

### Testing & Verification

| Category | Technology | Purpose |
| --- | --- | --- |
| CLI BDD techstack | `godog` (Cucumber for Go) | Run the executable **interface** Gherkin (`specs/truth/features/**`) via step definitions — E2E against the built binary. Instantiated in round 001 by the `tests/e2e/` suite |
| E2E runner / step definitions | Go test package at `tests/e2e/` (`godog.TestSuite`) | Loads `specs/truth/features/**`; builds the binary once per suite; per-scenario fresh `TELL_ME_HOME`; exit-code + stdout + stderr + filesystem assertions; round 005 extends the subprocess runner to inject a scripted stdin (a pipe) for the piped-prompt scenarios |
| Test strategy | E2E (black-box) for the acceptance path; fast unit tests for pure helpers | The acceptance path invokes the built `tellme` binary as a subprocess under a controlled environment and asserts exit code, stdout, stderr, and workspace effects; the pure resolution helpers are covered by fast, isolated unit tests |
| Local fake provider | `net/http/httptest` (in-process server) | Serves the network-path acceptance scenarios: the scenario's provider `URL` points at the fake; assertions cover the printed answer, the frozen failure class, and the fake's recorded request (round 004) |
| No-network verification (offline paths) | No-dial canary (recording sink) + differential no-egress sandbox | Prove the **offline paths** (`--version`, `-d` incl. its failure path, and no-prompt boot) never reach the network: (i) the offline paths must leave a recording sink — pointed at by the configured provider `URL` — with **zero** connections; (ii) the differential sandbox asserts byte-identical output with egress blocked (`unshare -n` where permitted, else a hostile DNS/proxy env; the privileged netns is **not** available on the local dev host). Round 004 **retired** the whole-binary build-graph capability guard ("no `net/http` in the closure" + no dialing symbol) because the prompt-bearing chat path now legitimately links `net/http`; the offline guarantee is now an offline-path behaviour claim, not a whole-binary absence claim |
| Version assertion | `VERSION=0.0.0-harness` sentinel | The suite builds with a sentinel and asserts the exact `--version` string (single target `main.version`), making FR-010 falsifiable instead of passing against the `dev` default |
| Host test harness | Go stdlib `testing` | Runs the godog suites and any supporting assertions; determinism — no `time.Sleep` for synchronization (ADR-036 parity) |
| Pure-helper unit tests | Go stdlib `testing` (table-driven) | Fast, isolated, offline tests for pure helpers — effective mode, effective provider, workspace idempotency (round 001/002 F9), `${VAR}` expansion, provider validation (round 003), request assembly + response normalization (round 004), and prompt combination + input/output-mode selection (round 005, folding issue #14) — complementing the E2E acceptance path |

### Build & Tooling

| Category | Technology | Purpose |
| --- | --- | --- |
| Build | `go build` (Makefile `build`) | Produce the `tellme` binary with the injected version |
| Formatting | `gofmt` / `go fmt` | Canonical source formatting |
| Static check | `go vet` | Minimum correctness gate (toolchain-native) |
| Static analysis | `staticcheck` | Secondary static gate (`unused` + SA/S classes) — direct analyzer, no policy artifact; the Makefile resolves it explicitly (`command -v` + `$GOPATH/bin` fallback) |
| Module hygiene | `go mod tidy` | Keep `go.mod` / `go.sum` consistent |
| Task runner | `make` | `fmt`, `tidy`, `build`, `test`, `lint`, `vulncheck`, `verify` |
| Lint aggregator | `golangci-lint` (with `errcheck` + `cyclop`) | Multi-linter gate; `errcheck` closes the round-001 unchecked-error residual; `cyclop` enforces a `max-complexity: 15` guard against complexity creep (round-004 PR #12 review follow-up). Policy artifact: a committed `.golangci.yml` |
| Dependency vulnerability scan | `govulncheck` | Fails the verification on a known vulnerability in dependencies |

## Adopted, Not Yet Instantiated

*(none — `godog` was instantiated in round 001: the `tests/e2e/` suite executes the interface Gherkin
under `specs/truth/features/**`.)*

## Not Introduced Yet

- `spf13/cobra` (subcommand framework) — deferred until subcommands (`browse`, `retry`) exist
- `spf13/viper` (config framework with env binding / watchers)
- `testify` (assertion / mock library)
- Provider SDKs (Google/Vertex, OpenAI, DeepSeek, Anthropic, Moonshot, Z.AI) — the provider transport uses stdlib `net/http`, not an SDK
- Gemini/Vertex and Anthropic provider adapters (only the OpenAI-compatible family ships this round)
- Streaming (SSE) response handling
- The full provider-agnostic `Thought` model and tool-call shapes
- `golang.org/x/term` — TTY detection uses a dependency-free stdlib char-device check instead (round 005)
- A Markdown/ANSI renderer and the `-r`/raw-output flag — tellme's output is plain and already equals the reference's `-r` output (round 005, Clarify Q1); rendered-output parity is a future slice
- TUI libraries (Bubble Tea, Lipgloss, Glamour)
- MCP client SDK
- SQLite / history persistence
- Web frontend and HTTP server — tellme has a single CLI end
