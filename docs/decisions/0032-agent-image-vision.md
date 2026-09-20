# ADR 0032 — Agent image vision: an explicit capability key, a content-sniffed image, and an inline image on the OpenAI-compatible wire

- **Status:** Accepted (**D4/D8 narrowed and §Forward RF-062-1 superseded by [ADR 0033](0033-gemini-image-vision.md)** — the Gemini family's `inlineData` image path; **§Forward RF-062-10 FULLY DELIVERED** — its `ToolSetSpec` seam half by [ADR 0039](0039-toolset-spec-capability-seam.md) (round 069) and its media-channel half by [ADR 0040](0040-media-channel-in-band.md) (round 070); **`D7a` (the media-channel mechanism) is SUPERSEDED by [ADR 0040](0040-media-channel-in-band.md)**; the rest of this ADR stands)
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

**D7a — The media CHANNEL is a per-call context collector (recorded mechanism; PR #129 fold F-062-3).** A tool attaches media by writing into a `*[]llm.MediaPart` the loop installs on the call's `context` (`llm.WithMediaCollector` / `llm.AttachMedia`; `internal/domain/llm/media.go`), so the loop can fold it onto a `user` message **without** widening the `Tool.Execute(…)(string, error)` port. This keeps the tool port and its resource contract unchanged, at the cost of two recorded facts: (1) the tool's full effect is **not** in its return value (a tool that attaches media mutates an ambient channel — the loop must know to install the collector); and (2) `internal/infrastructure/tools` imports `internal/domain/llm` (the conversation model) to write into it. The layer gate is **green** (the direction is legal under the pinned ranking), so this is a **cohesion**, not a gate, finding. The cleaner shape — a `domain/tools` media type returned from `Execute` (or a second method), translated by the loop — widens **every** tool's contract and is therefore **RF-062-10** (a next-round refactor candidate), not fold-sized. **SUPERSEDED by [ADR 0040](0040-media-channel-in-band.md) (round 070):** the channel is now **in-band** — `read_image` returns its media through the optional `tools.MediaTool` capability (`ExecuteMedia`), the loop type-asserts it, and the collector + this `domain/llm` import are removed.

**D7b — The offered/recorded tool sets are capability-aware (PR #129 folds F-062-1/F-062-5c).** The **offered** set is a function of the selected provider's `VISION` (D3); the **recorded/`--tool-usage`** surface lists the **union** of the base and capability-gated sets, so a tool that can be recorded is never invisible in the report (an accounting surface that cannot show a tool it counts is a defect). The E2E keeps the two authorities distinct: the offered-set assertion reads the **base** set, the all-zero tool-usage guard reads the **union**.

**D8 — Governance: this ADR stands alone; no existing ADR is amended.** It adds an orthogonal capability; it does not weaken ADR 0025/0031. **§Forward** (below) carries the deferred pieces.

**PR #129 folds (APPROVE WITH REQUIRED FOLDS — 2026-09-19).** The review found the feature sound (capability declared not guessed; text path byte-identical; refusals loud *and* diagnosable) and required five folds — none about the image path, all about two surfaces the round changed but did not carry: **(F-062-1)** the `--tool-usage` report now lists the capability-aware **union** (D7b) and the owning truth row + ledger record it; **(F-062-2)** the token estimate **counts media** (`EstimateTokens` adds a base64-expansion term per media part; `internal/domain/llm/token.go`), so the pre-flight payload figure and the budget view are no longer blind to an image (RF-062-9 **closed**); **(F-062-3)** the media channel is named here (D7a) and in the *Image filesystem tool* truth row; **(F-062-5a)** the two refusal E2E Thens are bound to the `read_image` tool result (matched by `tool_call_id`), not any message text; **(F-062-5b)** `STATUS.md` + the day's session summary are updated. Nits folded: the shared config writer + a `VISION` mutator in the E2E (F-062-5c); the refusal casing aligned to the readers' `ERROR:` (F-062-5d); the domain-model capability fact modelled as a `ToolGate` enum rather than a one-true boolean (F-062-5e, with the pre-existing four/two-family wording corrected). **F-062-4** (a `ToolSetSpec` seam instead of a bare positional `vision bool`) is recorded as a next-round refactor (**RF-062-10**).

**Fold verification (PR #129 at `744e10d`): FOLDS VERIFIED 5/5 — one truth fold-back TF-062-1.** The `--tool-usage` owning **interface** truth (`chat/dsl.md` — the `the review shows every tool with no uses` row) still stated the retired four-tool set; it is restated as the **recordable union** (the base ∪ capability-gated sets, eight tools) — and its `today:` enumeration was stale since rounds 029/033 — with the adjacent `accounting-for-the-tool-use.feature` comment and the `ui.FormatToolUsage` doc aligned, plus a ledger row (the round-057 TF-057-1 class). Cheap residual folds: **R-062-2** closed by a `unionToolNames` unit pin; **R-062-1** (the E2E union is a hand-kept duplicate of production's) and **R-062-3** (the report is capability-blind by design) recorded as **RF-062-12**/**RF-062-13**.

**§Forward (deferred, non-blocking).**

- **RF-062-1** — the **Gemini** family's `inline_data` image path (D4) — carries ADR 0031's closed-wire concerns. **SUPERSEDED — landed by [ADR 0033](0033-gemini-image-vision.md) (round 063):** the Gemini/Vertex family now carries the image as an `inlineData` blob and the loud refusal is retired.
- **RF-062-2** — the **Files-API upload leg** (32–64 MiB, `purpose=user_data`, turn-scoped upload + delete-on-exit) for images past the inline ceiling.
- **RF-062-3** — **multiple images in one turn** have **no aggregate bound** (the reference caps aggregate inline at 48 MiB); a single ≤ 32 MiB image stays under it, so the gap only opens with two or more images.
- **RF-062-4** — **`read_video`** and **`read_document`** (video is unsupported by DeepSeek; documents are a separate provider extract mechanism).
- **RF-062-5** — a **CLI `--image <path>`** prompt attachment (a different entry point; the user-message path is where a future flag would land).
- **RF-062-6** — **image persistence / replay**: the history step carries the tool's **text** only, so a resumed session replays without the image (reference parity).
- **RF-062-7** — a **capability-inference engine** (model-name probing or a served-API probe) — deliberately not built (D2).
- **RF-062-8** — an **aggregate** inline/request-body bound and a per-turn media count guard.
- **RF-062-9** — *(CLOSED by fold F-062-2)* the payload figure / budget check excluded media — now counted.
- **RF-062-10** — the **media-channel refactor**: return media from `Execute` (widening the tool contract via a `domain/tools` media type) instead of the per-call context collector (D7a); and the `ToolSetSpec` seam replacing the bare `vision bool` (F-062-4). **Split 2026-09-20 (round 069):** the **seam half is DELIVERED by [ADR 0039](0039-toolset-spec-capability-seam.md)** (the registry build takes one named `deps.ToolSetSpec`); the **media-channel half is DELIVERED by [ADR 0040](0040-media-channel-in-band.md) (round 070)** (`tools.MediaPart` + the `tools.MediaTool` capability; the collector removed). **RF-062-10 is now FULLY DELIVERED.**
- **RF-062-11** — the `(family, VISION)` mismatch is **static** and knowable at resolution time, so a resolution-time refusal could fail before a turn spends a round; D2 deliberately chose the provider-side refusal (recorded here as a note, not a defect).
- **RF-062-12** *(R-062-1)* — the E2E `recordableToolNames()` is a **hand-kept duplicate** of production's `unionToolNames(...)`; a third capability would need the test side edited too. A shared enumerator (exported from the composition root, or derived from the two DI registries) would close it.
- **RF-062-13** *(R-062-3)* — the `--tool-usage` report is **capability-blind by design** (it shows `read_image` for an operator with no vision provider) — the consequence of the union authority, recorded here (§Forward) and noted in the *Tool-usage accounting* truth row.
- **RF-062-14** *(PR #129 nit verification, `5255255480`)* — the `the request offered exactly the agent tools` sentence pins the **base offered set** (a provider that does not declare `vision`); its DSL row now says so (the offered set is capability-dependent, D3). A future pairing with a `vision` provider reds loudly (eight vs seven) — a safe, deliberate scope, not a gap.

## Consequences

### Positive

- tellme gains its **first non-text capability**: the agent can read a local screenshot/chart/photo and show it to a vision-capable provider, closing the reference gap for the operator's live `deepseek-flash` case.
- Capability is **declared, not guessed** — a stale naming rule cannot silently downgrade an image on a rotated model alias.
- The **text path is byte-identical** (D5/D7): a turn with no image is unchanged, so 060 rounds of text behaviour cannot regress.
- The offered tool list **tells the truth** (D3): the model is only offered what the selected provider can do.
- The failure modes are **loud** (D5/D6/D8): an unsupported type, an oversize image, or a media-on-Gemini turn fails visibly, never silently.

### Negative / Accepted Trade-offs

- **Family asymmetry** — *(CLOSED by [ADR 0033](0033-gemini-image-vision.md), round 063)* the capability key was honoured by the **openai** family only at this round; a `VISION: true` on a `gemini` provider turned an image turn into a loud provider error (D4/RF-062-1). **ADR 0033 lands RF-062-1**, so `VISION: true` now works on both families and the "set `VISION: true` only for an openai-family provider" guidance is retired.
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
