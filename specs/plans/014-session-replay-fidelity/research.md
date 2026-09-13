# Phase 0 Research: tellme Session-Replay Fidelity — Persist the Tool-Call Signature (Round 014)

Topic: persist the per-tool-step provider **signature** (the Gemini 3 `thoughtSignature`) in the session history, so a **resumed** Vertex/Gemini session replays its earlier tool steps faithfully instead of failing at the provider.

Today `tellme` drives Vertex Gemini end-to-end (round 013). Round-013 **Decision 10** made the adapter capture a `functionCall`'s `thoughtSignature` and echo it when the call is replayed **within** the turn — it rides through the loop because `AgentLoop` copies `resp.ToolCalls`. What remains is the **cross-process** case: on resume, `AgentLoop.BuildMessages` re-synthesises the tool call from `history.jsonl` with **no** signature → Vertex 400 (*"Function call is missing a thought_signature in functionCall parts…"*) → `the provider request failed` (exit 6). This round persists the signature in the history record and replays it on resume.

Scope note: the language (Go 1.26), module, CLI/config/testing layers, the `llm.Gateway` port, both adapters (OpenAI-compatible + Vertex/Gemini), the session-history store (append-only JSON-Lines, rounds 007–008), the agent tool loop (round 008), and the rest of the pipeline were locked in rounds 001–013. The system still has **one CLI end**; the round reaches **no new endpoint** and adds **no new module**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; `godog` running the built binary; E2E acceptance + pure-helper units) and are **not re-decided**. Clarify Round 1 (in `spec.md`) locked: **Q1** = a dedicated nullable `signature` step field (no generic container); **Q2** = unchanged best-effort when the signature is absent. **IN**: the persisted signature, its replay on resume, and the byte-identical back-compat guarantee. **OUT**: the other 014 candidates (the Google Gemini API family, Application Default Credentials, concurrent tool-call matching).

---

## Decision 1: Persist the signature on the tool step, not on a turn-level record

- **Decision**: The provider signature is a property of a single tool execution, so it is persisted **per tool step** — a new optional field on the existing `history.Step` (`{tool, arguments, result, signature}`) — never a turn-level map. A turn with several steps carries one signature per step.
- **Rationale**: The Vertex `thoughtSignature` is emitted per `functionCall` part; a turn can hold multiple `functionCall` parts, each with its own signature. A per-step field keeps the record shape aligned with the wire (one signature per replayed call) and needs no lookup. The `llm.ToolCall.Signature` the round-013 adapter already populates is the natural source.
- **Alternatives considered**:
  - **A turn-level `signatures[]` array parallel to `steps[]`** — same information but a second array to keep index-aligned with `steps`, a needless invariant — rejected.
  - **No persistence; re-derive on resume** — the signature is an opaque provider token with no local derivation — impossible.

## Decision 2: Capture in the loop, replay in `BuildMessages`

- **Decision**: `AgentLoop` records the signature when it builds the step (`history.Step{…, Signature: tc.Signature}`), and `AgentLoop.BuildMessages` replays it into the synthesised tool call it already emits on resume (`llm.ToolCall{ID: "call_step_<n>", Name, Arguments, Signature: s.Signature}`). The synthesised deterministic id is unchanged.
- **Rationale**: These are the two existing single points — the loop already turns a `ToolCall` into a `Step` (record) and a `Step` back into a `ToolCall` (replay); adding one field to each is the minimal, symmetric change and keeps the deterministic-id (TD-2) contract intact.
- **Alternatives considered**:
  - **Persist/replay in the store adapter** — the store is a JSON round-trip and holds no wire semantics; the mapping belongs in the loop — rejected.
  - **A new resume-only code path** — duplicates the existing replay — rejected.

## Decision 3: Round-trip through the existing store; omit when empty

- **Decision**: The new field round-trips through the existing append-only JSON-Lines store with no change to the storage mechanism (one line per completed turn, the ordered steps embedded). It is serialized with `omitempty` so a step with no signature (a tool-less turn has no step; an OpenAI-family step has an empty signature) writes **exactly** the round-013 shape — every existing `history.jsonl` line stays byte-identical.
- **Rationale**: Round-007 determinism treats the on-disk line as a fixed, append-only artifact; an always-present `"signature": ""` would rewrite every historical line's shape. `omitempty` makes the change strictly additive. The exact DBML column + JSON key is the **data owner's** (`/axb-data-plan`) call — this decision fixes the mechanism, not the name.
- **Alternatives considered**:
  - **Always emit the field (even empty)** — changes every line's bytes, breaking the round-007 determinism and the "existing history unchanged" guarantee (`US2`) — rejected.
  - **A sidecar file keyed by position** — a second artifact to keep in sync; rejected as heavier than an additive field.

## Decision 4: Provider-neutral, opaque, verbatim

