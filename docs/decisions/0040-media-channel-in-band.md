# ADR 0040 — The media channel in-band: return media from the tool contract

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0032](0032-agent-image-vision.md) (agent image vision — this ADR **completes its §Forward RF-062-10**, and supersedes its **D7a** collector mechanism), [ADR 0039](0039-toolset-spec-capability-seam.md) (the `ToolSetSpec` seam — this ADR **delivers its §Forward RF-069-1**), [ADR 0033](0033-gemini-image-vision.md) (the family-aware ceiling — unchanged), [ADR 0011](0011-layer-discipline-gate.md) (layer discipline — the intra-domain edge is legal), [ADR 0021](0021-tool-output-ctor-injection.md) (the `OutputSink` ctor-injection precedent for a tool's effect), round 062 (`specs/plans/062-agent-image-vision`), round 069 (`specs/plans/069-toolset-spec-capability-seam`), round 070 (`specs/plans/070-media-channel-in-band` — this ADR's round), issue [#142](https://github.com/gosharplite/tellme/issues/142)

## Context

`read_image` (round 062) must hand image bytes to the agent loop so the loop can fold them onto a `user` message. But `Tool.Execute` returns **text**:

```go
Execute(ctx context.Context, arguments string, budget ByteBudget) (string, error)
```

and that value is consumed as text by the loop's byte clamp, the `[Tool Result]` chrome line, the persisted `history.Step.Result`, the token estimator, and the tool-role message `Content`. **ADR 0032 D7a** therefore smuggled the media out-of-band: the loop installed a per-call `context` collector (`llm.WithMediaCollector`), the tool wrote into it (`llm.AttachMedia`), and the loop folded the collected slice back. D7a recorded the two costs:

1. **the tool's full effect is not in its return value** — the loop must *know* to install the collector; and
2. **`internal/infrastructure/tools` imports `internal/domain/llm`** (the conversation model) to build a `MediaPart`.

*(This is a **contract/type-honesty** limit, not a Go `string` limit — bytes can ride a string, but every textual consumer above would corrupt them, and the conversation model needs a typed media part.)*

## Decision

**D1 — The media type `MediaPart` moves to `domain/tools` (the single owner).** `tools.MediaPart{MIMEType string; Data []byte}`; `llm.Message.Media` becomes `[]tools.MediaPart`. The import is a **legal intra-domain** edge (both `internal/domain/**`; the tier table ranks all of `internal/domain/*` alike, and RULE-C flags only domain→non-domain), and it is **acyclic** (`domain/tools` does not import `domain/llm`).

**D2 — The media effect is exposed through a segregated capability interface, not a `Tool` widening.** `domain/tools` declares:

```go
type MediaTool interface {
    Tool
    ExecuteMedia(ctx context.Context, arguments string, budget ByteBudget) (text string, media []MediaPart, err error)
}
```

The agent loop **type-asserts** `tools.MediaTool`; when it holds, it runs `ExecuteMedia` (which returns the text result **and** the media it produced, **in-band**). Every other tool keeps the plain `Tool` contract **unchanged**. A capability interface is the repo's established shape (`history.Seeder`, `LoopObserver`, `tools.MCPClient`).

**D3 — The loop drops the collector.** `llm.WithMediaCollector` / `llm.AttachMedia` are **deleted** (with `internal/domain/llm/media.go`). The loop folds the returned media onto the **same** `user` message as before (media-first, immediately after the tool result) — the emitted turn is byte-identical.

**D4 — `read_image` implements the capability and stops importing `domain/llm`.** `readImage` gains `ExecuteMedia` (the real body) and keeps `Execute` as a thin delegate (it must satisfy `Tool` to be registered). After this change, **no `internal/infrastructure/tools` production file imports `internal/domain/llm`**.

**D5 — Serialization is unchanged.** The adapters (`gemini` `inlineDataParts`, `openai` `dataURI`) take `tools.MediaPart`; the wire is byte-identical on both families.

**D6 — Behaviour is byte-identical; no config, no UX.** The emitted turn, the wire, the chrome, `stdout`, the exit codes, the `--tool-usage` union, and the persisted records are unchanged. The acceptance is **structural** (existing pins green with **no assertion changed**).

**D7 — Governance: this ADR completes ADR 0032 RF-062-10 and delivers ADR 0039 RF-069-1.** It annotates ADR 0032's `Status` + `RF-062-10` (now **fully delivered**) and its **D7a** (the collector is **superseded** by this ADR), and ADR 0039's `Status` + `RF-069-1` (**delivered**). The **rejected shape** (a) is recorded below as a **settled rejection** — a decision, not an open forward item.

## Consequences

### Positive

- **The tool contract tells the truth**: a media-producing tool returns its media **in its return value** (`ExecuteMedia`); the loop no longer needs a hidden convention to install a collector.
- **The accidental cross-concern import disappears**: `internal/infrastructure/tools` no longer reaches into `internal/domain/llm`.
- **One media owner** (`tools.MediaPart`); one media path (the loop's fold); the wire is unchanged.
- The **RF-062-10 lineage ends** — no remaining half.

### Negative / Accepted Trade-offs

- **A capability interface rather than a uniform port** (the rejected shape (a)): a reader of `tools.Tool` alone does not see media; discovering it needs the `MediaTool` assertion. Accepted — the loop is the only consumer, and the alternative churns every tool + ~40 test sites for a capability one tool has.
- `read_image` now has **two execute methods** (`Execute` delegates to `ExecuteMedia`); a direct `Execute` caller sees text only. Accepted (the loop is the consumer).
- `llm` now imports `domain/tools` for the media element type. Accepted (legal, acyclic, and it removes the worse tools→llm coupling).

### Neutral

- No config key, tool schema, wire byte, exit code, or persisted record changes; the reason gate (ADR 0025) and the round-024 resource contract are untouched.

## Settled rejection (recorded — a decision, NOT a forward item)

**Shape (a): widen `Tool.Execute` to return a `Result{Text string; Media []MediaPart}` uniformly.** **Rejected** — it forces every tool implementation (7 + the MCP tool + registry stubs) and **~40 `Execute` test call sites** to change for a capability exactly **one** tool has today, and it would place an always-empty `Media` field on six tools. The capability interface (`MediaTool`) achieves the round's goal (the media is in `ExecuteMedia`'s return value; `infrastructure/tools` drops the `domain/llm` import) with localised, honest churn. **This is a decision of record — do not re-raise it as a follow-up.** *(The anti-muse rule: a structural round records the shape it did not take as settled, rather than leaving a live candidate — the exact mechanism that stretched RF-062-10 across rounds 062→070.)*

## Alternatives Considered

1. **Keep the collector (do nothing).** Rejected — it is RF-062-10's remaining half / RF-069-1 (#142).
2. **Shape (a), widen `Tool.Execute`.** Rejected — see the settled rejection above.
3. **A type alias `llm.MediaPart = tools.MediaPart`** (no test churn). Rejected — two spellings for one concept; the relocation is a one-time mechanical update, and an alias would hide the ownership.
4. **A new neutral `internal/domain/media` package.** Rejected — overkill for a two-field struct; `domain/tools` is its producer's home (and the operator's issue stipulates it).
5. **Have the loop translate `tools.MediaPart` → a private `llm.MediaPart`.** Rejected — duplicates the type (the repo's single-owner rule).

## Verification

- **Structural** — `llm.WithMediaCollector` / `llm.AttachMedia` absent (grep); no `internal/domain/llm` import under `internal/infrastructure/tools` (grep); the loop asserts `tools.MediaTool`.
- **Behaviour-identity** — the full unit suite + the godog E2E (`reading-a-local-image`, `calling-several-tools-in-one-round`) green with **no assertion changed**; the emitted turn byte-identical.
- **Falsifiability** — dropping the in-band media from the loop fold reds the image E2E.
- **Gates** — `gofmt`/`go vet`/`go build` clean; `go test -count=1 ./...` green; `make verify` **OK** (layer gate 0); `go.mod`/`go.sum` unchanged.

## References

- `internal/domain/tools/{tools.go,media.go}` (the `MediaTool` capability + `MediaPart`) · `internal/domain/llm/gateway.go` (`Message.Media`) · `internal/agent/agentloop.go` (the assertion + fold) · `internal/infrastructure/tools/image.go` (`ExecuteMedia`) · `internal/infrastructure/llm/{gemini,openai}/client.go` (serialization).
- [ADR 0032](0032-agent-image-vision.md) (D7a + RF-062-10) · [ADR 0039](0039-toolset-spec-capability-seam.md) (RF-069-1) · [ADR 0011](0011-layer-discipline-gate.md) · `specs/plans/070-media-channel-in-band/` (`spec.md` FR-001…FR-007; `research.md` D1…D7).

## §Forward (deferred, non-blocking)

> **⚠ Not open work.** A `§Forward` entry is a decision *deferred to a trigger* or a recorded divergence — **not** tasking. Do not re-raise absent its trigger.

- **RF-070-1** — a future third media producer (video/document) uses the **same** `MediaTool` capability; no new shape.
- **RF-070-2** — a reader of `tools.Tool` alone does not see media (the accepted trade of the settled-rejected shape (a)).
- **RF-070-3** — a direct `Execute` call on a media tool returns text only (the loop is the consumer).
