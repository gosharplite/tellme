# System Analysis Plan — round 011 (`011-persona-and-payload-estimate`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/011-persona-and-payload-estimate/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── sending-the-configured-persona.feature
│       └── estimating-the-payload-that-will-be-sent.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (persona + estimator inputs)
├── data/**                        # /axb-data-plan — NOOP this round (no persisted-state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/{*.feature, dsl.md}   # MODIFY — persona request contract + wire-faithful estimate
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **`NOOP`**: no persisted state changes — the session-history model is untouched.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── cli/
│   ├── cli.go                     # CHANGED — carry the resolved `PERSON` on the resolution; build the
│   │                              #   gateway with the persona; compute the pre-flight estimate over the
│   │                              #   WIRE payload (persona + tool declarations + messages)
│   └── turn_test.go               # EXTENDED — the pre-flight estimate reflects persona + tools + messages
├── config/
│   └── config.go                  # unchanged — `PERSON` is already modeled (`Config.Person`)
├── domain/
│   └── llm/
│       ├── token.go               # CHANGED — the estimator takes the wire payload (persona + tool
│       │                          #   declarations + messages); still a dependency-free heuristic
│       ├── gateway.go             # possibly widened — the provider port carries the persona shape used
│       │                          #   by the transport (client-level); exact seam a research/task detail
│       └── token_test.go          # EXTENDED — deterministic + responsive over an explicit payload
├── infrastructure/
│   └── llm/
│       ├── factory.go             # CHANGED — thread the resolved persona into the adapter (gateway-level)
│       └── openai/
│           ├── client.go          # CHANGED — emit a leading `system` message when the persona is set
│           └── client_test.go     # EXTENDED — the leading system message in the assembled request body

internal/agent/agentloop.go        # unchanged under gateway-level injection — the loop keeps using the
                                    #   same gateway instance for the main and tool-driven completions
                                   #   (no new domain port, no adapter, no dependency)

tests/e2e/
├── harness/cmd_helper.go          # CHANGED — expose the fake provider's recorded request (already recorded
│                                  #   since round 007) for the leading `system`-message assertion
└── steps/                          # NEW/CHANGED step files — persona-on-the-wire; wire-faithful estimate

specs/truth/features/cli/chat/
├── answering-a-single-prompt.feature   # MODIFY — the request now carries a leading system message
├── reporting-the-payload-status.feature # MODIFY — the estimate reflects the wire payload
└── dsl.md                               # MODIFY — persona request rows + wire-payload estimate rows

go.mod / go.sum                     # unchanged — stdlib-only (no tokenizer); no new dependency
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 011 changes **request content + the estimator's inputs**, not the system's ends. The persona is injected at the **transport** level (`internal/infrastructure/llm`) so a single gateway construction covers the main completion and every tool-driven completion — matching the reference's client-level injection — while the CLI carries the resolved `PERSON` on its resolution so the same value feeds both the gateway and the pre-flight estimate. The estimator (`internal/domain/llm/token.go`) is **widened in inputs** (persona + tool declarations + messages) but stays a dependency-free heuristic. There is **no** new domain port, adapter, persisted-state change, config field (`PERSON` already exists), or third-party dependency — consistent with `research.md` Decisions 1–8. The **request contract** and the **estimate** are pinned in the CLI interface truth (`specs/truth/features/cli/chat/**`) by `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface. Round 011 does not introduce a new system end; it changes the **CLI end**'s outbound request content and its pre-flight estimate.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the operator's configured `PERSON` is sent to the provider as a **leading `system` message** ahead of the conversation, on **every** request a prompt run makes (the main completion and any tool-driven completion); an empty `PERSON` sends no message, and the persona is request-only (never on `stdout`). The per-turn **pre-flight estimate** is computed over the **wire payload** — the persona message + the tool declarations + the conversation messages — so it is comparable to the provider's measured count. The answer stream (`stdout`) stays **byte-exact**; the payload-status **line format**, the measured line's source (`usage.prompt_tokens`), the frozen `tellme: {phrase}` class-phrase vocabulary, and the exit-code table (`0/2/3/4/5/6/7`) are **unchanged**.
   - Requirement evidence: `FR-001`–`FR-013`, `NFR-001`–`NFR-005`; acceptance features `sending-the-configured-persona.feature`, `estimating-the-payload-that-will-be-sent.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own). The outbound OpenAI-compatible request shape is described in `techstack.md` (a CLI-transport concern), not in an authored `contracts/**` artifact.
> - `/axb-data-plan` = **`NOOP`** — no persisted state changes; the session-history model (`history_entry`) is untouched; the token counts and the estimate remain display-only and are recomputed (research Decision 5).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **estimator** and the **persona injection** are **not** separate system interfaces: they are behaviour of the **CLI end** (its outbound request and its pre-flight diagnostic). Their executable contract is carried forward to `/axb-dsl-refine`.
> - The **provider** remains an **external dependency reached outbound by the CLI**; round 011 changes the request **content** the CLI authors (adds the persona `system` message) but introduces no new tellme-authored external contract beyond that.
> - **Truth amendments carried to `/axb-dsl-refine`**: **MODIFY** `specs/truth/features/cli/chat/answering-a-single-prompt.feature` so the request contract asserts the **leading `system` message** (and the existing "carries exactly the current user prompt / no earlier exchange" assertion is reworded to account for it); **ADD/MODIFY** a `chat` feature+DSL rows asserting the **wire-faithful estimate** (persona + tool declarations + messages; deterministic; responsive); and confirm **every round-011 acceptance rule is carried** by an interface feature (`acceptance-coverage`). The root `cli/dsl.md` class-phrase vocabulary stays **10**.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound request content is a CLI-transport concern recorded in `techstack.md`.
  - **Data** → **`NOOP`**: no persisted-state change; the estimate is display-only and recomputed.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin — (a) the persona request contract (leading `system` message; empty ⇒ none; rides every request) and (b) the wire-faithful estimate (persona + tool declarations + messages; deterministic; responsive) — as mechanically assertable interface Rules, while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
- Scheduling rationale: there is **one** interface and the behaviour is settled by `research.md` Decisions 1–8 and Clarify Round 1, so there is nothing to sequence; a single wave suffices. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a sole interface forms one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → pin the persona request contract and the wire-faithful estimate in `specs/truth/features/cli/chat/**`, and confirm `acceptance-coverage` for both round-011 acceptance features.

*Handoff payload (for the next phase)*: plan package `specs/plans/011-persona-and-payload-estimate`; truth root `specs/truth`; truth-delta `specs/plans/011-persona-and-payload-estimate/truth-delta.md`; interfaces `CLI end`; analysis focus as above; acceptance features `features/acceptance/sending-the-configured-persona.feature` + `features/acceptance/estimating-the-payload-that-will-be-sent.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the estimate scope (persona + tool declarations + messages) and the guarantee (inputs + determinism; no numeric equality); `research.md` Decisions 1–8 settled the persona shape/placement, the transport-level injection, the empty-persona behaviour, the estimator's widened inputs, the verification surface, and the no-dependency posture. **Open (non-blocking):** the estimation heuristic constants, the persona-plumbing seam, the existing-truth rewrite wording, and the persona-assertion granularity — all held as research/task/DSL-level defaults, none gating this round.)*
