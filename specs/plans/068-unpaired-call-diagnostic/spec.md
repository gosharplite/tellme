# Feature Specification: surface a Gemini/Vertex round's unpaired tool calls as a diagnostic (round 068)

**Feature Branch**: `068-unpaired-call-diagnostic`

**Created**: 2026-09-20

**Status**: Draft (specified + **clarify RESOLVED — Q1 → A · Q2 → A**)

**Input (operator, 2026-09-20, this session)**:

> *"Open 068-*, the goal is to resolve RF-067-1."*

**Forward item**: **ADR 0037 §Forward RF-067-1** — *"the unpaired-call observability is a returned-value accessor with **no live consumer**; the boundary drop stays silent at runtime. Surfacing it as a user-visible `[Tool …]`/`stderr` diagnostic is deferred (no adapter logging seam; a scope addition requiring operator confirmation)."* The operator has now confirmed the scope addition — this round lands it.

**Behaviour intent**: **MODIFY (add a user-visible diagnostic)** — when a Gemini/Vertex round's model turn requests **N** tool calls but the round produces **M < N** results, the unpaired calls are today dropped from the wire **silently** (ADR 0035/0036; the round-067 `UnpairedCallIDs` accessor accounts for them **in code** but has no live consumer). This round surfaces that accounting as a **user-visible `[Tool …]`-class diagnostic on `stderr`**, gated like the rest of the chrome. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Honest scope note (read first)

