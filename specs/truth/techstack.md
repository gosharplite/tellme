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
| Project layout | `cmd/tellme/` + `internal/{config,home,cli}` | Entrypoint plus single-responsibility packages, each owning one FR cluster and one E2E-observable behaviour (separation of concerns — not unit-testability; this round's strategy is E2E) |
| CLI flag parsing | `spf13/pflag` | GNU-style flags (`-c/--config`, `-d`, `--json`, `--version`) and usage-error classification |
| Version injection | `go build -ldflags "-X main.version=…"` | Bake the build version read by `--version`; the `version` var in `cmd/tellme/main.go` is the **only** version symbol and the single injection target |

### Configuration

| Category | Technology | Purpose |
| --- | --- | --- |
| Configuration format | YAML | Slice-local configuration input (`$TELL_ME_HOME/configs/<mode>.yaml`) |
| YAML parsing | `gopkg.in/yaml.v3` | Load the YAML configuration into a typed struct |
| Effective-value resolution | Hand-written Go (no framework) | Apply `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER` precedence over the file, and validate the selected provider (resolver algorithm lives in plan-side `research.md` Decision 3) |

### Testing & Verification

| Category | Technology | Purpose |
| --- | --- | --- |
| CLI BDD techstack | `godog` (Cucumber for Go) | Run the executable **interface** Gherkin (`specs/truth/features/**`) via step definitions — E2E against the built binary. **Adopted this round; first instantiated in `/axb-dsl-refine`** (see "Adopted, Not Yet Instantiated") |
| E2E runner / step definitions | Go test package at `tests/e2e/` (`godog.TestSuite`) | Loads `specs/truth/features/**`; builds the binary once per suite; per-scenario fresh `TELL_ME_HOME`; exit-code + stdout + stderr + filesystem assertions |
| Test strategy | E2E (black-box) | Invoke the built `tellme` binary as a subprocess under a controlled environment; assert exit code, stdout, stderr, and workspace effects |
| No-network verification | No-egress sandbox (privileged netns **or** unprivileged hostile-env fallback) + build-graph capability guard | Prove SC-004 (zero network on `-d`, incl. the failure path): the sandbox is a differential witness (assert identical output with egress blocked — `unshare -n` where permitted, else a hostile DNS/proxy env; the privileged netns is **not** available on the local dev host); the capability guard asserts **network-capability absence**, not bare package presence — no `net/http` in the binary's dependency closure and no dialing/listening symbol in the linked binary (`go tool nm`). Bare `net` / `net/netip` are linked transitively by `spf13/pflag` IP-flag parsing without performing I/O and are **not** indicators |
| Version assertion | `VERSION=0.0.0-harness` sentinel | The suite builds with a sentinel and asserts the exact `--version` string (single target `main.version`), making FR-010 falsifiable instead of passing against the `dev` default |
| Host test harness | Go stdlib `testing` | Runs the godog suites and any supporting assertions; determinism — no `time.Sleep` for synchronization (ADR-036 parity) |

### Build & Tooling

| Category | Technology | Purpose |
| --- | --- | --- |
| Build | `go build` (Makefile `build`) | Produce the `tellme` binary with the injected version |
| Formatting | `gofmt` / `go fmt` | Canonical source formatting |
| Static check | `go vet` | Minimum correctness gate (toolchain-native) |
| Static analysis | `staticcheck` | Secondary static gate (`unused` + SA/S classes) — direct analyzer, no policy artifact; the Makefile resolves it explicitly (`command -v` + `$GOPATH/bin` fallback) |
| Module hygiene | `go mod tidy` | Keep `go.mod` / `go.sum` consistent |
| Task runner | `make` | `fmt`, `tidy`, `build`, `test`, `verify` |

## Adopted, Not Yet Instantiated

- **`godog`** — the CLI BDD runner is a committed stack choice, but it executes **nothing** until
  `/axb-dsl-refine` produces the interface features + `DSLRow`s + step definitions (there is no
  `go.mod`, no Go code, and `specs/truth/features/**` does not exist yet). Listed here so a reader never
  reads an unexercised runner as exercised. (`techstack-complete` still holds: the current stack is
  complete *and* correctly labelled.)

## Not Introduced Yet

- `spf13/cobra` (subcommand framework) — deferred until subcommands (`browse`, `retry`) exist
- `spf13/viper` (config framework with env binding / watchers)
- `golangci-lint` (multi-linter pipeline; the intended next-slice aggregator) — deferred pending its
  curated `.golangci.yml` policy artifact, **not** for install cost (present in `$GOPATH/bin`)
- `govulncheck` (dependency vulnerability scan) — deferred pending a triage/policy posture, **not** for
  install cost
- `testify` (assertion / mock library)
- Provider SDKs and HTTP clients (Google/Vertex, OpenAI, DeepSeek, Anthropic, Moonshot, Z.AI)
- TUI libraries (Bubble Tea, Lipgloss, Glamour)
- MCP client SDK
- SQLite / history persistence
- Web frontend and HTTP server — tellme has a single CLI end
