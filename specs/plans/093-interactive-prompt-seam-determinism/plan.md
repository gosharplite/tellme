# System analysis — round 093 `093-interactive-prompt-seam-determinism`

**Owner**: `axb-system-analysis` (produces this `plan.md`; writes **no** `specs/truth/**`).
**Inputs**: `spec.md`, `research.md`, `truth-delta.md`, the current `specs/truth/**`.

## 1. Interface inventory

The round touches **one** system interface — the **CLI end** (a line-oriented terminal end: commands,
flags, stdin/stdout/stderr, exit codes). There is **no** HTTP/REST API and **no** web UI. The **observable
CLI surface is unchanged** (no new flag, no new output line, no new exit code); the change is confined to
the **E2E test seam** that drives the `-i` prompt (`tests/e2e/**`) and to the **semantics of one interface
Then** (a discriminating assertion).

| # | Interface | Kind | Description |
| --- | --- | --- | --- |
| I-1 | tellme CLI | `cli` | The `-i` interactive prompt is unchanged; the test seam that drives it becomes deterministic, and one interface Then's assertion becomes discriminating. |

## 2. Planner delegation (CLI-streamlined)

| Planner | Applicability | Disposition |
| --- | --- | --- |
| `/axb-api-plan` | none — a standalone CLI has no OpenAPI surface | **NOOP** (`specs/truth/contracts/**` untouched). |
| `/axb-data-plan` | none — the seam is in-memory test support; no persisted state changes | **NOOP** (`specs/truth/data/**` untouched). |
| `/axb-ui-plan` | none — a plain line-oriented CLI (no TUI surface change) | **Skipped** (`ui/**` untouched). |
| `/axb-dsl-refine` | **the CLI end's contract owner** | **INVOKED** — the `keeps the typed text` Then row's semantics + the feature header note. |

## 3. Dependency waves

| Wave | Order | Work |
| --- | --- | --- |
| W1 | 1 | `/axb-technical-research` updates `specs/truth/techstack.md` (the *Interactive TUI prompt harness* row) + **ADR 0063**. |
| W2 | 2 (after W1) | `/axb-dsl-refine` updates the interface truth (`chat/presenting-the-interactive-prompt.feature` note + `chat/dsl.md` row). |

The **CLI end is carried forward to its contract owner `/axb-dsl-refine`** (no API/data/UI planner
applicable) — the round-032 pattern.

## 4. Implementation surface (the `cli` interface's contract owner's plan)

| Surface | Change |
| --- | --- |
| `tests/e2e/harness/cmd_helper.go` | `paintGate(compose, fallback)` (content-aware gate); `runExecSynced` derives the gate from the compose chunk; the marker arg becomes the compose-less fallback. |
| `tests/e2e/steps/tui_keys.go` | `tuiKeyEnter = "\r"` + `typeText` (`\n → CR`); the typed-key helpers route through it. |
| `tests/e2e/steps/step_r038.go` | The empty-submit sequence routes its typed prompt through `typeText`. |
| `tests/e2e/steps/step_t011_chat_then_keeps_typed_text.go` | The multi-line separate-rows clause (`joinedRow`). |
| `tests/e2e/harness/cmd_helper_test.go`, `tests/e2e/steps/tui_keys_test.go`, `tests/e2e/steps/step_t011_chat_then_keeps_typed_text_test.go` | The three mechanism unit pins. |

**No `internal/**` or `cmd/**` change.**

## 5. Handoff

- **Truth root**: `specs/truth/`
- **Truth-delta**: `specs/plans/093-interactive-prompt-seam-determinism/truth-delta.md`
- **Truth owners**: `axb-technical-research` (`techstack.md`) · `axb-dsl-refine` (`features/cli/**`).

## 6. Domain model (ADR 0041 — load-bearing)

**Not modelled.** The change is a **test-seam** change: no product entity, relationship, attribute,
invariant, or scenario changes. The modelled `PromptInput` (the `-i` input surface) and the *An interactive
prompt* scenario describe the product, which is unchanged; the seam that drives it is test support. No
modelled behaviour changes, so `docs/domain-model/**` is **not** updated (the ADR 0041 same-PR rule does
not engage). `make modelith-check` must stay green (no drift introduced).

## 7. Records

- **ADR 0063** (`docs/decisions/0063-interactive-prompt-e2e-seam-determinism.md` + index).
- `specs/truth/techstack.md` (the *Interactive TUI prompt harness* row).
- `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` + `chat/dsl.md`.
