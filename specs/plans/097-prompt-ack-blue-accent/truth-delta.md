# Truth delta — round 097 `097-prompt-ack-blue-accent`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/techstack.md` — all rows **as they bear on this round's accent** | Checked — **no** technology change **for the accent**: it reuses the existing `internal/ui` colour mechanism (the SGR constants already in `internal/ui/colour.go`); no new dependency, no new dev-tool, no new gate, `go.mod`/`go.sum` unchanged. | NFR-001 / SC-004. |
| MODIFY | `specs/truth/techstack.md` — the ***Dependency vulnerability scan*** (`govulncheck`) row (round 096's row, `:145`) | **Reconcile the recorded toolchain floor**: it is now **Go ≥ 1.26.10**, not ≥ 1.26.9 — the newly published **`GO-2026-6809`** (`crypto/x509`, reached via `fmt.Fprintf` → `x509.Certificate.Verify`) is fixed only in **`go1.26.10`**, so a `go1.26.9` host reds the gate on that single stdlib advisory with **no repo change**. Current-state wording, same row, same owner. | Discovered **while running this round's `make verify`** (the red is ambient, not caused by the accent); ADR 0066 **D2** puts the floor on this row and its §Forward **RF-066-1** predicted exactly this drift ("the floor is recorded, not machine-checked"). Truth integrity requires the recorded floor to be current. |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — `the input capture is announced for the turn` (round 017 row) | Add the **scoping qualifier** the round-039 `[Tool Reason]` row already uses: the plain-form clause is scoped to a **non-terminal** `stderr`; on a terminal with `-r` off the line additionally carries tellme's own round-097 **blue** accent — a distinct, gated class (the colouring row below owns it). Visible text unchanged; the row still forbids `stdout`. | FR-001/FR-002/FR-003: the ack is now a coloured element on the gate, so the existing "carries a line matching `[HH:MM:SS] Input captured. Processing...`" check needs its fixture class stated, exactly as the round-039 reason row did. |
| ADD | `specs/truth/features/cli/chat/dsl.md` — new row `the input capture acknowledgement is shown in blue` | `Channel`: standard error. `Must-check`: on a colour-enabled terminal `stderr` the **whole** `[HH:MM:SS] Input captured. Processing...` line — timestamp included — is wrapped in the bright-blue SGR pair; `Should Not Happen`: plain on a colour-enabled terminal; present on `stdout`; present in `turns.log`; wrapped when the gate is off. | FR-001/FR-002/FR-005/FR-006: the colouring contract's own authority, mirroring the round-057/058 `the action line is shown in yellow` / `the tool output frame is shown in grey` rows. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — `the session chrome carries no colour` (round-054 row) | Extend the forbidden set with the round-097 bright-blue `\033[1;34m` in addition to the round-054 green, round-057 grey, and round-057 yellow. | EC-001 / SC-002: the negative must be able to **red** for the new accent, or the colour-off half is unwitnessed. |
| MODIFY | `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` | The feature's header comment widens from "accents **four** elements in green" to "four elements in green **plus** the acknowledgement in blue"; a **new Rule** ("The input-capture acknowledgement is blue at a terminal") with two Examples — one positional, one `-i` submit — pins the cross-surface uniformity (FR-005). | FR-001/FR-005; W3 of issue #204 (uniformity is the round's premise). |
| NOOP | `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature`, `reading-a-multi-line-prompt.feature`, `piped-prompt` / `presenting-the-turn` Examples | Checked — their steps are **unchanged**; the ack row they use keeps its semantics (the plain/non-terminal class), and the `-i` colour witness lives in the colouring feature (one authority per contract). | `dsl-single-authority`; no behavioural change on those surfaces. |

### PR-review fold (2026-10-22; second reviewer — comment `6334186002`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — `the session chrome carries no colour` (again) | Restated **exhaustively**: no ESC byte at all except the **one documented designed exception** — the `[Tool Output]` close path's unconditional default-state restore `\033[0m` (ADR 0007 TD-2). Replaces the hand-enumerated accent-code list, which had already been widened twice and could go vacuously green for a future accent. | PR #205 review finding 2 (intent); the proposed `substring "\x1b"` fix was rejected — it reds the pre-existing piped-command scenario because the unconditional restore is by design. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — `the input capture acknowledgement is shown in blue` | The uniformity clause now names the **three `turnOptions` construction sites** (`:295` · `:429` · `:479`) as the exhaustive surface set, so piped / `-b` are covered structurally by the positional Example. | PR #205 review finding 3. |
| MODIFY | `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` | A **third** surface Example — `A prompt read to end-of-input at a terminal is acknowledged in blue` — witnessing the Ctrl+D site (`:479`), the one construction site with no carrier. | PR #205 review finding 3 (the discriminating gap). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change; `turns.log` behaviour is unchanged (the ack was and remains excluded). | The round changes no state. |
| NOOP | `ui/**` | Checked — no plan-side UI surface; the `-i` editor is untouched. | Plain line-oriented chrome change; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0067-prompt-ack-blue-accent.md` (+ its `docs/decisions/README.md` index row) | Record the fifth chrome element: the input-capture acknowledgement in **bright blue**, why green was rejected (the `$` accent), and the recorded divergence (the reference greens this line). | `adr-index-consistent`; ADR 0023 D3 is operator-locked and immutable ⇒ a forward-pointer annotation on 0023, not an edit. |
| MODIFY | `docs/decisions/README.md` — the **ADR 0023** index row | Forward pointer to 0067 (its D3 element set is extended to five). Body of 0023 stays **verbatim**. | Immutability convention (ADR 0065 / 0057 precedent). |
| MODIFY | `docs/domain-model/tellme.modelith.yaml` + the re-rendered `.md` | The `Chrome` entity + the `chrome-colour-terminal-gated` invariant name the acknowledgement as a fifth accent carrying **bright blue**. | ADR 0041 — the model is load-bearing; `modelith-check` is a `verify` member and fails on drift. |
| MODIFY | `STATUS.md` (+ the day's session summary) | Round 097 delivered; the accent recorded as live state. | Session closeout hygiene (not truth). |
| MODIFY | **environment only (not a repo artifact)** | The kiosk Go toolchain is raised **user-scoped** `go1.26.9 → go1.26.10` (`~/sdk/go1.26.10` + the `~/.local/bin/go` wrapper), clearing `GO-2026-6809`. | ADR 0066 **D3**'s standing method: clear the module half in `go.mod` (nothing to do — no module advisory), satisfy the toolchain half in the environment; `RF-066-1`'s recorded-not-machine-checked limit. |
