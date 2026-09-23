# Technical Research — 084-rollback-durability-witness

**Plan Package**: `specs/plans/084-rollback-durability-witness`
**Spec**: [`spec.md`](spec.md)
**Anchor issue**: [#169](https://github.com/gosharplite/tellme/issues/169)
**Status**: complete — decisions **D1–D9** below are locked; residuals live in **ADR 0056 §Forward**.

> Decision-driven research for the round. The spec's goal is fixed by the issue + `aixbdd-tmg` **ADR 0006** (*no normative clause reaches truth unwitnessed*); this document decides the *technical* shape: the resolution (witness vs demote), the seam, the effect-strength calibration, the surfaces, and the scope guard.

---

## 1. Problem & grounding

Round 081 (ADR 0053) asserts, on **six live surfaces**, a durability/atomicity guarantee for the `-b`/`--back` rollback:

> *"durable and atomic: the surviving entries are written to a temp file + `fsync` + atomic `rename` over `history.jsonl`; a crash mid-rollback leaves the prior `history.jsonl` intact."*

Measured at the round start (`dev` @ `030e142`; each mutation applied then reverted):

| Mutation of `writeRaw` | Rollback suite |
| --- | --- |
| in-place truncate (no temp/rename) | **FAIL** — `TestFileStore_Rollback_FailureLeavesPriorHistoryIntact` reddens |
| remove `f.Sync()` (file `fsync`) | **green** — nothing reddens |
| remove `syncDir(...)` (dir `fsync`) | **green** — nothing reddens |

So the **file-`fsync` predication** has **no falsifier**. The class is the `GAPS.md` / `aixbdd-tmg#15` "claim-without-a-tripwire"; the obligation is now `aixbdd-tmg` **ADR 0006** (claim→witness). The grill round on `#15` (Q10) corrected the arithmetic: the unit of obligation is the **atomic effect claim**, and the directory `fsync` is named on **no** asserted surface — it is a *mechanism*, not an asserted claim.

Two further defects are recorded in the issue: the guarantee is **misfiled** (a behaviour guarantee asserted as a `techstack.md` row, which `techstack.md`'s own header forbids — now enforced by `axb-technical-research` **Rule 6**), and the anomaly is **also homed in the `history/dsl.md` prologue** (a second, ungoverned prose home — now barred by the ADR 0006 truth-prose boundary).

---

## 2. Decisions

### D1 — Resolve the gap by **witnessing** the clause (option (a)), not by demoting it (option (b))

The issue offered two resolutions: **(a)** make the clause executable via a mechanism seam, or **(b)** demote it out of every asserted surface into an accepted-unwitnessed record. This round adopts **(a)** — the file `fsync`-before-rename half becomes **falsifiable** — and applies **(b) narrowly** to the one genuinely-unwitnessable mechanism: the **best-effort directory `fsync`** (`syncDir`), whose error is swallowed.

**Rationale:** the guarantee is **true and valuable** (the shipped code is the textbook recipe); demoting it wholesale would trade a real guarantee for a bookkeeping record, and ADR 0006 explicitly prefers a witness when one is cheap. A store-level **seam** yields a discriminating pin at negligible cost (stdlib, hermetic, no behaviour change). The directory `fsync` is genuinely best-effort (a directory fsync is unsupported on some filesystems; the code deliberately ignores the error) — so it is *recorded*, not *asserted*, which is exactly ADR 0006's `accepted-unwitnessed` tier.

**Alternatives rejected:** *(b) demote-only* — loses a true guarantee and leaves no regression tripwire for the file `fsync`; *do nothing / rely on review* — the status quo the issue refutes (review caught temp+rename and missed `fsync`); *mutation testing* — disproportionate, and the issue's non-goal.

### D2 — The seam: an injectable `durableFS` on the store (fsync + rename + dir-fsync)

`internal/infrastructure/history/file_store.go` gains a small, unexported seam:

```go
type durableFS interface {
	Sync(f *os.File) error
	Rename(oldpath, newpath string) error
	SyncDir(dir string) error
}

type osDurableFS struct{}
func (osDurableFS) Sync(f *os.File) error        { return f.Sync() }
func (osDurableFS) Rename(o, n string) error      { return os.Rename(o, n) }
func (osDurableFS) SyncDir(dir string) error      { d, err := os.Open(dir); if err != nil { return err }; defer func() { _ = d.Close() }(); return d.Sync() }
```

`fileStore` gains one field (`fs durableFS`); `NewFileStore` sets the `os` default; `writeRaw` calls `s.fs.Sync(f)`, `s.fs.Rename(tmp, s.activePath())`, and `_ = s.fs.SyncDir(filepath.Dir(s.activePath()))`. A nil-guard accessor keeps the zero value (a bare struct literal) working with the `os` default.

**Why fsync *and* rename are in the seam:** the claim is the *order* "`fsync` **before** `rename`". A seam covering only `Sync` can witness that a sync happened, but not that it preceded the rename. Covering both lets the recording fake log an **ordered** event list (`sync → rename → syncdir`) — a claim-level witness for the whole "temp file + fsync + atomic rename" effect, not a per-mechanism proxy. `SyncDir` rides the seam so the best-effort behaviour (EC-002) is also pinnable.

**Why the store, not the composition root:** the seam is a *durability primitive of the rewrite*, owned by the adapter that performs the rewrite; no CLI/deps wiring is needed (the default is the `os` implementation, so production is unchanged — spec **NFR-002**). The seam is unexported (`durableFS`), so the public `history.Store` surface is unchanged.

### D3 — Effect-strength calibration: assert only the witnessed effect

The asserted clause becomes: *the survivors are written to a same-directory temp file, the temp file is `fsync`ed **before** it is atomically renamed over the active file* — i.e. **temp-file + fsync-before-rename + atomic rename**. The assertion of a *crash mid-rollback leaves the prior history intact* is retained **only** to the extent the mechanism justifies it (an in-place truncate is impossible ⇒ the prior file survives a crash **before** the rename); the stronger **power-loss durability** implication is **not** asserted (the directory `fsync` is best-effort).

**Rationale:** ADR 0006's effect-strength calibration (grill Q10) — do not assert a stronger effect than the witnessed mechanism supports.

### D4 — The directory `fsync` is recorded as an accepted-unwitnessed limit

`syncDir` (now `osDurableFS.SyncDir`) remains **best-effort** (behaviour unchanged): `writeRaw` ignores its error. The round records it as an explicit **accepted-unwitnessed** item in **ADR 0056 §Forward** (reason: a directory fsync is unsupported on some filesystems; the shipped code deliberately swallows the error), rather than asserting it. A unit pin witnesses the *best-effort* property itself (a dir-fsync error does not fail the rollback — EC-002).

**Alternatives rejected:** *make the dir fsync mandatory* (would fail rollbacks on filesystems without directory fsync — a behaviour change the spec forbids); *assert it as a guarantee* (the defect the issue names).

### D5 — Truth-surface correction (Rule 6 + the prose boundary)

- **`specs/truth/techstack.md` — *Session history store* row (MODIFY):** removes the **behaviour guarantee** ("…so a crash mid-rollback leaves the prior `history.jsonl` intact") and keeps the **technology** (a same-directory temp file, `File.Sync`, atomic `os.Rename`), pointing at the interface truth / ADR 0056 for the guarantee. This is required by `axb-technical-research` **Rule 6** (a behaviour guarantee MUST NOT be a `techstack.md` row).
- **`specs/truth/features/cli/history/dsl.md` prologue (MODIFY, owner `/axb-dsl-refine`):** drops the durability mechanism clause from the `-b` prologue ("The store's durable `Rollback` (temp-file + fsync + rename) is the single owner of the truncation." → "The store's `Rollback` is the single owner of the truncation."), per the ADR 0006 truth-prose boundary (a module prologue describes the *observable* contract, not an unwitnessed mechanism guarantee).
- **`docs/domain-model/tellme.modelith.{yaml,md}` — invariant `history-rollback-removes-complete-turns` (MODIFY, re-render):** the durability sentence is calibrated to the witnessed effect (temp file, `fsync` **before** the atomic rename) and drops the un-calibrated "crash mid-rollback" over-claim; the observable clauses (removes the last N complete turns atomically, clamps, MUST NOT touch the archive, surviving entries byte-unchanged) are unchanged.
- **`specs/truth/contracts/**` + `specs/truth/data/**` — NOOP** (no API surface; no persisted-shape change).

### D6 — Records: a new ADR 0056 + an ADR 0053 annotation

- **ADR 0056** (`docs/decisions/0056-rollback-durability-witness.md`) — records: the resolution (D1), the seam (D2), the effect-strength calibration (D3), the accepted-unwitnessed directory `fsync` (D4) + its `§Forward` item, the truth-surface corrections (D5), and the **claim→witness binding** (the durability clause ⇒ the seam pin `TestFileStore_Rollback_SyncsTempFileBeforeRename`). Indexed in `docs/decisions/README.md`.
- **ADR 0053** (`docs/decisions/0053-back-rollback-turns.md`) — its `Status` gains a forward pointer to ADR 0056 (the round that witnessed its durability clause and calibrated the dir-`fsync` limit). The ADR **body stays history** (never rewritten — `plan-package-frozen` ethos); only the `Status` annotation is added.
- **`GAPS.md` §7** — the open `fsync` defect is marked resolved (round 084 / ADR 0056); the §2 arithmetic corrected to the grill's claim-unit reading.

### D7 — Witness tiers: a mechanism-seam unit pin (and its mutations)

The clause is **unobservable** (spec `NFR-001`/US1): it is carried by **mechanism-seam unit pins** in `internal/infrastructure/history` (a recording `durableFS` fake), per ADR 0006's tier ladder (`effect/fault-injection → mechanism-seam → accepted-unwitnessed`). The round adds:

1. **`TestFileStore_Rollback_SyncsTempFileBeforeRename`** — logs the ordered events (`sync` on `*.tmp`, `rename` `*.tmp → history.jsonl`, `syncdir`); asserts `sync` occurs **before** `rename`. **Discriminating mutation:** remove the `s.fs.Sync(f)` call ⇒ the pin reddens.
2. **`TestFileStore_Rollback_SyncErrorLeavesPriorHistoryIntact`** (EC-001) — the fake's `Sync` errors ⇒ `Rollback` returns the error, no rename happens, the prior `history.jsonl` is intact, no temp file remains.
3. **`TestFileStore_Rollback_DirSyncErrorDoesNotFail`** (EC-002) — the fake's `SyncDir` errors ⇒ `Rollback` still succeeds (best-effort).
4. **`TestFileStore_Rollback_PreservesSurvivorBytesOnToolWrittenPath`** (EC-004) — append three entries through the store, snapshot the file, roll back one, assert the surviving bytes are **byte-identical**. **Fold F-084-1 correction:** this is a **companion guard**, not a discriminating witness — for canonical tool-written lines a decode+`json.Marshal` round-trip is byte-identical, so a re-marshal mutant leaves it green (measured). The **falsifiable** form of the "byte-unchanged" claim is the pre-existing hand-filled pin `TestFileStore_Rollback_DoesNotRewriteSurvivorBytes` (an unknown future field is lost under a re-marshal); the grill-flagged sub-clause is thereby **witnessed** by that pin, and the tool-written path is recorded as trivially true.

The **no-observable-change** is carried by the existing round-081 suite (all green, unchanged).

### D8 — Determinism, hermeticity, no new dependency

The change is a stdlib-only seam defaulting to `os`; POSIX-only; hermetic (the witness uses an in-memory recording fake — no real crash, no pty, no network); `go.mod`/`go.sum` unchanged (spec **NFR-002**). No new exit code; no new class phrase; the exit-code set stays **ten**.

### D9 — Scope guard (excludes) & residual risks

**Excluded:** a mutation-testing framework, a coverage threshold, a CI gate (spec A6); any change to `-b`/`--back` semantics, `history.Entry`/`history.Step` shapes, the offline readers, the interactive prompt, or the live chrome (spec I-1); any real crash simulation; a power-loss durability claim (D3); any edit of the frozen `specs/plans/081-back-rollback-turns/**` (spec A1). **Residual risks → ADR 0056 §Forward** (below).

---

## 3. Residual risks (homed in ADR 0056 §Forward)

- **RF-084-1** — the directory `fsync` stays **best-effort** and is recorded as accepted-unwitnessed (D4); a future round may make it mandatory (and witness it) if a filesystem without directory fsync is no longer supported.
- **RF-084-2** — the mechanism-seam pin witnesses the **call order** (`fsync` before `rename`), not the kernel's on-disk durability; a true power-loss witness would need an injected crash point (a heavier fault-injection tier, not adopted).
- **RF-084-3** — the `durableFS` seam is **unexported**; it is a test-only injection point (the production default is the `os` implementation, unchanged).
- **RF-084-4** — `Append`/`Archive`/the usage store still call `f.Sync()` directly (not through the seam); the claim is the **rollback** path only, and widening the seam to every write is out of scope.
- **RF-084-5** — the accepted-unwitnessed record lives in ADR 0056 §Forward; per the harness rules it is a **disclosure, not tasking** and is never re-raised without a fired trigger.
