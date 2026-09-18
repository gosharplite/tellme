# Truth Delta: 053-offline-session-config-and-turns-flag

**Plan Package**: `specs/plans/053-offline-session-config-and-turns-flag`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1 CLOSED** — **Q1 → (C1)** (`tellme` writes its own `turns.log` — its rendered chrome — and `-t` prints it); **Q2 → (A)** (an explicit `-c` that cannot be honoured fails; an absent default stays tolerant). No truth has been written yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **CLI flag parsing** row | **Round 053 (closes [#103](https://github.com/gosharplite/tellme/issues/103); ADR 0022)**: the flag set gains `-t`/`--turns`. | `spec.md` US2 / FR-005 |
| MODIFY | `specs/truth/techstack.md` — **Session lifecycle flags** row | **Round 053 (ADR 0022)**: the offline session commands (`-l`, prompt-less `--new`, `-t`) resolve the mode as `TELL_ME_MODE` → else the **`-c` config's `MODE`** → else the default → else `"butler"` (previously ignored `-c`); the `-c` read is mode-only (stays offline); an **explicit** `-c` that cannot be read **fails**; `-t` prints `turns.log`. | `spec.md` US1 / FR-001…FR-003, FR-009; `research.md` D1–D4 |
| MODIFY | `specs/truth/techstack.md` — **Turn log (`turns.log`)** row | **Round 053 (Q1 → (C1); ADR 0022)**: a per-session plain-text `output/<mode>/turns.log` holding **the subset of diagnostic lines routed through the chrome sink** (the per-call renderer's frame + tail); the input-capture ack, the `-i` echo, error phrases, the spinner, and the `[Tool …]` block are **not** persisted; written by an injected port (best-effort); read by `-t` (streaming); archived by `--new`. | `spec.md` US2 / FR-004, FR-006, FR-007; `research.md` D5; fold F-53-3 |
| MODIFY | `specs/truth/techstack.md` — **Prompt input** row (fold **F-53-5**) | The dispatch precedence now reads `--version` → `-d` → `-l` → `-t` → `--tool-usage`, and `-t` joins the flags that never read stdin — bringing the row into agreement with the Session-lifecycle-flags row (`truth-current`). | `spec.md` FR-005; PR #118 review F-53-5 |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the round adds **no** Makefile target — the gates ride the existing `verify` aggregate. | `spec.md` SC-004 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and no OpenAPI/HTTP surface. The round authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/data/data-model.dbml` — `turns_log_line` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry`: added the round-053 per-session **turn log** artifact — a plain-text line per rendered chrome line at `output/<mode>/turns.log`, read by `-t`, archived by `--new`. | `spec.md` US2 / FR-004, FR-006, FR-007; Q1 → (C1) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | **Round 053 (ADR 0022)**: added the `The session listed is the one the named configuration belongs to` Rule (two Examples: `a.yaml`/`b.yaml` differential) and the `A named configuration that cannot be read is refused` Rule. | `spec.md` US1; `plan.md` (the CLI end's contract owner) |
| ADD | `specs/truth/features/cli/history/reviewing-the-turn-log.feature` | **Round 053**: the `-t` turn-log read — the `-c`-resolved session's turn log (and the empty-log tolerance). | `spec.md` US2; `plan.md` |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | **Round 053**: new Given rows (config+session seed; config+turn-log seed), When rows (`-l -c`; `-t -c`), Then rows (`tellme lists the assistant message "{answer}"`; `tellme prints exactly the turn log line "{content}"`; `tellme prints nothing`). | `spec.md` US1/US2 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0022-offline-session-config-and-turns-log.md` (+ the `docs/decisions/README.md` index row) | Records D1–D5 (the mode-resolution precedence + explicit-`-c` failure; the widened `resolveWorkspace`; the `-t` flag; the `turns.log` artifact/writer) + a §Forward (RF-53-1…4). | `spec.md` FR-010; `research.md` D6/D7 |
