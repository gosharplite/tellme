# Technical research — round 093 `093-interactive-prompt-seam-determinism`

**Owner**: `axb-technical-research` (produces this `research.md`; owns `specs/truth/techstack.md`).
**Inputs**: `spec.md`, `truth-delta.md`, the current `specs/truth/**`, the codebase
(`tests/e2e/**`, `internal/ui/tui/prompt/**`).

## 1. Problem

Issue [#191](https://github.com/gosharplite/tellme/issues/191) homes the **round-082 O-082-1**
observation (recurred at the round-091 fold-verification, **RES-091-FV-4**): the interactive-prompt E2E
scenario `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature:53` fails **once** in a
full parallel run and passes on re-run / 5-of-5 in isolation. The recorded symptom is precise: *a
swallowed Enter → the captured line reads `line oneline` (two prompts joined)*.

Two distinct root causes sit behind that report; both are in the **E2E test seam**, not the product.

### Root cause A — the handshake's paint gate is not content-aware (the flake)

`tests/e2e/harness.RunInWithSyncedStdin` (round 023) splits the scripted keys into two chunks: the
**compose** keys written immediately, and the **terminal key** written only after the child's `stderr`
shows the painter **marker** `┌`. The intent is to avoid bubbletea's frame coalescing (with all keys
available at once, only the final frame renders and the editor box never reaches the capture). But the
marker `┌` is the **editor border**, present in *any* painted frame — including a frame flushed **before**
the input reader has applied the whole scripted composition. Under scheduling contention the child's
renderer flushes such a **partial** frame, the harness sees `┌`, writes the terminal key, and the remainder
of the composition **coalesces** with the terminal key — so the frame carrying the **complete** composed
text may never flush. The assertion (which reads the flat accumulated capture) then finds the typed text
missing → intermittent red.

**Measured (16-way concurrency, 960 runs, the shipped seam):** **599 / 960 red**; with a **content-aware**
gate (**gate on the composed text**, here its last visible line `line two`): **0 / 960 red**.

### Root cause B — the scripted line break is the wrong byte (the symptom)

The seam scripts a line break as a raw `\n` (LF). In the child, bubbletea decodes LF as **`KeyCtrlJ`**
(`key_names.go`: `keyLF → "ctrl+j"`, `keyCR → "enter"`), and the bubbles `textarea` binds
`InsertNewline` to **`enter` (CR)** and `ctrl+m` — **not** `ctrl+j`. So the LF is silently dropped and the
two typed lines are **joined** (`line oneline two`) — exactly the recorded symptom. The pre-093 assertion
required each line only as a **substring of some editor row**, and `line oneline two` contains both lines,
so the Example passed **vacuously**: the scenario never exercised two lines.

**Measured (16-way concurrency, 960 runs):** scripted LF ⇒ **960 / 960 joined** (the defect is
deterministic, not a flake); scripted **CR** ⇒ **960 / 960 two rows**.

A real terminal delivers **CR** for Enter (raw mode, `ICRNL` off); the product is correct. The seam was the
defect.

## 2. Options

| Area | Option | Verdict |
| --- | --- | --- |
| **A gate** | (A1) content-aware gate: release the terminal key only once the capture shows the composed text | **CHOSEN** (D1) |
| | (A2) keep the generic `┌` marker and inflate `markerDeadline` / add a wait | rejected — not a fix (a partial frame still satisfies `┌`); a wait is a sleep (`verify-no-test-sleep` spirit) |
| | (A3) retry a flaky run | rejected — paper-over (FR-2) |
| | (A4) product-side: flush a final frame on quit, or defer the round-023 clear-on-submit | rejected — a **product behaviour change**, out of scope; the clear is reference parity (round 023) |
| | (A5) deliver compose + key in one write and accept the coalescing | rejected — the composed frame is dropped; the assertion then needs a tolerance, which is the defect |
| **B newline** | (B1) encode the scripted logical line break as the terminal Enter byte (**CR**) at the seam | **CHOSEN** (D2) |
| | (B2) keep LF; product-side: bind `ctrl+j` to insert-newline | rejected — a product behaviour change; also wrong (a real terminal sends CR) |
| | (B3) keep LF; make the assertion tolerant (accept a joined row) | rejected — the round exists to make the joined state red |
| | (B4) write `\r` in the Gherkin instead of `\n` | rejected — the readable escape `\n` belongs in the readable Gherkin; the *seam* owns terminal-byte encoding |
| **C carrier** | (C1) the E2E repetition witness + mechanism unit pins | **CHOSEN** (D4) |
| | (C2) a harness ordering pin against a re-exec helper child | deferred — real value, but more machinery than the round needs; recorded as a forward option (RF-093-1) |

## 3. Decisions

### D1 — A content-aware paint gate (root cause A), single-owned by the harness

