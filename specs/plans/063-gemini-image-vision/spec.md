# Feature Specification: Gemini image vision — `inline_data` on the Gemini/Vertex wire (round 063)

**Feature Branch**: `063-gemini-image-vision`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from an **operator request** (no anchor issue). **Clarify IN PROGRESS** — **Q1 → A (reuse the single `VISION` key) LOCKED** (operator, 2026-09-19); **Q2 (the Gemini inline size ceiling)** is the remaining high-impact gap, asked one at a time (see *Clarify* below). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19, this session)**:

> *"Refer to tell-me-go, does vertex gemini have vision (read_image)?"* → (grounded answer: **yes** — the reference's Gemini adapter carries images natively as `inline_data` blobs, ungated) → *"We need to let tellme gemini to have read_image too."*

**Anchor issue**: none (operator request). **This round lands the recorded forward item `RF-062-1`** — *"the Gemini family's `inline_data` image path"* — from [`docs/decisions/0032-agent-image-vision.md`](../../../docs/decisions/0032-agent-image-vision.md) §Forward. It closes the family asymmetry ADR 0032 D4 accepted deliberately ("the operator guidance is to set `VISION: true` only for an openai-family provider until RF-062-1 lands").

**Behaviour intent**: **ADD (a second family's image serialization: `inline_data` on the Gemini/Vertex wire) + MODIFY (retire the round-062 loud Gemini media refusal).** Round 062 gave tellme its first non-text capability on the **OpenAI-compatible** family and made the **Gemini** family a **loud refusal** (no silent loss). This round makes the **same `read_image` capability** work on the **Gemini** family — the operator's live Vertex case — while keeping the text path byte-identical on **both** families.

---

## Grounded in the current system *(measured 2026-09-19, `dev` @ `cff2515`)*

| Site | Current shape |
| --- | --- |
| `internal/domain/llm/gateway.go` | `Message{Role, Content string, Media []MediaPart, ToolCalls, ToolCallID}` — **the round-062 media field already exists**; `MediaPart{MIMEType string; Data []byte}`; no further domain change is expected. |
| `internal/domain/llm/media.go` | The per-call media **channel**: `WithMediaCollector(ctx, *[]MediaPart)` + `AttachMedia(ctx, …)` (ADR 0032 D7a). |
| `internal/agent/agentloop.go:149-176` | Installs a FRESH collector per tool call; a call that attached media appends **ONE `user` message** (`llm.Message{Role:"user", Media: media}`, media-first) immediately **after** the round's `tool`-result message. |
| `internal/infrastructure/tools/image.go` | `read_image`: magic-byte sniff (JPEG/PNG/GIF/WebP), **32 MiB** inline ceiling with a loud oversize error, `llm.AttachMedia`. |
| `cmd/tellme/deps.go` (`assembleAgentTools(sink, vision)`) + `internal/cli/cli.go:745` | `read_image` is appended **iff** `res.Provider.Vision` — **the offered set is already a function of capability** (round 062 FR-007). |
| `internal/config/config.go` (`Provider.Vision bool yaml:"VISION"`) | The capability is **family-agnostic** today; nothing binds `VISION` to the OpenAI-compatible family *except* the Gemini adapter's refusal. |
| `internal/infrastructure/llm/gemini/client.go` | `buildContents` emits **text-only** parts, plus `functionCall` (role `model`) / `functionResponse` (role `user`); `Complete` opens with `if hasMedia(req.Messages) { return … c.wrap(fmt.Errorf("this provider family cannot carry images yet; remove VISION: true or use an OpenAI-compatible provider")) }` — **the loud refusal this round retires**. |
| `internal/infrastructure/llm/openai/client.go` (`messageContent`/`dataURI`) | The round-062 OpenAI wire shape: a content **array** with a leading text part + one inline base64 `image_url` data-URI per media part. |
| `specs/truth/techstack.md` (*Image content on the provider wire*) + `chat/reading-a-local-image.feature` + `chat/dsl.md` | The round-062 truth: the OpenAI image block; the offered-set capability gate; the loud refusals; **the Gemini refusal is asserted** (an E2E Rule + a unit pin `internal/infrastructure/llm/gemini/client_image_test.go`). |
| `tell-me-go` (reference) | The Gemini adapter maps `llm.Part.InlineData → toSDKBlob(p.InlineData)` → `genai.Blob{MIMEType, Data}`, serialized as the Vertex/Gemini **`inline_data`** blob (`mime_type` + base64 `data`); `normalizeUserTurnParts` reorders a **user** turn to `[InlineData][FunctionResponse][other]` **specifically** so Vertex does not 400 *"Requests ending with a model turn are not supported"* (#1441) — proof images are carried in production. `SupportsVision` (a capability flag) gates **only** the OpenAI transport; the Gemini path is **ungated** and carries the blob. |

**Reproduction of the gap (grounded):** run tellme with a Gemini/Vertex provider declaring `VISION: true` and a prompt asking it to look at a local PNG — `read_image` **is offered** (capability is family-agnostic) but the turn **fails loudly** at the adapter (*"this provider family cannot carry images yet…"*). The reference carries the same image on the Gemini wire; tellme cannot. This round closes that gap.

---

## Design (S-1 locked by Q1 → A; the rest proposed by grounding — Q2 pending the operator's confirmation)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **The capability key is unchanged: `VISION: true` (per provider), and it now works on the Gemini family.** The offered-set gate (round 062 D3/FR-007) needs **no change** — it is already family-agnostic; only the adapter's serialization + the refusal change. | **locked (Q1 → A)** |
| **S-2** | **The Gemini image is an `inline_data` blob** — a `{"inline_data":{"mime_type":<sniffed>,"data":<base64>}}` part on a **`user`** turn (the Vertex/Gemini REST shape; the reference's `genai.Blob` shape). | proposed |
| **S-3** | **The content sniff is shared, not duplicated** — the same magic-byte table (JPEG/PNG/GIF/WebP) and the same `read_image` tool serve both families; the kind resolution has **one owner**. | proposed |
| **S-4** | **Media placement keeps the functionCall→functionResponse pairing intact** (the Gemini parser rejects an unanswered `functionCall`; the reference merges `InlineData` **before** the `functionResponse` in the same `user` turn — #1441). The exact wire placement is a `/axb-technical-research` decision informed by the reference; the spec requires only that the request be **accepted** and the model **see** the image. | proposed |
| **S-5** | **The round-062 loud Gemini refusal is retired** (it existed because there was no image path; D8's "never silently lose an image" is preserved because the image now **is** carried). | proposed |
| **S-6** | **A family-appropriate inline ceiling with a loud oversize error** — the Vertex inline request has a documented bound; whether tellme keeps the 32 MiB ceiling for symmetry or adds a Gemini-specific one is **Q2**. | *pending Q2* |
| **S-7** | **The text path stays byte-identical on both families** — a Gemini request with no media is unchanged. | proposed |
| **S-8** | **Governance: extend ADR 0032 (or a new ADR?)** — RF-062-1 landing is a *decision* about the Gemini wire; the repo's pattern (round 057→058, 038→039) is a **new ADR** that extends the prior one and records the superseded forward item. The exact choice is a `/axb-technical-research` call. | proposed |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — **Text path byte-identical** on **both** families (a media-free Gemini request is byte-for-byte today's).
- **I-2** — **No security layer** (settled exclusion): `read_image` reads whatever path it is given.
- **I-3** — **The resource contract holds**: the universal `reason` gate (round 056) and the round-024 resource contract are unchanged.
- **I-4** — **Declared capability, no silent loss**: the image is either carried or the turn **fails loudly**; a declared-`VISION` provider never drops the image silently on either family.
- **I-5** — **The image is the file's own bytes** — no re-encoding/resizing; the base64 payload equals the file content byte-for-byte (the same sniffer + the same `read_image` tool produce the bytes).

---

## Clarify (one question at a time)

> Per `/axb-clarify`, only **high-impact** gaps are asked. **Q1 is answered**; **Q2 is open**; the third is left to research.

| # | Question | Why it is high-impact | Recommendation |
| --- | --- | --- | --- |
| **Q1** ✅ **ANSWERED (A)** | Is the Gemini image capability declared by the **same** `VISION` key (so `VISION: true` alone enables the Gemini image path), or by a **separate** Gemini key? | It decides whether the config surface grows and whether the round-062 offered-set semantics change. | **Operator chose (A) — reuse `VISION`** (2026-09-19). The key stays family-agnostic; `VISION: true` on a `gemini`/`vertex` provider now enables the image path; **no second key**; the offered-set gate is unchanged. Recorded as **S-1 (locked)** / **FR-007**. |
| **Q2** ⏳ **OPEN** | What **inline size ceiling** governs the Gemini path — one **shared** ceiling (**32 MiB**, the round-062/OpenAI number, kept global), or a **family-aware** ceiling (the Vertex inline request bound, enforced by `read_image` per selected provider so the oversize case is a loud **tool** error before the wire)? | It changes a numbered acceptance criterion, the oversize-refusal behaviour, and whether the tool's ceiling becomes capability-aware. | **Recommend (B) family-aware** with the value sourced/measured in `/axb-technical-research` (the honest bound, mirroring round 062's 32 MiB rationale); **(A)** keeps one global ceiling and lets the provider reject an oversize image (a loud *provider* error, not a loud *tool* error). |
| *(Q3 — not asked)* | The exact **wire placement** of `inline_data` relative to the tool result (merge into the functionResponse `user` turn — the reference's media-first shape — vs the round-062 separate trailing `user` message). | Technical shape; the spec requires acceptance + visibility only. | Left to `/axb-technical-research` (grounded in the reference's `[InlineData][FunctionResponse][other]` normalization, #1441). |

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the agent reads a local image with a Gemini/Vertex provider (Priority: P1)

As the **operator** running tellme with a `gemini`/`vertex` provider declaring `VISION: true`, I want the model to call `read_image` and **actually see the image**, so the screenshot/chart/photo tasks the reference supports work on my Vertex Gemini provider too.

**Why this priority**: it is the operator's request and the round's whole point.

**Independent verification**: with a fake Vertex-shaped provider capturing the request body, a tool-using turn that calls `read_image` produces a request whose content carries an `inline_data` blob whose base64 `data` decodes to the file's **exact** bytes with the **sniffed** `mime_type`; the turn completes with an answer. A falsifiability witness removes the serialization and the check goes red.

**Acceptance Scenarios**:

1. **Given** a Gemini/Vertex provider with `VISION: true` and a readable image (PNG/JPEG/GIF/WebP), **When** the model calls `read_image` with that `filepath`, **Then** the tool succeeds, the model-visible request carries the image as an `inline_data` blob, and the turn produces an answer.
2. **Given** the same, **When** the serialized request body is inspected, **Then** the blob's decoded bytes equal the file's bytes and its `mime_type` is the **sniffed** type (not the extension).
3. **Given** a media-free turn on the same provider, **When** the request body is compared to the pre-round shape, **Then** it is **byte-identical** (I-1).

**Functional Requirements**:

- **FR-001**: The Gemini/Vertex adapter MUST serialize a media-bearing message as an `inline_data` content part (`{"inline_data":{"mime_type":<sniffed>,"data":<base64>}}`) on a `user` turn (S-2).
- **FR-002**: The serialization MUST keep the conversation valid for the Vertex parser — in particular the `functionCall` → `functionResponse` pairing must not be broken by the image part (the reference's media-first user-turn shape, #1441) (S-4).
- **FR-003**: The MIME type MUST come from the **shared** content sniff (S-3) — the same magic-byte table the OpenAI path uses; the extension/declared MIME is never trusted.
- **FR-004**: The Gemini path MUST enforce its inline **size ceiling** with a **loud** tool error (never truncated/partial); the ceiling value is Q2.
- **FR-005**: A provider declaring `VISION: true` on the Gemini family MUST produce a **working** image turn (the round-062 loud refusal is retired) (S-5).
- **FR-006**: A media-free Gemini request MUST remain **byte-identical** to the pre-round shape (I-1).

### User Story 2 - one capability, two families, no new config surface (Priority: P2)

As the **operator**, I want the **same** `VISION` declaration and the **same** `read_image` tool to work on both families, so I do not carry a second capability spelling and cannot mis-declare one family for the other.

**Why this priority**: keeps the config and offered-set semantics coherent (S-1); independent of US1 (which is adapter-side).

**Independent verification**: with `VISION: true` and a Gemini provider, `read_image` is offered (unchanged from round 062) **and** the image is carried; with `VISION` off it is not offered (unchanged). Asserted against the production assembler and the adapter request body.

**Acceptance Scenarios**:

1. **Given** a Gemini provider with `VISION: true`, **When** the offered tool set is built, **Then** `read_image` is present (the round-062 gate is unchanged) **and** the image turn carries the blob.
2. **Given** a provider with no `VISION` key, **When** the offered set is built, **Then** `read_image` is absent (unchanged).
3. **Given** a `VISION: true` provider of **either** family, **When** the turn runs, **Then** no "cannot carry images" refusal is produced (US1) — the two families are at parity.

**Functional Requirements**:

- **FR-007**: The capability key MUST remain the single **`VISION`** boolean (no new key) (S-1, **Q1 → A locked**).
- **FR-008**: The offered-set gate (round 062 FR-007) MUST be **unchanged** — capability, not family, decides (S-1).
- **FR-009**: The round-062 Gemini **refusal** text/path MUST be retired, and the owning truth (techstack row / feature / `dsl.md` / the `client_image_test.go` pin) MODIFY-ed to the new behaviour.

### User Story 3 - the Gemini image path cannot silently rot and cannot regress the text path (Priority: P3)

As the **maintainer**, I want the Gemini image serialization proven by a hermetic, red-capable check and the Gemini text path kept byte-identical.

**Independent verification**: the wire-level pin fails (red) when the `inline_data` part is removed; the byte-identity check fails if the text path drifts.

**Acceptance Scenarios**:

1. **Given** the `inline_data` serialization is deleted (witness), **When** the gemini wire pin runs, **Then** it fails (red) — non-vacuity.
2. **Given** a media-free Gemini turn, **When** its request body is compared to the pre-round shape, **Then** they are identical (I-1).
3. **Given** an image past the ceiling, **When** `read_image` runs, **Then** the loud error fires and **no** request carrying the image is sent.

**Functional Requirements**:

- **FR-010**: The Gemini image path MUST be covered by a **hermetic** unit pin over the Vertex request body (fake `http.RoundTripper`) **and** an E2E carrier (a Vertex-shaped fake provider) — it MUST fail red if the blob is not serialized (SC-002). No real network/credential in the gate.
- **FR-011**: A media-free Gemini request MUST remain byte-identical (I-1) — pinned.
- **FR-012**: The oversize (FR-004) path MUST fail **before** any request carrying media is sent; the tool result is a model-visible recoverable error (I-3).

---

## Edge Cases

- **An image exactly at the ceiling / one byte over** — the ceiling is inclusive; one over is the loud error (boundary pinned both ways).
- **A `.txt`/truncated/0-byte file** — the shared sniff fails → the round-062 loud error (unchanged; the sniffer is shared, S-3).
- **A media-bearing message with the provider `VISION` off** — `read_image` is not offered (unchanged); a stale history naming it hits the existing "tool not available" path.
- **An image turn that also has text tool results** — the `functionCall`→`functionResponse` pairing must survive (FR-002); the exact ordering is the research decision (Q3).
- **`reason` gate** — a `read_image` call without a renderable reason is refused before it executes (unchanged, I-3).
- **Base64 expansion** — the request body grows ≈ 4/3; the Vertex inline bound (Q2) governs.
- **The `thoughtSignature` echo** — the Gemini functionCall replay carries `thoughtSignature`; the image part must not disturb it.
- **`VISION: true` on a Gemini provider with a text-only model** — an operator declaration error; the provider rejects it loudly (no guessing, I-4 — unchanged).
- **History replay** — media is not persisted (the step carries text; round-062 limitation, unchanged).

## Key Entities

- **The `inline_data` image part** — the new Gemini-wire content shape (sniffed MIME + base64 bytes).
- **The offered tool set** — unchanged: a function of the selected provider's `VISION` (one capability, two families).
- **The shared content sniff** — the magic-byte → MIME resolution used by both families.
- **The provider capability key (`VISION`)** — unchanged; now honoured by both families.

## Success Criteria

- **SC-001**: With a Gemini provider declaring `VISION: true`, a turn that calls `read_image` **completes** and its request carries the image as an `inline_data` blob (decoded bytes == file bytes; sniffed `mime_type`). *(Hermetic: fake Vertex-shaped provider. A live spot-check against a real Vertex endpoint is a non-gating closeout step, mirroring rounds 061/062.)*
- **SC-002**: A hermetic check proves the image reaches the Gemini wire; **removing the serialization turns it red**.
- **SC-003**: A media-free Gemini turn is **byte-identical** to the pre-round shape (I-1), pinned.
- **SC-004**: The offered set is unchanged (capability-driven); no "cannot carry images" refusal remains for a declared-`VISION` Gemini provider.
- **SC-005**: `make verify` + `go test -count=1 ./...` green (including the E2E contract); the topology/DSL audit adds **no** new findings; any dependency change is recorded.
- **SC-006**: Every changed behaviour (the blob, the retired refusal, the ceiling) has a reproduced falsifiability witness.

## Assumptions

- **A1**: The reference is the **parity precedent** for the Gemini shape (`InlineData` → blob → `inline_data`; media-first `user` turn; `role:user` for the functionResponse).
- **A2**: The **placement** (merge into the functionResponse `user` turn vs the round-062 separate trailing `user` message) is decided in `/axb-technical-research`, informed by the reference's `normalizeUserTurnParts` (`[InlineData][FunctionResponse][other]`, #1441). The spec requires only acceptance + visibility.
- **A3**: Plain line CLI round — `/axb-ui-plan` skipped; `/axb-api-plan` NOOP; `/axb-data-plan` expected **NOOP** (no persisted shape change — the round-062 `Message.Media` is in-memory).
- **A4**: Truth impact expected in `specs/truth/techstack.md` (*Image content on the provider wire* — the Gemini blob + the retired refusal; a `VISION`-honoured-by-both-families note), `specs/truth/features/cli/chat/reading-a-local-image.feature` (the Gemini journey / the retired refusal Rule), `chat/dsl.md`, and a **new ADR** extending ADR 0032 (S-8).
- **A5**: The round is testable **hermetically** (the existing fake provider + fake HTTP); the only non-hermetic step is the optional closeout live spot-check.
- **A6**: No new tool and no new tool-surface truth — `read_image` already exists and is capability-gated (round 062); this round is **adapter-side** plus the truth corrections.

## Out of scope (recorded forward items)

- **The Files-API upload leg (32–64 MiB, turn-scoped)** (round-062 RF-062-2) — still deferred; this round is inline-only for both families.
- **An aggregate cap for multiple images in one turn** (RF-062-3/RF-062-8) — unchanged.
- **`read_video` / `read_document`** (RF-062-4); **a CLI `--image` flag** (RF-062-5); **image persistence/replay** (RF-062-6); **a capability-inference engine** (RF-062-7); **the media-channel refactor** (RF-062-10); **the shared E2E enumerator** (RF-062-12).
- **The Microsoft/dimension guards** the reference applies on the OpenAI path (`media_dimensions.go`) — not adopted this round.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion) · **parallel tool calls** ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
