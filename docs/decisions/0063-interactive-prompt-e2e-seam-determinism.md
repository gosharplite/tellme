# ADR 0063 — The `-i` interactive-prompt E2E seam: a content-aware paint gate and a faithful Enter byte

- **Status:** Accepted
- **Date:** 2026-09-25
- **Deciders:** tellme owner (issue [#191](https://github.com/gosharplite/tellme/issues/191))
- **Related:** round-082 **O-082-1** (the first observation, `docs/archives/status/2026-09-23.md`) · round-091 **RES-091-FV-4** (the recurrence, homed on #191) · round 023 (the output-synchronized handshake) · round 016 (`TELL_ME_TUI_DEBOUNCE=0`) · **ADR 0024** (E2E throughput / `markerDeadline`) · the round-042/ADR-0012 hermetic-seam lineage · round 093 (`specs/plans/093-interactive-prompt-seam-determinism/`)

## Context

The `-i` interactive TUI prompt is driven end-to-end **without a pty**: the built binary runs under the
forced-terminal seam (`TELL_ME_FORCE_STDIN_TTY=1`) with a **scripted key sequence** on stdin. Round 023
fixed a frame-coalescing hazard by splitting the sequence into two chunks — the **compose** keys written
immediately, the **terminal key** (submit/abort) written only after the child's `stderr` shows a painter
marker — an output-synchronized handshake (`tests/e2e/harness.RunInWithSyncedStdin`) instead of a
wall-clock sleep. The chosen marker was the editor border `┌`.

Issue [#191](https://github.com/gosharplite/tellme/issues/191) records that
`specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature:53` fails **once** in a full
parallel run (round-082 **O-082-1**; recurred at round-091 fold-verification as **RES-091-FV-4**) with the
symptom *a swallowed Enter → the captured line reads `line oneline` (two prompts joined)*.

Investigation (round 093) found **two** root causes, both in the **test seam** (the product and its
`-i` surface are correct):

- **The paint gate is not content-aware.** The marker `┌` is the editor border, satisfied by **any** painted
  frame — including a frame flushed before the input reader has applied the whole composition. Under
  scheduling contention the renderer flushes such a **partial** frame, the harness releases the terminal
  key, and the rest of the composition **coalesces** with the key, so the frame carrying the complete
  composed text may never flush; the E2E assertion then finds the typed text missing. **Measured: 599 / 960
  red at 16-way concurrency.**
- **The scripted line break is the wrong byte.** The seam scripted Enter as a raw `\n`. bubbletea decodes LF
  as `KeyCtrlJ` (not `KeyEnter`), which the bubbles `textarea` does **not** bind to `InsertNewline`
  (`enter`/CR and `ctrl+m` only), so the line break was silently dropped and the two typed lines collapsed
  onto one row (`line oneline two`) — the recorded symptom. Worse, the pre-093 assertion only required each
  line as a **substring of some editor row**, and the joined row contains both, so the multi-line Example
  passed **vacuously**. A real terminal sends **CR** for Enter.

## Decision

**D1 — The handshake's paint gate is content-aware, single-owned by the harness.** `runExecSynced` derives
the gate from the **compose** chunk: `paintGate(compose, fallback)` returns the composition's **last
visible line** (control bytes, e.g. a leading submit key, are not visible text), falling back to the
caller's marker (the editor border) for a **compose-less** sequence (open / abort only). Because the
renderer flushes the whole composed frame, gating on the composition's **tail** guarantees the frame
carrying the complete composition has painted before the terminal key is written, so the key can no longer
coalesce with the composition.

**D2 — The scripted line break is the terminal Enter byte (CR).** The steps' key convention gains
`tuiKeyEnter = "\r"` and a `typeText` encoder (`\n → CR`) applied by every typed keystroke sequence. The
readable Gherkin keeps `\n`; the **seam** owns the terminal-byte encoding. The seam is now faithful to
what a real terminal delivers.

**D3 — The keep-the-typed-text assertion is discriminating.** `thenKeepsTypedText` requires that, when the
typed value spans several lines, **no single editor row carries two of them** — the joined row is the
defect. The Gherkin sentence is unchanged; the round strengthens its **implementation semantics** (the DSL
row is the authority) rather than manufacturing a new sentence.

**D4 — The seam change is witnessed deterministically.** Mechanism **unit pins**
(`paintGate` — a revert to a constant gate reddens; `typeText` — a revert to LF reddens; `joinedRow` — the
discriminating clause) plus the **E2E repetition witness** (N ≥ 50 consecutive green runs; a 16-way
concurrency hammer for measurement). A direct *ordering* unit pin (a re-exec helper child) is a forward
option (**RF-093-1**).

**D5 — No product change; no sleep/retry.** `internal/**` and `cmd/**` are untouched; no new flag, phrase,
exit code, or dependency; `go.mod`/`go.sum` unchanged; no `time.Sleep` in the seam
(`verify-no-test-sleep`); no new `make verify` member.

## Why an ADR

The **handshake contract** of the `-i` E2E harness is a rule future rounds depend on: any new scripted-key
scenario inherits it, and a future edit to the harness or a new TUI scenario must know that (a) the
terminal key is gated on the **composed text**, not a generic marker, and (b) a scripted **line break** is
the terminal **Enter (CR)** byte, not a line feed. Recording it prevents a silent re-introduction of the
round-191 flake and the vacuous multi-line assertion.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| Keep the generic `┌` marker; inflate `markerDeadline` / add a wait | Not a fix — a **partial** frame still satisfies `┌`; a wait is a sleep (forbidden as synchronization). |
| Retry a flaky run | Paper-over (issue #191 FR-2). |
| Product-side: flush a final frame on quit, or defer the round-023 clear-on-submit | A **product behaviour change**, out of scope; the clear-on-submit is reference parity (round 023). |
| Product-side: bind `ctrl+j` to insert-newline | A product behaviour change, and wrong — a real terminal sends CR for Enter. |
| Keep LF; make the assertion tolerant of a joined row | Non-discriminating — the round exists to make the joined state red. |
| Write `\r` in the Gherkin instead of `\n` | The readable escape belongs in the readable Gherkin; the **seam** owns terminal-byte encoding. |
| A harness ordering pin against a re-exec helper child | Real value, but more machinery than the round needs; recorded as **RF-093-1**. |

## Consequences

### Positive

- The `-i` scenario is **deterministic**: 0 / 960 red at 16-way concurrency (was 599 / 960); 50/50 green
  serially.
- The multi-line Example is **faithful**: the typed two lines are kept on two editor rows, and the joined
  state now **reddens** (was vacuous).
- The handshake rule and the Enter byte are **recorded** (this ADR + the DSL row + the feature note), so
  the flake cannot be silently re-introduced.

### Negative / Accepted Trade-offs

- The paint gate is a **visible substring** of the composition's last line; a composed line too long to
  render contiguously (soft-wrapped) would not satisfy it and the handshake falls back to
  `markerDeadline`. All shipped interactive composes are short or the two-line fixture. Recorded
  (**RF-093-2**).
- The ordering is pinned **by construction** (the `paintGate` unit pin + the repetition witness), not by a
  direct ordering unit test. Recorded (**RF-093-1**).

### Neutral

- No product behaviour, tool, wire shape, persisted record, config key, `.feature` step sentence, or
  `go.mod`/`go.sum` change (the multi-line Example's Gherkin is unchanged; only its semantics are
  strengthened). `docs/domain-model/**` is **not modelled** (ADR 0041 escape hatch — a test-seam change
  touches no modelled behaviour).

## Verification

- **Witnesses** — W-A: revert the gate to a constant marker ⇒ 607 / 960 red at 16-way concurrency (the
  `paintGate` pin reddens). W-B: revert the Enter byte to LF ⇒ the multi-line Example reddens (the joined
  row) and the `typeText` pin reddens. W-C: drop the separate-rows clause ⇒ the joined state passes (the
  `joinedRow` pin reddens). All reproduced then reverted.
- **Determinism** — 50/50 green serially; 0 / 960 red at 16-way concurrency (was 599 / 960).
- **Behaviour identity** — `internal/**`/`cmd/**` unchanged; the full unit suite + the godog E2E are green;
  E2E **scenarios 330** (unchanged) · **steps 2487** (unchanged; the strengthened Then is the same line).
- **Gates** — `make check` green (`make verify` incl. `verify-no-test-sleep`, `verify-architecture` at its
  0-violation baseline, `modelith-check`); `go.mod`/`go.sum` unchanged.

## References

- `tests/e2e/harness/cmd_helper.go` (`RunInWithSyncedStdin` / `paintGate` / `runExecSynced`) ·
  `tests/e2e/steps/tui_keys.go` (`tuiKeyEnter` / `typeText` / the type helpers) ·
  `tests/e2e/steps/step_t011_chat_then_keeps_typed_text.go` (`thenKeepsTypedText` / `joinedRow`) ·
  `tests/e2e/harness/cmd_helper_test.go`, `tests/e2e/steps/tui_keys_test.go`,
  `tests/e2e/steps/step_t011_chat_then_keeps_typed_text_test.go` (the round-093 pins) ·
  `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` + `chat/dsl.md`.
- Round 023 (the handshake) · round 016 (`TELL_ME_TUI_DEBOUNCE=0`) ·
  [ADR 0024](0024-e2e-suite-throughput.md) · round 093: `specs/plans/093-interactive-prompt-seam-determinism/`.

## §Forward (deferred, non-blocking)

> **⚠ Not open work.** A decision deferred to a trigger, or a recorded divergence — not tasking.

- **RF-093-1** — no direct **ordering** unit pin. The ordering ("the terminal key is delivered only after
  the composed frame painted") is pinned **by construction** via the `paintGate` unit pin (the gate *is*
  the ordering rule) plus the repetition witness. A stronger pin would run the handshake against a
  re-exec helper child that records the two chunks' arrival order.
- **RF-093-2** — the gate is a visible substring of the composition's last line; a **soft-wrapped** composed
  line would not satisfy it and the handshake would fall back to `markerDeadline` (a bounded, still-exits
  wait, not a hang).
- **RF-093-3** — the pre-093 gap was a **carrier-quality** defect (a non-discriminating assertion). The
  round strengthens this clause; tellme has **no** general "assertion is discriminating" gate (the
  round-089/090/091 record-hygiene lineage — a docs/record claim has no mechanical carrier).
