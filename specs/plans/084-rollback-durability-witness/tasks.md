# Tasks — 084-rollback-durability-witness

**Plan Package**: `specs/plans/084-rollback-durability-witness`
**Core Inputs**: [`spec.md`](spec.md) · [`plan.md`](plan.md) · [`research.md`](research.md) · [`truth-delta.md`](truth-delta.md) · `specs/truth/techstack.md` (*Session history store*) · `specs/truth/features/cli/history/dsl.md` · `docs/decisions/0056-rollback-durability-witness.md`

> One task per unit of work; a task is `[X]` only after its verification passes. The round changes **no** observable behaviour (spec I-1), so there is **no** Test-Alignment / Feature phase — the deliverable is a **witness** (upstream ADR 0006). The `[WITNESS]` task is done only when its falsifier is shown to **redden** (revert-and-re-green).

## Phase 1 — Setup & Foundational

- [X] **T001** — `internal/infrastructure/history/file_store.go`: add the unexported **`durableFS`** seam (`Sync(*os.File) error` · `Rename(old, new string) error` · `SyncDir(string) error`) with an `os`-backed default (`osDurableFS`) and a nil-guarded accessor; add the `fs durableFS` field to `fileStore`; `NewFileStore` sets the default; route **`writeRaw`** through `s.fs.Sync` → `s.fs.Rename` → `_ = s.fs.SyncDir` (replacing the direct `f.Sync()`/`os.Rename`/the free `syncDir`). **No behaviour change**; the public `history.Store` surface is unchanged. (D2; spec FR-001/NFR-002)

## Phase 2 — Witness Pins (Red → Green) — Phase 4W

**Goal**: bind the rollback durability clause to a mechanism-seam witness that **reddens** when the file `fsync` is removed (ADR 0006; spec I-4/NFR-001).

**Shared Must Read**: `spec.md` -> US1/NFR-001/EC-001/EC-002/EC-004 · `research.md` -> D2/D5 · `internal/infrastructure/history/file_store.go` (the seam) · `internal/infrastructure/history/file_store_rollback_test.go` (the round-081 pins)

**Boundary**: test-only; no product behaviour change beyond T001; the recording fake is in-package.

- [X] **T002** `[WITNESS]` — `internal/infrastructure/history/file_store_sync_test.go` (new): the recording `durableFS` fake + **`TestFileStore_Rollback_SyncsTempFileBeforeRename`** — assert the ordered events `sync(*.tmp)` → `rename(*.tmp → history.jsonl)` → `syncdir`. **Falsifier**: remove the `s.fs.Sync(f)` call ⇒ this pin reddens. `Dependencies: T001` · `Test Scope: internal/infrastructure/history` · `Target: file_store.go writeRaw`
- [X] **T003** `[WITNESS]` — `file_store_sync_test.go`: **`TestFileStore_Rollback_SyncErrorLeavesPriorHistoryIntact`** (EC-001) — the fake's `Sync` errors ⇒ `Rollback` returns the error, no rename occurs, the prior `history.jsonl` is byte-intact, and no temp file remains. `Dependencies: T001` · `Falsifier`: ignore the sync error ⇒ reddens.
- [X] **T004** `[WITNESS]` — `file_store_sync_test.go`: **`TestFileStore_Rollback_DirSyncErrorDoesNotFail`** (EC-002) — the fake's `SyncDir` errors ⇒ `Rollback` still succeeds (best-effort). `Dependencies: T001` · `Falsifier`: propagate the dir-sync error ⇒ reddens.
- [X] **T005** `[WITNESS]` — `file_store_sync_test.go`: **`TestFileStore_Rollback_PreservesSurvivorBytesOnToolWrittenPath`** (EC-004) — append three entries **through the store**, snapshot the file, roll back one, assert the surviving bytes are **byte-identical** to the first two original lines. `Dependencies: T001` · `Falsifier`: re-marshal the survivors instead of the raw copy ⇒ reddens.
- [X] **T006** — run the round-084 unit pins **+** the existing round-081 rollback suite — **Green**.

## Phase 3 — Truth & Records

- [X] **T007** — `specs/truth/techstack.md` — *Session history store* row: remove the behaviour guarantee; keep the mechanism (Rule 6) *(done at `/axb-technical-research`)*.
- [X] **T008** — `specs/truth/features/cli/history/dsl.md` — the `-b` prologue: drop the durability mechanism clause + the round-084 note *(done at `/axb-dsl-refine`)*.
- [X] **T009** — **ADR 0056** (`docs/decisions/0056-rollback-durability-witness.md` + index row) **witnesses ADR 0053** (its `Status` gains the pointer) *(done at `/axb-technical-research`)*.
- [X] **T010** — `docs/domain-model/tellme.modelith.{yaml,md}` — the `history-rollback-removes-complete-turns` invariant calibrated + re-rendered *(done at `/axb-technical-research`)*.
- [X] **T011** — `GAPS.md` — §2 arithmetic corrected to the grill's claim-unit reading; §7 records issue #169 closed by round 084.

