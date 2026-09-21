# Technical Research — skill frontmatter block scalars (round 075)

**Plan Package**: `specs/plans/075-skill-frontmatter-block-scalars`
**Owner**: `/axb-technical-research` (truth owner: `specs/truth/techstack.md`)
**Clarify**: resolved at specify time (**not escalated — 0 questions**). No new question rose here.

---

## Context (grounded 2026-09-21, `dev` @ `3dfc3a4`)

`internal/infrastructure/skills/loader.go` → `parseFrontmatter` is a **line-based** reader. For each frontmatter line it does `key, val, _ := strings.Cut(line, ":")` and takes `unquote(strings.TrimSpace(val))` — **the same line only**. For a YAML **block scalar**:

```yaml
---
name: tm-chat-ingroup
description: >
  Chat with other tellme agents …
  … reply in plain text.
---
```

the value after the colon is the **indicator** `>`, so `description` becomes the literal string `">"`. `renderSkills` (`internal/infrastructure/tools/skills.go`) then prints `- tm-chat-ingroup: > (/…/SKILL.md)`. The frontmatter is **valid YAML**; the reader is simply not YAML-aware. Observed on the in-tree catalog (six skills use `>`: three `tm-*` + three `domain-model-*`).

**The reference has the same limitation** — `tell-me-go` `internal/infrastructure/skills/file_repo.go` `parseSkill` is also line-based (`strings.HasPrefix(line, "description:")` → `strings.TrimPrefix`), so it would also yield `">"`. **This round is therefore a deliberate divergence *beyond* the reference — a robustness improvement, not parity.**

---

## 決策 1 — Resolve block scalars in the existing stdlib reader (no YAML dependency)

- **Decision**: extend `parseFrontmatter` to recognise a **block-scalar indicator** as a value and consume the following indented block, resolving it to text. Keep the reader **stdlib-only** (`strings`), matching the round-033 "minimal, stdlib-only" loader.
- **Rationale**: the fix is a small, local extension to one function; a real YAML parser (`gopkg.in/yaml.v3`) would be a new dependency for a single scalar-reader that only ever needs `name`/`description`. The reference's own reader is hand-rolled, so staying hand-rolled preserves the established shape (and I-4/NFR-001 default to no new dependency).
- **Alternatives considered**:
  - **Adopt `gopkg.in/yaml.v3`** — full YAML correctness, but a new dependency + a `techstack.md` dependency row for one reader that reads two string keys; a disproportionate change (rejected; recorded as an option, not taken).
  - **Author-side fix only** (reflow the skills to inline) — done separately for three `tm-*` template skills; it does not fix the general case (future / third-party `skills.sh`-style skills authored with `>`) and the operator directed the product fix (rejected as the product answer).

## 決策 2 — Recognise the indicator grammar; never mistake an inline value for one

- **Decision**: a value is a block scalar **iff** its trimmed text matches the indicator grammar `[>|][+-]?` (i.e. `>`, `|`, `>-`, `|-`, `>+`, `|+`). Anything else (`a > b`, `"…"`, a plain word) is treated as an inline scalar exactly as today.
- **Rationale**: `foo > bar` must stay a literal inline value (edge case in `spec.md`); only a *bare* indicator opens a block.
- **Alternatives considered**:
  - **Treat any value starting with `>`/`|` as a block** — would mis-read `> inline text` (rejected).

## 決策 3 — Fold / literal + chomping, resolved to the block's text (YAML 1.2)

