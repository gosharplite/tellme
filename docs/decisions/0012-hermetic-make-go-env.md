# ADR 0012 — A hermetic `make` Go-toolchain invocation environment

- **Status:** Accepted
- **Date:** 2026-09-18
- **Deciders:** tellme owner
- **Related:** issue [#96](https://github.com/gosharplite/tellme/issues/96) (the anchor — the *outer* half of round-042's F-2) ·
  PR [#95](https://github.com/gosharplite/tellme/pull/95) fold-review #2 (comment `5721537481`, Fold 2 — where the gap was found) ·
  **ADR 0011 D5** (round-042's gate child-env discipline — generalised here) ·
  round-020 `verify-cross-compile` + PR #46 review **TD1** (the `CGO_ENABLED=0` pin for one target — completed here) ·
  round 043 (`specs/plans/043-hermetic-make-go-env` — this ADR's round) ·
  **ADR 0010** (the falsifiability/verification doctrine the witnesses follow).

## Context

Every gate in the `Makefile` invokes the Go toolchain as a **child of `make`**, inheriting the ambient environment. A **persisted** (`go env -w`) or **exported** Go setting therefore reaches not just a gate's own tooling but the **outer** `go test` / `go vet` / `go build` themselves. Issue [#96](https://github.com/gosharplite/tellme/issues/96) reproduces it for `verify-architecture`:

```sh
$ printf 'GOFLAGS=-mod=vendor\n' > /tmp/goenv
$ GOENV=/tmp/goenv make verify-architecture
… inconsistent vendoring …      # exit 2, unrelated to the tree
```

Round 042's gate already filters the environment of **its own** child `go list` (`tools/arch/arch_test.go` `childEnv`, ADR 0011 **D5**), so the gate's *verdict* is hermetic — but the surrounding `go test`/`go vet` that `make` runs are not, and the same is true of **every** existing target. A developer/CI image that exports `GOFLAGS=-mod=vendor` or `-trimpath`, a `GOENV` file, `GO111MODULE=off`, or a stray `GOWORK` can redden the whole `verify` aggregate for reasons unrelated to the tree. Round 020's `verify-cross-compile` pinned `CGO_ENABLED=0` for **one** variable on **one** target (PR #46 TD1); this ADR generalises the concern to a repo-wide rule.

## Decision

**D1 — One hermetic boundary, at the top of the `Makefile`.** The neutralisation is defined **once** — a GNU-make `export`/`unexport` block placed near the top of the `Makefile` (above the `$(shell command -v …)` toolchain probes and before any recipe) — so **every** recipe, and every nested `go` invocation a scripted recipe spawns (e.g. the E2E harness behind `make test`, the witness behind `verify-no-network`), runs with the sanitised environment. A target added later is hermetic **by construction**. The rule is **not** applied per recipe (a `$(HERMETIC_ENV)` prefix per invocation would miss the nested spawns and drift as recipes are added).

**D2 — The neutralise set.** The environment `make` exports to its recipes neutralises:

| Variable | Value | Why |
| --- | --- | --- |
| `GOENV` | `off` | **Load-bearing.** Go falls back to the env **file** for a variable that is unset **or empty**, so unsetting `GOFLAGS` does **not** neutralise a persisted `go env -w GOFLAGS=-mod=vendor` (round-042 F-2). `off` is the documented value that disables the env-file fallback entirely. |
| `GOWORK` | `off` | A stray `go.work`/`GOWORK` changes the build context. |
| `GOFLAGS` | unset | Ambient build flags (e.g. `-mod=vendor`, `-trimpath`) must not apply. |
| `GO111MODULE` | unset | An ambient `off` must not disable modules. |
| `GOEXPERIMENT` | unset | Ambient experiment flags must not apply. |
| `GOOS` / `GOARCH` / `GOARM` | unset | A stale exported target must not redirect a host build (`verify-cross-compile` sets its own per target). |

**D3 — The preserve set.** Passed through unchanged, because a cold-cache or proxied build legitimately needs them:

- **warm-cache / runtime:** `PATH`, `HOME`, `GOPATH`, `GOMODCACHE`, `GOCACHE`;
- **network / checksum:** `GOPROXY`, `GOSUMDB`, `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE`.

*(An `env -i`-style wipe would fail with a *module cache not found* / resolution error — the `verify-cross-compile` PR #46 TD1 lesson, applied to the parent.)*

**D4 — `CGO_ENABLED` is left at the host default; it is not globally pinned.** A global `CGO_ENABLED=0` would be a real **behaviour change**: it would exclude any future cgo-tagged code (the recorded darwin `mach` CPU sampler is exactly such a candidate) from `make build`/`make test`, and diverge from a plain `go build`. The only cgo pinning remains `verify-cross-compile`'s **inline** `CGO_ENABLED=0`, which applies to its own recipe only — and the hermetic block **must not** clobber it (a recipe's inline assignment wins for that recipe).

**D5 — Scope: unconditional, all targets.** The block applies to `build`, `fmt`, `vet`, `tidy`, `test`, `lint`, `staticcheck`, `vulncheck`, every `verify-*`, and `verify` — no per-target opt-out. The two **mutating** targets (`fmt`, `tidy`) are covered and verified harmless (D7).

**D6 — Ownership: the `Makefile` block is the primary owner; round-042's `childEnv` is retained as defence-in-depth.** The gate (`tools/arch`) is documented as runnable **directly** (`go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch`), which **bypasses `make`**; its own `childEnv` filter therefore stays, covering that path. Cross-reference comments in both places name the primary owner; the two definitions MUST NOT drift silently (a drift witness — the two literal sets agree — is a folded option).

**D7 — Verification.** The acceptance evidence is (a) **four hostile-env witnesses** — `GOENV=<file: GOFLAGS=-mod=vendor>`, `GOFLAGS=-trimpath`, `GO111MODULE=off`, a stray `GOWORK=<path>` — asserted **red before** the change and **green after** (reproduced then reverted, ADR 0010 doctrine); and (b) **two positive controls**: `make verify-cross-compile` still builds + vets **4/4** targets under the hermetic env (proving the block removes only ambient inputs, not a recipe's intentional inline per-target override), and, in a clean env, `make tidy`/`make fmt` leave `go.mod`/`go.sum` unchanged and produce no source diff.

**D8 — Recorded residuals.** **(R1)** A bare `go test ./...` / `go build` run **without `make`** stays non-hermetic — hermeticity is a **`make`-boundary** property, not a toolchain property. **(R3)** `GOENV=off` also ignores a *legitimate* operator `go env -w` setting (e.g. a deliberately pinned `GOPROXY`); the escape hatch is an **explicitly exported** variable per invocation (`GOPROXY=… make verify`).

**D9 — Governance.** The policy is recorded here and in `specs/truth/techstack.md` (Build & Tooling: a **Hermetic toolchain invocation** row + a note on the **Task runner** row). This ADR is the citable home; `techstack.md` is the current-state truth.

## Alternatives considered

1. **Option B — document the requirement + a loud self-diagnosing guard** — rejected: it does not make the gates hermetic; the operator must still clean the environment.
2. **Option C — accept as a non-goal and record it** — rejected: leaves the fragility in place and contradicts round-042's "genuinely host-free" posture.
3. **A per-invocation `$(HERMETIC_ENV)` prefix on every recipe (A2)** — rejected: ~20 sites, a new recipe silently misses it, and it cannot reach a nested `go build` spawned inside a recipe's shell script.
4. **Neutralise only the module/work vars, leaving `GOOS`/`GOARCH`/`CGO_ENABLED` ambient (V3)** — rejected: a stale exported `GOOS` on a dev box would still miscolour a host build.
5. **Globally pin `CGO_ENABLED=0` (V2)** — rejected: a real behaviour change (excludes future cgo-tagged code) that violates the round's own "no gate's behaviour changes for a clean environment."
6. **Gate-family-only scope (S2)** — rejected: it fights the single-boundary mechanism and silently ambient-ises future targets.
7. **Delete `tools/arch` `childEnv` (one owner)** — rejected: it would break the gate's documented direct invocation and reopen frozen round-042 history.
8. **A `Makefile`↔Go shared set definition (D3)** — rejected: a new coupling mechanism disproportionate to a handful of literals.
9. **Unset `GOFLAGS` only, preserve `GOENV`** — rejected: defeated by the env-file fallback for the persisted case (the issue's headline reproduction would stay red).

## Consequences

- **New:** one top-of-`Makefile` `export`/`unexport` block; **ADR 0012** + its index row; a `techstack.md` (Build & Tooling) **Hermetic toolchain invocation** row + a note on the **Task runner** row.
- **No product code**, **no new dependency** (`go.mod`/`go.sum` unchanged); **no new target** (the existing targets change only how they launch the toolchain); POSIX-only.
- **No behaviour change for a clean environment:** every gate's verdict and every target's observable behaviour is unchanged, and a recipe's intentional inline per-target override still wins (D4/D7).
- **Recorded residuals:** bare-`go`-outside-`make` (R1); the `GOENV=off` env-file escape hatch (R3); `CGO_ENABLED` on a cgo-less host (not one of #96's cases — see `research.md`).
- **Forward items:** an optional drift witness that the `Makefile` set and `tools/arch` `childEnv` set agree; an optional `HERMETIC=0` opt-out; the pre-existing `make help` `verify-no-network` text tidy-up (round-032 R8d), unrelated.
- **Scope note:** this round is `043-*`; **R2 of [#92](https://github.com/gosharplite/tellme/issues/92)** (composition-root extraction) slides to `044-*` (the `STATUS.md` roadmap line is updated at closeout).
- **Atomic delivery:** the `Makefile` block + ADR 0012 + the truth row land in **one** delivery; the witnesses and positive controls are reproduced at the merged head.
- **Witnesses:** the four hostile-env cases ⇒ red before / green after; `verify-cross-compile` 4/4; `tidy`/`fmt` no-diff — reproduced, then reverted (ADR 0010 doctrine).
