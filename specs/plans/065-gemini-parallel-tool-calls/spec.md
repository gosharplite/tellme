# Feature Specification: Gemini/Vertex parallel tool calls — a round's tool results must share one turn (round 065)

**Feature Branch**: `065-gemini-parallel-tool-calls`

**Created**: 2026-09-20

**Status**: Draft — produced by `/axb-specify` from **anchor issue [#132](https://github.com/gosharplite/tellme/issues/132)** (a defect found by the round-063 closeout **live check**, 2026-09-20). No `specs/truth/**` file is written by this skill. Clarify: **not escalated** (see *Clarify strategy* — the goal and the fix are unambiguous; the remaining choices are technical, deferred to `/axb-technical-research` per the round-063 placement precedent).

**Input (operator, 2026-09-20, this session)**:

> *"Open round 065-*. **The goal is to close issue 132**."*

**Anchor issue**: [#132](https://github.com/gosharplite/tellme/issues/132) — *"Gemini/Vertex: a round with ≥2 parallel tool calls fails with HTTP 400 (function-response parts must share the function-call turn)"*. **The round's definition of done is closing #132.**

**Behaviour intent**: **MODIFY (the Gemini/Vertex request serialization when a model round carries ≥2 tool calls).** Today, a round whose model turn contains two or more `functionCall` parts produces two separate `user` turns each carrying one `functionResponse` part, and Vertex rejects the follow-up request. The round makes a Gemini round's tool results serialize as the wire requires — **all of the round's `functionResponse` parts in one `user` turn** (the media turns follow). Nothing else changes: the OpenAI-compatible wire, the single-call path (with or without media), and the media-free text path are **byte-preserved**.

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `fc44aa8`)*

| Site | Current shape |
| --- | --- |
| `internal/agent/agentloop.go` | Per tool call, the loop appends **one** `llm.Message{Role:"tool", Content:result, ToolCallID:tc.ID}` (lines ~145/171); a call that attached media appends **one extra** `llm.Message{Role:"user", Media:media}` **immediately after its own tool result** (line ~176). The **replay** path appends one `tool` message **per replayed step** (line ~426). |
| `internal/infrastructure/llm/gemini/client.go` `buildContents` | Maps **each** `Role:"tool"` message to its **own** `user` turn carrying a **single** `functionResponse` part; a media message maps to its own `user` turn (`inlineData` parts). So a two-call round becomes `[model: fcA, fcB] [user: frA] [user: inlineData(A)] [user: frB] [user: inlineData(B)]`. |
| `internal/infrastructure/llm/gemini/client_image_test.go` | `TestRequestBody_MultiCallRound_MediaTurnsInterleave` **pins the interleaved shape above** (the round-063 TD-063-1 pin) — it must be **updated** by this round. |
| `specs/truth/techstack.md` | the **Gemini/Vertex provider transport** row describes the current (community) mapping; it must state the batched-function-response rule. |
| `docs/decisions/0033-gemini-image-vision.md` (D2 + §Forward RF-063-7) | D2's per-message mapping + its "cannot arise by construction" scope note; **RF-063-7** names "round-scoped placement" (all of a round's `functionResponse`s contiguous, media turns after) as the scalable shape. This round lands it. |
| `tell-me-go` (reference) | its Gemini adapter **normalizes** the user turn (`normalizeUserTurnParts`) so a function-call turn's responses share one turn with media reordered — the parity precedent. |

**Reproduction (live, 2026-09-20 — the defect this round closes):** on the `dev` provider (`gemini-3.8-flash`, Vertex), a prompt asking for **2 × `read_files`** (no media) ran round 1 and then the **next** Vertex request failed:

```text
tellme: the provider request failed: provider dev: provider returned status 400:
Please ensure that the number of function response parts is equal to the
number of function call parts of the function call turn.
```

