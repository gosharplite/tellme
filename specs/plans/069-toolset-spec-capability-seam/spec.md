# Feature Specification: the `ToolSetSpec` capability seam — replace the positional registry scalars (round 069)

**Feature Branch**: `069-toolset-spec-capability-seam`

**Created**: 2026-09-20

**Status**: Draft (specified — clarify **not escalated**, see §Clarify strategy)

**Anchor**: issue [#140](https://github.com/gosharplite/tellme/issues/140) — **the round's DoD is closing it.**

**Input (operator, 2026-09-20, this session)**:

> *"Create a detail new issue for this. Open round 069, the goal is to close this new issue."*

**Behaviour intent**: **MODIFY (a construction seam; no behaviour change)** — replace the agent tool-registry construction's accumulated **positional scalars** (`sink domaintools.OutputSink, vision bool, providerType string`) with **one named, self-describing `ToolSetSpec`** (the capability set + the resolved limits). **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Honest scope note (read first)

- **This is a pure internal-shape refactor.** **No config-file change** (no key added/removed/renamed/re-defaulted; `VISION` and `TYPE` are read exactly as today) and **no UX change** (the offered tool set, `read_image` behaviour, the chrome, `stdout`, exit codes, and the `--tool-usage` union are all identical). The acceptance is **structural**: the existing pins stay green with **no assertion changed**.
- **RF-062-10 bundles TWO items** — (a) the `ToolSetSpec` seam (F-062-4) and (b) the **media-channel refactor** (return media from `Execute` instead of the per-call `context` collector). **This round lands (a) only**; (b) stays a recorded forward item (bigger blast radius; cohesion-motivated). Do not let the round absorb (b).
- **The seam is recorded overdue**: `NewToolRegistry` gained **one positional scalar per capability round** (`vision bool` in round 062, then `providerType` in round 063) **one round after** F-062-4 named the seam — the PR #130 review annotated it **overdue** and recommended landing the named spec "instead of a fourth scalar".

> **Round-number note.** The number **069** was first assigned to the **retracted** round `069-concurrent-tool-dispatch` (opened in error 2026-09-20, session 52 — never landed; branch deleted; issue [#139](https://github.com/gosharplite/tellme/issues/139) closed `not_planned`). Per the `axb-specify` naming rule (*max existing plan package + 1*; the integrated max was `068`), **069 is the correct next number and is free** — this package reuses it for the `ToolSetSpec` round.

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `5971822`)*

| Site | Current shape |
| --- | --- |
| `internal/app/deps/deps.go:70` | `NewToolRegistry func(sink domaintools.OutputSink, vision bool, providerType string) domaintools.Registry` — **three positional scalars**; the composition-root contract (ADR 0013/0019). |
| `cmd/tellme/deps.go:122` | `newToolRegistry(sink, vision, providerType)` → resolves `ceiling := infratools.ImageCeilingForFamily(infrallm.Family(providerType))` and calls the assembler. |
| `cmd/tellme/deps.go:105` | `assembleAgentTools(sink domaintools.OutputSink, vision bool, imageCeiling int) []domaintools.Tool` — `if vision { append(NewReadImageTool(imageCeiling)) }`. |
| `internal/cli/cli.go:750` | The prompt path: `dp.NewToolRegistry(prog.ToolOutput, res.Provider.Vision, res.Provider.Type)`. |
| `internal/cli/cli.go:998` | `renderToolUsage(env, newToolRegistry func(domaintools.OutputSink, bool, string) domaintools.Registry, …)` — the **anonymous type is duplicated**; builds **two** registries positionally (`(nil,false,\"\")` + `(nil,true,\"\")`) for the capability-aware **union** (round 062 fold F-062-1). |
| `internal/cli/cli.go:367` | Passes `dp.NewToolRegistry` into `renderToolUsage`. |
| `internal/config/config.go:106` | `Vision bool \\`yaml:\"VISION\"\\`` (default false; declared-never-inferred); `Type string \\`yaml:\"TYPE\"\\`` is the provider label. |

**Provenance:** round 062 (**ADR 0032**) added `vision bool` → PR #129 review **F-062-4** → **RF-062-10**; round 063 (**ADR 0033 D7**) added `providerType` → PR #130 review **RF-063-6** (annotated **overdue**).

---

## Design (proposed — the exact shape/placement is a technical choice for `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **One named capability value replaces the positional scalars** on the registry-construction path — a `ToolSetSpec` carrying the `[Tool Output]` sink, the **`vision`** capability gate, and the **resolved** family-aware `imageCeiling`. | **locked (round goal)** |
| **S-2** | **A new capability lands as a FIELD**, never a fourth positional argument — the seam's whole point. | locked |
| **S-3** | **The `renderToolUsage` anonymous function type is replaced by the named signature**, and its two capability variants become **named-field** constructions (no positional `false/true/\"\"`). | locked |
| **S-4** | **Where the type lives** (e.g. `internal/domain/tools` vs the composition-root `internal/app/deps`) and **who resolves** the ceiling (the composition root, as today) — a technical choice for `/axb-technical-research`. | proposed (research) |
| **S-5** | **Whether the spec carries the raw `providerType` or the already-resolved `imageCeiling`** — a technical choice (the resolved-ceiling form removes the `infrallm.Family` call from the builder). | proposed (research) |
| **S-6** | **Behaviour is byte-identical** — the offered set and the ceiling semantics are unchanged; the only delta is the Go signature + call sites. | locked (I-3) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — No config-file change** — no key added/removed/renamed/re-defaulted; every `configs/*.yaml` keeps working byte-for-byte; `VISION`/`TYPE` semantics unchanged.
- **I-2 — No UX change** — the offered tool set is a function of `VISION` alone; `read_image` behaviour identical; chrome/`stdout`/exit codes/`--tool-usage` union unchanged; no new flag, no output change, no new class phrase.
- **I-3 — Behaviour is byte-identical** — the acceptance is structural; the existing pins stay green with **no assertion changed**.
- **I-4 — Scope is the seam only** — the media-channel refactor (RF-062-10's other half) is **out**.
- **I-5 — No new dependency; stdlib-only; POSIX-only; hermetic.**
- **I-6 — The composition-root contract stays honest** — `deps.Dependencies` remains the single construction seam; the layer gate stays green (0-violation baseline).

---

## Clarify strategy

**NOT escalated (0 questions).** The goal is unambiguous — the operator defined it (*"the goal is to close the new issue"*, and the issue is scoped to the `ToolSetSpec` seam only, with the media-channel refactor explicitly out). No gap changes a user story, an acceptance criterion, or a high-impact truth behaviour: the round is a **structural** change whose acceptance is behaviour-identity. The residual decisions (**S-4** where the type lives, **S-5** raw `providerType` vs resolved `ceiling`) are **technical** and defer to `/axb-technical-research`.

**No `NEEDS CLARIFICATION` remains** (none was raised).

---

## User Stories (proposed)

### US1 — A capability lands as a field, not another positional scalar (Priority: P1)

The registry-construction path takes **one named, self-describing value**; adding a capability (a future gate, a third family) is a **new field** on that value, not a fourth positional argument plus a re-spelled function type at each call site.

**Why P1**: it is the round's whole goal — removing the seam that has already grown one scalar per capability round.

**Acceptance (proposed)**: no positional `bool`/`string` capability argument remains on the construction path (`NewToolRegistry` / `newToolRegistry` / `assembleAgentTools` / `renderToolUsage`); their call sites name fields; a new capability would slot in as a field.

### US2 — Behaviour is provably unchanged (Priority: P1)

The refactor is invisible: the offered tool set, `read_image`'s family ceiling, the `--tool-usage` union, the chrome, and the persisted records are all identical.

**Why P1**: a structural round earns its keep only if it changes nothing observable.

**Acceptance (proposed)**: the full existing suite (unit + godog E2E) stays **green with unchanged assertions**; `make verify` green; `go.mod`/`go.sum` unchanged.

---

## Functional Requirements (proposed)

- **FR-001** — `deps.NewToolRegistry` MUST take **one named capability value** (the `ToolSetSpec`) in place of the positional `(sink, vision bool, providerType)` scalars.
- **FR-002** — `cmd/tellme`'s `newToolRegistry` and `assembleAgentTools` MUST take that value; the **resolved** image ceiling is derived there (or carried on the spec — research decides), **exactly** as today.
- **FR-003** — `internal/cli`'s `renderToolUsage` MUST take the **named** registry-constructor signature (no duplicated anonymous function type) and build its two capability variants via **named fields**.
- **FR-004** — A **new capability** MUST be expressible as a **field** on the spec (the seam's invariant); no fourth positional scalar.
- **FR-005** — Behaviour MUST be byte-identical: the offered set, the `read_image` ceiling, the `--tool-usage` union, and the offline paths are unchanged (I-1/I-2/I-3).
- **FR-006** — The **config schema** MUST be untouched (I-1).

## Success Criteria (proposed)

- **SC-001** — No positional capability scalar remains on the construction path (grep-verifiable); the call sites are named.
- **SC-002** — The full suite is green with **no changed assertion** (unit + E2E, incl. `offering-the-agent-tools` / `reading-a-local-image` / `accounting-for-the-tool-use`).
- **SC-003** — `make verify` green; the layer gate stays at the 0-violation baseline; `modelith-check` unchanged; `go.mod`/`go.sum` unchanged.
- **SC-004** — RF-062-10 (the seam half) + RF-063-6 are delivered; the media-channel half stays a recorded forward item.

---

## Edge cases (proposed)

- The **offline `--tool-usage`** path builds **two** capability variants (base + `read_image`) — both must be constructible from the spec with named fields (the union semantics preserved).
- The **nil sink** (the round-031 assembler gate / the offline path) is a legitimate spec value (`Sink == nil`), as today.
- A **non-vision** provider (`VISION: false`) — the spec's gate is false; no `read_image`; the ceiling field is unused.
- A **provider whose label resolves to a family with a different ceiling** (openai-compatible 32 MiB vs gemini 14 MiB) — the resolved ceiling is unchanged.

## Key entities

`ToolSetSpec` (NEW — the named capability value) · `deps.Dependencies.NewToolRegistry` (the contract) · `assembleAgentTools` (the builder) · `renderToolUsage` (the offline call site) · `config.Provider.Vision` / `.Type` (the inputs).

## Assumptions

- A1 — The change is **internal only** (`internal/app/deps` + `cmd/tellme` + `internal/cli`, plus the truth rows); the adapters, the tools, and the config **schema** are untouched.
- A2 — The media-channel refactor is **out of scope** (its own round).
- A3 — The composition root keeps resolving the family ceiling (the spec carries the resolved value, or the builder derives it — research decides).
