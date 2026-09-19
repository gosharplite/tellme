# Plan — Gemini/Vertex parallel tool calls: a round's tool results share one turn (round 065)

**Plan Package**: `specs/plans/065-gemini-parallel-tool-calls`
**Truth root**: `specs/truth` · **Interface kind**: `cli` (plain line CLI — no `ui/**`, no TUI surface)
**Anchor issue**: [#132](https://github.com/gosharplite/tellme/issues/132)

## Interfaces inventoried

| Interface | Kind | Planner | Wave |
| --- | --- | --- | --- |
| The CLI chat end (`specs/truth/features/cli/chat/**`) | `cli` | **carried to its contract owner** `/axb-dsl-refine` (a line CLI has no API/data/UI planner) | 1 |

- `/axb-api-plan` — **NOOP** (no OpenAPI/HTTP surface; a single CLI end).
- `/axb-data-plan` — **NOOP** (checked): the round changes **no persisted shape** — the `Turn` line and its `steps` are unchanged, and media is in-flight only (ADR 0032 D7). The change is **wire-serialization inside the Gemini adapter**.
- `/axb-ui-plan` — **skipped** (plain line CLI; no screen/keybinding change).

## Waves

**Wave 1 (single, no dependencies)** — the CLI contract owner refines the round-065 acceptance journey into executable truth:
- a **new** `chat/calling-several-tools-in-one-round.feature` carrying three Rules — a multi-call round finishes (two files); a multi-picture round still shows the model every picture (reusing the round-062/063 `the request carried the image file` rows); a single call is unchanged;
- **`chat/dsl.md`** MODIFY-ed: a `## Given (round 065)` block (a Vertex provider scripting **two** tool calls in one model turn; a media/non-media pair) + a `## Then (round 065)` row (`the Gemini provider received the answer to both read requests together`) + a round-065 module note.

## The change surface (RD)

| # | Site | Change |
| --- | --- | --- |
| 1 | `internal/infrastructure/llm/gemini/client.go` | **MODIFY** `buildContents`: a model turn's **N** `functionCall` parts ⇒ the round's **N** tool results are **batched into one `user` `contents` entry** (N `functionResponse` parts, call order); the round's **media** messages are **buffered** and emitted **after** the batched turn (round-scoped). A `flush` boundary fires on the next model/plain-text turn and before the prompt. `inlineDataParts` and the `pending` FIFO are reused; no other helper changes. |
| 2 | `internal/infrastructure/llm/gemini/client_image_test.go` | **REWRITE** `TestRequestBody_MultiCallRound_MediaTurnsInterleave` → the batched shape; **ADD** a media-free multi-call batch pin; keep the N = 1 and media-free byte-identity pins. |

## Layer/architecture notes

- The **whole change is inside `internal/infrastructure/llm/gemini`** (the adapter) — no loop, port, tool, config, domain, or composition-root change; `verify-architecture` untouched (no new edge).
- The loop's per-call messages (and the OpenAI-compatible wire) are **unchanged** — the OpenAI family is byte-identical (I-1).
- **No dependency change** (`go.mod`/`go.sum` unchanged); stdlib `encoding/json` already imported; POSIX-only; hermetic.

## Test strategy

| Layer | Carrier |
| --- | --- |
| Unit (wire, primary) | `requestBody` pins: a **two-call** round ⇒ **one** `user` turn with **two** `functionResponse` parts (call order) + the media turns **after** it; a **three-call** media-free round ⇒ one turn with three parts; the **N = 1** media round ⇒ byte-identical to round 063 (contents len 3); a media-free conversation ⇒ byte-identical |
| Unit (contract) | the count invariant `len(functionResponse) == len(functionCall)` per driving model turn, asserted on the decoded `contents` |
| E2E | the new `chat/calling-several-tools-in-one-round.feature`: the Vertex-shaped fake scripts **two** `read_files` calls in one model turn; the run completes and the recorded request carries a **single** `user` turn with both `functionResponse` parts |
| Witness (reproduced then reverted) | reverting the batching to per-message turns reds the unit pin **and** the E2E shape assertion (non-vacuity) |

## Residual risks (non-blocking — see ADR 0035 §Forward)

`ToolCallID`-keyed pairing · one merged media turn · a fake-side contract check · a family-agnostic batching concept · the untouched RF-063-x items.
