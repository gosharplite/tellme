# tellme — Domain Model

This folder holds tellme's **canonical domain model** — a plain-language
description of *what the system is*: its entities, relationships, invariants, and
scenarios. It is authored as modelith `*.modelith.yaml` **sources** and rendered to
`*.modelith.md` (Markdown + Mermaid ER diagrams).

> **These are descriptive docs, not truth.** The single source of truth for
> behaviour lives under [`../specs/truth/**`](../specs/truth). On any conflict,
> **truth wins** and the model is corrected. The model is **subordinate** to the
> AIxBDD truth tree (ADR 0030 §D4).

## Files

| File | Role |
| --- | --- |
| `tellme.modelith.yaml` | **Product** model — the shipped system (source; edit this) |
| `tellme.modelith.md` | Rendered Markdown + Mermaid ER (generated — **never edit by hand**) |
| `quality.modelith.yaml` | **Quality process** model — the gate pipeline, ADR governance, triage (source) |
| `quality.modelith.md` | Rendered (generated) |
| `environment-management.modelith.yaml` | **Environment management** model — the external Niffler manager (source) |
| `environment-management.modelith.md` | Rendered (generated) |

## Toolchain

Built with **modelith** — the [`gosharplite/modelith`](https://github.com/gosharplite/modelith)
fork at `feat/self-domain-model` (a fork of [`stacklok/modelith`](https://github.com/stacklok/modelith)).
Install it as a **dev-tool binary** (it is **not** a Go module dependency —
`go.mod`/`go.sum` are unchanged):

```sh
go install github.com/gosharplite/modelith/cmd/modelith@feat/self-domain-model
```

*Observed version on the round-060 host:* `v0.0.0-20260815121344-b4153541cee8`
(the fork branch is unmerged, so a fork update may require a re-render — the
drift gate makes that a **visible red**, not silent rot).

## Quick commands

```sh
make modelith-lint     # validate every *.modelith.yaml (0 errors / 0 warnings)
make modelith-render   # regenerate every *.modelith.md from its YAML
make modelith-check    # drift gate: fail if a committed .md is stale
```

`make modelith-check` is a **zero-tolerance** member of `make verify`: **drift
fails**, and an **absent `modelith` binary fails** (naming the install command) —
it never silently skips. A host running `make verify` therefore needs the modelith
dev tool installed (ADR 0030 §D3).

## Authoring conventions

- **Entity keys** are PascalCase (`Session`, not `session`).
- **Backtick entity names** in freeform text (definitions, notes, invariants,
  scenario steps): `` `Session` ``. Do **not** backtick in structured fields that
  already imply an entity (`actors`, relationship `entity:`, entity keys).
- **`cardinality`** ∈ {`1:1`, `1:n`, `n:1`, `n:n`}; **`ownership`** ∈ {`owned`,
  `referenced`}.
- A **plain scalar must not begin with a backtick** — quote it (`` statement:
  "`X` …" ``) or reword.
- Follow the **3-pass build order**: **1 Skeleton** (name every entity, a crisp
  definition, relationships + cardinality) → **2 Behaviour** (`invariants` +
  `scenarios`) → **3 Refinement** (`attributes`, `enums`, `actions`, `glossary` —
  only where it adds clarity). The value is in the questions, not the typing.

## Lifecycle

The model is refreshed **alongside a round's truth changes** (the truth owner that
edits `specs/truth/techstack.md` or the CLI features updates the corresponding
model section in the same PR); the drift gate is the safety net. There is **no**
scheduled refresh pass, and the model is **not** a plan package (no `delivered`
freeze) — it always represents the *current* system (ADR 0030 §D6).