## Phase 4 — Gates (DoD)

- [X] **T012** — `gofmt -l .` + `goimports -l .` clean; `go vet ./...` clean.
- [X] **T013** — `make verify` **OK** (layer gate 0 · `modelith-check` ×3 no drift · `verify-fmt` · `verify-adr-index` · lint 0 · govulncheck clean · cross-compile).
- [X] **T014** — `go test -count=1 ./...` **green** (the E2E contract count **unchanged** — no new Gherkin); `go.mod`/`go.sum` unchanged.
- [X] **T015** — **Witness DoD**: remove the `s.fs.Sync(f)` call in `writeRaw` ⇒ **T002 reddens** (attributed); **revert** ⇒ the suite is green again. Recorded in `research.md`/`tasks.md`.

## Pre-Delivery 盤點與覆蓋對照

### 1. Pre-Delivery Orphan Coverage Sweep (孤立產物盤點)
- `truth-delta.md` 非 NOOP 項目：`docs/decisions/0056-*` + ADR 0053 Status + `techstack.md` *Session history store* + `docs/domain-model/**` + `history/dsl.md` prologue — all carried by T007–T010.
- `research.md` 已拍板 Decisions D1–D9：D2 (seam) → T001; D5 (witness pins) → T002–T006; D6 (records) → T009/T010; D8 (scope/gates) → T012–T014.
- `specs/truth/techstack.md` 異動章節：the *Session history store* row → T007.

### 2. Claim→Witness 盤點對照表 (Claim→Witness Ledger)

| Claim ID | 來源規格 / Truth 錨點 | 宣稱效果（Atomic Effect Claim） | 見證型態 | 綁定 Task / 決策記錄 | 鑑別性反證 | 狀態 |
|---|---|---|---|---|---|---|
| **CLM-084-1** | `spec.md` FR-001/NFR-001 · `techstack.md` *Session history store* · domain-model invariant | the rollback's **file `fsync` runs before the atomic `rename`** | `[WITNESS]` | T002 / ADR 0056 D2/D5 | remove `s.fs.Sync(f)` ⇒ T002 reddens | **witnessed** |
| **CLM-084-2** | `spec.md` FR-005/EC-002 · ADR 0056 §Forward | the **directory `fsync` is best-effort** (its error does not fail the rollback) | `[WITNESS]` (the best-effort property) + `accepted-unwitnessed` (the dir-durability *effect*) | T004 / ADR 0056 §Forward **RF-084-1** | propagate the dir-sync error ⇒ T004 reddens; the dir-durability effect itself is **accepted-unwitnessed** | witnessed (best-effort) / accepted-unwitnessed |
| **CLM-084-3** | `spec.md` EC-004 · domain-model invariant ("survivors byte-unchanged") | the survivors are copied **raw** (byte-unchanged) on the tool-written path | `[WITNESS]` | T005 / ADR 0056 D5 | re-marshal the survivors ⇒ T005 reddens | **witnessed** |
| **CLM-084-4** | `spec.md` EC-001 | a **sync error leaves the prior history intact** (no rename, temp removed) | `[WITNESS]` | T003 / ADR 0056 D5 | ignore the sync error ⇒ T003 reddens | **witnessed** |
| **CLM-084-5** | `spec.md` FR-003/FR-007 · `specs/truth/features/cli/history/**` | the observable rollback contract (clamp · no-op · archive-untouched · exit codes) | `[BDD-GREEN]` | the round-081 rollback Examples/suite (unchanged) | n/a (observed) | covered |

## Fold ledger

*(populated during the review-fold loop)*

## Falsifiability witnesses

| # | Mutant (applied then reverted) | Expected red | Measured |
| --- | --- | --- | --- |
| **W1** | remove the `s.fs.Sync(f)` call in `writeRaw` (the file-fsync mechanism) | `TestFileStore_Rollback_SyncsTempFileBeforeRename` reddens (the ordered durability events become `rename, syncdir`) | **confirmed** — `durability events = [rename:history.jsonl.tmp->history.jsonl syncdir:001], want (fsync BEFORE rename) [sync:…, rename:…, syncdir:…]`; reverted ⇒ green |
| **W2** | make the directory `fsync` mandatory (propagate `SyncDir`'s error) | `TestFileStore_Rollback_DirSyncErrorDoesNotFail` reddens | guard (green on shipped code; reds under the mutant) |

## Gate results (this round)

- `gofmt`/`goimports`/`go vet` clean · `make verify` **OK** (layer gate 0 · `modelith-check` ×3 no drift · lint 0 · govulncheck clean · cross-compile) · `go test -count=1 ./...` **green** (E2E **314 scenarios · 2354 steps — unchanged**) · `make test-race` **no data races** · `go.mod`/`go.sum` **unchanged**.
- The observable contract is unchanged (spec I-1): no Gherkin/step added or altered; the E2E count is identical to round 083.
