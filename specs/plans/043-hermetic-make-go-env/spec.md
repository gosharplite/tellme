# Feature Specification: a hermetic `make` Go-toolchain environment (round 043)

**Feature Branch**: `043-hermetic-make-go-env`

**Created**: 2026-09-18

**Status**: Draft — plan half only (specify → spec-by-example **NOOP** → technical-research → system-analysis). Anchor issue [#96](https://github.com/gosharplite/tellme/issues/96) — the **outer** half of round-042's **F-2**. Clarify round 1 locked Q1–Q8 (below).

> **Review fold (PR [#97](https://github.com/gosharplite/tellme/pull/97) review `5242446755` — B-1…B-4 · TD-1 · R-1…R-3; all folded before merge):** **B-1** `GOFLAGS=-trimpath` is **green before** the change, so it is a **neutralisation assertion**, not a witness — the round has **four** red→green witnesses (`GOENV` file, `GO111MODULE=off`, `GOWORK`, `GOTOOLCHAIN`; §D7/SC-001/FR-013 corrected). **B-2** the direct path's **outer** `go test`/`go vet` are **not** hermetic — `childEnv` keeps the gate's **verdict** hermetic only (§Q6/FR-008 corrected). **B-3** the two sites' invariant is **coverage**, not equality (they neutralise by different mechanisms); the non-re-set names (`GOARM`, `GOEXPERIMENT`, …) are a recorded residual. **B-4** the neutralise set now follows an explicit **inclusion criterion** (D2) and gains **`GOTOOLCHAIN`** (+ the micro-arch family, `GOFIPS140`, `GODEBUG`). **TD-1** parse-time `$(shell …)`/`$(eval …)`/command-line variables are outside the boundary (recorded). **R-1** `CGO_ENABLED` is *preserved from the caller*; scope is *unconditional for **ambient** input* (the command-line form is the hatch). **R-2** FR-008's **MUST** stands (the `plan.md`/T003 "optional" hedge is dropped). **R-3** the repo-wide scope claim's figure is corrected (§*Grounded*).

**Input**: Operator request to resolve [#96](https://github.com/gosharplite/tellme/issues/96) via the AIxBDD process. #96: *every gate in the `Makefile` invokes the Go toolchain as a child of `make`, inheriting the ambient environment; a **persisted** (`go env -w`) or exported Go setting therefore reaches the **outer** `go test`/`go vet`/`go build`, so a developer/CI image that exports `GOFLAGS=-mod=vendor` / `-trimpath`, a `GOENV` file, `GO111MODULE=off`, or a stray `GOWORK` can redden the whole `verify` aggregate for reasons unrelated to the tree.*

**Behaviour intent**: **ADD** a single **hermetic invocation boundary** to the `Makefile` so that **every** Go-toolchain invocation `make` launches runs with a **sanitised, explicit** environment — neutralising the build-context inputs an ambient/persisted Go env can inject, while preserving the warm-cache and network/checksum inputs a cold-cache or proxied build legitimately needs. This **generalises** round-042's `childEnv` discipline (ADR 0011 **D5**) from the gate's own child `go list` to **every** target, and completes round-020's `verify-cross-compile` `CGO_ENABLED=0` pin (PR #46 **TD1**) from one variable/one target to a repo-wide rule. It is a **tooling/truth slice** in the round-020 / round-032 / round-042 lineage: it changes **no** user-facing behaviour and **no** product code — it changes the `Makefile`, adds an ADR, and updates the technology-stack truth.

---

## Locked decisions (clarify round 1 — Q1–Q8)

| # | Decision |
| --- | --- |
| **Q1** | **Option A — neutralise at the invocation boundary.** Actually fix the class (repo-wide, once); rejected B (document + loud guard: does not make the gates hermetic) and C (non-goal). |
| **Q2** | **A1 — one top-of-`Makefile` `export`/`unexport` block.** Single owned definition; covers **every** recipe *and* the nested `go build` a scripted recipe (`make test`, `verify-no-network`) spawns — which per-recipe prefixes (A2) would miss. |
| **Q3** | **V1 (criterion-derived) — the neutralise/preserve sets below; `CGO_ENABLED` left untouched (preserved from the caller).** *Neutralise/replace*: `GOENV` (→ `off`), `GOWORK` (→ `off`), `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOTOOLCHAIN`/`GOFIPS140`/`GODEBUG` (unset), and the whole **target triple** — ambient `GOOS`/`GOARCH` **and** the micro-architecture family (`GOARM`, `GOARM64`, `GOAMD64`, `GO386`, `GOMIPS`, `GOMIPS64`, `GOPPC64`, `GORISCV64`, `GOWASM`) — unset. *Preserve*: `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE` **and** `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE`. **The set is derived from an explicit inclusion criterion** (neutralise the ambient *build context* — module mode · flags · workspace · experiments · target triple · toolchain selection · build-mode/runtime switches; preserve the *plumbing*), **not** an incident list: the PR #97 review (**B-4**) showed `GOTOOLCHAIN` (reproduced red) and the micro-arch family sit inside the class. **Not** V2 (global `CGO_ENABLED=0`: a real behaviour change — it would exclude future cgo-tagged code, e.g. the recorded darwin `mach` CPU sampler, and diverge from a plain `go build`); **not** V3 (leave `GOOS`/`GOARCH`/`CGO_ENABLED` ambient). |
| **Q4** | **S1 — the block is unconditional *for ambient input*; ALL targets hermetic** (`build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, `verify`). The **command-line** form (`make GOENV=<file> …`) is the deliberate escape hatch (**D5**, folded per PR #97 review **R-1(2)** — an `export`ed name loses to a command-line assignment). Rejected S2 (gate-family only: fights the A1 mechanism and reintroduces per-site drift). |
| **Q5** | **G1 — a new ADR (`0012`) + its index row, and a `specs/truth/techstack.md` (Build & Tooling) update.** The rule is project-level and future rounds/artifacts must cite it (the ADR 0011 precedent). |
| **Q6** | **D1 — keep round-042's `tools/arch` `childEnv` as defence-in-depth for the gate's *verdict*.** It keeps the gate's **own child `go list`** hermetic — i.e. the verdict — on the **documented direct invocation** (`go test -count=1 -tags=arch … ./tools/arch`) which bypasses `make`; the direct path's **outer** `go test`/`go vet` remain **non-hermetic** (the R1 residual; reproduced by the PR #97 review — **B-2**, corrected). The `Makefile` block is the **primary owner**; the two sites are **complementary mechanisms** whose invariant is **coverage**, not equality (**B-3**, corrected). Rejected D2 (delete `childEnv`: loses the direct-path verdict hermeticity and reopens frozen artifacts) and D3 (a `Makefile`↔Go shared list: a new coupling mechanism for a handful of literals). |
| **Q7** | **R1 recorded** (a bare `go test ./...` / `go build` *without* `make` stays non-hermetic — a recorded residual) · **R2 folded into verification** (`fmt`/`tidy` are mutating targets; a positive control confirms the block does not break them and `go.mod`/`go.sum` stay unchanged) · **R3 recorded** (`GOENV=off` also ignores a *legitimate* operator `go env -w`; an explicit exported var per invocation is the documented escape hatch). |
| **Q8** | **O1 — this is round `043-*`; R2 of [#92](https://github.com/gosharplite/tellme/issues/92) slides to `044-*`** (the `STATUS.md` roadmap line is updated at closeout). Rejected folding #96 into R2's package (`fresh-package-per-round`; two unrelated themes). |

---

## Grounded in the current system

- **Every `Makefile` recipe that invokes the Go toolchain inherits `make`'s environment.** A `grep` for the toolchain verbs (`go build|vet|test|fmt|mod|list|run` + the `$(STATICCHECK)`/`$(GOLANGCI)`/`$(GOVULNCHECK)` forms) resolves to **25 lines**, of which **16** are *executed* toolchain invocations (`verify-cross-compile`'s single recipe line carries **two** — `build` + `vet`); the other **10** are `make help` echoes (4), comments (2), or `ifeq` guards (4). **No** executed invocation sanitises the environment (PR #97 review **R-3** asked for a reproducible figure; measured here as 25 lines / 16 invocations, superseding the earlier "~20 sites").
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

**Independent verification**: with a hostile ambient Go env set, run the **four red→green cases** (`GOENV=<file: GOFLAGS=-mod=vendor>`, `GO111MODULE=off`, a stray `GOWORK`, `GOTOOLCHAIN=go1.99.9`) — before the change they **fail**; after, they are **green**; revert the block and they fail again (falsifiability). (`GOFLAGS=-trimpath` is a **neutralisation assertion**, not a witness — it is green before; PR #97 review **B-1**.)

**Acceptance Scenarios**:

1. **Given** `GOENV` pointing at a file containing `GOFLAGS=-mod=vendor`, **When** `make verify` runs, **Then** it is **green** (previously exit 2).
2. **Given** each of `GO111MODULE=off`, a stray `GOWORK=<path>`, and `GOTOOLCHAIN=go1.99.9`, **When** a toolchain target runs, **Then** the ambient value is **neutralised** and the target behaves as under a clean environment. **And** the green-before inputs (`GOFLAGS=-trimpath` and the long-tail D2 names) are asserted **absent from the recipe environment** (a neutralisation **assertion**, not a witness — PR #97 review **B-1**).
3. **Given** the aggregate `make verify`, **When** it runs under any of the above hostile envs, **Then** it exits 0 and every member gate's verdict matches its clean-env verdict.

**Functional Requirements**:

- **FR-001**: Every `make` target that invokes the Go toolchain (`build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, `verify`) MUST run it with a **hermetic environment**: the **neutralise set** neutralised, the **preserve set** preserved.
- **FR-002 (neutralise set, criterion-derived)**: the environment `make` exports to its recipes MUST neutralise — with an explicit value where an unset/empty value would be defeated by the env-file fallback — `GOENV` (set to `off`), `GOWORK` (set to `off`), `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOTOOLCHAIN`/`GOFIPS140`/`GODEBUG` (unset), and the ambient **target triple** — `GOOS`/`GOARCH` and the micro-architecture family `GOARM`/`GOARM64`/`GOAMD64`/`GO386`/`GOMIPS`/`GOMIPS64`/`GOPPC64`/`GORISCV64`/`GOWASM` (unset ⇒ host-native). Membership follows the **inclusion criterion** (neutralise the ambient build context; preserve the plumbing), **not** an incident enumeration: `GOTOOLCHAIN` (reproduced red) and the micro-arch family are in-class (PR #97 review **B-4**).
- **FR-003 (preserve set)**: `PATH`, `HOME`, `GOPATH`, `GOMODCACHE`, `GOCACHE` (the warm-module-cache set) and `GOPROXY`, `GOSUMDB`, `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE` (the network/checksum set) MUST be preserved so a cold-cache or proxied build still resolves.
- **FR-004**: `CGO_ENABLED` MUST be **preserved from the caller** (the block neither sets nor unsets it — it is **not** globally pinned); the only cgo pinning remains `verify-cross-compile`'s inline `CGO_ENABLED=0` **for its own targets** (PR #46 TD1), which MUST keep working.
- **FR-005**: The neutralisation MUST be defined **once**, at the invocation boundary (a single top-of-`Makefile` block), not duplicated per recipe — so a target added later is hermetic **by construction** for **recipes and their descendants** (parse-time `$(shell …)`/`$(eval …)`/command-line variables are outside the boundary — FR-009d), and a per-recipe inline override (e.g. `verify-cross-compile`'s `GOOS=…`) still wins for that recipe.
- **FR-006**: For a **clean** environment, **no** target's observable behaviour changes; **no** gate's semantics change; `go.mod`/`go.sum` stay unchanged; no new dependency is added.
- **FR-007**: The rule MUST be recorded in a new **ADR 0012** (`docs/decisions/0012-hermetic-make-go-env.md`) with its `docs/decisions/README.md` index row, and in `specs/truth/techstack.md` (Build & Tooling) as a citable row (with the `Task runner` row noting the hermetic environment).
- **FR-008**: Round-042's `tools/arch` `childEnv` filter MUST be **retained** so the gate's **verdict** stays hermetic on the **direct** invocation of the gate (its own child `go list`), **and** the ADR MUST record that the direct path's **outer** `go test`/`go vet` remain **non-hermetic** (R1). The two sites MUST carry cross-referencing documentation, and the ADR MUST state the **coverage** invariant — *every name in FR-002's set is either re-set explicitly by `childEnv` or recorded as a known non-covered name on the direct path* — **not** an equality invariant (the mechanisms differ; PR #97 review **B-2/B-3**).
- **FR-009**: The residuals MUST be recorded (in ADR 0012's Consequences and the truth row): **(a)** a bare `go test ./...` / `go build` run **without `make`** stays non-hermetic — hermeticity is a **`make`-boundary** property, not a toolchain property — including the gate's direct invocation path's **outer** `go test`/`go vet` (R1); **(b)** the escape hatch is **per variable class** — an already-`export`ed name via a command-line assignment (`make GOENV=<file> verify`), a plumbing name via an explicit export (`GOPROXY=… make verify`), and the unexported names only by invoking the toolchain directly or editing the block (R3); **(c)** inside the frozen round-042 guard `childEnv` **drops but does not re-set** `GOARM`/`GOEXPERIMENT` (and the names added this round), so on the direct path an env-file value for those still reaches the gate's child (R4); **(d)** parse-time `$(shell …)`/`$(eval …)`/command-line variables are outside the boundary — keep `go` out of `$(shell …)`, and a future Go env name outside the FR-002 criterion's categories is out of scope until added (R5).
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

- **FR-011**: The ADR MUST state the **inclusion criterion** and the neutralise set (FR-002), the **preserve set**, the **`GOENV=off`** rationale (the env-file fallback defeats an unset/empty value — round-042 F-2), the **`CGO_ENABLED` preserved-from-the-caller** stance, the **unconditional-for-ambient-input / all-targets** scope (with the command-line hatch), the **coverage** invariant relating the `Makefile` set to round-042's `childEnv` (**not** equality), the **four red→green witnesses + the neutralisation assertions** (FR-013), and the residuals R1/R3/R4/R5.

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

- **FR-013**: The round MUST pin the neutralise set (**criterion-derived**, FR-002), the preserve set, the `CGO_ENABLED` stance, and the scope explicitly in `research.md`/`plan.md`, and MUST provide the falsifiability witnesses — **four red→green** cases (`GOENV=<file: GOFLAGS=-mod=vendor>`, `GO111MODULE=off`, a stray `GOWORK`, `GOTOOLCHAIN=go1.99.9`) — **plus the neutralisation assertions** for the green-before inputs (`GOFLAGS=-trimpath` and the long-tail names), **plus** the positive controls (cross-compile 4/4; `tidy`/`fmt` no-diff). A green-before input MUST NOT be presented as a witness (PR #97 review **B-1**). Shape: **ADD** — the round adds a `Makefile` boundary + ADR + truth row; it changes no existing truth row's meaning beyond the `Task runner` note.

#### Non-Functional Requirements

- **NFR-004**: `make verify` (including `verify-no-test-sleep`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `verify-architecture`, `lint`, `govulncheck`) and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL rows (`/axb-dsl-refine` **NOOP** — A6).

### Key Entities

- **Hermetic boundary**: the single top-of-`Makefile` environment definition that governs how every recipe's Go-toolchain invocation is launched.
- **Neutralise set**: the environment inputs replaced or cleared, **criterion-derived** (FR-002): `GOENV`→`off`, `GOWORK`→`off`; `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOTOOLCHAIN`/`GOFIPS140`/`GODEBUG` unset; and the ambient target triple — `GOOS`/`GOARCH` **and** the micro-architecture family (`GOARM`/`GOARM64`/`GOAMD64`/`GO386`/`GOMIPS`/`GOMIPS64`/`GOPPC64`/`GORISCV64`/`GOWASM`) — unset.
- **Preserve set**: the environment inputs passed through unchanged (warm-cache: `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE`; network/checksum: `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE`).

## Success Criteria *(mandatory)*

- **SC-001**: The four red→green hostile-env cases are **neutralised** (each fails before the change, passes after): `GOENV=<file: GOFLAGS=-mod=vendor>` (incl. the **aggregate** `make verify` ⇒ exit 0 with every member green), `GO111MODULE=off`, a stray `GOWORK`, and `GOTOOLCHAIN=go1.99.9` — each **fails before** the change and **passes after** (reproduced then reverted, ADR 0010 doctrine); `GOFLAGS=-trimpath` and the long-tail names are asserted **absent from the recipe environment** (a neutralisation **assertion**, not a witness — PR #97 review **B-1**). (covers FR-001, FR-002, FR-003)
- **SC-002**: For a **clean** environment, every gate's verdict is unchanged and no target's observable behaviour changes; `go.mod`/`go.sum` are unchanged; no new dependency. (covers FR-006, NFR-002)
- **SC-003**: **Positive control — cross-compile**: `make verify-cross-compile` still builds + vets 4/4 targets under the hermetic env (its inline per-target `GOOS`/`GOARCH`/`CGO_ENABLED=0` still wins for its recipe). (covers FR-004, FR-012)
- **SC-004**: **Positive control — mutating targets**: under a clean env, `make tidy` and `make fmt` leave `go.mod`/`go.sum` unchanged and produce no source diff. (covers FR-006, Q7-R2)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the hermetic invocation policy and the `Task runner` note; **ADR 0012** + its index row are present; **no production Go behaviour** changes; the topology audit is unchanged/green. (covers FR-007, FR-010, FR-011, NFR-004)

## Edge Cases

- **A target that legitimately overrides per-target** (`verify-cross-compile`) → its inline `GOOS`/`GOARCH`/`CGO_ENABLED=0` MUST still win (SC-003).
- **A nested `go build` spawned by a scripted recipe** (`make test`, `verify-no-network`) → covered by the `make` export (the reason A1 was chosen over per-recipe prefixes).
- **A `go` invocation in a parse-time expansion** (`$(shell …)`/`$(eval …)`, an `include`d makefile, or a make **command-line** variable) → **outside** the boundary (`research.md` D1/TD-1): an `unexport`ed name is still visible to `$(shell …)`'s child, and a command-line `VAR=…` is not exported to recipes at all. Keep `go` out of `$(shell …)`; record any future name by the FR-002 criterion.
- **The existing `$(shell command -v …)` toolchain probes** → they read `PATH` only (a preserved name); the block MUST be placed **above** them and MUST NOT break tool resolution (`staticcheck`/`golangci-lint`/`govulncheck` still resolve from `PATH`; NFR-002).
- **A deliberate command-line override** → verified: `make GOFLAGS=-mod=vendor vet` is **neutralised** (a command-line var is not exported to recipes), while `make GOENV=<file> vet` **escapes** (an already-`export`ed name loses to the command-line assignment) — the latter is the documented escape hatch (`research.md` D8/R3; PR #97 review **R-1(2)**).
- **A directly-run gate bypassing `make`** → `tools/arch` keeps its own `childEnv`, which keeps the **verdict** hermetic; the direct path's outer `go test`/`go vet` are non-hermetic (FR-008/R1).
- **A cold module cache / a proxied environment** → the preserve set keeps `GOPROXY`/`GOSUMDB`/… and the cache roots, so resolution behaves as today.
- **An operator who *wants* their env-file setting** → the documented escape hatch, **per variable class** (`research.md` D8/R3): an already-`export`ed name via a command-line assignment (`make GOENV=<file> verify`); a plumbing name via an explicit export (`GOPROXY=… make verify`); the unexported names only by invoking the toolchain directly or editing the block.
- **The round must not touch the frozen guard's behaviour** → `tools/arch/arch_test.go` changes **comment-only** (round-042 frozen history; the one cross-reference comment, `research.md` scope note).

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
