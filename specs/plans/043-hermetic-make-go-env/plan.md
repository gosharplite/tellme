# System Analysis Plan — round 043 (`043-hermetic-make-go-env`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/043-hermetic-make-go-env/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks (NOT this branch)

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Build & Tooling) ✓ done

docs/decisions/
├── 0012-hermetic-make-go-env.md   # governance — ADD (the neutralise/preserve policy) ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (the change is a dev surface, not the `tellme`
CLI contract). No `contracts/**` change — `/axb-api-plan` **NOOP**. No `data/**` change —
`/axb-data-plan` **NOOP**. No `ui/**` artifact.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
Makefile                            # CHANGED — one top-of-file hermetic `export`/`unexport` block (above the `$(shell command -v …)` probes)
docs/decisions/0012-hermetic-make-go-env.md    # NEW — the neutralise/preserve policy; index row in docs/decisions/README.md
specs/truth/techstack.md            # MODIFY — Build & Tooling: a Hermetic toolchain invocation row + a note on the Task runner row ✓ done
tools/arch/arch_test.go             # a cross-reference COMMENT only (behaviour unchanged — round-042 frozen history) — FR-008 MUST (not optional)
go.mod / go.sum                     # unchanged — no new dependency
internal/** , cmd/** , tests/**      # unchanged — no product code / no CLI behaviour change
```

**Structure Decision**: Round 043 is a **build/quality-pipeline change**, not a runtime-interface change.
It adds **one hermetic invocation boundary** to the `Makefile` (a top-of-file `export`/`unexport` block)
that neutralises the build-context inputs an ambient/persisted Go env can inject, while preserving the
warm-cache and network/checksum inputs a cold-cache/proxied build needs. There is **no** new target, **no**
new endpoint, **no** persisted state, **no** CLI behaviour change, and **no** new dependency, consistent
with `research.md` D1–D12. The truth changes are `specs/truth/techstack.md` (Build & Tooling) + the
governance **ADR 0012**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round changes the **build / quality pipeline**,
which is **not** a system boundary (backend / frontend / CLI): the `tellme` CLI end's observable
behaviour is unchanged, so there is no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (no persisted/in-runtime state; the change is build tooling).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the change is a dev surface (`make`), not the `tellme` binary's CLI contract).
> - `/axb-ui-plan` = **skipped** (no UX surface).
> - `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — a build-tooling hardening is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's truth changes (`techstack.md` +
ADR 0012) are RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/043-hermetic-make-go-env`; truth root
`specs/truth`; truth-delta `specs/plans/043-hermetic-make-go-env/truth-delta.md`; interfaces: **none**;
the round's delivery is the `Makefile` hermetic block + ADR 0012 + the `techstack.md` truth rows.

---

### The boundary the round installs

**Neutralise/replace** — derived from the **D2 inclusion criterion** (*neutralise the ambient build
context; preserve the plumbing*): `GOENV` → `off`; `GOWORK` → `off`; `GOFLAGS`, `GO111MODULE`,
`GOEXPERIMENT`, `GOTOOLCHAIN`, `GOFIPS140`, `GODEBUG` → **unset**; the ambient **target triple** —
`GOOS`, `GOARCH` **and** the micro-architecture family (`GOARM`/`GOARM64`/`GOAMD64`/`GO386`/`GOMIPS`/
`GOMIPS64`/`GOPPC64`/`GORISCV64`/`GOWASM`) → **unset**.

**Preserve** — `PATH`, `HOME`, `GOPATH`, `GOMODCACHE`, `GOCACHE` (warm-cache) **and** `GOPROXY`,
`GOSUMDB`, `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE` (network/checksum).

**`CGO_ENABLED`** — **preserved from the caller** (the block neither sets nor unsets it — *not* globally
pinned); the only cgo pinning remains `verify-cross-compile`'s inline `CGO_ENABLED=0` **for its own
recipe** (PR #46 TD1), which the block MUST NOT clobber (RULE: a recipe's inline assignment wins for that
recipe).

**Why `GOENV=off` is load-bearing**: Go falls back to the env **file** for a variable that is unset **or
empty**, so unsetting `GOFLAGS` does not neutralise a persisted `go env -w GOFLAGS=-mod=vendor` — the
issue's exact failing case (round-042 F-2, ADR 0011 D5). Policy: **ADR 0012**.

**Ownership (two mechanisms, one invariant).** The `Makefile` block is the **primary owner**; round-042's
`tools/arch` `childEnv` (ADR 0011 D5) is retained as **defence-in-depth** for the gate's **verdict** on its
documented **direct** invocation (which bypasses `make`; the direct path's **outer** `go test`/`go vet`
remain non-hermetic — R1). Cross-reference comments in both places. The two sites neutralise by
**different mechanisms** (the block *disables the env file* + unsets; `childEnv` *re-sets explicit values*),
so their invariant is **coverage** — *every neutralised name is re-set by `childEnv` or recorded as a non-covered class
(R4)* — **not** equality (PR #97 review B-3); R4's list is exact and closed (14 names).

**Boundary limit**: `export`/`unexport` govern recipes + descendants, **not** parse-time `$(shell …)`/
`$(eval …)`/`include`d makefiles/command-line variables — keep `go` out of `$(shell …)` (TD-1/R5).

**Verified by**: **four red→green witnesses** (`GOENV=<file: GOFLAGS=-mod=vendor>`, `GO111MODULE=off`,
a stray `GOWORK`, `GOTOOLCHAIN=go1.99.9`; red → green, reverted) + **neutralisation assertions** for
green-before inputs (`GOFLAGS=-trimpath` + the long-tail names — asserted absent from the recipe env, *not*
witnesses; PR #97 review B-1) + **aggregate-level** evidence (`GOENV=<file> make verify` ⇒ exit 0) + two
positive controls (`verify-cross-compile` 4/4 under the hermetic env; `tidy`/`fmt` no-diff in a clean env).

---

### Gating blockers

*(none — the operator locked the theme (resolve [#96](https://github.com/gosharplite/tellme/issues/96)) and
answered clarify Q1–Q8 one at a time. No open decision gates the round. The **plan half is complete**;
`/axb-tasks` and `/axb-implement` are intentionally **out of scope for this branch**.)*
