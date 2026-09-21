# Tasks — MCP tool-name presentation (round 077)

**Plan Package**: `specs/plans/077-mcp-tool-name-presentation`
**Anchor**: [#155](https://github.com/gosharplite/tellme/issues/155) · **ADR**: [0049](../../decisions/0049-mcp-tool-name-discoverability.md)

## Setup

- **T001** Confirm the baseline is green on the round branch (`gofmt -l .` clean · `go test -count=1 ./internal/infrastructure/mcp/` ok · `make verify` OK). *(no new dependency, no Makefile change — setup otherwise empty)*

## Foundational

> None — the change is stdlib-only and confined to the MCP adapter's `Description()`; no shared infrastructure is needed.

## Phase 3 — Test Alignment & Implementation

- **T002** [ALIGN] Extend `internal/infrastructure/mcp/tool_test.go` (or the existing pins) with the round-077 expectation **before** the change:
  - a **non-empty** server description → the offered description is `Call this tool as "<wire name>". <server text>` (server text byte-identical, appended-to not rewritten);
  - an **empty** server description → the fallback names the **callable wire name** (not the bare name) and states the server;
  - a **truncated** (64-byte hash) name → the note carries the **post-truncation** name, equal to `Name()`.
  Run → **RED** (the current code has no note; the fallback names the bare tool).
- **T003** [GREEN] Implement in `internal/infrastructure/mcp/tool.go`:
  - a named `callNameNoteFormat` constant (one home) + a computed `description` field on `Tool` (`NewTool` builds it once: `fmt.Sprintf(callNameNoteFormat, name) + body`, where `body` = the server description, else the fallback `"MCP tool " + name + " from server " + server`);
  - `Description()` returns the computed field.
  Run → **GREEN**.
- **T004** [GREEN] Ping the **schema-unchanged** invariant: a pin asserting `Parameters()` (the envelope) is byte-identical to the pre-round shape for a sample tool (the change is description-only).
- **T005** [REFACTOR] Keep the note's template a single constant; ensure the name interpolated is `t.name` (post-truncation); no other package changes.

## Phase 4 — Verification & Regression

- **T006** [WITNESS W1] Remove the note → the T002 non-empty pin REDs; revert. (falsifiability)
- **T007** [WITNESS W2] Note `def.Name` instead of `t.name` → the callable-name pin REDs; revert.
- **T008** [WITNESS W3] Restore the bare-name fallback → the empty-description pin REDs; revert.
- **T009** [WITNESS W4] Note the untruncated name on a long tool → the truncated-name pin REDs; revert.
- **T010** [E2E] `/axb-dsl-refine` carrier: an interface Rule/Example observing the offered MCP declaration's description names the callable wire name (via the hermetic fake MCP server + the fake provider), plus the `dsl.md` Then.
- **T011** [GATES] `gofmt -l .` + `goimports -l .` clean · `go vet ./...` clean · `go test -count=1 ./...` green · `make verify` OK · `go.mod`/`go.sum` unchanged · topology audit — no new errors.
- **T012** [RECORDS] Update `STATUS.md` + the day summary; record the ADR 0049 forward items.

## Verification ledger

| Task | Status |
| --- | --- |
| T001 | ✅ baseline green on the round branch |
| T002 | ✅ pins written (`tool_r077_test.go`), run RED first |
| T003 | ✅ `callNameNoteFormat` + computed `description` in `NewTool`; GREEN |
| T004 | ✅ `TestMCPToolName_OfferedSchemaIsUnchanged` (envelope/schema unchanged) |
| T005 | ✅ one constant home; name interpolated is `t.name` |
| T006–T009 | ✅ witnesses run then reverted — see below |
| T010 | ✅ E2E carrier: the new Rule + `step_r077_mcp_name.go` + the `dsl.md` Then row |
| T011 | ✅ `gofmt`/`goimports` clean · `go vet` clean · `go test -count=1 ./...` green (incl. E2E **288 scenarios**) · `make verify` **OK** · `go.mod`/`go.sum` unchanged · topology audit **the same 5 pre-existing errors, none new** (52 features · 410 module rows · 2117 steps) |
| T012 | pending (closeout) |

### Witnesses (reproduced then reverted)

- **W1** — neutralize the note (empty `callNameNoteFormat`) ⇒ the **E2E `The offered declaration names the callable wire name`** reddens at the intended Then; the unit build also reddens (`vet` printf check).
- **W2** — the note names `def.Name` (bare) instead of `t.name` ⇒ reddens **three** unit pins — `…DescriptionIsNotePlusServerTextByteExact`, `…EmptyDescriptionFallbackNamesTheCallableName`, `…NoteNamesTheTruncatedWireName` — **plus** the E2E `The offered declaration names the callable wire name`.
- **W3** — restore the bare-name fallback ⇒ `…EmptyDescriptionFallbackNamesTheCallableName` reddens **and (post-fold TD-1) the E2E `A server that ships no description is still offered by its callable name` reddens** at the force-bearing Then.
- **W4** — the note carries the **post-truncation** wire name (`Name()`); covered by W2's truncated-name failure.
- **Folds (PR #157 architect review)** — **F-1** byte-exact description pin over a whitespace/CRLF fixture (a `TrimSpace` of the server text, or interposed text, now reddens; previously green); **F-2** the envelope property-set pin reddens on a stray `hint` property (round-056's structural pin alone stayed green); **F-3** the fallback provenance is asserted on the fallback SEGMENT (dropping `from server <server>` now reddens); **TD-1** the empty-description carrier (`mcptest.Options{EmptyDescription: true}`) makes US2/FR-004 E2E-reachable (the bare fallback now reddens the E2E).

