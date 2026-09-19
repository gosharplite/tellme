# Plan — Gemini image vision: `inlineData` on the Gemini/Vertex wire (round 063)

**Plan Package**: `specs/plans/063-gemini-image-vision`
**Truth root**: `specs/truth` · **Interface kind**: `cli` (plain line CLI — no `ui/**`, no TUI surface)

## Interfaces inventoried

| Interface | Kind | Planner | Wave |
| --- | --- | --- | --- |
| The CLI chat end (`specs/truth/features/cli/chat/**`) | `cli` | **carried to its contract owner** `/axb-dsl-refine` (a line CLI has no API/data/UI planner) | 1 |

- `/axb-api-plan` — **NOOP** (no OpenAPI/HTTP surface; a single CLI end).
- `/axb-data-plan` — **NOOP** (checked): the round adds **no persisted shape** — the image stays in-memory/in-flight (ADR 0032 D5/D7); a `read_image` step persists its **text** result only. The round is **adapter-side** (the Gemini transport) + the family-aware ceiling input; the in-memory `llm.Message.Media` is unchanged and outside the dbml's modelled scope (persisted local-state logs only).
- `/axb-ui-plan` — **skipped** (plain line CLI; no screen/keybinding changes).

## Waves

**Wave 1 (single, no dependencies)** — the CLI contract owner refines the round-063 acceptance journey into executable truth:
- the existing **`chat/reading-a-local-image.feature`** MODIFY-ed with a **Gemini journey** (Rules: the Gemini model is shown the image; one declaration serves both families; the family-aware loud refusals) — reusing the round-062 Then sentences widened to be family-aware;
- **`chat/dsl.md`** MODIFY-ed: the round-062 image/offered/refusal Then rows become **family-aware** (any image wire), and a new `## Given (round 063)` block adds the three Gemini-provider Givens;
- **no** new tool, config key, or offered-set change (the gate is unchanged — Q1 → A).

## The change surface (RD)

| # | Site | Change |
| --- | --- | --- |
| 1 | `internal/infrastructure/llm/gemini/client.go` | **ADD** the `inlineData` serialization in `buildContents` (a media-bearing message → one `user` `contents` entry with `inlineData` blob parts, after the tool-result turn); **RETIRE** the `hasMedia` refusal; delete the now-unused `hasMedia` |
| 2 | `internal/infrastructure/tools/image.go` | the tool takes a **resolved image byte ceiling** input (`NewReadImageTool(maxBytes int)`), enforced in place of the hardcoded `imageMaxBytes` constant (round 063 D4/D7) |
| 3 | `internal/app/deps/deps.go` + `cmd/tellme/deps.go` | the registry builder gains the resolved ceiling alongside the vision flag; the composition root resolves the family → ceiling (single-owned) |
| 4 | `tests/e2e/fakeprovider/fakeprovider.go` | `ToolNamesAt` becomes **family-aware** (also read `tools[].functionDeclarations[].name`) so the offered-set Thens work on the Vertex wire (round 061 only used a declaration-scan) |
| 5 | `tests/e2e/steps/step_r062_image.go` | the image/refusal/offered Then helpers become **family-aware** (scan `inlineData` as well as `image_url`; read the `read_image` result on either wire shape) |
| 6 | `tests/e2e/steps/step_r063_image.go` (new) | the three Gemini-provider Givens (Vertex config + service account + `VISION`; a scripted `read_image`; blind vs seeing) |
| 7 | `internal/infrastructure/llm/gemini/client_image_test.go` | **INVERT** the round-062 refusal pin to a "media ⇒ the `inlineData` blob" pin |
| 8 | `internal/infrastructure/llm/gemini/client_test.go` (pins) | the wire-level pin: the `inlineData` part present on a `user` `contents` entry **after** the tool result, `data` decodes to the file's exact bytes, `mimeType` sniffed; a media-free control is **byte-identical** |
| 9 | `docs/domain-model/tellme.modelith.{yaml,md}` | refresh the capability fact (`Tool.gate`/`ImageContent`/`Provider.vision` wording: carried by **both** families); `make modelith-check` green (ADR 0030) |

## Layer/architecture notes

- The **ceiling** (a resolved scalar) rides the existing composition-root injection (ADR 0013) — no new package, no cross-layer edge; `verify-architecture` untouched.
- The **wire shape** stays inside `internal/infrastructure/llm/gemini` (the adapter) — one concern per seam.
- **No dependency change** (`go.mod`/`go.sum` unchanged); stdlib only (`encoding/base64`, `encoding/json` already imported); POSIX-only.
- The loop's placement (ADR 0032 D7 — a trailing `user` media message) is **reused unchanged**; the adapter renders it as its own Vertex `user` turn (D2: the #1441 hazard avoided by construction).

## Test strategy

| Layer | Carrier |
| --- | --- |
| Unit (wire) | a fake `http.RoundTripper` over `gemini.Complete`: the `inlineData` part present on a `user` turn after the tool result, `data` decodes to the file's **exact** bytes, `mimeType` the **sniffed** type; a media-free control is **byte-identical** (I-1) |
| Unit (retired refusal) | the round-062 `client_image_test.go` **inverted**: a media-bearing request no longer refuses (asserts the blob) |
| Unit (ceiling) | the family-aware boundary both ways (at / one byte over the injected ceiling), and the composition-root family→ceiling mapping |
| Unit (offered set) | unchanged round-062 pins (capability-driven) — `read_image` offered iff `VISION` |
| E2E | the truth feature's new Gemini journey (the Vertex-shaped fake records the request; the image reaches that wire; the offered set; the loud family-aware refusals) |

Witnesses (reproduced then reverted): (a) removing the `inlineData` serialization turns the gemini unit pin **RED**; (b) moving the injected ceiling flips the oversize boundary RED; (c) [context] the round-062 byte-identity control fails if the Gemini text path drifted.
