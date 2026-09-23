# GAPS.md — witness gaps in the round process (working notes)

> **Status**: open · report-only · **not** a `specs/truth/**` artifact and **not** a process gate.
> **Origin**: a `tellme`-side investigation while executing `SESSION-BOOTSTRAP.md` (2026-09-23).
> **Scope**: **one class only** — *a claim with no tripwire* (a normative clause that nothing can
> falsify). The separate *deferral-curation* problem (a forward item parked on a bootstrap-read
> surface becoming a recurring "muse") is **out of scope** here; see `ADR §Forward` curation rules.
> **Filed upstream**: [`aixbdd-tmg#15`](https://github.com/gosharplite/aixbdd-tmg/issues/15).

---

## 1. The class

**Claim-without-a-tripwire** = a normative sentence (an `FR`/`NFR`/`SC`/`MUST`, or a promoted truth
row) that is asserted on a surface engineers trust, but for which **no input exists that would make
it fail**. `green ≠ falsifiable`: a suite of passing tests gives no evidence about a clause none of
them can redden.

**Discriminator (one question):**

> *Which input makes this claim false — is that input in the suite?*
> No input ⇒ no witness ⇒ the claim is unguarded.

An unwitnessed-but-authoritative sentence is worse than an absent one: it reads as binding, so future
rounds reason **from** it ("durability is handled") and any silent regression is invisible. It also
fails `truth-current` **silently**, which is the one thing an executable-spec process exists to prevent.

---

## 2. Worked example — round 081 `fsync` (the case that started this)

**The claim** (round 081, `-b` rollback; ADR 0053), stated on **six live surfaces**:

> *"Rollback is durable and atomic: the survivors are written to a temp file + `fsync` + atomic
> `rename` over `history.jsonl`; a crash mid-rollback leaves the prior `history.jsonl` intact."*

Surfaces: `specs/truth/techstack.md` (*Session history store* row) · `specs/truth/features/cli/history/dsl.md`
(module prologue, line 20) · `docs/decisions/0053-back-rollback-turns.md` (D3 + Consequences) ·
`docs/decisions/README.md` (index row) · `docs/domain-model/tellme.modelith.yaml` / `.md`
(`history-rollback-removes-complete-turns`) · + the frozen plan package.

**Claim decomposed** into independent clauses: `{temp file, fsync file, rename, fsync dir, crash-safety}`.

**First-party mutation check** (2026-09-23, current tree; each mutation applied then reverted; tree
confirmed clean):

| Mutation of the shipped `writeRaw` | Rollback suite |
| --- | --- |
| in-place truncate (no temp file / no rename) | **FAIL** — `TestFileStore_Rollback_FailureLeavesPriorHistoryIntact` reddens ✅ |
| remove `f.Sync()` (file `fsync`) | **ok — all green** ❌ |
| remove `syncDir(...)` (directory `fsync`) | **ok — all green** ❌ *(no asserted claim to falsify: the dir `fsync` is a **mechanism**, named on no asserted surface — see the corrected Finding below; the marker reads "no tripwire needed", not "open defect")* |

**Finding.** The review fold (`F-081-3`) added a failure-path pin that reddens — **but only for the
temp+rename half**. **The `fsync` predication has no falsifier** — corrected by the upstream grill
round (`aixbdd-tmg#15` Q10): the unit of obligation is the **atomic effect claim**, not the mechanism
list, and **none** of the six surfaces names the directory `fsync` (the code's `syncDir` is
best-effort, its error swallowed) — so the unwitnessed predication is **one**: the file
`fsync`-before-`rename` guarantee. `git blame` shows the mechanism shipped in round 081's single
implementation commit (`45f43e7`).

**Two distinct defects, not one:**

1. **Not witnessed** — no test reddens if the `fsync` half is removed.
2. **Misclassified / misfiled** — the guarantee is asserted as *truth* (a `techstack.md` row) although
   it is (a) a **behaviour guarantee**, which `techstack.md`'s own header says belongs to
   spec/interface truth, and (b) **unwitnessed**, so by the executable-spec premise it is not truth at
   all but a **decision** → its home is the ADR (decision + an explicit *accepted-unwitnessed*
   disclosure in `§Forward`). `ADR 0053 §Forward` has nine items and records **none** of this.

**Not the implementers' fault.** The shipped code is the textbook recipe (temp file → file `fsync` →
`rename` → dir `fsync`). `syncDir`'s error is swallowed, so dir-level durability is best-effort —
a minor over-claim, recorded, not a defect.

---

## 3. Inventory — how many similar hints exist

Method: keyword scan over `docs/session-summary/**` and `STATUS.md`; counted at *finding* and
*lesson* level (raw per-day counts conflate the positive "Witnesses (reproduced then reverted)"
records with defects, so they are not used).

