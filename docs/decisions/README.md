# Decisions (ADRs)

Lightweight, immutable **decision records** for this project's own governed artifacts — mirroring the
upstream `aixbdd-tmg` `decisions/` convention. An ADR here also serves as a **named home** for a
project-language declaration (see the upstream `axb-gherkin-and-dsl/STANDARDS.md` "Project Language"
clause, added by [aixbdd-tmg PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8)).

## When to write one

- A change to a project-level rule, convention, or a language / scope declaration.
- A decision that other artifacts (or future rounds) depend on and must be able to cite.
- The resolution of a question escalated to, or answered by, an upstream host rule.

Typo- and editorial-only fixes do not need an ADR.

## Naming & lifecycle

- One file per decision: `docs/decisions/NNNN-<kebab-slug>.md`, zero-padded, ascending.
- `Status` is one of `Proposed` / `Accepted` / `Superseded by NNNN` / `Rejected`.
- Immutable once `Accepted` (except the `Status` line and this index) — supersede with a new ADR
  rather than rewriting history.

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-project-language.md) | Project language: English artifact declaration | Accepted |
| [0002](0002-first-presentation-dependency.md) | First presentation dependency: glamour | Accepted |
| [0003](0003-terminal-detection-isatty.md) | Terminal detection: a real isatty (`golang.org/x/term`) | Accepted |
| [0004](0004-user-global-prompt-log.md) | User-global interactive prompt log (`~/.tellme/global_prompts.jsonl`) | Accepted |
| [0005](0005-tool-call-log-parity.md) | Tool-call log parity: per-call frame cadence, CLI-computed per-call estimate, rune-safe rendering | Accepted (**D7 superseded by [0009](0009-spinner-dual-timer-and-streaming-liveness.md)**; **D1 amended by [0014](0014-yield-policy-owner.md)**) |
| [0006](0006-tool-reason-fold-and-cap.md) | Tool reason fold, trim, and cap: the model-authored reason joins its sibling sanitize+cap family | Accepted |
| [0007](0007-terminal-control-sanitization.md) | Terminal-control sanitization of the `[Tool Output]` block: a new stderr presentation invariant, and a deliberate reference divergence | Superseded by [0008](0008-terminal-safe-lines-and-blank-line-grouping.md) |
| [0008](0008-terminal-safe-lines-and-blank-line-grouping.md) | Terminal-safe `[Tool …]` line policy (generalized) + live turn-output blank-line grouping | Accepted |
| [0009](0009-spinner-dual-timer-and-streaming-liveness.md) | Spinner dual elapsed timer + streaming liveness (supersedes ADR 0005 **D7** — the whole-block pause — only) | Accepted |
| [0010](0010-test-deadline-decoupling.md) | Test deadlines: a test must not hardcode a tight wall-clock budget that is not its subject | Accepted |
| [0011](0011-layer-discipline-gate.md) | Layer-discipline gate: the pinned layer rule + a fail-on-stale violation baseline | Accepted |
| [0012](0012-hermetic-make-go-env.md) | A hermetic `make` Go-toolchain invocation environment (generalises ADR 0011 **D5**; round-020 TD1 is the precedent — its pin stays recipe-local) | Accepted |
| [0013](0013-composition-root-injection.md) | Composition-root extraction: an injected, domain-typed `Dependencies` seam (R2 of #92) | Accepted (**D2**'s unexported-typed, nil-defaulted `RunTUIPrompt` seam *narrowed* by [0017](0017-cli-tui-prompt-decoupling.md)) |
| [0014](0014-yield-policy-owner.md) | Yield-policy owner + `LoopObserver` hook split (R3 of #92; amends ADR 0005 **D1**) | Accepted |
| [0015](0015-loop-presentation-port.md) | Loop presentation port: the loop owns the schedule, `internal/ui` owns the tool-line rendering + the blank-reason predicate (R4 of #92) | Accepted |
| [0016](0016-application-import-ceiling.md) | Application import-ceiling gate (RULE-E): the application tiers' downward-import allow-list (R5 of #92) | Accepted |
| [0017](0017-cli-tui-prompt-decoupling.md) | De-couple `internal/cli` from the TUI prompt: an injected domain `tui.Prompter` port (R5.2 of #92; closes review-deferral F-4) | Accepted |
| [0018](0018-cli-agent-contracts-extraction.md) | De-couple `internal/cli` from the turn loop (re-cut sub-slice 1): extract the loop's crossing contracts to `internal/domain/agent` (R5.3 of #92; baseline unchanged — the edge move is sub-slice 2) | Accepted |
| [0019](0019-agentloop-construction-inversion.md) | De-couple `internal/cli` from the turn loop (sub-slice 2): invert the `AgentLoop` construction/execution into an injected domain port (R5.4 of #92; baseline **2 → 1**; the `→ ui` edge retained) | Accepted |
| [0020](0020-cli-ui-decoupling.md) | Close [#101](https://github.com/gosharplite/tellme/issues/101): de-couple `internal/cli` from `internal/ui` (RULE-E baseline **1 → 0**, the terminal state) + the F-6/F-7/F-8 seam resolutions (R5.5 of #92) | Accepted |
| [0021](0021-ride-alongs-and-records.md) | Ride-alongs: the command tool's construction-time output sink + the suggester's single-owned selection policy; the durable home of the three `#92` records (closes [#115](https://github.com/gosharplite/tellme/issues/115) + [#116](https://github.com/gosharplite/tellme/issues/116)) | Accepted |
| [0022](0022-offline-session-config-and-turns-log.md) | Offline session commands honour `-c` (+ an explicit-`-c` failure) + a per-session `turns.log` and a `-t` reader (closes [#103](https://github.com/gosharplite/tellme/issues/103)) | Accepted |
| [0023](0023-list-default-and-chrome-colour.md) | `-l` takes an optional count (bare `-l` = 1) + green chrome accents on a terminal (four elements; supersedes round-017 D3 for them) | Accepted |
| [0024](0024-e2e-suite-throughput.md) | E2E suite throughput: parallel scenarios by default (`Concurrency=4` + the `TELL_ME_E2E_CONCURRENCY` seam) + a `test-fast` subset that is never the gate | Accepted |
