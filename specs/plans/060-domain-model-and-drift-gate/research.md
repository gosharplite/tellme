# Phase 0 Research: tellme domain model + drift gate (Round 060)

**Plan package**: `specs/plans/060-domain-model-and-drift-gate`
**Anchor**: operator request (no anchor issue) — *"I want tellme to have domain model."*
**Clarify**: round 1 closed — Q1 → 3 (three models), Q2 → 1 (adopt the modelith fork + `make modelith-*` targets), Q3 → 1 (`modelith-check` is a zero-tolerance `make verify` member; an absent binary hard-fails). Q4/Q5 converged as assumptions A6/A7.
**Shape**: **ADD** a descriptive domain model (`docs/domain-model/**`) + a modelith toolchain (`make modelith-lint|render|check`) + a drift gate (a `make verify` member) + the truth/ADR records. **No product code; `go.mod`/`go.sum` unchanged.**

---

## Decision 1: Adopt the modelith fork; the YAML is the source, the `.md` is generated (clarify Q2 → 1)

The three models are authored as `*.modelith.yaml` **sources** and **rendered** to `*.modelith.md` (Markdown + embedded Mermaid ER diagram) by **modelith** — the `gosharplite/modelith` fork at `@feat/self-domain-model` (a fork of `stacklok/modelith`). Authoring tooling = the three pre-loaded skills `domain-model-author` (build/update by conversation) · `domain-model-context` (load the model into a session) · `domain-model-lint` (read-only review). The `domain-model-author` **3-pass build order** is followed: **1 Skeleton** (name every entity, a crisp 2–4-sentence `definition`, `relationships` + `cardinality`) → **2 Behaviour** (`invariants` + `scenarios` that exercise every entity — where the real behaviour lives) → **3 Refinement** (`attributes`, `enums`, `actions`, `glossary`, `ownership` — only where it adds clarity). Conventions: PascalCase entity keys; backticked entity names in freeform text; `cardinality` ∈ {`1:1`,`1:n`,`n:1`,`n:n`}; `ownership` ∈ {`owned`,`referenced`}.

## Decision 2: File locations — three models under `docs/domain-model/`

| Model | Path | Subject |
| --- | --- | --- |
| Product | `docs/domain-model/tellme.modelith.yaml/.md` | the shipped `tellme` (entities below) |
| Quality | `docs/domain-model/quality.modelith.yaml/.md` | tellme's quality process (`make verify` gates, E2E, topology audit, ADR governance, triage) |
| Environment management | `docs/domain-model/environment-management.modelith.yaml/.md` | the **external** Niffler manager (`tellme.sh`) + `ait-<tag>` environments + personas + hot-swap |

The **environment** model is kept **beside** the others (not in a `docs/architect/environments/` subtree as the reference does) — tellme has **no** `docs/architect/**` tree, and the trio reads as one "domain model" folder. Recorded divergence: this model describes an **external** system (the Niffler `tellme.sh` at `…/mbp-johndoe-niffler/` / `…/beta-niffler/`), not part of this repo's runtime — noted in the model's `description` + this research + the truth row.

## Decision 3: The product model's entity set (the shipped surface)

From the shipped tree (`internal/**`, `cmd/tellme`), the product model MUST cover, at minimum: **`Session`**, **`Turn`**, **`Provider`** (+ the family/registry shape), **`Tool`**, **`ToolCall`**, the **context/`Context`** assembly, **`History`** (+ turns log), **`Skill`**, **`MCP` server/tool**, **`Config`**, and the **presentation/Chrome** surface (turn chrome, `[Tool …]` blocks, payload status lines, spinner, colour). It MUST NOT carry the reference's deliberately-excluded concepts — `SafePath`, `SecurityManager`, `UserInteractor`, Windows paths, `pipe_commands` (README *Design Intent & Direction*; ADR lineage rounds 008/021). Invariants where the behaviour lives: e.g. the **universal *no reason, no go* gate** (ADR 0025), the **payload status** shape (ADR 0022/0027), the **terminal-only colour gate** (ADR 0023/0027/0028), the **1 Hz resource-sample** cadence (ADR 0029).

## Decision 4: The quality model's subject (tellme's *actual* process)

The quality model MUST model tellme's process, **not** a copy of the reference's: the `Makefile` **gate catalog** (`verify-no-test-sleep` · `verify-no-network` · `vet` · `verify-cross-compile` 4/4 · `verify-mcp-sdk-confinement` · `verify-architecture` (the Go-guard form + baseline) · `lint` · `vulncheck`, and the `test` target incl. the godog E2E + `test-fast`), the Gherkin/DSL **topology audit**, the **ADR governance**, and the **triage loop**. **Recorded divergence:** tellme has **no** `NonFixCatalog` (no `docs/architect/INTENTIONAL_NON_FIXES.md`) and, *before this round*, **no** modelith gate. The quality model documents tellme's reality — it must not claim a `NonFixCatalog` tellme does not have.

## Decision 5: Makefile wiring + the absent-binary policy (clarify Q3 → 1)

Three targets, POSIX-only, mirroring the reference's form but **without** its `go run …@branch` fallback (which would fetch over the network — non-hermetic):

```
MODELITH := $(shell command -v modelith 2>/dev/null)
MODELITH_MODELS := docs/domain-model/tellme.modelith.yaml \
                   docs/domain-model/quality.modelith.yaml \
                   docs/domain-model/environment-management.modelith.yaml
modelith-lint:   # for m in $(MODELITH_MODELS): modelith lint $$m
modelith-render: # for m in $(MODELITH_MODELS): modelith render $$m
modelith-check:  # require $(MODELITH) non-empty (else: named install instruction, exit 1);
                 # for m: modelith render --check $$m
```

