# Truth Delta: 007-session-history-persistence

**Plan Package**: `specs/plans/007-session-history-persistence`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added a `Session history store` row (append-only JSON-Lines `history.jsonl` under the per-mode session workspace; written after the turn completes; answer text verbatim) and a `Session lifecycle flags` row (`--new` archive + `-l N` inspect, no provider request for `-l`); extended the `CLI flag parsing` row (added `--new`, `-l`) and reworded the `Prompt input` row's dispatch precedence to `--version` → `-d` → `-l` → (`--new`) prompt turn → boot. | Round-007 research Decisions 1, 3, 4, 5; Clarify Q1/Q2. |
| MODIFY | `specs/truth/techstack.md` | **Reasoning & Provider Transport**: added a `Conversation context` row (`llm.Request` prior-message field + hand-written assembly; empty history leaves the round-004 single-message request byte-identical) and extended the `Request assembly` row to send the `messages` array (resumed history + current prompt). | Round-007 research Decision 2; Clarify Q1 (auto-resume). |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: extended the `E2E runner / step definitions` row (filesystem assertions over `history.jsonl` / `history.archive.jsonl`), the `Local fake provider` row (records the request `messages` array — the resume witness), the `Pure-helper unit tests` row (history path/append/reload/archive + `-l` selection + positive-count validation), and the `No-network verification (offline paths)` row (offline set now includes `-l` and a prompt-less `--new`). | Round-007 research Decision 8; NFR-001. |
| MODIFY | `specs/truth/techstack.md` | **Not Introduced Yet**: reworded the `SQLite / history persistence` bullet to `SQLite / database-backed persistence` (history is a plain JSON-Lines file, not a database) and added a bullet for `History summarisation and token-budget pruning` (deferred to the agent-tools round) plus the out-of-scope note (pinning, `-b`/`--retry`, `SafePath`). | Round-007 research Decisions 1, 3; settled scope exclusions. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; round 007 changes the operator terminal contract and the persisted session store, not any API surface. The outbound provider request carries prior messages inside its existing `messages` array — no new external contract. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/data/data-model.dbml` | New data truth — the persisted session history: one `history_entry` table (a completed turn's `prompt` + `answer`) keyed by `(location, position)`, plus the `history_location` enum (`active` / `archived`) capturing the `--new` archive. Append-only, ordered; no stored ID/timestamp. | Round 007 introduces persisted system state (local JSON storage); Clarify Round 2 Q1 → invoke `/axb-data-plan`; `data-model-covers-all-state`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/remembering-the-conversation.feature` | New interface feature — Rule *A completed prompt turn is stored in the session history* + Rule *A prompt run carries the persisted conversation to the provider* + Rule *A prompt run with no persisted conversation carries no earlier exchange* (1 Example each). | Carries acceptance `remembering-the-conversation` Rules 1–2 (persist + auto-resume) plus the persistence facet of FR-001/FR-003. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added Then rows `tellme stored the exchange "{prompt}" and "{answer}" in the session history`, `the request carried the earlier exchange "{prompt}" and "{answer}"`, and `the request carried no earlier exchange`. | Round-007 persist + resume assertions (the fake's recorded request is the resume witness). |
| ADD | `specs/truth/features/cli/history/starting-a-fresh-session.feature` | New interface feature — Rule *A fresh session begins with no earlier conversation* + Rule *The previous conversation is retained when a fresh session starts* (1 Example each). | Carries acceptance `starting-a-fresh-conversation` Rules 1–2 (`--new` reset; archive retention). |
| ADD | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | New interface feature — Rule *The operator can list the most recent messages* + Rule *Listing a session with no persisted history lists nothing* (1 Example each). | Carries acceptance `inspecting-the-session-history` Rules 1–2 (`-l N`; empty). |
| ADD | `specs/truth/features/cli/history/dsl.md` | New module DSL — 2 When rows (`the operator starts a fresh session with "--new"`, `the operator asks tellme to list the last {count} messages`) + 5 Then rows (active-empty, archive-retains, lists-last, lists-none, no-provider-request). | Round-007 `--new` / `-l` session-lifecycle behaviour. |
| MODIFY | `specs/truth/features/cli/dsl.md` | Added one **cross-module** Given row (`the session history already holds the exchanges:` — a DataTable fixture writing `history.jsonl`), used by both `chat` and `history`. | The history fixture is shared across two modules → interface root (unique-authority). |
| NOOP | `specs/truth/features/cli/chat/answering-a-single-prompt.feature` | Checked; left unchanged. Its fresh-workspace assertions ("exactly one request", "prompt-less run makes no request") still hold under persistence; the new persist/resume behaviour lives in the new `chat/remembering-the-conversation.feature`. | Reconciled the plan's predicted MODIFY — the round-004 feature's contract is unchanged; the round-007 chat change is additive (new feature + rows). |

*Topology audit (`axb-gherkin-and-dsl`, `audit_feature_dsl_topology.py --root specs/truth/features/cli`): **PASSED** — 13 features, 6 modules, 11 root rows, 77 module rows, 344 steps, 0 errors / 0 warnings.*
