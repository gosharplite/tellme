# Tasks — round 097 `097-prompt-ack-blue-accent`

**Plan Package**: `specs/plans/097-prompt-ack-blue-accent`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `features/acceptance/colouring-the-prompt-capture.feature`, `specs/truth/features/cli/chat/colouring-the-session-chrome.feature`, `specs/truth/features/cli/chat/dsl.md`, `docs/decisions/0067-prompt-ack-blue-accent.md`
**Type**: presentation round — one accent on one chrome line; no provider/state change.

## Phase 1: Setup

- [X] **T001** — Created the round branch `097-prompt-ack-blue-accent` off `dev` `0f72b06`; created the plan package (`spec.md`, `checklists/requirements.md`, `truth-delta.md`, `features/acceptance/`).

## Phase 2: Foundational

- [X] **T002** — Baseline: confirmed the ack is emitted with the colour adapter **hard-wired false** (`internal/cli/cli.go:858` → `dp.NewLines(false)`) while every other accent-bearing line reads `chromeColour`; confirmed the gate `chromeColour` (`:1116`) = `spinnerGate` and that all five prompt surfaces set `chrome: true` (`:295`, `:429`, `:479`). *(research D2; FR-002)*

## Phase 3: Test Alignment & Implementation

- [X] **T003** — **RED**: the truth contract (`colouring-the-session-chrome.feature` + its two new Rules, and the widened `the session chrome carries no colour` row) fails — the ack is plain on a colour-enabled terminal and the negative row cannot yet distinguish blue. *(FR-001/FR-002/FR-005)*
- [X] **T004** — `internal/ui/turn.go` — add `formatInputCapturedColour(t time.Time, colour bool) string`: the plain path returns `FormatInputCaptured(t)` **verbatim**; the colour path wraps the **whole** line (timestamp included) via `blue(...)`. `FormatInputCaptured` stays the single plain entry point, signature unchanged. *(FR-001/FR-003; ADR 0067 D3)*
- [X] **T005** — `internal/ui/render_ports.go` — `Lines.InputCaptured` becomes a **value receiver** `(l Lines)` returning `formatInputCapturedColour(t, l.colour)` (the `PayloadMeasured` / `ToolReason` shape). *(FR-001; ADR 0023 D5)*
- [X] **T006** — `internal/cli/cli.go` — the ack call site passes the **existing** gate: `emitInputCaptured(env, dp.NewLines(chromeColour(opts, env)))` (was `dp.NewLines(false)`). No new predicate. *(FR-002; research D2)*
- [X] **T007** — `tests/e2e/steps/step_r054_colour.go` — widen `thenChromeNoColour`'s forbidden set with the round-097 `\x1b[1;34m`. *(EC-001/SC-002)*
- [X] **T008** — `tests/e2e/steps/step_r097_ack_blue.go` — the Then `the input capture acknowledgement is shown in blue`: asserts the bright-blue pair around the **whole** ack line (a regex that pins the timestamp *inside* the wrap). *(FR-001)*
- [X] **T009** — **GREEN**: the new Examples pass on both surfaces (positional + `-i` submit), the `-r` Example passes (line present, plain), and the widened negative row still passes on the piped default. *(FR-001/FR-002/FR-004/FR-005)*

## Phase 4W: Witness Pins

- [X] **T010** — **[WITNESS] W1 — the accent is discriminating.** Remove `blue(...)` from `formatInputCapturedColour` (return the plain form) ⇒ the positive Examples **red** with "missing the blue acknowledgement accent". Restore ⇒ green. *(CLM-001; ADR 0006)*
- [X] **T011** — **[WITNESS] W2 — the negative is discriminating.** Add `\x1b[1;34m` to the emitted plain ack (i.e. colour it unconditionally) ⇒ the piped / `-r` Examples **red** with "must carry no colour". Restore ⇒ green. *(CLM-002)*
- [X] **T012** — **[WITNESS] W3 — the gate is the real gate, not `opts.chrome` alone.** Force the colour regardless of `raw` (drop `!opts.raw` from `chromeColour`) ⇒ the `-r` Example **red**. Restore ⇒ green. *(CLM-003; ADR 0067 D5)*
- [X] **T013** — **[WITNESS] W4 — uniformity.** Delete the accent from the `-i` path only (i.e. gate the colour on `!opts.echo`) ⇒ the `-i` Example **red** while the positional Example stays green. Restore ⇒ both green. *(CLM-004; ADR 0067 D4)*
- [X] **T014** — **[WITNESS] unit pins** — `formatInputCapturedColour` colour-off ⇒ **byte-identical** to `FormatInputCaptured`; colour-on ⇒ the pair wraps the whole line including the timestamp; `wrap`'s empty-safety holds. *(FR-003; CLM-005)*