`modelith-check` is added to the **aggregate `verify`** target (Q3 = 1): **zero-tolerance** — drift ⇒ fail; **absent `modelith` ⇒ fail** naming `go install github.com/gosharplite/modelith/cmd/modelith@feat/self-domain-model`. The gate **never** silently skips (no warn-tier). Verified subcommands (`modelith --help`): `lint` (structural/semantic/completeness) · `render` (`--check` = non-zero on drift). Recorded consequence: a host running `make verify` MUST have the modelith dev tool installed (README documents the one-line install).

## Decision 6: Fork pinning + the version-drift residual

The fork branch is **unmerged**; a rendering change could make a committed `.md` look stale **without** a YAML edit. This round **pins** the install form (`@feat/self-domain-model`) + the observed version (`v0.0.0-20260815121344-b4153541cee8`) in a `docs/domain-model/README.md` row and the truth row. **Residual (forward item):** a fork-branch move may require a re-render; the drift gate makes that a **visible red**, not silent rot. Pinning to a commit hash is a **recorded forward option**, not adopted now.

## Decision 7: Governance — ADR 0030 + the technology-stack truth (clarify Q2/Q3)

**ADR 0030** (`docs/decisions/0030-domain-model-and-modelith-toolchain.md`) records: (a) the adoption of the modelith fork + the three models; (b) the **amendment of ADR 0011 D10** — its "no modelith toolchain (no `modelith-layers` analogue)" position is **superseded for the model + drift gate** (tellme now *has* a modelith toolchain, but still ships **no** `modelith-layers` architecture gate — the Go-guard `verify-architecture` is retained); (c) the gate policy (zero-tolerance `verify` member; absent-binary hard-fail). Index row added to `docs/decisions/README.md`. `specs/truth/techstack.md` MODIFY: a new **Domain model** row (the three models + the toolchain + the drift gate) and the **Task runner** `verify` aggregate gains `modelith-check`. The ADR 0011 *Layer-discipline gate* row's recorded divergence line is corrected in the same fold (it currently claims "no modelith toolchain").

## Decision 8: Authority boundary — descriptive docs, truth wins (clarify Q4 → A6/FR-008)

The models are **descriptive docs**, **not** AIxBDD `TruthArtifact`s (they are not under `specs/truth/**`, have no single-owner skill, and are not gate-consumed as truth). On any conflict with the truth tree (`specs/truth/techstack.md`, `specs/truth/data/data-model.dbml`, `specs/truth/features/cli/**`), **truth wins** and the model is corrected. The drift gate protects *model-internal* consistency (YAML ↔ rendered `.md`), **not** model-vs-code (that is the reference's advisory `modelith-drift`/`modelith-layers`, deliberately **not** adopted — out of scope).

## Decision 9: Lifecycle (clarify Q5 → A7)

No separate scheduled refresh pass. The model is refreshed **alongside a round's truth changes** (the truth owner that already edits `techstack.md`/the CLI features updates the corresponding model section in the same PR); the drift gate is the safety net. The model is not a plan package and has **no** `delivered` freeze (it represents the *current* system, like truth).

## Decision 10: Non-BDD tooling round — interface planners are NOOP

Precedent: rounds 020/031/041/042/043/055. `/axb-spec-by-example` = **NOOP** (no user-facing business journey) · `/axb-api-plan` = **NOOP** (no HTTP surface) · `/axb-data-plan` = **NOOP** (the model is a docs artifact, not runtime/persisted state) · `/axb-dsl-refine` = **NOOP** (the model + gate are a **dev/docs surface**, not the `tellme` CLI contract; no new Gherkin/DSL rows, topology audit unchanged) · `/axb-ui-plan` skipped. The **acceptance carrier** is the model + `modelith lint` + `modelith render --check` + the drift gate — **not** "the E2E suite is green" (the round-009 *"green suite = false confidence"* trap).

## Decision 11: No new dependency; POSIX-only; `go.mod`/`go.sum` unchanged

`modelith` is a **dev-tool binary** prerequisite (like `golangci-lint`/`govulncheck`), **not** a Go module dependency. `go.mod`/`go.sum` stay unchanged. POSIX-only (Linux/macOS) — consistent with tellme's POSIX-only direction.

## Decision 12: Testing / BDD techstack unchanged

No new test framework, no new gate semantics beyond adding a `verify` member. `make verify` (incl. all pre-existing members), the godog E2E, and the Gherkin/DSL topology audit MUST stay green.

---

## Residual risks / forward links

- **RF-060-1** — fork-branch version drift (a re-render may be needed on a fork update); pinning to a commit hash is a recorded forward option (Decision 6).
- **RF-060-2** — the environment model describes an **external** system (the Niffler `tellme.sh`), not this repo's runtime; it is a model *about* it, recorded as a divergence (Decision 2).
- **RF-060-3** — the reference's advisory `modelith-drift` / `modelith-layers` (code↔model alignment) are **not** adopted; tellme's model is descriptive docs, gated only for internal drift (Decision 8).
- **RF-060-4** — the quality model documents tellme's reality, which **diverges** from the reference (no `NonFixCatalog`; a lighter gate catalog) — divergence recorded, not smoothed over (Decision 4).
- **RF-060-5** — `make verify` now requires the modelith dev tool; a host without it fails the gate by design (Q3 = 1).

## Not applicable here

- `/axb-spec-by-example`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`, `/axb-ui-plan` — all **NOOP** (Decision 10).
- No product code, no flag/exit-code/stream change, no new `go.mod` dependency, no new Gherkin/DSL row.
