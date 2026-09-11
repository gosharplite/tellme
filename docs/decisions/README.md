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
