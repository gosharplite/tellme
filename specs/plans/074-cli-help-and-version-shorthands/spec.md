# Feature Specification: `-h` help and `-v` version shorthands (round 074)

**Feature Branch**: `074-cli-help-and-version-shorthands`

**Created**: 2026-09-21

**Status**: Draft (specified — clarify **pending**; see §Clarify strategy)

**Anchor**: **operator request** — no anchor issue.

**Input (operator, 2026-09-21)**:

> *"Please add simple `-h` and `-v` flags for tellme."*

…following the operator's comparison of tellme's CLI surface with the reference's. The reference (cobra-based) exposes `-h, --help` and `-v, --version` on every command; tellme ships a single pflag flag set with **no `-h`/`--help` and only the long `--version`** (so `tellme -h`, `tellme --help`, and `tellme -v` are today unrecognized flags → the usage block + `tellme: the command-line usage is invalid` on `stderr`, **exit 2**).

**Behaviour intent**: **ADD (two CLI-surface shorthands)** — tellme gains a **`-h` help** flag and a **`-v` version** shorthand, with no other behaviour change. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The reference's shape** (`tell-me-go -h`, verified on the installed binary): the help text goes to **`stdout`**, prints the flag list, and **exits 0**; `tell-me-go -v` prints `tell-me-go version …` and exits 0. cobra auto-adds both. tellme has no subcommands and no cobra — the round must reproduce the *observable contract*, not the framework.
- **`--version` already exists** (`internal/cli/cli.go` `fs.BoolVar(&o.version, "version", …)`; the run path prints `tellme {version}` to `stdout` and returns `Success`). `-v` is therefore a **shorthand addition**, and the long form must keep working unchanged.
- **`-h` is new.** tellme currently has **no** help flag: an unrecognized flag is a **usage error** — pflag prints the flag-usage block, then tellme prints `tellme: the command-line usage is invalid` on `stderr` and returns the pinned usage code **2** (`specs/truth/features/cli/usage/unsupported-cli-usage.feature`). The round must not disturb that failure path.
- **Help is a *successful*, offline, prompt-less action** — like `--version` and `-d`: no provider request, no turn chrome, no spinner, no stdin read. Its placement in the dispatch precedence must be decided (research/plan): today the precedence is `--version` → `-d` → `-l` → `-t` → `--tool-usage`.
- **The frozen class-phrase vocabulary** (`the eleven frozen class phrases`, interface-root `dsl.md`) must not widen: help is **not** an error, so it must **not** print a `tellme: …` phrase; the usage-error path keeps its single phrase.

---

## Grounded in the current system *(measured 2026-09-21, `dev` @ `5206f9f`)*

| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go` — `parseFlags` | one `pflag.FlagSet("tellme", ContinueOnError)`, `SetOutput(stderr)`; flags: `-c/--config`, `-d/--diagnostics`, `--version`, `-r/--raw`, `--new`, `-l/--list`, `-t/--turns`, `-i/--interactive`, `--tool-usage`. **No `-h`/`--help`, no `-v`.** |
| `internal/cli/cli.go` — `run` | `--version` first: `fmt.Fprintf(env.stdout, "tellme %s\n", version); return Success`. Then `dispatchReporting` (`-d` → `-l` → `-t` → `--tool-usage`), then the prompt / boot paths. |
| `internal/cli/cli.go` — `emitUsageError` | `fmt.Fprintln(w, "tellme: the command-line usage is invalid"); return UsageError` (pinned code **2**). pflag prints the flag-usage block to `stderr` *before* this on a parse error. |
| `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | `Rule: tellme reports its build version` — `--version` → `tellme prints the build version`, no chrome/status/spinner, `tellme exits successfully`. |
| `specs/truth/features/cli/usage/unsupported-cli-usage.feature` | every unrecognized flag → `tellme refuses to proceed` + `tellme explains on stderr that "the command-line usage is invalid"` + `tellme exits with the usage error code`. |

---

## Design (the shape is a `/axb-technical-research` decision; residual choices marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | tellme gains a **`-v`** shorthand for the existing `--version`; `-v` behaves exactly as `--version` (stdout `tellme {version}`, exit 0, offline, no chrome). | locked (operator) |
| **S-2** | tellme gains a **`-h`** flag that prints the flag-usage block and exits **0** (a *successful*, offline, prompt-less action). | locked (operator) |
| **S-3** | Does `-h` also carry the **long form `--help`**? (The reference has `-h/--help`; a short-only help flag would be an unusual, arguably untidy surface.) | **clarify Q1** |
| **S-4** | Which **stream** does help write to, and which **exit code** — `stdout` + **0** (GNU/cobra/reference convention) or `stderr` + a dedicated code? | **clarify Q2** |
| **S-5** | **What** does the help block contain — the current pflag flag list only, or a richer block (usage line + flags, as the reference prints a `Usage:`/`Flags:` section)? | **research decision** (D-x) |
| **S-6** | The **dispatch precedence** (where `-h` sits relative to `--version`/`-d`/`-l`/…) and the exact renderer are technical. | **research decision** (D-x) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — No regression to the usage-error path** — an unrecognized flag still refuses with the pinned phrase and exit **2**; the help flag must not weaken it.
- **I-2 — Help is offline and prompt-less** — no provider request, no turn chrome, no spinner, no stdin read; it exits immediately.
- **I-3 — The frozen phrase vocabulary is unchanged** — help prints **no** `tellme: …` class phrase (it is not an error); exactly one phrase still belongs to the usage-error path.
- **I-4 — `--version` is unchanged; `-v` is additive** — existing `--version` behaviour (stdout, exit 0) is preserved byte-for-byte.
- **I-5 — No new dependency; stdlib + the existing pflag; POSIX-only; hermetic.**

