# System analysis — round 098 `098-param-swap-guard`

**Owner**: `/axb-system-analysis` (RD) · **Input**: `spec.md` · `research.md` · `truth-delta.md` ·
`specs/truth/**`

## 1. Interfaces in scope

| Interface | Kind | Change |
| --- | --- | --- |
| `cli` | executable contract (`specs/truth/features/cli/**`) | **NOOP** — no command, flag, exit code, rendered byte or persisted state changes. The E2E contract is the behaviour-identity witness (**334 / 2519 unchanged**), not a change surface. |
| **build** (the Makefile + `tools/**` surface) | build/tooling contract (`specs/truth/techstack.md`) | **MODIFY** — a new `verify-params` member of `make verify`; a new `tools/params/` guard package. |

There is no third interface: this round touches no HTTP/API surface (a CLI end), no persisted state, and
no governed UX surface.

## 2. Analysis waves

- **Wave 0 — the script (must land first).** `tools/params/{doc.go,params_test.go,allowlist.txt}`. The
  scan is committed **before** any fix, so its red run is the specification (SC-001).
- **Wave 1 — the domain seam shapes.** `internal/domain/render` (`ReadyStats`, `PayloadStatus`, the three
  `Lines` port methods) and `internal/domain/llm` (`CostInputs`). Consumers: `internal/ui` (the adapters),
  `internal/cli` (the call sites). Pure, dependency-free, compiles alone.
- **Wave 2 — the application seam shapes.** `internal/app/suggestions` (`WorkspaceQuery`), `internal/cli`
  (`sessionRef`, `cliDeps`, `spinnerLine`, `loopSignals`, `actionText`, `tokenBound`, `usageCall`,
  `retrySpec`) and the `internal/cli` file split.
- **Wave 3 — the infrastructure seam shapes.** `internal/config` (`mcpFieldRef`),
  `internal/infrastructure/{llm/gemini,llm/openai}` (`requestSpec`), `internal/infrastructure/{mcp,di}`
  (`RemoteSpec` + the `ClientFactory` seam), `internal/infrastructure/telemetry` (`darwinMachVM`,
  `darwinSysctlVM`), `internal/infrastructure/tools` (`treeSpec`, `captureTarget`, `capturedPipes`).
- **Wave 4 — the gate joins `verify`.** `Makefile` (`verify-params` + `verify-params-report` +
  `verify-params-update`, and the aggregate). Then the truth/model/ADR records.

Wave 2 depends on Wave 1 only through `render.Lines`' signatures; the waves are otherwise disjoint
packages, so the compiler arbitrates each wave independently (Red-Green per package).

## 3. Dependency direction (must not move)

`internal/domain/**` ← `internal/app/**` ← `internal/infrastructure/**` ← `internal/agent` ←
`internal/ui` ← `internal/cli`, with `cmd/**`, `tests/**`, `tools/**` exempt. This round:

- adds **no** import edge (the `internal/cli` split is **same-package**; `tools/params` is stdlib-only
  `go/parser`, exempt);
- changes **two** domain port contracts **in place** (`render.Lines`, `suggestions.WorkspaceSource`) —
  shapes, not direction;
- leaves `verify-architecture`'s baseline **header-only** — that is the proof the slice is legal (W9).

## 4. `/axb-api-plan` and `/axb-data-plan`

- **`/axb-api-plan` — NOOP.** A standalone CLI has no OpenAPI/endpoint surface; the round changes no
  transport contract. (Recorded as a NOOP in `truth-delta.md` on the owner's behalf.)
- **`/axb-data-plan` — NOOP.** No persisted-state shape changes: `history.jsonl`, `turns.log`,
  `tokens.log`, `tools-count.jsonl`, the prompt log and the MCP tool cache all keep their bytes and
  semantics.

## 5. `/axb-ui-plan`

**Skipped** — a plain line-oriented round with no governed terminal UX surface. The `-i` TUI is untouched.

## 6. Test strategy

| Layer | Carrier |
| --- | --- |
| The gate's rule logic | synthetic-input self-tests inside the gate (`arms` · `unkeyed` · `allow-list`), each pinned by a name-set assertion |
| The gate's red-first property | W1 (pre-round tree) → W2 (this tree) |
| Each arm's liveness | W3 (leaf, scalar) · W4 (non-scalar) · W5 (interface) · W10 (arm (d)) |
| The allow-list | W6 (synthetic stale entry) |
| The cache-vacuity rule | W7 (`-count=1` against a warm cache) |
| Behaviour identity | W8 (full E2E contract, unchanged counts) |
| Structure | W9 (`verify-architecture` header-only · `modelith-check` · `test-race`) |

No sleeps, no network, no retries (ADR 0012 / `verify-no-test-sleep` / `verify-no-network`).

## 7. Risk register

| Risk | Mitigation |
| --- | --- |
| The refactor silently changes a rendered byte | W8 (334 / 2519 unchanged); the byte-producing formatters keep their format strings verbatim (only field access changes) |
| A seam struct reintroduces the hazard as a field list | Out of scope by the rule (struct **fields** are not parameters) and recorded in ADR 0068 D7/D11; unkeyed **construction** is what arm (d) governs |
| The allow-list becomes a dumping ground | Empty at delivery; refuse-to-grow; a recorded reason per entry; W6 proves the stale direction |
| The gate rots into an advisory | It is a `make verify` member (zero-tolerance), not an advisory — the ADR 0064 distinction is cited in ADR 0068 D9 |
