# ADR 0061 — Operator-declared direction: no security layer · no Windows · bash-first · POSIX-only

- **Status:** Accepted
- **Date:** 2026-09-25
- **Deciders:** tellme owner (issue [#186](https://github.com/gosharplite/tellme/issues/186))
- **Related:** rounds 008 (Clarify Q3) and 021 (D4) — the "no security/consent layer" settled exclusion · rounds 012/015 — POSIX-only · round 024 — the bash-first `execute_command` · **ADR 0043** (`search_files`, the small-surface bar) · **ADR 0032** (the capability-gated `read_image`) · **ADR 0060** (context control — the operator-manual posture, a sibling "small surface" decision) · round 090 (`specs/plans/090-record-hygiene-offered-set-and-direction`)

## Context

`tellme` is a deliberate **re-specification** of `tell-me-go` (a multi-provider reasoning agent CLI), not
a feature-for-feature port. Since its first rounds it has been shaped by **operator-declared directions**
that the operator stated repeatedly in-session and that were echoed in `README.md` ("Design Intent &
Direction") and `STATUS.md`. The directions were **never recorded in an ADR** — `README.md` even asserted
that direction changes are "recorded here", making a narrative README the *de facto* home of a durable
governance decision. That is the wrong home: an ADR is the durable, citable record
(`docs/decisions/README.md`: *"a decision that other artifacts or future rounds depend on and must be able
to cite"*). The operator decided (issue #186) that the direction **belongs in an ADR**, and that
`README.md` is **not** the source of truth for directions.

## Decision

**D1 — No security layer.** tellme ships **no** `SafePath` authorization, consent prompt, command
whitelist, or path boundary. Tools read, write, and execute whatever they are given. The reference's
`SecurityManager` / `UserInteractor` are **absent by decision**. The resulting risk (destructive commands,
out-of-tree writes) is an **explicitly accepted operator decision, not an oversight** — the reference's
guardrails were in practice always bypassed, so they were pure overhead that repeatedly failed tool calls
for no protection gained.

**D2 — No Windows.** tellme targets **bash on POSIX only** — no path translation, `cmd`/PowerShell
wrappers, Windows built-in probing, or separator handling. The cross-platform tax is dropped entirely.

**D3 — Bash-first execution.** `execute_command` (run through `bash -c`) is a **first-class primitive**,
not a gated escape hatch. Because bash already provides piping and redirection, a separate `pipe_commands`
tool is **not offered**.

**D4 — POSIX-only surfaces.** The interactive surfaces (the round-012 plain reader and the round-015 TUI)
are POSIX-only; there is no Windows variant.

**D5 — A deliberately small tool surface (the consequence).** A dedicated agent tool earns its place only
if it beats bash on a real axis — **context boundedness**, **determinism / a testable contract**, or
**reliability** (exact-content writes / exact-block replacements). The shipped surface is therefore
`execute_command` plus only the tools that clearly clear that bar: the reader family
(`list_files`/`read_files`/`get_tree` — round 021), the bounded in-file content search `search_files`
(round 071; **ADR 0043**), the write pair (`write_file`/`replace_text` — round 029), and the read-only
`list_skills` (round 033); a provider that declares the `vision` capability is additionally offered
`read_image` (round 062; **ADR 0032**). Tools that merely duplicate a trivial shell command are omitted
rather than carried for parity.

**D6 — This ADR is the direction's authoritative home.** `README.md` (and `STATUS.md`) **summarise** the
direction and **point at this ADR**; they do not assert it independently. A change to a direction is
recorded by a **superseding ADR**, never by editing the README summary.

## Consequences

- The direction is now **citable and durable**: a future round that touches (say) a tool-surface or
  platform question can cite ADR 0061 instead of inferring intent from a README narrative.
- The **accepted risk** of D1 is explicit: an operator (or an agent) must not "fix" the absence of a
  security layer as though it were a defect — that is the settled, deliberate position.
- D5 is a **load-bearing, testable** consequence: the offered set is single-sourced to the live registry
  and checked (round 090's carrier over `agentTools()`), so the "small surface" cannot silently grow.
- `README.md`'s direction section is now a summary; a direction change that edited only the README would
  be **non-authoritative** (the ADR controls).
- No product behaviour changes; `--version` stays `dev`; no new flag/phrase/exit code.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| Keep the direction in `README.md` (status quo) | A README is narrative prose with no governance status; the operator ruled it out — a durable decision needs an ADR that other artifacts can cite. |
| Record the direction in `STATUS.md` | `STATUS.md` is live/rotating state, not a decision record; it already *points* at the direction's home. |
| Record it as an `ADR §Forward` item on an existing ADR | A forward item is a *deferral/disclosure*, not a decision; the direction is a settled decision, not deferred work. |
| Leave it unrecorded (echo the reference) | The direction is a tellme-specific divergence (no security layer; POSIX-only), so the reference asserts nothing to echo. |

## Forward (non-blocking)

> **⚠ Not open work.** A decision deferred to a trigger, or a recorded divergence — not tasking.

- **RF-061-1** — the small-surface bar (D5) is a **judgement**, applied per tool at its round; it is not
  a mechanical admission test. The `search_files` / `read_image` inclusions are the precedents.
- **RF-061-2** — D1's accepted risk has no guardrail by construction; a future operator wanting one would
  record a **superseding ADR** (and reconcile the "no security layer" truth rows), not patch this one.
- **RF-061-3** — the direction's enforcement surfaces are the round-090 carrier (the offered set) and the
  POSIX-only build/verify gates; a *general* "small surface" gate is not adopted (cf. ADR 0060 RF-089-6).
