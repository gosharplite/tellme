# Feature Specification: Gemini/Vertex usage drops the cached-token count (round 072)

**Feature Branch**: `072-gemini-cached-token-usage`

**Created**: 2026-09-20

**Status**: Draft (specified — clarify **not escalated by default**; see §Clarify strategy)

**Anchor**: issue [#149](https://github.com/gosharplite/tellme/issues/149) — **the round's DoD is closing it.**

**Input (operator, 2026-09-20)**:

> *"I briefly switch to gemini and something wrong! The hit is always zero, this is too expensive! Please investigate why by referencing tell-me-go."* → (investigation) → *"Create a detail github bug issue on this. Open round 072, the goal is to close this new issue."*

**Behaviour intent**: **MODIFY (a provider-adapter defect fix)** — the Gemini/Vertex adapter must decode the response's `cachedContentTokenCount` (and `thoughtsTokenCount`) into `llm.Usage`, so the post-turn metrics line, the cost computation, and the persisted usage record stop billing the entire prompt at the MISS rate. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **This is a real, reachable, expensive defect** — not a defensive/unreachable case. On the Gemini/Vertex family, `H` is always `0`, so `miss = prompt − 0 = prompt` and `ComputeCost` charges the whole prompt at the `MISS` rate (config `gemini-3.8-flash`: `HIT 0.075 / MISS 0.75 / COMP 3.75` — **MISS is 10× HIT**), inflating the reported cost by up to **~10×**.
- **The reference shows the intended shape** (`tell-me-go/internal/infrastructure/llm/gemini/metrics.go:17-28`): `m.CachedTokens = um.CachedContentTokenCount` and `m.ThinkingTokens = um.ThoughtsTokenCount` (when > 0).
- **In-tellme contrast**: the **OpenAI-compatible** adapter already parses the equivalent fields (`openai/client.go:278-281,299-309`), so DeepSeek shows real hits. The two adapters disagree — this round makes the Gemini one consistent.
- **The OpenAI-compatible wire is frozen** — the change is confined to `internal/infrastructure/llm/gemini/**` (+ tests + the truth row + an ADR); the emitted turn is byte-identical.

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `5b1bfa1`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/llm/gemini/client.go:532-535` | the `usageMetadata` decode struct — **three** fields (`promptTokenCount`, `candidatesTokenCount`, `totalTokenCount`); **no** `cachedContentTokenCount`, **no** `thoughtsTokenCount`. |
| `internal/infrastructure/llm/gemini/client.go:574-580` | the `llm.Usage` construction — sets `PromptTokens`/`CompletionTokens`/`TotalTokens`; **leaves `CachedTokens` and `ThinkingTokens` zero**. |
| `internal/domain/llm/gateway.go:77-84` | `llm.Usage{Reported, PromptTokens, CachedTokens, CompletionTokens, ThinkingTokens, TotalTokens}` — the type **already has** the two fields. |
| `internal/cli/call_renderer.go:205-206` | `miss := c.PromptTokens - c.CachedTokens`; `cost := llm.ComputeCost(pricing, miss, c.CachedTokens, c.CompletionTokens, c.ThinkingTokens)`. |
| `internal/domain/llm/pricing.go:18` | `(miss·Miss + hit·Hit + (completion+thinking)·Comp) / 1e6`. |
| `internal/infrastructure/llm/openai/client.go:278-281,299-309` | the **working** sibling: reads `prompt_tokens_details.cached_tokens` and `completion_tokens_details.reasoning_tokens`; stores the **exclusive** completion (`completion − reasoning`). |
| Reference `tell-me-go/internal/infrastructure/llm/gemini/metrics.go:17-28` | `CachedTokens = um.CachedContentTokenCount`; `ThinkingTokens = um.ThoughtsTokenCount` when `> 0` (treated **disjoint** from candidates — no subtraction). |

**Reference `usageMetadata` field names** (proto-JSON / SDK): `promptTokenCount`, `candidatesTokenCount`, `totalTokenCount`, **`cachedContentTokenCount`**, **`thoughtsTokenCount`** (also `trafficType`, `toolUsePromptTokenCount` — not needed here).

---

## Design (the shape is a `/axb-technical-research` decision; residual choices marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | The Gemini adapter decodes `cachedContentTokenCount` into `llm.Usage.CachedTokens`. | locked (round goal) |
| **S-2** | The Gemini adapter decodes `thoughtsTokenCount` into `llm.Usage.ThinkingTokens`. | locked (round goal) |
| **S-3** | The OpenAI-compatible wire and the emitted turn are **unchanged** (byte-identity); the change is family-local. | locked (I-1) |
| **S-4** | **Disjointness rule**: is `candidatesTokenCount` **inclusive** of `thoughtsTokenCount` (subtract, like the OpenAI adapter) or **disjoint** (no subtraction, like the reference)? | **research decision** (D-x) |
| **S-5** | A response with **no** `usageMetadata` (or a zero field) leaves the counts at 0 — no crash, `Reported` stays the gate. | locked (I-3) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Byte-identity** — the request body, the emitted turn, and the OpenAI-compatible wire are unchanged; only the Gemini **usage decode** changes.
- **I-2 — The cost formula is untouched** — `ComputeCost`/`miss = prompt − cached` are reused as-is; correctness comes from populating `CachedTokens`, not from a new formula.
- **I-3 — Degenerate responses are safe** — a missing/partial `usageMetadata` yields zeros (the existing `Reported` gate), never a panic or a negative miss.
- **I-4 — No new dependency; stdlib-only; POSIX-only; hermetic.**
- **I-5 — The four surfaces agree** — the metrics line, the `Ready` summary totals, the persisted `tokens.log` (`UsageRecord`), and the session accumulator all derive from the one `llm.Usage` value.

---

## Clarify strategy

**Not escalated by default (0 questions).** The defect and the fix are unambiguous and grounded in the issue + the reference; the one residual choice (**S-4**, the `thoughtsTokenCount` disjointness rule) is **technical** and defers to `/axb-technical-research` (it must decide + pin it). *If research finds the rule materially changes the reported figures the operator relies on (e.g. `Th` billed double), it MAY escalate ONE question rather than assume.* **No `NEEDS CLARIFICATION` remains.**

---

## User Stories (proposed)

### US1 — A Gemini turn reports its real cache hits (Priority: P1)

A Gemini/Vertex response that carries `cachedContentTokenCount` surfaces it as the turn's `H`; the cached portion is billed at the HIT rate and the miss is `prompt − cached`.

**Why P1**: it is the round's whole goal — the ~10× over-charge.

**Acceptance (proposed)**:

1. **Given** a Vertex response whose `usageMetadata` is `{promptTokenCount: 390564, cachedContentTokenCount: 389538, candidatesTokenCount: 100, totalTokenCount: 390664}`, **When** the turn completes, **Then** the metrics line reads `H: 389538` and the computed cost uses `miss = 1026`.
2. **Given** `promptTokenCount == cachedContentTokenCount`, **When** the turn completes, **Then** the metrics line reads `M: 0`.

### US2 — Thinking tokens are surfaced (Priority: P1)

A Gemini response carrying `thoughtsTokenCount` surfaces it as `Th`, billed consistently with the chosen disjointness rule (S-4).

**Why P1**: `dev` declares `THINKING_LEVEL: HIGH`; `Th: 0` on every turn is wrong and mis-bills.

**Acceptance (proposed)**:

1. **Given** a Vertex response with `thoughtsTokenCount: 4096`, **When** the turn completes, **Then** the metrics line reads `Th: 4096` and `C`/`Th` follow the pinned rule.

### US3 — The accounting surfaces stay consistent (Priority: P2)

The metrics line, the `Ready` summary totals, and the persisted `tokens.log` record agree.

**Why P2**: the defect was silent; the fix must be visible everywhere the operator reads it.

**Acceptance (proposed)**:

1. **Given** two Gemini turns with known `H`, **When** the session is read (`-l`/`--tool-usage`-analogue / `tokens.log`), **Then** the session `H`/`M` totals equal the sum of the per-turn counts.

---

## Functional Requirements (proposed)

- **FR-001** — The Gemini adapter MUST decode `cachedContentTokenCount` into `llm.Usage.CachedTokens`.
- **FR-002** — The Gemini adapter MUST decode `thoughtsTokenCount` into `llm.Usage.ThinkingTokens`, per the pinned disjointness rule (S-4).
- **FR-003** — A response with no `usageMetadata`, or with any field absent/zero, MUST leave the corresponding counts at 0 and MUST NOT produce a negative miss (I-3).
- **FR-004** — The cost formula and the `miss = prompt − cached` derivation MUST be reused unchanged (I-2).
- **FR-005** — The request body, the emitted turn, and the OpenAI-compatible wire MUST be unchanged (I-1).
- **FR-006** — The change MUST be confined to the Gemini family (+ tests, the truth row, and a new ADR).
- **FR-007** — The metrics line, the `Ready` summary, and the persisted `UsageRecord` MUST all reflect the corrected counts from the single `llm.Usage` value (I-5).

## Success Criteria (proposed)

- **SC-001** — The `H: 0`-always symptom is gone: the Gemini usage mapping is **unit-pinned** (the adapter decodes `cachedContentTokenCount`/`thoughtsTokenCount`) and witnessed **E2E** (the fake Vertex provider scripts `usageMetadata`, and the run's metrics line shows the cached count).
- **SC-002** — The reference parity holds: the new mapping matches `tell-me-go`'s `metrics.go` (modulo the S-4 rule the round pins and records).
- **SC-003** — `make verify` green; `go.mod`/`go.sum` unchanged; the OpenAI-compatible wire byte-identical.
- **SC-004** — The over-charge is demonstrably closed: a documented witness shows the corrected cost is the HIT-weighted figure (not the MISS-rate figure).

## Edge cases (proposed)

- `usageMetadata` **absent** → all counts 0 (`Reported` false) — unchanged behaviour.
- `cachedContentTokenCount` **absent/0** → `H: 0` legitimately (a genuine cold cache).
- `cachedContentTokenCount == promptTokenCount` → `miss = 0`.
- `thoughtsTokenCount` **absent/0** → `Th` omitted (the existing zero-suppression).
- A **malformed/negative** count → floored at 0 (mirror the OpenAI adapter's `max(0, …)` precedent).
- A **tool-call round** (no text answer) still reports usage.

## Key entities

`llm.Usage` (the target) · the Gemini adapter's `usageMetadata` decode struct · `ComputeCost`/`Pricing` · `UsageRecord` (`tokens.log`) · the metrics-line/`Ready` renderers · the fake Vertex provider (E2E fixture).

## Assumptions

- **A1** — The fix is adapter-local (`internal/infrastructure/llm/gemini/**` + tests + the truth row + the ADR); the loop, the CLI renderers, the pricing formula, and the config are untouched.
- **A2** — Vertex/Gemini reports the cached count on the `generateContent` response's `usageMetadata` under `cachedContentTokenCount` (confirmed by the reference SDK usage).
- **A3** — The metrics line, the `Ready` summary, and the persisted record already derive from `llm.Usage` (they do — `call_renderer.go`/`usage.go`), so a correct `Usage` fixes all four surfaces at once.

## Out of scope

- Any **new** pricing/accounting feature (a cache-hit-ratio display, a cost breakdown table) — only the mis-decode is fixed.
- The **OpenAI-compatible** adapter (already correct).
- Provider-side cache **management** (creating/deleting cached contents) — tellme relies on implicit caching.
