# Research — Agent image vision: `read_image` to the provider wire (round 062)

**Plan Package**: `specs/plans/062-agent-image-vision`
**Anchor issue**: none (operator request; supersedes the reference's stale DeepSeek vision assumption **in tellme**)
**Truth owner**: `/axb-technical-research` — updates `specs/truth/techstack.md`

## Problem (grounded)

tellme is **text-only end to end** (measured on `dev` @ `902642a`): `internal/domain/llm/gateway.go`'s `Message.Content` is a `string`; `internal/infrastructure/llm/openai/client.go` emits `{"role":…,"content":<string>}`; `internal/infrastructure/llm/gemini/client.go` emits `{text: …}` parts; `internal/domain/tools.Tool.Execute` returns a `string`; and no media tool is registered. So there is **no path** by which an image can enter a turn — the model cannot see a picture the operator already has on disk.

The reference (`tell-me-go`) *can*, but its capability gate is a **model-ID substring** (`strings.Contains(model,"vision")`) and its **ADR-070** assumed `deepseek-v4-flash` was text-only. The operator's official DeepSeek doc inverts that: *"The `deepseek-flash` model accepts images alongside text … The legacy model name `deepseek-v4-flash-vision-exp` is still accepted, but the model has been retired and its requests are served by the latest Flash model as well. Supported image formats: JPEG, PNG, GIF, and WebP. The format is detected from the actual file content, not from the file name or the declared MIME type."* tellme therefore must **not** copy the substring rule; capability is **declared**, not inferred (Q2).

Entering point: an **agent tool** (operator choice A) — the model decides when to look, exactly like the reference.

## Decisions taken upstream (`/axb-clarify`, one at a time → `spec.md` S-1…S-8)

| # | Decision |
| --- | --- |
| Q1 → 1 | **OpenAI-compatible family only**; the Gemini `inline_data` path is a forward item. |
| Q2 → 1 | **Explicit per-provider `VISION` config key** (bool, default off); no model-name inference. |
| Q3 → 1 | **`read_image` only** (no video, no documents). |
| Q4 → 1 | **The capability decides the offered surface** — `read_image` is hidden when `VISION` is off. |
| Q5 → 1 | **Inline only, 32 MiB ceiling; oversize = a loud tool error** (the Files-API leg is a forward item). |

## Decisions (this phase)

### D1 — Entry point: a `read_image` agent tool on the existing `domain/tools.Tool` port

- **Decision**: add one agent tool `read_image(filepath, reason)` (plus the shared resource params) on the existing port; the model calls it like any other tool. No CLI `--image` flag this round.
- **Rationale**: reference parity (its `read_image`); the model decides when to look; reuses the whole round-024/056 tool contract (resource bound, `reason` gate) unchanged (I-3).
- **Alternatives considered**: (a) a CLI `--image <path>` prompt attachment — deferred (a different entry point; the user message path is the Gemini/`user`-only concern, not needed for the tool journey); (b) a provider-level "attach last screenshot" implicit behaviour — rejected (magic, not agentic).

### D2 — Family scope: serialize as inline base64 `image_url` blocks (OpenAI-compatible only)

- **Decision**: the image is carried on the OpenAI-compatible wire as a content block `{"type":"image_url","image_url":{"url":"data:<mime>;base64,<data>"}}`. The Gemini family is **out of scope**.
- **Rationale**: it is the `deepseek` family's documented inline path (the operator's live case); the Gemini `inline_data` shape carries the same closed-wire concerns as round 061 / ADR 0031 and deserves its own round.
- **Alternatives considered**: (a) both families — rejected (Q1 → 1: doubles the transport surface); (b) an external-URL mode — deferred (a different mechanism).

### D3 — Capability: an explicit per-provider `VISION` key

