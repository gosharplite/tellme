# Requirements checklist — round 092 `092-single-source-agent-tool-set`

Anchor: [#189](https://github.com/gosharplite/tellme/issues/189). **DoD = close it.**

| # | Requirement | Status | Carrier |
| --- | --- | --- | --- |
| FR-001 | one canonical owner of the base set; the root + the e2e enumerator derive from it | ✅ | `infratools.NewAgentBaseTools`; both callers delegate (I-1) |
| FR-002 | the e2e enumerator is bound to the owner (reddens on a divergent inline copy) | ✅ | `tests/e2e/steps/tool_usage_test.go` — `TestRegisteredToolNamesIsTheCanonicalBaseSet` |
| FR-003 | the production base set is bound to the owner (same shape) | ✅ | `cmd/tellme/deps_test.go` — `TestAgentToolsIsTheCanonicalBaseSet` |
| FR-004 | the capability gate is unchanged (`read_image` only for vision; the union enumerator keeps it) | ✅ | existing `TestCompositionResolvesTheFamilyAwareImageCeiling` + the unchanged `recordableToolNames()` |
| NFR-001 | no behaviour change; set/order byte-identical; no `go.mod`/`go.sum` change; stdlib-only | ✅ | full suite green, no assertion changed (I-2) |
| NFR-002 | `make verify` green (architecture 0-violation) · E2E unchanged (330 · 2487) | ✅ | `make check` |

## Witness / falsifiability (W)

| W | Claim | Mutation that reddens it |
| --- | --- | --- |
| **W-1** | the e2e enumerator derives from the canonical owner | re-inline the base constructors in `registeredToolNames()` (dropping one) ⇒ `TestRegisteredToolNamesIsTheCanonicalBaseSet` reddens |
| **W-2** | the canonical owner is load-bearing (the doc carrier binds to it) | remove a tool from `NewAgentBaseTools` ⇒ round-090's `TestOfferedSetDocMatchesTheLiveRegistry` reddens (doc (8) vs live (7)) |
| **W-3** | the production base set derives from the canonical owner | re-inline the base constructors in `assembleAgentTools()` (diverging) ⇒ `TestAgentToolsIsTheCanonicalBaseSet` reddens |
| **W-4** | behaviour identity | the full unit suite + the godog E2E stay green with **no assertion changed**; E2E counts unchanged |

## DoD (issue #189)

- [x] the e2e enumerator is bound to (single-sourced with) `agentTools()` — via one canonical owner.
- [x] the placement decision is recorded (**ADR 0062**).
- [x] `make check` green · E2E green · `go.mod`/`go.sum` unchanged.
- [ ] human-reviewed + merged (**no Copilot review; only a human merges**).
