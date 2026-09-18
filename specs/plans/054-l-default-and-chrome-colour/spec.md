# Feature Specification: `-l` defaults to `1` + green chrome accents (round 054)

**Feature Branch**: `054-l-default-and-chrome-colour`

**Created**: 2026-09-19

**Status**: Draft — **clarify round 1: Q1 (the colour gate) LOCKED by direction + precedent** (see below); no open `NEEDS CLARIFICATION`. Produced by `/axb-specify` from operator tasking.

**Input (operator, 2026-09-19)**: *"Round 054 = the `-l` default-1 parity fix. I also want some text color. `[Tool Reason]` lines are green … `butler` (MODE) in Payload lines (two) are green. `438165` (tokens) is green in `[06:25:11] Payload: 438165/1000000 tokens - butler - deepseek-flash`. `$1.0607` (3rd cost) is green. Check the color green used in tell-me-go."* Then: *"1. bundle both. 2. Confirm the four items, accept tellme is different, green the whole line of `[Tool Reason]`."*

**Behaviour intent**: **MODIFY (user-facing)** — two independent changes: **(a)** bare `-l` gains a default of `1` (reference parity); **(b)** the diagnostic chrome gains **green accents** on four elements, emitted only on a terminal. `stdout`, exit codes, the DSL phrase vocabulary, and the offline `turns.log` content are **unchanged**.

---

## Grounded in the current system

