# Technical research — round 089 `089-context-control-position`

**Topic**: reconcile the stale `-b`/`--retry` clause in `specs/truth/techstack.md` and record tellme's
**context-control position** as a durable decision — while changing **no** behaviour.

**Owner**: `axb-technical-research` (no new technology; no new dependency).

---

## D1 — The defect: a superseded present-tense claim in the truth tree

`specs/truth/techstack.md` (the *History summarisation* bullet, `## Not Introduced Yet`) reads, verbatim:

> *"…so do history **pinning**, the undo/retry flags (`-b`/`--retry`), and `SafePath`/consent (out of
> tellme scope, not merely deferred)"*

But **`-b`/`--back` shipped in round 081 (ADR 0053)** — the *Session lifecycle flags* row documents
it as delivered (default-1, `-l`/`-b` compose, `-b "p"` rolls back then runs). So the bullet
**conflates** two different things under one label:

| Item | True status today |
| --- | --- |
| `-b` / `--back` | **Delivered** (round 081 / ADR 0053; refined by 082/086) — **not** an exclusion |
| `--retry` (roll back the last user message + resend) | **Not implemented** — a genuine non-introduction |

The truth tree is the *single source of current system behaviour* (`truth-current`); a row that reads
as excluding a **shipped** feature is a **self-contradiction inside the truth tree** — the same class
as round 088's **F-088-1** (a `techstack.md` row contradicting another). Grounding: verified on `dev`
@ `53753a0` — the clause is present-tense; `grep` finds **no** `--retry` in `internal/`, `cmd/`,
`tests/`, or `specs/truth/`; and every **other** `-b`/`--retry` mention in the repo is either inside a
**frozen** `specs/plans/NNN-*/` package (correctly historical) or accurate (the round-078 *Provider
request retry* row records the reference's flag as **not adopted**).

## D2 — The fix: split the clause; keep each surviving exclusion true

Options:
- **(a) Delete the whole clause** — **rejected**: `--retry`, pinning, pruning, and `SafePath`/consent
  are *genuinely* still exclusions; deleting them would be a new truth defect.
- **(b) Just delete `-b`** — **rejected (thin)**: the bullet would still conflate *"the undo/retry
  flags"* conceptually with the delivered `-b`; the reader would not learn that `-b` **is** shipped
  context control.
- **(c) Split the clause** — **adopted**: `-b`/`--back` is removed from the exclusion list and given a
  forward pointer to the *Session lifecycle flags* row / ADR 0053; `--retry` is kept and stated
  accurately as a real non-introduction; the remaining exclusions are untouched.

The edit is a **status move** (`-b`: exclusion → delivered), not a behaviour claim: it changes **no**
delivered capability's description, and it adds **no** new capability.

## D3 — Option (b): the decision needs a durable home; an ADR is owed

The **position** is not merely "summarisation is unbuilt"; it is a **stance**: tellme *chooses*
operator-manual context control over the reference's (a) `summarize_history` tool and (b) automatic
self-healing summarisation. Nothing durable records **why**, nor frames it as a **deliberate
divergence**. Per `docs/decisions/README.md` (*"a decision that other artifacts or future rounds
depend on and must be able to cite"*), this warrants an **ADR** — **0060** (0059 was round 088).

The ADR records, precisely:
- **The decision**: tellme does **not** port `summarize_history` (on-demand **or** automatic);
  supported context control is the **operator-manual** `-l` (inspect) + `-b` (roll back) pair;
  `MAX_HISTORY_TOKENS` is **displayed** and caps the tool-result bound but **prunes nothing**;
  token-budget pruning, history pinning, and `SafePath`/consent stay settled exclusions.
- **The rationale**: the reference's compression is **lossy-but-information-retaining**; tellme's
  `-l`/`-b` is **coarse and destructive** (whole-turn removal) but **deterministic, inspectable, and
  operator-controlled** — consistent with tellme's postures (no auto-magic; small surface;
  determinism; operator-in-the-loop). The **round-021 removal** of the summarisation *agent tool*
  (*tool-surface parity* — the read trio) is the precedent.
- **The revisit trigger**: revisit **only** if a measured session is shown to **routinely** exceed the
  effective budget **and** the operator's manual `-l`/`-b` control is demonstrated insufficient — a
  future round whose anchor is a live issue, not this ADR.

## D4 — Do **not** assert an equivalence claim (FR-004)

The issue's W3 warns that an *equivalence/coverage* claim ("`-l`+`-b` cover the summarisation role")
would need a **discriminating** carrier, and to **not assert it** if none can be made non-vacuous.
Decision: **do not assert equivalence.** The ADR records the **non-equivalence** (compression vs
coarse operator trim) — which is the honest reading — and cites the **existing** carrier for the only
mechanical half that already exists:

- `specs/truth/features/cli/chat/offering-the-agent-tools.feature` carries **Rule** *"a summarisation
  tool must not be offered"* (with its `chat/dsl.md` row), i.e. the "tellme offers no summarisation
  tool" claim is **already falsifiable in the E2E** — a mutant that re-offers such a tool reddens it.
  The ADR **cites** this carrier; the round adds **no** new claim that would need one.

## D5 — The witness (falsifiability)

- **W1 (the stale clause)** — a **reproducible** grep: pre-fix, `specs/truth/techstack.md` lists `-b`
  in the out-of-scope set; post-fix it does not (SC-001). Reproduced pre-fix and re-run post-fix. (A
  carried, on-demand check — like round 085's audit, it is not a `make verify` member; the round adds
  no gate, honouring the repo's `topology-audit-not-a-gate` posture.)
- **W2 (the decision is durable)** — `make verify-adr-index` (a **standing** `make verify` member)
  proves the ADR is indexed once and its number unique; the ADR file itself **is** the witness
  (quality-model `ADR`/`DecisionIndex` invariants).
- **W3** — **not used** (D4: no equivalence claim asserted).

## D6 — Scope guard

No product code, no `.feature`, no step definition, no `go.mod`/`go.sum`, no new dependency, no `make
verify` change. Frozen `specs/plans/NNN-*/**` history is **not** rewritten (`plan-package-frozen`);
`docs/domain-model/**` is **not** modelled (ADR 0041 escape hatch).
