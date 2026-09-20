# Tasks — the `ToolSetSpec` capability seam (round 069)

**Plan Package**: `specs/plans/069-toolset-spec-capability-seam`
**Inputs**: `spec.md` · `research.md` (D1–D7) · `plan.md` (§W1–W4) · `truth-delta.md`
**Anchor**: issue [#140](https://github.com/gosharplite/tellme/issues/140)

**Scope reminder**: a **pure internal-shape refactor**. No config change, no UX change, no truth behaviour change. The "RED" state is the **compile break** the signature change causes; the "GREEN" state is the existing suite passing **with no assertion changed**.

---

## Phase 1 — Setup

_None._ (No new dependency, no config key, no new file beyond the one named type.)

## Phase 2 — Foundational

_None._

## Phase 3 — Test Alignment & Implementation

- [x] **T001** [CODE] Declare `deps.ToolSetSpec` (`Sink domaintools.OutputSink`, `Vision bool`, `ProviderType string`) in `internal/app/deps/deps.go`; change the `Dependencies.NewToolRegistry` field to `func(spec ToolSetSpec) domaintools.Registry` (+ the D5 invariant in the doc comment).
- [x] **T002** [CODE] `cmd/tellme/deps.go`: `assembleAgentTools(spec deps.ToolSetSpec)` resolves the ceiling **inside** the `if spec.Vision` branch; `newToolRegistry(spec deps.ToolSetSpec)` is a thin wrapper; `agentTools()` passes `deps.ToolSetSpec{}`. **No positional scalar remains.**
- [x] **T003** [CODE] `internal/cli/cli.go`: the prompt-path call site constructs `deps.ToolSetSpec{Sink: prog.ToolOutput, Vision: res.Provider.Vision, ProviderType: res.Provider.Type}`; `renderToolUsage` takes the **named** signature `func(deps.ToolSetSpec) domaintools.Registry` and builds its two union variants as `deps.ToolSetSpec{}` / `deps.ToolSetSpec{Vision: true}`.
- [x] **T004** [TEST ALIGN] Update the test call sites (a compile-driven change, no assertion edits): `cmd/tellme/deps_test.go` (3 sites → named-field constructions), `internal/cli/testdeps_test.go` + `internal/cli/cli_test.go` (the injected registry func type).
- [x] **T005** [TRUTH] `specs/truth/techstack.md` MODIFY ×2 (*Composition root* + *Image filesystem tool*) + `docs/decisions/0039-toolset-spec-capability-seam.md` + the index row.

## Phase 4 — Verification & Regression

- [x] **T006** [GATES] `gofmt`/`go vet`/`go build` clean; `go test -count=1 ./...` green (incl. the godog E2E); `make verify` **OK**; `go.mod`/`go.sum` unchanged.
- [x] **T007** [WITNESS] Structural witness (no positional capability scalar on the path — grep) + falsifiability witnesses (reproduced then reverted).

---

## Implementation ledger

### Structural witness (SC-001) — no positional capability scalar remains

```
$ grep -rn 'func(domaintools.OutputSink, bool, string)\|NewToolRegistry(nil\|newToolRegistry(nil' --include=*.go .
(no matches)
```

The construction path is now `NewToolRegistry(spec deps.ToolSetSpec)` → `newToolRegistry(spec)` → `assembleAgentTools(spec)`; the offline `renderToolUsage` takes `func(deps.ToolSetSpec) domaintools.Registry`. Every call site names fields:

- `internal/cli/cli.go` — `deps.ToolSetSpec{Sink: prog.ToolOutput, Vision: res.Provider.Vision, ProviderType: res.Provider.Type}`
- `internal/cli/cli.go` (`renderToolUsage`) — `deps.ToolSetSpec{}` / `deps.ToolSetSpec{Vision: true}`
- `cmd/tellme/deps.go` (`agentTools`) — `deps.ToolSetSpec{}`

### Behaviour-identity (SC-002) — no assertion changed

The only test edits are **signature/constructor** updates (T004); **no assertion changed**. The full suite (unit + the godog E2E, incl. `offering-the-agent-tools` / `reading-a-local-image` / `accounting-for-the-tool-use` / the offline `--tool-usage` pin) is **green**.

### Falsifiability witnesses (reproduced then reverted)

- **(a) un-gate vision** — in `assembleAgentTools`, replace `if spec.Vision {` with `if true {`: the existing offered-set pins red (`cmd/tellme` `TestNewToolRegistryOffersAgentTools` — a non-vision provider must NOT offer `read_image`; and the E2E `offering-the-agent-tools` journey). Reverted.
- **(b) wrong ceiling** — in `assembleAgentTools`, hard-code the resolved ceiling to `1` (any real image exceeds it): the E2E `reading-a-local-image` journey reds — `the picture's kind comes from its content, not its name` (the request carries no image block) and `a file that is not a picture is refused, whatever it is called` (the refusal reads *"too large … exceeds the 0 MiB limit"* instead of not-a-picture). **Reproduced then reverted** — the ceiling **wiring** is carried by the E2E journey (there is no unit pin on the wiring itself; the unit pins cover `ImageCeilingForFamily`'s table, not its consumption — a pre-existing gap, recorded as **RF-069-5**).

### Gates (T006)

```
gofmt -l .                      (clean)
go vet ./...                    (clean)
go build ./...                  (clean)
go test -count=1 ./...          ok (all packages incl. tests/e2e)
make verify                     verify: OK
git diff --stat go.mod go.sum   (unchanged)
```
