# Feature Specification: Gemini/Vertex tool-call id follow-ups — surface unpaired calls + provider-issued `functionCall.id` preference (round 067)

**Feature Branch**: `067-toolcall-id-followups`

**Created**: 2026-09-20

**Status**: Draft (specified)

**Input (operator, 2026-09-20, this session)**:

> *"Open round 067, the goal is to close https://github.com/gosharplite/tellme/issues/136"*

**Anchor issue**: [#136](https://github.com/gosharplite/tellme/issues/136) — *"Round 067 anchor — Gemini/Vertex tool-call id follow-ups: surface unpaired calls (RF-066-7) + provider-issued `functionCall.id` preference (RF-066-2)"*. **The round's definition of done is closing #136.**

**Behaviour intent**: **MODIFY (the Gemini/Vertex adapter's tool-call id handling) — a *hardening / parity* round, NOT a user-visible defect fix.** It is the small, self-contained continuation of round 066 (**ADR 0036**), covering its two deferred Gemini/Vertex id forward items — both live in **`internal/infrastructure/llm/gemini`**:

- **RF-066-7** — **surface the unpaired calls** of an `M < N` round (an id accessor) instead of the current silent drop, so the still-open `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual can be replaced by a mutant-killing pin (**RF-066-8** retired).
- **RF-066-2** — **provider-issued `functionCall.id` preference**: read the Vertex response's own `candidates[].content.parts[].functionCall.id` and **prefer** it over the synthetic `call_<n>`, completing the reference-parity axis round 066 deliberately deferred.

**No `specs/truth/**` file is written by this skill.** Clarify: **not escalated** (see *Clarify strategy* — the goal and the two changes are fully grounded in #136; the surviving choices are **technical**, deferred to `/axb-technical-research`).

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `30f54a1`; #136 states the same at `3637ec2`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/llm/gemini/client.go` `parseResponse` | Builds `llm.ToolCall{ID: fmt.Sprintf("call_%d", callIdx), …}` — a **synthetic, positional** id. The Vertex response's own `functionCall.id`, if any, is **unread** (the decode struct's `FunctionCall` has only `Name` + `Args`). |
| `internal/infrastructure/llm/gemini/client.go` `roundBuilder.flush` | Emits the buffered round's **`M`** `functionResponse` parts (call order) and clears `pending` — a call left **unpaired** at a round boundary (a round that yielded `M < N` results) contributes no part and is **silently dropped** (no error, log, or accessor) — behaviourally identical to the round-065 `pending[:0]`. |
| `internal/infrastructure/llm/gemini/client.go` `roundBuilder.bind` | Already pairs each result to its call **by `ToolCallID`** (identity, order-independent), with the FIFO name match as the fallback — the round-066 core. `callEntry` already carries `id` + `name`. |
| `internal/agent/agentloop.go` | Live path: `llm.Message{Role:"tool", Content:result, ToolCallID:tc.ID}` per call (`~:145/171`). **Replay** path: a deterministic `call_step_<n>` id per replayed step (`~:413-426`). Every producer sets a `ToolCallID`. |
| `internal/infrastructure/llm/openai/client.go` | The OpenAI-compatible wire **already** emits `"id": tc.ID` on `tool_calls` (`:126`) and `tool_call_id: m.ToolCallID` on `role:"tool"` (`:136-137`). The id **value** therefore flows to **both** families. |
| `internal/infrastructure/llm/gemini/client_ids_test.go` + the round-065 pins | `TestRequestBody_ToolPartsCarryIDs`, `…_OutOfOrderResultsPairByIdentity`, `…_EmptyToolCallIDOmitsID`, `…_UnmatchedToolCallIDFallsBackToFIFO`, `…_ReplayedStepIDsPairByIdentity`, and the round-065 `TestRequestBody_ShortRound_DropsUnpairedNames` (`N=2 M=1`) pin the round-066 id axis + the boundary drop. |
| `specs/truth/techstack.md` | the *Vertex/Gemini adapter* row records the round-066 id-link + id-keyed pairing (FIFO fallback). The id **provenance** clause is the round-067 subject. |
| `tell-me-go` (reference) | `fromSDKFunctionCall` reads the provider `f.ID` **first**, else a deterministic `gemini-call-<index>-<name>`; **empty id = invalid** (`isInvalidToolPart` strips such parts). Its executor coincidentally copies `call.ID` into every response. |

**Why this is hardening, not a defect:** tellme's loop is **sequential** and appends results in call order, so today the synthetic `call_<n>` id and the positional pairing are always correct; the id **value** is not observable by the user through any shipped surface. RF-066-2 becomes load-bearing only when the id axis is relied on beyond tellme's own bookkeeping (a provider that keys on its own id; a future concurrent dispatch — [#36](https://github.com/gosharplite/tellme/issues/36) item 3); RF-066-7 is the round-066 accounting gap ADR 0036 §Forward recorded.

---

## Design (the two goals are the round's settled scope; the residual choices are proposed — pending `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **When a Gemini round yields `M < N` results, the adapter makes the unpaired calls' ids observable** (never a silent drop). This is the change that lands #136 item A (**RF-066-7**). | **locked (round goal / #136)** |
| **S-2** | **The provider's own `functionCall.id` is preferred** when the Vertex response carries a non-empty one; otherwise the existing deterministic id is kept. This is #136 item B (**RF-066-2**). | **locked (round goal / #136)** |
| **S-3** | **The observability form of S-1 is a technical choice** — a single-owned **accessor** returning the unpaired call ids (e.g. `roundBuilder.unpaired() []string`) exposed as a **returned value**, a **diagnostic** (a `[Tool …]`-class `stderr` note / an `slog`-style warning), or **both** — settled by `/axb-technical-research`. The **loop/ports MUST stay untouched** unless research shows an accessor cannot be owned inside the adapter. | proposed (research) |
| **S-4** | **The fallback id spelling is a technical choice** — keep tellme's `call_<n>` or adopt the reference's `gemini-call-<index>-<name>` — but **whichever is chosen, the id MUST be deterministic for a given turn** (replay stability, ADR 0036 D3 / round-014 fidelity). | proposed (research) |
| **S-5** | **The cross-family id question is an explicit, recorded decision** — the id **value** flows to the OpenAI-compatible `tool_call_id` too. Either **(i) the provider id is Gemini-local** (never leaks to the OpenAI wire — the loop keeps its own ids for that family; the **conservative default**) or **(ii) the id is unified** and the OpenAI `tool_call_id` value changes (a deliberate, recorded divergence from the round-066 byte-freeze **I-1**). Settled by `/axb-technical-research` + recorded in the round's ADR / ADR-0036 D-note. | proposed (research; default **(i)**) |
| **S-6** | **A provider-id-absent response keeps today's behaviour** (the deterministic id) — **no new failure mode**; an **empty** provider id is treated as absent (never an empty id on the wire — ADR 0036 D2/D4). | proposed |
| **S-7** | **No new dependency / no new tool / no new config key / no port change** — stdlib-only (`encoding/json`); confined to **`internal/infrastructure/llm/gemini`** (plus pins, plus the round-067 truth row). | proposed |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The OpenAI-compatible wire is byte-preserved** *unless* S-5 is explicitly taken as **(ii)** and recorded; under the default (i) its request serialization is **byte-identical** and its existing tests stay green and unchanged.
- **I-2 — The round-066 batched shape is preserved** — a round of **N** calls still serializes to **one** `user` turn carrying the produced `functionResponse` parts (call order), media turns after it (ADR 0035 D1/D3); round 067 does **not** change the turn/part structure.
- **I-3 — The media-free Gemini text path is byte-preserved** (no tool parts ⇒ no id change).
- **I-4 — No silent media loss** (ADR 0032 D8 / ADR 0033): every image a round attached is carried on the wire or the turn fails loudly — a change MUST NOT satisfy a shape/accounting check by dropping media.
- **I-5 — Round-014 replay fidelity is preserved:** a resumed turn's persisted `steps` replay with the loop's deterministic `call_step_<n>` on **both** the call and the result, so the round pairs **id-primary**; the FIFO fallback stays the defensive path for an id-less result. **No regression to FIFO-only.**
- **I-6 — No security layer, POSIX-only, hermetic.** No new dependency; every gate offline (ADR 0012).
- **I-7 — The universal reason gate (ADR 0025) and the round-024 tool resource contract are untouched.**
- **I-8 — The round's `id` is omitted when empty** (never `id:""`) and a **foreign** id (not equal to the bound call's id) stays omitted (ADR 0036 D2/D4).

---

## Clarify strategy

**Not escalated.** Per the clarify-escalation rule only a gap that changes the user-story split, requirement attribution, the main flow, scope, or the formal acceptance criteria warrants `/axb-clarify`. Here the goal is unambiguous (**close #136**), both changes are fully grounded (the issue carries the current behaviour, the reference comparison, and the proposed change), and the surviving choices — the observability form (S-3), the fallback spelling (S-4), and the cross-family id scope (S-5) — are **technical**, disclosed as S-3/S-4/S-5 and in *Assumptions* rather than asked. This mirrors the round-066 precedent (its S-4/S-6 were deferred to research). No `NEEDS CLARIFICATION` remains.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - a Gemini round's unpaired calls are surfaced, not silently dropped (Priority: P1)

As the **operator** running `tellme` against a **Gemini/Vertex** provider, when a round's model turn requests **N** calls but the round yields **M < N** results, I want the **unpaired calls' ids to be observable** (an accessor / a diagnostic), so that the boundary drop is **accounted** and never a silent loss — retiring the round-066 `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual (RF-066-8).

**Why this priority**: it is the concrete capability #136 adds (item A) and it turns a recorded gap into an owned, testable property; it is independent of the provider-id question (US2) and can ship alone.

**Independent verification**: build the request body for a round that yields `M < N` results and assert the produced `M` parts are named correctly **and** the unpaired calls' ids are observable through the round's single-owned accessor (or the chosen diagnostic surface); a witness that suppresses the accessor reds the check.

**Acceptance Scenarios**:

1. **Given** a Gemini round with **N** calls and **M < N** results, **When** the body is built, **Then** the batched `user` turn carries the **`M`** parts the round produced (I-2) **and** the `N − M` unpaired calls' ids are **observable** (never a silent drop).
2. **Given** a round whose every call is paired (**M = N**), **When** the body is built, **Then** no unpaired call is reported (the accessor yields none) — the accounting does not invent a gap.
3. **Given** the **media-free text** round / the single-call round, **When** the body is built, **Then** the emitted body is unchanged apart from the round-066 id axis already in place (I-2/I-3).

**Functional Requirements**:

- **FR-001**: When a round yields `M < N` results, the adapter MUST make the **unpaired call ids observable** — it MUST NOT drop them silently.
- **FR-002**: The observability MUST have a **single owner** (one accessor or one diagnostic site on the round), so the accounting cannot drift between call sites.
- **FR-003**: The emitted batched turn MUST still carry the **`M` parts the round produced** (I-2); the round boundary still prevents a later round's part from mispairing.
- **FR-004**: The surfaced unpaired ids MUST be **deterministic** (a stable order — call order) so a witness can assert them.
- **FR-005**: The **replay** path MUST keep its id-primary pairing (I-5); surfacing unpaired calls MUST NOT change the pair-by-id behaviour.

**Non-Functional Requirements**:

- **NFR-001**: The observability MUST be **offline/hermetic** in tests (no network, no credential) and MUST NOT write to `stdout` (a diagnostic goes to `stderr` only, if chosen).

### User Story 2 - the adapter prefers the provider's own `functionCall.id` (Priority: P2)

As the **operator**, I want the Gemini/Vertex adapter to use the **provider-issued** `functionCall.id` when the Vertex response carries one (falling back to the deterministic id otherwise), so that tellme's round is bound by the **same identity the provider uses** — the reference's `f.ID`-first provenance — instead of always synthesising a positional `call_<n>`.

**Why this priority**: it is the reference-parity axis (#136 item B, RF-066-2) and it is only observable through the request body (no user-visible surface today), so it ranks below the accounting fix that makes the id axis trustworthy.

**Independent verification**: drive a Vertex response carrying `functionCall.id` and assert the emitted `functionCall`/`functionResponse` parts carry **that** id; drive a response **without** one and assert the deterministic fallback (deterministic across builds); a witness that removes the preference reds the first leg.

**Acceptance Scenarios**:

1. **Given** a Vertex response with a **non-empty** `candidates[0].content.parts[].functionCall.id`, **When** the adapter builds the next request body, **Then** the emitted call/result parts carry **that** id (provider preference).
2. **Given** a Vertex response with **no** provider id (or an **empty** one), **When** the body is built, **Then** the deterministic fallback id is used (S-4), **identical across builds** for the same turn, and **no** empty `id` key is emitted (I-8).
3. **Given** the chosen cross-family decision (S-5), **When** the id is produced, **Then** the OpenAI-compatible `tool_call_id` is **byte-frozen** under the default **(i)** (Gemini-local) — or the divergence is explicitly recorded under **(ii)** (I-1).

**Functional Requirements**:

- **FR-006**: `parseResponse` MUST read `candidates[0].content.parts[].functionCall.id` and **prefer** it when present and non-empty.
- **FR-007**: When the provider id is **absent or empty**, the adapter MUST use the **deterministic fallback** (S-4), **stable for a given turn** (no per-build randomness).
- **FR-008**: The adapter MUST **never** emit an empty `id` on the wire (I-8); an empty provider id is treated as absent.
- **FR-009**: The **cross-family id scope** (S-5) MUST be decided explicitly and recorded; under the default **(i)** the OpenAI-compatible wire stays **byte-identical** (I-1).
- **FR-010**: The provider-id preference MUST hold on **both** the live path and the **replay** path, without breaking I-5 (a replayed round still pairs id-primary and produces a stable body).

**Non-Functional Requirements**:

- **NFR-002**: The id value flow MUST remain **stdlib-only** (`encoding/json`) with no new dependency (I-6).

---

## Edge Cases

- **`M = N`** (every call paired) — no unpaired call reported (US1 Scenario 2).
- **`M < N`** — the produced `M` parts emitted; the `N − M` unpaired ids observable (US1 Scenario 1).
- **`M = 0`** (no results for the round) — the batched turn is omitted (no parts), and **all N** call ids are unpaired/observable.
- **Provider id absent** — deterministic fallback (S-4).
- **Provider id empty string** — treated as absent; no empty `id` key (FR-008).
- **Two calls with the same provider id** (a degenerate provider response) — the pairing is order-independent by identity; the behaviour MUST be deterministic and **not** a silent mispair (recorded if it needs an explicit rule — S-4/S-5).
- **Replay path** — `call_step_<n>` on both sides; id-primary pairing (I-5), unchanged.
- **A result with no `ToolCallID`** — FIFO fallback (round 066), unchanged; no `id` key emitted.
- **Media on some/all calls** — no media loss (I-4); the batched turn carries all produced results and the media turns follow.
- **The OpenAI-compatible family** — byte-frozen under S-5 (i) (I-1).
- **A round at a boundary (a new model turn / a text turn / the prompt)** — the previous round is flushed; its unpaired calls are accounted at the flush.

## Key Entities

- **A model round** — one assistant turn carrying N `functionCall` parts plus the results/media it produces; the unit whose pairing and accounting this round refines.
- **A call id** — the identity linking a `functionCall` part to its `functionResponse` part; its **provenance** is this round's US2 subject.
- **An unpaired call** — a call of a round that received no result by the round boundary; its id is the observable this round adds (US1).
- **The Gemini request `contents`** — the serialized wire body; its construction is the round's subject (an unchanged shape; an id whose value may now come from the provider).
- **A `functionResponse` part** — one tool result, bound to its call by id (fallback: FIFO name) — unchanged.

## Success Criteria

- **SC-001**: For a round that yields `M < N` results, the produced `M` parts are emitted and the `N − M` unpaired call ids are **observable** through the round's single owner (US1). *(Unit pin over the built request body + the accessor/diagnostic.)*
- **SC-002**: A provider response carrying `functionCall.id` produces parts carrying **that** id; a response without one produces the **deterministic fallback** (US2). *(Unit pins.)*
- **SC-003**: The **media-free text** path and the **OpenAI-compatible** wire stay **byte-identical** under the default S-5 (i); the **tool-bearing Gemini** shape is unchanged apart from the id values (I-1/I-2/I-3). *(Regression pins.)*
- **SC-004**: The round-066 pins (`TestRequestBody_ToolPartsCarryIDs`, `…_OutOfOrderResultsPairByIdentity`, `…_EmptyToolCallIDOmitsID`, `…_UnmatchedToolCallIDFallsBackToFIFO`, `…_ReplayedStepIDsPairByIdentity`) and the round-065 batch pins stay green (I-2/I-5/I-8). *(Regression pins.)*
- **SC-005**: A red-capable carrier proves the change: suppressing the unpaired-id accessor reds SC-001; removing the provider-id preference reds the preference leg of SC-002; a non-deterministic fallback reds the stability leg of SC-002.
- **SC-006**: `make verify` + `go test -count=1 ./...` green (incl. the E2E contract); the topology/DSL audit adds **no** new findings; no new dependency; `go.mod`/`go.sum` unchanged.
- **SC-007**: The `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual is **retired** — replaced by (or explicitly closed against) a pin that **kills** the partial-drop mutant (RF-066-8).

## Assumptions

- **A1**: The change is **family-local** — `internal/infrastructure/llm/gemini` (+ pins, + the round-067 truth row). The OpenAI-compatible wire is **byte-frozen** under the S-5 default (i).
- **A2**: **This is a hardening/parity round** — no user-visible behaviour changes through a shipped surface today, *unless* the unpaired-call observability is deliberately surfaced as a **user-visible diagnostic** (S-3); if so, `/axb-spec-by-example` + `/axb-dsl-refine` carry a **small interface Rule/Example** instead of the expected **NOOP** (the research decides).
- **A3**: The exact **observability form** (S-3), **fallback spelling** (S-4), and **cross-family id scope** (S-5) are **technical** decisions for `/axb-technical-research`, bounded by I-1 (OpenAI wire frozen unless (ii) is explicitly taken and recorded) and I-5 (replay stable).
- **A4**: **Truth impact expected** — `specs/truth/techstack.md` (the *Vertex/Gemini adapter* row: the id-provenance clause + the unpaired-call accounting; the *Image content on the provider wire (Gemini/Vertex)* row if the id axis touches the batched-shape wording), and a **new ADR** (or an **ADR-0036 forward-annotation + D-note**) recording the id-provenance + cross-family decision. `/axb-api-plan` **NOOP**; `/axb-data-plan` **NOOP**.
- **A5**: **Testable hermetically.** The primary carrier is a **unit pin over the built request body** (the round-065 D5 / round-066 precedent) plus the accessor/diagnostic pin; the round-065/066 E2E journeys drive multi-call rounds end to end unchanged. `/axb-spec-by-example` is **expected NOOP** (no user-visible change — the 042/043/045–052 + 066 structure-round precedent); `/axb-dsl-refine` is **expected NOOP or a small MODIFY** (only if the observability is user-visible — S-3). **No** real network/credential.
- **A6**: No new tool, no config key, no port change, no new dependency (`encoding/json` only); POSIX-only; hermetic (ADR 0012).
- **A7**: Byte-identity is claimed **only** for the media-free text path and the OpenAI-compatible wire; the tool-bearing Gemini body stays **shape-identical** (the id **values** are the sole possible difference), continuing the round-066 narrowing.

## Out of scope (recorded forward items)

- **A fake-side contract check** (the E2E fake rejecting a `functionResponse`/`functionCall` count mismatch) — **RF-066-3** / RF-065-3.
- **Concurrent tool execution** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — **RF-066-4**; round 067 makes the id axis **more** ready, it does **not** add concurrency (sequential tools remain a settled exclusion).
- **An E2E carrier for the wire `id`** — **RF-066-9** (no E2E asserts the wire `id`; the pins are fixture-based).
- **The branch-binary live-check convention** — **RF-066-10** (a mid-round live check MUST build the **branch** binary and record `go version -m` provenance).
- **A family-agnostic round-batching/id concept** in the domain — **RF-065-4**.
- **The `ToolSetSpec` seam** (RF-062-10 / RF-063-6) — a separate structural round.
- **RF-063-10** is **RETIRED** (2026-09-20) — **not** part of this round.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion).
