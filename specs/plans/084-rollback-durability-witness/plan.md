# System Analysis — 084-rollback-durability-witness

**Plan Package**: `specs/plans/084-rollback-durability-witness`
**Spec**: [`spec.md`](spec.md) · **Research**: [`research.md`](research.md) · **Truth delta**: [`truth-delta.md`](truth-delta.md)
**Anchor issue**: [#169](https://github.com/gosharplite/tellme/issues/169)

## 1. Interface inventory

| Interface | Kind | Present? | Analysis planner |
| --- | --- | --- | --- |
| CLI (terminal: `tellme -b [N]` — the rollback path; the observable contract is **unchanged**) | `cli` | **yes — the round's only end** | none (a plain line-oriented CLI) — carried to its contract owner **`/axb-dsl-refine`** (the `history` module's `dsl.md` prologue carries the durability mechanism clause this round removes) |
| API (`contracts/**`) | `backend` | no | **NOOP** — no OpenAPI/HTTP surface |
| Data (`data/**`) | — | no shape change | **NOOP** — `history.jsonl` / `history.Entry` / `history.Step` shapes unchanged |
| Web UI | `frontend` | no | **NOOP** — no web UI |

**Waves:** one wave, one interface (the CLI). No dependency ordering is needed (a single interface, and no new observable behaviour).

## 2. What the round touches (code map)

| Layer | File | Change |
| --- | --- | --- |
| adapter | `internal/infrastructure/history/file_store.go` | add the unexported `durableFS` seam (`Sync`/`Rename`/`SyncDir`, `os`-defaulted, nil-guarded); route `writeRaw` through it (D2) — production behaviour unchanged |
| adapter tests | `internal/infrastructure/history/file_store_sync_test.go` (new) | the recording `durableFS` fake + the four mechanism-seam pins (D5): sync-before-rename, sync-error aborts, dir-sync-error best-effort, survivor-bytes on the tool-written path |
| E2E | — | **no change** — the observable contract is unchanged (spec I-1), so no Gherkin/step is added or altered |
| documentation | `docs/decisions/0056-*.md` (+ index), `docs/decisions/0053-*.md` (Status), `specs/truth/techstack.md` (row), `specs/truth/features/cli/history/dsl.md` (prologue), `docs/domain-model/**` | the witness record + the calibrated/removed prose |

## 3. Truth surfaces

| Artifact | Owner | Action |
| --- | --- | --- |
| `docs/decisions/0056-rollback-durability-witness.md` (+ index row) | `/axb-technical-research` | **ADD** — witnesses + calibrates ADR 0053 |
| `docs/decisions/0053-back-rollback-turns.md` — `Status` | `/axb-technical-research` | **MODIFY** — a forward pointer to ADR 0056 (body stays history) |
| `specs/truth/techstack.md` — *Session history store* | `/axb-technical-research` | **MODIFY** — remove the behaviour guarantee; keep the mechanism (Rule 6) |
| `docs/domain-model/tellme.modelith.{yaml,md}` — `history-rollback-removes-complete-turns` | `/axb-technical-research` | **MODIFY** + render — calibrate the durability sentence (ADR 0041) |
| `specs/truth/features/cli/history/dsl.md` — the `-b` prologue | `/axb-dsl-refine` | **MODIFY** — drop the durability mechanism clause from the prologue (truth-prose boundary) |
| `specs/truth/contracts/**` | `/axb-api-plan` | **NOOP** |
| `specs/truth/data/**` | `/axb-data-plan` | **NOOP** |

## 4. Boundaries & guards

- **Layer discipline**: the change is confined to `internal/infrastructure/history` (an adapter); the seam is unexported and `os`-defaulted; no new cross-layer edge; `verify-architecture` 0 is preserved.
- **Single owner of the durability primitives**: the store adapter owns the `fsync`/`rename`/dir-`fsync` seam (D2); no CLI/deps wiring is required.
- **No observable behaviour change (spec I-1)**: `-b`/`--back` semantics, the entry shapes, the offline readers, the interactive prompt, and the live chrome are untouched; the exit-code set stays **ten**; the frozen class-phrase vocabulary is unchanged; **no new Gherkin**.
- **Stdlib-only, POSIX-only, no new dependency, hermetic**: `go.mod`/`go.sum` unchanged; the witness is an in-memory recording fake (no real crash, no pty, no network).
- **The witness discriminates (spec I-4)**: the file-`fsync`-removal mutation reddens the seam pin; the failure is attributed; the mutation is reverted and re-green.
- **The claim→witness ledger (spec I-7)**: recorded in `tasks.md` (`/axb-tasks`).
