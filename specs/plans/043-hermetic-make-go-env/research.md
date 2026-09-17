# Phase 0 Research: a hermetic `make` Go-toolchain environment (Round 043)

Topic: give the `Makefile` a single **hermetic invocation boundary** so every Go-toolchain invocation `make` launches runs with a sanitised, explicit environment — neutralising the build-context inputs an ambient or persisted Go env can inject (`GOENV` file, `GOFLAGS`, `GO111MODULE`, `GOWORK`, ambient `GOOS`/`GOARCH`) while preserving the warm-cache and network/checksum inputs a cold-cache or proxied build needs — resolving issue [#96](https://github.com/gosharplite/tellme/issues/96) (the **outer** half of round-042's **F-2**).

Scope note: the language (`Go 1.26`), module, CLI flag layer, config layer, testing harness, provider transports, skills, MCP, presentation, and build gates were locked in rounds 001–042. The round adds **no new system end**, **no external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions are answered by the standing `techstack.md` (single CLI end; BDD techstack = `godog` on the built binary; E2E + unit strategy) and are **not re-decided** (D12). **IN**: the `Makefile` boundary, ADR 0012, the truth rows, and the falsifiability/positive-control witnesses. **OUT**: any product code, any bare-`go`-outside-`make` hermeticity, and any existing gate's semantics.

> **Lineage:** this generalises round-042's `tools/arch` `childEnv` (ADR 0011 **D5**) from the gate's own child `go list` to **every** target, and completes round-020's `verify-cross-compile` `CGO_ENABLED=0` pin (PR #46 **TD1**) from one variable/one target to a repo-wide rule. Origin: PR [#95](https://github.com/gosharplite/tellme/pull/95) fold-review #2 (comment `5721537481`, Fold 2).

---

## Decision 1: Mechanism — one top-of-`Makefile` `export`/`unexport` block (clarify Q2 → A1)

- **Decision**: define the hermetic environment **once**, near the top of the `Makefile` (above the `$(shell command -v …)` toolchain probes and before any recipe), using GNU-make `export`/`unexport`: `export GOENV := off`, `export GOWORK := off`, and `unexport GOFLAGS GO111MODULE GOEXPERIMENT GOOS GOARCH GOARM` (the precise list is D2). Every recipe — and every nested `go` invocation a scripted recipe spawns — inherits it.
- **Rationale**: the failure class is **repo-wide**, so a single owned definition is the only form that does not drift. A1 also covers the **nested** `go build` that `make test` (via the E2E harness) and `verify-no-network` spawn — which per-recipe `$(HERMETIC_ENV)` prefixes (A2) would **miss**, because those spawns happen inside a shell script, not on a recipe line. A new target added later is hermetic **by construction**.
- **Alternatives considered**:
  - **A2 — an explicit `$(HERMETIC_ENV)` prefix per invocation** — ~20 edit sites, a new recipe silently misses it, and it does not reach nested spawns — rejected.
  - **A3 — both** — redundant once A1 covers nested spawns; more surfaces to keep in sync — rejected.
- **Note on `make VAR=…` overrides**: a contributor who deliberately wants an ambient setting can export it explicitly for the invocation (`GOPROXY=… make verify`) — the documented escape hatch (D8).

## Decision 2: The neutralise set, the preserve set, and the `CGO_ENABLED` stance (clarify Q3 → V1)

- **Decision**:
  - **Neutralise/replace** — `GOENV` → `off`; `GOWORK` → `off`; `GOFLAGS`, `GO111MODULE`, `GOEXPERIMENT` → **unset**; ambient `GOOS`, `GOARCH`, `GOARM` → **unset** (host-native).
  - **Preserve** — `PATH`, `HOME`, `GOPATH`, `GOMODCACHE`, `GOCACHE` (warm-cache set) **and** `GOPROXY`, `GOSUMDB`, `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE` (network/checksum set).
  - **`CGO_ENABLED`** — left at the **host default** (not globally pinned).
- **Rationale**: the sets mirror round-042's `childEnv` discipline, extended to the preserve set the parent invocation needs. Preserving `GOPROXY`/`GOSUMDB`/… is essential: with a **cold** module cache (or behind a corporate proxy) the aggregate must still resolve modules; an `env -i`-style wipe would fail with a cache/resolution error (the `verify-cross-compile` **PR #46 TD1** lesson, in the parent).
- **Alternatives considered**:
  - **V2 — also pin `CGO_ENABLED=0` globally** — a **real behaviour change**: it would exclude any future cgo-tagged code (the recorded darwin `mach` CPU sampler is exactly such a candidate) from `make build`/`make test`, and diverge from a plain `go build`. Violates #96's own "no gate's behaviour changes for a clean environment" criterion — rejected.
  - **V3 — neutralise only the module/work vars; leave `GOOS`/`GOARCH`/`CGO_ENABLED` ambient** — a stale exported `GOOS=linux` on a macOS dev box would still miscolour `make build`/`test` (same family as the issue) — rejected.

## Decision 3: Scope — unconditional, all targets (clarify Q4 → S1)

- **Decision**: the block applies to **every** target — `build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, and `verify`. No per-target opt-out.
- **Rationale**: A1 is unconditional by nature; scoping it to a gate-family subset (S2) would require *undoing* it per target (re-exporting ambient values), reintroducing exactly the per-site drift A1 removes. And a target added later is hermetic by construction.
- **Verification folded for the widest surface (Q7-R2)**: `fmt`/`tidy` are **mutating** targets; a **positive control** (D7) confirms the block does not break them (`gofmt`/`go fmt` use no build flags; `go mod tidy` still resolves via the preserved `GOPROXY`/`GOSUMDB`/… and leaves `go.mod`/`go.sum` unchanged in a clean env).
- **Alternatives considered**: **S2 — gate/test family only** (leaves `build`/`fmt`/`tidy` ambient) — fights A1 and silently ambient-ises future targets — rejected; **S3 — S1 + an explicit `HERMETIC=0` opt-out** — a reasonable add-on, recorded as a forward option, not needed for this round.

## Decision 4: `GOENV=off` is load-bearing (not decoration)

- **Decision**: the block sets `GOENV=off`, **and** this is documented as the mechanism that makes the whole thing work.
- **Rationale**: Go resolves a variable from the **env file** whenever it is **unset or empty** in the process environment, so simply deleting or emptying `GOFLAGS` does **not** neutralise a persisted `go env -w GOFLAGS=-mod=vendor` — the exact failing case in [#96](https://github.com/gosharplite/tellme/issues/96). Round-042 review **F-2** established this for the child `go list`; this round elevates it to the parent. `GOENV=off` is the documented Go value that disables the env-file fallback entirely.
- **Alternatives considered**: **unset `GOFLAGS` only** — defeated by the env file for the persisted case (would leave the issue's headline reproduction red) — rejected.

## Decision 5: Round-042's `childEnv` is retained as defence-in-depth; primary owner is the `Makefile` (clarify Q6 → D1)

- **Decision**: keep `tools/arch/arch_test.go`'s `childEnv` filter in place; add a short cross-reference **comment** in both the `Makefile` block and the guard, naming the `Makefile` block as the primary owner and `childEnv` as the cover for the **direct** invocation of the gate. The two definitions MUST NOT drift silently.
- **Rationale**: the gate is documented (`tools/arch/doc.go`, the baseline header, the `techstack.md` row) as runnable **directly** (`go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch`), which bypasses `make`; deleting `childEnv` would regress that documented path. Doing so would also reopen frozen round-042 history. A `Makefile`↔Go shared list (D3) is a new coupling mechanism disproportionate to a handful of literals.
- **Alternatives considered**: **D2 — delete `childEnv`, single owner** — breaks the documented direct path and reopens frozen artifacts — rejected; **D3 — a shared definition the `Makefile` sources** — awkward make↔Go coupling — rejected. A cheap **drift witness** (e.g. a task/step asserting the two literal sets agree) is folded into `/axb-tasks` as an option (D7).

## Decision 6: Governance — ADR 0012 + the technology-stack truth (clarify Q5 → G1)

- **Decision**: record the policy in a new **ADR 0012** (`docs/decisions/0012-hermetic-make-go-env.md` + its `docs/decisions/README.md` index row) **and** MODIFY `specs/truth/techstack.md` (Build & Tooling): a new **Hermetic toolchain invocation** row stating the neutralise/preserve sets and the `GOENV=off` rationale, plus a note on the **`Task runner`** row that `make` runs the toolchain hermetically.
- **Rationale**: per `docs/decisions/README.md`, an ADR is for "a project-level rule … that other artifacts (or future rounds) depend on and must be able to cite" — exactly this. The ADR 0011 precedent (round 042's repo-wide build rule) is the direct analogue. `techstack.md` remains the **current-state truth**.
- **Alternatives considered**: **G2 — truth row only, no ADR** — loses the decision/rejected-options/consequences structure the repo's ADR policy expects for a repo-wide rule — rejected; **G3 — + a `SESSION-BOOTSTRAP.md` §2.2 note** — scope creep; §2.2 is the *reference's* guard set, not this repo's — rejected.

## Decision 7: Verification — four hostile-env witnesses + two positive controls

- **Decision**: the acceptance evidence is (a) **four hostile-env witnesses** — `GOENV=<file: GOFLAGS=-mod=vendor>`, `GOFLAGS=-trimpath`, `GO111MODULE=off`, a stray `GOWORK=<path>` — asserted **red before** the change and **green after** (reproduced then reverted, ADR 0010 doctrine); and (b) **two positive controls**:
  - **cross-compile**: `make verify-cross-compile` still builds + vets **4/4** targets under the hermetic env — proving the block neutralises *ambient* env **without** clobbering a recipe's intentional inline per-target `GOOS`/`GOARCH`/`CGO_ENABLED=0` (FR-012);
  - **mutating targets**: under a **clean** env, `make tidy` and `make fmt` leave `go.mod`/`go.sum` unchanged and produce no source diff (Q7-R2).
  - Optionally a **drift witness** asserting the `Makefile` block's list and `tools/arch` `childEnv`'s list agree (D5).
- **Rationale**: the two "positive controls" are what separate a *hermetic* change from a *destructive* one — they prove we removed only the ambient inputs and kept the intentional ones and the mutating targets working.
- **Alternatives considered**: **rely on `make verify` green alone** — does not exercise the failing inputs (vacuous) — rejected; **an E2E Gherkin carrier** — no user-facing CLI behaviour to express (D9) — rejected.

## Decision 8: Recorded residuals (clarify Q7 → R1 + R3)

- **Decision**: record two residuals in ADR 0012's Consequences and the truth row: **(R1)** a bare `go test ./...` / `go build` run **without `make`** stays non-hermetic — hermeticity is a **`make`-boundary** property, not a toolchain property; **(R3)** `GOENV=off` also ignores a *legitimate* operator `go env -w` setting (e.g. a deliberately pinned `GOPROXY`), so the escape hatch is an **explicitly exported** variable per invocation.
- **Rationale**: R1 bounds the claim honestly (an "all Go invocations are hermetic" claim would be false); R3 documents the accepted cost of the mechanism that is required to fix the persisted case (D4), with a concrete workaround.
- **Alternatives considered**: **preserve `GOENV` and only neutralise `GOFLAGS`** — leaves the persisted case red (D4) — rejected.

## Decision 9: Non-BDD tooling round — `/axb-spec-by-example` and the interface planners are NOOP

- **Decision**: no new system end and no change to the BDD techstack or strategy. `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey); `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP** (no persisted/runtime state — the change is build tooling); `/axb-dsl-refine` = **NOOP** (the change is a **dev surface** (`make`), not the `tellme` CLI end); `/axb-ui-plan` = **skipped**.
- **Rationale**: the round's verification is the `Makefile` invocation behaviour + falsifiability witnesses, not a Gherkin scenario; forcing CLI Gherkin risks an `acceptance-coverage` mismatch. Precedent: rounds 020/031/036/041/042.
- **Alternatives considered**: **author plan-side acceptance Gherkin for `make verify`** — no user-facing CLI behaviour to express — rejected.

## Decision 10: No new dependency; POSIX-only; `go.mod`/`go.sum` unchanged

- **Decision**: the change is `make` + the Go toolchain only; no new module; POSIX-only (GNU-make shaped, matching the repo's POSIX/bash-only stance); `go.mod`/`go.sum` unchanged.
- **Rationale**: the repo is POSIX-only by operator decision (`README.md` *Design Intent & Direction*); the `Makefile` is already GNU-make shaped.

## Decision 11: Round ordering — this is `043-*`; R2 of #92 slides to `044-*` (clarify Q8 → O1)

- **Decision**: take #96 as round `043-hermetic-make-go-env` now; R2 of [#92](https://github.com/gosharplite/tellme/issues/92) (composition-root extraction) becomes `044-*`. The `STATUS.md` roadmap line is updated at closeout.
- **Rationale**: #96 is independent and small (a `Makefile` change + ADR + truth); it is a clean hardening win to land before the larger R2 refactor; nothing in R2 is blocked by it (round 042 already delivered the gate-first precondition).
- **Alternatives considered**: **O2 — hold #96 to `044-*`** — needlessly delays a small independent fix — rejected; **O3 — fold into R2's package** — `fresh-package-per-round`; mixes a tooling change with a product refactor — rejected.

## Decision 12: Testing / BDD techstack unchanged

- **Decision**: no change to the BDD techstack (`godog` on the built binary) or the test strategy (E2E + unit); the three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are not re-decided.
- **Rationale**: the round adds no interface and no scenario; re-deciding them would be ceremony.

---

## Residual risks / forward links

- **Bare `go` outside `make`** stays non-hermetic (D8/R1) — recorded; a global shim is out of scope.
- **`GOENV=off` vs a legitimate env-file setting** (D8/R3) — the escape hatch is an explicit exported var per invocation.
- **Two definitions of the sets** (`Makefile` block + `tools/arch` `childEnv`, D5) — mitigated by cross-reference comments; an optional drift witness is folded into `/axb-tasks`.
- **A future recipe that *wants* an ambient build flag** — `make VAR=…`/an explicit export is the escape hatch; the block owns the default.
- **`CGO_ENABLED` on a cgo-less host** — an ambient `CGO_ENABLED=1` without a C compiler still breaks a *non-cross* `make build`; that is **not** one of #96's cases (the issue scopes `GOFLAGS`/`GOENV`/`GO111MODULE`/`GOWORK`), and pinning `0` globally is the rejected V2 — recorded.
- **`make help` `verify-no-network` text** — a pre-existing stale-help item (round-032 R8d), unrelated; not touched here.
- **No behaviour change** — `stdout`/`stderr`, the class-phrase vocabulary, flags, exit codes, every existing gate's semantics, and `internal/**`/`cmd/**`/`tests/**` are unchanged; the round touches the `Makefile`, `docs/decisions/**`, and `specs/truth/techstack.md`.