- **Decision**: The persisted field is a provider-agnostic, opaque string — tellme never parses, transforms, or expires it; only providers that supply a signature populate it, and only Vertex/Gemini does today. The OpenAI-compatible family leaves it empty (and its request shape is unchanged).
- **Rationale**: The history model is provider-neutral (`data-model-covers-all-state`); a Vertex-specific token is stored as an opaque blob under a neutral name so the model does not fork per provider, and the OpenAI path is provably unaffected (`FR-007`, `SC-004`).
- **Alternatives considered**:
  - **A provider-tagged structured container** — a generic `metadata`/`extensions` map (Clarify Q1 Option 2) reserves space for providers the project has not built; it is more surface than the round needs — deferred (Q1 = Option 1).
  - **A Vertex-named field (`thoughtSignature`)** — leaks a provider into the neutral model — rejected.

## Decision 5: Legacy histories keep the unchanged best-effort replay

- **Decision**: When a resumed history's step has **no** stored signature (a pre-round-014 `history.jsonl`, or a step written by a signature-less provider), tellme replays it as-is and adds **no** pre-flight detection. On a provider that requires one, the provider's own rejection surfaces as the existing `the provider request failed` (exit 6).
- **Rationale**: Clarify Q2 = Option 1. The provider's 400 is already correctly classified as the frozen provider-failure phrase; a pre-flight "tool steps but no signature" check would couple the loop to a provider-specific requirement and invent a new failure mode for a case the round's fix does not create. New histories are replayable; old ones behave exactly as today.
- **Alternatives considered**:
  - **Pre-flight detection with a clearer message** (Clarify Q2 Option 2) — earlier signal, but a new coupling + a new message for no behavioural gain — rejected.

## Decision 6: Hermetic two-process resume witness

- **Decision**: Verify the round end-to-end without real egress by reusing the existing in-process fake provider (which already serves the Vertex `:generateContent` shape and the OpenAI shape, and records the request). The witness runs the built binary **twice** over one `TELL_ME_HOME`: the first run performs the tool-using turn (persisting the step + signature); the second run resumes and re-prompts; the fake's recorded **second** request is asserted to carry the replayed `functionCall` **with** its signature. The pure layer is additionally pinned by a unit test: the history `Step` JSON round-trip preserves the signature and `BuildMessages` emits it into the synthesised `ToolCall`.
- **Rationale**: `NFR-003` + the round-007/013 precedent — the fake's recorded request is the direct oracle, and the resume witness (round 007) already drives a second process over the same history. Asserting at both layers (unit replay + E2E resume) mirrors round-010's "both layers" discipline and makes the fix falsifiable (removing the field fails both).
- **Alternatives considered**:
  - **E2E only** — would not pin the `BuildMessages` mapping in isolation — rejected.
  - **A live Vertex assertion** — non-hermetic, non-deterministic, and offline-hostile — rejected (a live run remains a manual confirmation, as in round 013).

## Decision 7: No new module; BDD techstack and strategy unchanged

- **Decision**: The round adds **no** third-party dependency (`go.mod`/`go.sum` unchanged) and does not change the BDD techstack (`godog` running the built binary) or the strategy (E2E acceptance + fast pure-helper units). The change is a struct field plus its `json` tag, its two mapping sites, and tests.
- **Rationale**: The work is plain Go over the existing store and loop; the harness already records requests.
- **Alternatives considered**:
  - **A JSON-schema/versioning library for the record** — unnecessary for one additive optional field — rejected.

## Decision 8: The three AIxBDD must-ask questions remain settled

- **Decision**: No new system end; BDD techstack = `godog`; strategy = E2E + pure-helper units — unchanged this round, per the standing `techstack.md`. No `/axb-clarify` round is owed for them.
- **Rationale**: The round adds an optional field to an existing CLI end's persisted record; it introduces no new interface, service, runner, or end, and no new system end.
- **Alternatives considered**:
  - **Re-open the techstack questions** — no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **DBML column + JSON key (Decision 3)**: the field is proven as an additive optional step field; its exact DBML column name/type and JSON key are the **data owner's** determination (`/axb-data-plan` MODIFY on `history_step`) — held open here.
- **A generic provider-metadata container (Decision 4)**: if a later provider needs bespoke persisted state, revisit a `metadata`/`extensions` map; deferred (Q1 = Option 1).
- **Concurrent tool calls (out of scope)**: `buildContents` uses a FIFO `pending` queue for tool names, sound for today's **sequential** loop; explicit tool-call-ID→name tracking is a forward item if concurrency arrives (issue #34, non-primary).
- **Legacy-history re-mint (Decision 5)**: pre-round-014 histories stay un-replayable on Gemini (unchanged `the provider request failed`, exit 6); a future "re-synth/repair missing signatures" path is explicitly out of scope.
- **Exact replay-id + signature interaction**: the synthesised `call_step_<n>` id is unchanged; the signature is attached to the same synthesised `ToolCall` (implementation detail, pinned by the unit test).
- **Deferred, still out of scope**: the Google Gemini API family (inline key), Application Default Credentials, Anthropic, streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, `SafePath`/consent.
