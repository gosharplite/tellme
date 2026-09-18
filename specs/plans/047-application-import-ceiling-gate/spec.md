# Feature Specification: application import-ceiling gate (RULE-E) + application-tier import baseline (round 047)

**Feature Branch**: `047-application-import-ceiling-gate`

**Created**: 2026-09-18

**Status**: Draft — plan package created by `/axb-specify`. Anchor issue [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)). Clarify round 1 **locked** (**Q1 → A** a gate-first slice; **Q2 → A** a normative application-tier allow-list in the gate's tier table, default-deny, fail-on-stale allow-list; **Q3 → A** RULE-E binds **both** application tiers; **Q4 → A** the sanctioned set = `domain` + stdlib + `config` + `home` + `app/**`; **Q5 → A** ADR `0016` + slug `047-application-import-ceiling-gate`) — see §Locked decisions.

**Input**: Issue [#101](https://github.com/gosharplite/tellme/issues/101) — **R5** of the [#92](https://github.com/gosharplite/tellme/issues/92) gate-first split. #92 **AC2** second clause: *"`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities."* #101 records that stronger ambition (R2 mechanised only the **first** clause — the 7 RULE-B edges) and names its **precondition**: *"the gate needs a **new rule** (a candidate **RULE-E**: an allow-list for the application tiers' downward imports) **plus a new ADR** adjudicating it, shipped **with its enablers** (round-040 TD-1 lesson)."* R5 ordinal is a placeholder; this round is **R5.1** (the gate + its enablers), the first slice of the R5 programme.

**Behaviour intent**: **ADD** a new **RULE-E** rule to the existing layer-discipline guard (`tools/arch`, ADR 0011) — an **application-tier import ceiling**: an application tier (`internal/app/**`, `internal/cli`) may import, beyond stdlib, only `internal/domain/**`, `internal/config`, `internal/home`, and `internal/app/**` — plus the committed **baseline** of the not-yet-removed residual edges (a ratchet). It is a **tooling/gate slice** in the round-042 (`verify-architecture`) lineage: it changes **no** user-facing behaviour and **no** product code — it adds a rule, its baseline, its truth record, and its ADR.

---

## Locked decisions (clarify round 1 — Q1 … Q5)

> Locked one at a time; the ranked agenda is enumerated in the tasking. The issue states the **problem** and the **precondition**; the mechanism is chosen here.

| # | Decision |
| --- | --- |
| **Q1 → (A) gate-first slice** | Round 047 ships **RULE-E + ADR 0016 + a committed ratchet baseline** (tooling/truth only, **zero product code**); the *de-coupling* of the residual edges happens in **later slices**, each shrinking the baseline. Mirrors round **042** (gate + 8-entry baseline, no product code). #101's own precondition (*rule + ADR shipped with its enablers*) is satisfied by landing RULE-E **with a baseline** so `dev` stays green. Rejected: **(B)** full de-coupling in one round (R5 is "a multi-round programme, not a slice" — #101) and **(C)** rule + easy seams now (mixes gate work with behavioural-adjacent refactors, weakening the behaviour-free review surface). |
| **Q2 → (A) normative allow-list in the tier table** | RULE-E's sanctioned set is a **normative** section of the gate's embedded **tier table** (ADR 0011 **D7** — the single normative ranking source), **default-deny** for an unsanctioned application-tier internal import. Not-yet-removed residues live in the **existing** committed `tools/arch/baseline.txt` ratchet (they can only shrink). **Fail-on-stale applies to allow-list entries too**: a sanctioned entry that the application tiers no longer use **fails**, forcing the list to shrink to truth. Rejected: **(B)** a separate committed manifest (a second normative source — D7 tension) and **(C)** edge-level hardcoded allow-list (no machine-readable home). |
| **Q3 → (A) bind both application tiers** | RULE-E binds **`internal/app/**` (tier 2) and `internal/cli` (tier 6)** — the same target set as RULE-B, matching #101's own RULE-E sentence ("the application tiers' downward imports"). Consequence: the `internal/app/** → internal/config` edge is in scope for the adjudication (Q4 sanctions it). Rejected: **(B)** `internal/cli` only (the literal AC2 reading — narrower, and would make RULE-E and RULE-B govern different sets). |
| **Q4 → (A) sanctioned set as measured** | **Sanctioned** (permanently allowed by RULE-E): `internal/domain/**` + stdlib + **`internal/config`** + **`internal/home`** (ADR 0011 D1 tier-1 shared utilities) + **`internal/app/**`** (tier-2 application utilities — `deps` + `suggestions`). **Residual** (baselined now, inverted in later slices): `internal/cli → internal/agent`, `internal/cli → internal/ui`, `internal/cli → internal/ui/tui/prompt`. Matches #92 AC2 verbatim (*"`domain/*`, stdlib, and application utilities"*). Rejected: **(B)** the tighter set (residualise `config`/`home`/`suggestions` too) — larger programme, and inconsistent with the already-pinned tier-1 utility ranking. |
| **Q5 → (A) round identity** | ADR number **`0016`** (confirmed next free — `0015` is the latest indexed); package + branch slug **`047-application-import-ceiling-gate`**; **F-4/F-6/F-7/F-8** stay deferred (they belong to the de-coupling slices, recorded on [#101](https://github.com/gosharplite/tellme/issues/101)); [#101](https://github.com/gosharplite/tellme/issues/101) stays **OPEN** (047 is slice R5.1), its body refreshed at closeout to track the slice chain; only closes when AC1–AC5 land. |

**Mechanism note (RD detail, `research.md`).** RULE-E is added to the **existing** `tools/arch` guard (`//go:build arch`) — **no new Makefile target**; `make verify` already includes `verify-architecture`. The rule reuses the proven machinery: module-root-anchored `go list` (ADR 0011 **D4**), the `CROSS_TARGETS` union + filtered child env (**D5**), the merged production+test graph for the rule and the production-only SCC pass (**D8**), the deterministic sorted `baseline.txt` (**D9**), and the fail-on-stale ratchet (**D3**). The exact table shape, the self-test coverage assertion, and the guard file layout are RD decisions.

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `b42f868` (static import scan; the gate re-measures at implementation).

### The sanctioned set and today's residual edges

Every `internal/**` import of the two application tiers, by package edge:

| Application tier | `internal/**` import | RULE-E verdict |
| --- | --- | --- |
| `internal/cli` | `internal/domain/{agent,history,llm,tools,metrics}` | sanctioned (domain) |
| `internal/cli` | `internal/config`, `internal/home` | sanctioned (tier-1 utilities) |
| `internal/cli` | `internal/app/deps`, `internal/app/suggestions` | sanctioned (tier-2 application utilities) |
| `internal/cli` | `internal/agent` | **RESIDUAL** (RULE-E) |
| `internal/cli` | `internal/ui` | **RESIDUAL** (RULE-E) |
| `internal/cli` | `internal/ui/tui/prompt` | **RESIDUAL** (RULE-E) |
| `internal/app/**` | `internal/domain/{history,llm,metrics,suggestions,tools}` | sanctioned (domain) |
| `internal/app/**` | `internal/config` | sanctioned (tier-1 utility) |

⇒ **RULE-E baseline = 3** distinct package edges, all `internal/cli → {internal/agent, internal/ui, internal/ui/tui/prompt}`. The `internal/cli` **test** files import `internal/ui` only — the same package edge, so they add **no** new entry (consistent with the existing gate's production+test merged graph). `internal/app/**` has **no** residual edge.

### The pre-existing baselines

`tools/arch/baseline.txt` is **header-only (0 entries)** today: RULE-A/B/C report **0 new / 0 stale / 0 cycles** (rounds 042–046 ratcheted 8 → 0). RULE-E is a **new rule**, so its 3 residual edges are newly recorded entries — the total committed baseline becomes **3** (all RULE-E), ratcheting to 0 across the later R5 slices.

### The rule the round extends (ADR 0011)

- **D1** ranks the tiers: 0 `domain` · 1 `config`/`home` · 2 `app` · 3 `infrastructure` · 4 `agent` · 5 `ui`/`ui/tui/**` · 6 `cli`; `cmd/**`, `tests/**`, `tools/**` exempt.
- **D2** pins the two-part predicate (RULE-A direction · RULE-B application target rule · RULE-C domain purity · RULE-D default-deny).
- **D3** makes the baseline a **fail-on-stale ratchet**: a new violation fails; a stale entry fails; removing an entry for an unfixed violation fails.
- **D7** makes the gate's embedded **tier table** the **single normative ranking source**; a self-test asserts the table classifies every governed package.

RULE-E is a **fifth rule** (an application-tier import **ceiling**), mechanically distinct from RULE-A/B/C/D (a **direction** rule; a **target** rule; a **purity** rule; a **coverage** rule): it is an **allow-list ceiling** on downward imports — the ambition RULE-B could not express.

### Invariants that must survive

- **The gate stays green on `dev` at delivery** (RULE-E lands with its 3-entry baseline; RULE-A/B/C stay 0; 0 cycles).
- **The sanctioned set is one normative source** (the tier table, D7); prose cites it, never re-states it.
- **No product code** changes; **no** Gherkin/DSL row changes; **no** user-facing behaviour (flags, exit codes, streams).
- The guard must not itself introduce a governed violation (it lives in the exempt `tools/**`).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - an unsanctioned application-tier import fails the standard gate (Priority: P1)

As a maintainer, I want `make verify` to fail when an application tier (`internal/app/**` or `internal/cli`) imports an internal package outside the sanctioned set (`domain`, `config`, `home`, `app/**`), so the [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** second clause is enforced mechanically instead of by review.

**Why this priority**: it is the round's entire reason to exist (#101's precondition) and the falsifiable prerequisite for the later de-coupling slices.

**Independent verification**: add a *new* unsanctioned import to a governed application tier (e.g. `internal/cli → internal/infrastructure/mcp`, or a fresh `internal/cli → internal/telemetry`), run `make verify` — it exits non-zero and names `internal/cli -> <import>`; revert — exits 0.

**Acceptance Scenarios**:

1. **Given** the current tree, **When** `make verify` runs, **Then** RULE-E runs, reports **0** violations beyond the baseline, and the aggregate exits 0.
2. **Given** a change that adds a new unsanctioned import to `internal/cli` or `internal/app/**`, **When** `make verify` runs, **Then** the gate exits non-zero and names the offending `source -> import`.
3. **Given** an `internal/**` package matching **no** tier, **When** the gate runs, **Then** it **fails** (default-deny, RULE-D preserved).
4. **Given** a governed application tier importing only sanctioned packages, **When** the gate runs, **Then** RULE-E reports no violation for it.

**Functional Requirements**:

- **FR-001**: The layer-discipline guard MUST gain a **RULE-E** rule: an **application tier** (`internal/app/**`, `internal/cli`) MUST import, beyond stdlib, **only** `internal/domain/**`, `internal/config`, `internal/home`, and `internal/app/**`; any other `internal/**` import is a violation (`source -> import`, module-relative, ASCII ` -> `).
- **FR-002**: RULE-E MUST bind **both** application tiers (Q3) and MUST reuse the guard's existing machinery — **module-root-anchored** `go list`, the **`CROSS_TARGETS` union** with the filtered child env, and the merged **production + test** graph — so a test-only unsanctioned import is caught and an OS-gated import cannot hide.
- **FR-003**: The sanctioned set MUST be a **normative** section of the guard's embedded **tier table** (ADR 0011 **D7**) and MUST be **default-deny**: a governed application tier importing an internal package not in the sanctioned set is a violation; the rule MUST NOT treat "unknown" as "allowed".
- **FR-004**: RULE-E MUST be a member of the aggregate `make verify` (it rides the existing `verify-architecture` target) and MUST NOT weaken or re-order any existing rule's verdict.

**Non-Functional Requirements**:

- **NFR-001**: The gate MUST remain **deterministic** (sorted, diff-friendly) and **host-independent** (identical verdict on every host), adding no dependency (`go.mod`/`go.sum` unchanged) and the guard must be **stdlib-only**.

### User Story 2 - the residual edges are baselined and the baseline is a fail-on-stale ratchet (Priority: P1)

As a maintainer, I want the 3 residual application-tier edges recorded in the committed baseline so `dev` is green **today**; each later de-coupling slice must **remove** an entry (visible in the diff); and any divergence from truth — a new violation, or a stale entry — must **fail**, so the baseline can never rot.

**Why this priority**: it is what makes RULE-E *landable now* (Q1-A) while keeping teeth; without it RULE-E reddens `dev` on day one.

**Independent verification**: read `tools/arch/baseline.txt` — it lists exactly `internal/cli -> internal/agent`, `internal/cli -> internal/ui`, `internal/cli -> internal/ui/tui/prompt`; run the gate — green; remove an entry for an unfixed edge and re-run — **fails**; fix an edge but leave its entry and re-run — **fails** (stale).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** it is **green** and the baseline lists the 3 RULE-E residuals (re-measured from the gate, not transcribed).
2. **Given** an unfixed residual whose baseline entry was removed, **When** the gate runs, **Then** it **fails** (the baseline is not a blanket allow-list — ADR 0011 D3 preserved).
3. **Given** a residual that is fixed but whose baseline entry remains, **When** the gate runs, **Then** it **fails** and names the stale entry.
4. **Given** the baseline at 0 (after the later slices), **When** a stale entry is added, **Then** the gate **fails** (the anti-bypass rule holds at 0).

**Functional Requirements**:

- **FR-005**: `tools/arch/baseline.txt` MUST be extended with the 3 RULE-E residual edges, in the gate's own deterministic format (sorted in Go, module-relative, ASCII ` -> `), **generated from the gate** (`make verify-architecture-update`), never transcribed; RULE-A/B/C stay at **0**; the gate is **green on `dev` at delivery**.
- **FR-006**: The baseline MUST remain a **fail-on-stale ratchet** (ADR 0011 D3): a violation not listed **fails**; a listed-but-no-longer-violating entry **fails**; an entry removed while its violation persists **fails**; the baseline only shrinks toward truth.
- **FR-007**: The round MUST NOT fix any residual edge (that is the later slices); it MUST ship **zero product code** and MUST NOT change any existing gate's behaviour beyond adding RULE-E.

**Non-Functional Requirements**:

- **NFR-002**: The gate + baseline + tier-table extension MUST land as **one atomic delivery** (round-040 TD-1 precedent); the committed baseline MUST be the gate's own output at the delivered head.

### User Story 3 - the rule and the sanctioned set are recorded and single-sourced (Priority: P2)

As a maintainer/operator, I want RULE-E, the sanctioned set, and the baseline policy recorded in `specs/truth/techstack.md` (Build & Tooling) and in a new ADR, so the rule is discoverable, citable by the later slices, and cannot silently drift from the machine source.

**Why this priority**: it bounds the round in truth and governance. It is a documentation/governance constraint on Stories 1–2.

**Independent verification**: read `docs/decisions/0016-*.md` + the index row; read `specs/truth/techstack.md` (Build & Tooling) — RULE-E, its sanctioned set, and its baseline policy are recorded (citing ADR 0016 and the tier table as the normative source); the guard self-test asserts the tier table classifies every governed package **and** every sanctioned entry is live (fail-on-stale, Q2-A).

**Acceptance Scenarios**:

1. **Given** the technology-stack truth, **When** the round lands, **Then** the Build & Tooling section records RULE-E + the sanctioned set + the fail-on-stale baseline policy, citing ADR 0016.
2. **Given** the guard's tier table, **When** the self-test runs, **Then** it asserts every governed package is classified (RULE-D coverage) **and** every sanctioned allow-list entry is still used (an unused sanctioned entry **fails**).
3. **Given** the pre-round gate catalog, **When** the round lands, **Then** every existing rule's verdict is unchanged and `make verify` gains no new target (RULE-E rides `verify-architecture`).

**Functional Requirements**:

- **FR-008**: A new **ADR `0016`** (`docs/decisions/0016-application-import-ceiling.md` + the `docs/decisions/README.md` index row) MUST record RULE-E: the rule, the sanctioned set, the default-deny + fail-on-stale policy, **what the rule is *not*** (D10-style scope statement — it binds only the application tiers; it constrains **`internal/**`** imports only, so stdlib/third-party are out of its scope; RULE-A/B/C/D are unchanged), and its relation to ADR **0011** (the tier table it extends, the baseline ratchet it reuses) and **0013**.
- **FR-009**: `specs/truth/techstack.md` (**Build & Tooling**) MUST be updated — **MODIFY**, not NOOP — the Layer-discipline gate row MUST name RULE-E, the sanctioned set, and the residue ratchet, citing ADR 0016 + the tier table as the normative source; the `verify` aggregate member is unchanged (`verify-architecture`).
- **FR-010**: The guard self-test MUST assert (a) the tier table classifies **every** governed package (RULE-D coverage — carried forward) and (b) **fail-on-stale for the allow-list**: every sanctioned entry is imported by at least one governed application-tier package (an unused sanctioned entry **fails**).

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only**; the round adds **no** new third-party dependency (`go.mod`/`go.sum` unchanged).

### Edge Cases

- **Gate lands before its baseline** → forbidden: the rule + baseline + tier-table extension land as **one atomic delivery** (NFR-002, round-040 TD-1).
- **A *new* unsanctioned import while the 3 residuals are baselined** → **fails** on the new edge (FR-006).
- **A *fixed* residual whose baseline entry remains (stale)** → **fails**, naming the entry (FR-006).
- **An unused sanctioned allow-list entry** → **fails** (fail-on-stale allow-list, FR-010 / Q2-A).
- **An `internal/**` package matching no tier** → **fails** (default-deny, FR-003).
- **A cycle** → the SCC pass **fails** (cycles have no baseline; ADR 0011 D8 preserved) — RULE-E adds entries to the same ratchet, never a cycle exemption.
- **`internal/cli` test files importing `internal/ui`** → the same package edge as production (no new entry); the merged production+test graph is governed (FR-002).
- **A non-`internal` import (stdlib / third-party)** → **out of RULE-E's scope** (the rule constrains `internal/**` imports; a *recorded residual*): today no application-tier package imports a third-party module, and stdlib is trivially allowed. The scope statement lives in the ADR (FR-008) and is re-adjudicated if a third-party application-tier import ever appears.
- **OS/build-tag-gated files** → the `CROSS_TARGETS` union (FR-002) catches an OS-gated unsanctioned import; custom build-tag-gated files remain a recorded out-of-scope residual (ADR 0011 D6).
- **The round's own guard files** → the guard + its self-test must not themselves introduce a governed violation (they live in the exempt `tools/**`).

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-011**: The round MUST be **tooling + truth only**: it MUST add RULE-E, its self-tests, the baseline entries, the ADR, and the truth/spec records, and MUST NOT modify any production Go behaviour, any `stdout`/`stderr` behaviour, any flag, any exit code, or any existing rule's verdict.
- **FR-012**: The round MUST record the **residual edges** and the **sanctioned set** explicitly in `research.md`/`plan.md`, re-measure the baseline from the gate before freezing it, and MUST leave the de-coupling of the 3 residuals **out of scope** (deferred to the later R5 slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)) along with **F-4/F-6/F-7/F-8**.
- **FR-013**: The round MUST include **falsifiability witnesses** reproduced then reverted (ADR 0010): (a) a new unsanctioned application-tier import ⇒ RULE-E reds; (b) an unused sanctioned allow-list entry ⇒ the coverage self-test reds; (c) a stale baseline entry at 3 (and at the eventual 0) ⇒ the gate reds. Shape: **ADD** a rule + baseline entries + truth/ADR records; it changes no existing truth row's meaning beyond the Layer-discipline gate row.

#### Non-Functional Requirements

- **NFR-004**: The round's witness MUST be **the gate + unit seams**, NOT the E2E suite (a green suite alone is false confidence — the round-009 trap, per #92 AC5); `make verify` (incl. `verify-architecture`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL row.
- **NFR-005**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.

### Key Entities

- **RULE-E** — the new guard rule: an application-tier import **ceiling** (domain + stdlib + `config` + `home` + `app/**`), default-deny.
- **The sanctioned set** — the normative allow-list, embedded in the gate's **tier table** (ADR 0011 D7).
- **The baseline** — `tools/arch/baseline.txt`; gains the 3 RULE-E residual edges; a **fail-on-stale ratchet** that only shrinks toward 0.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make verify` runs RULE-E; on `dev` it is **green** with the baseline listing exactly the **3** RULE-E residuals (re-measured from the gate at freeze); RULE-A/B/C stay **0** and cycles **0**. (covers FR-001, FR-003, FR-005)
- **SC-002**: A deliberately added **new** unsanctioned application-tier import makes RULE-E **fail** and name the offending `source -> import` — reproduced as a falsifiability witness, then reverted (ADR 0010). (covers FR-002, FR-006, FR-013)
- **SC-003**: The baseline is a self-policing **ratchet**: removing an entry for an unfixed residual fails; a stale entry (a fixed residual left in the baseline) fails; an **unused sanctioned allow-list entry** fails (fail-on-stale allow-list). (covers FR-006, FR-010)
- **SC-004**: The gate's verdict is identical across hosts (`CROSS_TARGETS` union), deterministic/sorted, asserts **0 cycles**, treats an unranked `internal/**` package as a failure (default-deny), and adds no dependency (`go.mod`/`go.sum` unchanged). (covers FR-002, FR-003, NFR-001, NFR-005)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records RULE-E + the sanctioned set + the ratchet policy (citing ADR 0016); **ADR 0016** + its index row are present; every existing rule's verdict is unchanged; **no production Go behaviour** changes; the Gherkin/DSL topology audit is unchanged/green. (covers FR-008, FR-009, FR-011, NFR-004)
- **SC-006**: The 3 residuals + F-4/F-6/F-7/F-8 are recorded as out-of-scope on [#101](https://github.com/gosharplite/tellme/issues/101) (a live issue, not a frozen package); no frozen `specs/plans/NNN-*` package is touched. (covers FR-012)

## Assumptions

- **A1 (gate-first)** — R5 is a **programme**; 047 is **R5.1** (the rule + ADR + baseline). The de-coupling of the 3 residuals is later slices (Q1-A).
- **A2 (weak E2E carrier)** — the round changes no `tellme` CLI behaviour, so the E2E suite is a **weak acceptance carrier**; the witness is **the gate + its self-tests** (NFR-004). `/axb-spec-by-example` is therefore **NOOP** — precedent: rounds 020/031/036/041/042/043/044/045/046.
- **A3 (rule form is RD)** — the exact tier-table shape, the sanctioned-set representation, and the guard file layout are **RD** decisions (`research.md` D-x), not FRs.
- **A4 (ADR required)** — **ADR 0016** records RULE-E (FR-008); the number is confirmed in research.
- **A5 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP** (the baseline is a repo artifact, not runtime state); `/axb-ui-plan` **skipped** (not user-facing).
- **A6 (`/axb-dsl-refine` NOOP)** — no feature Rule, Example, step, or `DSLRow` change (the gate is a dev surface, not the `tellme` CLI contract); the topology audit stays unchanged.
- **A7 (scope guard)** — see FR-011/FR-012; the round touches only `tools/arch/**`, `docs/decisions/**`, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's `session-summary.md`.

## Out of scope (recorded)

- The **de-coupling** of the 3 residual edges (`cli → agent`, `cli → ui`, `cli → ui/tui/prompt`) and each port's home (`internal/domain` vs `internal/app`) — later R5 slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101).
- **F-4/F-6/F-7/F-8** (the PR #104 review deferrals: `Options.RunTUIPrompt` unexport; narrow seams ≤2 fields; named `Discovery{Closer}`; `OutputSink` → interface) — later slices.
- Ruling **off** `internal/config` / `internal/home` / `internal/app/suggestions` as sanctioned utilities (Q4-A sanctions them).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting adapters or domain business logic; the yield policy (ADR 0014) is untouched.
- The reference's `modelith-layers` analogue; custom build-tag-gated imports (ADR 0011 D6 residual); the topology audit's stale-semantics blind spot → [#92](https://github.com/gosharplite/tellme/issues/92).
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#101](https://github.com/gosharplite/tellme/issues/101) — the round's anchor (**R5**, the gate precondition + AC1–AC5 + F-4/F-6/F-7/F-8).
- [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** (second clause) + **AC5** (witness = gate + unit seams).
- Recorded by round **044** (R2, [#100](https://github.com/gosharplite/tellme/issues/100)) `/axb-clarify` **Q5** — the strict form is this follow-up; PR [#104](https://github.com/gosharplite/tellme/pull/104) review `5243584043` — F-4/F-6/F-7/F-8.
- Round **042** ([#93](https://github.com/gosharplite/tellme/issues/93), ADR 0011) — the gate + baseline precedent; round **046** ([#108](https://github.com/gosharplite/tellme/issues/108), ADR 0015) — ratcheted the RULE-A baseline to **0**.

## References

- [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) · [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) — the guard + tier table + baseline.
- [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) (RULE-A/B/C/D + D3 ratchet + D7 normative table) · [`0013`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0013-composition-root-injection.md) (the injected `Dependencies` seam) · [`0015`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0015-loop-presentation-port.md).
- [`internal/cli/cli.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/cli.go) · [`internal/app/deps/deps.go`](https://github.com/gosharplite/tellme/blob/dev/internal/app/deps/deps.go) — the application tiers under RULE-E.
- [`specs/truth/techstack.md`](https://github.com/gosharplite/tellme/blob/dev/specs/truth/techstack.md) — Build & Tooling (the gate row).
