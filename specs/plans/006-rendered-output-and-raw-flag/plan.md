# System Analysis Plan — round 006 (`006-rendered-output-and-raw-flag`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/006-rendered-output-and-raw-flag/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── rendering-the-answer.feature
│       └── controlling-the-rendered-width.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (glamour renderer, -r flag, wrap width)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` or `data/**` truth artifacts in this round — see the `NOOP` notes for `/axb-api-plan` and `/axb-data-plan` below.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # entrypoint; supplies os.Stdin/Stdout/Stderr, the real TTY detector, and constructs the renderer (composition root). Single `-X main.version` target

internal/
├── cli/                           # dispatch (--version → -d → prompt turn → boot)
│   ├── cli.go                     #   CHANGED — adds the `-r`/`--raw` flag; threads the renderer seam + the stdout TTY probe;
│   │                              #   selects rendered (default) vs raw (`-r`) output; passes the resolved wrap width
│   └── exitcode.go                # unchanged this round (0/2/3/4/5/6)
├── ui/                            # NEW — output-rendering adapter (glamour)
│   └── renderer.go                #   Markdown→ANSI renderer behind a small seam; build options (WithStandardStyle +
│                                  #   GLAMOUR_STYLE + WithEmoji); LaTeX→Unicode sanitize; rendered/raw byte handling;
│                                  #   graceful degrade to raw on init failure (ADR-007 parity); WithWordWrap(width)
├── config/                        # CHANGED — add `WrapWidth` (`WRAP_WIDTH`) + `TELL_ME_WRAP_WIDTH` binding; validate `>= 0`
├── home/                          # unchanged this round
├── domain/llm/                    # unchanged this round (no provider request-shape change)
└── infrastructure/llm/openai/     # unchanged this round

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
├── harness/cmd_helper.go          # CHANGED — an ANSI-aware stdout predicate (rendered vs raw: literal-markdown-marker presence)
├── steps/                         # NEW step files for rendering, raw output, and wrap width
└── network_guard_test.go          # unchanged
go.mod / go.sum                    # CHANGED — adds github.com/charmbracelet/glamour (+ transitive lipgloss/termenv/reflow)
Makefile                           # unchanged (no new gate); `fmt`/`tidy`/`build`/`test`/`lint`/`vulncheck`/`verify` stand
```

**Structure Decision**: The change stays **inside the existing CLI surface** and adds **one new adapter package** for rendering. The renderer is an output adapter (glamour), so it is constructed at the **composition root** (`cmd/tellme`) and injected into `internal/cli` behind a small seam — following the round-003/004 pattern (an injected dependency so the pure render/raw mode selection and the wrap-width resolution stay unit-testable without a real terminal or a real glamour renderer). The provider path (`internal/domain/llm` + `internal/infrastructure/llm/openai`) is untouched; `internal/home` is untouched. `internal/config` gains the `WRAP_WIDTH` input field and its env binding. The E2E harness gains an ANSI-aware output predicate (today it asserts captured bytes; rendered-vs-raw must key on literal-markdown-marker presence/absence, not exact bytes). This is the project's **first presentation dependency** (`go.mod` grows), which is the round's deliberate, ratified departure.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface (round 006 changes how the answer is **displayed** on the existing end — it introduces **no** new end, **no** new outbound provider contract, and **no** persisted state):

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: **standard output** — the answer rendered as formatted Markdown by default, or printed verbatim as plain text under `-r`/`--raw`; the new `-r`/`--raw` flag; the `WRAP_WIDTH` configuration and `TELL_ME_WRAP_WIDTH` environment override for the rendered width; the **stdout terminal probe** gating tellme's own presentation (`UseColor = isTTY && !raw`). **Standard error** (frozen class phrases, incl. the degrade warning) and the exit-code table (`0`/`2`/`3`/`4`/`5`/`6`) are otherwise unchanged.
   - Requirement evidence: `FR-001`–`FR-009`, `NFR-001`–`NFR-003`; acceptance features `rendering-the-answer.feature`, `controlling-the-rendered-width.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own; round 006 does not change the provider request shape).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - `/axb-data-plan` = **`NOOP`** (round 006 introduces no persisted or in-memory state). `WRAP_WIDTH` is configuration **input**, not state — consistent with the ratified round-001 position (grill #3) that the config file is a CLI input contract owned by the CLI end, not a `data/**` artifact.
> - The provider remains an **external dependency reached outbound by the CLI**; its request/response contract is unchanged from round 004 and is not re-analysed here.
> - **Truth amendment carried to `/axb-dsl-refine`**: round 006 **amends round-005 FR-007** — under reference parity the answer is rendered regardless of stream, and the non-terminal suppression applies to tellme's **own** presentation only. The `piping-the-answer-out` interface feature's "redirected answer is the answer text alone / carries no decoration" Examples must move their byte-exact assertion to the `-r` path (a **MODIFY**).

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; no provider request-shape change.
  - **Data** → **`NOOP`**: no persisted or in-memory state introduced (`WRAP_WIDTH` is config input, not state).
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: extend `specs/truth/features/cli/**` with the output-presentation behaviour — default rendered output (Markdown→ANSI) on all streams; the `-r`/`--raw` flag printing the answer verbatim; the `WRAP_WIDTH`/`TELL_ME_WRAP_WIDTH` rendered width (rendered-only; `0` = default; negative rejected); the stdout terminal probe gating tellme's own presentation — and **MODIFY** the round-005 `chat/piping-the-answer-out` feature + DSL rows so the byte-exact/plain assertion is tied to `-r` (the FR-007 amendment).
- Scheduling rationale: there is a single interface, so no dependency ordering is needed — it settles entirely from the upstream sources (`spec.md` §US1–US3, `research.md` Decisions 1–7). Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a lone interface forms a single wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no persisted state).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the output-presentation behaviour (`rendering-the-answer`, `controlling-the-rendered-width`) plus the **round-005 `piping-the-answer-out` MODIFY**, in `specs/truth/features/cli/**` (module feature + DSL rows; interface-root `dsl.md` if any row becomes cross-module).

*Handoff payload (for the next phase)*: plan package `specs/plans/006-rendered-output-and-raw-flag`; truth root `specs/truth`; truth-delta `specs/plans/006-rendered-output-and-raw-flag/truth-delta.md`; interface `CLI end`; analysis focus as above; acceptance features `features/acceptance/rendering-the-answer.feature` + `features/acceptance/controlling-the-rendered-width.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the output-mode contract, the dependency posture, and the wrap-width scope; the byte-level parity details are settled by `/axb-technical-research` Decisions 1–7. The only known consequence — the round-005 FR-007 amendment — is a recorded, ratified truth `MODIFY`, not a blocker.)*
