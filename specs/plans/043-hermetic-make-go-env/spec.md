# Feature Specification: a hermetic `make` Go-toolchain environment (round 043)

**Feature Branch**: `043-hermetic-make-go-env`

**Created**: 2026-09-18

**Status**: Draft — plan half only (specify → spec-by-example **NOOP** → technical-research → system-analysis). Anchor issue [#96](https://github.com/gosharplite/tellme/issues/96) — the **outer** half of round-042's **F-2**. Clarify round 1 locked Q1–Q8 (below).

**Input**: Operator request to resolve [#96](https://github.com/gosharplite/tellme/issues/96) via the AIxBDD process. #96: *every gate in the `Makefile` invokes the Go toolchain as a child of `make`, inheriting the ambient environment; a **persisted** (`go env -w`) or exported Go setting therefore reaches the **outer** `go test`/`go vet`/`go build`, so a developer/CI image that exports `GOFLAGS=-mod=vendor` / `-trimpath`, a `GOENV` file, `GO111MODULE=off`, or a stray `GOWORK` can redden the whole `verify` aggregate for reasons unrelated to the tree.*

**Behaviour intent**: **ADD** a single **hermetic invocation boundary** to the `Makefile` so that **every** Go-toolchain invocation `make` launches runs with a **sanitised, explicit** environment — neutralising the build-context inputs an ambient/persisted Go env can inject, while preserving the warm-cache and network/checksum inputs a cold-cache or proxied build legitimately needs. This **generalises** round-042's `childEnv` discipline (ADR 0011 **D5**) from the gate's own child `go list` to **every** target, and completes round-020's `verify-cross-compile` `CGO_ENABLED=0` pin (PR #46 **TD1**) from one variable/one target to a repo-wide rule. It is a **tooling/truth slice** in the round-020 / round-032 / round-042 lineage: it changes **no** user-facing behaviour and **no** product code — it changes the `Makefile`, adds an ADR, and updates the technology-stack truth.

---

## Locked decisions (clarify round 1 — Q1–Q8)

| # | Decision |
| --- | --- |
| **Q1** | **Option A — neutralise at the invocation boundary.** Actually fix the class (repo-wide, once); rejected B (document + loud guard: does not make the gates hermetic) and C (non-goal). |
| **Q2** | **A1 — one top-of-`Makefile` `export`/`unexport` block.** Single owned definition; covers **every** recipe *and* the nested `go build` a scripted recipe (`make test`, `verify-no-network`) spawns — which per-recipe prefixes (A2) would miss. |
| **Q3** | **V1 — the neutralise/preserve sets below; `CGO_ENABLED` left at the host default.** *Neutralise/replace*: `GOENV` (→ `off`), `GOFLAGS`, `GO111MODULE`, `GOEXPERIMENT` (unset), `GOWORK` (→ `off`), ambient `GOOS`/`GOARCH`/`GOARM` (unset). *Preserve*: `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE` **and** `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE`. **Not** V2 (global `CGO_ENABLED=0`: a real behaviour change — it would exclude future cgo-tagged code, e.g. the recorded darwin `mach` CPU sampler, and diverge from a plain `go build`); **not** V3 (leave `GOOS`/`GOARCH`/`CGO_ENABLED` ambient — a stale exported `GOOS` on a dev box would still miscolour `make build`/`test`). |
| **Q4** | **S1 — the block is unconditional; ALL targets hermetic** (`build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, `verify`). Rejected S2 (gate-family only: fights the A1 mechanism and reintroduces per-site drift). |
| **Q5** | **G1 — a new ADR (`0012`) + its index row, and a `specs/truth/techstack.md` (Build & Tooling) update.** The rule is project-level and future rounds/artifacts must cite it (the ADR 0011 precedent). |
| **Q6** | **D1 — keep round-042's `tools/arch` `childEnv` as defence-in-depth.** It covers the **documented direct invocation** of the gate (`go test -count=1 -tags=arch … ./tools/arch`) which bypasses `make`; the `Makefile` block is the **primary** owner. Cross-reference comments both ways. Rejected D2 (delete `childEnv`: breaks the documented direct path) and D3 (a `Makefile`↔Go shared list: a new coupling mechanism for a handful of literals). |
| **Q7** | **R1 recorded** (a bare `go test ./...` / `go build` *without* `make` stays non-hermetic — a recorded residual) · **R2 folded into verification** (`fmt`/`tidy` are mutating targets; a positive control confirms the block does not break them and `go.mod`/`go.sum` stay unchanged) · **R3 recorded** (`GOENV=off` also ignores a *legitimate* operator `go env -w`; an explicit exported var per invocation is the documented escape hatch). |
| **Q8** | **O1 — this is round `043-*`; R2 of [#92](https://github.com/gosharplite/tellme/issues/92) slides to `044-*`** (the `STATUS.md` roadmap line is updated at closeout). Rejected folding #96 into R2's package (`fresh-package-per-round`; two unrelated themes). |

---

## Grounded in the current system

- **Every `Makefile` recipe that invokes the Go toolchain inherits `make`'s environment.** `grep -nE '\bgo (build|vet|test|fmt|mod|list|run)\b|\$\(GOLANGCI\)|\$\(STATICCHECK\)|\$\(GOVULNCHECK\)' Makefile` resolves to **~20 sites** across `build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, the `verify-*` family, and `verify`. None of them sanitises the environment.
- **The failure is reproducible today** (#96, verified for `verify-architecture`):
  ```sh
  $ printf 'GOFLAGS=-mod=vendor\n' > /tmp/goenv
  $ GOENV=/tmp/goenv make verify-architecture
  … inconsistent vendoring …        # exit 2, unrelated to the tree
  ```
- **Round 042 already sanitises its *own* child** (`tools/arch/arch_test.go` `childEnv`: drops `GOOS`/`GOARCH`/`GOARM`/`CGO_ENABLED`/`GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOWORK`, then re-adds explicit non-empty `GOFLAGS=-mod=readonly`, `GO111MODULE=on`, `GOWORK=off`). The gate's **verdict** is hermetic; the surrounding `go test`/`go vet` that `make` runs are **not**.
- **Round 020 pinned exactly one variable for one target** (`verify-cross-compile`: an inline `CGO_ENABLED=0 GOOS=… GOARCH=… go build`; PR #46 review **TD1**) so an ambient `CGO_ENABLED=1` cannot break a cross build. #96 generalises the concern.
- **`GOENV=off` is load-bearing, not decoration.** Deleting or emptying `GOFLAGS` in the recipe environment does **not** neutralise a *persisted* `go env -w GOFLAGS=-mod=vendor`: Go falls back to the env **file** for a variable that is unset **or empty** (round-042 review **F-2**, ADR 0011 D5). `GOENV=off` is the documented Go mechanism that disables the env-file fallback entirely.
- **Two existing gate shapes** model the round: the **Makefile grep** (`verify-mcp-sdk-confinement`) and the **make → `go test -run …`** delegation (`verify-no-network`). The round adds **no** target — it changes how the *existing* targets launch the toolchain.
- **`specs/truth/techstack.md` (Build & Tooling)** already records the `Task runner` (`make`) and the `verify` aggregate, so this is a real truth **MODIFY** (not a NOOP).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A hostile ambient Go env cannot redden any `make` target (Priority: P1)

As a maintainer/CI operator, I want every `make` target to run the Go toolchain with a sanitised, explicit environment, so an exported or persisted Go setting in my shell/CI image (`GOFLAGS=-mod=vendor`, `GOFLAGS=-trimpath`, `GO111MODULE=off`, a `GOENV` file, a stray `GOWORK`, a stale `GOOS`) cannot make `make verify` fail for reasons unrelated to the tree.

**Why this priority**: it is the round's entire reason to exist; today **no** target sanitises the ambient environment.

**Independent verification**: with a hostile ambient Go env set, run the four #96 cases (`GOENV=<file: GOFLAGS=-mod=vendor>`, `GOFLAGS=-trimpath`, `GO111MODULE=off`, a stray `GOWORK`) — before the change they **fail**; after, they are **green**; revert the block and they fail again (falsifiability).

**Acceptance Scenarios**:

1. **Given** `GOENV` pointing at a file containing `GOFLAGS=-mod=vendor`, **When** `make verify` runs, **Then** it is **green** (previously exit 2).
2. **Given** each of `GOFLAGS=-trimpath`, `GO111MODULE=off`, and a stray `GOWORK=<path>`, **When** a toolchain target runs, **Then** the ambient value is **neutralised** and the target behaves as under a clean environment.
3. **Given** the aggregate `make verify`, **When** it runs under any of the above hostile envs, **Then** it exits 0 and every member gate's verdict matches its clean-env verdict.

**Functional Requirements**:

- **FR-001**: Every `make` target that invokes the Go toolchain (`build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, `verify`) MUST run it with a **hermetic environment**: the **neutralise set** neutralised, the **preserve set** preserved.
- **FR-002 (neutralise set)**: the environment `make` exports to its recipes MUST neutralise — with an explicit value where an unset/empty value would be defeated by the env-file fallback — `GOENV` (set to `off`), `GOFLAGS`, `GO111MODULE`, `GOEXPERIMENT` (unset), `GOWORK` (set to `off`), and the ambient build-context `GOOS`/`GOARCH`/`GOARM` (unset ⇒ host-native).
- **FR-003 (preserve set)**: `PATH`, `HOME`, `GOPATH`, `GOMODCACHE`, `GOCACHE` (the warm-module-cache set) and `GOPROXY`, `GOSUMDB`, `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE` (the network/checksum set) MUST be preserved so a cold-cache or proxied build still resolves.
- **FR-004**: `CGO_ENABLED` MUST be left at the **host default**; the only cgo pinning remains `verify-cross-compile`'s inline `CGO_ENABLED=0` **for its own targets** (PR #46 TD1), which MUST keep working.
- **FR-005**: The neutralisation MUST be defined **once**, at the invocation boundary (a single top-of-`Makefile` block), not duplicated per recipe — so a target added later is hermetic **by construction**, and a per-recipe inline override (e.g. `verify-cross-compile`'s `GOOS=…`) still wins for that recipe.
- **FR-006**: For a **clean** environment, **no** target's observable behaviour changes; **no** gate's semantics change; `go.mod`/`go.sum` stay unchanged; no new dependency is added.
- **FR-007**: The rule MUST be recorded in a new **ADR 0012** (`docs/decisions/0012-hermetic-make-go-env.md`) with its `docs/decisions/README.md` index row, and in `specs/truth/techstack.md` (Build & Tooling) as a citable row (with the `Task runner` row noting the hermetic environment).
- **FR-008**: Round-042's `tools/arch` `childEnv` filter MUST be **retained** as defence-in-depth (it covers the documented **direct** invocation of the gate, which bypasses `make`), with a cross-reference comment in both the `Makefile` block and the guard. The two definitions MUST NOT be allowed to drift silently.
- **FR-009**: The two residuals MUST be recorded (in ADR 0012's Consequences and the truth row): **(a)** a bare `go test ./...` / `go build` run **without `make`** stays non-hermetic (hermeticity is a `make`-boundary property, not a toolchain property); **(b)** `GOENV=off` also ignores a *legitimate* operator `go env -w` setting — an explicitly exported variable per invocation (`GOPROXY=… make verify`) is the documented escape hatch.
- **FR-010**: The round MUST be **tooling + truth only**: it MUST change the `Makefile`, add ADR 0012, and update the truth, and MUST NOT modify any production Go behaviour, any flag, any exit code, any stream contract, or any existing gate's semantics.

**Non-Functional Requirements**:

- **NFR-001**: The change MUST be **deterministic** (no wall-clock/timing dependence; the block is a static environment definition).
- **NFR-002**: **`make` + the Go toolchain only** — no new module dependency; `go.mod`/`go.sum` unchanged.

---

### User Story 2 - The rule is recorded and citable (Priority: P2)

As a maintainer, I want the neutralise/preserve policy recorded in an ADR and in the technology-stack truth, so the rule is discoverable, future rounds can cite it, and a `Makefile` reader can find *why* the block exists.

**Why this priority**: it bounds the round in truth and prevents the rule from becoming folklore in a comment.

**Independent verification**: read `docs/decisions/0012-hermetic-make-go-env.md` (+ the index row) and `specs/truth/techstack.md` (Build & Tooling) — the neutralise/preserve sets, the `GOENV=off` rationale, the `CGO_ENABLED` stance, the scope, and the residuals are stated.

**Acceptance Scenarios**:

1. **Given** the technology-stack truth, **When** the round lands, **Then** Build & Tooling records the hermetic invocation policy and the `Task runner` row notes the hermetic environment.
2. **Given** the decisions index, **When** the round lands, **Then** ADR 0012 is present and indexed.

**Functional Requirements**:

- **FR-011**: The ADR MUST state the **neutralise set**, the **preserve set**, the **`GOENV=off`** rationale (the env-file fallback defeats an unset/empty value — round-042 F-2), the **`CGO_ENABLED` host-default** stance, the **unconditional all-targets** scope, the **defence-in-depth** relationship to round-042's `childEnv`, and the two recorded residuals (FR-009).

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only** (the repo is bash/POSIX-only; the `Makefile` is GNU-make shaped); no Windows variant.

---

### User Story 3 - No behaviour change for a clean environment (Priority: P3)

As a maintainer, I want the round to be a **pure hardening**: under a clean environment every gate keeps its verdict, the cross-compile and safety gates keep working, and the mutating targets (`fmt`, `tidy`) are unaffected.

**Why this priority**: it protects the "no gate's behaviour changes for a clean environment" acceptance criterion #96 itself states.

**Independent verification**: run `make verify` (and the individual targets) under a **clean** env before and after — identical verdicts; run `make tidy`/`make fmt` — no unexpected diff; run `make verify-cross-compile` — still 4/4.

**Acceptance Scenarios**:

1. **Given** a clean environment, **When** `make verify` runs, **Then** every member gate's verdict is unchanged and the aggregate is green.
2. **Given** a clean environment, **When** `make verify-cross-compile` runs, **Then** it still builds + vets 4/4 targets (its inline per-target `GOOS`/`GOARCH`/`CGO_ENABLED=0` still wins for that recipe).
3. **Given** a clean environment, **When** `make tidy` (and `make fmt`) run, **Then** `go.mod`/`go.sum` are unchanged and no source diff is produced.

**Functional Requirements**:

- **FR-012**: The gate MUST NOT clobber a recipe's **inline** per-target environment assignment (the `verify-cross-compile` `CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build` form MUST keep overriding for its own recipe).

## Requirements *(mandatory)*

### Global Requirements

#### Functional Requirements

- **FR-013**: The round MUST pin the neutralise set, the preserve set, the `CGO_ENABLED` stance, and the scope explicitly in `research.md`/`plan.md`, and MUST provide the falsifiability witnesses (the four hostile-env cases) plus the positive controls (cross-compile 4/4; `tidy`/`fmt` no-diff). Shape: **ADD** — the round adds a `Makefile` boundary + ADR + truth row; it changes no existing truth row's meaning beyond the `Task runner` note.

#### Non-Functional Requirements

- **NFR-004**: `make verify` (including `verify-no-test-sleep`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `verify-architecture`, `lint`, `govulncheck`) and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL rows (`/axb-dsl-refine` **NOOP** — A6).

### Key Entities

- **Hermetic boundary**: the single top-of-`Makefile` environment definition that governs how every recipe's Go-toolchain invocation is launched.
- **Neutralise set**: the environment inputs replaced or cleared (`GOENV`, `GOFLAGS`, `GO111MODULE`, `GOEXPERIMENT`, `GOWORK`, ambient `GOOS`/`GOARCH`/`GOARM`).
- **Preserve set**: the environment inputs passed through unchanged (warm-cache: `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE`; network/checksum: `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE`).

## Success Criteria *(mandatory)*

- **SC-001**: The four hostile-env cases are **green**: `GOENV=<file: GOFLAGS=-mod=vendor>`, `GOFLAGS=-trimpath`, `GO111MODULE=off`, and a stray `GOWORK` each leave `make verify` (and a toolchain target) exit 0 — reproduced as falsifiability witnesses, then reverted (ADR 0010 doctrine). (covers FR-001, FR-002, FR-003)
- **SC-002**: For a **clean** environment, every gate's verdict is unchanged and no target's observable behaviour changes; `go.mod`/`go.sum` are unchanged; no new dependency. (covers FR-006, NFR-002)
- **SC-003**: **Positive control — cross-compile**: `make verify-cross-compile` still builds + vets 4/4 targets under the hermetic env (its inline per-target `GOOS`/`GOARCH`/`CGO_ENABLED=0` still wins for its recipe). (covers FR-004, FR-012)
- **SC-004**: **Positive control — mutating targets**: under a clean env, `make tidy` and `make fmt` leave `go.mod`/`go.sum` unchanged and produce no source diff. (covers FR-006, Q7-R2)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the hermetic invocation policy and the `Task runner` note; **ADR 0012** + its index row are present; **no production Go behaviour** changes; the topology audit is unchanged/green. (covers FR-007, FR-010, FR-011, NFR-004)

## Edge Cases

- **A target that legitimately overrides per-target** (`verify-cross-compile`) → its inline `GOOS`/`GOARCH`/`CGO_ENABLED=0` MUST still win (SC-003).
- **A nested `go build` spawned by a scripted recipe** (`make test`, `verify-no-network`) → covered by the `make` export (the reason A1 was chosen over per-recipe prefixes).
- **`$(shell command -v …)` parse-time probes** → they read `PATH` only; the block MUST be placed above them and MUST NOT break tool resolution (NFR-002; `staticcheck`/`golangci-lint`/`govulncheck` still resolve from `PATH`).
- **A directly-run gate bypassing `make`** → `tools/arch` keeps its own `childEnv` (FR-008); other gates run only via `make`.
- **A cold module cache / a proxied environment** → the preserve set keeps `GOPROXY`/`GOSUMDB`/… and the cache roots, so resolution behaves as today.
- **An operator who *wants* their env-file setting** → the documented escape hatch: export the variable explicitly for that invocation (FR-009b).
- **`make GOFLAGS=… verify`** (a deliberate command-line override) → out of scope; the block owns the default, and an explicit per-invocation export is the escape hatch (FR-009b).
- **The round must not touch the frozen guard** → `tools/arch/**` is unchanged (round-042 frozen history; only a **comment** may be added there to cross-reference — recorded as a scope note in `research.md`).

## Assumptions

- **A1 (atomicity)** — the `Makefile` block lands **together with** the ADR + the truth row (one PR).
- **A2 (weak E2E carrier)** — the round changes no `tellme` CLI behaviour, so the E2E suite is a **weak acceptance carrier** here; the witness is the **scripted hostile-env Makefile invocation**, not "the suite is green". Per this, **`/axb-spec-by-example` is NOOP** (no user-facing business journey) — precedent: rounds 020/031/036/041/042.
- **A3 (mechanism detail is RD)** — the exact GNU-make syntax (`export`/`unexport` forms, placement above the `$(shell …)` probes) is an **RD** decision (`research.md` D1), not an FR.
- **A4 (ADR required)** — **ADR 0012** records the policy (Q5 → G1; a project-level rule other rounds may cite).
- **A5 (no other interfaces)** — `/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted/runtime state); `/axb-dsl-refine` = **NOOP** (the change is a **dev-surface** (`make`) environment boundary, not the `tellme` CLI contract); `/axb-ui-plan` **skipped**.
- **A6 (scope guard)** — the round touches **only** the `Makefile`, `docs/decisions/**`, and `specs/truth/techstack.md` (+ the plan package). No `internal/**`, `cmd/**`, `tests/**`, `tools/**` **behaviour** change.

## Out of scope (recorded)

- **Bare `go` invocations outside `make`** (`go test ./...`, `go build`) — stay non-hermetic (FR-009a residual).
- **A global toolchain shim/wrapper** — a much larger scope; not asked for.
- **Pinning `CGO_ENABLED=0` globally** — rejected (Q3/V2: a real behaviour change vs future cgo-tagged code).
- **Changing the `go` toolchain, any gate's semantics, or any product behaviour.**
- **Modifying the frozen round-042 guard** (`tools/arch/**` behaviour) — retained as-is (FR-008).
- **R2 of [#92](https://github.com/gosharplite/tellme/issues/92)** (composition-root extraction) — the next round (`044-*`) per Q8/O1.
