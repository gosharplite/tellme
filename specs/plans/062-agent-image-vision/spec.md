# Feature Specification: agent image vision — `read_image` to the provider wire (round 062)

**Feature Branch**: `062-agent-image-vision`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from an **operator request** (no anchor issue). **Clarify CLOSED** (5 decisions, one at a time — **Q1 → 1**, **Q2 → 1**, **Q3 → 1**, **Q4 → 1**, **Q5 → 1**; recorded below). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19, this session)**:

> *"Can tell-me-go read image with deepseek-flash?"* → (grounded answer: the reference gates vision on a model-ID substring, and its ADR-070 assumed `deepseek-v4-flash` was text-only) →
> the operator supplied the **official DeepSeek doc**: *"The deepseek-flash model accepts images alongside text … The legacy model name deepseek-v4-flash-vision-exp is still accepted, but the model has been retired and its requests are served by the latest Flash model as well. Supported image formats: JPEG, PNG, GIF, and WebP. The format is detected from the actual file content, not from the file name or the declared MIME type."* →
> *"We need to refer tell-me-go and make tellme to have vision."*
> → entry-point selection: **A — an agent tool**.

**Anchor issue**: none (operator request; the round supersedes the reference's stale capability assumption **in tellme**, and records the divergence).

**Behaviour intent**: **ADD (a new agent-tool capability + a new provider-capability config key + a new media content shape on the existing OpenAI-compatible wire).** tellme today is **text-only end to end**: the domain `llm.Message.Content` is a `string`, both transports emit text-only `parts`/`content`, the domain `tools.Tool` result is a `string`, and no tool produces or attaches media. This round makes the agent able to **read a local image and show it to a vision-capable provider** — the first non-text capability in tellme — while keeping the existing text path byte-identical.

---

## Grounded in the current system *(measured 2026-09-19, `dev` @ `902642a`)*

| Site | Current shape |
| --- | --- |
| `internal/domain/llm/gateway.go` | `Message{ Role, Content string, ToolCalls, ToolCallID }` — the conversation is **text-only**; `Request{Prompt, Messages, Tools}`; no media/part type exists anywhere in the domain. |
| `internal/infrastructure/llm/openai/client.go` | `requestBody` emits `{"role":…, "content": <string>}` messages; the tool-result message is `{"role":"tool","tool_call_id":…,"content":<string>}`. No `image_url` block, no content-array form. |
| `internal/infrastructure/llm/gemini/client.go` | `buildContents` emits `{text: …}` parts only; tool results are `{"role":"user","parts":[{"functionResponse":…}]}`. No `inline_data`. |
| `internal/domain/tools/tools.go` | `Tool.Execute(ctx, arguments string, budget ByteBudget) (string, error)` — a tool result is **a string**; no binary/media channel. |
| `internal/infrastructure/tools/*.go` | The tool surface is the reader trio + write pair + `execute_command` + `list_skills`; **no media tool**. `isBinary` (binary.go) probes a NUL byte but never identifies an image. |
| `internal/domain/llm/pricing.go` + `internal/config/config.go` | The provider registry carries `TYPE`/`MODEL`/`URL`/`API_KEY`/`MAX_TOKENS`/`THINKING_*`/`HEADERS`; there is **no capability key** — capability today is implied by the family (`ProviderFamily`). |
| `specs/truth/features/cli/chat/offering-the-agent-tools.feature` + `chat/dsl.md` | The offered tool set is an **exact, tested** truth surface (round 029: "six tools"; round 033: "seven"). A capability-dependent tool changes this truth. |
| `cmd/tellme/deps.go` `agentTools()` + round-031 `TestAgentToolSchemasAreWellFormed` | tellme's own declared tools are built in one non-overridable assembler and gated for `required ⊆ properties`. A new tool + a capability input touch both. |
| `tell-me-go` (reference) | `llm.Part{Text, InlineData *Blob, FunctionCall, FunctionResponse}`; `tools.ToolResult{Text, BinaryData []BinaryData}`; a `read_image` tool returning `BinaryData`; `executor.AssembleResponse` folds media parts into the model-visible conversation (**media parts first**, role `user` for Gemini, `functionResponse` after); the OpenAI client serializes media as `image_url` blocks. Capability resolved by model-id substring (`vision`) — **stale** per the operator's doc. |

**Reproduction of the gap (grounded):** run tellme (this repo) with any vision-capable provider and a prompt that asks it to look at a local PNG — there is **no tool** to read it and **no wire shape** to carry it, so the model cannot see the image at all. The reference can (with a vision model); tellme currently cannot. This round closes that gap for the OpenAI-compatible family.

---

## Settled design *(all decisions locked in `/axb-clarify`, one question at a time)*

| # | Decision (source) |
| --- | --- |
| **S-1** | **Entry point = an agent tool (`read_image`).** The model decides when to look at a file, exactly like the reference; no CLI `--image` flag in this round. **(Operator 2026-09-19, entry-point choice A.)** |
| **S-2** | **Family scope = the OpenAI-compatible family only (Q1 → 1).** The image is serialized as inline base64 `image_url` block(s) on the OpenAI-compatible wire — the family `deepseek` (incl. `deepseek-flash`) belongs to. The **Gemini** family's `inline_data` path is **out of scope** and is a recorded forward item (it carries the same closed-wire concerns as round 061 / ADR 0031). |
| **S-3** | **Capability = an explicit per-provider config key (Q2 → 1).** Vision is declared, not inferred: a provider entry carries `VISION: true` (bool, **default false**). tellme MUST NOT use a model-name/substring heuristic (the reference's rule is stale — the operator's doc retires `…-vision-exp`) and MUST NOT gate on the Gemini-style closed-wire analysis. |
| **S-4** | **Tool surface = `read_image` only (Q3 → 1).** No `read_video` (DeepSeek supports images only) and no `read_document` (a separate provider file-extract mechanism, not vision). Both are recorded forward items. |
| **S-5** | **Capability decides the offered surface (Q4 → 1).** With `VISION` off (**the default**), `read_image` is **not offered** to the model at all; with `VISION` on, it **is**. The offered tool set becomes a **function of the selected provider's capability** — an explicit truth change (round-029/033 territory). |
| **S-6** | **Size ceiling = 32 MiB, inline only; oversize ⇒ a loud tool error (Q5 → 1).** The wire accepts inline base64 images up to 32 MiB; the reference's Files API upload leg (32–64 MiB, `purpose=user_data`, turn-scoped cleanup) is **out of scope** and a recorded forward item. An oversize image is a **loud** tool error naming the limit — never silently truncated or sent. |
| **S-7** | **MIME is sniffed from content, never the extension (operator doc, verbatim).** JPEG/PNG/GIF/WebP are recognized from **magic bytes**; the declared extension and any client-side MIME guess are not trusted. An unrecognized/unsupported type is a loud tool error naming the detected/unknown type. |
| **S-8** | **Governance: a new ADR (next free number 0032).** It records: the explicit `VISION` capability key (and why the reference's substring rule is not copied), the OpenAI-compatible-only scope, the inline-only 32 MiB ceiling + loud oversize error, the media placement in the folded-back conversation, and §Forward (Gemini `inline_data`, the Files API leg, video, documents, the CLI `--image` flag). **No ADR is amended** (this ADR does not weaken ADR 0031; it adds an orthogonal capability). |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — **Text path byte-identical.** With no image in play, the request body is **byte-identical** to today's (round 004–061 behaviour unchanged); a text-only turn must not regress.
- **I-2** — **No security layer.** `read_image` reads whatever path it is given; there is **no** path boundary or consent gate (operator-declared direction). It is *not* a `SafePath` port.
- **I-3** — **The resource contract holds.** `read_image` honours the universal `reason` gate (round 056 / ADR 0025) and the round-024 resource contract (a tool default timeout; the loop clamp is the backstop for the **text** result).
- **I-4** — **Capability is declared, never guessed.** A provider without `VISION: true` never receives an image (the tool is hidden, S-5), so no turn can silently lose or downgrade an image (Q4).
- **I-5** — **The image is the file's own bytes.** No re-encoding, resizing, or metadata stripping in this round; the base64 payload equals the file content byte-for-byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the agent reads a local image with a vision-enabled provider (Priority: P1)

As the **operator** running tellme with a `VISION: true` provider (e.g. `deepseek-flash`) and a prompt like *"what is in ./shot.png?"*, I want the model to call `read_image` and **actually see the image**, so tellme can do the screenshot / chart / photo tasks the reference supports.

**Why this priority**: it is the operator's request and the round's whole point — the first non-text capability in tellme.

**Independent verification**: with a fake OpenAI-compatible provider capturing the request body, a tool-using turn that calls `read_image` produces a request whose image content block decodes to the file's exact bytes with the sniffed MIME type; the turn completes with an answer. A falsifiability witness removes the serialization and the check goes red.

**Acceptance Scenarios**:

1. **Given** a provider with `VISION: true` and a readable image file (PNG/JPEG/GIF/WebP), **When** the model calls `read_image` with that `filepath`, **Then** the tool succeeds and the model-visible request carries the image, and the turn produces an answer.
2. **Given** the same, **When** the serialized request body is inspected, **Then** it contains an image content block whose decoded bytes equal the file's bytes and whose MIME type is the **sniffed** type (not the extension).
3. **Given** a file whose name says `.png` but whose content is not an image, **When** `read_image` runs, **Then** it returns a loud tool error (the content is not a supported image) — the extension is not trusted.

**Functional Requirements**:

- **FR-001**: tellme MUST offer a `read_image` agent tool with a required `filepath` argument and the universal required `reason` argument (I-3), **only when** the selected provider is vision-enabled (S-5).
- **FR-002**: `read_image` MUST read the named local file and attach its content as **image media** to the model-visible request for the OpenAI-compatible family (S-1, S-2).
- **FR-003**: The MIME type MUST be resolved from the file's **content** — the magic bytes for JPEG (`FF D8 FF`), PNG (`89 50 4E 47 0D 0A 1A 0A`), GIF (`GIF87a`/`GIF89a`), and WebP (`RIFF????WEBP`) — never from the file extension or a declared MIME (S-7). An unrecognized content type MUST be a loud tool error naming the detected/unknown type.
- **FR-004**: The image MUST be serialized as an inline base64 content block (an `image_url` data-URI for the OpenAI-compatible wire) (S-2, S-6).
- **FR-005**: An image file larger than **32 MiB** MUST be a **loud** tool error naming the limit; the file MUST NOT be silently truncated, partially sent, or sent anyway (S-6).

---

### User Story 2 - the capability gate is explicit and honest (Priority: P2)

As the **operator**, I want tellme to offer `read_image` **only** when the selected provider can accept an image, so that a wrong/incomplete provider never produces a turn that silently drops or downgrades the image, and the offered tool list tells the model the truth.

**Why this priority**: it is the safety property that makes US1 trustworthy on a real fleet of providers; independent of US1 (US1 could be satisfied with the tool always offered).

**Independent verification**: with `VISION` off the offered set does **not** contain `read_image`; with `VISION` on it does — both asserted against the production assembler.

**Acceptance Scenarios**:

1. **Given** a provider with **no** `VISION` key (the default), **When** the offered tool set is built, **Then** `read_image` is **absent**.
2. **Given** a provider with `VISION: true`, **When** the offered tool set is built, **Then** `read_image` is **present** with a well-formed schema (`required ⊆ properties`, round 031).
3. **Given** a text-only turn under a `VISION: true` provider, **When** no image is read, **Then** the request body is byte-identical to today's (I-1).

**Functional Requirements**:

- **FR-006**: Provider capability MUST be an explicit per-provider config key `VISION` (boolean, **default false**); tellme MUST NOT infer vision from a model name/substring, a family, or a served-API shape (S-3).
- **FR-007**: When the selected provider's `VISION` is false, `read_image` MUST NOT be offered to the model (S-5); the offered tool set becomes a function of the selected provider's capability.
- **FR-008**: When `VISION` is true, `read_image` MUST appear in the offered set and its declaration MUST be well-formed (`required ⊆ properties`, round 031) and carry the workspace-wide resource contract.
- **FR-009**: The round-029/033 **offered-tool-set** truth (`offering-the-agent-tools.feature` + `chat/dsl.md`) MUST be MODIFY-ed to state that the set is capability-dependent and to add `read_image` under a vision-enabled provider.

---

### User Story 3 - the capability cannot silently rot, and the text path cannot regress (Priority: P3)

As the **maintainer**, I want the image serialization to be proven by a hermetic, red-capable check and the text path to stay byte-identical, so the capability cannot silently break and the round does not disturb 060 rounds of text behaviour.

**Why this priority**: recurrence/regression protection; a corollary of US1/US2.

**Independent verification**: the wire-level pin fails (red) when the image block is removed; the byte-identity check fails if the text path drifts; both ride the existing test suite.

**Acceptance Scenarios**:

1. **Given** the image serialization is deleted (witness), **When** the wire pin runs, **Then** it fails (red) — non-vacuity.
2. **Given** a text-only turn, **When** its request body is compared to the pre-round shape, **Then** they are identical (I-1).
3. **Given** an image larger than 32 MiB, **When** `read_image` runs, **Then** the loud error fires and **no** request is sent with the image.

**Functional Requirements**:

- **FR-010**: The image path MUST be covered by a **hermetic** check: a unit pin over the OpenAI-compatible request body (fake `http.RoundTripper`) AND an E2E carrier over the existing fake provider — it MUST fail (red) if the image block is not serialized (SC-002). No real network/credential is used by the gate.
- **FR-011**: A text-only turn's request body MUST remain **byte-identical** to the pre-round shape (I-1) — pinned.
- **FR-012**: An oversize image (FR-005) and an unrecognized type (FR-003) MUST fail **before** any request carrying media is sent; the tool result is a model-visible recoverable error, not a process failure (I-3).

---

## Edge Cases

- **An image exactly at 32 MiB / one byte over** — the ceiling is inclusive at 32 MiB; one byte over is the loud error (boundary pinned both ways).
- **A `.txt`/`.bin`/truncated file** — sniffing fails → loud error naming the type; no request media.
- **A 0-byte file** — sniffing fails → loud error.
- **A readable image but the provider has `VISION` off** — `read_image` is not offered, so the model cannot call it (S-5); if a stale history names it, the loop's existing "tool not available" `ErrIncomplete` path applies (unchanged).
- **An image read on a turn that also has text tool results** — media-first ordering in the folded-back conversation, so the provider does not invalidate the turn (the reference's `AssembleResponse` two-pass ordering is the precedent — the exact placement is a `/axb-technical-research` decision).
- **The `reason` gate** — a `read_image` call without a renderable reason is refused before it executes (round 056 / I-3), exactly like every tool.
- **Very large base64 expansion** — 32 MiB of file becomes ≈ 42.7 MiB of base64; the aggregate inline cap concern (reference: 48 MiB) is noted in research and, if needed, is a forward item (no aggregate bound is introduced unless research shows it is required).
- **A provider entry with `VISION: true` but a non-image-capable model** — a **declaration error** by the operator; the provider rejects it (loud provider error), by design (no guessing, I-4).
- **History replay** — a resumed session replays the stored steps; media is **not** persisted (the step carries text), so a resumed tool-using-with-image turn replays the text result only (recorded limitation; the reference is the same).

## Key Entities

- **The image media part** — the new model-visible content shape carrying an image (MIME + bytes) attached to the folded-back conversation.
- **The offered tool set** — the ordered tools tellme offers the model, now a **function of the selected provider's capability**.
- **The provider capability key (`VISION`)** — the explicit, per-provider declaration of whether an image may be sent.
- **The content sniff** — the magic-byte → MIME resolution (JPEG/PNG/GIF/WebP).
- **The image wire block** — the inline base64 `image_url` content block on the OpenAI-compatible wire.

## Success Criteria

- **SC-001**: With a vision-enabled provider, a turn that calls `read_image` **completes** and its request carries the image (decoded bytes == file bytes; sniffed MIME). *(Hermetic: fake provider capture. A live spot-check against a real vision endpoint is a non-gating closeout step, mirroring round 061/031's manual live check.)*
- **SC-002**: A hermetic check proves the image reaches the wire; **removing the serialization turns it red** (falsifiability witness).
- **SC-003**: A text-only turn is **byte-identical** to the pre-round shape (I-1, pinned).
- **SC-004**: With `VISION` off, `read_image` is **not offered**; with it on, it **is** — both asserted against the production assembler (FR-007).
- **SC-005**: `make verify` + `go test -count=1 ./...` green (including the E2E contract); the topology/DSL audit adds **no** new findings; any dependency change is recorded (`go.mod`/`go.sum`).
- **SC-006**: Every changed behaviour (the wire block, the offered-set gate, the ceiling, the sniff) has a reproduced falsifiability witness.

## Assumptions

- **A1**: The reference is the **parity precedent** for the shape (`read_image` tool, `BinaryData` → media part, `image_url` blocks, media-first ordering), adapted to tellme's architecture and the operator's doc.
- **A2**: The **placement** of the image in the folded-back conversation (media-first content array on the tool message vs a following user message) is decided in `/axb-technical-research`, informed by the reference (`executor.AssembleResponse` → role `user`, media first) and DeepSeek's *"images in user messages only"* constraint. The spec requires only that the request be accepted and the model see the image (FR-002).
- **A3**: This is a **plain line CLI** round — `/axb-ui-plan` is skipped; `/axb-api-plan` is NOOP (no HTTP surface); `/axb-data-plan` is expected **NOOP or a small MODIFY** (the in-memory message-part model).
- **A4**: Truth impact is expected in `specs/truth/techstack.md` (a vision/image row + the config-key row), `specs/truth/data/data-model.dbml` (the in-memory conversation model gains a media part), and `specs/truth/features/cli/chat/**` (the offered-set feature + `dsl.md` + a new image journey).
- **A5**: A **new ADR (next free number: 0032)** records the capability key, the scope, and §Forward (S-8); **no ADR is amended**.
- **A6**: The round is testable **hermetically** (the existing fake provider + fake HTTP), so no live endpoint or credential is required by any gate. The only non-hermetic step is the optional closeout live spot-check.
- **A7**: Capability is **declared per provider** (S-3); the round does not add a capability-inference engine.

## Out of scope (recorded forward items)

- **The Gemini family's `inline_data` image path** (Q1 → 1) — a recorded forward item; it carries the closed-wire projection concerns of round 061 / ADR 0031.
- **Video (`read_video`) and documents (`read_document`)** (Q3 → 1) — video is unsupported by DeepSeek; documents are a separate extract mechanism.
- **The Files API upload leg (32–64 MiB, `purpose=user_data`, turn-scoped cleanup)** (Q5 → 1) — inline-only this round.
- **A CLI `--image <path>` attachment on the prompt** (entry-point choice) — the agent tool is the only entry point this round.
- **Image generation** (`create_image`) and **re-encoding/resizing/metadata handling** (I-5).
- **Persisting media in history / replaying an image across sessions** — the step carries text (a recorded limitation).
- **A capability-inference engine** (model-name or served-API probing) — capability is declared (S-3).
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion) · **parallel tool calls** ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
