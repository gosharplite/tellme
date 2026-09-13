# System Analysis Plan — round 012 (`012-interactive-multiline-prompt`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/012-interactive-multiline-prompt/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── capturing-a-multi-line-prompt.feature
│       └── cancelling-an-interactive-prompt.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (interactive prompt read)
├── data/**                        # /axb-data-plan — NOOP this round (no persisted-state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/{usage,chat}/…         # MODIFY/ADD — the interactive prompt-read contract
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **`NOOP`**: no persisted state changes — the session model is untouched.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── cli/
│   ├── cli.go                     # CHANGED — on the empty-prompt path with a TTY stdin (and no
│   │                              #   terminal-less subcommand), print the multi-line hint to stderr
│   │                              #   and read the prompt to EOF, bounded 1 MiB; empty/cancel ⇒ no
│   │                              #   request. Reuses the injected isTTY seam + the SIGINT context
│   ├── prompt.go                  # possibly extended — the interactive read helper (bounded io.ReadAll)
│   └── *_test.go                  # EXTENDED — the interactive read driven through the injected isTTY
│                                  #   seam (hint to stderr; read to EOF; empty/cancel ⇒ no request)
└── ...                            # unchanged — no new domain port, adapter, config field, or dependency

tests/e2e/
├── harness/cmd_helper.go          # possibly extended — feed a piped stdin (already supported) for the
│                                  #   negative assertion; no pty
└── steps/                          # NEW step file — the negative assertion (a piped run prints no
                                   #   reading announcement)

specs/truth/features/cli/
└── {chat,usage}/…                 # MODIFY/ADD — the interactive prompt-read contract + DSL rows

go.mod / go.sum                     # unchanged — stdlib-only; no new dependency
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 012 adds one **CLI input capability**. The interactive read is implemented in `internal/cli` behind the **existing** injected terminal-detection seam (`runtimeEnv.isTTY`), reusing the round-005 bounded-read pattern (`io.ReadAll`/`io.LimitReader`, 1 MiB) and the round-005 SIGINT context for cancellation. It introduces **no** new domain port, adapter, config field, persisted-state change, or third-party dependency. Because the E2E harness runs the **built binary** sub-process (pty-less), the interactive branch is verified at the **unit** layer through the injectable seam; the E2E suite asserts only the observable **negatives** (a piped run prints no reading announcement; the pipe/positional/non-prompt paths are byte-identical to prior rounds). The **executable contract** is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface. Round 012 does not introduce a new system end; it adds an **interactive prompt-read capability** to the **CLI end**.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: when `tellme` resolves to an **empty prompt**, with **no positional arguments and no piped stdin**, and stdin is a **terminal**, the CLI prints the multi-line hint (`[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]`) to **`stderr`** and reads the prompt from stdin to **EOF** (`Ctrl+D`), bounded at 1 MiB; the captured (trimmed) text is the prompt of exactly one reasoning turn — the same downstream as a positional/piped prompt. `Ctrl+C` cancels, and an empty/whitespace submission sends **no request**. **POSIX-only** (no Windows variant). `stdout` stays **byte-exact**; the piped/positional paths, the frozen `tellme: {phrase}` vocabulary, the exit-code table (`0/2/3/4/5/6/7`), the payload-status line, and the cross-stream ordering contract are **unchanged**.
   - Requirement evidence: `FR-001`–`FR-011`, `NFR-001`–`NFR-005`; acceptance features `capturing-a-multi-line-prompt.feature`, `cancelling-an-interactive-prompt.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own).
> - `/axb-data-plan` = **`NOOP`** — no persisted state changes; the session-history model is untouched.
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **interactive reader** is **not** a separate system interface: it is behaviour of the **CLI end** (its input handling). Its executable contract is carried forward to `/axb-dsl-refine`.
> - **Verification seam**: because the E2E harness is pty-less (round 005's named pin; reaffirmed round 006), the **positive** interactive read is asserted at the **unit** layer via the injected terminal seam; the E2E suite carries the **negative/unchanged** facts (a pipe prints no reading announcement; pipe/positional/non-prompt paths unchanged). The positive path is a **named pin** in the executable truth, mirroring round 005.
> - **Truth amendments carried to `/axb-dsl-refine`**: **ADD/MODIFY** a `cli` feature (likely `usage` or `chat`) + `dsl.md` rows pinning — the interactive read (how invoked, read to EOF, bounded), the hint on `stderr` (and absent on a pipe), and the empty/cancel contract (no request; exit `0`); and confirm **every round-012 acceptance rule is carried** by an interface feature (`acceptance-coverage`). The root `cli/dsl.md` class-phrase vocabulary stays **10**.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own.
  - **Data** → **`NOOP`**: no persisted-state change.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the interactive prompt-read contract (invocation, read-to-EOF, hint on `stderr`, empty/cancel ⇒ no request) and the observable negatives (a pipe prints no announcement), while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
- Scheduling rationale: there is **one** interface and the behaviour is settled by `research.md` Decisions 1–8 and Clarify Round 1, so there is nothing to sequence; a single wave suffices. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a sole interface forms one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → pin the interactive prompt-read contract and the observable negatives in `specs/truth/features/cli/**`, and confirm `acceptance-coverage` for both round-012 acceptance features.

*Handoff payload (for the next phase)*: plan package `specs/plans/012-interactive-multiline-prompt`; truth root `specs/truth`; truth-delta `specs/plans/012-interactive-multiline-prompt/truth-delta.md`; interfaces `CLI end`; analysis focus as above; acceptance features `features/acceptance/capturing-a-multi-line-prompt.feature` + `features/acceptance/cancelling-an-interactive-prompt.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the invocation (bare no-prompt TTY), the hint stream (`stderr`), and the empty/cancel contract; `research.md` Decisions 1–8 settled the seam, the bounded read, the POSIX-only posture, and the verification approach (unit seam + E2E negatives + named pin). **Open (non-blocking):** the reader's exact dispatch insertion point, the hint colour/gating, the empty-EOF exit code, and the optional `TELL_ME_FORCE_STDIN_TTY` fallback seam — all held as research/DSL/task-level defaults, none gating this round.)*
