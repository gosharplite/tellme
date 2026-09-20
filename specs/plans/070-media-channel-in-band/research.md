# Technical Research — the media channel in-band (round 070)

**Plan Package**: `specs/plans/070-media-channel-in-band`
**Spec**: `spec.md` (anchor issue [#142](https://github.com/gosharplite/tellme/issues/142))
**Input (operator)**: *"Create a detail new issue for this. Open round 070, the goal is to close the new issue."*

**Grounding**: measured on `dev` @ `d6d606f`.

---

## Problem restated (the two ADR 0032 D7a costs)

A media-producing tool (`read_image`) cannot put its media in its result, so round 062 smuggled it out-of-band:

- **`Tool.Execute` is text-only** — `Execute(ctx, arguments string, budget ByteBudget) (string, error)` (`internal/domain/tools/tools.go:65`). Its value is consumed as **text** by the loop's `clampBytes`, the `[Tool Result]` chrome line (folded + rune-capped), the persisted `history.Step.Result`, the token estimator, and the tool-role message `Content`.
- **The per-call collector** — the loop installs a fresh `llm.WithMediaCollector` on the call context (`internal/agent/agentloop.go:153-154`), the tool writes with `llm.AttachMedia` (`internal/infrastructure/tools/image.go:136`), and the loop folds the collected slice onto a `user` message.
- **The consequence** — (1) the tool's full effect is **not** in its return value (the loop must *know* to install the collector); (2) `internal/infrastructure/tools` imports `internal/domain/llm` (the conversation model) merely to build a `MediaPart`.

*(Not a Go `string` limitation — bytes can ride a string; it is a **contract/type-honesty** limit: no field exists for a distinct media value, and every textual consumer above would corrupt it.)*

---

## Decisions

### D1 — The media type `MediaPart` moves to `domain/tools` (single owner)

`tools.MediaPart{MIMEType string; Data []byte}` becomes the **single** media type (relocated from `internal/domain/llm/media.go`). `llm.Message.Media` becomes `[]tools.MediaPart`. This is **arch-clean**: both packages are `internal/domain/**`, so the edge is a **legal intra-domain** import (RULE-C flags only domain→non-domain; the tier table ranks all `internal/domain/*` at the same tier). It is **acyclic** (`domain/tools` does not import `domain/llm`).

*Why not keep the type in `domain/llm`?* A tool would then have to import the conversation model to name its own output type — the exact coupling this round removes. The tool's output artifact is a tool-domain concept; the conversation model references it (the operator's issue #142 stipulates "a `domain/tools` media type").

### D2 — Shape: a segregated capability interface `tools.MediaTool` (NOT a `Tool` widening)

```go
// domain/tools/media.go
type MediaPart struct { MIMEType string; Data []byte }

// domain/tools/tools.go — the optional capability contract
type MediaTool interface {
    Tool
    ExecuteMedia(ctx context.Context, arguments string, budget ByteBudget) (text string, media []MediaPart, err error)
}
```

The agent loop **type-asserts** `tools.MediaTool` and, when it holds, executes via `ExecuteMedia` (which returns the text **and** the media, **in-band**); every other tool keeps the plain `Tool` contract **unchanged**.

**Rejected — shape (a): widen `Tool.Execute` to return a `Result{Text, Media}`.** Rejected because it forces **every** tool (7 implementations + the MCP tool + the registry stubs) and **~40 `Execute` test call sites** to change for a capability that **exactly one** tool has today; and a uniform "result struct" would put an always-empty `Media` field on six tools. Recorded here as a **settled rejection** (S-4 / the anti-muse rule) — a capability interface is the repo's established shape (`history.Seeder`, `LoopObserver`, `tools.MCPClient`), so a *second* shape is not an open question.

*Why the interface embeds `Tool`:* a `MediaTool` is a `Tool` that also produces media; embedding makes the assertion self-documenting and lets the registry keep holding `Tool` values unchanged.

### D3 — The loop drops the collector and folds the returned media

```go
var media []tools.MediaPart
if mt, ok := tool.(tools.MediaTool); ok {
    result, media, terr = mt.ExecuteMedia(tctx, tc.Arguments, tools.ByteBudget(byteBudget))
} else {
    result, terr = tool.Execute(tctx, tc.Arguments, tools.ByteBudget(byteBudget))
}
…
if len(media) > 0 { turn = append(turn, llm.Message{Role: "user", Media: media}) }
```

The fold is **identical** to today (media-first `user` message immediately after the tool result) — I-1. No `llm.WithMediaCollector`; no per-call sink.

### D4 — `read_image` implements the capability; `Execute` stays (it is registered as a `Tool`)

`readImage` gains `ExecuteMedia` (the real body, returning text + `[]tools.MediaPart`) and keeps `Execute` as a thin delegate (`text, _, err := ExecuteMedia(...)`). It no longer imports `internal/domain/llm` — only `internal/domain/tools`. **No `internal/infrastructure/tools` production file imports `internal/domain/llm`** (SC-001).

### D5 — `internal/domain/llm/media.go` is deleted

The whole file (the collector + the old `MediaPart`) goes; `llm.Message.Media` references `tools.MediaPart`. The adapters (`gemini` `inlineDataParts`, `openai` `dataURI`) take `[]tools.MediaPart` / `tools.MediaPart` — **serialization is unchanged** (I-1).

### D6 — Behaviour is byte-identical; no config, no UX

The emitted turn, the wire (both families), the chrome, `stdout`, the exit codes, the `--tool-usage` union, and the persisted records are unchanged. The acceptance is **structural** (existing pins green with **no assertion changed**).

### D7 — Governance: a new ADR (0040)

Record the in-band media channel; **annotate ADR 0032 §Forward `RF-062-10` as fully delivered** and **ADR 0039 §Forward `RF-069-1` as delivered**; record the rejected shape (a) as a **settled rejection**. This closes the RF-062-10 lineage (its seam half: round 069; its media half: this round).

---

## Truth impact (techstack.md)

- **MODIFY** — *Image filesystem tool (`read_image`)*: the media **channel** is now **in-band** (returned from `ExecuteMedia` via `tools.MediaPart`; the loop folds it), and the round-062 D7a collector note is removed; `internal/infrastructure/tools` no longer imports `internal/domain/llm`.
- **MODIFY** — *Agent tool loop* (or a tool-contract row): the loop type-asserts the optional `tools.MediaTool` capability and drops the per-call collector.
- **NOOP** — ev.ry other row (no behaviour, no dependency, no config key).

---

## Residual risks (forward items)

- **RF-070-1** — `MediaPart` moved to `domain/tools`, so **`MediaPart` is now a domain/tools concept**; a future third producer (video/document) uses the same capability interface — no new shape.
- **RF-070-2** — the capability interface means a reader of `tools.Tool` alone does not see media; discovering media requires the `MediaTool` assertion. Accepted (the loop is the only consumer). *(This is the deliberate trade of the rejected shape (a).)*
- **RF-070-3** — the tool's `Execute` on a media tool returns the text without media; a direct `Execute` caller (a unit test) sees text only — the loop is the intended consumer.

## Verification (expected — finalised in `tasks.md`)

- **Structural** — `llm.WithMediaCollector`/`llm.AttachMedia` absent (grep); `internal/infrastructure/tools` free of `internal/domain/llm` (grep); `tool.(tools.MediaTool)` present in the loop.
- **Behaviour-identity** — the full unit suite + the godog E2E (`reading-a-local-image`) green with **no assertion changed**; the emitted turn identical.
- **Falsifiability** — dropping the in-band media from the loop fold reds the image E2E; the collector's absence is grep-verifiable.
- **Gates** — `gofmt`/`go vet`/`go build` clean; `go test -count=1 ./...` green; `make verify` **OK**; `go.mod`/`go.sum` unchanged.
