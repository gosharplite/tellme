# Truth Delta: 073-list-role-headers-and-rendered-body

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` RUN (2026-09-21)** — `research.md` D1–D9 + **ADR 0045** + `techstack.md` MODIFY ×4 + `docs/domain-model` MODIFY. Clarify resolved at specify time (Q1 → A · Q2 → A · Q3 → B). `/axb-api-plan` + `/axb-data-plan` to record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Session lifecycle flags* | the `-l` listing presents each message as a role header line (`[USER]`/`[MODEL]`) + body (**model** rendered by glamour, **operator** verbatim) + one blank line after every message; header accented blue/magenta only on a terminal `stdout` with `-r` off; count/selection semantics unchanged; tool activity stays omitted; width resolved best-effort | `spec.md` US1–US3, FR-001…FR-009; `research.md` D1–D7 |
| MODIFY | `specs/truth/techstack.md` — *Raw output flag* | `-r`/`--raw` now also governs the `-l` listing (model body verbatim, no colour), so the in-group relay recipe stays byte-plain | `spec.md` US1 FR-006 (clarify Q1 → A); `research.md` D8 |
| MODIFY | `specs/truth/techstack.md` — *Terminal detection* | the listing's colour gate wires a dedicated **`stdout`** probe + the `TELL_ME_FORCE_STDOUT_TTY` seam (tellme's first stdout probe, scoped to the listing); the answer path still wires no stdout probe, so Obs 1 stays OPEN | `spec.md` US2 FR-004/FR-005; `research.md` D3 |
| MODIFY | `specs/truth/techstack.md` — *Output rendering* | the listing reuses the one glamour renderer (same style/sanitizer/word-wrap/degraded fallback) for the listed model body — one rendering policy, not two | `spec.md` FR-009; `research.md` D6 |
| ADD | `docs/decisions/0045-list-role-headers-and-rendered-body.md` (+ index row) | the decision record: the port seam, the role enum, the stdout probe, the colour codes, the per-message shape, the one renderer, the best-effort width, the `-r` interplay, and the recorded divergences | `spec.md` SC-001…SC-004; `research.md` D1–D9 |
| MODIFY | `docs/domain-model/tellme.modelith.{yaml,md}` — `Session` | the offline readers' "mode-only config parse" phrasing becomes "an offline parse; `-l` additionally resolves the rendered width **best-effort** (degrades to the renderer default)" — rendered via `make modelith-render` | ADR 0041 (the model is load-bearing); `research.md` D7 |
| NOOP (checked) | the *Payload status line* / *Turn chrome* / *Post-turn status lines* rows | the listing writes no payload/metrics/chrome line and never enters `turns.log` (I-3) | `spec.md` NFR-002 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | The persisted `history.jsonl` / `history.archive.jsonl` shape is **unchanged** — the role naming and the rendering are presentation-only (I-2). | `spec.md` I-2; `research.md` D2 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | the listing Rules/Examples become the new per-message shape (header + rendered model body + verbatim operator body + blank separator + terminal-gated accents + `-r` raw) | `spec.md` US1–US3, FR-001…FR-009 |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | rewrite the listing Then-rows (`role: content` → the header/render/separator contract) + the new Given rows | `spec.md` FR-001…FR-009 |
| NOOP (checked) | `specs/truth/features/cli/dsl.md` (interface root) | the new sentence patterns are `history`-module-scoped unless shared (decided at refine) | `plan.md` W-x |