### `-l` arity (US1)
| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go` `parseFlags` | `fs.IntVarP(&o.list, "list", "l", 0, …)` — no `NoOptDefVal`; `o.listSet = fs.Changed("list")` |
| `dispatchReporting` | `if f.list <= 0 { return emitUsageError(…) }` — so bare `-l` (no value) is a pflag **usage error** (`flag needs an argument`), and `-l 0`/negative is a usage error |
| reference (`tell-me-go` `chat_command.go:91-92`) | `fs.Lookup("last").NoOptDefVal = "1"` **plus** `sanitizeArgs`/`consumeOptionalIntFlag` pre-pass — bare `-l` ⇒ `1`; `-l 5` consumes the `5`; `-l` before a non-integer leaves the token |

### Chrome (US2) — the exact lines today (`internal/ui`)
- `[Tool Reason]` — `ui.ToolReason(t, reason)` → `[HH:MM:SS] [Tool Reason] <reason>` (plain, single line).
- `Payload` (estimated) — `ui.FormatPayloadStatus(..., estimated=true)` → `[HH:MM:SS] Payload: ~<tokens>/<budget> tokens - <mode> - <model>`.
- `Payload` (measured) — same formatter, `estimated=false`, no `~`.
- `Ready` — `ui.FormatReady(...)` → `╰─⠿ Ready ($<lastCall> $<turn> $<session> - M: … H: … O: … - <hit%>%)`.

**tellme is plain text by design** (round-017 Decision 3: *"no ANSI this round"*); the E2E asserts these lines with **regexes** (`tests/e2e/steps/{payload_status,post_turn_status}.go`) over a **non-terminal** stderr. The reference's code is `colorGreen = "\033[0;32m"` (`tell-me-go/internal/ui/colors.go`), gated by `SetUseColor(isTTY && !raw)`.

---

## Clarify round 1

| # | Question | Decision |
| --- | --- | --- |
| **Q1** | **When is the green emitted?** — `tellme`'s chrome is plain-text today and the E2E reads non-terminal stderr with regexes. | ✅ **LOCKED (direction + precedent)**: colour is emitted only when the **diagnostic stream (`stderr`) is a terminal AND `-r` is off** — the same gate the round-019 spinner already uses (`spinnerGate(opts, env.stderrIsTerminal())`), and the reference's `SetUseColor(isTTY && !raw)`. On a redirected/piped `stderr`, under `-r`, and on every offline path the chrome stays **plain** (byte-identical to today). **Operator may veto** — it is recorded as a decision, not a hidden assumption. |
| **Q2** | **Which elements are green, and how does `tellme`'s placement differ from the reference?** | ✅ **LOCKED by the operator**: the **four** elements below, in tell-me-go's green `\033[0;32m`; **accepted divergence** — only the session cost matches the reference's own layout. |
| **Q3** | **Does `turns.log` carry the colour bytes?** | ✅ **LOCKED (round-053 RF-53-1 / the "control-free" spec)**: `turns.log` stays **plain text** — the colour is applied only to the `stderr` sink; the turn-log tee keeps the uncoloured chrome. No ANSI enters the session artifact. |

### Q2 → the four green accents (LOCKED)

| # | Element | Where | tellme vs reference |
| --- | --- | --- | --- |
| **1** | `[Tool Reason]` line — the **whole line** (`[HH:MM:SS] [Tool Reason] <reason>`) | the per-call reason line(s) + the post-call tail group | reference greens only the value's line in **gray**; tellme greens the whole line (**accepted divergence**) |
| **2** | the **`MODE`** token (`butler`) in **both** `Payload` lines (estimated `~` + measured) | `FormatPayloadStatus` | reference prints the mode in **gray** (**divergence**) |
| **3** | the **measured token number** (`438165`) | the measured `Payload` line only (the `~`-estimated **pre-flight** number stays uncoloured) | reference colours the token number **yellow/red by budget ratio, never green** (**divergence**) |
| **4** | the **third (session) cost** (`$1.0607`) in `╰─⠿ Ready (…)` | `FormatReady` | **matches** the reference (`renderer.go:459`) |

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - bare `-l` lists the last message (Priority: P1)

As an **operator**, I want `tellme -l` (no value) to behave like `-l 1`, so the flag matches the reference and the SOP's shorthand works without a value.

**Why this priority**: it is the round's parity item; it is small and independently verifiable.

**Independent verification**: `tellme -l` on a seeded session prints the last message and exits 0 (identical to `-l 1`); `tellme -l 5` still prints five; `tellme -l "prompt"` treats the token as a **prompt**, not a count.

**Acceptance Scenarios**:

1. **Given** a session with persisted messages, **When** the operator runs `tellme -l` (no value), **Then** it lists the **last one** message and exits successfully.
2. **Given** the same session, **When** the operator runs `tellme -l 5`, **Then** it lists the last five (the explicit value is honoured; unchanged).
3. **Given** `tellme -l` followed by a non-numeric token (a prompt), **When** it runs, **Then** the count defaults to 1 and the token is **not** consumed as the count.
4. **Given** `tellme -l 0` or a negative, **When** it runs, **Then** it is still a **usage error** (unchanged).

**Functional Requirements**:

- **FR-001**: Bare `-l` / `--list` MUST default to `1` (reference parity) — implemented as `NoOptDefVal = "1"` **plus** an args pre-pass that consumes an adjacent integer and leaves a non-integer token untouched (pflag's `NoOptDefVal` alone would parse `-l 5` as `-l=1` + positional `5`).
- **FR-002**: An explicit non-positive value (`-l 0`, `-l -1`) MUST remain a **usage error** (round-007 unchanged).

---

### User Story 2 - the chrome carries green accents on a terminal (Priority: P2)

As an **operator**, I want the diagnostic chrome to accent four elements in green on a terminal, so a long session is easier to scan — while redirected/piped output and the session `turns.log` stay plain.

**Why this priority**: it is an operator-requested readability change; it must not disturb the byte-contracts.

**Independent verification**: with `stderr` a terminal (the `TELL_ME_FORCE_STDERR_TTY` seam) the four elements carry `\033[0;32m…\033[0m`; with `stderr` redirected, under `-r`, and in `turns.log`, the bytes are plain.

**Acceptance Scenarios**:

1. **Given** the diagnostics are shown at a terminal, **When** a tool-using turn runs, **Then** each `[Tool Reason]` line is wrapped in green, the `MODE` and the measured token count in each `Payload` line are green, and the **session cost** in `╰─⠿ Ready` is green.
2. **Given** the diagnostics are **not** a terminal (piped/redirected), **When** the same turn runs, **Then** the chrome is **byte-identical to today** (no ANSI).
3. **Given** `-r` (raw), **When** the turn runs on a terminal, **Then** the chrome is plain (no ANSI).
4. **Given** a prompt turn on a terminal, **When** the session's `turns.log` is read, **Then** it holds the **plain** chrome (no ANSI bytes).

**Functional Requirements**:

- **FR-003**: The chrome MUST colour, **only when `stderr` is a terminal and `-r` is off**: (i) every `[Tool Reason]` line (whole line); (ii) the `MODE` token in both `Payload` lines; (iii) the measured token number in the measured `Payload` line; (iv) the session (third) cost in the `╰─⠿ Ready` line.
- **FR-004**: The colour MUST be the reference's green `\033[0;32m` … `\033[0m` (the code only; the **element choice** is tellme's, per Q2 — a recorded divergence).
- **FR-005**: `stdout` MUST stay byte-exact; the `tellme: {phrase}` vocabulary and exit codes MUST be unchanged.
- **FR-006**: `turns.log` MUST stay **plain text** (Q3): the colour is applied to the `stderr` sink only; the turn-log tee records the uncoloured chrome (round-053 RF-53-1).

---

## Edge Cases

- **`-l` followed by a prompt** — the token must not be consumed as the count (`tellme -l hello` ⇒ count 1, prompt `hello`); mirrors the reference's `consumeOptionalIntFlag` guard.
- **`-l` combined with other flags** — `tellme -l -r` ⇒ `-l 1 -r`; `tellme -l -t` ⇒ `-l 1` wins (the round-053 precedence).
- **`-l=3` / `-l3`** — an attached value is honoured (unchanged).
- **Colour gate precedence** — non-terminal `stderr` beats everything; `-r` off is required; the four elements are coloured independently of the spinner (which has its own gate).
- **The `[Tool Reason]` value after sanitization** — the round-036/039 fold+cap+sanitize still runs first; colour wraps the **rendered** line (no control bytes from the value can escape the wrapper).
- **ANSI in an answer / tool value** — the answer is `stdout` (unchanged); tool values stay sanitized; our own colour never enters `stdout`.
- **`turns.log` and colour** — the tee must not receive ANSI (FR-006).
- **E2E is non-terminal** — so all existing regex assertions over `stderr` remain valid; the colour is asserted via the `TELL_ME_FORCE_STDERR_TTY` seam only.

## Key Entities

- **The `-l` default** — `NoOptDefVal` + the args pre-pass in `parseFlags`.
- **The chrome colour policy** — one gate (stderr TTY && !raw) applied to the four elements; `internal/ui` owns the bytes.
- **The colour code** — `ui.colorGreen = "\033[0;32m"`.

## Success Criteria

- **SC-001**: `tellme -l` ≡ `tellme -l 1`; `-l 5`/`-l 0` behaviours unchanged.
- **SC-002**: On a terminal, the four elements are green; piped/`-r`/`turns.log` are plain.
- **SC-003**: `stdout` byte-exact; `specs/truth/**` updated (CLI usage row + the history feature for `-l`; a new acceptance Rule for the colour, if the PM/DSL owners carry it).
- **SC-004**: `gofmt`/`go vet` clean · `make verify` **OK** · `go test -count=1 ./...` green (incl. the godog E2E) · `go.mod`/`go.sum` unchanged.

## Assumptions

- **A1**: The colour gate is `stderr`-TTY && `!raw` (Q1) — the spinner-gate precedent + the reference.
- **A2**: The four elements and the whole-line `[Tool Reason]` (Q2); the divergence from the reference's own layout is accepted.
- **A3**: `turns.log` stays plain (Q3).
- **A4**: `/axb-api-plan` + `/axb-data-plan` are **NOOP**; `/axb-spec-by-example` is **invoked** (a user-visible behaviour change); `/axb-dsl-refine` adds the `-l` bare-form Example/`DSLRow` and the colour Example.
- **A5**: ADR 0023 records the round (the `-l` default + the colour policy + the recorded divergence).

## Out of scope (recorded forward items)

- Colouring **any other** element (the `Ready` label, the metrics line's other numbers, the turn header, the spinner).
- The self-diagnosing retrieve (round-053 RF-53-4); `--json` listings.
- Windows / a security layer / conversation pruning / tool concurrency (settled exclusions).
