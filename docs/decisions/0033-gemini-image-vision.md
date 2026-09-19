# ADR 0033 — Gemini image vision: the `inlineData` blob on the Gemini/Vertex wire, and a family-aware inline ceiling

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0032](0032-agent-image-vision.md) (agent image vision — the capability key, the content sniff, the OpenAI-compatible `image_url` block, and the loud Gemini refusal this ADR **lands forward item RF-062-1** for and **narrows** D4/D8 of), `tell-me-go` (the parity precedent: its Gemini adapter carries `InlineData → Blob` ungated, and `normalizeUserTurnParts` heals the `[InlineData][FunctionResponse][other]` ordering, #1441), [ADR 0031](0031-provider-supported-schema-surface.md) (the closed-wire provider surface — the sibling concern, untouched here), [ADR 0013](0013-composition-root-injection.md) (the composition-root injection the resolved ceiling rides), round 024 (the tool resource contract), round 062 (`specs/plans/062-agent-image-vision` — the capability), round 063 (`specs/plans/063-gemini-image-vision` — this ADR's round)

## Context

Round 062 (ADR 0032) gave tellme its first non-text capability: a `read_image` agent tool, an explicit per-provider `VISION` key, an inline base64 `image_url` block on the **OpenAI-compatible** wire, and — because the **Gemini/Vertex** family had no image path — a **loud refusal** at the Gemini adapter (`this provider family cannot carry images yet; remove VISION: true or use an OpenAI-compatible provider`). ADR 0032 D4 recorded the Gemini `inline_data` shape as forward item **RF-062-1**, and its Consequences told the operator to set `VISION: true` only for an OpenAI-compatible provider until that item landed.

The operator asked whether the reference's Vertex Gemini can see images (it can), then asked tellme to match. Grounding confirmed that **everything except the transport already exists and is family-agnostic**: the domain carrier (`llm.Message.Media []MediaPart`), the per-call media channel (`llm.WithMediaCollector`/`AttachMedia`), the loop's placement (a trailing `user` message after the tool result), the `read_image` tool (content sniff + ceiling + attach), and the offered-set gate (offered **iff** the selected provider declares `VISION`). Only the Gemini adapter's serialization was missing — and the capability gate needed **no** change, because it was already family-agnostic; the *refusal*, not the gate, was family-specific.

A one-at-a-time clarify settled two decisions: the capability key is **reused** (Q1 → A), and a **family-aware** inline ceiling is enforced by the tool (Q2 → B). The `inlineData` *placement* was left to this phase.

## Decision

**D1 — The Gemini/Vertex adapter serializes media as an `inlineData` proto-JSON blob; the round-062 refusal is retired.** The capability plumbing is family-agnostic, so this round is **adapter-side** (one `buildContents` branch, the removal of the `hasMedia` refusal) plus the truth corrections. No new tool, no new capability key, no domain change, no new dependency. ADR 0032 D8's "never silently lose an image" is now satisfied **by carrying** the image rather than by refusing it.

**D2 — Placement: the media message becomes its own `user` `contents` entry, after the tool-result turn.** In `buildContents`, a message with media maps to **one** Vertex `contents` entry `{"role":"user","parts":[<a text part when the message has text>, <one `inlineData` part per media part>]}` — media lead the turn — emitted **after** the tool result's `functionResponse` `user` turn. Rationale: the loop **already** appends a separate trailing `user` media message (ADR 0032 D7), so this is the **minimal** change and needs no cross-message correlation; and it **avoids by construction** the in-turn hazard the reference heals (`functionResponse` before `inlineData` — #1441), because the inline data never shares a turn with a `functionResponse`. A `user` turn carrying only `inlineData` (optionally text) is a valid `Part` set, and consecutive `user` turns are accepted (`contents` is a list; roles may repeat). The reference's **merge** into the tool-result turn is **rejected** here: tellme's loop does not co-locate the media with the tool result, so merging would force the adapter to correlate a media message with the preceding `functionResponse` turn for no observable gain.

**D3 — The wire keys are camelCase `inlineData`/`mimeType`.** Vertex REST accepts proto-JSON, and tellme's Gemini adapter already emits camelCase for **every** other key (`functionCall`, `thoughtSignature`, `maxOutputTokens`, `systemInstruction`, `generationConfig`), matching the SDK's json tags. The base64 is `base64.StdEncoding`. (Snake `inline_data`/`mime_type` is the alternate REST form; consistency wins. The spec's `inline_data` is the **concept** name; this decision pins the literal.)

**D4 — A family-aware inline ceiling, enforced by `read_image` (Q2 → B).** `read_image` gains a resolved **image byte ceiling** input (file bytes). A single-owned lookup maps the family: **OpenAI-compatible = 32 MiB** (unchanged, round 062) and **Gemini/Vertex = 14 MiB** — a conservative value **derived** from the documented ≈20 MB inline-data **request-size** bound (a raw ceiling of 14 MiB base64-expands to ≈18.7 MiB ≈ 19.6 MB, under the bound, so a file the tool accepts cannot be rejected purely for size). Oversize is a **loud tool error naming the limit**, emitted **before** the wire, on **both** families — preserving round 062's diagnostic property. The composition root resolves and injects the value (the same seam that carries the vision flag). The value lives in **one** place so it moves trivially if the documented figure or its unit is corrected.

**D5 — Capability stays `VISION`; the offered surface is unchanged.** No config change (Q1 → A): `VISION` remains the single, family-agnostic declaration, `read_image` is offered **iff** it is true, and the round-062 offered-set truth is untouched.

**D6 — One content sniff, shared by both families.** The Gemini path reuses the existing `read_image` tool and its magic-byte sniff verbatim; the adapter only base64-encodes the bytes the tool produced. A second sniff would be a drift source.

**D7 — The seam is minimal; RF-062-10 remains the refactor candidate.** The assemblage path widens by one scalar (`NewReadImageTool(maxBytes int)` + the assembler input). A `ToolSetSpec` replacing the bare `vision bool` stays the recorded next-round refactor (ADR 0032 RF-062-10); this round adds a scalar rather than pre-empting it.

**D8 — Verification: a hermetic Gemini wire pin, the retired-refusal pin inverted, and an E2E carrier.** A unit pin over the Vertex request body (fake `http.RoundTripper`) asserts the `inlineData` part is present, its `data` decodes to the file's **exact** bytes, and its `mimeType` is the **sniffed** type; the round-062 `client_image_test.go` inverts from "media ⇒ refusal" to "media ⇒ the blob"; an E2E carrier drives the path through the existing Vertex-shaped fake provider; witnesses: removing the serialization reds the unit pin, and moving the injected ceiling reds the boundary Example. No new `make verify` member.

**D9 — Governance: this ADR extends ADR 0032 and supersedes its forward item RF-062-1.** It **narrows ADR 0032 D4/D8** (the family asymmetry / the loud Gemini refusal) — ADR 0032's body is **not** edited beyond annotating its `Status` line and the §Forward RF-062-1 entry with a supersession pointer (the ADR-0027/0028 precedent). ADR 0032 remains an `Accepted` historical record.

**D10 — The Gemini text path stays byte-identical.** A media-free Gemini request serializes exactly as today; the media branch is additive.

**D11 — No new dependency; stdlib-only; hermetic.** `encoding/base64` + `encoding/json` only; POSIX-only; every gate offline.

## §Forward (deferred, non-blocking)

- **RF-063-1** — the Gemini ceiling figure (14 MiB) is **derived, not measured**; confirm the documented bound and its unit at the closeout live check, then move the single constant if it differs. The acceptance fixture expresses "larger than the Gemini image size limit" **relative** to the limit, so it follows automatically.
- **RF-063-2** — the standalone-`user`-turn placement (D2) is argued valid and hazard-free but **unproven live**; the closeout live check is its witness (the round-059/061 class).
- **RF-063-3** — an **aggregate** inline/request bound for **multiple** images in one turn (inherits round-062 RF-062-3/RF-062-8) — unchanged on both families.
- **RF-063-4** — the **Files-API upload leg** for images past the inline ceiling (round-062 RF-062-2) — still deferred; inline-only on both families.
- **RF-063-5** — the reference's **image-dimension guard** (longest-edge) and its per-family media plumbing are not adopted; a future dimension/format policy is a separate decision.
- **RF-063-6** — ADR 0032 **RF-062-10** (the `ToolSetSpec` seam / media returned from `Execute`) remains the next structural refactor; this round adds one scalar to the existing seam.

## Consequences

### Positive

- The operator's Vertex/Gemini case reaches **parity** with the OpenAI-compatible case: the same `read_image`, the same `VISION` key, the model **sees** the image.
- The change is **small and local** — one adapter branch + one retired refusal + one injected scalar — because the round-062 plumbing was already family-agnostic.
- Capability remains **declared, not guessed**; the offered list still tells the truth; the failure modes stay **loud** (a family-aware oversize refusal before the wire).
- The **text path is byte-identical on both families**.

### Negative / Accepted Trade-offs

- **Inline-only** — an image over the family ceiling cannot be sent (a loud error) where the reference would upload it (RF-063-4).
- **The Gemini ceiling is a derived constant** until the live confirmation (RF-063-1); a conservative choice may refuse an image the endpoint would in fact accept.
- **Placement unproven live** until the closeout check (RF-063-2).
- **No aggregate bound** for multiple images (RF-063-3).
- The adapter's media branch assumes the loop's placement (a separate trailing `user` message); a future co-location change would revisit D2.

### Neutral

- No domain, config, history, or estimator change; `llm.MediaPart` is reused as-is (the round-062 estimate already counts media, so the payload figure reflects an image on both families).
- Image **generation** and **persistence** remain out of scope (round-062 RF-062-5/RF-062-6).

## Alternatives Considered

1. **A second Gemini capability key.** Rejected (Q1 → A) — the key is already family-agnostic; a second spelling would let an entry claim one family's capability for the other.
2. **One global 32 MiB ceiling.** Rejected (Q2 → A, the operator's call) — it would make oversize a *provider* error on Gemini, the weaker diagnostic round 062 deliberately avoided.
3. **Merge the image into the tool-result `functionResponse` turn (the reference's normalized shape).** Rejected (D2) — tellme's loop does not co-locate media with the tool result; merging adds stateful correlation for no observable gain, and the standalone user turn avoids the #1441 hazard by construction.
4. **Snake_case `inline_data` keys.** Rejected (D3) — inconsistent with the adapter's other keys; the API accepts both.
5. **Measure the Gemini bound with a live probe now.** Rejected for the round — no credential and the E2E suite is network-guarded; recorded as RF-063-1 with a conservative derived value.
6. **Keep the loud Gemini refusal and add an opt-in.** Rejected — the refusal exists only because there was no path; the operator asked for parity.
7. **Introduce the `ToolSetSpec` seam now.** Deferred (D7 / RF-063-6) — bigger blast radius, not required by the round.

## Verification

- **Unit** — a pin over the Vertex request body (fake `http.RoundTripper`): the `inlineData` part is present on a `user` turn after the tool result, its `data` decodes to the file's **exact** bytes, and its `mimeType` is the **sniffed** type; a media-free control asserts **byte-identity**; the family-aware ceiling boundary (at / one byte over) both ways; the retired-refusal pin inverted.
- **E2E** — the Vertex-shaped fake provider drives `read_image` end to end on a `gemini` provider (the recorded request carries the image; the offered set is unchanged; the loud family-aware refusals appear), riding `go test` / `make test`.
- **Falsifiability** — removing the `inlineData` serialization turns the unit pin **red**; moving the injected ceiling reds the boundary Example.
- **Live (non-gating, closeout)** — one real Vertex turn with an image (witnesses D2's placement and confirms D4's bound) — RF-063-1/RF-063-2.

## References

- Operator request (2026-09-19): *"does vertex gemini have vision (read_image)?"* → *"We need to let tellme gemini to have read_image too."*
- `tell-me-go` Gemini adapter (`inline_data` blob; `normalizeUserTurnParts` `[InlineData][FunctionResponse][other]`, #1441) — the parity precedent.
- [ADR 0032](0032-agent-image-vision.md) (the capability, the OpenAI-compatible block, RF-062-1 the round lands) · `specs/plans/063-gemini-image-vision/` (`spec.md` S-1…S-8; `research.md` D1…D11).
