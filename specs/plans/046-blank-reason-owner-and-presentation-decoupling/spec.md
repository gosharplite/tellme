# Feature Specification: blank-reason owner + the loop's presentation de-coupling (round 046)

**Feature Branch**: `046-blank-reason-owner-and-presentation-decoupling`

**Created**: 2026-09-18

**Status**: Draft — plan + truth half. Anchor issue [#108](https://github.com/gosharplite/tellme/issues/108) — **R4** of [#92](https://github.com/gosharplite/tellme/issues/92). Clarify round 1 (**C-R4-1 … C-R4-4**, below) **not yet locked** — see §Clarify.

**Input**: Issue [#108](https://github.com/gosharplite/tellme/issues/108) (R4 of the #92 gate-first split). #92's scope ledger item **#4**: *"Blank-reason predicate — single owner (3 sites, 1 dead) + drop the loop→`ui` predicate coupling"*, with #92 **AC6** (*"each policy has one named owner, with a unit witness on the real path"*). R1 ([#93](https://github.com/gosharplite/tellme/issues/93), round 042, ADR 0011) shipped the gate; R2 ([#100](https://github.com/gosharplite/tellme/issues/100), round 044, ADR 0013) ratcheted the baseline **8 → 1**; R3 ([#105](https://github.com/gosharplite/tellme/issues/105), round 045, ADR 0014) owned the **yield** axis and left the last entry deliberately. R4 **owns the last baseline entry** (`internal/agent -> internal/ui`, RULE-A) and the **blank-reason predicate**.

**Behaviour intent**: **MODIFY (structural — behaviour-preserving)**. Give the *"does this reason render a `[Tool Reason]` row?"* predicate **one owner**, evaluated **once** on the real path, and remove the loop's dependence on `internal/ui` so the layer-discipline baseline ratchets **1 → 0** — with **zero** change to any flag, exit code, stream contract, formatting, or the DSL vocabulary. The round's witness is **unit pins + the gate** (a green E2E suite alone is false confidence — the round-009 trap, per #92).

---

## Clarify (round 1 — C-R4-1 … C-R4-4; **not yet locked**)

> The candidates are enumerated on [#108](https://github.com/gosharplite/tellme/issues/108). The issue states the **problem**; the mechanism is chosen one question at a time.

| # | Candidate | Options |
| --- | --- | --- |
| **C-R4-1** | **The mechanism that removes the loop's `internal/ui` import** (`agentloop.go` uses it for 4 formatters **and** the predicate) | **(A)** inject a **domain-typed presentation port** (the four formatters + the predicate) implemented in `internal/ui`; the loop keeps its own `Stderr` writes + the yield bracket — minimal, **ADR 0014 untouched** · **(B)** the loop **emits semantic tool-line hooks** through `LoopObserver` and the CLI `call` half formats + writes via `internal/ui`; the loop keeps the yield bracket — **ADR 0014 untouched**, larger blast radius · **(C)** as (B) but the presenter owns the yield wrapping too — **amends ADR 0014** |
| **C-R4-2** | **The predicate's single owner + the dead site P3** | a: one evaluation on the real path + **delete** P3 (single owner) · b: one evaluation + keep P3 as documented defence-in-depth · c: other |
| **C-R4-3** | **The witness** | a: unit pins (hostile-fixture reason rows + the loop's line ordering) + the gate's **1 → 0**; **no** new E2E Example (weak-carrier precedent) · b: also an E2E Example |
| **C-R4-4** | **Governance** | a: a new **ADR 0015** recording the presentation-ownership rule (the loop emits semantics; the presenter owns formatting + the predicate), amending nothing · b: extend an existing ADR · c: no ADR |

---

## Grounded in the current system

Static read 2026-09-18 @ `dev` `1d36509` (recorded on [#108](https://github.com/gosharplite/tellme/issues/108)).

### The single illegal edge

`internal/agent/agentloop.go:19` imports `github.com/gosharplite/tellme/internal/ui`. The R1 gate ranks `internal/agent` = tier **4** and `internal/ui` = tier **5**, so this is an **upward** import → **RULE-A**. It is the **only** entry in `tools/arch/baseline.txt` — the ratchet's last tooth.

### The 3 predicate sites

| # | Site | Role | Note |
| --- | --- | --- | --- |
| **P1** | `agentloop.go` `reasonsOf` | filters each call's reason for the grouped post-call tail | the **production** filter; appends the **raw** value |
| **P2** | `agentloop.go` `logAction` | decides whether the call's begin block emits a `[Tool Reason]` line | the **production** filter for the begin line |
| **P3** | `cli/call_renderer.go` `OnCallEnd` → `emit` | re-checks each tail reason before printing | **DEAD** — P1 already filtered, so it never drops (a documented defence-in-depth leftover; round-036 review TD-1) |

The four **formatter** call sites (`logEngine` → `FormatToolEngine`; `logAction` → `FormatToolReason`, `FormatToolAction`; `logResult` → `FormatToolResult`) all sit inside `withToolLog`, which brackets them with the round-045 yield pair (`YieldIndicator()` / `RestoreIndicator()`).

### The ≈3×/call re-sanitize

For one call carrying a reason, `ui.toolReasonText` (the single transform, round-039 RF-1) runs **≥3 times**: P1's check, P2's check, and `FormatToolReason` — plus P3's check and the tail's own `FormatToolReason`. The transform is pure, so this is waste rather than a defect; but it is the evidence that the *policy* is **re-stated per site** instead of owned.

### Invariants that must survive

- The tool-line stream (`[Tool Engine]` / `[Tool Reason]` / `[Tool Action]` / `[Tool Result]`) and the round-039 **blank-line grouping** are **byte-identical**, in the same order, with the same yields around them.
- `ADR 0014`: the yield **policy** is owned by `ui.YieldController` and the port's `YieldIndicator`/`RestoreIndicator` pair is the loop's **route** — R4 must not silently undo it (a mechanism that moves the wrapping is an explicit **ADR amendment**).
- `ADR 0013`: the composition root (`cmd/tellme` + `cli.Options`) stays the only assembly site; `agentTools()` stays parameterless + read-free (round-033 FR-009).
- The gate's own doctrine (ADR 0011): the baseline **only shrinks**; a compliant code change and the baseline edit land **together**.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the blank-reason predicate has one owner (Priority: P1)

As a maintainer, I want *"does this reason render a `[Tool Reason]` row?"* answered in exactly one place and evaluated once per reason, so the rule cannot drift between the loop and the presenter (and the ≈3×/call re-run disappears).

**Why this priority**: it is the ledger item (#92 scope #4) and the AC6 obligation (*one named owner + a unit witness on the real path*).

**Independent verification**: the predicate has a **single** owning definition; every site that needs the answer consumes that owner's decision rather than re-deriving it; the dead site P3 no longer re-checks; and a hostile-fixture unit pin shows an empty / whitespace-only / escape-only / over-cap reason producing the same rendered outcome as today.

**Acceptance Scenarios**:

1. **Given** a call whose `reason` is `""`, `"   "`, `"\n"`, or escape-only, **When** the call's begin block and the grouped tail are rendered, **Then** no `[Tool Reason]` line appears for it (the blank-reason rule unchanged) and **no** other line's position or content changes.
2. **Given** a call whose `reason` needs folding (`\r`/`\n`) or capping, **When** it is rendered, **Then** the emitted bytes are exactly today's (fold → trim → sanitize → cap, one U+2026 inside the cap).
3. **Given** the round head, **When** the predicate's evaluations are counted for one tool call carrying a reason, **Then** the transform runs **once** per consumer decision, not ≥3 times.

**Functional Requirements**:

- **FR-001**: The blank-reason predicate MUST have **one owning definition**; the loop and the presenter MUST NOT each carry their own copy of the rule or of its transform.
- **FR-002**: The predicate MUST be evaluated on the **real path** (the loop's per-call path), not only at a dead/defensive site (`#92` AC6).
- **FR-003**: The **dead** site P3 (`cli/call_renderer.go` `OnCallEnd`) MUST NOT remain a duplicate authority — it is either removed (single owner) or explicitly re-expressed as a consumer of the owner (per **C-R4-2**).

**Non-Functional Requirements**:

- **NFR-001**: The consolidation MUST be **behaviour-preserving** — every rendered reason byte is identical, including the fold/trim/sanitize/cap chain and the blank-line grouping.

### User Story 2 - the loop no longer owns presentation (Priority: P1)

As a maintainer, I want `internal/agent` to stop importing `internal/ui`, so the layer-discipline ratchet reaches its terminal state (**0** violations) and the loop owns only the think→act→observe semantics.

**Why this priority**: it is the round's **falsifiable DoD** — the R1 gate proves it (#92 AC1/AC2/AC3, AC6).

**Independent verification**: `grep` shows no `internal/ui` import in `internal/agent/**` (production or test); `tools/arch/baseline.txt` is **header-only**; `make verify-architecture` is green with **0 new / 0 stale / 0 cycles**.

**Acceptance Scenarios**:

1. **Given** the round head, **When** the R1 gate enumerates the module graph, **Then** it reports **0** violations and the baseline file holds **no** entry line.
2. **Given** a regression that re-introduces an `internal/agent -> internal/ui` import, **When** the gate runs, **Then** it **fails** (a new violation) — the ratchet has teeth at 0 (the anti-bypass rule, round-042 review F-1).
3. **Given** the loop's diagnostic path, **When** a tool round runs, **Then** the same four line kinds appear on `stderr`, in the same order, with the same yields around them.

**Functional Requirements**:

- **FR-004**: `internal/agent/**` MUST import **no** `internal/ui` package (production or test files), and the round MUST ship the corresponding `tools/arch/baseline.txt` shrink (**1 → 0**) in the same commit-range (the gate ships with its enabler).
- **FR-005**: The mechanism MUST be chosen per **C-R4-1**; if it moves any yield wrapping, that is recorded as an explicit **ADR 0014 amendment**, never a side effect.
- **FR-006**: No new layer violation may be introduced anywhere (e.g. the loop must not import `internal/cli`; `internal/domain/**` stays 100% pure).

**Non-Functional Requirements**:

- **NFR-002**: The port/seam that replaces the direct `internal/ui` use MUST be **network-free** and **presentation-free at the domain tier** (no `internal/ui` types in `internal/domain/**`), per ADR 0005 D1/D2 and ADR 0013.
- **NFR-003**: The gate's own invocation contract MUST be honoured (`make verify` / `verify-architecture` green: `go vet -tags=arch`, the `-count=1` gate, 0 new / 0 stale / 0 cycles).

### User Story 3 - zero behavioural change (Priority: P2)

As a maintainer, I want the round to be a **pure structural refactor**: under the same inputs, every flag, exit code, stream byte, and formatting outcome is identical, and the full regression is green.

**Why this priority**: it bounds the round to structure and protects the #92 *"no user-facing change"* constraint.

**Independent verification**: `make verify` green (incl. the R1 gate); godog E2E green; `stdout` byte-exactness unchanged; DSL vocabulary unchanged; topology audit unchanged.

**Acceptance Scenarios**:

1. **Given** a supported prompt path (tool-using and tool-less), **When** the round head runs it, **Then** `stderr` and `stdout` are byte-identical and every exit code is unchanged.
2. **Given** the full `make verify` + E2E, **When** it runs, **Then** it is green (no behavioural/stream-contract regression).
3. **Given** the R1 gate at 0, **When** it runs, **Then** it stays green at the terminal state.

**Functional Requirements**:

- **FR-007**: The round MUST NOT change any user-facing CLI behaviour: no flag, exit code, stream contract, formatting, DSL row, or stored record.
- **FR-008**: The round's witness MUST be **unit pins + the gate**, NOT the E2E suite (per **C-R4-3**), and MUST include **falsifiability witnesses** reproduced then reverted (ADR 0010) — at minimum: (a) re-introducing the `internal/agent -> internal/ui` import ⇒ the gate reds; (b) un-single-owning the predicate (re-adding a second evaluation with a divergent rule) ⇒ the hostile-fixture pin reds; (c) a stale baseline line at 0 ⇒ the gate reds (the anti-bypass rule).

**Non-Functional Requirements**:

- **NFR-004**: `make verify` (incl. `verify-architecture`, `verify-mcp-sdk-confinement`, 4/4 cross-compile, lint, govulncheck), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; **no** new Gherkin/DSL row (`/axb-dsl-refine` **NOOP**).

---

## Requirements *(mandatory)*

### Global Requirements

#### Functional Requirements

- **FR-009**: The round MUST record the ownership decision in an **ADR** (`docs/decisions/0015-*.md` + the `docs/decisions/README.md` index row) per **C-R4-4** — the rule: *the agent loop emits semantics; the presenter owns the tool-line formatting and the blank-reason predicate* — and MUST state its relation to ADRs 0005 (tool-call-log parity, D1 partitions *rendering*), 0013 (composition-root injection), and 0014 (yield-policy owner).
- **FR-010**: `specs/truth/techstack.md` MUST be updated for the changed seam (`truth-current`) — **MODIFY**, not NOOP; the affected rows name the loop's tool-line rendering and the `internal/ui` coupling (e.g. **Agent tool loop**, and any row naming `internal/agent`'s presentation use).
- **FR-011**: Scope guard — the round MUST touch only: `internal/agent/**`, the chosen seam's home (`internal/domain/agent/**` and/or `internal/cli/**`), `internal/ui/toolcall*.go` (only if the chosen mechanism relocates the predicate/formatters), the affected `_test.go` files, `tools/arch/baseline.txt` (**shrink only**), `docs/decisions/**`, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's `session-summary.md`. It MUST NOT change adapter behaviour, domain business logic, flags, exit codes, or stream contracts.

#### Non-Functional Requirements

- **NFR-005**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.
- **NFR-006**: **POSIX-only** (the repo is bash/POSIX-only); no Windows variant.

### Key Entities

- **The blank-reason predicate** — *"does this `reason` render a `[Tool Reason]` row?"*; today `ui.ToolReasonRenders` (derived from the single transform `ui.toolReasonText`), evaluated at P1/P2/P3.
- **The four tool-line formatters** — `ui.FormatToolEngine/FormatToolReason/FormatToolAction/FormatToolResult` (the loop's other `internal/ui` use).
- **`agentport.LoopObserver`** — the loop's observer port (round 019/034/045); extended or reused by the chosen mechanism.
- **`tools/arch/baseline.txt`** — the RULE-A/B/C ratchet; goes **1 → 0** this round.

### Truth obligations (explicit — `truth-current`)

- The **Agent tool loop** row (`specs/truth/techstack.md`) names the loop's diagnostic rendering today; under R4 the loop no longer formats — that row is a real **MODIFY**.
- Any row naming the `internal/agent -> internal/ui` coupling (e.g. the *Not Introduced Yet* / presentation notes) MUST be reconciled.

`/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted/runtime state change); `/axb-dsl-refine` = **NOOP** (no feature/DSL change).

## Success Criteria *(mandatory)*

- **SC-001**: The blank-reason predicate has one owning definition; the loop and the presenter do not each carry a copy; the transform runs **once** per consumer decision. (covers FR-001, FR-002, FR-003)
- **SC-002**: `internal/agent/**` imports **no** `internal/ui`; `tools/arch/baseline.txt` is **header-only**; the R1 gate is green (**0 new / 0 stale / 0 cycles**). (covers FR-004, FR-006, NFR-003)
- **SC-003**: The chosen mechanism is implemented and its yield-route relation to **ADR 0014** is recorded (amendment or invariance). (covers FR-005)
- **SC-004**: Full regression green with **zero** behavioural / stream-contract change — `make verify`, godog E2E, exit codes, `stderr`/`stdout` byte-exactness, DSL vocabulary, topology audit. (covers FR-007, NFR-004)
- **SC-005**: Falsifiability witnesses (a)/(b)/(c) reproduced and reverted. (covers FR-008)
- **SC-006**: The **ADR** + its index row are present and `specs/truth/techstack.md` carries the MODIFY row(s); `go.mod`/`go.sum` unchanged. (covers FR-009, FR-010, NFR-005)

## Edge Cases

- **A reason that is absent from the arguments JSON** → `toolReason` returns `""` → no `[Tool Reason]` line (unchanged).
- **A reason that sanitizes to nothing** (escape-only) → renders no line (unchanged); the sanitizer's window semantics are untouched (ADR 0007/0008).
- **An over-cap reason** → rune-safe cap with one U+2026 inside the cap (unchanged).
- **A tool-less turn** → no tool-line hooks fire, no blank is added (round-039 B1 preserved).
- **The final call's deferred tail** → still deferred past the answer; nothing renders at its call-end (round-035 / ADR 0005 G5 preserved).
- **A gated-off spinner (nil presenter)** → the mechanism's injected seam is nil-safe, so the turn renders without an indicator (round-019 FR-006 preserved).
- **A future second consumer of the predicate** → it must consume the owner, not re-derive the rule (the ADR records the owner).
- **The baseline at 0** → a *stale* baseline line still fails; an emptied baseline with violations still fails (the anti-bypass rule preserved).

## Assumptions

- **A1 (weak E2E carrier)** — the round changes no `tellme` CLI behaviour, so the E2E suite is a **weak acceptance carrier**; the witness is **unit pins + the gate** (FR-008). `/axb-spec-by-example` is therefore **NOOP** (no new user-facing business journey) — precedent: rounds 020/031/036/041/042/043/044/045.
- **A2 (mechanism is RD)** — the exact seam shape, where it lives, and the test-harness placement are **RD** decisions (`research.md` D-x), not FRs.
- **A3 (ADR required)** — an ADR records the ownership rule (FR-009); the exact number (candidate **0015**) is confirmed in research.
- **A4 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**; `/axb-ui-plan` **skipped** (not user-facing).
- **A5 (`/axb-dsl-refine` NOOP)** — no feature Rule, Example, step, or `DSLRow` change (the topology audit stays unchanged); the acceptance carrier is the unit pins + the gate.
- **A6 (scope guard)** — see FR-011.
- **A7 (R5 relation)** — R5 ([#101](https://github.com/gosharplite/tellme/issues/101)) remains the *deeper* `internal/cli` strict de-coupling; R4 closes only the `internal/agent -> internal/ui` edge.

## Out of scope (recorded)

- **R5** ([#101](https://github.com/gosharplite/tellme/issues/101)) — `internal/cli` imports `domain` + stdlib + sanctioned application utilities only, and the PR #104 review deferrals **F-4/F-6/F-7/F-8**.
- The **ride-alongs** (suggestion-selection `set(items, cursor)`; `NewCommandTool(sink)` ctor injection) — parked.
- The `#92` **records**: the permanent E2E narrowing (round 036), the one-concurrent-block limit + the `End`-while-write-stalled residual (round 040).
- Changing user-facing CLI behaviour, flags, exit codes, or formatting.
- Re-architecting the loop/coordinator/observer beyond the presentation edge; the yield **policy** (ADR 0014) is R3's, not re-opened.
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#92](https://github.com/gosharplite/tellme/issues/92) scope ledger **#4**; **AC1/AC2/AC3/AC6**.
- [#108](https://github.com/gosharplite/tellme/issues/108) — the round's anchor (grounded inventory + the clarify candidates).
- [#93](https://github.com/gosharplite/tellme/issues/93) → round 042 (ADR 0011) — the gate; [#100](https://github.com/gosharplite/tellme/issues/100) → round 044 (ADR 0013) — 8 → 1; [#105](https://github.com/gosharplite/tellme/issues/105) → round 045 (ADR 0014) — the yield owner, `C-R3-4` deferred 1 → 0 to R4.
- PR [#75](https://github.com/gosharplite/tellme/pull/75) (round 036) — the blank-reason suppression predicate + the dead third site; PR [#81](https://github.com/gosharplite/tellme/pull/81) (round 039) — the loop→`ui` predicate + the ≈3×/call re-sanitize.

## References

- [`internal/agent/agentloop.go`](https://github.com/gosharplite/tellme/blob/dev/internal/agent/agentloop.go) — the single `internal/ui` import; `logEngine`/`logAction`/`logResult`/`withToolLog`/`reasonsOf`.
- [`internal/ui/toolcall.go`](https://github.com/gosharplite/tellme/blob/dev/internal/ui/toolcall.go) — `toolReasonText`, `ToolReasonRenders`, the four formatters.
- [`internal/cli/call_renderer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/call_renderer.go) — P3 (the dead tail guard).
- [`internal/domain/agent/observer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/domain/agent/observer.go) — the observer port.
- [`internal/cli/composite_observer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/composite_observer.go) — the loop's single observer; the yield route.
- [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) · [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) — the gate + the last baseline entry.
- [`docs/decisions/0005-tool-call-log-parity.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0005-tool-call-log-parity.md) (D1 partitions *rendering*) · [`0011`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) · [`0013`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0013-composition-root-injection.md) · [`0014`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0014-yield-policy-owner.md).
