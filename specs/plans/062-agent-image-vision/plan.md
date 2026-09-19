# Plan — Agent image vision: `read_image` to the provider wire (round 062)

**Plan Package**: `specs/plans/062-agent-image-vision`
**Truth root**: `specs/truth` · **Interface kind**: `cli` (plain line CLI — no `ui/**`, no TUI surface)

## Interfaces inventoried

| Interface | Kind | Planner | Wave |
| --- | --- | --- | --- |
| The CLI chat end (`specs/truth/features/cli/chat/**`) | `cli` | **carried to its contract owner** `/axb-dsl-refine` (a line CLI has no API/data/UI planner) | 1 |

- `/axb-api-plan` — **NOOP** (no OpenAPI/HTTP surface; a single CLI end).
- `/axb-data-plan` — **NOOP** (checked): the round adds **no persisted shape** — a `read_image` step persists its **text** result only (`history_step.result`), and the image itself is **in-memory, in-flight, never written** (ADR 0032 §Forward RF-062-6). The in-memory conversation (`llm.Message`) is not within the dbml's modelled scope (the file models the **persisted** local-state logs only).
- `/axb-ui-plan` — **skipped** (plain line CLI; no TUI surface added — the tool's result is a chrome-free `[Tool …]` exchange, and no screen/keybinding changes).

## Waves

**Wave 1 (single, no dependencies)** — the CLI contract owner refines the acceptance journey into executable truth:
- a new journey for reading a local image with the agent (`chat/reading-a-local-image.feature`), carrying Rules for: the image reaching the wire (content-sniffed kind), the capability-dependent offered set, and the loud refusals;
- the **offered-set** feature (`chat/offering-the-agent-tools.feature` + `chat/dsl.md`) MODIFY-ed to state that the set is a **function of the selected provider's capability** and to add `read_image` under a vision-enabled provider (round 029/033 territory);
- the new DSL rows for the new Given/Then sentences.

## The change surface (RD)

| # | Site | Change |
| --- | --- | --- |
| 1 | `internal/domain/llm/gateway.go` | `Message` gains an optional `Media []MediaPart` (`MediaPart{ MIMEType string; Data []byte }`); a media-less message is unchanged (research D5) |
| 2 | `internal/config/config.go` (+ `Provider`) | the provider entry gains `VISION bool` (default false); exposed on the resolved provider the factory reads (research D3) |
| 3 | `internal/infrastructure/llm/openai/client.go` | `requestBody` serializes a media-bearing message as a content **array** (leading text part, then one `image_url` data-URI block per media part); media-less messages stay string-content (byte-identical) (research D2/D5) |
| 4 | `internal/infrastructure/llm/gemini/client.go` | if a message carries media, return a **loud** `*llm.ProviderError` — the family has no `inline_data` path this round (no silent drop) (research D8) |
| 5 | `internal/agent/agentloop.go` | a tool that produced media appends **one `user` message** (media-first) after the round's `tool`-result message(s) (research D4/D7); the text-only path is unchanged |
| 6 | `internal/domain/tools/tools.go` (or a small sibling) | the `Tool` result gains an optional media channel so a tool can attach an image (the exact shape — a sibling `Media()` output vs an extended return — is an implementation choice for `/axb-tasks`) |
| 7 | `internal/infrastructure/tools/image.go` (new) | the `read_image` tool: read → **content sniff** (JPEG/PNG/GIF/WebP magic bytes) → 32 MiB ceiling check → attach media; loud errors for an unknown type / oversize (research D6/D7) |
| 8 | `cmd/tellme/deps.go` | `agentTools()` gains a **vision flag**; `read_image` is assembled only when the selected provider declares `VISION: true`; the round-031 well-formedness gate covers both variants (research D9) |
| 9 | `tests/e2e/steps/**` (new stepdefs) + the fake provider | the E2E carrier: the recorded request body carries the image; the offered set by capability; the loud refusals |
| 10 | `internal/infrastructure/tools/image_test.go`, `internal/infrastructure/llm/openai/client_test.go` (pins) | the hermetic unit pins: the wire block (exact bytes, sniffed MIME), byte-identity control, the sniff table, the ceiling boundary, media-on-Gemini refusal (research D11) |

## Layer/architecture notes

- The capability **key** lives in config; the **sniff + ceiling + tool** live in `internal/infrastructure/tools`; the **wire shape** lives in the openai adapter; the **placement** logic lives in the loop. Each seam is one concern.
- The domain change is **one optional field** on `llm.Message` (+ a `MediaPart` value) — no new domain package, no cross-layer edge; `verify-architecture` should be untouched (the assembler stays in `cmd/tellme`).
- No dependency change (`go.mod`/`go.sum` unchanged); stdlib only (no image library — magic bytes only); POSIX-only.

## Test strategy

| Layer | Carrier |
| --- | --- |
| Unit (sniff) | a table pin: JPEG/PNG/GIF/WebP magic bytes → MIME; an unknown/empty content → the loud error |
| Unit (wire) | a fake `http.RoundTripper` over `openai.Complete`: the image block present, decodes to the file's **exact** bytes, the **sniffed** MIME; a text-only control is **byte-identical** to the pre-round body (I-1) |
| Unit (ceiling + refusal) | the 32 MiB boundary both ways; a non-image content; media-on-Gemini → the loud `*llm.ProviderError` |
| Unit (capability) | the assembler offers/omits `read_image` by the flag; the round-031 `TestAgentToolSchemasAreWellFormed` still passes for both variants |
| E2E | the truth feature's new journey (fake provider records the request; the tool offered/omitted by capability; the loud refusals) |

Witnesses (reproduced then reverted): (a) removing the image serialization turns the openai unit pin RED; (b) inverting the capability flag flips the offered-set assertion RED; (c) dropping the ceiling check lets an oversize image reach the request (asserted RED).
