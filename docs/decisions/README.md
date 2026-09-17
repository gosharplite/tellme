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
| [0005](0005-tool-call-log-parity.md) | Tool-call log parity: per-call frame cadence, CLI-computed per-call estimate, rune-safe rendering | Accepted (**D7 superseded by [0009](0009-spinner-dual-timer-and-streaming-liveness.md)**) |
| [0006](0006-tool-reason-fold-and-cap.md) | Tool reason fold, trim, and cap: the model-authored reason joins its sibling sanitize+cap family | Accepted |
| [0007](0007-terminal-control-sanitization.md) | Terminal-control sanitization of the `[Tool Output]` block: a new stderr presentation invariant, and a deliberate reference divergence | Superseded by [0008](0008-terminal-safe-lines-and-blank-line-grouping.md) |
| [0008](0008-terminal-safe-lines-and-blank-line-grouping.md) | Terminal-safe `[Tool …]` line policy (generalized) + live turn-output blank-line grouping | Accepted |
| [0009](0009-spinner-dual-timer-and-streaming-liveness.md) | Spinner dual elapsed timer + streaming liveness (supersedes ADR 0005 **D7** — the whole-block pause — only) | Accepted |
