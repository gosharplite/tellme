# Phase 0 Research: layer-discipline gate + violation baseline (Round 042)

Topic: add a **layer-discipline verification gate** to `make verify` that enforces an **import-direction rule** over the module's layers, shipping with a **committed baseline** of the layer violations that already exist, so `dev` is green today, a **new** illegal import fails the standard gate, and the existing count can only **ratchet down** to 0 across R2–R4 ([#92](https://github.com/gosharplite/tellme/issues/92) → [#93](https://github.com/gosharplite/tellme/issues/93)).

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` + stdlib `testing`), provider transports, skills, MCP, and the presentation packages were locked in rounds 001–041. The system still has **one CLI end**, adds **no new system end**, **no external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; BDD techstack = `godog` on the built binary; E2E + unit strategy) and are **not re-decided** (Decision 10). **IN**: the gate, its baseline, its self-tests, and the truth rows. **OUT**: fixing any baselined violation (R2–R4), any `tellme` binary behaviour, and any existing gate's semantics.

---

## Decision 1: A **broad** import-direction rule over a pinned layer ranking (clarify Q1 → Option 2)

- **Decision**: the gate enforces a general *no-upward-import* rule over a **pinned layer ranking** (low → high): `internal/domain/**` (0, pure) → `internal/config`, `internal/home` (1, shared utilities) → `internal/app/**` (2) → `internal/infrastructure/**` (3) → `internal/agent` (4) → `internal/ui`, `internal/ui/tui/**` (5) → `internal/cli` (6); `cmd/tellme` (7, composition) and `tests/**` are **exempt**. A governed package's internal import of a **higher** tier is a **violation**. Consequence: `internal/agent → internal/ui` **is** a violation, so the baseline is **8**.
- **Rationale**: the operator chose the broad form (the reference's `verify-architecture` shape) — a general rule catches future misuse, not just the one `cli → infrastructure` class. Ranking `ui` above `agent` is the honest reading of presentation-vs-application direction; `config`/`home` are bottom shared utilities, so `infrastructure → config` (which the module ships) stays **legal** (otherwise the baseline would balloon for a non-defect).
- **Alternatives considered**:
  - **Narrow target-rule** (`cli`/`ui`/`app` ↛ `infrastructure` + domain purity; baseline = 7) — smallest, but ignores the `agent → ui` edge the operator chose to flag — rejected (Q1/Option 2).
  - **Broad with ranks chosen to keep 7** (rank `agent` ≥ `ui`) — future-proof but pre-empts R3/R4's ownership of the loop→`ui` coupling by declaring it legal — rejected.

## Decision 2: The guard is a **build-tagged Go guard test** invoked from a `Makefile` `verify-architecture` target

- **Decision**: implement the gate as a Go test guarded by a build tag (`//go:build arch`) that (a) obtains the module's import graph from the **Go toolchain itself** (`go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...`, parsed from stdlib `os/exec` + `strings`), (b) filters to governed packages and internal imports, (c) applies the Decision-1 ranking, and (d) diffs the violation set against the baseline (Decision 3). A new `Makefile` target `verify-architecture` runs `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch` and is added to the `verify` aggregate + `make help`.
- **Rationale**: using the **Go toolchain's own resolution** (`go list`) is precise — it respects build tags, `internal/` visibility, and the real package set, so a build-tagged illegal import cannot hide (a naive per-file grep would miss it). It mirrors the reference (`go test -tags=arch … -strict-arch=true` + `modelith-layers`) and keeps the guard stdlib-only (no `golang.org/x/tools`). The `-tags=arch` gate keeps the guard out of the default `go test ./...` run. The guard package carries a trivial untagged `doc.go` so `go build ./...` never sees a "build constraints exclude all Go files" error.
- **Alternatives considered**:
  - **Makefile grep** (the `verify-mcp-sdk-confinement` style) — simplest, but cannot resolve package tiers, cannot respect build tags (a build-tagged illegal import hides), and would mis-rank nested packages — rejected.
  - **Pure `go/build` stdlib walk** — avoids `go list`, but re-implements package resolution and can diverge from the real build graph — rejected.

## Decision 3: A committed, sorted **baseline** with **fail-on-stale** (clarify Q3 → Option 1)

- **Decision**: commit a baseline file listing the currently-known violations as **sorted lines** `source-path → imported-path` (the gate's own output form; generated from the gate, never hand-transcribed), parsed by the guard. Behaviour: a violation **not** in the baseline **fails** (new violation); a baseline entry that no longer violates **fails** (stale) and the message says "remove it from the baseline"; a violation removed from the baseline while still present **fails**.
- **Rationale**: fail-on-stale makes the baseline a self-policing **ratchet** — the count can only shrink toward truth, and each fix (R2–R4) **must** edit the baseline (visible in the diff), which is exactly what makes [#92](https://github.com/gosharplite/tellme/issues/92)'s "→ 0" claim falsifiable.
- **Alternatives considered**:
  - **Warn-on-stale** — the baseline can silently rot (stale lines accumulate) — rejected (Q3/Option 1).
  - **A JSON baseline** — heavier and less diff-friendly than sorted lines for a one-entry-per-violation ratchet — rejected.

## Decision 4: Package scope — production `internal/**`; `_test.go` governed (adds nothing); `cmd/**` + `tests/**` exempt (clarify Q2, resolved by measurement)

- **Decision**: the gate governs **production `internal/**`** packages; `internal/**` **test imports** are also evaluated (measured 2026-09-17: they add **no** entries), and `cmd/tellme` (composition top) and `tests/**` (test-support) are **exempt**.
- **Rationale**: measured, the choice does **not** change the baseline — no `internal/**` test file imports anything upward, and the only outward harness importer is `tests/e2e`(+`steps`), which is test-support by design. Exempting `cmd`/`tests` matches the composition-root/test-support intent and avoids a false-positive surface.
- **Alternatives considered**:
  - **Production-only (exclude all `_test.go`)** — equivalent baseline; a gratuitous blind spot — rejected.
  - **Govern `tests/**` too** — the harness legitimately imports `internal/infrastructure/**`; governing it would add noise for test-support code — rejected.

## Decision 5: Wiring + truth record

- **Decision**: add the `verify-architecture` target to the `Makefile` (**member of `make verify`**; add to `.PHONY` + `make help`). Record the gate, the layer ranking, and the baseline policy in **`specs/truth/techstack.md` (Build & Tooling)** — a new **layer-discipline gate** row — and add the member to the **Task runner** row's `verify` aggregate list.
- **Rationale**: the gates live in the `Makefile`; a new gate is a real techstack **MODIFY** (the `verify` aggregate is named in truth), so the round must record it (no unevidenced NOOP).
- **Alternatives considered**:
  - **A standalone target not in `verify`** — a new violation would pass the standard gate — rejected.
  - **Record only in the plan package** — the frozen package does not carry current truth — rejected (the round-035 session lesson: a durable home).

## Decision 6: No new dependency; hermetic and host-independent; deterministic

- **Decision**: the gate uses only the Go toolchain + `make` + stdlib (`os/exec`, `strings`, `sort`); it makes **no** network request (warm module cache), does **not** depend on the host `GOOS`/`GOARCH` or an ambient `CGO_ENABLED`, does **not** use `time.Sleep`, and emits a stable, sorted report. `go.mod`/`go.sum` are unchanged.
- **Rationale**: the guard reads the module's own graph (identical on every host); hermeticity mirrors `verify-cross-compile`'s rationale.
- **Alternatives considered**:
  - **A third-party import-graph library (`golang.org/x/tools/go/packages`)** — a new dependency for no gain — rejected.

## Decision 7: The gate must not forbid the legitimate seams

- **Decision**: the guard MUST NOT flag `agentTools()` (the parameterless, read-free assembler the round-031 well-formedness gate iterates) or any other currently-legal edge under the pinned ranking; the baseline (Decision 3) is the mechanism for the existing-but-unfixed edges, and the guard's self-test must confirm the guard's own package adds no violation.
- **Rationale**: a gate that breaks a legitimate seam would be undone immediately; the baseline handles the known-unfixed edges.
- **Alternatives considered**:
  - **An allow-list of "fine" edges as code** — duplicates the baseline — rejected.

## Decision 8: Governance — **no new ADR** for R1 (recorded)

- **Decision**: R1 records **no** new ADR. The layer ranking, the baseline, and the fail-on-stale policy live in `techstack.md` (Build & Tooling). The ADR obligation from [#92](https://github.com/gosharplite/tellme/issues/92) attaches to **R3**'s yield policy, not here.
- **Rationale**: R1 settles no decision that other artifacts need to *cite as a superseding rule*; the techstack row is the durable home. (If a later round needs the layer model as a citable rule, it adds an ADR then.)
- **Alternatives considered**:
  - **A short layer-model ADR now** — proportionate only if the ranking is contended; deferring keeps R1 minimal — rejected (deferrable).

## Decision 9: Testing & BDD techstack unchanged; `/axb-dsl-refine` is `NOOP`

- **Decision**: no new system end and no change to the BDD techstack or strategy. `/axb-api-plan` = `NOOP`, `/axb-data-plan` = `NOOP` (the baseline is a repo artifact, not runtime state), `/axb-dsl-refine` = `NOOP` (the gate is a **dev surface**, not the `tellme` CLI end — no user-facing CLI contract changes; the acceptance carrier is the gate + its self-test, spec A3), `/axb-ui-plan` = skipped, `/axb-spec-by-example` = **NOOP** (no user-facing business journey — precedent: rounds 020/031/036/041).
- **Rationale**: the round's verification is the gate's own exit code + falsifiability witnesses, not a Gherkin scenario; forcing CLI Gherkin for a `make verify` gate would be ceremony and would risk an `acceptance-coverage` mismatch.
- **Alternatives considered**:
  - **Author plan-side acceptance Gherkin for `make verify`** — no user-facing CLI behaviour to express — rejected (mirrors round 020's `/axb-spec-by-example` = skipped).
  - **Add a CLI-truth `dsl.md` row for the gate** — the gate is not the `tellme` binary's CLI contract — rejected.

## Decision 10: Testing & BDD techstack must-asks settled (unchanged)

- **Decision**: the three AIxBDD must-ask questions (system ends; BDD techstack; test strategy) remain those of the standing `techstack.md`; the round adds no CLI interface truth, no API surface, and no data model.
- **Rationale**: a build/quality-pipeline change with a single CLI end already fixed.
- **Alternatives considered**:
  - **Re-open the techstack** — nothing in this round changes the stack — rejected.

---

## Residual risks / forward links

- **Ranking judgements** — `config`/`home` ranked bottom (so `infrastructure → config` is legal) and `ui` ranked above `agent` (so `agent→ui` is a violation, baseline entry 8). If a later round disagrees, it edits the ranking + re-measures the baseline (a `techstack` MODIFY).
- **`go list` cost** — the guard shells to `go list`; a cold module cache resolves modules like any build (no new module, no new service). Runtime is a few seconds — bounded.
- **Guard self-exclusion** — the guard package (`tools/arch`) must not itself violate the rule (it imports only stdlib + `os/exec`); its self-test pins this.
- **Build-tag hygiene** — the guard is `-tags=arch`-excluded from `go test ./...`; an untagged `doc.go` keeps the directory buildable without the tag.
- **Ratchet breadth** — the 8th entry (`agent → ui`) is R3/R4's to remove; the count reaches **0 across R2–R4** (how [#92](https://github.com/gosharplite/tellme/issues/92) AC1 is read).
- **`modelith-layers` analogue** — the reference's domain-model-as-code layer check is **not** adopted (tellme has no modelith toolchain) — recorded, not a gap.
- **No behaviour change** — `stdout`/`stderr`, the class-phrase vocabulary, flags, exit codes, and every existing gate's semantics are unchanged; the round touches the `Makefile`, a new guard package, a baseline file, and a repo doc/truth row.
- **Audit blind spot** (recorded forward item) — the topology audit checks **feature → row** only; a renamed symbol/string can leave a stale row green (its own round, [#92](https://github.com/gosharplite/tellme/issues/92)).