---

## Clarify strategy

**Escalate (1–3 questions, one at a time).** Two choices materially change the visible surface and the acceptance:

- **Q1 → S-3** — does `-h` ship with the long form **`--help`** (recommended; the reference has both), or is it `-h`-only?
- **Q2 → S-4** — help to **`stdout` + exit 0** (the conventional/reference contract), or to `stderr` + a dedicated help code?

**Vetoable assumptions (disclosed, not asked):**

- **A1** — `-v` mirrors `--version` exactly (no separate output shape).
- **A2** — "simple" = the help block is the **flag list** (no subcommand/exit-code essay), matching the reference's `Flags:` section without cobra's `Available Commands`.
- **A3** — help is a **prompt-less successful action** (it does not read stdin, does not contact a provider), like `--version`.

**No `NEEDS CLARIFICATION` beyond Q1–Q2**; the residual technical choices (S-5/S-6) defer to `/axb-technical-research`.

---

## User Stories (proposed)

### US1 — The operator can ask for help with `-h` (Priority: P1)

When the operator runs `tellme -h`, tellme prints its flag usage and exits successfully (offline, no provider request, no chrome).

**Why P1**: it is the operator's core request — discoverability of the flag surface.

**Acceptance (proposed)**:

1. **Given** a runnable tellme installation, **When** the operator runs `tellme -h`, **Then** tellme prints the flag-usage block and exits successfully (code 0), with no `tellme: ` error phrase and no provider request.

**Functional requirements (FR)**:

- **FR-001**: tellme MUST accept `-h` and, on its presence, print the flag-usage block and exit `0`.
- **FR-002**: The help output MUST list every flag the process accepts, each with its short form (where one exists).
- **FR-003**: `-h` MUST be an offline, prompt-less, successful action (no provider request, no turn chrome, no spinner).

**Non-functional requirements (NFR)**:

- **NFR-001**: Help MUST NOT print any frozen `tellme: …` class phrase.

---

### US2 — The operator can ask for the version with `-v` (Priority: P1)

`-v` behaves exactly as `--version`.

**Why P1**: the operator explicitly asked for it; it is the reference's shorthand.

**Acceptance (proposed)**:

1. **Given** a runnable tellme installation, **When** the operator runs `tellme -v`, **Then** tellme prints the build version to `stdout` and exits successfully — identically to `--version`.

**Functional requirements (FR)**:

- **FR-004**: tellme MUST accept `-v` as a shorthand for `--version`; `-v` MUST produce the same output as `--version` and exit `0`.
- **FR-005**: `--version` MUST keep its existing behaviour (no regression).

---

### US3 — An unrecognized flag is still a usage error (Priority: P2)

The new flags must not weaken the usage-error contract.

**Why P2**: guard against a naive `-h`/`-v` addition swallowing the verdict.

**Acceptance (proposed)**:

1. **Given** the usage-error suite, **When** an unrecognized flag is passed, **Then** the run still refuses with the pinned phrase and the usage exit code (2).

**Functional requirements (FR)**:

- **FR-006**: An unrecognized flag MUST still refuse with `tellme: the command-line usage is invalid` and exit `2`.

**Non-functional requirements (NFR)**:

- **NFR-002**: The change MUST NOT introduce a new dependency, a network access, or a non-deterministic byte.

---

### Edge cases

- When `-h` is combined with a prompt or `-c`, help MUST still take precedence and exit 0 (a help request is terminal, like `--version`).
- When both `-h` and `-v` are given, the precedence MUST be deterministic (research decides).
- When help is piped (non-terminal stdout), the output MUST be the plain block (no terminal detection needed).

## Requirements

### Global requirements

#### Functional requirements

- **FR-007**: The dispatch precedence MUST place the help flag with the other terminal, prompt-less actions (`--version`/`-d`) — it must never fall through to a prompt turn.

### Key entities

- **Action flag**: a boolean flag that makes the process print something and exit (e.g. `--version`, `-d`, `-l`), as opposed to a turn-shaping flag (e.g. `-r`, `--new`).

## Success criteria

### Measurable outcomes

- **SC-001**: `tellme -h` prints the flag block and exits 0 (E2E carrier); `tellme -v` prints the version and exits 0 (E2E carrier).
- **SC-002**: The existing usage-error and `--version` scenarios all still pass unchanged.
- **SC-003**: `make verify` + `go test -count=1 ./...` are green; no `go.mod` change; the topology audit adds no new error.

## Assumptions

- **A1** — `-v` mirrors `--version` exactly.
- **A2** — the help block is the flag list (reference-shaped), not a command/exit-code essay.
- **A3** — help is a prompt-less successful action.
- **A4** — the frozen phrase vocabulary and the usage-error path are untouched (I-1/I-3).
