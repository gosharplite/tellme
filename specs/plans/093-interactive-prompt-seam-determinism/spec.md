# Round 093 — `093-interactive-prompt-seam-determinism`

**Theme**: make the `-i` interactive-prompt **E2E seam** deterministic and faithful. Two root causes in the
scripted-key test seam make `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature`
intermittently red, and make its multi-line Example vacuous:

- **(A) the paint gate is not content-aware.** The output-synchronized handshake
  (`tests/e2e/harness.RunInWithSyncedStdin`) delivers the terminal key once the child's stderr shows a
  **generic border marker** (`┌`) — a marker any painted frame satisfies, including a frame flushed *before*
  the reader has applied the whole scripted composition. Under scheduling contention the renderer flushes
  such a **partial** frame, the harness writes the terminal key, and the remainder of the composition
  **coalesces** with the terminal key, so the frame carrying the **complete** composed text is never
  flushed. The E2E assertion (which reads a flat accumulated capture) then finds the typed text missing →
  intermittent red. Measured: **599 / 960** runs red at 16-way concurrency.
- **(B) the scripted line break is the wrong byte.** The seam scripts a line break as a raw `\n` (LF). In
  the child, bubbletea decodes LF as `KeyCtrlJ`, which the bubbles textarea does **not** bind to
  `InsertNewline` (it binds CR/`enter` and `ctrl+m`), so the newline is silently dropped and the two typed
  lines are **joined** (`line oneline two`) — the exact symptom recorded on issue #191. The `keeps the typed
  text` assertion passes anyway (both lines remain substrings of the joined row), so the defect is
  invisible: the multi-line Example never exercises two lines.

**Anchor issue**: [#191](https://github.com/gosharplite/tellme/issues/191). **DoD = close it.**

---

## 1. Why this round

Issue [#191](https://github.com/gosharplite/tellme/issues/191) homes the **round-082 O-082-1** observation
(recurred at the round-091 fold-verification, **RES-091-FV-4**): the interactive-prompt E2E scenario
`presenting-the-interactive-prompt.feature:53` fails **once** in a full parallel run, passes on re-run and
5/5 in isolation. An intermittently-red E2E scenario erodes the E2E suite's role as **the gate**
(`make test` always runs the full contract) and can mask a real regression as "the known flake".

The issue's **evidence** names the symptom precisely — *a swallowed Enter → the captured line reads
`line oneline` (two prompts joined)* — and its **requirements sketch** demands: determinism under
repetition (FR-1), a **root-cause** fix, not a retry/sleep paper-over (FR-2; `verify-no-test-sleep`), and a
**deterministic carrier** (a unit pin on the submit/teardown ordering, "a flat capture cannot observe the
timing", **plus** the E2E scenario passing reliably under repetition).

## 2. The change

1. **Content-aware paint gate (root cause A).** The seam's second chunk (the terminal key) is delivered only
   once the capture shows the **composed text**, not merely the editor border. The gate marker is derived
   from the scripted composition (the last visible line of the compose keys), falling back to the border for
   a compose-less (open/abort-only) sequence. This guarantees the **composed** frame has painted before the
   terminal key is written, so the terminal key can no longer coalesce with the composition.
2. **Faithful Enter byte (root cause B).** The seam's scripted line break is the byte a real terminal sends
   for Enter — **CR** (`\r`) — so the product's textarea inserts a newline and keeps the lines apart
   (matching the scripted intent "types `line one\nline two`").
3. **Discriminating carrier.** The `keeps the typed text` assertion is strengthened to require each typed
   line on its **own** editor row, so the pre-fix joined state (`line oneline two`) reddens instead of
   passing vacuously.
4. **A deterministic unit pin** on the paint-gate derivation / ordering (the flat capture cannot observe the
   timing).

No product behaviour change: the fix is confined to the E2E test seam (`tests/e2e/harness/**`,
`tests/e2e/steps/**`) plus its assertion; `go.mod`/`go.sum` unchanged; no new flag, phrase, or exit code.

## 3. Requirements

### Story 1 — the multi-line prompt scenario is deterministic and faithful (P1)

- **FR-001** the `-i` scripted-key handshake MUST deliver the terminal key only after the **composed text**
  has painted (a content-derived paint gate; not the generic border, not a sleep/retry).
  [Verification Intent: unobservable → harness/steps unit pin + the E2E repetition witness]
- **FR-002** the seam's scripted line break MUST be delivered as the terminal Enter byte (CR), so the
  product inserts a line break rather than dropping it. [Verification Intent: observable → acceptance
  scenario 1]
- **FR-003** the `keeps the typed text` assertion MUST require each typed line on its own editor row (the
  joined state reddens). [Verification Intent: observable → acceptance scenario 1]
- **NFR-001** the seam MUST NOT use a sleep/retry as synchronization (root-cause fix; `verify-no-test-sleep`
  green) and MUST NOT change product behaviour; `go.mod`/`go.sum` unchanged.
  [Verification Intent: unobservable → `make verify` + review]
- **NFR-002** `make test` (the E2E gate) MUST be green; the E2E counts are unchanged unless a deterministic
  pin is added (state any delta). [Verification Intent: observable → `make check`]

## 4. Invariants

- **I-1 — Root-cause, not paper-over.** No sleep, retry, or relaxed assertion; determinism comes from the
  content-aware gate.
- **I-2 — Product behaviour identity.** `internal/**` and `cmd/**` are unchanged; the fix is test-seam only;
  no new flag/phrase/exit code.
- **I-3 — Faithful seam.** The scripted keys emulate what a real terminal delivers (Enter = CR); the
  multi-line Example truly keeps two lines.
- **I-4 — Determinism under repetition.** The scenario is green across N consecutive runs (both serial and
  under concurrency), where it was intermittently red before.
- **I-5 — Frozen history untouched.** `specs/plans/NNN-*/**` (delivered) is never edited.

## 5. Scope

**In**: the harness paint gate + the seam's Enter byte (`tests/e2e/harness/cmd_helper.go`,
`tests/e2e/steps/tui_keys.go`, the When steps) + the discriminating assertion (`thenKeepsTypedText`) +
the deterministic unit pin + the interface-truth row/feature-note reconciliation + the design ADR.
**Out**: any product behaviour change; the `-i` surface itself; the E2E counts (unless a pin adds Examples —
state it); frozen plan packages.

## 6. Success criteria

- **SC-001** the multi-line Example passes on its **own** (two editor rows) and cannot pass on the joined
  state. [Verification Intent: observable → acceptance scenario 1]
- **SC-002** the handshake is deterministic: **N≥50** consecutive runs (incl. under concurrency) are green,
  where the pre-fix seam measured ~62 % red at 16-way concurrency. [Verification Intent: unobservable →
  the repetition witness + the unit pin]
- **SC-003** `make check` green; E2E counts unchanged (state the delta, if a pin adds Examples).
  [Verification Intent: observable → `make check`]
- **SC-004** `go.mod`/`go.sum` unchanged; no sleep/retry introduced. [Verification Intent: unobservable →
  `make verify`]
- **SC-005** the design decision is recorded (an ADR) and the interface truth reflects the strengthened
  assertion. [Verification Intent: observable → `verify-adr-index` + the truth review]

## 7. Assumptions

- **A1** The operator's round instruction ("Open a new aixbdd round, the goal is to close #191") grants the
  intent to fix the seam; the residual design choices (the gate derivation, the Enter byte) are RD-owned.
  No `/axb-clarify` needed (0 questions).
- **A2** The product's interactive surface (real terminal) is correct: a real terminal sends CR for Enter,
  so no product change is required — the seam was the defect.
- **A3** No `NEEDS CLARIFICATION`.