| Tier | Meaning | Count |
| --- | --- | --- |
| **Finding level** (a review finding id + a witness-gap phrase) | distinct defect records | **11** |
| **Lesson level** (the explicit "recurrence meme" line in process notes) | self-diagnoses | **7–8** (rounds 075, 076, 078, 079, 080×2, 081) |
| **Review chains** returning `APPROVE WITH REQUIRED FOLDS` | review passes finding ≥1 required fold | **43** |
| **Day density** | days with ≥1 witness-class mention | **12 of 14** (`2026/09/10–23`) |
| `STATUS.md` | class mentions | 2 (hygiene pass moved narrative to archives) |

**Finding-level instances (the concrete list):**

```
075  F-1   D6 overclaimed the witness set + FR-005 had no carrier
075  N-1   the DSL 不該發生 clause had no carrier
073  F-1   the rendered-model-body E2E carrier was unfalsifiable
078  F-1   the pairing/no-usage halves unwitnessed
078  F-2   the accounting MUST unwitnessed
078  F-3   the reason×unknown ordering unwitnessed
079  F-2   the persisted-`calls` claim had no E2E carrier
080  F-080-1  the hole-#2 fix had no witness (mutation left the suite green)
080  F-080-2  the persisted-`calls` claim had no E2E carrier
081  F-081-1  the durability/atomicity claim had no discriminating witness
081  F-081-2  the `-l`×`-b` compose was unwitnessed
081  F-081-3  the decode-failure claim had no carrier
082  F-082-2  SC-002 had no carrier
082  F-1   D6 overclaimed the witness set + FR-005 had no carrier
```

**Lesson-level lines** — the same rule rewritten each round:

- 075 — *"A witness that cannot fail is not a witness."*
- 076 — *"A claimed invariant with no witness is a round-trip waiting to happen."*
- 078 — *"A claim must have a witness that can redden."*
- 079 — *"A claimed E2E assertion must have a carrier that can redden."*
- 080 — *"A fix that cannot fail under any test is not a fix."*
- 080 — *"A claimed E2E assertion must have a carrier."*
- 081 — *"A claimed invariant with no discriminating witness is a defect."*

**Reading.** ~**one witness-class finding per round, sustained** — every round 075→082 produced at
least one, and the last five produced a lesson-level statement in near-identical words. The process
**re-derives** the rule instead of **enforcing** it. And because every entry above is a *review*
finding, the density is a **floor**, not a total: it counts only what review caught — the `fsync`
clause is exactly the member of the class that review **missed** and that surfaced only under mutation.

---

## 4. Root cause — where the pipeline has no carrier

Audit of `aixbdd-tmg/skills/` (16 skills). The pipeline converts **observable** behaviour into
executable truth (Gherkin `Rule`/`Example` + `dsl.md` row + stepdef) and that path is genuinely
self-policing. **Non-observable** claims have no carrier anywhere:

| Skill | Gap |
| --- | --- |
| `axb-specify` | Origin of every claim. `templates/spec.template.md` has a *story-level* `獨立驗證方式`; **no per-`FR`/`NFR` verification slot**. |
| `axb-tasks` | Owns the only forward-traceability check (`rules/Pre-Delivery覆蓋檢驗與孤立產物盤點判準.md`, ADR 0004). Sweeps **consumption** (*"is every truth row / decided Decision referenced by a task?"*) — **never falsifiability**. |
| `axb-implement` | `rules/完成定義-驗證與回寫判準.md` Rule 1: done = *implemented + verified*; "verified" means **the test passes**, never **the test can fail**. |
| `axb-bdd` | Holds the repo's **only** falsifiability rule (`rules/red-失敗訊號建立判準.md`), but scoped to interface **slices**; a unit-tier witness has no rule to cite. |
| `axb-gherkin-and-dsl` / `axb-dsl-refine` | Own `acceptance-coverage` for **observable** acceptance only (`STANDARDS.md` §7). An unobservable NFR is **silently dropped** at conversion. |
| `axb-technical-research` | Owns `techstack.md`, whose header says *"behaviour contracts are spec/interface truth"*; nothing in `rules/techstack-artifact最小必要資訊判準.md` forbids a **behaviour guarantee** from being filed as a techstack row. |
| `axb-truth-delta` | Records truth changes with a fixed four-column table (`動作`/`Truth 規格`/`改動摘要`/`原因`) — **no witness column**, so the claim→witness map is not derivable from truth. |

**Net.** The repo already uses a *falsifier* concept — ADR 0001 / domain-model invariant
`acceptance-atomic-rules` ("if a result can fail while another still holds … split") — but only at the
**Rule-atomicity** level. There is **no equivalent at the claim level**, and no carrier at all for
claims that cannot be expressed as observable Gherkin.

---

## 5. Proposed fix — a claim→witness obligation

> **No normative clause may be promoted to truth without a witness.** Every `FR`/`NFR`/`SC` clause is
> classified **observable** (carried by a falsifiable test) or **unobservable** (must name a witness
> tier — a unit pin / fault-injection seam — **or** be recorded as **accepted-unwitnessed** in the ADR
> `§Forward`). A clause that is neither witnessed nor recorded is a **decision**, and belongs in the
> ADR — it must not be asserted as truth.

