# Feature Specification: the media channel in-band — return media from `Execute` (round 070)

**Feature Branch**: `070-media-channel-in-band`

**Created**: 2026-09-20

**Status**: Draft (specified — clarify **not escalated by default**; see §Clarify strategy)

**Anchor**: issue [#142](https://github.com/gosharplite/tellme/issues/142) — **the round's DoD is closing it.**

**Input (operator, 2026-09-20, this session)**:

> *"Create a detail new issue for this. Open round 070, the goal is to close the new issue."*

**Behaviour intent**: **MODIFY (a structural/cohesion refactor; no behaviour change)** — make a media-producing tool's effect **in-band** (returned from `Execute` via a `domain/tools` media value, translated by the loop), **retiring the per-call `context` collector** (`llm.WithMediaCollector` / `llm.AttachMedia`) and the `internal/infrastructure/tools → internal/domain/llm` import it forced. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first — this closes the RF-062-10 lineage, and it is NOT a halving

- **This round completes `ADR 0032 §Forward RF-062-10`** (the *media-channel* half). Its *seam* half was delivered by round 069 / **ADR 0039**; the media half is currently carried as **ADR 0039 §Forward `RF-069-1`**. After this round, **RF-062-10 has no remaining half**. (The seam and the media channel were the two halves of one forward item — the reason the item spanned rounds 062→069; this ends it.)
- **Do not halve it again.** Scope is the whole media channel in one round: the in-band return path **and** the collector's deletion. If the round discovers it must split, that is a **STOP-and-re-decide** signal (issue `INV-6`) — not a new "part 2".
- **No config change, no UX change** — pure internal shape; the wire and all observable behaviour are **byte-identical**.
- **Anti-muse discipline (this round's own commitment):** the **not-taken shape** (see S-1) is recorded as a **settled decision** (a `not_planned`-style rejection in the ADR), **never** left as a live forward item. That is the exact failure mode that turned RF-062-10 into a multi-round thread. *(The accurate framing: this is a **contract/type-honesty** limit, not a Go `string` limit — image bytes *can* ride a string, but every textual consumer of the result would corrupt them, and the conversation model needs a typed media part.)*

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `d6d606f`)*

| Site | Current shape |
| --- | --- |
| `internal/domain/tools/tools.go:65` | `Execute(ctx context.Context, arguments string, budget ByteBudget) (string, error)` — **text-only**. |
| `internal/domain/llm/media.go` | `MediaPart{MIMEType, Data}` + `WithMediaCollector(ctx, *[]MediaPart)` + `AttachMedia(ctx, …)` — the **per-call context collector**. |
| `internal/agent/agentloop.go:153-154` | the loop installs a **fresh** collector per call (`var media []llm.MediaPart; tctx = llm.WithMediaCollector(tctx, &media)`), then folds `media` onto `turn = append(turn, llm.Message{Role:"user", Media: media})` (`:172-174`). |
| `internal/infrastructure/tools/image.go:136` | `llm.AttachMedia(ctx, llm.MediaPart{MIMEType: mime, Data: data})`. |
| `internal/domain/llm/gateway.go:50` | `llm.Message{ …, Media []MediaPart }` — the conversation model already carries a **typed** media field. |
| `internal/infrastructure/llm/{gemini,openai}/client.go` | `inlineDataParts` / `dataURI` serialize `llm.MediaPart` (unchanged). |

**Consumers that treat the `Execute` result as text** (why bytes cannot simply ride the string): the loop's raw-byte clamp (`clampBytes`), the `[Tool Result] <snippet>` chrome line (`logResult`, folded + rune-capped), the persisted `history.Step.Result` (JSONL), the token estimator, and the tool-role message `Content`.

---

## Design (the shape is a `/axb-technical-research` decision; the *rejected* shape is recorded settled)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **The media effect becomes in-band** — a media-producing tool returns its media through `Execute` (a `domain/tools` value), and the loop translates it onto the `user` message. The context collector is **deleted**. Shape **(a) widen the `Tool` port** vs **(b) a second, optional interface** is a technical choice. | **locked (round goal)**; shape = research |
| **S-2** | **The media type is owned by `domain/tools`** — so `internal/infrastructure/tools` no longer imports `internal/domain/llm` for it (cost 2 removed). | locked |
| **S-3** | **The emitted turn is byte-identical** — a `tool` message (same text) then a `user` message carrying the media (media-first). No wire change. | locked (I-1) |
| **S-4** | **The not-taken shape is a settled rejection** — recorded in the new ADR, **not** a forward item. | locked (anti-muse) |
| **S-5** | **The reason gate, the resource contract, the per-call timeout, and the observer hooks are unchanged.** | locked (I-3) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Behaviour byte-identical** — the emitted turn is unchanged; no wire change.
- **I-2 — No config change, no UX change** — chrome, `stdout`, exit codes, and the `--tool-usage` union unchanged.
- **I-3 — The reason gate (ADR 0025), the round-024 resource contract, and the per-call timeout are unchanged.**
- **I-4 — The layer gate stays at the 0-violation baseline**; if shape (a) is taken, the `Tool` truth row is updated with the change.
- **I-5 — No new dependency; stdlib-only; POSIX-only; hermetic.**
- **I-6 — No halving** — the media channel lands in one round; a discovered split is a STOP-and-re-decide, not a "part 2".

---

## Clarify strategy

**Not escalated by default (0 questions).** The goal is unambiguous (an operator-chosen forward item with a clear shape) and the outcome is behaviour-identity — the residual choice (which shape) is **technical** and defers to `/axb-technical-research`, with the rejected shape recorded settled (S-4). *If `/axb-technical-research` finds that the shape choice changes the formal acceptance criteria (e.g. shape (a) forces every tool's contract to change in a user-visible-to-RD way that the operator should choose), it MUST escalate a single question (one at a time) rather than assume.* **No `NEEDS CLARIFICATION` remains.**

---

## User Stories (proposed)

### US1 — A tool's media effect is in its return value (Priority: P1)

A media-producing tool returns its media through `Execute`; the loop translates it. No ambient `context` channel, no "the loop must know to install a collector", and `internal/infrastructure/tools` no longer reaches into the conversation model (`internal/domain/llm`).

**Why P1**: it is the round's whole goal — the in-band contract and the removal of the accidental cross-layer import.

**Acceptance (proposed)**: `llm.WithMediaCollector`/`llm.AttachMedia` are gone; `internal/infrastructure/tools` does not import `internal/domain/llm` (grep); an image-bearing turn still completes with the same emitted turn.

### US2 — Behaviour is provably unchanged (Priority: P1)

The refactor is invisible: the emitted turn, the wire, the chrome, and the records are identical.

**Why P1**: a structural round earns its keep only if it changes nothing observable.

**Acceptance (proposed)**: the full existing suite (unit + godog E2E, incl. `reading-a-local-image`) stays **green with no assertion changed**; `make verify` green.

---

## Functional Requirements (proposed)

- **FR-001** — A media-producing tool MUST return its media through the tool contract (a `domain/tools` media value), not via an ambient `context` collector.
- **FR-002** — The loop MUST translate the returned media onto the same `user` message it emits today (media-first, after the tool result) — byte-identical.
- **FR-003** — `llm.WithMediaCollector` / `llm.AttachMedia` MUST be **deleted**; the loop MUST NOT install a per-call collector.
- **FR-004** — The `domain/tools` media type MUST be the owner, so `internal/infrastructure/tools` no longer imports `internal/domain/llm`.
- **FR-005** — The reason gate, the resource contract, the timeout, and the observer hooks MUST be unchanged (I-3).
- **FR-006** — The not-taken shape MUST be recorded as a settled decision in the ADR (S-4).
- **FR-007** — The config schema and all UX surfaces MUST be untouched (I-2).

## Success Criteria (proposed)

- **SC-001** — `llm.WithMediaCollector`/`llm.AttachMedia` absent (grep); `internal/infrastructure/tools` free of `internal/domain/llm` (grep).
- **SC-002** — The full suite green with **no changed assertion**; the emitted turn byte-identical.
- **SC-003** — `make verify` green (layer gate 0); `go.mod`/`go.sum` unchanged.
- **SC-004** — ADR 0032 **RF-062-10 fully delivered** + ADR 0039 **RF-069-1 delivered**; the rejected shape recorded settled.

---

## Edge cases (proposed)

- A **non-media** tool call (the six readers/writers/`execute_command`/`list_skills`) — its contract is unchanged (shape (b)) or updated uniformly (shape (a)).
- **`execute_command`** (the single-writer `[Tool Output]` sink) alongside a media tool — unaffected.
- A tool that attaches **no** media but could (the return path yields an empty media slice).
- A **unit test calling a tool directly** — today `AttachMedia` is a silent no-op without a collector; under the in-band shape the media is simply in the return value (the caller decides).
- The **media fold order** (media `user` message immediately after the tool result) — must be preserved.

## Key entities

`tools.Tool` (the port) · the new `domain/tools` media value · `llm.Message.Media` (the destination) · `read_image` (the sole producer today) · `agentloop` (the translator) · `MediaPart` (the current `domain/llm` type — relocated or re-exported).

## Assumptions

- A1 — The change is **loop/tools-local** (`internal/domain/tools` + `internal/agent` + `internal/infrastructure/tools` + the truth rows); the adapters, the CLI, and the config are untouched.
- A2 — `read_image` is the **only** media producer today; the shape must nevertheless accommodate a future second producer.
- A3 — The emitted turn (text result + a `user` media message) is the invariant the round preserves.
