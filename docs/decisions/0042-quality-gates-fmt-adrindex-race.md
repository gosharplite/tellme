# ADR 0042 — Quality gates: format check, ADR-index gate, race target

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** operator request (a quality pass; no anchor issue) · **ADR 0011** (layer gate + `verify` ratchet form) · **ADR 0012** (hermetic `make` env — a gate must not be a spurious-red generator) · **ADR 0026** (governance) · **ADR 0030** (modelith; `modelith-check`) · **ADR 0041** (model drift guard) · the reference `tell-me-go` (`Makefile` `check`/`check-full`, `verify-adr-index`, `test-race`)

## Context

A quality audit (2026-09-20) found three cheap, precise gaps, each verified green on `dev` @ `8a986bd`:

1. **No format check in the pipeline.** `make fmt` *mutates* (`go fmt ./...`); nothing *checks* formatting, so a non-gofmt-clean file is caught only by the manual closeout step.
2. **No ADR-index gate.** `docs/decisions/README.md` is hand-maintained; nothing enforces that every ADR on disk is listed, or that no number is claimed twice. (It is currently consistent — 41 files, 41 rows — but only by diligence; the ADR-0041 pass had to edit it by hand.)
3. **No race detector anywhere — and no CI.** tellme has genuinely concurrent code (the round-040 UI coordinator: a mutexed writer + a spinner admit goroutine; the telemetry sampler; the round-055 parallel E2E harness), yet nothing runs `-race`. The subprocess-based E2E contract cannot instrument the built binary, so a data race is exactly the class it cannot catch.

## Decision

**D1 — `verify-fmt` (format check) joins `make verify`.** A `gofmt -l` that is non-empty fails the gate ("run `make fmt`"). Toolchain-native (no dependency), fast, hermetic, zero false positives.

**D2 — `verify-adr-index` (ADR-index consistency) joins `make verify`.** Every `docs/decisions/[0-9]*.md` file must be listed in `docs/decisions/README.md`, and no ADR number may be claimed twice. Fast, hermetic, zero false positives.

**D3 — `test-race` ships as a standalone target, NOT a `verify` member.** A `-race` run is expensive (measured **~67 s** for the package-by-package `./...` loop; a single `go test -race ./...` is ~28.6 s, ~11.6 s for `./internal/...`), so it stays out of the fast `verify` aggregate and is run on demand (a **pre-push / closeout** check, the reference's `check-full` posture). It iterates **package-by-package** (an AI-safe per-package loop, not one `go test -race ./...`) so a single slow/contended package cannot time the whole run out — the loop's per-package build overhead is the cost of that safety; scope defaults to `./...` and is overridable with `RACE_PKGS`.

**D4 — the audit's other candidates stay rejected.** `vet`'s duplication (it is the toolchain-native minimal gate; `staticcheck` was redundant, `vet` is only *wasteful* — kept, ADR 0042 records the distinction); coverage / reachability-orphan tooling (**declined**, #144 — E2E-subprocess paths read 0 %); the 5 pre-existing topology-audit DSL errors (the audit script is external — `aixbdd-tmg` — and the check is *carried*, not gated).

## Consequences

- `make verify` grows by **two fast members** (`verify-fmt`, `verify-adr-index`) — measured **13.9 s** end-to-end, still a hermetic pre-check.
- **No new dependency.** `verify-fmt` uses the toolchain's `gofmt`; `verify-adr-index` is `grep`/`sed`; `test-race` uses the toolchain's `-race`.
- A formatting or index regression now fails the pipeline rather than surviving to closeout.
- **`make verify` still does NOT run the E2E contract** (that is `make test`); the race target is the pre-push addition. The check-vs-verify distinction is unchanged.

## §Forward

- **RF-042-1** — no `make check` / `make check-full` aggregate ships yet (`verify` + `test` + `test-race`); deferred (purely additive, but a naming/scope decision of its own).
- **RF-042-2** — `verify-fmt` covers `gofmt` only, not `goimports`/`gofumpt`; a formatter-set decision is deferred (`golangci-lint fmt` v2 is the vehicle if wanted).
- **RF-042-3** — `test-race` is not wired into any automated runner (no CI); it relies on the closeout/pre-push discipline.

## References

- `Makefile` (`verify-fmt`, `verify-adr-index`, `test-race`, the `verify` aggregate)
- **ADR 0011** (the `verify` ratchet form) · **ADR 0012** (hermeticity) · **ADR 0030** (`modelith-check`) · **ADR 0041** (model drift guard)
- reference `tell-me-go/Makefile` (`verify-adr-index`, `test-race`, `check`/`check-full`)
