# Truth delta — round 093 `093-interactive-prompt-seam-determinism`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop` proves the
area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*Interactive TUI prompt harness* row) | Add the round-093 clause: the round-023 output-synchronized handshake's paint gate is **content-aware** (the terminal key is released only once the capture shows the **composed text**; the editor border `┌` is only the compose-less fallback); a scripted **line break** is delivered as the terminal **Enter byte (CR)**; the `keeps the typed text` assertion requires each line on its **own** editor row. Test-seam only; no product change; no sleep/retry; `go.mod`/`go.sum` unchanged. | Record the seam contract change (ADR 0063) on the techstack surface; **no technology change** (same `bubbletea`/`bubbles`/`lipgloss`, stdlib-only). |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other row changes. | No new dependency/technology. |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` (`the interactive prompt keeps the typed text "{text}"` Then row) | Strengthen the 必查: each line of `{text}` sits on its **OWN** editor row — no single row may carry two of the lines (the joined state reddens); the `來源` notes a line break is delivered as the terminal Enter byte (CR). | The acceptance Rule "The editor keeps each typed line on its own row" is carried here; the pre-093 semantics passed vacuously on a joined row (issue #191). |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` (header note + the multi-line Example comment) | Add the round-093 note: a scripted line break is the terminal Enter byte (CR); `keeps the typed text` requires each line on its own row; the scripted keys are delivered after the composed frame paints. The **Gherkin step sentences are unchanged** (the multi-line Example's Then is the strengthened carrier). | Record the seam + strengthened-assertion contract; the Example is the round's carrier. |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change (the seam is in-memory test support). | Test-seam-only round. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change (the `-i` surface itself is unchanged). | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0063-interactive-prompt-e2e-seam-determinism.md` | **ADR 0063** — the `-i` E2E seam: a content-aware paint gate + a faithful Enter byte + a discriminating assertion. | Record the seam contract (its durable home). |
| MODIFY | `docs/decisions/README.md` | The **ADR 0063** index row (`Accepted`). | `verify-adr-index` (`adr-index-consistent`). |
| MODIFY | `tests/e2e/harness/cmd_helper.go` | Content-aware paint gate (`paintGate`); the marker arg becomes the compose-less fallback. | FR-001 (root cause A). |
| MODIFY | `tests/e2e/steps/tui_keys.go`, `tests/e2e/steps/step_r038.go` | `tuiKeyEnter` (CR) + `typeText`; every typed sequence routes through it. | FR-002 (root cause B). |
| MODIFY | `tests/e2e/steps/step_t011_chat_then_keeps_typed_text.go` | The separate-rows clause (`joinedRow`) in `thenKeepsTypedText`. | FR-003 (discriminating carrier). |
| ADD | `tests/e2e/harness/cmd_helper_test.go`, `tests/e2e/steps/tui_keys_test.go`, `tests/e2e/steps/step_t011_chat_then_keeps_typed_text_test.go` | The three mechanism unit pins. | NFR-001/SC-002 (deterministic carrier). |
| NOOP | `docs/domain-model/**` | Checked — **not modelled** (ADR 0041 escape hatch). The model's `PromptInput`/`Chrome` entities and the interactive-prompt scenario are unchanged; a test-seam change touches no modelled behaviour. | A seam change touches no modelled behaviour. |
