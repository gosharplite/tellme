# System Analysis Plan — the media channel in-band (round 070)

**Plan Package**: `specs/plans/070-media-channel-in-band`
**Inputs**: `spec.md` · `research.md` (D1–D7) · `truth-delta.md` · `specs/truth/**`
**Anchor**: issue [#142](https://github.com/gosharplite/tellme/issues/142)

---

## 1. Interface inventory

| Interface | Kind | Where | Planner |
| --- | --- | --- | --- |
| The **agent tool loop**'s media path | `cli` | `internal/agent` (the `MediaTool` assertion + fold) | carried to its contract owner **`/axb-dsl-refine`** (no API/data/UI planner applies) |
| `read_image`'s result | `cli` | `internal/infrastructure/tools` | same |

Both are **one CLI end**: an internal construction/execution seam. No HTTP/OpenAPI surface, no persisted state, no user-facing UX surface. The change is **not user-visible** → no new acceptance journey (the `042/043/047/049/069` structural-round precedent).

## 2. Waves

| Wave | Delegates | Outcome |
| --- | --- | --- |
| **W1** | `/axb-api-plan` | **NOOP** — a single CLI end; no `specs/truth/contracts/**`. |
| **W2** | `/axb-data-plan` | **NOOP** — no persisted state; media is never persisted (ADR 0032). |
| **W3** | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI; no TUI change. |
| **W4** | `/axb-dsl-refine` (the CLI contract owner) | **NOOP** — no user-visible behaviour change; `reading-a-local-image.feature` stays green **unchanged**. |

## 3. Affected files (the change set as landed)

```text
internal/domain/tools/media.go        ADD     MediaPart (moved from domain/llm)
internal/domain/tools/tools.go        CHANGED + the MediaTool capability interface
internal/domain/llm/gateway.go        CHANGED Message.Media []tools.MediaPart (+ import)
internal/domain/llm/media.go          DELETE  the collector + the old MediaPart
internal/infrastructure/tools/image.go CHANGED ExecuteMedia (in-band) + a delegating Execute; drop the llm import
internal/agent/agentloop.go           CHANGED assert tools.MediaTool; drop the collector
internal/infrastructure/llm/{gemini,openai}/client.go  CHANGED serialize []tools.MediaPart
docs/decisions/0040-*.md (+ index)    ADD     the ADR; annotate ADR 0032 (RF-062-10 + D7a) + ADR 0039 (RF-069-1)
specs/truth/techstack.md              MODIFY  the *Image filesystem tool* row
specs/plans/070-media-channel-in-band/**  ADD
tests                                 CHANGED media type refs + image_test.go → ExecuteMedia
```

**No** change to `internal/config/**`, the CLI, the adapters' wire bytes, or `go.mod`.

## 4. Contract / layering notes

- The new edge `internal/domain/llm → internal/domain/tools` is a **legal intra-domain** import (RULE-C flags only domain→non-domain; all `internal/domain/*` rank alike). It is **acyclic** (`domain/tools` does not import `domain/llm`).
- `internal/infrastructure/tools` **loses** its `internal/domain/llm` import (the round's second cost) — it imports `internal/domain/tools` (already did) for `MediaPart`/`MediaTool`.
- The `Tool` port is **unchanged** — the capability is a **second, optional** interface (`MediaTool`), so the six non-media tools (and their ~40 `Execute` test sites) are untouched. *(The rejected shape — widening `Tool.Execute` — is settled in ADR 0040.)*
