# Phase 0 Research: layer-discipline gate + violation baseline (Round 042)

Topic: add a **layer-discipline verification gate** to `make verify` that enforces an **import-direction rule** over the module's layers, shipping with a **committed baseline** of the layer violations that already exist, so `dev` is green today, a **new** illegal import fails the standard gate, and the existing count can only **ratchet down** to 0 across R2–R4 ([#92](https://github.com/gosharplite/tellme/issues/92) → [#93](https://github.com/gosharplite/tellme/issues/93)).

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` + stdlib `testing`), provider transports, skills, MCP, and the presentation packages were locked in rounds 001–041. The system still has **one CLI end**, adds **no new system end**, **no external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; BDD techstack = `godog` on the built binary; E2E + unit strategy) and are **not re-decided** (D10). **IN**: the gate, its baseline, its self-tests, the rule ADR, and the truth rows. **OUT**: fixing any baselined violation (R2–R4), any `tellme` binary behaviour, and any existing gate's semantics.

> **Review fold (PR #94 architect, B-1/B-2 · TD-1…TD-4 · RF-1…RF-3):** the first draft of D1 stated a **one-part** rule ("any governed package's import of a higher tier") that **cannot reproduce the 7 `cli → infrastructure` entries** it claimed as its baseline (they are *downward* imports) — so the baseline was not derivable from the rule. D1 is now a **two-part predicate** (D1 below; ADR 0011 D2). D2's "respects build tags" claim was **false** and is **withdrawn** (TD-2/D6). D6's host-independence is now the **`CROSS_TARGETS` union** (TD-1/D5). D8 adds **ADR 0011** (RF-1). New: D4 anchors the enumeration (B-2), D11 default-deny (TD-3), D12 baseline format (RF-2), D13 one normative source (RF-3), D14 the acyclicity assertion (TD-4).

---

## Decision 1: A **two-part** import-direction rule over a pinned layer ranking (clarify Q1 → Option 2)

- **Decision**: the gate enforces, over a **pinned layer ranking** (low → high) — `internal/domain/**` (0, pure) → `internal/config`, `internal/home` (1, shared utilities) → `internal/app/**` (2) → `internal/infrastructure/**` (3) → `internal/agent` (4) → `internal/ui`, `internal/ui/tui/**` (5) → `internal/cli` (6); `cmd/tellme`, `tests/**`, `tools/**` **exempt** — the predicate: **(A)** no **upward** import (imported tier > importer tier); **(B)** the **application tiers** (`internal/app/**`, `internal/cli`) MUST NOT import `internal/infrastructure/**`; **(C)** `internal/domain/**` imports only `internal/domain/**` + stdlib; **(D)** an `internal/**` package matching **no** tier is a violation (default-deny).
- **Rationale**: the operator chose the broad form (the reference's `verify-architecture` shape) — a general rule catches future misuse, not just the one `cli → infrastructure` class. **The predicate must be stated in two parts**: the 7 `cli → infrastructure` edges are **downward** imports that part (A) leaves legal, so part (B) — the target rule [#93](https://github.com/gosharplite/tellme/issues/93) mechanised ("`internal/cli`: domain + stdlib + application utilities only") — is what makes the 7-row baseline **derivable from the rule**. Ranking `ui` above `agent` makes `agent → ui` a part-(A) violation (the 8th entry); ranking `config`/`home` at tier 1 keeps `infrastructure → config` (which the module ships) **legal**.
- **Measured (2026-09-17, whole module, `dev` `3f8ec08`)**: 30 packages · **8** violations · **0** unranked governed packages · **0** cycles. The 8 = 7 `internal/cli -> internal/infrastructure/{history, llm, skills, telemetry, tools, di, mcp}` (part B) + `internal/agent -> internal/ui` (part A).
- **Alternatives considered**:
  - **One-part "no upward import"** — cannot reproduce the 7; with fail-on-stale it would red the gate on the first run — rejected (B-1).
  - **Narrow target-rule only** (baseline 7) — ignores the `agent → ui` edge Q1 asked to flag — rejected.
  - **Broad with ranks chosen to keep 7** (rank `agent` ≥ `ui`) — pre-empts R3/R4's ownership of the loop→`ui` coupling — rejected.

## Decision 2: The guard is a build-tagged Go guard test invoked from a `Makefile` `verify-architecture` target

- **Decision**: implement the gate as a Go test guarded by `//go:build arch` (package `tools/arch/`, with an untagged `doc.go` so `go build ./...` never sees "build constraints exclude all Go files") that obtains the module's import graph from the **Go toolchain itself** via stdlib `os/exec` running `go list -f '{{.ImportPath}}|{{join .Imports " "}}' ./...`, then applies D1's predicate and D3's baseline diff. A new `Makefile` target `verify-architecture` runs `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch` and joins the `verify` aggregate (+ `.PHONY` + `help`).
- **Rationale**: the Go toolchain's own resolution is precise and stdlib-only (no `golang.org/x/tools/go/packages`); `go list` respects `internal/` visibility and the real package set. The `-tags=arch` gate keeps the guard out of the default `go test ./...` run. Mirrors the reference (`go test -tags=arch … -strict-arch=true`).
- **Alternatives considered**:
  - **Makefile grep** — cannot resolve tiers or the build context; mis-ranks nested packages — rejected.
  - **Pure `go/build` stdlib walk** — re-implements package resolution and can diverge from the real build graph — rejected.

## Decision 3: A committed, sorted **baseline** with **fail-on-stale** (clarify Q3 → Option 1)

- **Decision**: commit `tools/arch/baseline.txt` — sorted lines `<source> -> <import>` (module-relative, ASCII delimiter; see D12) parsed by the guard. Behaviour: a violation **not** in the baseline **fails**; a baseline entry that no longer violates (**stale**) **fails** (*"remove it from the baseline"*); a baselined violation removed from the baseline **fails**.
- **Rationale**: fail-on-stale makes the baseline a self-policing **ratchet** — the count only shrinks toward truth, and each fix (R2–R4) must edit the baseline (visible in the diff), which is what makes [#92](https://github.com/gosharplite/tellme/issues/92)'s "→ 0" claim falsifiable.
- **Alternatives considered**: **warn-on-stale** (baseline rots) — rejected; **a JSON baseline** (heavier, less diff-friendly) — rejected.

## Decision 4: The enumeration is anchored to the module root, and the graph is self-tested (B-2)

- **Decision**: the guard resolves the **module root once** (`go list -m -f '{{.Dir}}'`, or `runtime.Caller` + walk up to `go.mod`) and runs the child `go list` with `cmd.Dir = <moduleRoot>`; its self-test **asserts the enumeration itself** (expected package count + the presence of known governed packages — `internal/cli`, `internal/agent`, `internal/ui`, `internal/ui/tui/prompt`, `internal/domain/llm`) **before** any ranking assertion, asserts the child's exit status and **non-empty** output, and passes **no** `-e` (an error must be an error, never an empty graph).
- **Rationale**: `go list ./...` is **CWD-relative**, and a Go test binary runs with **CWD = its package directory**, so a guard that shells `go list ./...` from the test would enumerate only `./tools/arch/…` ⇒ an empty/near-empty graph ⇒ either all baseline entries read stale, or a **vacuous green** gate (the round-009 *"green suite = false confidence"* trap in gate form). The self-test must therefore assert the graph, not just a trivial "no self-violation" (which an empty graph satisfies).
- **Alternatives considered**: **rely on `make` to set CWD** — brittle; the guard must be correct when run directly (`go test ./tools/arch`) — rejected.

## Decision 5: Host-independence is the union over `CROSS_TARGETS` (TD-1)

- **Decision**: the guard evaluates the import graph as the **union** over `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` (round 020's `CROSS_TARGETS`) — 4 × `go list` with `GOOS`/`GOARCH` set — so an **OS-gated** illegal import cannot hide, and the single committed baseline is genuinely host-free.
- **Rationale**: `go list {{.Imports}}` resolves in the **host** build context: files excluded by `GOOS`/`GOARCH` land in `.IgnoredGoFiles` and their imports vanish (measured: on darwin, `internal/infrastructure/telemetry` reports `ignored=system_metrics_linux.go`, and its import set is OS-dependent). The verdict is identical across GOOS × CGO **today** only because both sampler files import downward — an OS-gated upward import would go red on one host, green on another, against a single baseline. The union closes that.
- **Alternatives considered**: **downgrade FR-004/SC-004 to "host-context-relative"** — honest but loses the property; the union is cheap (seconds) and matches `verify-cross-compile`'s rationale — rejected in favour of the union.

## Decision 6: Custom build-tag-gated files are out of scope (recorded) — the "respects build tags" claim is withdrawn (TD-2)

- **Decision**: the guard passes **no** `-tags` to the child `go list`; tag-gated files (e.g. `//go:build arch`) are therefore invisible and **out of scope** — a recorded residual (tellme's only custom tag today is `arch`, on the guard's own exempt `tools/**` file). GOOS-conditional compilation is covered by D5.
- **Rationale**: `go list` without `-tags` hides the import (measured in a scratch module: a `//go:build arch` file's import is absent from `.Imports` and listed in `.IgnoredGoFiles`; `go list -tags=arch …` surfaces it), and a `go test -tags=arch` invocation does **not** propagate its tag to a child `go list` (only `GOFLAGS`/env does). The earlier D2 claim ("it respects build tags … so a build-tagged illegal import cannot hide") was **false as written**; it is **withdrawn** and replaced by this scoping (also corrected in `techstack.md` — the round-036/041 lesson: a claim that entered as prose and was never measured).
- **Alternatives considered**: **pass the tag set explicitly** — possible, but then the guard must decide tag scope (and would rank its own `arch` file, in an exempt tree); recorded as a future option — rejected for R1's minimal scope.

## Decision 7: Wiring + truth record + ADR (RF-1)

- **Decision**: add `verify-architecture` to the `Makefile` (member of `make verify`; `.PHONY` + `help`). Record the gate + predicate + baseline policy in **`specs/truth/techstack.md` (Build & Tooling)** (new row; the **Task runner** row's `verify` list gains the member), and record the rule durably in a **new ADR 0011** (`docs/decisions/0011-layer-discipline-gate.md` + its index row).
- **Rationale**: the gates live in the `Makefile`; a new gate is a real techstack **MODIFY** (no unevidenced NOOP); and per `docs/decisions/README.md`, the layer rule + baseline policy is exactly "a project-level rule … that other artifacts (or future rounds) depend on and must be able to cite" — **R2–R4 will cite it, and this PR is proof it is contentious** (the same rule read two ways). *(This replaces the earlier "no ADR" position.)*
- **Alternatives considered**: **no ADR, rule only in truth prose** — leaves the rule without a citable home and keeps it narrative rather than a predicate — rejected.

## Decision 8: Default-deny for unranked governed packages (TD-3)

- **Decision**: an `internal/**` package matching **no** tier is a **violation** ("unranked governed package"); non-`internal` trees are the explicit exemption list (`cmd/**`, `tests/**`, `tools/**`). The guard's self-test asserts the tier table classifies **every** governed package.
- **Rationale**: this round ships the **first** unranked package (`tools/arch`) — exempt by the `tools/**` rule — and a future `internal/<new>` could otherwise be silently ignored (fail-open). A gate whose default is "unknown ⇒ allowed" hollows out the future-proofing Q1/Option 2 was chosen for.
- **Alternatives considered**: **silently ignore unranked** — fail-open — rejected.

## Decision 9: One normative ranking source + a consistency assertion (RF-3)

- **Decision**: the guard's embedded **tier table** is the single normative machine-readable source of the ranking; `techstack.md` states the rule as a short **predicate** and cites ADR 0011 + the table; the guard's self-test asserts the table covers every governed package (D8), so prose and machine source cannot silently diverge.
- **Rationale**: truth prose + a code table = two copies of one rule ⇒ drift, in a round whose purpose is anti-drift (`truth-single-owner`). The self-test ties them.
- **Alternatives considered**: **a second unit assertion parsing the truth row** — fragile (parsing prose) — rejected; the single-source + coverage assertion is simpler.

## Decision 10: Deterministic baseline format (RF-2)

- **Decision**: baseline lines are **sorted in Go** (`sort.Strings`, byte-wise — never shell `sort`, whose collation is locale-dependent), use **module-relative package paths** (the `github.com/gosharplite/tellme/` prefix stripped, matching the gate's own output), and an **ASCII** delimiter ` -> `. The baseline is **generated from the gate**, never transcribed.
- **Rationale**: a locale-dependent or path-form-mismatched baseline would itself be host-dependent and emit spurious new/stale diffs (NFR-001/SC-004), and would not match the gate's own output (FR-006/NFR-003).
- **Alternatives considered**: **shell `sort`** / **`→`** / **module paths** — all rejected for the above reasons.

## Decision 11: Acyclicity is asserted, not assumed (TD-4)

- **Decision**: the guard additionally runs a stdlib **SCC (Tarjan)** over the governed graph and asserts **0 cycles**. Cycles have **no baseline** — the count must be 0 (today it is 0). AC6 (domain purity) is likewise an **explicit** self-test assertion, not merely a consequence of tier 0.
- **Rationale**: issue [#93](https://github.com/gosharplite/tellme/issues/93)'s **AC4** names "0 circular references"; a **direction-only** rule can never detect a **same-tier** cycle (`internal/infrastructure/llm` ↔ `…/mcp`; `internal/ui` ↔ `internal/ui/tui/prompt` are the shape), so a separate pass is required to make AC4 a real assertion rather than a measurement. AC2's "7" is revised to "8" by Q1 — stated on the durable surface (issue [#93](https://github.com/gosharplite/tellme/issues/93)) at this gate.
- **Alternatives considered**: **record the narrow scope** (direction-only; acyclicity out of scope) — leaves AC4 vacuous; the SCC pass is cheap — rejected.

## Decision 12: Testing & BDD techstack unchanged; `/axb-dsl-refine` is `NOOP`

- **Decision**: no new system end and no change to the BDD techstack or strategy. `/axb-api-plan` = `NOOP`, `/axb-data-plan` = `NOOP` (the baseline is a repo artifact, not runtime state), `/axb-dsl-refine` = `NOOP` (the gate is a **dev surface**, not the `tellme` CLI end), `/axb-ui-plan` = skipped, `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — precedent: rounds 020/031/036/041).
- **Rationale**: the round's verification is the gate's own exit code + falsifiability witnesses, not a Gherkin scenario; forcing CLI Gherkin would be ceremony and risk an `acceptance-coverage` mismatch.
- **Alternatives considered**: **author plan-side acceptance Gherkin for `make verify`** — no user-facing CLI behaviour to express — rejected.

---

## Residual risks / forward links

- **Custom build-tag-gated imports** are out of scope (D6) — none today; a future tagged file needs a decision.
- **`go list` cost** — the union is 4 × `go list` (seconds); a cold module cache resolves modules like any build (no new module/service).
- **Guard self-exclusion** — `tools/arch` is exempt and imports only stdlib + `os/exec`; the self-test pins that the guard adds no violation.
- **`-tags=arch` compiled by no other gate** — `go vet ./...`/`verify-cross-compile` never compile the guard file (host-compiled only): a compile error surfaces only when the gate runs (accepted; recorded in `techstack.md`/ADR 0011 so a future "why isn't this vetted?" has an answer).
- **Ratchet breadth** — the 7 entries are R2's and the 8th is R3/R4's; the count reaches **0 across R2–R4** (how [#92](https://github.com/gosharplite/tellme/issues/92) AC1 is read).
- **`modelith-layers` analogue** — the reference's domain-model-as-code layer check is **not** adopted (tellme has no modelith toolchain) — recorded, not a gap.
- **Ranking judgements** — `config`/`home` at tier 1 (so `infrastructure → config` is legal) and `ui` above `agent` (so `agent → ui` is a violation). A later round may disagree; that is a ranking edit + a baseline re-measure (`techstack`/ADR MODIFY).
- **AC4's cycle clause** — now asserted (D11); the 0-cycle reading is no longer just a measurement.
- **Audit blind spot** (recorded forward item) — the topology audit checks **feature → row** only; a renamed symbol/string can leave a stale row green (its own round, [#92](https://github.com/gosharplite/tellme/issues/92)).
- **No behaviour change** — `stdout`/`stderr`, the class-phrase vocabulary, flags, exit codes, and every existing gate's semantics are unchanged; the round touches the `Makefile`, a new guard package, a baseline file, an ADR, and a truth row.
