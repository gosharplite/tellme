# ADR 0056 — Witnessing the rollback durability guarantee (a mechanism seam, effect-strength calibration, and the accepted-unwitnessed directory fsync)

- **Status:** Accepted
- **Date:** 2026-09-23
- **Deciders:** tellme owner
- **Related:** [ADR 0053](0053-back-rollback-turns.md) (the durable `Rollback` whose clause this **witnesses and calibrates**; its `Status` gains the forward pointer), [ADR 0006](https://github.com/gosharplite/aixbdd-tmg/blob/main/decisions/0006-claim-witness-obligation.md) (upstream — *no normative clause reaches truth unwitnessed*), [ADR 0041](0041-domain-model-drift-guard.md) (the load-bearing domain model), [ADR 0012](0012-deterministic-hermetic-tests.md) (determinism/hermeticity); issue [#169](https://github.com/gosharplite/tellme/issues/169); `GAPS.md` / [aixbdd-tmg#15](https://github.com/gosharplite/aixbdd-tmg/issues/15) (the claim-without-a-tripwire class); round 084 (`specs/plans/084-rollback-durability-witness`)

## Context

Round 081 (ADR 0053) shipped a **durable** `history.Store.Rollback(n)` and asserted its guarantee — *"the surviving entries are written to a temp file + `fsync` + atomic `rename` over `history.jsonl`; a crash mid-rollback leaves the prior `history.jsonl` intact"* — on **six live surfaces**: `specs/truth/techstack.md` (*Session history store*), `specs/truth/features/cli/history/dsl.md` (prologue), `docs/decisions/0053-back-rollback-turns.md` (D3 + Consequences), `docs/decisions/README.md` (index row 0053), `docs/domain-model/tellme.modelith.{yaml,md}` (`history-rollback-removes-complete-turns`), and the frozen `specs/plans/081-back-rollback-turns/**`.

The only pins on the rollback path were the failure-path pin `TestFileStore_Rollback_FailureLeavesPriorHistoryIntact` (fold `F-081-3`) and `…DoesNotRewriteSurvivorBytes` (`TD-081-2`). Measured first-party (`dev` @ `030e142`, 2026-09-23; each mutation applied then reverted):

| Mutation of `internal/infrastructure/history.fileStore.writeRaw` | Rollback suite |
| --- | --- |
| in-place truncate (no temp file / no rename) | **FAIL** — the fold's pin reddens |
| remove `f.Sync()` (file `fsync`) | **green** — no input reddens |
| remove `syncDir(...)` (directory `fsync`) | **green** — no input reddens |

So the **file-`fsync` predication** was asserted on six surfaces with **no falsifier** — the `GAPS.md` / `aixbdd-tmg#15` "claim-without-a-tripwire" class, now governed by upstream **ADR 0006** (*no normative clause may be promoted to truth without a witness; an unwitnessable claim is a decision, recorded in the ADR and not asserted as truth*). Two further defects were recorded: the guarantee was **misfiled** (a behaviour guarantee asserted as a `techstack.md` row, which that file's own header forbids — now `axb-technical-research` **Rule 6**), and it had a second, ungoverned prose home (the `history/dsl.md` prologue — now barred by the ADR 0006 truth-prose boundary). The adversarial grill round on `#15` (Q10) corrected the arithmetic: the unit of obligation is the **atomic effect claim**, and the directory `fsync` is named on **no** asserted surface — it is a *mechanism*, not an asserted claim.

## Decision

**D1 — resolve the gap by witnessing the clause, not by demoting it.** The file `fsync`-before-rename half becomes **falsifiable** (option (a) of the issue); the human-meaningful guarantee is kept **and** guarded. The **best-effort directory `fsync`** — genuinely unwitnessable as a guarantee — is recorded as **accepted-unwitnessed** (option (b) applied narrowly).

**D2 — the seam: an injectable `durableFS` on the store.** `internal/infrastructure/history/file_store.go` gains an **unexported** seam:

```go
type durableFS interface {
	Sync(f *os.File) error
	Rename(oldpath, newpath string) error
	SyncDir(dir string) error
}
```

with an `os`-backed default (`osDurableFS`) and a nil-guard accessor. `fileStore` gains one field; `NewFileStore` sets the default; **`writeRaw`** calls `s.fs.Sync(f)` → `s.fs.Rename(tmp, activePath)` → `_ = s.fs.SyncDir(dir)`. The public `history.Store` surface and production behaviour are **unchanged**. Including `Rename` in the seam lets the witness assert the **order** (`sync` before `rename`) — a claim-level witness of the whole "temp file + fsync + atomic rename" **effect**, not a per-mechanism proxy.

**D3 — effect-strength calibration.** The asserted clause is exactly the witnessed effect: *the survivors are written to a same-directory temp file, which is `fsync`ed **before** it is atomically renamed over the active file*. The stronger **power-loss durability** implication is **not** asserted (the directory `fsync` is best-effort).

**D4 — the directory `fsync` is an accepted-unwitnessed limit.** `osDurableFS.SyncDir` stays **best-effort** (behaviour unchanged; the error is ignored), and is recorded as an accepted-unwitnessed item in **§Forward** rather than asserted. A unit pin witnesses the *best-effort* property itself (a dir-fsync error does not fail the rollback).

**D5 — the witness the round adds.** The clause is **unobservable**, so it is carried by **mechanism-seam unit pins** in `internal/infrastructure/history` (a recording `durableFS` fake), per ADR 0006's tier ladder:

1. **`TestFileStore_Rollback_SyncsTempFileBeforeRename`** — the fake logs the durability events; the pin asserts the **order relation** — the `sync` (on the temp file) **precedes** the `rename` — **and** that a directory-`fsync` event is invoked (the two requirements held together — fold R-FV-084-1). **Discriminating mutations:** remove the `s.fs.Sync(f)` call ⇒ reddens; move the sync after the rename ⇒ reddens; drop the `SyncDir` call ⇒ reddens.
2. **`TestFileStore_Rollback_SyncErrorLeavesPriorHistoryIntact`** — the fake's `Sync` errors ⇒ `Rollback` returns the error, no rename occurs, the prior `history.jsonl` is intact, and no temp file remains.
3. **`TestFileStore_Rollback_DirSyncErrorDoesNotFail`** — the fake's `SyncDir` errors ⇒ `Rollback` still succeeds (best-effort), pinning D4.
4. **`TestFileStore_Rollback_PreservesSurvivorBytesOnToolWrittenPath`** — three entries appended **through the store**, one rolled back; the surviving bytes are **byte-identical**. **It is a companion guard, not a witness** (fold F-084-1): canonical tool-written lines round-trip byte-identically under a decode+`json.Marshal`, so a re-marshal mutant leaves it green. The grill-flagged "MUST leave the surviving entries byte-unchanged" sub-clause is witnessed by the **hand-filled** pin `TestFileStore_Rollback_DoesNotRewriteSurvivorBytes` (an unknown future field is lost under a re-marshal).

**D6 — truth-surface corrections.** The behaviour guarantee is removed from `techstack.md` (mechanism kept, guarantee routed to interface truth / this ADR — Rule 6); the durability mechanism clause is dropped from the `history/dsl.md` prologue (owner `/axb-dsl-refine`); the domain-model invariant `history-rollback-removes-complete-turns` is calibrated (drops the un-calibrated "crash mid-rollback" over-claim; keeps the observable clauses) and re-rendered.

**D7 — records.** ADR 0056 (+ index); ADR 0053's `Status` gains a forward pointer to this ADR (its body stays history); `techstack.md` MODIFY; the `history/dsl.md` prologue MODIFY (`/axb-dsl-refine`); `docs/domain-model/**` MODIFY (re-render); `contracts/**` + `data/**` NOOP; `GAPS.md` §2/§7 corrected.

**D8 — scope guard.** No mutation-testing framework, no coverage threshold, no CI gate; no change to `-b`/`--back` semantics, the `history.Entry`/`history.Step` shapes, the offline readers, the interactive prompt, or the live chrome; no new dependency; stdlib-only; POSIX-only; hermetic. The exit-code set stays **ten**; the frozen phrase vocabulary is unchanged.

## Consequences

- The durability guarantee is now **guarded**: removing the file `fsync` from the rollback path reddens `TestFileStore_Rollback_SyncsTempFileBeforeRename`.
- Every surviving surface asserts **exactly** the witnessed effect (`techstack.md` records the mechanism only; the domain-model invariant is calibrated); the directory `fsync` is an explicit **accepted-unwitnessed** limit (D4/§Forward), not a hidden promise.
- The observable rollback contract is **unchanged** — a reader of `-b` output cannot tell this round shipped — so the round adds no Gherkin and the E2E contract count is unchanged.
- The misfiled guarantee leaves `techstack.md` (Rule 6 compliance) and the `dsl.md` prologue (truth-prose boundary), and gains a durable home: **this ADR** (decision) + the interface/domain-model surfaces that assert the *observable* contract.
- No new dependency; no `go.mod`/`go.sum` change.

## Alternatives considered

- **Demote-only (issue option (b)) — remove the clause from every surface and record it.** Rejected: it loses a **true** guarantee and leaves no regression tripwire for the file `fsync`; ADR 0006 prefers a cheap witness when one exists (D1).
- **A seam covering only `Sync`.** Rejected: it witnesses that a sync happened, not that it preceded the rename — a weaker claim (D2).
- **Fault injection (a real injected crash point).** Rejected as disproportionate; the mechanism-seam tier is endorsed by ADR 0006's ladder and is hermetic (D5, §Forward RF-084-2).
- **Make the directory `fsync` mandatory (and witness it).** Rejected: it would fail rollbacks on filesystems without directory fsync — a behaviour change the spec forbids (D4).
- **Rely on the review-fold loop (status quo).** Rejected by the issue's evidence: review caught temp+rename and missed `fsync`; the same class recurred across rounds.

## Forward

- **RF-084-1** — the directory `fsync` stays **best-effort** and is recorded as accepted-unwitnessed; a future round may make it mandatory (and witness it) if a filesystem without directory fsync is no longer supported.
- **RF-084-2** — the mechanism-seam pin witnesses the **call order** (`fsync` before `rename`), not the kernel's on-disk durability; a true power-loss witness would need an injected crash point (the heavier fault-injection tier), not adopted.
- **RF-084-3** — the `durableFS` seam is **unexported** — an injection point compiled into production with the `os` default (not "test-only"); production behaviour is unchanged.
- **RF-084-4** — `Append`/`Archive` and the usage store still call `f.Sync()` directly (not through the seam); the claim is the **rollback** path only — widening the seam to every write is out of scope.
- **RF-084-5** — this accepted-unwitnessed record is a **disclosure, not tasking**: per the harness rules it is never re-raised without a fired trigger.
