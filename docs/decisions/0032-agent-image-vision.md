# ADR 0032 — Agent image vision: an explicit capability key, a content-sniffed image, and an inline image on the OpenAI-compatible wire

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** `tell-me-go` **ADR-070** (DeepSeek vision + `FileUploadMode`; its model-ID substring capability rule is the reference this ADR **deliberately does not copy** — the operator's current DeepSeek doc retires `deepseek-v4-flash-vision-exp` and makes the rolling `deepseek-flash` alias image-capable), [ADR 0013](0013-composition-root-injection.md) (the composition-root injection the capability + tool are wired through), [ADR 0025](0025-mcp-tool-call-reason.md) (the universal `reason` gate every tool honours), [ADR 0031](0031-provider-supported-schema-surface.md) (the closed-wire provider surface — the sibling concern for the **Gemini** family, deliberately out of scope here), round 024 (the tool resource contract), round 029 (`specs/plans/029-agent-write-tools` — the last tool added), round 062 (`specs/plans/062-agent-image-vision` — this ADR's round)

## Context

tellme is **text-only end to end** (measured 2026-09-19, `dev` @ `902642a`): the domain conversation message is `{Role, Content string}`; both transports emit text-only `content`/`parts`; the domain `tools.Tool.Execute` returns a `string`; and no media tool is registered. There is therefore **no path** by which an image can enter a turn — the model cannot see a picture the operator already has on disk.

The operator asked tellme to read images (as the reference tool can) and supplied the official DeepSeek doc:

> *"The `deepseek-flash` model accepts images alongside text … The legacy model name `deepseek-v4-flash-vision-exp` is still accepted, but the model has been retired and its requests are served by the latest Flash model as well. Supported image formats: JPEG, PNG, GIF, and WebP. The format is detected from the actual file content, not from the file name or the declared MIME type."*

This inverts the reference's assumption: `tell-me-go` decides "can this model see?" with `strings.Contains(model, "vision")` and its ADR-070 assumed `deepseek-v4-flash` was text-only. Copying that rule into tellme would make its default provider silently drop an image. The operator also chose the entry point (**an agent tool**) and settled four further decisions in a one-at-a-time clarify (family scope · capability determination · tool surface · offered surface · size ceiling).

## Decision

**D1 — Entry point is an agent tool, `read_image(filepath, reason)`.** The model decides when to look at a file, exactly like the reference; there is no CLI `--image` flag this round. The tool rides the existing `domain/tools.Tool` port, so it inherits the round-024 resource contract and the round-056 universal `reason` gate unchanged.

**D2 — Capability is an explicit per-provider config key, `VISION` (boolean, default false).** tellme gates image capability **only** on this key — never on a model name/substring, a family, or a served-API probe. Rationale: the reference's substring rule is **stale** (the operator's doc retires the `…-vision-exp` suffix and makes the rolling alias image-capable), so any naming heuristic is born-rotten; an explicit key is deterministic, rot-free, and mirrors tellme's established config-gated pattern (round 024's config-gated `CONTEXT_WINDOW`). A provider entry that declares `VISION: true` but whose model cannot accept images is an **operator declaration error** — the provider rejects it loudly (no guessing).

**D3 — The capability decides the offered surface.** With `VISION` off (**the default**), `read_image` is **not offered** to the model at all; with it on, it **is**. The offered tool set therefore becomes a **function of the selected provider's capability** — the honest answer to "what can this model do?" — and no turn can silently lose or downgrade an image. The round-031 well-formedness guarantee (`required ⊆ properties`, over the non-overridable assembler) is preserved for the new tool and for both capability variants.

