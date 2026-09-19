# Research — Gemini/Vertex parallel tool calls: a round's tool results must share one turn (round 065)

**Plan Package**: `specs/plans/065-gemini-parallel-tool-calls`
**Anchor issue**: [#132](https://github.com/gosharplite/tellme/issues/132) — *"Gemini/Vertex: a round with ≥2 parallel tool calls fails with HTTP 400 (function-response parts must share the function-call turn)"*
**Truth owner**: `/axb-technical-research` — updates `specs/truth/techstack.md`

## Problem (grounded)

A **live check** during the round-063 closeout (2026-09-20, `coder` peer, `dev` provider `gemini-3.8-flash` / Vertex) showed that a turn in which the model requests **two tools at once** fails on the **follow-up** request:

```text
tellme: the provider request failed: provider dev: provider returned status 400:
Please ensure that the number of function response parts is equal to the
number of function call parts of the function call turn.
```

Exit **6**. It reproduces with **2 × `read_files`** (no media) and **2 × `read_image`**; a **single** `read_image` succeeds (`exit 0`, the model described the image). The defect is therefore **media-agnostic**.

The cause is a **serialization split** between the loop and the Gemini adapter:

- `internal/agent/agentloop.go` appends **one** `llm.Message{Role:"tool", Content:…, ToolCallID:…}` **per tool call** (the live path, ~`:145`/`:171`; the replay path, ~`:426`), and — for a call that attached media — **one extra** `llm.Message{Role:"user", Media:…}` **immediately after its own tool result** (~`:176`).
- `internal/infrastructure/llm/gemini/client.go`'s `buildContents` maps **each** `Role:"tool"` message to its **own** `user` `contents` entry carrying a **single** `functionResponse` part; a media message maps to its own `user` entry (`inlineData`).

So a two-call round serializes to `[model: fcA, fcB] [user: frA] [user: inlineData(A)] [user: frB] [user: inlineData(B)]`. Vertex requires the `functionResponse` parts answering a function-call turn to sit in **one** turn; the request is rejected.

Provenance (both pre-existing, **not** introduced by round 063): the per-call `tool` message is round 008 (`5a37fe4`); the per-message `functionResponse` `user` turn is round 013 (`de79fc0`); round 062 added the per-call media message and round 063 (`e8e0880`) added the `inlineData` branch. The **OpenAI-compatible** family is unaffected — it uses separate `role:"tool"` messages, which its wire accepts.

ADR 0033 already named the fix: **RF-063-7** — "round-scoped placement", i.e. all of a round's `functionResponse`s contiguous (then the media turns). This round lands it.

## Decisions taken upstream

No `/axb-clarify` was needed: the goal (close #132) is unambiguous, the defect is reproduced, and the fix shape is named in the issue. The residual choices are **technical** and are settled here.

## Decisions (this phase)

### D1 — Scope: batch a round's tool results into one `user` turn; family-local

- **Decision**: the Gemini/Vertex adapter serializes a model round's **N ≥ 2** tool results as **one** `user` `contents` entry carrying **N** `functionResponse` parts (in call order). The change is **confined to `internal/infrastructure/llm/gemini`**; the loop, the message model, the ports, the tools and the config are **not** changed. The OpenAI-compatible wire is untouched.
- **Rationale**: the OpenAI-compatible wire is *correct* today (one `role:"tool"` message per `tool_call_id`), so a loop-side change would either alter a working wire (violating **I-1**) or add a family dimension to the loop it does not need. The adapter already owns the `contents` shape and already keeps a **`pending` name queue**; the batching is a local, provable transform (a pure unit-testable function).
- **Alternatives considered**:
  - **(a) Loop-side**: append one combined tool message per round (and defer media). Rejected — it changes the **OpenAI wire** (the media messages' position) and the persisted step shape, for no benefit the adapter cannot provide; larger blast radius (history/`steps`, both adapters, the replay path).
  - **(b) A family-agnostic "batch the round" domain concept** — rejected (a round is a wire concern; the OpenAI family neither needs nor wants it). Recorded as a forward item.

### D2 — The batching rule: a round = one model turn with N calls + its N results

- **Decision**: in `buildContents`, when a `model` turn carries **N** `functionCall` parts, the adapter takes the **next N** `Role:"tool"` messages as that round's results and emits them as **one** `user` entry with **N** `functionResponse` parts, in **call order** (the existing FIFO name-matching is kept: the i-th `functionCall`'s name pairs with the i-th result; `ToolCallID` is not required to be present). Media messages encountered **within** the round window are **deferred** and emitted **after** the batched function-response turn (D3).
- **Rationale**: the numbering is exactly what Vertex checks (`len(functionResponse) == len(functionCall)` per turn); grouping by the driving model turn is the natural, stateless rule, and it works identically on the **live** and **replay** paths (both produce the same message sequence).
- **Alternatives considered**: (a) match by `ToolCallID` — **not required** (the loop's native path leaves `ToolCallID` set, but the existing FIFO already pairs correctly and a `ToolCallID`-keyed map would need a fallback anyway; the FIFO stays, and this preserves today's single-call behaviour byte-for-byte); (b) group by "consecutive tool messages" — **insufficient** (a media message sits *between* frA and frB, so the round's results are not adjacent — D3 exists precisely for this).

### D3 — Media placement: standalone `user` turns, after the batched function-response turn (round-scoped)

- **Decision**: a media message stays its **own** `user` `contents` entry (round-063 D2/D3, unchanged), but the round's media entries are emitted **after** the round's single batched function-response turn. A two-call round with two images therefore serializes to `[model: fcA, fcB] [user: frA, frB] [user: inlineData(A)] [user: inlineData(B)]`.
- **Rationale**: keeps the round's `functionResponse`s **contiguous** (the Vertex contract) and keeps every media entry **standalone** (so the `functionResponse`-before-`inlineData` hazard cannot arise) — the exact "round-scoped placement" ADR 0033 RF-063-7 named. It also preserves the **single-call** shape exactly (`[model: fcA] [user: frA] [user: inlineData(A)]` — I-2).
- **Alternatives considered**:
  - **(a) Merge the media into the batched function-response turn** (the reference's normalized shape: `inlineData` parts first, then the `functionResponse`s, in one turn). Rejected — it changes the single-call shape (breaking I-2 and the round-063 pin), reintroduces the ordering hazard the reference has to heal, and forces the adapter to correlate media with its call.
  - **(b) One merged media `user` turn for the whole round** — a viable micro-variant; rejected for the smallest diff (one media message already maps to one turn; merging would add a second grouping rule for no observable gain). Recorded as a forward item.

### D4 — The wire contract is stated and pinned

- **Decision**: the adapter guarantees **`len(functionResponse parts in a turn) == len(functionCall parts in the driving model turn)`** for every round; a round with **N = 1** is byte-identical to today; a media-free conversation keeps its current shape exactly.
- **Rationale**: it is the Vertex error's own wording, so pinning it makes the regression unrepeatable.
- **Alternatives considered**: pinning only the concrete two-call example — rejected (a shape-count invariant is the honest claim).

### D5 — Verification: a unit pin over the request body (primary) + the E2E carrier (updated pin)

- **Decision**:
  - **(a) Unit (primary, red-capable)**: a pin over the built Vertex request body asserting, for a two-call round, **one** `user` turn carrying **two** `functionResponse` parts (names in call order) and the media turns **after** it; plus a **three-call** case; plus the **N = 1** and **N = 0** byte-identity controls.
  - **(b) The round-063 pin is superseded**: `TestRequestBody_MultiCallRound_MediaTurnsInterleave` (which pins today's interleave) is **rewritten** to the batched shape — a shape pin, not a behaviour claim (S-5).
  - **(c) E2E**: the Vertex-shaped fake provider already scripts a **multi-call** model turn (`vertexMultiToolCallBody`, round 019) — the acceptance journeys drive it and assert the turn completes and (for images) that both pictures reached the fake.
  - **(d) Witnesses**: reverting the batching to per-message turns reds the unit pin (non-vacuity); a witness that drops media to satisfy the count reds the two-image journey.
- **Rationale**: the transform is pure ⇒ the unit pin is a complete proof; the E2E proves the wiring and the operator-visible outcome. **No new `make verify` member.**
- **Alternatives considered**: (a) E2E-only — rejected (weaker red-capability); (b) teaching the fake to *reject* a malformed Vertex body — a nice belt-and-braces but **not required** (the unit pin is sharper and the fake is deliberately permissive); recorded as a forward item.

### D6 — `ToolCallID` stays optional; the FIFO name pairing is retained

- **Decision**: the round's `functionResponse` names continue to come from the FIFO pairing of the driving model turn's calls; a tool message without a `ToolCallID` (the replay path) is paired the same way.
- **Rationale**: minimally invasive; already sound for the sequential loop; changing to a `ToolCallID` map would be a separate concern (the reference's concurrent-matching forward item, `#36` item 3).
- **Alternatives considered**: a `ToolCallID → name` map with a FIFO fallback — deferred (not needed; recorded).

### D7 — The single-call and text paths are byte-preserved (I-2/I-3)

- **Decision**: `buildContents`' change is **additive to the multi-call case**: with **N ≤ 1** the emitted `contents` are byte-identical to today (a control asserts it), and a media-free conversation is unchanged.
- **Rationale**: rounds 013–063 text behaviour and round 063's single-image shape must not regress.
- **Alternatives considered**: none (a hard requirement).

### D8 — Governance: a **new ADR** (0035), superseding ADR 0033 **RF-063-7** and annotating its D2 note

- **Decision**: a new ADR records the batched-function-response rule + the round-scoped media placement, **supersedes ADR 0033's forward item RF-063-7**, and annotates ADR 0033 **D2**'s multi-call scope note (the "pre-existing trait this round does not introduce" clause is now resolved). ADR 0033 stays an `Accepted` historical record; only its `Status` line and the named clauses are annotated (the ADR-0027/0028/0032 precedent).
- **Rationale**: the repo's standing rule — a durable wire decision is an ADR, and an `Accepted` ADR is immutable but for annotation.
- **Alternatives considered**: (a) amend ADR 0033 in place — rejected (immutability + a distinct decision); (b) no ADR — rejected (a cross-cutting wire contract).

### D9 — No new dependency; stdlib-only; hermetic

- **Decision**: `encoding/json` only (already imported); POSIX-only; no new `go.mod` entry; every gate stays hermetic (ADR 0012). The live re-check runs at the closeout (the round-059/061/063 precedent).
- **Rationale**: round 062/063 precedent; SC-006.

## Truth impact (owner summary → `truth-delta.md`)

| Owner | Spec | Action |
| --- | --- | --- |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | **MODIFY** — the round-063 TD-063-1 qualification is replaced by the batched rule (all of a round's `functionResponse` parts in one turn; media turns after). |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | **MODIFY** — the adapter batches a round's tool results (the `functionResponse` count == the `functionCall` count). |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Local fake provider* | **MODIFY** — note the fake's existing multi-call scripting is the carrier for the parallel-round journeys (no behaviour change). |
| `/axb-technical-research` | `docs/decisions/0035-*.md` (+ index; an ADR 0033 annotation) | **ADD** — the decision + the supersession of RF-063-7. |
| `/axb-api-plan` | `specs/truth/contracts/**` | **NOOP** (no HTTP surface). |
| `/axb-data-plan` | `specs/truth/data/data-model.dbml` | **NOOP** (no persisted shape change). |
| `/axb-dsl-refine` | `specs/truth/features/cli/chat/using-a-tool.feature` (or `driving-a-vertex-gemini-model.feature`) + `chat/dsl.md` | **MODIFY** — a parallel-round Rule/Example (the turn completes; the recorded request carried the round's results together). |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): **NOOP** — this round changes no modelled entity or invariant (the wire serialization is not a domain concept); `make modelith-check` must stay green.

## Residual risks (non-blocking — forwarded)

- **The `ToolCallID`-keyed pairing** (D6) — the FIFO is retained; an explicit id map is a forward item (matching `#36` item 3).
- **A single merged media turn** (D3 variant (b)) — not adopted; a micro-variant a future round may take if a wire issue appears.
- **The fake provider does not reject a malformed Vertex body** (D5 variant (b)) — the unit pin is the carrier; a fake-side contract check is a forward item.
- **RFC 9457-style error surfacing** — the 400 already surfaces the provider's `error.message`; unchanged.
- **The other RF-063-x items** (RF-063-1 the derived ceiling, RF-063-3 aggregate bound, RF-063-4 the Files-API leg, RF-063-5 dimension guard, RF-063-8 the family-blind fixture, RF-063-9 the ceiling placement, RF-063-10 the meta-Rule clean-up) remain forward items — **not** touched by this round.

## Why this is the right shape for tellme

The defect is a **wire-serialization split** owned entirely by the adapter: the loop's per-call messages are correct for the OpenAI family, and the Gemini adapter's job is to render the conversation in the shape its wire requires. Batching a round's `functionResponse` parts in the adapter is the smallest change that closes [#132](https://github.com/gosharplite/tellme/issues/132), keeps the OpenAI wire and the single-call Gemini path byte-identical, keeps every media entry standalone, and is provable by a pure unit pin plus the existing multi-call E2E scripting. The larger questions (id-keyed pairing, a fake-side contract check, the other RF-063-x items) stay named, durable forward items.
