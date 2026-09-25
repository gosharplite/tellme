# Round 089 — `089-context-control-position`

**Theme**: reconcile a **superseded present-tense claim** in the truth tree — `specs/truth/techstack.md`
still lists the undo flag **`-b`** as an out-of-scope exclusion though round 081 shipped it — **and**
record tellme's **context-control position** as a durable decision (**ADR 0060**): tellme does **not**
port the reference's `summarize_history`; its supported context control is the **operator-manual**
`-l` (inspect) + `-b` (roll back) pair, with token-budget pruning / history pinning / `SafePath`-consent
remaining settled exclusions. **Truth + record only** — no product code, no behaviour change.

**Anchor issue**: [#184](https://github.com/gosharplite/tellme/issues/184). **DoD = close it.**

---

## 1. Why this round

tellme's truth tree (`specs/truth/**`) is the **single source of current system behaviour**
(`truth-current`): it must describe the system **as it is**, never a prior round's view. Two coupled
items on the **context-control** surface break or leave a gap in that contract:

1. **(a) A stale truth claim.** `specs/truth/techstack.md` (the *History summarisation* bullet under
   *Not Introduced Yet*) asserts, present-tense, that the **undo/retry flags (`-b`/`--retry`)** are
   *"out of tellme scope, not merely deferred"*. But **`-b`/`--back` shipped in round 081
   (ADR 0053)** — the *Session lifecycle flags* truth row documents it as delivered. The clause
   therefore mixes **one delivered capability (`-b`)** with **one genuine non-introduction
   (`--retry`)** and reads as a false exclusion — the same **superseded present-tense claim** class as
   round 088's **F-088-1** (a self-contradiction inside one truth surface).
2. **(b) An unrecorded position.** tellme has **no history summarisation** — round 008 introduced it
   as an on-demand agent tool, **round 021 removed it** (*tool-surface parity*) — and **no
   token-budget pruning** (a settled exclusion). Both are stated as *Not Introduced*, but tellme's
   **replacement stance** is recorded **nowhere durable**: no ADR declares that context control is
   **operator-manual** (`-l` inspect, `-b` roll back) and that this is a **deliberate divergence**
   from the reference (which *compresses* older turns — lossy but information-retaining — while
   tellme *deletes* whole turns — coarse and destructive).

The two are coupled: a round that records (b) **touches the same `techstack.md` bullet** that carries
the stale clause, so `truth-current` obliges (a) to be reconciled **in the same PR**.

## 2. The change

1. **(a) Reconcile `specs/truth/techstack.md`** — split the conflated clause in the *History
   summarisation* bullet:
   - **`-b`/`--back`** — remove from the *settled-exclusion* list (it is **delivered**, round 081 /
     ADR 0053) and add a forward pointer to the *Session lifecycle flags* row;
   - **`--retry`** — keep as a **real non-introduction**, stated accurately (roll back the last user
     message + resend; the reference ships it, tellme does not);
   - **history pinning · token-budget pruning · `SafePath`/consent** — unchanged (genuinely settled
     exclusions).
2. **(b) Record the decision — ADR 0060** (`docs/decisions/0060-context-control-position.md`): tellme
   does not port `summarize_history`; context control is operator-manual (`-l` inspect, `-b` roll
   back); token-budget pruning and history pinning remain settled exclusions; a **revisit trigger** is
   stated. Add its **index row** to `docs/decisions/README.md`.

No `.feature`, `.go`, `go.mod`/`go.sum`, or `make verify` change.

## 3. Requirements

- **FR-001** `specs/truth/techstack.md` MUST **no longer list `-b`** as an out-of-scope / settled
  exclusion; the *History summarisation* bullet references `-b`'s **delivered** status (ADR 0053 /
  the *Session lifecycle flags* row).
- **FR-002** `specs/truth/techstack.md` MUST still record **`--retry`** as a **non-introduction**,
  distinctly from `-b`.
- **FR-003** The context-control position MUST be recorded as an **ADR** (number **0060**),
  **indexed once** in `docs/decisions/README.md` (`verify-adr-index` green), with a **revisit
  trigger**.
- **FR-004** The round MUST **not** assert an **unfalsifiable equivalence/coverage claim**
  ("`-l`+`-b` *are* `summarize_history`"). The ADR records a **non-equivalence** (compression vs
  coarse operator trim) and cites the **existing** E2E carrier for the "no summarisation tool is
  offered" half (`specs/truth/features/cli/chat/offering-the-agent-tools.feature`).
- **FR-005** Each surviving exclusion in the reconciled clause (`--retry`, history pinning,
  token-budget pruning, `SafePath`/consent) MUST remain **true** of the shipped system.

## 4. Invariants

- **I-1 — No product code / no behaviour change.** The round touches only `specs/plans/**`,
  `specs/truth/techstack.md`, `docs/decisions/**` (the new ADR + the index), and the governance docs
  (`STATUS.md` / the day summary). No `.go` file, no `.feature` file, no `go.mod`/`go.sum`.
- **I-2 — `truth-current` satisfied.** After the round, no **non-frozen** truth surface carries a
  present-tense claim that places `-b` in the out-of-scope set. Frozen `specs/plans/NNN-*/**` history
  is **not** rewritten (`plan-package-frozen`).
- **I-3 — `truth-single-owner`.** The `techstack.md` edit is made by its owner
  (`/axb-technical-research`) and recorded in this package's `truth-delta.md`.
- **I-4 — E2E unchanged and green.** No `.feature`/step-definition change, so
  `go test -count=1 ./...` stays green with the same scenario/step counts.
- **I-5 — `make verify` unaffected** (no gate added, removed, or re-scoped); stdlib-only / POSIX-only
  (trivially: no code, no dependency).

## 5. Scope

**In**: the `techstack.md` clause reconciliation (`-b` out, `--retry` kept accurate); the context-control
**ADR 0060** + its index row; the ADR-0053 index back-pointer.
**Out**: any product-code or `.feature` change; introducing `summarize_history`, automatic
summarisation, token-budget pruning, or history pinning (**settled exclusions** — this round records,
it does not reintroduce); porting `--retry`; any other `Not Introduced Yet` item.

## 6. Success criteria

- **SC-001** `grep -n -- '--retry' specs/truth/techstack.md` finds **no** line that lists `-b` as an
  out-of-scope exclusion; the *History summarisation* bullet names `-b` as delivered (ADR 0053) and
  `--retry` as a non-introduction.
- **SC-002** `docs/decisions/0060-context-control-position.md` exists, its `# ADR 0060` heading is
  unique, and `make verify-adr-index` is green.
- **SC-003** `make check` (verify + test) is green; E2E scenario/step counts unchanged.
- **SC-004** No unfalsifiable equivalence claim is asserted; the ADR states the **non-equivalence**
  and cites the existing `offering-the-agent-tools.feature` carrier.

## 7. Assumptions

- **A1** The operator's instruction *"open a round, the goal is to close #184"* **grants the
  intent** the issue recorded as pending for item (b) — so the decision **content is locked** by the
  issue and no `/axb-clarify` is required (0 questions).
- **A2** The **only** non-frozen truth surface carrying the stale clause is `techstack.md` (verified:
  every other `-b`/`--retry` mention is inside a **frozen** `specs/plans/NNN-*/` package or is
  accurate — e.g. the round-078 *Provider request retry* row correctly records the reference's
  `--retry` as **not adopted**). Frozen history is left verbatim (`plan-package-frozen`).
- **A3** No `NEEDS CLARIFICATION`.
