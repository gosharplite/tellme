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
It is a **dev-tool binary**, **not** a Go module dependency — `go.mod`/`go.sum` are unchanged.

### Install (the single source of the route)

The fork **declares the upstream module path** (`github.com/stacklok/modelith`), so it is
**not** installable via `go install github.com/gosharplite/modelith/...@<ref>` at **any**
ref — the branch query form is rejected (`invalid version … disallowed version string`) and
the pseudo-version form fails (`module declares its path as: github.com/stacklok/modelith`).
The working route is a **local clone of the fork, checked out at an immutable commit, built
locally** (a local module build is unaffected by the declared-path mismatch):

```sh
git clone https://github.com/gosharplite/modelith && cd modelith && git checkout b4153541cee8 && go install ./cmd/modelith
```

`$(MODELITH_INSTALL)` in the `Makefile` (and the message `make modelith-check` prints) quotes
this same route; the ADR/truth prose cites this section rather than restating it.

- **Pinned identifier (immutable):** commit `b4153541cee8` — the `feat/self-domain-model` tip
  this round was authored against. Branch-tracking is an explicit, documented **upgrade**, not
  the default (a fork move changes the renderer ⇒ a committed `.md` renders differently ⇒
  `make verify` reds on a re-install; that is the observable cost of the pin — TD-060-1).
- **Known-good tool (provenance):** `github.com/stacklok/modelith v0.0.0-20260815121344-b4153541cee8`,
  commit `b4153541cee8`, `vcs.modified=false` (the build the three committed `.md` files were
  rendered with). *Provenance is recorded here; it is not a claim about which tool the gate
  "supports" — the gate runs whatever `modelith` is on `PATH`.*

## Quick commands

```sh
make modelith-lint     # validate every *.modelith.yaml (0 errors / 0 warnings)
make modelith-render   # regenerate every *.modelith.md from its YAML
make modelith-check    # drift gate: fail if a committed .md is stale
make modelith-drift    # ADVISORY (never fails): a modeled entity with no code anchor
```

`make modelith-check` is a **zero-tolerance** member of `make verify`: **drift
fails**, and an **absent `modelith` binary fails** (naming the install route) —
it never silently skips. A host running `make verify` therefore needs the modelith
dev tool installed (ADR 0030 §D3).

`make modelith-drift` is different: **advisory**, **not** a `make verify` member,
**never failing** (see *Drift guard* below).

## Drift guard (advisory)

The model is **load-bearing**: it is *descriptive docs, subordinate to truth*, but
it is **maintained** — a round that changes **modelled behaviour** updates
`docs/domain-model/**` **in the same PR** (ADR 0041). Two guards keep it honest:

| Guard | Kind | What it catches |
| --- | --- | --- |
| `make modelith-check` | **gate** (`make verify` member, zero-tolerance) | the rendered `.md` is **stale** vs its `.yaml` source (YAML↔MD generation drift) |
| `make modelith-drift` | **advisory** (never fails, not a `verify` member) | a **modeled entity whose concept has vanished from the code** (a stale model entry — the round-time rule above is the primary guard; this is its aid) |

`make modelith-drift` (script `scripts/modelith-drift.sh`) checks, for every
modeled **entity**, **enum**, and **glossary** term, whether **any** of its *code
anchors* — its own name, a backticked identifier in its definition, or one of its
enum values — still appears in the production Go sources (comments included;
test files excluded). Zero anchors ⇒ a warning. It is deliberately **lenient**
(one live anchor clears the entry) and **code-model-only**
(`tellme.modelith.yaml`); the quality and environment models are not code-backed.
A term that models a **deliberate absence** (e.g. `NoSecurityLayer`) is
hand-excepted in the script.

**Why there is no name-diff check (the anti-muse record).** The reference ships
advisory `modelith-drift` / `modelith-layers` gates that flag **new** exported Go
identifiers with no model entry. A direct port was **measured** on this repo and
rejected: comparing the model's names against the exported types under
`internal/domain/**` produced **~45 of 57 types flagged (~79 % false positives)** —
ports (`Store`, `Sink`, `Reader`, `Source`, `Prompter`, `Loop`, `Lines`, …) and
value types (`Request`, `Response`, `Result`, `Step`, `Entry`, …) vastly outnumber
*modeled concepts*, and the reverse direction flagged **12 legitimate entities**
(named differently in code, logical enums, behavioural roles). Shipping it would
recreate the **retired** "noisy advisory surface becomes a permanent muse" failure
(the `RF-063-10` / `RF-068-1` class the curation rule exists to prevent). The
forward check above is **precise** (zero findings on the current tree, and it
catches a synthetic stale entry); a name-diff check would need a hand-curated
entity↔symbol manifest, which is a **separate, unwarranted** decision today.

**What neither guard catches (be honest):** *semantic* model rot — the model
describing behaviour the code does not have (e.g. skills modelled as auto-injected
vs the shipped on-demand surface; media modelled as an attached collector vs the
shipped in-band return). That class has **no mechanical carrier**; it is caught by
the **round-time rule** (the truth owner that touches a modelled area updates the
model), backed by review. Keeping the model load-bearing is a **process**
commitment, not a gate.

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
model section in the same PR); the `modelith-check` gate is the YAML↔MD safety
net and `make modelith-drift` is the advisory staleness aid. **A round that
changes modelled behaviour MUST touch `docs/domain-model/**`** (or record, in its
plan package, why the change is not modelled) — the model is load-bearing
(ADR 0041). There is **no** scheduled refresh pass, and the model is **not** a plan
package (no `delivered` freeze) — it always represents the *current* system
(ADR 0030 §D6).