**Minimal cut (changes three skills + one scoping rule):**

1. **`axb-specify`** — per-clause classification: every `FR`/`NFR` carries a verification intent
   (`observable → acceptance Rule` / `unobservable → witness tier` / `accepted-unwitnessed`); the
   completeness self-check asserts no clause is unclassified.
2. **`axb-tasks`** — extend the Orphan Coverage Sweep from *artifacts* to **claims**: every clause
   maps to a task delivering a **falsifier**, or carries an explicit accept-record; emit a
   claim→witness ledger in `tasks.md`.
3. **`axb-implement`** — a *witness* task is done only when the falsifier is shown to **redden**
   (revert-the-fix / mutation), or the absence is recorded. Converts "green" into "guarded".
4. **`axb-technical-research`** — a **behaviour guarantee/invariant** MUST NOT be asserted as a
   `techstack.md` row; route it to executable truth or an ADR decision.

*(Enforcement ergonomics, optional: `axb-bdd` generalise the red-signal rule to unit pins;
`axb-dsl-refine`/`axb-gherkin-and-dsl` route unobservable NFRs explicitly; `axb-truth-delta` add an
optional `Witness` column; `axb-constitution` land the invariant in the shared constitution.)*

**Not** mutation testing, **not** a coverage threshold, **not** a CI gate (prose can't be parsed —
same scope decision as ADR 0004: a `MUST` rule + template). Full rationale + alternatives:
[`aixbdd-tmg#15`](https://github.com/gosharplite/aixbdd-tmg/issues/15).

---

## 6. Honest limits

- **Coverage is unknown.** This file enumerates the *caught* instances from the summaries. No scan has
  been run over the live `specs/truth/**` for unguarded prose claims — and truth has **no witness
  field** to enumerate against (`grep` finds only ad-hoc "… is a unit pin (`Test…`)" mentions: 13 in
  `techstack.md`, 0 in `data-model.dbml`).
- **Per-clause discrimination is the method.** A claim is usually several clauses glued by "and";
  the gap lives in the clause no test feeds. Reading the prose cannot reveal this — only mutation.
- **Not every clause can or should be witnessed.** The proposal does not require a witness for every
  claim; it requires the **choice** to be *visible and recorded*, not hidden as prose-truth.
- **These findings are the reviewer's class, not the author's.** Each was found *after* the fact
  (review-fold), never at authoring — the argument for moving the check to authoring time.

## 7. Status

- **Closed instance — round 083 (2026-09-23; ADR 0055):** the multi-media tool round (`aixbdd-tmg#15` class in the `tellme` product) was an *unguarded normative clause* — ADR 0032 D7's round-scoped media placement had **no** falsifier (the only wire-order witness, `toolExchangeChronologyOK`, asserted precedence, never contiguity, and its fixtures used no media tool). Round 083 fixed the loop's placement and added the missing tripwire — a loop-tier 3-media order pin + the shared E2E contiguity witness carried by the multi-image fixture — both of which **redden** under the pre-fix ordering (measured W1: 1 loop pin + 1 E2E Example). This is the class **caught at authoring/review** (a round self-diagnosed its own tripwire gap), the intended remedy for this file.


- **Filed**: `aixbdd-tmg#15` (this proposal) — **closed (completed)** by [`aixbdd-tmg#16`](https://github.com/gosharplite/aixbdd-tmg/pull/16) (merged 2026-09-23; **ADR 0006**), which lands the claim→witness obligation across the pipeline (`axb-specify` verification intent + `EC-nnn` · `axb-tasks` `[WITNESS]` lane + Claim→Witness ledger · `axb-implement` three-gate DoD · `axb-bdd` two-authority split · `axb-technical-research` Rule 6 · the `dsl.md` prose boundary · `axb-constitution` Rule 4). The `tellme` skills are synced with it.
- **Closed instance — round 084 (2026-09-23; ADR 0056):** the `fsync` gap is **fixed** as **option (a)** of the issue — a store-level **`durableFS` seam** (`Sync`/`Rename`/`SyncDir`, `os`-defaulted) binds the durability clause to a **mechanism-seam pin** (`TestFileStore_Rollback_SyncsTempFileBeforeRename`) that **reddens** when the file `fsync` is removed; the asserted **effect is calibrated** (temp file + `fsync`-before-rename + atomic rename; no power-loss claim); the best-effort **directory `fsync`** is recorded as an **accepted-unwitnessed** limit (ADR 0056 §Forward **RF-084-1**); the misfiled guarantee leaves `techstack.md` (Rule 6) and the `dsl.md` prologue (truth-prose boundary), and the domain-model invariant is calibrated. It is a **fresh round** — no re-open of the frozen 081 package. **Closes [#169](https://github.com/gosharplite/tellme/issues/169)**; the rollback's `fsync` clause is **no longer** an unwitnessed open defect. *(Scope: the **rollback** clause — `Append`/`Archive` still call `f.Sync()` directly and are out of scope; ADR 0056 §Forward **RF-084-4**.)*