Exit **6**. The **identical** 400 reproduces with **2 × `read_image`**; a **single** `read_image` succeeds (`exit 0`, the model described the image). The defect is therefore **media-agnostic** and **pre-existing** (per-call tool message = round 008; per-message `functionResponse` user turn = round 013; round 063 added only the `inlineData` branch). The OpenAI-compatible family is unaffected (it uses separate `role:"tool"` messages, which is legal there).

---

## Design (S-1 is the round's settled goal; the rest is proposed — technical choices pending `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **A Gemini/Vertex round's tool results serialize as ONE `user` turn carrying all `functionResponse` parts** (in call order), so the follow-up request is accepted. This is the fix that closes [#132](https://github.com/gosharplite/tellme/issues/132). | **locked (round goal / #132)** |
| **S-2** | **Media (`inlineData`) turns follow the batched function-response turn** — round-scoped placement (RF-063-7's scalable shape): all of a round's `functionResponse`s contiguous, then the media. | proposed |
| **S-3** | **The fix is family-local** — the Gemini/Vertex adapter only. The OpenAI-compatible wire, the loop's message model, the ports, the tools, and the config are **unchanged** unless `/axb-technical-research` shows a loop-side ordering change is necessary (then the OpenAI wire must stay byte-preserved — I-1). | proposed |
| **S-4** | **Placement details are a technical decision** — whether the round's media rides **one** merged `user` turn or **one turn per call**, and whether the loop's append order changes, is settled by `/axb-technical-research` (the round-063 placement precedent). Acceptance is **behavioural**. | proposed (research) |
| **S-5** | **The round-063 pin is superseded** — `TestRequestBody_MultiCallRound_MediaTurnsInterleave` is updated to the new shape (a shape pin, not a behaviour claim). | proposed |
| **S-6** | **No new dependency / no new tool / no new config key / no port change** — stdlib-only; the change is confined to the Gemini adapter (+ a pin update, + possibly a loop ordering tweak). | proposed |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The OpenAI-compatible wire is byte-preserved.** Its request serialization is correct today (one `role:"tool"` message per `tool_call_id`); the fix MUST NOT change it. Its existing tests stay green and unchanged.
- **I-2 — The single-call path is byte-preserved on Gemini.** A round with exactly **one** tool call (with or without media) serializes exactly as today (the round-063 D2 shape stands).
- **I-3 — The media-free text path is byte-preserved** on both families.
- **I-4 — No silent media loss** (ADR 0032 D8 / ADR 0033): every image the model attached is carried on the wire or the turn fails loudly; never dropped.
- **I-5 — No security layer, POSIX-only, hermetic.** No new dependency; every gate offline (ADR 0012).
- **I-6 — The universal reason gate (ADR 0025) and the round-024 tool resource contract are untouched.**

---

## Clarify strategy

**Not escalated.** Per the clarify-escalation rule, only a gap that changes the user-story split, requirement attribution, the main flow, scope, or the formal acceptance criteria warrants `/axb-clarify`. Here the goal is unambiguous (**close #132**), the defect and its fix are fully grounded and reproduced, and the surviving choices are **technical** — the exact placement shape, and where the batching lives — which the repo's process assigns to `/axb-technical-research` (the round-063 precedent: *"the `inlineData` placement was left to this phase"*). They are disclosed as **S-2/S-3/S-4** and in *Assumptions* rather than asked. No `NEEDS CLARIFICATION` remains.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - a Gemini turn that issues several tool calls completes (Priority: P1)

As the **operator** running `tellme` against a **Gemini/Vertex** provider, I want a turn in which the model requests **more than one** tool call in the same round to complete normally, so that a multi-call round (e.g. `read_files` on two paths, or two `read_image` calls) is usable instead of failing with an HTTP 400.

**Why this priority**: it is the round's entire point (closing [#132](https://github.com/gosharplite/tellme/issues/132)); the defect makes any parallel-tool round on Gemini unusable.

**Independent verification**: drive a Gemini-shaped round with **two** tool calls through the fake provider (hermetic **and** live); before the round the follow-up request is rejected (400), after the round the turn completes and the model's answer reflects **both** tool results. A falsifiability witness restores the per-message serialization and the check goes red.

**Acceptance Scenarios**:

1. **Given** a Gemini/Vertex provider and a turn in which the model requests **two `read_files`** calls in one round, **When** the round's results are sent back, **Then** the request is accepted (no 400), the turn completes, and the model's answer reflects **both** results.
2. **Given** the same provider and a turn in which the model requests **two `read_image`** calls in one round, **When** the round completes, **Then** the request is accepted and the model's answer reflects **both** images.
3. **Given** a Gemini turn whose model round requests **three** tool calls, **When** the results are sent back, **Then** the request is accepted and the round completes.
4. **Given** a Gemini turn that resumes a **stored** multi-call round (replayed from history), **When** the next request is built, **Then** it is accepted (the replay path serializes the round's results the same way).

**Functional Requirements**:

- **FR-001**: When a Gemini/Vertex model round carries **N ≥ 2** tool calls, the adapter MUST serialize the round's **N** tool results as **one** `user` turn carrying **N** `functionResponse` parts (in call order) — never N separate `user` turns.
- **FR-002**: The number of `functionResponse` parts in a function-call turn's response MUST equal the number of `functionCall` parts in that turn (the Vertex contract).
- **FR-003**: The batching MUST hold on **both** the live path and the **replay** (resumed-history) path.
- **FR-004**: A round's attached **media** MUST still be carried on the wire (I-4) — as `inlineData` part(s) in `user` turn(s) placed after the batched function-response turn (S-2).
- **FR-005**: The **single-call** round (N = 1) MUST serialize exactly as today (I-2).
- **FR-006**: The **OpenAI-compatible** family's request wire MUST be **byte-identical** to today (I-1).
- **FR-007**: A round in which some calls fail, or a call is refused for a missing reason (ADR 0025), MUST still serialize its **N** results (the error/refusal text is a `functionResponse` like any other).
- **FR-008**: A round mixing **native** and **MCP** tool calls MUST serialize all of its results together.

### User Story 2 - a multi-call Gemini round still shows the model every image (Priority: P2)

As the **operator**, when a turn reads **two images** on a Gemini provider, I want the model to actually **see both** images, so that vision works on the multi-call path and not only the single-call one.

**Why this priority**: it is the stronger, result-level form of US1's acceptance and the check the round-063 live verification asked for (RF-063-7's "two images in one turn"); it guards against a fix that makes the request *valid* by dropping media.

**Independent verification**: a Gemini round that attaches two images completes **and** the model's answer describes **both** images (greedy parity with the single-image case); a witness that drops media to satisfy the shape turns it red.

**Acceptance Scenarios**:

1. **Given** a Gemini round that reads **two** distinguishable images, **When** the turn completes, **Then** the model's answer reflects the content of **both** images (US1 Scenario 2 strengthened to the visible-content level).

**Functional Requirements**:

- **FR-009**: Every image attached by a round's tool calls MUST reach the model on the Gemini wire (I-4); the round's media MUST NOT be dropped or merged into a `functionResponse` in a way that loses it.

---

## Edge Cases

- **N = 1** (single call) — byte-preserved (I-2): the pin and the round-063 shape stand.
- **N = 0** — no tool round; the text path is byte-preserved (I-3).
- **Media on only some calls** — only those calls contribute `inlineData`; the batched function-response turn carries all N results.
- **Every call attaches media** — all images carried (US2).
- **A tool error / a reason-less refusal** — the (error/refusal) text is a `functionResponse`; N stays N (FR-007).
- **Two rounds in a turn** — each round is batched independently; the per-round boundary is respected.
- **A replayed (resumed) multi-call round** — the replay path batches too (FR-003); note media is **not** persisted (ADR 0032 D7), so a resumed round carries the tool text without the images (unchanged).
- **A call without a `ToolCallID` on the replay** — the existing FIFO name-matching is kept; the fix MUST NOT regress it.
- **An OpenAI-compatible provider** — unchanged (FR-006).

## Key Entities

- **A model round** — one assistant turn carrying N `functionCall` parts plus the N tool results (and any media) it produces; the unit that must serialize as one function-call turn + one function-response turn.
- **The Gemini request `contents`** — the serialized wire body; its construction is the round's subject.
- **A `functionResponse` part** — one tool result, keyed to the corresponding `functionCall`.
- **A media (`inlineData`) part** — an attached image (unbounded by this round; the family-aware ceiling standing).

## Success Criteria

- **SC-001**: A Gemini round with **2 × `read_files`** completes (`exit 0`) and the model's answer reflects both files — the reported 400 is gone. *(Hermetic via the fake provider; also re-verified live at the closeout.)*
- **SC-002**: A Gemini round with **2 × `read_image`** completes **and** the model describes **both** images (US2).
- **SC-003**: The **single-call** Gemini image path and the media-free text path are **byte-identical** to the pre-round serialization (a regression pin).
- **SC-004**: The **OpenAI-compatible** request wire is byte-identical (its tests unchanged) (I-1).
- **SC-005**: A red-capable carrier proves the fix: restoring the per-message serialization turns the SC-001/SC-002-class check **red**.
- **SC-006**: `make verify` + `go test -count=1 ./...` green (incl. the E2E contract); the topology/DSL audit adds **no** new findings; no new dependency.

## Assumptions

- **A1**: The fix is **family-local** (the Gemini/Vertex adapter). The OpenAI-compatible wire is correct as-is (I-1) — unless `/axb-technical-research` shows a loop-side change is cleaner, in which case I-1 still binds.
- **A2**: The exact placement (one merged media turn vs one per call; any loop ordering change) is a **technical** decision for `/axb-technical-research` (S-4); acceptance is behavioural (SC-001/SC-002).
- **A3**: Plain-CLI round — `/axb-ui-plan` skipped; `/axb-api-plan` **NOOP** (no API surface); `/axb-data-plan` **NOOP** (no persisted shape changes). `/axb-dsl-refine` is **expected to MODIFY** the CLI loop/media interface truth (a multi-call Rule/Example over the fake provider) — or NOOP if the carrier is unit-level; the research decides.
- **A4**: Truth impact expected — `specs/truth/techstack.md` (**Gemini/Vertex provider transport** row: the batched-function-response rule), and a **new ADR** (the placement decision supersedes ADR 0033 **RF-063-7** and annotates its D2 scope note; the ADR-0032/0033 forward-annotation precedent).
- **A5**: Testable **hermetically** — the E2E fake provider drives a multi-call round; `/axb-dsl-refine` adds no new dispatch; **no** real network/credential. The live re-check runs at the closeout.
- **A6**: No new tool, no config key, no port change, no new dependency.

## Out of scope (recorded forward items)

- **A family-agnostic "batch a round's tool results" concept** in the domain/ports — the fix is Gemini-local unless research shows otherwise.
- **The OpenAI-compatible family** — unchanged (it is correct).
- **The reference's exact `normalizeUserTurnParts` shape** (media reordered **before** the merged function responses in one turn) — tellme keeps standalone media turns (ADR 0033 D2).
- **RF-063-1** (the 14 MiB Gemini ceiling is derived, not measured) · **RF-063-3** (no aggregate multi-image bound) · **RF-063-4** (the Files-API upload leg) · **RF-063-5** (no image-dimension guard) · **RF-063-8** (the family-blind oversize E2E fixture) · **RF-063-9** (`ImageCeilingForFamily` fails open / stringly-typed) · **RF-063-10** (PM-owned: stop authoring comment-only meta-Rules) — all remain forward items.
- **The `ToolSetSpec` seam** (RF-062-10 / RF-063-6) — a separate structural round.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion).
