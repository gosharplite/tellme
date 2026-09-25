# ADR 0060 — Context control: no `summarize_history` port; operator-manual `-l`/`-b`

- **Status:** Accepted
- **Date:** 2026-09-25
- **Deciders:** tellme owner (issue [#184](https://github.com/gosharplite/tellme/issues/184))
- **Related:** **ADR 0053** (the `-b`/`--back` rollback this decision builds on — **delivered round 081**), **ADR 0023** (the `-l` listing), **ADR 0045**/**0054**/**0057** (listing presentation), **ADR 0026** (round-close tags); the reference's `summarize_history` tool + automatic self-healing summarisation; round 008 (introduced the summarisation tool), round 021 (removed it — tool-surface parity), round 089 (`specs/plans/089-context-control-position`); `specs/truth/techstack.md` (*History summarisation* bullet; *Session lifecycle flags* row)

## Context

tellme carries a long, **append-only** session history (`history.jsonl`) and reports a payload budget
(`MAX_HISTORY_TOKENS`) against the model's window. The budget is **displayed** and (since round 024)
caps the **tool-result** bound — but it **prunes nothing**: tellme performs **no** automatic history
summarisation, **no** token-budget pruning, and **no** history pinning. The reference (`tell-me-go`)
instead ships both a `summarize_history` **agent tool** and automatic "self-healing" summarisation: it
**compresses** older turns into a summary, keeping a long session inside the window **lossily but
information-retaining**.

tellme introduced the summarisation tool in round 008, then **removed it in round 021**
(*tool-surface parity* — the agent tool surface is the reference's read trio plus the later write/shell
tools). What tellme offers for controlling context is **operator-manual**: `-l`/`--list` **inspects**
the last N persisted messages (with role headers, backward turn indices, and a per-turn tool count),
and `-b`/`--back [N]` **rolls back** the last N turns (round 081 / **ADR 0053**) — a **coarse,
destructive, whole-turn** trim, but **deterministic, inspectable, and fully operator-controlled**.

Two facts were **unrecorded**: (i) this is a **deliberate stance**, not an omission; and (ii) it is a
**recorded divergence** from the reference. Separately, `specs/truth/techstack.md` carried a
**superseded present-tense claim** — it still listed the undo/retry flags (`-b`/`--retry`) as an
out-of-scope exclusion though `-b` had shipped (the round-088 **F-088-1** class: a self-contradiction
inside the truth tree). This ADR records the stance **and** anchors the truth fix (issue #184).

## Decision

**D1 — tellme does not port `summarize_history`.** Neither as an on-demand **agent tool** nor as
**automatic** self-healing summarisation. The reference's compression capability is deliberately **not
re-created**. (The round-021 tool removal set the precedent; this ADR makes the position durable.)

**D2 — Supported context control is operator-manual: `-l` inspect + `-b` roll back.** The operator
inspects the session (`-l`) and, when the conversation should shrink, removes whole turns (`-b
[N]`). The operator may also start fresh (`--new`, which **archives** rather than prunes). These are
the **only** in-product context-control operations.

**D3 — The budget does not prune.** `MAX_HISTORY_TOKENS` is **displayed** and caps the effective
budget that drives the **tool-result** bound; it introduces **no** conversation pruning. **Token-budget
pruning** and **history pinning** remain **settled exclusions** (as does `SafePath`/consent, out of
tellme scope).

**D4 — This is a deliberate divergence, and it is *not* an equivalence claim.** tellme's `-l`/`-b`
and the reference's `summarize_history` are **not equivalent**: compression is **lossy but
information-retaining** and mid-session; `-b` is **coarse, whole-turn, and destructive** and requires a
separate invocation (and, in the `-b "p"` form, a re-run). The round records the divergence; it does
**not** claim `-l`+`-b` "cover" the summarisation role (no unfalsifiable coverage claim).

**D5 — The `techstack.md` claim is reconciled.** The *History summarisation* bullet no longer lists
`-b` as an exclusion (`-b` is **delivered** — ADR 0053; forward pointer to the *Session lifecycle
flags* row) and keeps `--retry` as a **genuine non-introduction** (roll back the last user message +
resend — the reference ships it, tellme does not). `truth-current` holds after the edit.

**D6 — The durable witness.** The *"tellme offers no summarisation tool"* half is **already carried**
by the E2E: `specs/truth/features/cli/chat/offering-the-agent-tools.feature` (Rule *"a summarisation
tool must not be offered"*, with its `chat/dsl.md` row). This ADR **cites** that carrier; it adds no
new claim requiring one. The ADR's own durability is witnessed by `make verify-adr-index` (a standing
`make verify` member) plus this index row.

## Consequences

- tellme's context control is **explicit**: inspect with `-l`, trim with `-b`, restart with `--new`.
  A long session is bounded by the **operator**, not by a background summariser.
- The cost is **accepted**: tellme has **no** automatic way to keep a very long session inside the
  window without losing turns (compression is unavailable); the operator must trim or restart.
- The truth tree no longer contradicts itself: `-b` reads as **delivered**, `--retry` as a genuine
  non-introduction.
- Future rounds gain a **citable** home for "why no summarisation?" — and a stated **revisit trigger**
  (below), so the position can be reopened **only** on evidence, without re-litigating a settled choice
  on a whim.
- No product code, no new flag, no new class phrase, no new exit code; `go.mod`/`go.sum` unchanged.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| Port `summarize_history` as an agent tool | Reverses round 021 (*tool-surface parity*); the reference's tool surface diverges from tellme's deliberate small set, and the operator-in-the-loop control (`-l`/`-b`) is the supported surface. |
| Automatic self-healing summarisation | Introduces auto-magic and a non-deterministic context transformation — against tellme's postures (determinism, small surface, operator control); also a much larger change than a truth/record round. |
| Add token-budget pruning | Same objection; pruning is a **settled exclusion** and would silently drop conversation the operator did not choose to drop. |
| Add a **soft warning** when the payload approaches the budget | Not adopted **this round** (out of scope); if a need is shown, it is a small future round — see the revisit trigger. |
| Delete the whole `-b`/`--retry` clause | Would drop **genuine** exclusions (`--retry`, pinning, pruning, `SafePath`/consent) — a new truth defect. |

## Forward (non-blocking)

> **⚠ Not open work.** A decision deferred to a trigger, or a recorded divergence — not tasking.

- **RF-089-1 (revisit trigger).** Revisit **only** if a **measured** session is shown to **routinely**
  exceed the effective budget **and** the operator's manual `-l`/`-b` control is demonstrated
  insufficient. Then a future round opens from a **live issue** (not from this ADR) and decides then.
- **RF-089-2 (the non-equivalence is permanent until a carrier exists).** Any future claim that
  `-l`/`-b` "cover" summarisation must supply a **discriminating** carrier; until then the ADR records
  the **non-equivalence** only.
- **RF-089-3 (`--retry` remains unimplemented).** The reference's user-facing `--retry` (roll back the
  last user message + resend, with confirmation) is **not** ported; `-b "p"` is the nearest tellme
  form (retry **by hand**). Its own decision is deferred (see round 081 `RF-081-7`).
- **RF-089-4 (pruning absence is not mechanically gated).** "No token-budget pruning" rests on the
  E2E + review, not on a dedicated gate (the repo's `topology-audit-not-a-gate` posture); a future
  pruning feature would have to update this ADR and the truth row together.
- **RF-089-5 (`docs/domain-model` not engaged).** The decision changes no modelled behaviour; the
  product model already records `-b`/`--back` and has no summarisation/pruning entity (ADR 0041).
