# Requirements checklist — round 093 `093-interactive-prompt-seam-determinism`

Anchor: [#191](https://github.com/gosharplite/tellme/issues/191). **DoD = close it.**

| # | Requirement | Status | Carrier |
| --- | --- | --- | --- |
| FR-001 | the `-i` scripted-key handshake delivers the terminal key only after the **composed text** painted (content-derived gate; no sleep) | ✅ | harness gate + the steps unit pin (SC-002) |
| FR-002 | the seam's scripted line break is the terminal Enter byte (CR), so the product inserts a newline | ✅ | acceptance scenario 1 (two separate editor rows) |
| FR-003 | the `keeps the typed text` assertion requires each line on its own editor row (the joined state reddens) | ✅ | `thenKeepsTypedText` strengthened + E2E Example |
| NFR-001 | no sleep/retry (`verify-no-test-sleep`); no product change; `go.mod`/`go.sum` unchanged | ✅ | `make verify` + review (I-2/I-3) |
| NFR-002 | `make check` green; E2E counts unchanged (or state the delta) | ✅ | `make check` |

## Witness / falsifiability (W)

| W | Claim | Mutation that reddens it |
| --- | --- | --- |
| **W-A** | the paint gate is content-aware (the terminal key waits for the composed text) | revert the gate to the generic `┌` marker ⇒ the harness/steps unit pin reddens **and** the E2E flakes under repetition (the pre-fix measurement) |
| **W-B** | the scripted line break is a real newline | revert the seam's line break to a raw LF ⇒ the multi-line Example reddens (the joined single row, FR-003) |
| **W-C** | the assertion is discriminating | drop the "separate rows" requirement ⇒ the joined state passes (the pinned regression `B` returns green) |
| **W-D** | behaviour identity (product) | `internal/**` + `cmd/**` unchanged; the whole unit suite + E2E stay green with no product assertion changed |

## DoD (issue #191)

- [ ] the `-i` submit/capture race is fixed and the scenario is deterministic under repetition (≥50 runs).
- [ ] root-cause fix (not a sleep/retry); `make check` green; E2E green.
- [ ] a deterministic carrier is added (unit pin + the E2E repetition witness).
- [ ] human-reviewed + merged (**no Copilot review; only a human merges**).

**備註**: `NEEDS CLARIFICATION` — none. Ready for `/axb-spec-by-example`.
