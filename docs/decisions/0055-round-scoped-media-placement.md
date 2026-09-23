# ADR 0055 — Round-scoped media placement (a tool round's media is folded once, after all tool results)

- **Status:** Accepted
- **Date:** 2026-09-23
- **Deciders:** tellme owner
- **Related:** [ADR 0032](0032-agent-image-vision.md) (D7 — the **media-placement clause this clarifies**: its body read round-scoped, its heading did not), [ADR 0033](0033-gemini-image-vision.md) (the Gemini `inlineData` wire), [ADR 0035](0035-gemini-parallel-tool-calls.md) (the batched function-response turn), [ADR 0040](0040-media-channel-in-band.md) (the in-band `tools.MediaTool` media channel), [ADR 0048](0048-recoverable-unknown-tool-name.md) (a fold-back path that yields no media), [ADR 0052](0052-failed-turn-partial-persistence.md) (the failed-turn persistence the pre-fix 400 would trigger); issue [#167](https://github.com/gosharplite/tellme/issues/167); `GAPS.md` / [aixbdd-tmg#15](https://github.com/gosharplite/aixbdd-tmg/issues/15) (the claim-without-a-tripwire class); round 083 (`specs/plans/083-round-scoped-media-placement`)

## Context

A model response may request **several** tools in one round. Each `read_image` call returns its media **in-band** (round 070; ADR 0040), and the agent loop folds it onto the active turn.

ADR 0032 **D7** ("Media placement: a `user` message after the tool result, media-first") has an **ambiguous heading** but a **round-scoped body**: *"When a round executes a tool that produced media, the loop appends **one `user` message** whose content array leads with the image block(s), immediately **after** the round's `tool`-result message(s)."* The **pre-083 implementation** read it per-call: `internal/agent/agentloop.go` appended a media `user` message **inside** the per-call loop (after each `tool` result), so a round with N ≥ 2 media-producing calls emitted:

```text
assistant(tool_calls:[c1,c2,c3])
tool(r1)  user(img1)  tool(r2)  user(img2)  tool(r3)  user(img3)
```

The **OpenAI-compatible** adapter relays each message **verbatim**, so this interleaved non-`tool` message reached the wire — and the provider family requires an assistant `tool_calls` message to be answered by its `tool` messages **contiguously** (`An assistant message with 'tool_calls' must be followed by tool messages responding to each 'tool_call_id' …`). The **next** request of the turn was therefore rejected `400` (exit 6), and the round-080 failure path then persisted the completed steps. A **single** media call (`assistant, tool, user(img)`) is a complete block and worked; the shipped rounds 062/063 were verified with one image only, so the per-call placement was never exercised at N ≥ 2.

The **Gemini/Vertex** adapter (`roundBuilder.consume`/`flush`) buffers media and emits it after the batched function-response turn — accidentally correct by construction.

No test could redden on the bug: the only wire-order witness, `toolExchangeChronologyOK` (`tests/e2e/steps/wire_tools.go`), asserted that a `tool` message is **preceded** by an assistant-with-tool_calls and a user message — never **contiguity** — and its fixtures never used a media tool (`GAPS.md` / aixbdd-tmg#15 "claim-without-a-tripwire").

## Decision

**D1 — a round's media is folded ONCE, after the per-call loop (round-scoped).** `internal/agent.AgentLoop.Run` accumulates the media each executed call returns (`roundMedia = append(roundMedia, media...)` at the call site) and, **after** the per-call loop and before `notifyToolsEnd()`, appends **one** `user` message when any media was collected:

```go
if len(roundMedia) > 0 {
    turn = append(turn, llm.Message{Role: "user", Media: roundMedia})
}
```

This makes the implementation match ADR 0032 D7's body — every `tool` message of a round is **contiguous** — and delivers the round's media **once**, **after** the results, in **call order**.

**D2 — single-call output is byte-identical.** For N = 1 the emitted order stays exactly `assistant(tool_calls), tool(result), user(media)`; the shipped single-image path is unchanged.

**D3 — the Gemini adapter is unchanged; recorded cardinality consequence.** The Gemini/Vertex adapter still buffers media and emits it **after** the batched function-response turn. Because the loop now hands it **one** media message per round, the Gemini multi-media path emits **one** media turn carrying N `inlineData` parts (in call order) instead of N single-part media turns — a benign shape change for the multi-media path only; the batched function-response turn and the standalone-`inlineData` guarantee are untouched. (Carried by `TestRequestBody_RoundScopedMedia_OneTurnTwoParts`.)

**D4 — the loop owns the active-turn message order.** The fix lives in the loop (the site that knows a round's media and owns the turn's chronology), not in each adapter; the adapters keep their existing job (render the messages verbatim / buffer per family). No adapter change is needed.

**D5 — the witness the round adds (the missing tripwire).** The tripwire that **reddens** under the pre-fix ordering is: (1) a **loop-tier unit pin** over the emitted message order for a **3-media** round (`assistant(tool_calls)`, then all `tool` results contiguous, then one media `user` message of three parts) — `TestRunRoundScopedMedia_ThreeCalls_FoldedOnceAfterResults`; and (2) an **E2E** contiguity witness — the **shared** `toolExchangeChronologyOK` (extended with contiguity) **as exercised by the round-083 multi-image fixture** (`step_r083_media_round.go`). No second contiguity predicate is defined (single owner). The single-media pin (`…SingleCall_ByteIdentical`), the no-media pin, the **resumed-session** Example, and the **Gemini multi-media** Example are **guards**, not tripwires: they are green pre- and post-fix (N = 1 is byte-identical; the replay path carries no media; the Gemini adapter already buffered). Measured W1 (the true pre-fix per-call placement): **one** loop-tier pin red (`ThreeCalls_…`, `emitted messages = 8 want 6`) and **one** E2E Example red (`314 → 313 passed`).

**D6 — placement-only.** No `history.jsonl` schema, `history.Store`, `-b`/`--back`, interactive-prompt, or live-chrome change; media is never persisted (`history.Step` stores the tool's **text** result), so a resumed session replays contiguous text steps and the images remain in-flight only. No new exit code (the set stays **ten**); the frozen class-phrase vocabulary is unchanged.

**D7 — scope guard.** Excluded: media persistence / payload re-budget; image dedupe / downscale / per-round byte accounting; a `--image` flag or config toggle; any change to `read_image`'s contract or the family-aware inline ceiling. Also **not adopted** as the primary witness: an *enforcing* fake provider that returns `400` on a broken tool block (a shared-fake mutation that duplicates a directly-stateable assertion) — kept a forward option.

**D8 — records.** ADR 0055 (+ index) **clarifies ADR 0032 D7** (the round-scoped reading of its body is now the implementation); `techstack.md` *Agent tool loop* + *Image content on the provider wire (OpenAI-compatible)* + *Image content on the provider wire (Gemini/Vertex)* MODIFY; the `chat` CLI feature (`reading-a-local-image`) + `chat/dsl.md` rows (`/axb-dsl-refine`); `docs/domain-model/**` **MODIFY** — the two narrative statements that described a per-call fold (the `ImageContent` description and the *Reading a local image* scenario step 3) are corrected to "after the **round's** tool result(s)" (the model is bootstrap-read and load-bearing, ADR 0041); `contracts/**` + `data/**` **NOOP**.

## Consequences

- A turn with **≥ 2** `read_image` calls in one round **completes** on the OpenAI-compatible family (no 400): all `tool` results are answered contiguously, then the round's images ride one `user` message.
- The single-image path stays **byte-identical**; the Gemini adapter is unchanged (with the recorded one-media-turn cardinality consequence for N ≥ 2).
- The `assistant(tool_calls) → tool… → user(media)` order is now a stated invariant of the loop's active turn, matching the replay path (`BuildMessages`), which already obeyed it.
- The class defect (`a normative clause no input could falsify`) is closed for this surface by the loop-tier pin + the E2E contiguity witness (D5).
- No new dependency; stdlib-only; POSIX-only; `go.mod`/`go.sum` unchanged.

## Alternatives considered

- **N trailing media messages after the loop** — restores contiguity, but emits N wire messages where one suffices and is not byte-identical to today for N = 1; rejected (D1).
- **Re-order in each adapter** (make the OpenAI adapter buffer like the Gemini one) — duplicates the fix per family and leaves the loop emitting an invalid order for any future adapter; rejected (D4).
- **An enforcing fake provider** (400 on a broken tool block) as the witness — mutates a shared fake used by every scenario and reproduces a rule the assertion states directly; rejected as the primary witness (D7), kept a forward option.
- **Per-call media placed after *all* results by index** (e.g. `user(img1)`… still separate) — no benefit over one message; rejected (D1).