- **Decision**: the provider entry gains `VISION` (boolean, **default false**). tellme gates image capability **only** on this key — never on a model name/substring, a family, or a served-API probe.
- **Rationale**: the reference's substring rule is **stale** (the operator's doc retires `…-vision-exp`); an explicit key is deterministic, rot-free, and matches tellme's established config-gated pattern (round 024's config-gated `CONTEXT_WINDOW`).
- **Alternatives considered**: (a) a corrected model-name allowlist — rejected (rots); (b) no gate — rejected (a wrong provider would fail or silently lose the image).

### D4 — Media placement in the folded-back conversation: a `user` message after the tool result

- **Decision**: when a round executes a tool that produced media, the loop appends **one `user` message** whose content array leads with the image block(s), immediately **after** the round's `tool`-result message(s); the image is hence on a **user** message. Media lead the message (media-first), matching the reference's `executor.AssembleResponse` ordering.
- **Rationale**: DeepSeek's doc accepts images in **user** messages only (so a media-bearing `tool` message is not safe), and the reference's own assembly puts `InlineData` on a **`user`** turn with the `functionResponse` after it. A dedicated user message satisfies both, and never leaves the `tool` result unanswered.
- **Alternatives considered**: (a) a content array on the `tool` message — rejected (not user-only per the DeepSeek doc); (b) attaching the image to the *next* request's prompt — rejected (the model may answer instead of calling again).

### D5 — Domain media carrier: `llm.Message` gains an optional media list

- **Decision**: `internal/domain/llm.Message` gains `Media []MediaPart` (`MediaPart{ MIMEType string; Data []byte }`). A message with **no** media serializes **exactly** as today (string `content`) — **I-1 byte-identity**; a message **with** media serializes as a content **array** (a leading text part when the message has text, then one `image_url` block per media part). `Request` needs no new field (media rides the `Messages` already folded by the loop).
- **Rationale**: the smallest change that keeps the text path untouched and carries the image; the media list is a domain value (adapter-agnostic), so the wire shape stays in the adapter.
- **Alternatives considered**: (a) a full `[]Part` union (text/image/tool) replacing `Content string` — rejected (a large refactor touching the history replay, the estimator, and both transports, for one feature); (b) a raw-bytes `any` on the message — rejected (untyped, un-testable).

### D6 — Type detection: magic bytes, never the name

- **Decision**: a pure helper resolves the MIME from the file's **content** — JPEG (`FF D8 FF`), PNG (`89 50 4E 47 0D 0A 1A 0A`), GIF (`GIF87a`/`GIF89a`), WebP (`RIFF` + bytes 8–11 `WEBP`). The extension and any declared MIME are **not** trusted; an unrecognized content type is a **loud** tool error naming the detected/unknown type.
- **Rationale**: the operator's doc: *"detected from the actual file content, not from the file name or the declared MIME type"*; it is also the more honest behaviour for a mislabelled file.
- **Alternatives considered**: (a) extension → MIME — rejected (the doc says otherwise; a `.png` that is text must fail loudly); (b) `http.DetectContentType` — a reasonable stdlib base but it also asserts non-image types (`text/plain`), so a small explicit table is clearer and testable.

### D7 — Size ceiling: 32 MiB inline, oversize is a loud error

- **Decision**: an image whose bytes exceed **32 MiB** is a **loud** tool error naming the limit, emitted **before** any request carrying the image; the file is never truncated or partially sent. The reference's Files-API upload leg (32–64 MiB, `purpose=user_data`, turn-scoped cleanup) is a **forward item**.
- **Rationale**: the OpenAI-compatible wire accepts inline base64 images up to 32 MiB (ADR-070); the upload leg is real machinery (an upload call, `file_id` references, a delete-on-exit path, its own failure modes) that the operator chose to defer (Q5 → 1).
- **Alternatives considered**: (a) the upload leg (reference parity) — deferred; (b) a tunable/larger inline ceiling — rejected (the wire still rejects > 32 MiB inline; a bigger number would be a lie).

### D8 — The Gemini adapter refuses media loudly (no silent drop)