- **Decision**: the block's content lines are those following the indicator that are **blank or indented** (stopping at the first non-blank, non-indented line — the next key). The block's indentation is the first non-blank content line's leading whitespace, stripped from every line. Then:
  - **folded (`>`)** — consecutive non-empty lines join with a **single space**; a blank line becomes a **newline** (paragraph break);
  - **literal (`|`)** — lines join with a **newline** (breaks preserved);
  - **chomping** — the indicator's `-`/`+`/default is accepted syntactically, but the resolved value is `strings.TrimSpace`d, which **normalizes trailing-newline chomping**: for the trimmed scalar, `>`/`>-`/`>+` (and `|`/`|-`/`|+`) are **indistinguishable**. The chomping branches preserve YAML fidelity (and would matter for an untrimmed value); they are **inert** for the frontmatter scalar. *(fold F-2.)*
  and the resolved value is `strings.TrimSpace`d (so a single-line description carries no stray break, while a multi-paragraph block keeps its internal breaks).
- **Rationale**: matches the semantics a skill author expects; the trim keeps `description` a clean scalar for the one-line listing entry, mirroring the existing trim behaviour (A3).
- **Alternatives considered**:
  - **Preserve the raw block verbatim** — would leak a trailing newline into the `- name: desc (loc)` line and break the listing format (rejected).
  - **Full YAML folding subtleties (indentation-based folding, leading-space preservation)** — out of scope for a two-key reader; the fold/literal/blank-line/chomp subset covers every authored case (rejected as over-engineering).

## 決策 4 — Malformed block ⇒ empty ⇒ the file is not a skill (best-effort, unchanged)

- **Decision**: an indicator with **no indented body** resolves to the empty string, so the file's `name`/`description` check fails and the file is **skipped silently** — the existing best-effort rule (NFR-002) is preserved; the run never fails.
- **Rationale**: I-2/NFR-002; a malformed skill file must not break a turn.

## 決策 5 — The `Skill` value type and the `list_skills` surface are unchanged

- **Decision**: only the **parsed value** of `description` (and `name`) changes; `internal/domain/skills.Skill`, the `list_skills` tool, its bounds, its rendering, and its path-sorted order are untouched.
- **Rationale**: the round is a reader-robustness fix, not a surface change; it keeps the blast radius to one function + its tests.

## 決策 6 — Witnesses: unit pins for the grammar/folding + one E2E journey

- **Decision** (the pins as shipped — corrected at fold **F-1**): unit pins are `TestLoadFoldedBlockScalarDescription` (folded `>`), `TestLoadLiteralFoldedJoining` (literal `|` + a `>-` fold, the latter documenting chomping acceptance), `TestLoadMultiParagraphFoldedBlockScalar` (blank line ⇒ newline — **FR-005**), `TestLoadQuotedScalarsUnchanged` (double/single-quoted regression), `TestLoadCRLFBlockScalar` (CRLF + folded), `TestLoadInlineGreaterThanIsNotABlock` (inline `>` stays literal), `TestLoadBlockScalarWithNoBodyIsSkipped`, and `TestLoadBlockScalarName`. The E2E adds one journey — a skill authored with a folded description is listed by its folded text — whose Then asserts the expected entry **and** (fold **N-1**) that the entry is not the bare indicator; this is the `>` falsifiability witness (dropping the fix reddens it, showing `">"`).
- **Rationale**: the parsing matrix is a unit-level fact; the listing effect is the acceptance-level carrier. **Chomping** has no independent witness by construction — the mandated `TrimSpace` normalizes trailing-newline chomping (D3, fold **F-2**), so no assertion on it can fail; the chomping branches are fidelity/defensive, not a verified behaviour.

---

## Residual risks / forward items (for the ADR §Forward)

- **RF-075-1** — the reader supports the block-scalar subset (`>`/`|` + chomping + blank-line folding), not the full YAML block grammar (indentation indicators like `>2`, leading-space preservation). A future round adopting `yaml.v3` would subsume it.
- **RF-075-2** — `name` is parsed with the same rule but has no E2E carrier (only the motivating `description` case); a block-scalar `name` is unit-pinned only.
- **RF-075-3** — the reference (`tell-me-go`) is **not** fixed; this remains a recorded divergence, not a parity claim. If the reference is ever aligned, this ADR's divergence note should be revisited.
