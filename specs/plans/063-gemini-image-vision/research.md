# Research — Gemini image vision: `inline_data` on the Gemini/Vertex wire (round 063)

**Plan Package**: `specs/plans/063-gemini-image-vision`
**Anchor issue**: none (operator request — lands the recorded forward item **RF-062-1** from ADR 0032 §Forward)
**Truth owner**: `/axb-technical-research` — updates `specs/truth/techstack.md`

## Problem (grounded)

Round 062 gave tellme its first non-text capability on the **OpenAI-compatible** family and made the **Gemini** family a **loud refusal** (ADR 0032 D4/D8): `internal/infrastructure/llm/gemini/client.go`'s `Complete` opens with `if hasMedia(req.Messages) { return … "this provider family cannot carry images yet; remove VISION: true or use an OpenAI-compatible provider" }`. Everything else the capability needs already exists and is **family-agnostic**:

- the domain carrier `llm.Message.Media []llm.MediaPart{MIMEType, Data}` (`internal/domain/llm/gateway.go`);
- the per-call media channel `llm.WithMediaCollector`/`llm.AttachMedia` (`internal/domain/llm/media.go`, ADR 0032 D7a);
- the loop's placement — a call that attached media appends **one** `llm.Message{Role:"user", Media: media}` **after** the round's `tool` result (`internal/agent/agentloop.go:149-176`);
- the `read_image` tool (content sniff + ceiling + `AttachMedia`, `internal/infrastructure/tools/image.go`);
- the offered-set gate — `read_image` is offered **iff the selected provider declares `VISION`** (`cmd/tellme`'s `assembleAgentTools(sink, vision)`; `internal/cli/cli.go:745`).

So the **only** missing piece is the Gemini adapter's **serialization**: the reference maps `InlineData → toSDKBlob` → an `inline_data` blob and, crucially, reorders a user turn to `[InlineData][FunctionResponse][other]` (`normalizeUserTurnParts`, #1441) because the Vertex parser invalidates a user turn whose `functionResponse` **precedes** an `inlineData` part. Its `SupportsVision` flag gates **only** the OpenAI transport; the Gemini path carries the blob ungated.

## Decisions taken upstream (`/axb-clarify`, one at a time → `spec.md` S-1…S-8)

| # | Decision |
| --- | --- |
| Q1 → A | **Reuse the single, family-agnostic `VISION` key** — no second key; the offered-set gate is unchanged; `VISION: true` on a `gemini`/`vertex` provider enables the image path. |
| Q2 → B | **A family-aware inline ceiling**, enforced by **`read_image`** (it takes the selected provider's resolved ceiling) so oversize is a loud **tool** error before the wire, on **both** families. |
| Q3 (not asked) | The `inlineData` **placement** in the Vertex `contents` — decided here (D2). |

## Decisions (this phase)

### D1 — Scope: serialize media on the Gemini wire; retire the refusal

- **Decision**: the Gemini adapter **serializes** a media-bearing message (D2/D3); the round-062 loud refusal and its helper (`hasMedia`, `client_image_test.go`) are **retired**. No new tool, no new capability key, no domain change — the round is **adapter-side** plus truth corrections.
- **Rationale**: the capability plumbing is already family-agnostic (grounded above); only the transport was missing. "never silently lose an image" (ADR 0032 D8) is now satisfied **by carrying** the image, not by refusing.
- **Alternatives considered**: (a) a second capability key — rejected (Q1 → A); (b) keep the refusal and add a `--gemini-images` opt-in — rejected (the operator asked for parity, and the refusal exists only because there was no path).

### D2 — Placement: the media message becomes its **own `user` contents entry**, after the tool-result turn

- **Decision**: in `buildContents`, a message with `len(m.Media) > 0` maps to **one `contents` entry** `{"role":"user","parts":[<text part when the message has text>, <one inlineData part per media part>]}` — i.e. the round-062 trailing `user` message (agentloop.go) becomes its own Vertex `user` turn, emitted **after** the `functionResponse` `user` turn of the tool result. Media lead the turn's parts.
- **Rationale**: the loop **already** appends a separate trailing `user` message (ADR 0032 D7), so serializing it as its own turn is the **minimal** change and needs no cross-message correlation. It **avoids by construction** the in-turn hazard the reference heals (`functionResponse` before `inlineData` — #1441): the inline data never shares a turn with a `functionResponse`. A `user` turn carrying only `inlineData` (and optionally text) is a valid Vertex `Part` set; consecutive `user` turns are accepted (Vertex `contents` is a list, roles may repeat).
- **Alternatives considered**: (a) **merge** the inline data into the tool result's `functionResponse` turn, media-first — the reference's **normalized** shape — rejected: tellme's loop does not co-locate the media with the tool result, so merging would require the adapter to correlate a media message with the preceding `functionResponse` turn (stateful, error-prone) for no observable gain; (b) attach the image to the **next** request's prompt — rejected (the model may answer instead of calling again).
- **Residual (recorded)**: a **live** Vertex confirmation of the standalone-`user`-turn ordering is the round's one non-gating closeout check (the round-059/061 precedent — a real-endpoint behaviour a hermetic fake cannot prove).

### D3 — The wire shape: an `inlineData` proto-JSON blob (`mimeType` + base64 `data`)

- **Decision**: each media part becomes `{"inlineData":{"mimeType":<sniffed>,"data":"<base64-StdEncoding>"}}`. The literal keys are **camelCase** (`inlineData`/`mimeType`) — Vertex REST accepts proto-JSON, and tellme's adapter **already** uses camelCase for every other key it emits (`functionCall`, `thoughtSignature`, `maxOutputTokens`, `systemInstruction`, `generationConfig`), matching the SDK's json tags. The base64 is `base64.StdEncoding` (round 062's choice; Vertex's `inlineData.data` is standard base64 in proto-JSON).
- **Rationale**: consistency with the adapter's existing wire spellings; the SDK-tag spelling is `inlineData`/`mimeType` (snake `inline_data`/`mime_type` is the alternate REST form — the spec's `inline_data` is the **concept** name; this decision pins the literal).
- **Alternatives considered**: (a) snake_case `inline_data`/`mime_type` — rejected for key-spelling consistency within the adapter; the API accepts both, so consistency wins.

### D4 — The family-aware inline ceiling (Q2 → B)

- **Decision**: `read_image` gains a resolved **image byte ceiling** input (measured in **file bytes**, the unit the tool already uses). A single-owned lookup maps the family to its ceiling:
  - **OpenAI-compatible: 32 MiB** — **unchanged** (round 062 / the DeepSeek doc's inline limit).
  - **Gemini/Vertex: 14 MiB** (a conservative, derived value — see below) — enforced **before** the wire as a loud tool error naming the limit.
  The composition root resolves the family and injects the value (the same seam that already carries the vision flag).
- **Derivation + source**: Vertex/Gemini documents an inline-data **request-size** bound of ≈**20 MB**. Because the tool measures **raw** file bytes while the wire payload is **base64** (≈ 4/3 of raw), a raw ceiling of **14 MiB** yields an encoded body of ≈ 18.7 MiB ≈ 19.6 MB — under the documented bound, so a file the tool accepts cannot be rejected purely for size. This is deliberately **conservative** (a loud refusal beats a provider 400).
- **Rationale**: honest and single-owned; keeps round 062's "oversize ⇒ a loud **tool** error naming the limit, never sent" property on **both** families (Q2 → B). The value lives in **one** place, so it moves trivially if the documented figure/unit is corrected.
- **Alternatives considered**: (a) one global 32 MiB ceiling — **the operator rejected it** (Q2 → A: it would make oversize a *provider* error on Gemini, the weaker diagnostic round 062 deliberately avoided); (b) measuring the bound with a live probe — **not available** hermetically in this environment (no credential; the E2E suite is network-guarded); (c) applying the documented bound to the **encoded** length — deferred (would need a second unit in the tool seam; the raw-bytes conservative derivation needs none).
- **Residual (recorded)**: the exact documented bound and its **unit** (total request vs inline bytes; decimal MB vs MiB) must be **confirmed** at the closeout live check; if it differs, the single constant moves and the acceptance fixture's "larger than the Gemini image size limit" (expressed **relative** to the limit — no literal) follows automatically. Recorded in ADR 0033 §Forward.

### D5 — Capability stays `VISION`; the offered surface is unchanged (Q1 → A)

- **Decision**: no config change. `VISION` remains the single, family-agnostic declaration; `read_image` is offered **iff** `VISION` is true; the round-062 offered-set truth (`offering-the-agent-tools.feature` + the `dsl.md` rows) is **unchanged**.
- **Rationale**: Q1 → A; the key was already family-agnostic (the refusal, not the gate, was family-specific).
- **Alternatives considered**: a separate Gemini key — rejected (Q1).

### D6 — One content sniff, shared by both families (no duplication)

- **Decision**: the Gemini path reuses the **existing** `read_image` tool and its magic-byte sniff verbatim — the MIME/staleness/boundary logic has **one owner**; the adapter only base64-encodes the bytes the tool already produced.
- **Rationale**: S-3 / I-5; a second sniff would be a drift source, and the reference likewise sniffs once (its `read_image` precedes both transports).
- **Alternatives considered**: an adapter-side sniff — rejected (duplicate authority).

### D7 — The tool's ceiling seam (the natural continuation of RF-062-10)

- **Decision**: widen the assemblage path minimally — `NewReadImageTool(maxBytes int)` (and the assembler's input) carries the resolved ceiling; no struct/seam refactor is required this round. The already-recorded **RF-062-10** (`ToolSetSpec` replacing the bare `vision bool`) is the **next** refactor candidate; this round adds one more scalar alongside it rather than pre-empting the refactor.
- **Rationale**: smallest change; keeps the round adapter-side + one injected scalar; the layer gate stays green.
- **Alternatives considered**: introduce `ToolSetSpec` **now** — deferred (RF-062-10; not required by the round, and a bigger blast radius in the composition root).

### D8 — Verification: a hermetic Gemini wire pin + the retired refusal pin; the E2E carrier

- **Decision**: (a) a **unit pin** over the Vertex request body (fake `http.RoundTripper` capturing the JSON) asserts the `inlineData` part is present, its `data` **decodes to the file's exact bytes**, and its `mimeType` is the **sniffed** type; (b) the round-062 `client_image_test.go` **inverts** from "media ⇒ refusal" to "media ⇒ the blob"; (c) an **E2E carrier** (the existing Vertex-shaped fake provider) drives `read_image` end to end on a `gemini` provider; (d) **witnesses**: removing the serialization reds the unit pin (non-vacuity), and lowering/raising the injected ceiling reds the boundary Example. No new `make verify` member.
- **Rationale**: the transform is pure ⇒ a unit pin is a complete proof; the E2E proves the wiring; round 062's shape (D11) is reused.
- **Alternatives considered**: E2E-only — rejected (a unit pin is the sharper red-capable carrier).

### D9 — Governance: **ADR 0033** (extends ADR 0032; supersedes its RF-062-1)

- **Decision**: **ADR 0033** records the Gemini `inlineData` serialization, the standalone-`user`-turn placement (and why it differs from the reference's merge), the camelCase key spelling, and the family-aware ceiling with its **derivation + residual**. It **supersedes ADR 0032's forward item RF-062-1** and **narrows ADR 0032 D4/D8** (the family-asymmetry/loud-refusal clause) — recorded in ADR 0033 and by **annotating ADR 0032's Status/§Forward** (round 058's ADR-0027-annotation precedent). No assumption of ADR 0032 is otherwise edited.
- **Rationale**: the repo's standing rule (a durable decision is an ADR; the family scope changed); ADR 0032 stays an `Accepted` historical record.
- **Alternatives considered**: (a) amend ADR 0032 in place — rejected (an `Accepted` ADR is immutable; the 0027/0028 precedent annotates the superseded clause and adds a new ADR); (b) no ADR — rejected (a cross-cutting wire decision).

### D10 — The Gemini text path stays byte-identical (I-1)

- **Decision**: a media-free Gemini request serializes **exactly** as today; `buildContents`' media branch is additive. A text-only control asserts byte-identity.
- **Rationale**: I-1; 060 rounds of text behaviour must not regress.
- **Alternatives considered**: none (a hard requirement).

### D11 — No new dependency; stdlib-only; hermetic

- **Decision**: `encoding/base64` + `encoding/json` only (both already imported); POSIX-only; no new `go.mod` entry; every gate stays hermetic.
- **Rationale**: round 062's precedent; SC-005.

## Truth impact (owner summary → `truth-delta.md`)

| Owner | Spec | Action |
| --- | --- | --- |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Image content on the provider wire* | **MODIFY** — the Gemini `inlineData` blob; the standalone-`user`-turn placement; the retired Gemini refusal; the family-aware ceiling. |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Image filesystem tool (`read_image`)* | **MODIFY** — the injected, family-aware ceiling. |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | **MODIFY** — the adapter now carries `inlineData` media (and no longer refuses). |
| `/axb-technical-research` | `specs/truth/techstack.md` — *Provider entry schema* | **MODIFY** — the `VISION` note: honoured by **both** families. |
| `/axb-technical-research` | `docs/decisions/0033-*.md` (+ index; ADR 0032 annotation) | **ADD** — the decision + the supersession of RF-062-1. |
| `/axb-api-plan` | `specs/truth/contracts/**` | **NOOP** (no HTTP surface). |
| `/axb-data-plan` | `specs/truth/data/data-model.dbml` | **NOOP** (no persisted shape change; the image is in-flight). |
| `/axb-dsl-refine` | `chat/reading-a-local-image.feature` + `chat/dsl.md` | **MODIFY** — the Gemini journey + the retired-refusal/loud-family-ceiling sentences. |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): to be refreshed for the round-063 capability fact (the `ImageContent`/`Provider.vision` wording now says the image is carried by **both** families), landed in the implementation phase; `make modelith-check` must stay green.

## Residual risks (non-blocking — forwarded)

- **The Gemini ceiling figure** — derived (conservative) from a documented ≈20 MB inline request bound; **not measured** here (no credential / hermetic gates). Confirm at the closeout live check; one owner ⇒ cheap to move (D4/ADR 0033 §Forward).
- **The standalone-`user`-turn ordering** — argued valid + hazard-free (D2) but **unproven live**; the closeout live check is its witness.
- **No aggregate bound** for multiple images in one turn (round-062 RF-062-3/RF-062-8) — unchanged, and the Gemini side inherits it.
- **The Files-API upload leg** (round-062 RF-062-2) — still deferred; inline-only on both families.
- **`read_image` reads whatever path it is given** (no security layer — a settled exclusion).

## Why this is the right shape for tellme

The capability plumbing (carrier, channel, placement, tool, gate) is already family-agnostic; only the transport was missing. One adapter branch (a `user` turn with `inlineData` parts), one retired refusal, one injected family-aware ceiling, and the truth corrections buy **full parity** on the operator's Vertex case with no new tool, no new key, no new dependency, and a byte-identical text path on both families. The two genuinely-larger pieces (the Files-API upload leg and an aggregate multi-image bound) stay named, durable forward items rather than silent gaps.