The **shipped agent loop appends exactly one `tool` result per requested call** (`internal/agent/agentloop.go`: `for _, tc := range resp.ToolCalls { … turn = append(turn, llm.Message{Role:"tool", ToolCallID: tc.ID}) }`). So in production today **M == N always**, and the unpaired-call case (`M < N`) **cannot arise through the shipped paths** — which is exactly why round 067 recorded the accessor as having *no live consumer*. This round therefore lands a **defensive diagnostic**: it fires only when a round's `prior` genuinely carries fewer results than calls — e.g. a hand-built/partial/corrupted resumed conversation, or a **future out-of-order/concurrent dispatch** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3, the ADR 0035/0036 motivation). Its value is **safety-net observability**, not a frequent user-facing signal; the round MUST say so (no over-claiming a live-surface change).

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `e864d9a`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/llm/gemini/client.go` `roundBuilder.flush` / `unpaired()` / `UnpairedCallIDs` | The `M < N` boundary drop is **accountable in code**: `flush()` records the unpaired ids (`dropped`), and `UnpairedCallIDs(prior)` returns them (call order). **No live consumer** — its only referents are the round-067 tests. The emitted batched turn is unchanged (it carries the `M` parts produced). |
| `internal/infrastructure/llm/gemini/client.go` `Complete` | Builds the request body via `requestBody(...)` → `buildContents(...)`. The adapter owns **no logging/stderr seam** (the CLI owns the chrome). |
| `internal/domain/tools/outputsink.go` | An existing **neutral domain port** (`tools.OutputSink`: `Begin`/`Writer`/`End`/`Enabled`) — the `[Tool Output]` block's sink, **construction-time injected** into the command tool (`NewCommandTool(sink)`; ADR 0021), wired at `cmd/tellme/deps.go`. A precedent for an injected diagnostic seam. |
| `internal/cli/cli.go` | The chrome (`stderr`) + the per-session `turns.log` writer (`env.turnsLog`, round 053; ADR 0022) are CLI-owned. The `[Tool Output]` block is **not** routed to `turns.log` (ADR 0022 D5 narrowing). Colour is **terminal-gated** (`chromeColour`; round 054 ADR 0023). |
| `internal/ui/**` | The `[Tool …]` line formatters live here (single-owned, sanitized/capped; ADR 0008/0015/0027/0028). |
| `specs/truth/techstack.md` | the *Vertex/Gemini adapter* row records the round-067 unpaired-call accounting ("no live consumer; surfacing it is the forward item RF-067-1"). |

---

## Design (proposed — the **surface** and **loudness** are the round's 2 clarify questions; the *where* is a technical choice for research)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **When a Gemini/Vertex round yields `M < N` results, the unpaired calls' ids are surfaced as a user-visible diagnostic** (never only a silent/in-code account). This is the change that resolves **RF-067-1**. | **locked (round goal)** |
| **S-2** | **The diagnostic is informational, not a failure** — the turn proceeds and the emitted body is unchanged (the unpaired calls still contribute no part, ADR 0035/0036). It is NOT a terminal provider error. **(Q2 → A, locked.)** | **locked (Q2 → A)** |
| **S-3** | **The surface is a `[Tool …]`-class line on `stderr`, terminal-gated colour (ADR 0023), `stdout` untouched.** **(Q1 → A, locked.)** | **locked (Q1 → A)** |
| **S-4** | **The diagnostic is NOT routed to the per-session `turns.log`** — `-t` output is unchanged (the `[Tool Output]`-block precedent, ADR 0022 D5). **(Q1 → A, locked.)** | **locked (Q1 → A)** |
| **S-5** | **Scope is family-local: the Gemini/Vertex family only.** The OpenAI-compatible family has no round-boundary drop (it emits one `tool` message per `tool_call_id`), so it is untouched (`I-1`). | proposed |
| **S-6** | **The detection seam is a technical choice for `/axb-technical-research`** — e.g. an injected diagnostic sink on the Gemini adapter (the `tools.OutputSink`/ADR-0021 construction-time-injection precedent) vs a loop-side detection via the `LoopObserver` port vs a returned value the caller renders. Bounded by: the adapter owns no logging seam today, the port contracts stay frozen where possible, and the wiring is CLI-owned. | proposed (research) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The OpenAI-compatible wire is byte-preserved** and the OpenAI-compatible adapter is untouched (family-local).
- **I-2 — The round-066/067 Gemini batch shape + id-link are preserved** — a round of N calls still serializes to one `user` turn with the produced `M` parts; the diagnostic adds **no** wire change (it is `stderr`/`turns.log` only).
- **I-3 — The media-free Gemini text path is byte-preserved**; no media loss (ADR 0032/0033).
- **I-4 — `stdout` is untouched** — the answer stream stays byte-exact; the diagnostic is a `stderr` (chrome) line only.
- **I-5 — Replay fidelity holds** (ADR 0036 D3); the diagnostic never changes pairing.
- **I-6 — No new dependency; POSIX-only; hermetic** (ADR 0012). No new config key required.
- **I-7 — `M == N` remains the silent happy path** — the diagnostic fires **only** when `M < N`; a normal round emits nothing new.

---

## Clarify strategy

**Escalated — 2 questions, BOTH RESOLVED (Q1 → A · Q2 → A).** RF-067-1 was *operator-gated* precisely because it adds a **user-visible surface**: per the clarify-escalation rule, the surface, its routing, and its loudness **change the formal acceptance criteria**, so they were asked rather than assumed. Interview format: **one question at a time**.

- **Q1 → A — SURFACE & ROUTING.** The diagnostic is a **`[Tool …]`-class line on `stderr`**, terminal-gated colour (ADR 0023), **never routed to `turns.log`** (the `[Tool Output]`-block precedent, ADR 0022 D5). `stdout` is untouched. *(A distinct `[Warning]` prefix (option C) and the `turns.log` routing (option B) were not taken.)*
- **Q2 → A — LOUDNESS.** **Informational** — the turn proceeds and the emitted body is unchanged; the diagnostic is a warning, **not** a terminal provider-error failure. *(Option B, a loud failure with the frozen provider-error class, was not taken; it stays a recorded forward item.)*

The remaining item (the **detection seam** and exact rendering) is **technical** and defers to `/axb-technical-research` (S-6). **No `NEEDS CLARIFICATION` remains** (both locks are folded in).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - an operator sees when a Gemini round left a call unanswered (Priority: P1)

As the **operator** running `tellme` against a **Gemini/Vertex** provider, when a round's model turn requests **N** tool calls but the round produces **M < N** results, I want the **unpaired calls to be surfaced as a diagnostic** (not silently dropped), so I can see that the round left a call unanswered and that the request body carried fewer responses than calls.

**Why this priority**: it is the whole round (RF-067-1). It is independent and testable through the built binary's `stderr`.

**Independent verification**: drive a Gemini round whose `prior` carries `M < N` (a fixture), assert the diagnostic appears on `stderr` (per Q1's surface), and assert the emitted request body is otherwise unchanged (I-2); a normal `M == N` round emits **no** new line (I-7).

**Acceptance Scenarios** *(Q1 → A · Q2 → A)*:

1. **Given** a Gemini/Vertex round with **N** calls and **M < N** results, **When** the turn builds its request, **Then** a diagnostic naming the unpaired call(s) appears on **`stderr`** (surface per Q1) and the request body is unchanged (I-2).
2. **Given** a round whose every call is paired (**M == N**) — the shipped happy path — **When** the turn runs, **Then** **no** new diagnostic is emitted (I-7).
3. **Given** the media-free / single-call / OpenAI-compatible paths, **When** the turn runs, **Then** `stdout` and the wires are byte-identical to before (I-1/I-3/I-4).

**Functional Requirements**:

- **FR-001**: When a Gemini/Vertex round yields `M < N` results, the adapter/system MUST surface the unpaired call ids as a **user-visible diagnostic** (never only a silent/in-code account).
- **FR-002**: The diagnostic MUST be **`stderr`-only**; `stdout` MUST stay byte-exact (I-4).
- **FR-003**: The diagnostic MUST be **informational** — the turn proceeds and the request body is unchanged (I-2) *(Q2 → A)*.
- **FR-004**: When `M == N` (the shipped path), the system MUST emit **no** new diagnostic (I-7).
- **FR-005**: The diagnostic MUST be **deterministic** (call order) so a witness can assert it.
- **FR-006**: The diagnostic MUST be a **`[Tool …]`-class line on `stderr`** and MUST **NOT** be routed to the per-session `turns.log` (`-t` output unchanged) — **Q1 → A**.
- **FR-007**: The diagnostic MUST be **informational** — the turn proceeds and the emitted request body is unchanged; it MUST NOT abort the turn — **Q2 → A**.

**Non-Functional Requirements**:

- **NFR-001**: The diagnostic MUST respect the existing chrome gate (terminal-gated colour; never colour into a non-terminal or into `turns.log` — ADR 0023).
- **NFR-002**: Hermetic, POSIX-only, stdlib-only; no new dependency, no new config key (I-6).

---

## Edge Cases

- **`M == N`** (shipped happy path) — no diagnostic (I-7).
- **`M = 0`** (no results for a round) — all N call ids unpaired → the diagnostic names them (call order).
- **Multiple rounds with unpaired calls** — the diagnostic reports in call order, per round / aggregated (finalise at research).
- **Media-bearing round** — the diagnostic does not alter the batched/media wire (I-2/I-3).
- **Non-terminal `stderr` / `-r`** — the diagnostic is **plain** (no colour; terminal-gated, ADR 0023) but **still emitted** (a diagnostic is not decoration). It is **never** written to `turns.log` (Q1 → A).
- **OpenAI-compatible provider** — unchanged, no diagnostic (I-1/S-5).

## Key Entities

- **An unpaired call** — a call of a round that received no result by the round boundary; its id is what the diagnostic names.
- **The diagnostic** — the new user-visible `stderr` line (and, per Q1, possibly a `turns.log` line).
- **The round** — a model turn's N calls + their M results; the unit whose boundary drop the diagnostic reports.
- **The Gemini request `contents`** — unchanged by this round (the diagnostic is out-of-band).

## Success Criteria

- **SC-001**: With `M < N`, the unpaired call ids are surfaced on `stderr` (per Q1) — a red-capable carrier (drive the built binary with a fixture and assert the line). *(E2E or CLI-tier pin.)*
- **SC-002**: With `M == N`, **no** new diagnostic is emitted (I-7). *(Negative pin.)*
- **SC-003**: `stdout` is byte-exact and the Gemini/OpenAI wires are byte/shape-preserved (I-1/I-2/I-3/I-4). *(Regression pins.)*
- **SC-004**: The falsifiability witness: removing the diagnostic reds SC-001.
- **SC-005**: `make verify` + `go test -count=1 ./...` green; topology audit adds no new findings; `go.mod`/`go.sum` unchanged.

## Assumptions

- **A1**: The round is **family-local** (Gemini/Vertex) — the concept does not exist for the OpenAI-compatible family (S-5/I-1).
- **A2**: The diagnostic is **informational** (a warning), not a terminal failure — the turn proceeds (S-2) *(pending Q2)*.
- **A3**: The detection seam + the exact rendering are **technical** decisions for `/axb-technical-research` (S-6), bounded by I-1…I-7 — the adapter owns no logging seam today, so the seam is likely an injected sink (the `tools.OutputSink`/ADR-0021 precedent) or a loop-side observer event.
- **A4**: **Truth impact expected** — `specs/truth/techstack.md` (the *Vertex/Gemini adapter* row + the *Agent tool loop* row as needed) and a **new ADR** (or an **ADR-0037 forward-annotation**) recording the diagnostic. `/axb-api-plan` **NOOP**; `/axb-data-plan` **NOOP** (no persisted shape change).
- **A5**: **This is not NOOP for `/axb-spec-by-example` / `/axb-dsl-refine`** — unlike rounds 066/067, the diagnostic is **user-visible**, so a small acceptance Rule/Example + a CLI interface feature/DSL row are expected (the round-059/060 precedent for a stderr diagnostic carrier).
- **A6**: Hermetic, stdlib-only, POSIX-only; no new dependency/config key (I-6).

## Out of scope (recorded forward items)

- **Concurrent tool execution** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — the real producer of an out-of-order result set; not added here.
- **RF-067-3** (an E2E carrier for the wire `id`) · **RF-067-5** (the reference fallback spelling) · **RF-067-6** (duplicate-id handling) · **RF-067-7** (the `history.Step` guarding note) — other ADR-0037 forward items.
- **The `ToolSetSpec` seam** (RF-062-10 / RF-063-6) — a separate structural round (recorded).
- **A loud failure** variant (Q2 → B) — not taken; recorded as a forward item.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion).