- **Decision**: the OpenAI-compatible adapter carries media (D2); the **Gemini** adapter, if it ever receives a media-bearing message, returns a **loud** `*llm.ProviderError` ("this provider family cannot carry images yet") rather than dropping the image. The `VISION` key is honoured by the **openai** family only this round (operator guidance: set `VISION: true` only for an openai-family provider).
- **Rationale**: I-4 — a capability the operator declared must never silently lose the image; a loud failure is the honest alternative to an out-of-scope `inline_data` implementation.
- **Alternatives considered**: (a) silently drop media on Gemini — rejected (silent loss, the exact failure the round exists to prevent); (b) implement `inline_data` now — rejected (Q1 → 1).

### D9 — The offered set is a function of capability; the round-031 gate stays

- **Decision**: `agentTools()` gains a **vision flag** input; `read_image` is assembled only when it is true. The round-031 well-formedness gate (`required ⊆ properties`, over the non-overridable assembler) is preserved for the new tool and both assembler variants.
- **Rationale**: S-5/FR-007 + round-031's guarantee (a strict provider rejects a malformed declaration).
- **Alternatives considered**: (a) always offer + a recoverable refusal — rejected (the model wastes a round; the offered list would lie about capability); (b) a nil-tool stub that errors — rejected (same lie).

### D10 — Governance: a new ADR 0032; no ADR amended

- **Decision**: **ADR 0032** records the `VISION` capability key (and why the reference's substring rule is not copied), the OpenAI-compatible-only scope, the 32 MiB inline ceiling + loud oversize, the media placement, and §Forward (Gemini `inline_data`, the Files-API leg, video, documents, the CLI `--image` flag, image persistence). No existing ADR is amended (this adds an orthogonal capability; it does not weaken ADR 0031/0025).
- **Rationale**: tellme's standing policy (every durable decision is an ADR; adr-index-consistent).
- **Alternatives considered**: (a) no ADR ("just a feature") — rejected (the capability gate is a durable, cross-cutting decision); (b) amend ADR 0070-style rules — not applicable (tellme has no such ADR).

### D11 — Verification: a hermetic unit pin + an E2E carrier; no new `make verify` member

- **Decision**: a unit pin over the OpenAI-compatible request body (a fake `http.RoundTripper` capturing the JSON) asserts the image block is present, decodes to the file's exact bytes, and carries the **sniffed** MIME; an E2E carrier drives the same path through the existing fake provider; a text-only control asserts byte-identity. No new `make verify` member.
- **Rationale**: the transform is pure, so a unit pin is a complete proof; it rides the existing `go test` / `make test` (the aggregate gate keeps its size).
- **Alternatives considered**: (a) a new `verify-*` gate — rejected (no aggregate growth); (b) an E2E-only proof — rejected (a unit pin is the sharper, red-capable carrier).

## Residual risks (non-blocking — forwarded)

- **Base64 expansion**: 32 MiB of file becomes ≈ 42.7 MiB of base64 in the request body; the reference's aggregate inline cap is 48 MiB, so one maximal image stays under it — but **multiple** images in one turn are unbounded this round (a forward item: an aggregate bound).
- **No image persistence**: the history step carries the tool's **text** result only, so a resumed session replays without the image (reference parity; a recorded limitation).
- **The Gemini family** silently cannot be vision-enabled until the `inline_data` round (D8 makes it a loud failure, not a loss).
- **`read_image` reads whatever path it is given** (no security layer — a settled exclusion, I-2); a huge *non-image* file is read then rejected by the sniffer (bounded by the sniff-window read, not the whole file — an implementation detail for `/axb-tasks`).

## Why this is the right shape for tellme

The reference's converter is a full typed per-family schema model; tellme's minimalism (one optional `Media` list on the message + one table-driven sniffer + one content-block serialization) buys the same observable capability for the one family the operator actually uses, keeps the text path byte-identical, and leaves the two genuinely-larger pieces (the Gemini `inline_data` shape and the Files-API upload leg) as named, durable forward items rather than silent gaps.
