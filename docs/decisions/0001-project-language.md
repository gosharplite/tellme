# ADR 0001 — Project language: English artifact declaration

- **Status:** Accepted
- **Date:** 2026-09-11
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** [aixbdd-tmg#6](https://github.com/gosharplite/aixbdd-tmg/issues/6) →
  [aixbdd-tmg PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8)
  (`skills/axb-gherkin-and-dsl/STANDARDS.md` → "Project Language" +
  `decisions/0002-project-language-override.md`); tellme grill round #4 residual **R2**.

## Context

Upstream `axb-gherkin-and-dsl/STANDARDS.md` §2/§3 mandated **繁體中文** for feature filenames and
Gherkin parameter keys **unconditionally**. tellme authors its artifacts in **English**, so it shipped
an **unratified deviation** — recorded only in `STATUS.md` and the daily log (a status log, not a named
home). Upstream PR #8 (ADR 0002) added a **Project Language** clause: the default stays 繁體中文, but a
project **MAY** override it, and the declaration **MUST** live in a named home — in order: the project's
`.agents/constitution/shared.md`, a project ADR (`decisions/NNNN-*.md`), or a `spec.md` explicit
constraint.

## Problem

tellme had **no named home** for a language declaration: it deliberately skipped `/axb-constitution`
(no `.agents/constitution/`) and kept no `decisions/` directory. The English decision therefore lived
only in status/journal prose — which the clause explicitly does not accept.

## Decision

tellme declares **English** as its artifact language. The declaration lives in **this project ADR**
(`docs/decisions/0001-project-language.md`) — the second named home in the upstream order, chosen because
tellme has no constitution and a `spec.md` constraint would be per-round (wrong scope for a
project-wide declaration).

Scope (per the upstream clause):

- **Project-declared → English:** feature filenames (§2), Gherkin parameter keys (§3), Gherkin
  sentences / keywords / prose, and DSL prose.
- **Fixed → Chinese, never translated:** the §4/§5 DSL meta-schema column tokens
  (`DSL 句型`, `Gherkin 參數`, `Data Table 參數`, `預設參數`, `StepDef 實作語意`) and the §5.2 channel
  labels (`呈現結果` / `權威狀態` / `再讀確認` / `跨視角` / `不該發生`) and the §5.1 Given / When
  sub-labels. The topology audit (`audit_feature_dsl_topology.py`) requires the first DSL column header
  to be exactly `DSL 句型`; translating any of these makes the audit silently zero-match.

## Alternatives considered

1. **`.agents/constitution/shared.md`** — the strongest (highest-priority) home, but tellme deliberately
   skipped `/axb-constitution`; starting a constitution now is out of scope for this decision.
2. **A `spec.md` explicit constraint** — weakest home: it is per-round, whereas the language
   declaration is project-wide. Rejected.
3. **Do nothing (leave it in `STATUS.md`)** — the defect the clause exists to close. Rejected.

## Consequences

- tellme's English artifacts are now **warranted** (the deviation is ratified) and this ADR is the
  citation a downstream reviewer can point at.
- The mechanical audit still passes: every `dsl.md` keeps the fixed §4/§5 meta-schema tokens unchanged.
- Immutable once `Accepted` — a future language change supersedes this ADR rather than editing it.
