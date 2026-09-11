# Phase 0 Research: tellme Follow-up Cleanups (Round 002)

Topic: the round-002 cleanup slice — (1) **remove the `--json` diagnostic flag** (reverses round-001
FR-013), (2) **freeze the operator-facing failure contract** (exact stderr messages + numeric exit
codes), (3) **F9** — fast unit tests for the pure resolution helpers, and (4) **quality-gate hardening**
(an ignored-error gate + a dependency-vulnerability gate).

Scope note: the language (**Go**), module, layout, and base toolchain are already fixed by round-001
`specs/truth/techstack.md`. This research settles only the round-002 **deltas** — the test-strategy
amendment (F9), the two adopted gates, and the unit-test tooling. The `--json` removal and the
message/exit-code freeze are **behaviour** (spec + interface truth), not technology: they land in
`spec.md` and `specs/truth/features/cli/**`, so the only techstack row they touch is the CLI flag list
(reconciled here).

## Decision 1: Test strategy — E2E acceptance path plus fast unit tests for pure helpers (amends round-001 Decision 5)

- **Decision**: Keep round-001's **E2E black-box** strategy as the *acceptance* path, and **add
  table-driven unit tests** (Go stdlib `testing`) for the **pure resolution helpers** — the effective
  mode, the effective selected provider, provider-registry membership, and workspace creation/reuse.
  The unit tests **complement** the E2E path; they never replace it.
- **Rationale**: round-001 Decision 5 chose E2E-only because every round-001 scenario is defined at the
  process boundary. That reasoning still holds for the *acceptance* path — but it never forbade, and
  (review finding **F9**) ought not to forbid, **directly** testing the pure helpers, which is where the
  env-over-file precedence matrix and workspace idempotency actually live. Those helpers are pure
  functions of strings/paths, so unit tests are cheap, deterministic, offline, and give a sharper
  failure signal than a subprocess round-trip. The E2E suite stays the source of truth for the
  acceptance journeys (`acceptance-coverage`); the unit layer is an additive safety net.
- **Alternatives considered**:
  - **Keep E2E-only (D5 unchanged)** — fewer moving parts, but leaves the precedence matrix and
    idempotency covered only indirectly through the CLI boundary; F9 explicitly asks otherwise.
  - **Move to unit-first** — rejected: the acceptance journeys (exit codes, streams, workspace effects)
    are defined at the process boundary, so E2E must remain the acceptance path.

## Decision 2: Static-analysis and vulnerability gates — adopt `golangci-lint` (with `errcheck`) and `govulncheck` (settles round-001 Decision 7's deferrals)

- **Decision**: Adopt **`golangci-lint`** — with **`errcheck`** enabled — as the lint **aggregator**
  gate, and **`govulncheck`** as the **dependency-vulnerability** gate. Both are wired into the Makefile
  (`make verify` / `make check`) and require a small committed **`.golangci.yml`** policy artifact.
- **Rationale**: round-001 Decision 7 deferred both — naming `golangci-lint` the *"intended next-slice
  aggregator"* — and logged an explicit **unchecked-error residual**: the code does file/YAML I/O
  (`os.ReadFile`, `yaml.Unmarshal`, `os.Stat`, `os.MkdirAll`), and `staticcheck` does **not** cover
  `errcheck`'s class. Adopting `golangci-lint` with `errcheck` closes that residual (spec NFR-004);
  `govulncheck` adds the dependency-vulnerability gate (spec NFR-005). Round-001's feasibility pass
  confirmed both binaries are already present in `$GOPATH/bin`, so there is **no install cost** — only
  the one-time `.golangci.yml` policy artifact.
- **Alternatives considered**:
  - **Standalone `errcheck` binary only** — narrower, and `golangci-lint` (the named aggregator)
    subsumes it.
  - **Keep deferring both** — leaves a known error-handling blind spot and no vulnerability gate,
    contrary to this round's NFRs.
  - **`staticcheck` alone** — already adopted, but it does not check unchecked errors, which is exactly
    the residual in scope.

## Decision 3: Unit-test tooling — Go stdlib `testing` (no new dependency)

- **Decision**: The pure-helper unit tests use the Go standard library **`testing`** package with
  **table-driven** cases and `t.Run` subtests; **no** assertion/mocking library is added.
- **Rationale**: the helpers are pure functions of strings/paths with no external seam, so stdlib
  `testing` is sufficient and adds no dependency; the E2E suite already uses stdlib `testing` as its
  host harness, so the unit layer keeps the same tool and style. Determinism carries over (no
  `time.Sleep` for synchronization — `verify-no-test-sleep` parity).
- **Alternatives considered**:
  - **`testify` (assert/require)** — ergonomic, but adds a dependency for trivial equality checks;
    unnecessary at this size.

## Residual risks / forward links

- The `--json` removal **reverses round-001 FR-013**; round-001's plan package stays frozen history. The
  affected interface truth (`specs/truth/features/cli/diagnostics/**`, `.../usage/**`) is a
  `DELETE`/`MODIFY` owned by `/axb-dsl-refine`, not a techstack concern.
- `.golangci.yml` linter selection is a **policy choice**; this round enables `errcheck` at minimum.
  Broader linter selection can be revisited as the codebase grows (round-001 D7's reason for deferring
  the aggregator was precisely "a curated policy artifact").
- Whether the unit layer should later extend to `internal/cli`'s flag parsing is **deferred**; this
  round covers only the pure resolution helpers (F9's stated scope).
- The exact stderr strings and numeric exit-code values are frozen by this round as **interface
  contract** (recorded in `specs/truth/features/cli/**` by `/axb-dsl-refine`), not here.
