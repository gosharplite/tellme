# System Analysis — skill frontmatter block scalars (round 075)

**Plan Package**: `specs/plans/075-skill-frontmatter-block-scalars`

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The skills catalog / `list_skills` prompt surface | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `chat` module gains a block-scalar Rule/Example + a Given/Then pair |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state change (the catalog is not persisted) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The frontmatter reader | the implementation | `internal/infrastructure/skills/loader.go` — `parseFrontmatter` resolves a block-scalar `name`/`description` (indicator `[>|][+-]?`; fold/literal + chomping; `TrimSpace`) |
| **W2** | The truth row | `/axb-technical-research` (done) | `specs/truth/techstack.md` §Skills *Skills catalog (load)* MODIFY (`research.md` D1–D6; **divergence note**) |
| **W3** | The executable CLI contract | `/axb-dsl-refine` | the `chat` module Rule/Example + a Given that authors a folded block scalar + a Then that reads the described text |
| **W4** | The plan-side acceptance | `/axb-spec-by-example` (done) | `features/acceptance/listing-a-block-scalar-skill.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **user-visible** (the description text in the `list_skills` result), so `/axb-spec-by-example` is **NOT** NOOP.

## 4. Unchanged surfaces (invariants)

- Inline (quoted and unquoted) `name`/`description` values resolve byte-identically; a non-skill Markdown file is skipped; a duplicate name is first-wins; a missing/empty dir yields an empty catalog (`spec.md` I-1).
- The loader never fails a turn; a malformed/unsupported file is skipped silently (`spec.md` I-2).
- The listing contract is unchanged — name + description + location, path-sorted; an empty catalog is a *result*, not an error (`spec.md` I-3).
- No new dependency; stdlib-only; POSIX-only; hermetic (`spec.md` I-4).
- The catalog stays single-source (`<TELL_ME_HOME>/docs/skills/`); no `.skills/`, no `skillssh` tools (`spec.md` I-5).

## 5. Domain model (ADR 0041)

**Not modelled, and that is recorded here** (the same-PR rule's escape hatch): the round changes the **parsing** of a frontmatter scalar, not a modelled entity or invariant. The domain model's `Skill` entity describes a guidance block's `name`/`description`/`location` and its on-demand delivery — it does **not** model the frontmatter *syntax* the loader reads, so there is nothing to update and `modelith-check` stays green. (`techstack.md`'s *Skills catalog (load)* row is the truth home for the reader's behaviour.)
