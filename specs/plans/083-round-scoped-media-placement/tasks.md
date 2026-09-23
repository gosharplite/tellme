# Tasks — 083-round-scoped-media-placement

**Plan Package**: `specs/plans/083-round-scoped-media-placement`
**Spec**: [`spec.md`](spec.md) · **Plan**: [`plan.md`](plan.md) · **Research**: [`research.md`](research.md)

> One task per unit of work; a task is `[X]` only after its verification passes. The test-alignment layer precedes the feature work; refactor happens under green.

## Phase 1 — Setup & Foundational

- [X] **T001** — `internal/agent/agentloop.go`: accumulate each executed call's media in `roundMedia []tools.MediaPart` inside the per-call loop (replacing the per-call `user` media append at `:216`) and, **after** the per-call loop (before `notifyToolsEnd()`), append **one** `llm.Message{Role: "user", Media: roundMedia}` when any media was collected (D1/D5; ADR 0055).

## Phase 2 — Test Alignment & Implementation (Red → Green)

- [X] **T002** — `internal/agent/agentloop_media_test.go` (new): pin the emitted message order for a **3-media** round — `assistant(tool_calls)`, then **all three** `tool` results **contiguous**, then **one** media `user` message carrying three parts (the loop-tier tripwire). **Red** before T001.
- [X] **T003** — `internal/agent/agentloop_media_test.go`: pin the **single-media** round order (`assistant(tool_calls), tool(result), user(media)` — byte-identical, I-3) and the **no-media** round (unchanged). **Red** before T001 (the single-media case is already green; the 3-media case is red).
- [X] **T004** — `tests/e2e/steps/wire_tools.go`: extend `toolExchangeChronologyOK` to assert **contiguity** — a `tool` message may be preceded by an assistant-with-tool_calls **or** another `tool` message, never a `user`/media message (D4/D5). This is the E2E contiguity tripwire.
- [X] **T005** — `tests/e2e/steps/step_r083_media_round.go` (new): the OpenAI multi-image Given `a configured provider "…" that can take images and whose endpoint asks tellme, in one step, to read "{a}", "{b}" and "{c}" and then answers with "{answer}"` (scripts ONE response of three `read_image` calls via `fakeprovider.Reply.Tools`) + the resumed-history Given `the session history already holds a turn in which the agent read "{a}", "{b}" and "{c}"` (appends one `history.Entry` with three `read_image` steps). **Red** before T001.
- [X] **T006** — `tests/e2e/steps/step_r083_media_round.go`: the Thens — `the tool results of the step were answered together, before the pictures were shown` (contiguity + media-after on the recorded request), `the step showed the three pictures together after the results` (one media message of three images), `the replayed step answered every tool result together` (the replayed conversation is contiguous). **Red** before T001.
- [X] **T007** — Run the round-083 E2E Examples for the new Rule (three pictures, one message, resumed, the Gemini companion) + the single-picture Examples — **Green**.

## Phase 3 — Truth & Records

- [X] **T008** — `specs/truth/features/cli/chat/reading-a-local-image.feature`: the round-083 Rule + 3 Examples (done at `/axb-dsl-refine`).
- [X] **T009** — `specs/truth/features/cli/chat/dsl.md`: the round-083 note + the 5 new rows + the round-063 supersession pointer (done at `/axb-dsl-refine`).
- [X] **T010** — `specs/truth/techstack.md` ×3 MODIFY (*Agent tool loop*; *Image content on the provider wire (OpenAI-compatible)*; *… (Gemini/Vertex)*) (done at `/axb-technical-research`).
- [X] **T011** — **ADR 0055** (`docs/decisions/0055-round-scoped-media-placement.md` + index) **amends ADR 0032 D7** (done at `/axb-technical-research`).
- [X] **T012** — `docs/domain-model/**` **NOT modelled** (recorded in `plan.md` §3; `modelith-check` stays green).

## Phase 4 — Gates

- [X] **T013** — `gofmt -l .` + `goimports -l .` clean; `go vet ./...` clean.
- [X] **T014** — `make verify` **OK** (layer gate 0 · `modelith-check` ×3 no drift · `verify-fmt` · `verify-adr-index` · lint 0 · govulncheck clean · cross-compile).
- [X] **T015** — `go test -count=1 ./...` **green** (E2E incl. the new Examples); `go.mod`/`go.sum` unchanged.

## Fold ledger

_Pending the round-083 architectural review (the `architect` peer)._

## Falsifiability witnesses

| # | Mutant | Expected red |
| --- | --- | --- |
| **W1** | move the media append back **inside** the per-call loop (the pre-fix code) | the loop-tier 3-media unit pin (T002) **and** the E2E contiguity Then (T006) red |
| **W2** | emit N trailing media messages (one per call after the loop) instead of one | the `one media message` Then (T006 `the step showed the three pictures together after the results`) reds |
| **W3** | put the media message **before** the tool results | the contiguity Then (T006) reds |
