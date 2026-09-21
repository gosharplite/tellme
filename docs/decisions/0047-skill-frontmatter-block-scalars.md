# ADR 0047 — Skill frontmatter: resolve YAML block-scalar descriptions

**Status**: Accepted (round 075)

**Date**: 2026-09-21

**Related**: round 033 (the minimal, on-demand skills catalog + `list_skills`) · `specs/truth/techstack.md` §Skills (*Skills catalog (load)*) · `specs/truth/features/cli/chat/listing-the-available-skills.feature` · the reference `tell-me-go` `internal/infrastructure/skills/file_repo.go` (`parseSkill`).

## Context

The skills loader (`internal/infrastructure/skills/loader.go` → `parseFrontmatter`) is a **line-based** frontmatter reader: it takes, for `name:` / `description:`, only the text after the colon on the same line. A SKILL.md authored with a YAML **block scalar**:

```yaml
---
name: tm-chat-ingroup
description: >
  Chat with other tellme agents …
  … reply in plain text.
---
```

therefore resolves `description` to the literal indicator `">"`, and `list_skills` (via `renderSkills`) prints `- tm-chat-ingroup: > (/…/SKILL.md)` — the agent cannot tell what the skill is from the listing (it must open the file to judge relevance). The frontmatter is **valid YAML**; the reader is simply not YAML-aware.

The motivating evidence is the in-tree catalog (six skills use `>`: three `tm-*`, three `domain-model-*`). The operator directed tellme be fixed to read block scalars (the docs-side reflow of the three `tm-*` template skills was done separately in the seed repo; the three `domain-model-*` are out of scope — not owned).

**The reference shares the limitation** — `tell-me-go`'s `parseSkill` is likewise line-based, so this round is a **deliberate divergence *beyond* the reference: a robustness improvement, not parity.**

## Decision

1. **Resolve block scalars in the reader — stdlib-only.** `parseFrontmatter` recognises a block-scalar indicator as a value and consumes the following indented block, resolving it to text. The reader stays **stdlib-only** (`strings`); **no YAML dependency** is added (matches the round-033 minimal loader; a library for a two-string-key reader is disproportionate) (D1).

2. **The indicator grammar is exact.** A value is a block scalar **iff** its trimmed text matches `[>|][+-]?` (`>`, `|`, `>-`, `|-`, `>+`, `|+`). Anything else — `a > b`, a quoted string, a plain word — stays an **inline** scalar, resolved exactly as before (D2). *(This is what keeps `description: a > b` a literal value.)*

3. **Fold / literal + chomping, resolved to the block's text** (D3): content lines are the following blank-or-indented lines down to the first non-blank, non-indented line (the next key); the first non-blank line's indent is stripped. **Folded (`>`)** joins consecutive non-empty lines with a **single space** and turns a blank line into a **newline**; **literal (`|`)** keeps newlines; **chomping** is default **clip** (one trailing newline), `-` **strip**, `+` **keep**. The resolved value is `strings.TrimSpace`d, so a one-line description has no stray break (keeps the `- name: desc (loc)` listing line intact) while a multi-paragraph block keeps its internal breaks.

4. **Malformed block ⇒ empty ⇒ the file is not a skill.** An indicator with no indented body resolves to `""`, so the file's `name`/`description` requirement fails and it is **skipped silently** — the existing best-effort rule is preserved; the run never fails (D4, NFR-002).

5. **The `Skill` value type and the `list_skills` surface are unchanged.** Only the parsed value of `description` (and `name`) changes; `internal/domain/skills.Skill`, the tool, its bounds, its rendering, and its path-sorted order are untouched (D5).

6. **Recorded divergence from the reference.** The truth row states that tellme now resolves block scalars **and** that the reference's line-based `parseSkill` does not — a deliberate divergence, not parity (D6).

## Consequences

- A skill authored with `description: >` (or `|` / chomping) is listed with its **real text**, so the catalog is self-describing: the agent can judge relevance from the listing instead of opening every file.
- **No regression**: inline (quoted and unquoted) values resolve byte-identically; `a > b` stays literal; a non-skill Markdown file is still skipped; duplicate names stay first-wins; a missing/empty/unreadable directory still yields an empty catalog (I-1).
- The loader stays **best-effort** (I-2), **stdlib-only / POSIX-only / hermetic** (I-4), and the catalog stays **single-source** under `<TELL_ME_HOME>/docs/skills/` (I-5).
- The blast radius is one function (`parseFrontmatter`) + its unit pins + one E2E journey; no domain-model change (the model does not model frontmatter syntax — recorded in `plan.md` §5, ADR 0041's escape hatch).

## Forward (non-blocking)

- **RF-075-1** — the reader supports the block-scalar *subset* (`>`/`|` + chomping + blank-line folding), not the full YAML block grammar (indentation indicators `>2`, leading-space preservation). Adopting `gopkg.in/yaml.v3` in a future round would subsume it.
- **RF-075-2** — `name` is parsed with the same rule but has **no** E2E carrier (only the motivating `description` case); a block-scalar `name` is unit-pinned only.
- **RF-075-3** — the reference (`tell-me-go`) is **not** fixed; this remains a recorded divergence, not a parity claim. Revisit the divergence note if the reference is ever aligned.