## Verification

- [X] **T015** — `make check` (verify + full E2E) green; E2E scenario/step counts reconciled; `make test-race` no races; `make modelith-check` no drift; `make verify-adr-index` consistent; the topology audit reports the same error/warning state. *(SC-003/SC-004; I-1/I-2/I-3/I-4)*
- [X] **T016** — Record: `truth-delta.md` rows finalised; ADR 0067 + the 0023 forward pointer; the domain model (YAML + rendered MD); `STATUS.md` + the day log.

## Delivery

- [X] **T017** — Commit the round branch; open the round PR into `dev` (no Copilot review; only a human merges).

---

## Claim → Witness ledger

| Claim ID | Source spec / Truth anchor | Claimed effect | Witness type | Bound Task / Decision | Discriminating falsifier | Status |
| --- | --- | --- | --- | --- | --- | --- |
| CLM-001 | FR-001 · SC-001 · `the input capture acknowledgement is shown in blue` | on a colour-enabled terminal the ack line is wrapped in `\033[1;34m…\033[0m`, timestamp inside the wrap | `[WITNESS]` W1 + the positive Examples | T008 / T009 / T010 | deleting the `blue(...)` call leaves the ack plain ⇒ the Example reddens (mutation run) | verified |
| CLM-002 | FR-002 · EC-001 · SC-002 · `the session chrome carries no colour` | off the gate (piped `stderr`, `-r`) the ack carries **no** SGR byte | `[WITNESS]` W2 + the negative Examples | T007 / T011 | colouring the ack unconditionally ⇒ the piped and `-r` Examples redden (mutation run) | verified |
| CLM-003 | FR-004 · EC-002 · ADR 0067 D5 | under `-r` the ack is **still printed**, merely uncoloured | `[WITNESS]` W3 (the `-r` Example) | T012 | dropping `!opts.raw` from the gate ⇒ the `-r` Example reddens (mutation run) | verified |
| CLM-004 | FR-005 · ADR 0067 D4 | the accent lands on **every** prompt surface (positional **and** `-i` submit) | `[WITNESS]` W4 (the two cross-surface Examples) | T009 / T013 | gating the colour on `!opts.echo` ⇒ the `-i` Example reddens while the positional one stays green (mutation run) | verified |
| CLM-005 | FR-003 · I-2 · NFR-001 | the colour-off path is byte-identical; no palette growth | `[WITNESS]` unit pins | T014 | wrapping the plain path (or adding a new constant) reddens the byte-identity pin | verified |
| CLM-006 | FR-006 · EC-004 · `the session turn log carries no decoration` | the accent never reaches `turns.log` (a separate plain writer) | existing E2E row (already run on a colour-enabled turn) | T015 | the ack is written to `env.turnsLog` ⇒ the row reddens | verified (existing carrier) |

## Pre-Delivery Orphan Coverage Sweep

- `truth-delta.md` non-NOOP items: the ack announced row (MODIFY), the blue row (ADD), the no-colour row (MODIFY), the colouring feature (MODIFY) — each covered by T003/T007/T009.
- `research.md` Decisions 1–6: D1 → T004/T014; D2 → T006; D3 → ADR 0067 + T016; D4 → T003/T009; D5 (modelled) → T016; D6 (out of scope) → the scope boundary recorded in the spec §8 A4 and this ledger's exclusions.
- `spec.md` FR-001..FR-006 / NFR-001 / SC-001..SC-004 / EC-001..EC-004: all bound above.
- Acceptance feature `colouring-the-prompt-capture.feature`: all 7 Examples map onto truth rows (`… is shown in blue` ×4, `… is announced` + `carries no colour` ×2, `turn log carries no decoration` ×1). No orphan acceptance rule.
- No orphan artifacts.