**D4 — Family scope is the OpenAI-compatible family only; the image is an inline base64 `image_url` block.** The image is serialized on the OpenAI-compatible wire as `{"type":"image_url","image_url":{"url":"data:<mime>;base64,<data>"}}` (the `deepseek` family's documented inline path). The **Gemini** family's `inline_data` shape is **out of scope** this round (its closed-wire concerns are ADR 0031's sibling); the Gemini adapter, if it ever receives a media-bearing message, returns a **loud** `*llm.ProviderError` rather than dropping the image (**no silent loss** — the failure the round exists to prevent is not re-introduced on the other family).

**D5 — The image type is resolved from the file's content, never its name.** A pure, table-driven sniffer recognizes JPEG (`FF D8 FF`), PNG (`89 50 4E 47 0D 0A 1A 0A`), GIF (`GIF87a`/`GIF89a`), and WebP (`RIFF…WEBP`) from **magic bytes**. The declared extension and any client-side MIME guess are **not** trusted. An unrecognized content type is a **loud** tool error naming the detected/unknown type — a `.png` whose bytes are text must fail, not be sent mislabelled.

**D6 — Size ceiling is 32 MiB, inline only; oversize is a loud tool error.** An image larger than the wire's **32 MiB** inline limit is a loud tool error emitted **before** any request carries it — never truncated, and never partially sent. The reference's Files-API upload leg (32–64 MiB, `purpose=user_data`, turn-scoped cleanup) is a **forward item**, not this round.

**D7 — Media placement: a `user` message after the tool result, media-first.** The domain conversation message gains an optional media list (`llm.Message.Media []MediaPart{ MIMEType, Data }`). When a round executes a tool that produced media, the loop appends **one `user` message** whose content array leads with the image block(s), immediately **after** the round's `tool`-result message(s). Rationale: DeepSeek accepts images in **user** messages only (so a media-bearing `tool` message is unsafe), and the reference's own `executor.AssembleResponse` puts `InlineData` on a **`user`** turn with the `functionResponse` after it — the media-first ordering is copied. A message with **no** media serializes **exactly** as today, so a text-only turn stays **byte-identical**.

**D8 — Governance: this ADR stands alone; no existing ADR is amended.** It adds an orthogonal capability; it does not weaken ADR 0025/0031. **§Forward** (below) carries the deferred pieces.

**§Forward (deferred, non-blocking).**

- **RF-062-1** — the **Gemini** family's `inline_data` image path (D4) — carries ADR 0031's closed-wire concerns.
- **RF-062-2** — the **Files-API upload leg** (32–64 MiB, `purpose=user_data`, turn-scoped upload + delete-on-exit) for images past the inline ceiling.
- **RF-062-3** — **multiple images in one turn** have **no aggregate bound** (the reference caps aggregate inline at 48 MiB); a single ≤ 32 MiB image stays under it, so the gap only opens with two or more images.
- **RF-062-4** — **`read_video`** and **`read_document`** (video is unsupported by DeepSeek; documents are a separate provider extract mechanism).
- **RF-062-5** — a **CLI `--image <path>`** prompt attachment (a different entry point; the user-message path is where a future flag would land).
- **RF-062-6** — **image persistence / replay**: the history step carries the tool's **text** only, so a resumed session replays without the image (reference parity).
- **RF-062-7** — a **capability-inference engine** (model-name probing or a served-API probe) — deliberately not built (D2).
- **RF-062-8** — an **aggregate** inline/request-body bound and a per-turn media count guard.

## Consequences

### Positive

- tellme gains its **first non-text capability**: the agent can read a local screenshot/chart/photo and show it to a vision-capable provider, closing the reference gap for the operator's live `deepseek-flash` case.
- Capability is **declared, not guessed** — a stale naming rule cannot silently downgrade an image on a rotated model alias.
- The **text path is byte-identical** (D5/D7): a turn with no image is unchanged, so 060 rounds of text behaviour cannot regress.
- The offered tool list **tells the truth** (D3): the model is only offered what the selected provider can do.
- The failure modes are **loud** (D5/D6/D8): an unsupported type, an oversize image, or a media-on-Gemini turn fails visibly, never silently.

### Negative / Accepted Trade-offs

- **Family asymmetry** — the capability key is honoured by the **openai** family only this round; a `VISION: true` on a `gemini` provider turns an image turn into a loud provider error (recorded, D4/RF-062-1). The operator guidance is to set `VISION: true` only for an openai-family provider until RF-062-1 lands.
- **Inline-only** — an image just over 32 MiB cannot be sent (a loud error) where the reference would upload it (RF-062-2).
- **Base64 expansion** — 32 MiB of file becomes ≈ 42.7 MiB of request body; with more than one image there is no aggregate guard (RF-062-3/RF-062-8).
- **No persistence** — a resumed session replays the tool's text, not the image (RF-062-6).
- The sniffer recognizes exactly **four** formats (the operator-doc set); a future format (e.g. AVIF) needs a table row.

### Neutral

- `llm.Message` gains one optional field; the text-only path, the estimator, the history replay, and the Gemini transport are otherwise untouched.
- Image **generation** (`create_image`) remains out of scope; this is a read capability only.

## Alternatives Considered

1. **Copy the reference's model-name substring rule.** Rejected — stale (the operator's doc retires `…-vision-exp`); it would drop images on tellme's default provider.
2. **Both families in this round.** Rejected (Q1 → 1) — the Gemini `inline_data` shape doubles the transport surface and carries ADR 0031's class of closed-wire concerns.
3. **A CLI `--image` flag as the entry point.** Deferred (RF-062-5) — the agent tool is the operator's chosen entry point; a flag is a follow-up.
4. **A content array on the `tool` message.** Rejected — DeepSeek accepts images in user messages only; a user message after the tool result is the safe shape (D7).
5. **Always offer `read_image` and return a recoverable refusal when the provider cannot see.** Rejected (Q4 → 1) — the offered list would lie about the capability and the model would waste a round.
6. **Silently drop media on the Gemini family.** Rejected — silent loss is precisely the failure class this round removes; a loud error is the honest alternative (D4).
7. **A full `[]Part` union replacing `Message.Content string`.** Rejected for this round — a large refactor (history replay, estimator, both transports) for one feature; the optional media list is the smaller change that keeps the text path byte-identical.

## Verification

- **Unit** — a pin over the OpenAI-compatible request body (a fake `http.RoundTripper` capturing the JSON): the image block is present, decodes to the file's **exact** bytes, and carries the **sniffed** MIME; a text-only control asserts **byte-identity**; the sniffer's four formats + an unknown content type; the 32 MiB boundary both ways; media-on-Gemini refuses loudly.
- **E2E** — the existing fake provider drives `read_image` end to end (the request carries the image; the offered set contains/omits `read_image` by capability; the loud refusals appear), riding `go test` / `make test` (no new `make verify` member).
- **Falsifiability** — removing the image serialization turns the unit pin red; hiding/offering `read_image` by capability is asserted against the production assembler; the byte-identity control fails if the text path drifts.

## References

- Operator's DeepSeek doc (2026-09-19): *`deepseek-flash` accepts images; `deepseek-v4-flash-vision-exp` retired; JPEG/PNG/GIF/WebP; type detected from content.*
- `tell-me-go` ADR-070 (DeepSeek vision + `FileUploadMode`) — the parity precedent, with the capability rule this ADR replaces.
- `specs/plans/062-agent-image-vision/` — this ADR's round (`spec.md` S-1…S-8, `research.md` D1…D11).
