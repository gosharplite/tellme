# System Analysis — 083-round-scoped-media-placement

**Plan Package**: `specs/plans/083-round-scoped-media-placement`
**Spec**: [`spec.md`](spec.md) · **Research**: [`research.md`](research.md) · **Truth delta**: [`truth-delta.md`](truth-delta.md)
**Anchor issue**: [#167](https://github.com/gosharplite/tellme/issues/167)

## 1. Interface inventory

| Interface | Kind | Present? | Analysis planner |
| --- | --- | --- | --- |
| CLI (terminal: `tellme "<prompt>"`, the tool loop, the wire the fake records) | `cli` | **yes — the round's only end** | none (a plain line-oriented CLI) — carried to its contract owner **`/axb-dsl-refine`** (the `chat` module's `reading-a-local-image` feature + `chat/dsl.md`) |
| API (`contracts/**`) | `backend` | no | **NOOP** — no OpenAPI/HTTP surface |
| Data (`data/**`) | — | no shape change | **NOOP** — media is in-flight only; `history.Step` stores the tool's text result |
| Web UI | `frontend` | no | **NOOP** — no web UI |

**Waves:** one wave, one interface (the CLI). No dependency ordering is needed (a single interface).

## 2. What the round touches (code map)

| Layer | File | Change |
| --- | --- | --- |
| loop | `internal/agent/agentloop.go` | accumulate the round's media; append **one** `user` media message **after** the per-call loop (D1) |
| loop tests | `internal/agent/*_test.go` (new media-order pins) | the 3-media round order + the single-media order (D5) |
| adapter (unchanged) | `internal/infrastructure/llm/openai/client.go` | no change (relays the loop's order, now valid) |
| adapter (unchanged) | `internal/infrastructure/llm/gemini/client.go` | no change (still buffers media after the batched function-response turn; the loop now hands it one media message) |
| E2E steps | `tests/e2e/steps/wire_tools.go` | extend `toolExchangeChronologyOK` to assert **contiguity** (a `tool` may follow an assistant-with-tool_calls **or** another `tool`, never a `user`) |
| E2E steps | `tests/e2e/steps/step_r083_media_round.go` (new) | the OpenAI multi-image Given(s) + the contiguity/media-after Thens + the resumed-history Given |

## 3. Truth surfaces

| Artifact | Owner | Action |
| --- | --- | --- |
| `docs/decisions/0055-round-scoped-media-placement.md` (+ index) | `/axb-technical-research` | **ADD** — clarifies ADR 0032 D7 |
| `specs/truth/techstack.md` — *Agent tool loop*, *Image content on the provider wire (OpenAI-compatible)*, *… (Gemini/Vertex)* | `/axb-technical-research` | **MODIFY** (×3) |
| `specs/truth/features/cli/chat/reading-a-local-image.feature` + `chat/dsl.md` | `/axb-dsl-refine` | **MODIFY** — a new media-round Rule + rows |
| `specs/truth/contracts/**` | `/axb-api-plan` | **NOOP** |
| `specs/truth/data/**` | `/axb-data-plan` | **NOOP** |
| `docs/domain-model/**` | `/axb-technical-research` (F-083-4) | **MODIFY** — the `ImageContent` description + the *Reading a local image* scenario step 3 narrated a per-call fold; corrected to the round-scoped fold + re-rendered (the model is bootstrap-read + load-bearing, ADR 0041) |

## 4. Boundaries & guards

- **Layer discipline**: the change is confined to `internal/agent` (tier 4) + E2E steps; no new cross-layer edge; `internal/agent` keeps its injected `ToolLineRenderer`; the lock (`verify-architecture` 0) is preserved.
- **Single owner of the order**: the loop owns the active-turn message order (D4); the adapters keep their existing relativity (verbatim / per-family buffering).
- **Stdlib-only, POSIX-only, no new dependency**: `go.mod`/`go.sum` unchanged; the E2E is hermetic (a scripted fake, no pty, no network).
- **No storage/presentation change**: `history.jsonl`, `history.Store`, `-b`/`--back`, the offline readers, the prompt, and the live chrome are untouched; exit-code set stays **ten**; the frozen class-phrase vocabulary is unchanged.
