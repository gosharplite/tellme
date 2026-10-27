# Truth delta — round 098 `098-param-swap-guard`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop` proves
the area was checked).

## /axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` — a new ***Parameter-shape gate*** row in **Build & Tooling**, immediately after the *Layer-discipline gate* row | Record the gate as shipped current state: `make verify-params` = `go vet -tags=quality ./tools/params` + `golangci-lint run --build-tags=quality ./tools/params/...` + `go test -count=1 -tags=quality -run TestVerifyParameterShapes ./tools/params`; a `go/parser` walk over `internal/**` production `FuncDecl`s **and** top-level interface methods with a **structural** type key; arms (a) `n ≥ 4` + a run ≥ 2, (b) a run ≥ 3, (c) a run ≥ 2 func-typed (**DI idiom, not a swap rule**), (d) a same-package named-struct literal with ≥ 2 unkeyed elements; the `context.Context` / `interface{}` / `any` / lone-trailing-variadic exemptions; the **fail-on-stale, refuse-to-grow** allow-list (`tools/params/allowlist.txt`, empty at delivery); `-count=1` kept for **parity** with `verify-architecture` — **not** its cache-vacuity rationale, which does not hold here (this gate reads the tree in-process; the architecture gate shells a `go list` child process) — and arm (d) **attributes elided literal types** (`[]T{{…}}`, `map[K]T{"k": {…}}`) with function-local structs collected; and the recorded boundaries with their measurements (result lists 11 · type aliases 0 · top-level interfaces only · the `n ≥ 4` boundary leaving 6 interface + 33 `FuncDecl`). Policy: **ADR 0068**. | FR-001…FR-007 / NFR-001 / NFR-002: a new gate is a technology-stack fact, and `techstack.md` is the gate's truth home (the ADR 0042 / ADR 0066 precedent — a gate is described where the toolchain is described). |
| MODIFY | `specs/truth/techstack.md` — the ***Task runner*** row (`make` aggregates) | Add **`verify-params`** to the enumerated `make verify` aggregate (`… + verify-architecture + verify-params + modelith-check + …`). | SC-003: the aggregate list is the authoritative statement of what `make verify` runs; a member missing from it would make the row false. |
| NOOP | `specs/truth/techstack.md` — all other rows (Language · dependencies · MCP SDK · domain model · hermetic invocation · E2E harness rows) | Inspected, nothing needed: no new dependency (`go.mod`/`go.sum` unchanged), no dev-tool binary (stdlib `go/parser` only), no change to the language, the SDK confinement, or the E2E harness. | NFR-002 / SC-006. |

## /axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/cli/**/*.feature` + every `dsl.md` row (chat · listing · diagnostics · tools) | Inspected, nothing needed: the round changes no command, flag, exit code, rendered byte, `turns.log` line or persisted-state shape. The CLI's executable contract is **byte-identical**, which is exactly what W8 asserts (the full E2E contract unchanged at **334 scenarios / 2519 steps**). A build-target guard is not a godog surface: the E2E harness drives the **built binary**, and this round changes the binary's internals, not its behaviour (the round-094 / round-096 precedent). | I-1 (behaviour-neutral) / NFR-004: no product-observable behaviour change, so no truth-feature may claim one. |

## /axb-api-plan (owner: `specs/truth/contracts/**`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Not applicable, by analysis decision: a standalone CLI has no OpenAPI/endpoint surface, and the round changes no transport contract. Skipped by `/axb-system-analysis`. | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |

## /axb-data-plan (owner: `specs/truth/data/**`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Not applicable, by analysis decision: no persisted-state shape changes — `history.jsonl`, `turns.log`, `tokens.log`, `tools-count.jsonl`, the shared prompt log and the MCP tool cache keep their bytes and semantics. Skipped by `/axb-system-analysis`. | The round changes no state. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0068-parameter-shape-gate.md` (+ its `docs/decisions/README.md` index row) | Record the two hazards, arms (a)–(d), arm (c) as a **DI idiom** (not a swap rule), the interface scope, the `n ≥ 4` boundary with its measured remainder, result lists out of scope (11), syntactic type identity (0 aliases), `n` counting exempt params, the allow-list / fail-on-stale / refuse-to-grow design, the `-count=1` rationale and its `doc.go` origin, and the **measurement-bug lesson** (the script owns the numbers). | DoD 11 / `adr-index-consistent`: a durable decision needs an ADR; no existing ADR body is edited (0068 is new; 0011/0012/0020/0041/0064 are **cited**). |
| MODIFY | `docs/decisions/0056-rollback-durability-witness.md` + `docs/decisions/0058-mcp-tool-cache.md` (**one link each**) | Fix the dead `0012-deterministic-hermetic-tests.md` link → `0012-hermetic-make-go-env.md` (the file that exists). | DoD 12 / issue #210 §7: `verify-adr-index` checks index membership only, never links. |
| MODIFY | `docs/domain-model/quality.modelith.yaml` + the re-rendered `.md` | The `QualityPipeline` gate list gains `verify-params`; the `QualityGate` examples name the parameter-shape gate. | ADR 0041 / DoD 4: the model is load-bearing, and this round **is** a process-model change. |
| NOOP | `docs/domain-model/tellme.modelith.yaml` (+ its `.md`) | Resolved, not assumed: `usage-summed-for-the-ready-summary` concerns **values**, not parameter shapes, and the round changes no product behaviour. Re-rendered — byte-identical (verified by `modelith-check`). | ADR 0041 (a round must **resolve** the model question, not skip it). |
| MODIFY | `tools/params/` (new package) · `Makefile` (3 new targets + the `verify` aggregate) · `internal/**` (the 46 fix sites) · `internal/cli/{boot,turn,reporting,workspace,diagnostics,interactive}.go` (new) | The round's implementation: the guard, its wiring, the seam structs, the split. | FR-001…FR-012 / SC-001…SC-006. |
| MODIFY | `STATUS.md` (+ the day's session summary) | Round 098 delivered; the gate recorded as live state. | Session closeout hygiene (not truth). |

### PR-review fold (2026-10-27; reviewer comment [`6436228558`](https://github.com/gosharplite/tellme/pull/211#issuecomment-6436228558))

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — the ***Parameter-shape gate*** row | (i) **`-count=1` is kept for parity, not for cache-vacuity** — this gate reads the tree in-process, which Go **does** key into the test cache; the `tools/arch` rationale (a `go list` child process) was inherited wrongly and is now explicitly disclaimed on the row. (ii) Arm (d) now **attributes elided literal types** (`[]T{{…}}`, `map[K]T{"k": {…}}`) and collects **function-local** struct types. | The row stated a **false** rationale (the round's own witness contradicted it) and an arm narrower than its published definition. Truth must be current and correct, not merely present. |
| MODIFY | `specs/truth/techstack.md` — same row (the boundary measurements) | Unchanged in substance; the red-set total it implies is now **46** (40 signatures + 6 literals), each figure re-measured from the script. | `script-owns-the-numbers`. |
| NOOP | `specs/truth/features/cli/**`, `specs/truth/contracts/**`, `specs/truth/data/**` | Inspected, nothing needed — the fold changed a **gate's** correctness and the plan package's records only. No command, flag, exit code, rendered byte, persisted state or transport contract moved; the E2E contract is unchanged at **334 / 2519**. | I-1 (behaviour-neutral). |
