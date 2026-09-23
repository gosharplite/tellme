# Technical Research — 083-round-scoped-media-placement

**Plan Package**: `specs/plans/083-round-scoped-media-placement`
**Spec**: [`spec.md`](spec.md)
**Anchor issue**: [#167](https://github.com/gosharplite/tellme/issues/167)
**Status**: complete — decisions **D1–D9** below are locked; residuals live in **ADR 0055 §Forward**.

> Decision-driven research for the round. The spec (behaviour) is fixed by the issue + the operator directive; this document decides the *technical* shape: where the round's media is folded, the exact N ≥ 2 emitted shape, the witness tiers, the truth/record surfaces, and the scope guard.

---

## 1. Problem & grounding

A model response may request **several** tools in one round. Each `read_image` call returns its media **in-band** (round 070 / ADR 0040) and the loop folds it onto the active turn. The loop today appends, **inside the same per-call loop**, the `tool` result and then — when that call attached media — a `user` media message (`internal/agent/agentloop.go:211`/`:216`). The OpenAI-compatible adapter relays each message **verbatim** (`requestBody`), so a round with N ≥ 2 media calls emits `assistant(tool_calls), tool(r1), user(img1), tool(r2), user(img2), …` — an interleaved non-`tool` message. The provider requires the `tool_calls` block to be answered **contiguously**, so the **next** request of the turn is rejected `400` (exit 6). The Gemini/Vertex adapter buffers media and emits it after the batched function-response turn, so it is correct by construction; the single-call case (`assistant, tool, user(media)`) is a complete block and works.

Measured at the round start (`dev` @ `426a93f`):

| Site | Shape |
| --- | --- |
| `internal/agent/agentloop.go` `Run` | `turn = append(turn, llm.Message{Role:"assistant", ToolCalls: resp.ToolCalls})` (`:143`); per call: `turn = append(turn, llm.Message{Role:"tool", Content: result, ToolCallID: tc.ID})` (`:211`) then `if len(media) > 0 { turn = append(turn, llm.Message{Role:"user", Media: media}) }` (`:216`). |
| `internal/agent/agentloop.go` `BuildMessages` | the replay projection: `user(prompt)`, per step `assistant(tool_call)`, `tool(result)`, then `assistant(answer)` — already contiguous (media is not persisted). |
| `internal/infrastructure/llm/openai/client.go` `requestBody`/`messageContent` | relays each `llm.Message` verbatim; a media-bearing message becomes a content **array** (optional text part + one `image_url` data URI per media part). |
| `internal/infrastructure/llm/gemini/client.go` `roundBuilder.consume`/`flush` | buffers a media-bearing message into `mediaTurns`; `flush()` emits the batched function-response turn, then the media turns. |
| `tests/e2e/steps/wire_tools.go:169` `toolExchangeChronologyOK` | asserts a `tool` message is **preceded** by an assistant-with-tool_calls and a user message — **no** contiguity, and its fixtures never use a media tool ⇒ **nothing can redden today** (`GAPS.md` / `aixbdd-tmg#15`). |

`read_image` returns media via the optional `tools.MediaTool` capability (`ExecuteMedia(ctx, args, budget) (text, []tools.MediaPart, err)`); the loop type-asserts it and collects `media`. The mid-loop `continue` paths (an unknown tool name, a reason-less refusal) run no tool and yield **no** media, so they contribute nothing.

---

## 2. Decisions

### D1 — Fold a round's media ONCE, after the per-call loop (round-scoped), on the loop's active turn

`Run` accumulates `var roundMedia []tools.MediaPart` **inside** the per-call loop (`roundMedia = append(roundMedia, media...)` at the call site) and, **after** the loop (before `notifyToolsEnd()`), appends **one** message when any media was collected:

```go
if len(roundMedia) > 0 {
    turn = append(turn, llm.Message{Role: "user", Media: roundMedia})
}
```

This makes every `tool` message of the round **contiguous** (spec **FR-001**), delivers the round's N media parts in **one** `user` message **after** the results (spec **FR-002**), and keeps the emitted parts in **call order** (deterministic — spec **NFR-001**). The message is appended to `turn` before the next loop iteration, so the **next** request carries it (after all results), which is exactly the shape the wire requires.

**Why one message and not N messages after the loop:** N trailing media messages would also restore contiguity, but the issue's sketch (and the multi-part content array the adapters already build) is a **single** media message carrying all parts; one message is the smaller surface and is byte-identical to today for N = 1. N messages is rejected (more wire messages for no benefit; the adapters already render an arbitrary media list as parts/blocks of one message).

**Why the loop, not the adapter:** the issue's root cause is the **loop's order**, and the loop is where a round's media is known. Making each adapter re-order would duplicate the fix across families and leave the loop emitting an invalid order for any future adapter. The loop is the single owner of the active-turn message order (spec **FR-007**); the adapters keep their existing job (render the messages verbatim / buffer per family).

### D2 — Single-call output is byte-identical

For N = 1 the emitted order stays exactly `assistant(tool_calls), tool(result), user(media)` — the same messages, in the same order, with the same bytes. The shipped single-image path (rounds 062/063) is unchanged (spec **FR-005** / **I-3**); the existing `reading-a-local-image.feature` Examples stay green.

### D3 — Gemini placement is unchanged; a benign cardinality consequence is recorded

The Gemini/Vertex adapter still buffers media and emits it **after** the batched function-response turn (spec **FR-006** / **I-4**). Because the loop now hands it **one** media message per round (D1), the Gemini multi-media path emits **one** media turn carrying N `inlineData` parts (in call order) instead of N single-part media turns. This is a **shape change for the Gemini multi-media path only** — the batched function-response turn is untouched, every `inlineData` turn stays **standalone** (no `inlineData`-before-`functionResponse`), and the turn still completes. It is recorded in the ADR; it is not a failure and needs no adapter change.

### D4 — Witness tiers: a loop-tier unit pin + an E2E contiguity witness, both failing pre-fix

This round is the `GAPS.md` / `aixbdd-tmg#15` "claim-without-a-tripwire" class — no test can redden on the bug today. The round adds the tripwire at two tiers (spec **NFR-004** / **I-7**):

1. **Loop tier (`internal/agent`)** — a unit pin over the emitted message order for a **3-media round**: assert `assistant(tool_calls)`, then **all three** `tool` results (contiguous), then **one** media `user` message carrying three parts; and a companion pin that the **single-media** round keeps `assistant, tool, user(media)`. The loop's existing tests already drive a fake gateway/registry, so this needs no new harness.
2. **E2E tier** — extend `toolExchangeChronologyOK` to assert **contiguity** (no non-`tool` message between a round's `tool` results) and add a fixture that scripts a **multi-image round** on the OpenAI-compatible family, plus a **resumed-session** variant and a **Gemini multi-media** companion. The existing single-image Examples remain the N = 1 carrier.

**Mutation witness (reproduced then reverted):** move the media append back inside the per-call loop (the pre-fix code) ⇒ `TestRunRoundScopedMedia_ThreeCalls_FoldedOnceAfterResults` reddens (`emitted messages = 8, want 6`) **and** the E2E `The model inspects three pictures in one step` reddens (`314 → 313 passed`) via the shared `toolExchangeChronologyOK`. The single-media pin, the no-media pin, the resumed Example, and the Gemini companion are **guards** (green pre- and post-fix) — see D5.

The E2E contiguity check is **single-owned** by the shared `toolExchangeChronologyOK` (`wire_tools.go`, extended with contiguity), which the round-083 multi-image fixture is the **carrier** for (the helper is invoked by the round-083 Thens); no second contiguity predicate exists (round-083 fold F-083-6 / TD-083-1). The **resumed-session** and **Gemini multi-media** Examples are **guards** (green pre- and post-fix), not tripwires (TD-083-2/3).

**Alternative rejected — an *enforcing* fake provider** that returns `400` when a request's tool block is broken (so the E2E would observe a real rejection): considered, and **not** adopted as the primary witness — it mutates a **shared** fake used by every scenario (a risk to unrelated suites) and duplicates a rule the assertion states directly; the assertion witness is deterministic and localised. It stays a §Forward option.

### D5 — Canonical message order (the invariant, stated once)

For one round with calls `c1…cN` (of which the media-producing ones yield `m1…mk`):

```text
assistant(tool_calls: c1…cN)
tool(result: c1)
…                       (every result of the round, contiguous)
tool(result: cN)
user(media: m1…mk)       (one message; only when k > 0; in call order)
```

No `user`/`assistant`/media message may sit between `tool(result: c1)` and `tool(result: cN)` (spec **I-1**). The replay path (`BuildMessages`) already obeys this (it carries no media); the active-turn path now does too.

### D6 — Truth & records

- **ADR 0055** (`docs/decisions/0055-round-scoped-media-placement.md`) — **clarifies ADR 0032 D7** (D7's *body* reads round-scoped; its *heading* + the pre-083 **implementation** were per-call; this ADR makes the implementation match the body and adds the back-pointer to ADR 0032's `Status` + D7); indexed in `docs/decisions/README.md`. Records D1–D6 + the Gemini cardinality consequence (D3) + the scope guard.
- **`specs/truth/techstack.md`** — MODIFY the **Agent tool loop** row (the round-083 media-fold-once clause) + the **Image content on the provider wire (OpenAI-compatible)** row (the round's placement + contiguity) + the **Image content on the provider wire (Gemini/Vertex)** row (the recorded multi-media cardinality consequence).
- **`specs/truth/features/cli/chat/reading-a-local-image.feature`** — a new Rule ("several pictures in one step are shown together, after every tool result") + `chat/dsl.md` rows; **`/axb-dsl-refine`** owns this.
- **`docs/domain-model/**`** — **MODIFY** (round-083 fold F-083-4): the model *does* narrate the placement (the `ImageContent` description + the *Reading a local image* scenario step 3 both said "after the tool result"), so the two sentences are corrected to "after the **round's** `tool` result(s)" and the model re-rendered (the model is bootstrap-read + load-bearing, ADR 0041). The modelled entities/invariants are unchanged (the placement is in-flight, never persisted).
- **`specs/truth/contracts/**`** + **`specs/truth/data/**`** — **NOOP** (no API surface; no persisted-state shape change).

### D7 — Determinism, hermeticity, no new dependency

The change is a single accumulation + one append over the already-executed round; stdlib-only; POSIX-only; no new dependency; `go.mod`/`go.sum` unchanged. The E2E scripts a multi-tool-call response via the existing `fakeprovider.Reply.Tools` seam and asserts the recorded wire — no pty, no network.

### D8 — Verification strategy

- **Unit (`internal/agent`)**: the 3-media round order (D4.1) + the single-media order (D2); the media parts are in call order.
- **Unit (optional, `internal/infrastructure/llm/openai`)**: `messageContent` renders a multi-part media list as one content array (an existing concern; no new behaviour).
- **E2E (`tests/e2e/steps/`)**: the contiguity helper extension + the multi-image fixture (fresh + resumed) + the Gemini companion.
- **Falsifiability witnesses** (reproduced then reverted): (a) move the media append inside the per-call loop ⇒ the loop-tier unit pin + the E2E contiguity witness red; (b) emit N trailing media messages instead of one ⇒ the `one media message` assertion reds (a shape pin); (c) reorder media by a map ⇒ the call-order pin reds (defensive).

### D9 — Scope guard (excludes) & residual risks

**Excluded:** media persistence / payload re-budget (the images still ride the next round's active-turn messages); image dedupe / downscale / per-round byte accounting; a `--image` flag or config toggle; any change to `read_image`'s contract or the family-aware inline ceiling. No new exit code; no new class phrase; the exit-code set stays **ten**. **Residual risks → ADR 0055 §Forward** (below).

---

## 3. Residual risks (homed in ADR 0055 §Forward)

- **RF-083-1** — the round's media still rides the **next** round's request (active-turn growth); a payload/dedupe mitigation is out of scope.
- **RF-083-2** — the E2E contiguity witness asserts a **wire property** (a predicate), not a byte-golden; the byte identity for N = 1 is carried by the existing single-image Examples + a unit pin.
- **RF-083-3** — the Gemini multi-media cardinality consequence (one media turn of N parts) is now **carried by `TestRequestBody_RoundScopedMedia_OneTurnTwoParts`** (added in the round-083 fold F-083-5); the live end-to-end Gemini path is still only companion-guarded.
- **RF-083-4** — the *enforcing-fake* alternative (a 400 on a broken tool block) is not adopted (D4).
- **RF-083-5** — a round that mixes media-producing and media-free calls is covered by construction (only the producing calls contribute), but no dedicated Example scripts a mixed round.
- **RF-083-6** — `BuildMessages` (the replay path) carries no media, so a resumed session replays text results only; unchanged and by design.
