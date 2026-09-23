# ADR 0012 — A hermetic `make` Go-toolchain invocation environment

- **Status:** Accepted
- **Date:** 2026-09-18
- **Deciders:** tellme owner
- **Related:** issue [#96](https://github.com/gosharplite/tellme/issues/96) (the anchor — the *outer* half of round-042's F-2) ·
  PR [#95](https://github.com/gosharplite/tellme/pull/95) fold-review #2 (comment `5721537481`, Fold 2 — where the gap was found) ·
  **ADR 0011 D5** (round-042's gate child-env discipline — generalised here) ·
  round-020 `verify-cross-compile` + PR #46 review **TD1** (the `CGO_ENABLED=0` pin for one target — the **precedent** generalised here as an **invocation** rule; the pin itself stays recipe-local, **D4**) ·
  round 043 (`specs/plans/043-hermetic-make-go-env` — this ADR's round) ·
  PR [#97](https://github.com/gosharplite/tellme/pull/97) review (comment `/pullrequestreview-5242446755` — B-1…B-4 + TD-1/R-1…R-3 folded below) ·
  PR [#99](https://github.com/gosharplite/tellme/pull/99) review (`/pullrequestreview-5242608457` — N-1…N-3 folded below; `/pullrequestreview-5242647672` re-review FOLD-ACCEPTED) ·
  **ADR 0010** (the falsifiability/verification doctrine the witnesses follow).

## Context

Every gate in the `Makefile` invokes the Go toolchain as a **child of `make`**, inheriting the ambient environment. A **persisted** (`go env -w`) or **exported** Go setting therefore reaches not just a gate's own tooling but the **outer** `go test` / `go vet` / `go build` themselves. Issue [#96](https://github.com/gosharplite/tellme/issues/96) reproduces it for `verify-architecture`:

```sh
$ printf 'GOFLAGS=-mod=vendor\n' > /tmp/goenv
$ GOENV=/tmp/goenv make verify-architecture
… inconsistent vendoring …      # exit 2, unrelated to the tree
```

Round 042's gate already filters the environment of **its own** child `go list` (`tools/arch/arch_test.go` `childEnv`, ADR 0011 **D5**), so the gate's *verdict* is hermetic — but the surrounding `go test`/`go vet` that `make` runs are not, and the same is true of **every** existing target. A developer/CI image that exports `GOFLAGS=-mod=vendor` or `-trimpath`, a `GOENV` file, `GO111MODULE=off`, a stray `GOWORK`, or `GOTOOLCHAIN=…` can redden the whole `verify` aggregate for reasons unrelated to the tree. Round 020's `verify-cross-compile` pinned `CGO_ENABLED=0` for **one** variable on **one** target (PR #46 TD1); this ADR generalises the concern to a repo-wide rule.

## Decision

**D1 — One hermetic boundary, at the top of the `Makefile`.** The neutralisation is defined **once** — a GNU-make `export`/`unexport` block placed near the top of the `Makefile` (above the `$(shell command -v …)` toolchain probes and before any recipe) — so **every** recipe, and every descendant process a recipe spawns (e.g. the E2E harness's `go build` behind `make test`, the witness behind `verify-no-network`), runs with the sanitised environment. A target added later is hermetic **by construction**. The rule is **not** applied per recipe (a `$(HERMETIC_ENV)` prefix per invocation would miss descendant spawns and drift as recipes are added).

**Scope of the boundary (what it does *not* govern).** `export`/`unexport` govern the environment of **recipes and their descendants** — and **nothing else**. A `$(shell …)` / `$(eval …)` child (and `include`d makefiles) receives make's **original environment verbatim**: a makefile `export` does **not** reach it, and `unexport` does **not** strip it. Parse-time expansion is therefore **outside** the boundary — for **every** name, not merely the unexported ones. Measured (GNU Make 4.3, this block in place; PR #99 review **N-1** corrected an earlier mis-statement of this clause):

```
ambient: GOENV=<file> GOFLAGS=-mod=vendor GOTOOLCHAIN=go1.99.9
recipe:      GOENV=[off]            GOFLAGS=[]            GOTOOLCHAIN=[]          ← neutralised
parse-time:  GOENV=[<file>]         GOFLAGS=[-mod=vendor] GOTOOLCHAIN=[go1.99.9]   ← ambient, untouched
```

The command-line form behaves as follows (verified): a `VAR=…` on the make **command line** is not exported to recipes (`make GOFLAGS=-mod=vendor vet` ⇒ neutralised), while an already-`export`ed name is **overridden** by it (`make GOENV=<file> vet` ⇒ the file wins — the escape hatch, D8/R3).

**Consequence:** keep `go` **out of** `$(shell …)` — a parse-time `go` invocation escapes the boundary entirely (it sees the ambient environment, whatever the name). Recorded as a residual (D8/R5).

**D2 — The neutralise set, derived from an explicit inclusion criterion (not an incident list).** The criterion — the rule a reader applies to decide whether the *next* name belongs:

> **Neutralise the ambient *build context*** — anything that changes **what** the Go toolchain builds or **which** toolchain builds it: module mode, build flags, workspace, experiments, target triple (OS/arch **and** micro-architecture level), toolchain selection, and build-mode switches / runtime defaults. **Preserve the *plumbing*** — anything that only lets that build *happen*: the warm cache/runtime roots and the network/checksum configuration.

Under that criterion the round neutralises:

| Variable(s) | Category | Why |
| --- | --- | --- |
| `GOENV` → `off` | env file | **Load-bearing.** Go falls back to the env **file** for a variable that is unset **or empty**, so unsetting `GOFLAGS` does **not** neutralise a persisted `go env -w GOFLAGS=-mod=vendor` (round-042 F-2). `off` disables the env-file fallback entirely. |
| `GOFLAGS` | module mode / flags | e.g. `-mod=vendor`, `-trimpath` — they change the build. |
| `GO111MODULE` | module mode | An ambient `off` disables modules. |
| `GOEXPERIMENT` | experiments | Gates `goexperiment.*` build constraints (changes the file set). |
| `GOWORK` → `off` | workspace | A stray `go.work` redirects the build (reproduced: exit 2). |
| `GOTOOLCHAIN` | toolchain selection | An exported/persisted pin selects a **different toolchain** (reproduced: `GOTOOLCHAIN=go1.99.9` ⇒ exit 2). Unset ⇒ Go's `auto` ⇒ selection is driven by the **tree's** `go.mod`, not the caller's shell. |
| `GOOS`, `GOARCH` | target triple | A stale exported target redirects a host build. |
| `GOARM`, `GOARM64`, `GOAMD64`, `GO386`, `GOMIPS`, `GOMIPS64`, `GOPPC64`, `GORISCV64`, `GOWASM` | target triple (micro-architecture level) | The **micro-arch** family — "host-native" for an unset `GOOS`/`GOARCH` does **not** cover it (PR #97 review R-1(1)/B-4). |
| `GOFIPS140` | build-mode switch | Selects the FIPS 140 module at build time. |
| `GODEBUG` | runtime defaults | Sets the built binary's/tests' default debug behaviour. |

**D3 — The preserve set.** Passed through unchanged, because a cold-cache or proxied build legitimately needs them:

- **cache / runtime plumbing:** `PATH`, `HOME`, `GOPATH`, `GOMODCACHE`, `GOCACHE`;
- **network / checksum:** `GOPROXY`, `GOSUMDB`, `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE`.

*(An `env -i`-style wipe would fail with a *module cache not found* / resolution error — the `verify-cross-compile` PR #46 TD1 lesson, applied to the parent.)*

**D4 — `CGO_ENABLED` is *preserved from the caller*; it is not globally pinned.** The block neither sets nor unsets it, so an ambient `CGO_ENABLED` reaches the recipe as-is. Pinning `0` globally would be a real **behaviour change**: it would exclude any future cgo-tagged code (the recorded darwin `mach` CPU sampler is exactly such a candidate) from `make build`/`make test`, and diverge from a plain `go build`. The only cgo pinning remains `verify-cross-compile`'s **inline** `CGO_ENABLED=0`, which applies to its own recipe only — and the hermetic block **must not** clobber it (a recipe's inline assignment wins for that recipe).

**D5 — Scope: unconditional for *ambient* input — all targets.** The block applies to `build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, and `verify`; there is no per-target opt-out. It is **unconditional for ambient *Go-env* input** (the exceptions are **make-level** inputs — `-e` (**R6**, reproduced) and a `MAKEFILES`-injected `override export`, same class but contrived); the deliberate **command-line** form (`make GOENV=<file> …`) is the escape hatch (D8/R3) — that is the opt-out, so no separate `HERMETIC=0` switch is introduced. The two **mutating** targets (`fmt`, `tidy`) are covered and verified harmless (D7).

**D6 — Ownership: the `Makefile` block is the primary owner; round-042's `childEnv` is retained as defence-in-depth for the gate's *verdict*.** The gate is documented as runnable **directly** (`go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch`), bypassing `make`. On that path `childEnv` keeps the gate's own child `go list` hermetic — i.e. **the verdict** stays host- and env-file-neutral — but the direct path's **outer** `go test`/`go vet` remain **non-hermetic** (an ambient `GOENV=<file>` still reaches them; reproduced: exit 1). That is R1, not a claim `childEnv` does not make.

The two sites are **complementary mechanisms, not duplicates**: the `Makefile` **disables the env file** (`GOENV=off`) and unsets the ambient names; `childEnv` **re-sets explicit non-empty values** (`GOOS`/`GOARCH` per target, `CGO_ENABLED=0`, `GOFLAGS=-mod=readonly`, `GO111MODULE=on`, `GOWORK=off`) that win over the file (round-042 **F-2**). **Equality of the two sets is therefore the wrong invariant** — the right one is **coverage**: *every name in D2's neutralise set must be either re-set explicitly by `childEnv` or recorded as a non-covered class (D8/R4)*. This is a **maintenance invariant recorded here**, not a witness (it cannot fail on its own).

**D7 — Verification.** The acceptance evidence is:

- **Four red→green witnesses** (reproduced **red before** the change and green after, then reverted — ADR 0010 doctrine; a witness must be able to fail):
  1. `GOENV=<file: GOFLAGS=-mod=vendor>` — red: *inconsistent vendoring* (exit 2);
  2. `GO111MODULE=off` — red: GOPATH-mode resolution failure (exit 2);
  3. `GOWORK=<stray go.work>` — red: *directory prefix . does not contain modules listed in go.work* (exit 2);
  4. `GOTOOLCHAIN=go1.99.9` — red: *toolchain not available* (exit 2).
- **Neutralisation assertions** (a *green-before* input is not a witness — the round-043 review **B-1** correction): `GOFLAGS=-trimpath` and the long-tail names are asserted **absent from the recipe environment** (a `make`-level probe), which is exactly the property the round claims; they are not counted as witnesses.
- **Aggregate-level evidence:** `GOENV=<file: GOFLAGS=-mod=vendor> make verify` ⇒ **exit 0** with every member green (cross-compile 4/4 · `verify-architecture` ok · lint 0 · govulncheck 0 reachable) — stronger than a single target.
- **Two positive controls:** `make verify-cross-compile` still builds + vets **4/4** targets under the hermetic env (the block removes only ambient inputs, not a recipe's intentional inline per-target override); and, in a clean env, `make tidy`/`make fmt` leave `go.mod`/`go.sum` unchanged and produce no source diff.

**D8 — Recorded residuals.**

- **R1** — a bare `go test ./...` / `go build` run **without `make`** stays non-hermetic (hermeticity is a `make`-boundary property, not a toolchain property). This includes the gate's documented **direct** invocation's outer `go test`/`go vet` (D6).
- **R3** — the escape hatch, **per variable class**: an already-`export`ed name (`GOENV`, `GOWORK`) is overridden by a command-line assignment (`make GOENV=<file> verify` — the file then wins, and it re-injects a whole env file, not one variable); the **unexported** names (`GOFLAGS`, `GO111MODULE`, `GOEXPERIMENT`, `GOTOOLCHAIN`, `GOOS`/`GOARCH`/`GOARM*`/micro-arch, `GOFIPS140`, `GODEBUG`) have **no in-make hatch** — the hatch is to invoke the toolchain directly (`go …`) or edit the block. `GOENV=off` also means a *legitimate* operator `go env -w` setting is ignored by `make` targets; an explicit exported variable per invocation (`GOPROXY=… make verify`) is the hatch for a plumbing name.
- **R4** — inside the **frozen** round-042 guard, `childEnv` **drops but does not re-set** most of the D2 set, so on the **direct** path an **env-file**-supplied value for those still reaches the gate's child `go list` (the round-042 F-2 shape *inside* the guard). The **non-covered list is exact and closed** — the **14** D2 names `childEnv` does not re-set: `GOENV`, `GOEXPERIMENT`, `GOTOOLCHAIN`, `GOFIPS140`, `GODEBUG`, and the micro-arch family `GOARM`/`GOARM64`/`GOAMD64`/`GO386`/`GOMIPS`/`GOMIPS64`/`GOPPC64`/`GORISCV64`/`GOWASM` (every D2 name except `GOFLAGS`/`GO111MODULE`/`GOWORK`/`GOOS`/`GOARCH`, which `childEnv` re-sets explicitly). Recorded, not fixed (round 042 is frozen); a future round may extend `childEnv` by citation.
- **R5** — the boundary governs recipes + descendants; parse-time `$(shell …)`/`$(eval …)`/command-line variables are outside it (D1). Keep `go` out of `$(shell …)`. A **future** Go env name outside the D2 criterion's categories is out of scope until added (the criterion is the rule for deciding).
- **R6** — make's **`-e` mode** (`make -e`, or an ambient `MAKEFLAGS=-e` / `GNUMAKEFLAGS=-e`) lets the ambient environment **override** the makefile, so it **resurrects the two `export`ed names** (`GOENV`, `GOWORK`) — reproduced: `GOENV=<file: GOFLAGS=-mod=vendor> … make -e vet` ⇒ **exit 2** (issue #96's reproduction reinstated); the **17 `unexport`ed names are unaffected**. This narrows D5's "unconditional" to **ambient *Go-env* input** — `-e` is a **make-mode** input. Recorded, **not** closed (PR #99 review **N-5**): **option (ii)** (`override export GOENV := off`) would close it but destroys the R3 escape hatch (`make GOENV=<file> …`), and **option (iii)** (a fail-loud `$(error …)`) adds a new failure mode — both are a future decision, not this round's. **Breadth of the make-level class** (PR #99 review **N-6**, reproduced): an ambient **`MAKEFILES`** pointing at a makefile with `override export GOENV=<file>` also re-enables it (contrived — it needs an authored `override` file, unlike `-e`, which resurrects whatever ambient value already exists); an `override export GOFLAGS=…` injected the same way is **still defeated** by the `unexport`; and **`MAKEOVERRIDES` has no channel** here because the `Makefile` never recurses via `$(MAKE)`.

**D9 — Governance.** The policy is recorded here and in `specs/truth/techstack.md` (Build & Tooling: a **Hermetic toolchain invocation** row + a note on the **Task runner** row). This ADR is the citable home; `techstack.md` is the current-state truth.

## Alternatives considered

1. **Option B — document the requirement + a loud self-diagnosing guard** — rejected: it does not make the gates hermetic; the operator must still clean the environment.
2. **Option C — accept as a non-goal and record it** — rejected: leaves the fragility in place and contradicts round-042's "genuinely host-free" posture.
3. **A per-invocation `$(HERMETIC_ENV)` prefix on every recipe (A2)** — rejected: ~16 executed invocation sites (25 `grep` lines; PR #97 review **R-3**), a new recipe silently misses it, and it cannot reach a descendant `go build` spawned inside a recipe's shell script.
4. **Neutralise only the module/work vars, leaving `GOOS`/`GOARCH`/`CGO_ENABLED` ambient (V3)** — rejected: a stale exported `GOOS` on a dev box would still miscolour a host build.
5. **Globally pin `CGO_ENABLED=0` (V2)** — rejected: a real behaviour change (excludes future cgo-tagged code) that violates the round's own "no behaviour change for a clean environment."
6. **Gate-family-only scope (S2)** — rejected: fights the single-boundary mechanism and silently ambient-ises future targets.
7. **Delete `tools/arch` `childEnv` (one owner)** — rejected: it would break the gate's documented direct invocation's *verdict* and reopen frozen round-042 history.
8. **A `Makefile`↔Go shared set definition** — rejected: a new coupling mechanism disproportionate to a handful of literals.
9. **Unset `GOFLAGS` only, preserve `GOENV`** — rejected: defeated by the env-file fallback for the persisted case (the issue's headline reproduction would stay red).
10. **An **incident-derived** 8-name set, or an **equality** invariant between the two sites** — rejected (PR #97 review B-3/B-4): the set must follow the D2 criterion (it now includes `GOTOOLCHAIN`, the micro-arch family, `GOFIPS140`, `GODEBUG`) and the sites' invariant is **coverage**, not equality.
11. **Keep a "drift witness" that the two sets agree** — rejected: it would fail by design (the mechanisms differ); the coverage invariant is recorded as a maintenance invariant (D6), not a witness.
12. **A `HERMETIC=0` opt-out switch** — rejected: redundant with (and contradictory to D5's) command-line escape hatch.

## Consequences

- **New:** one top-of-`Makefile` `export`/`unexport` block; **ADR 0012** + its index row; a `techstack.md` (Build & Tooling) **Hermetic toolchain invocation** row + a note on the **Task runner** row.
- **No product code**, **no new dependency** (`go.mod`/`go.sum` unchanged); **no new target** (the existing targets change only how they launch the toolchain); POSIX-only.
- **No behaviour change for a clean environment:** every gate's verdict and every target's observable behaviour is unchanged, and a recipe's intentional inline per-target override still wins (D4/D7).
- **The neutralise set is criterion-derived (D2), not a fixed count** — a future name is adjudicated by the D2 criterion, and adding one is an ADR amendment (this ADR is immutable once Accepted except its `Status` line + the index → a superseding ADR or an accepted amendment round). *This ADR was folded once before merge (PR #97 review B-1…B-4 · TD-1 · R-1…R-3), which is why D2 carries the criterion and D7 the corrected witness count.*
- **Recorded residuals:** R1 (direct path / bare `go`), R3 (hatch per variable class), R4 (`childEnv`'s non-re-set names on the direct path), R5 (`$(shell …)`/parse-time + future names).
- **Forward items:** extend `childEnv` with the D2 names when round-042's guard is next touched (R4). The pre-existing `make help` `verify-no-network` text tidy-up (round-032 **R8d**) was **RESOLVED** in round 085 (`085-dsl-topology-reconciliation`; closes [#174](https://github.com/gosharplite/tellme/issues/174)) — the help line + the target header comment now describe the shipped **offline-path witness**, and the `techstack.md` R8d bullet is retired.
- **Scope note:** this round is `043-*`; **R2 of [#92](https://github.com/gosharplite/tellme/issues/92)** (composition-root extraction) slides to `044-*` (the `STATUS.md` roadmap line is updated at closeout).
- **Atomic delivery:** the `Makefile` block + ADR 0012 + the truth row land in **one** delivery; the witnesses and positive controls are reproduced at the merged head.
- **Witnesses:** the four red→green cases + the neutralisation assertions + `verify-cross-compile` 4/4 + `tidy`/`fmt` no-diff — reproduced, then reverted (ADR 0010 doctrine).
