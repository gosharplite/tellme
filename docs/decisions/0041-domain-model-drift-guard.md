# ADR 0041 — Domain-model drift guard (advisory): the model is load-bearing

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** operator request (a docs/quality pass; no anchor issue) · **ADR 0030** (the model + the YAML↔MD `modelith-check` gate; its **D5**/**RF-060-3** "the reference's advisory code↔model gates are not adopted" is **qualified here**) · **ADR 0011 D10** (layer gate; unchanged) · reference `tell-me-go` `scripts/modelith-drift.sh` / `modelith-layers.sh` (not adopted — see D3) · the curation rule (the `RF-063-10`/`RF-068-1` "noisy advisory surface becomes a muse" lesson).

## Context

Round 060 (ADR 0030) gave tellme its canonical domain model (`docs/domain-model/*.modelith.{yaml,md}`) and a **YAML↔MD generation** gate (`make modelith-check`, a zero-tolerance `make verify` member). That gate protects the model from *internal* drift only — it does not know anything about the code.

Between rounds 069–070 the model **rotted silently**: it kept describing a `Skill → Context` automatic-injection relationship (false — the shipped surface is on-demand, round 033 Q1) and an "attached" media channel (superseded — the media is returned in-band, round 070 / ADR 0040). Each round's truth owner edited `specs/truth/**` but **not** the model, and no gate noticed. The rot was found by a hand audit (2026-09-20), not by the pipeline.

The operator then asked whether any quality gate was *coverage-like* and should be removed, and (after the audit) directed the highest-value fix: keep the model honest.

## Decision

**D1 — the model is load-bearing.** It remains *descriptive docs, subordinate to truth* (ADR 0030 §D4 — on conflict, truth wins), **but it is maintained**: **a round that changes MODELLED behaviour updates `docs/domain-model/**` in the same PR**, or records in its plan package why the change is not modelled. This is a **process** commitment; where it is carried is `docs/domain-model/README.md` → *Lifecycle* + *Drift guard*.

**D2 — an advisory `make modelith-drift` guard (the stale-entry direction).** A POSIX-only script (`scripts/modelith-drift.sh`) checks, for every modeled **entity**, **enum**, and **glossary** term, whether **any code anchor** — its own name, a backticked identifier in its definition, or an enum value — still appears in the production Go sources (comments included; `_test.go` excluded). Zero anchors ⇒ a warning. It is **advisory**: it **never fails** and is **not** a `make verify` member; scope is the **code model** only (`tellme.modelith.yaml`). It is deliberately **lenient** (one live anchor clears an entry) and keeps a tiny hand-exception list for a term modelling a **deliberate absence** (`NoSecurityLayer`). `docs/domain-model/README.md` is the authority for the guard; the `Makefile` target is its entry point.

**D3 — the reference's NAME-DIFF advisory gates stay NOT adopted, on measurement.** `tell-me-go` ships `modelith-drift` (new exported Go identifier with no model entry) and `modelith-layers` (modeled entity not in `internal/domain/`). A direct port was **measured** on this repo and **rejected**:
- comparing the model's names against exported types under `internal/domain/**` flagged **~45 of 57 types (~79 % false positives)** — ports (`Store`, `Sink`, `Reader`, `Source`, `Prompter`, `Loop`, `Lines`, `Indicator`, …) and value types (`Request`, `Response`, `Result`, `Step`, `Entry`, …) vastly outnumber *modeled concepts*;
- the reverse direction flagged **12 legitimate entities** (named differently in code, logical enums, behavioural roles).
Shipping it would recreate the **retired** "noisy advisory surface becomes a permanent muse" failure (the `RF-063-10`/`RF-068-1` class). The D2 guard is the **precise** subset that survives measurement (zero findings on the tree; it catches a synthetic stale entry). A faithful name-diff gate would need a hand-curated entity↔symbol manifest — a separate, unwarranted decision today (**RF-041-1**).

**D4 — semantic rot has no mechanical carrier (an honest limitation).** Neither `modelith-check` nor the D2 guard catches the 069–070 class: the model describing behaviour the code does **not** have (an *injection* relationship that no longer exists; an *attached* media channel that became in-band). There is no mechanical anchor for "this relationship is false". That class is guarded by the **D1 round-time rule** and review. Keeping the model load-bearing is therefore a **process** commitment, not a gate.

## Consequences

- `make verify` is **unchanged** (no new member; `modelith-check` stays the only model gate). `make modelith-drift` is available for a developer or closeout to run on demand.
- **No new dependency** — the guard is bash + `grep`/`awk`/`sed`, POSIX-only, hermetic (reads the tree only).
- The guard is an **aid**, not a gate: a zero finding means "no modeled term lost its code anchor", not "the model is correct" (D4).
- ADR 0030 §D5 / RF-060-3 remain true as written (**the reference's** gates are not adopted); they are **qualified** — tellme now ships its **own**, narrower advisory drift guard. Recorded on ADR 0030's `Status` line and its RF-060-3 bullet.

## §Forward

- **RF-041-1** — a hand-curated **entity↔symbol manifest** would enable a reliable name-diff direction (a new domain concept with no model entry); deferred as a separate decision (it introduces a maintained artifact of its own).
- **RF-041-2** — the guard's leniency: a modeled term kept alive only by a **comment** or an unrelated token in an anchor set is not flagged; tightening it trades false negatives for false positives.
- **RF-041-3** — the guard does not read `specs/truth/**` (a truth-only change with no model update is out of scope; truth is the authority, not the model).

## References

- `docs/domain-model/README.md` (→ *Drift guard*, *Lifecycle*) — the guard's authority
- `scripts/modelith-drift.sh` · `Makefile` (`modelith-drift`, `MODELITH_CODE_MODEL`)
- **ADR 0030** (the model + `modelith-check`) · **ADR 0032 D7a** / **ADR 0040** (the media-channel lineage) · **ADR 0033** (skills on-demand, round 033 Q1)
- The curation rule (`SESSION-BOOTSTRAP.md` Agent Rule 11; `SESSION-CLOSEOUT.md` Rule 17) — the anti-muse premise of D3