`runExecSynced` derives the gate from the **compose** chunk: `paintGate(compose, fallback)` returns the
composition's **last visible line** (control bytes — e.g. a leading submit key — are not visible text),
falling back to the caller's marker (the editor border) for a **compose-less** sequence (open / abort
only). Because the renderer flushes the *whole* composed frame, gating on the composition's tail guarantees
that the frame carrying the **complete** composition has painted before the terminal key is written — the
key can no longer coalesce with the composition. The rule is **single-owned** in the harness (the
`fallbackMarker` argument keeps its meaning; the steps' call site is unchanged).

*Trade-off:* the gate is a **visible substring** of the last line; a composed line too long to render
contiguously (soft-wrapped) would not satisfy it and the handshake would fall back to `markerDeadline`.
All shipped interactive composes are short single-line values or the two-line fixture, so this is a
**recorded** limit (RF-093-2), not a defect.

### D2 — The scripted line break is the terminal Enter byte (CR, root cause B)

`tests/e2e/steps/tui_keys.go` gains `tuiKeyEnter = "\r"` and `typeText(text)` = `\\n → CR`; every typed
keystroke sequence routes through it (`tuiKeysTypeAndAbort`, `tuiKeysTypeAndSubmit`,
`tuiKeysTypeAcceptAbort`, and the round-038 empty-submit sequence). The Gherkin keeps the readable `\n`;
the seam owns the terminal-byte encoding. A real terminal sends CR for Enter, so the seam is now
**faithful**.

### D3 — Make the assertion discriminating

`thenKeepsTypedText` gains a multi-line clause: no single editor row may carry two of the typed lines
(the joined `line oneline two` row is the defect). The clause is extracted into the pure helper
`joinedRow(rows, parts)` so a unit pin can redden on the joined state. The sentence
`the interactive prompt keeps the typed text "{text}"` is **unchanged** in the Gherkin — the round
**strengthens its implementation semantics** (the DSL row is the authority) rather than manufacturing a
new sentence (the round-089 lesson: *cite the carrier, don't manufacture one*).

### D4 — Carriers

- **Mechanism unit pins** (deterministic, no timing): the harness `paintGate` pin (a revert to a constant
  gate reddens; the ordering rule is pinned); the steps `typeText` pin (a revert to LF reddens); the steps
  `joinedRow` pin (the discriminating clause).
- **The E2E repetition witness** (the flake is a timing race; a flat capture cannot observe the timing):
  N ≥ 50 consecutive green runs of the feature, plus — for measurement — a 16-way concurrency hammer. The
  pre-fix seam measured ~62 % red at 16-way concurrency; the fixed seam measured 0/960.

A stronger **ordering** pin (a re-exec helper child asserting the key is written only after the composed
text) is a **forward option** (RF-093-1).

### D5 — No product change, no sleep/retry

`internal/**` and `cmd/**` are untouched; no new flag/phrase/exit code; `go.mod`/`go.sum` unchanged; no
`time.Sleep` in the seam (`verify-no-test-sleep`); no new `make verify` member.

### D6 — Records

- **ADR 0063** — the seam's determinism rule (a durable, citable home for the handshake contract).
- `specs/truth/techstack.md` — the *Interactive TUI prompt harness* row (the round-093 clause).
- `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` (a round-093 header note + a
  carrier comment) and `chat/dsl.md` (the `keeps the typed text` row's separate-rows clause + the CR note).
- `docs/domain-model/**` — **not modelled** (ADR 0041 escape hatch): a test-seam change touches no modelled
  behaviour.

## 4. Risks & mitigations

| Risk | Mitigation |
| --- | --- |
| The content gate is satisfied by a **partial** frame that already ends with the composition's last line | The gate is the composition's **tail**; a partial frame cannot carry the tail before it carries the text (a prefix of the composition is missing the tail). |
| A future compose line wraps and the gate never fires | `markerDeadline` bounds the wait and the key is still delivered; recorded (RF-093-2). |
| The strengthened assertion breaks the single-line accept-suggestion Example | `joinedRow` is only applied when the value spans several lines; the single-line path is unchanged (verified: the full E2E is green). |
| The CR change affects another typed scenario | Only the multi-line fixture carries a `\n`; every other typed compose is a single line, where `typeText` is the identity (verified: the full E2E is green). |

## 5. Residual risks (forward)

- **RF-093-1** — no direct *ordering* unit pin (a re-exec helper child); the ordering is pinned **by
  construction** via the `paintGate` unit pin + the repetition witness.
- **RF-093-2** — the gate is a visible substring of the composition's last line; a soft-wrapped line would
  fall back to the deadline.
- **RF-093-3** — the pre-093 gap (a vacuous multi-line assertion) was a **carrier-quality** defect: the
  round strengthens this one clause; a general "assertion is discriminating" gate does not exist (the
  round-089/090/091 record-hygiene lineage).
