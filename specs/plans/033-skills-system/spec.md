# Feature Specification: skills system — load & list pre-loaded skills (round 033)

**Feature Branch**: `033-skills-system`

**Created**: 2026-09-16

**Status**: Draft — scope resolved from an operator request. Clarify locked in-session (Q1, Q3; Q2 moot).

**Input**: Operator request: *"I want tellme to have a skills system, but not the `skillssh` toolkit complication. tellme only needs to load skills in `ait-*/docs/skills`. Is it possible? … If I ask tellme to list all pre-load skills, will tellme be able to list skills?"*

**Scope note**: a **new-capability** round. tellme gains a **minimal skills system**: it loads the skill definitions under `<TELL_ME_HOME>/docs/skills/` and surfaces them to the agent **on demand** through a read-only **`list_skills`** tool. Skill **content is NOT auto-injected** into the prompt — the agent lists the skills and, when it wants one, opens that skill's file with the existing `read_files` tool. The reference (`tell-me-go`) additionally ships a **skills.sh ecosystem** (a `.skills/` source + `search_skills` / `install_skill` / `remove_skill`) and **automatic relevance-based injection**; **both are explicitly out of scope** here (the operator asked to drop the `skillssh` "complication"). tellme currently has **no** skill subsystem at all — it never reads `docs/skills/`.

**Clarify (locked in-session, one decision at a time)**:

- **Q1 → 1** — surface model: **on-demand only**. tellme lists the skills and the agent reads a skill's content via the existing reader; **no automatic prompt injection and no relevance selector**.
- **Q2 → moot** — with Q1 → 1 there is no injected block, so the manifest-vs-selector question does not apply.
- **Q3 → 1** — the listing is delivered as a **model-callable, read-only `list_skills` tool** (not an offline CLI flag).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List the environment's pre-loaded skills (Priority: P1)

As an operator whose environment carries a `docs/skills/` library, I want tellme to load those skill definitions and let the agent list them, so that when I ask tellme to list its pre-loaded skills it answers with the actual set that is present — rather than claiming it has none.

**Why this priority**: This is the capability itself — the whole point of the round. Without loading and listing, nothing else in the round has value; it is the first value to prove and the foundation the other story builds on.

**Independent verification**: place a known set of skill definitions under `<TELL_ME_HOME>/docs/skills/`, run a prompt that asks tellme to list its skills, and confirm the answer names exactly that set (each skill's name and description).

**Acceptance Scenarios**:

1. **Given** a skills directory containing N skill definitions, **When** the agent is asked to list its skills, **Then** tellme returns each loaded skill's **name** and **description** (N entries).
2. **Given** no skills directory, or an empty one, **When** the agent is asked to list its skills, **Then** tellme reports that no skills are loaded (no failure).
3. **Given** a skills directory with a skill that has nested non-skill Markdown (e.g. its `rules/*.md`), **When** the agent is asked to list its skills, **Then** only the skill definitions appear — the nested non-skill Markdown is not listed.

**Functional Requirements (FR)**:

- **FR-001**: tellme MUST load skill definitions from `<TELL_ME_HOME>/docs/skills/` (resolved through the existing runtime-home resolver) when a prompt-bearing turn runs; a missing, empty, or unreadable directory MUST yield an **empty catalog** and MUST NOT fail the run.
- **FR-002**: tellme MUST expose the loaded catalog to the model through a read-only agent tool (`list_skills`) that returns, for each skill, at least its **name** and its **description**.
- **FR-003**: Skill discovery MUST be **recursive** and MUST NOT mistake non-skill Markdown files (a skill's `rules/*.md`, `STANDARDS.md`, examples, etc.) for skills.
- **FR-004**: When no skills are loaded, the tool MUST report the **empty catalog** (not an error), so the agent can answer truthfully that none are present.

**Non-Functional Requirements (NFR)**:

- **NFR-001**: Loading MUST be **best-effort**: a malformed or unreadable individual skill file is skipped (with a diagnostic) and MUST NOT fail the run or the boot.

---

### User Story 2 - Use a listed skill on demand (Priority: P2)

As an operator, I want the agent to be able to open a listed skill's content on demand, so the skill actually gets used within a turn — without tellme having to pre-inject any skill into the prompt.

**Why this priority**: Listing a skill only has value once the agent can act on it. This story delivers the on-demand round-trip (list → open → follow). It builds on US1 and reuses tellme's existing file reader, so it ranks below US1 — but it is a first-class requirement, not a nicety, because it is what makes the skills *usable* under the on-demand model.

**Independent verification**: from the `list_skills` output, open one named skill's file with the existing `read_files` tool and confirm the skill's Markdown content is returned; separately confirm that a run which never opens a skill contains no skill content in the outbound request.

**Acceptance Scenarios**:

1. **Given** the `list_skills` output names a skill's readable location, **When** the agent reads that location with the existing file reader, **Then** that skill's Markdown content is returned.
2. **Given** a run in which no skill is opened, **When** the turn is sent to the provider, **Then** the request carries **no** skill content (nothing was injected), and the pre-flight payload estimate is unchanged by the mere presence of skills.

**Functional Requirements (FR)**:

- **FR-005**: The listing MUST include each skill's **readable location** (path) so the agent can open it with the existing file reader.
- **FR-006**: tellme MUST **NOT** inject skill content into the assembled conversation or the outbound request automatically — no skill block is added on any turn, and the pre-flight estimate is unchanged by the presence of skills.

**Non-Functional Requirements (NFR)**:

- **NFR-002**: The on-demand read path MUST reuse the existing file-reader tool; the round MUST NOT add a second file-reading tool or a new read contract.

---

### Edge cases

- **A skills directory holds nested skill directories and their `rules/` / `templates/` Markdown** → only the skill definition files are listed; rule/template Markdown is not.
- **A skill file with missing or invalid frontmatter** → it is skipped (not listed), best-effort; the other skills are unaffected.
- **Two skill files declare the same name** → tellme MUST handle it deterministically (e.g. the first is kept, the duplicate is skipped with a warning) and MUST NOT fail.
- **The skills directory is absent, empty, or unreadable** → an empty catalog, no failure (FR-001/FR-004).
- **The catalog is very large** → the `list_skills` result MUST be bounded by the ordinary tool resource contract, like every other tool (FR-008).
- **The offline paths** (`--version`, `-d`, `--tool-usage`, prompt-less `--new`, boot) → MUST be unaffected (no skill loading observable, no new output).

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain several stories or cannot be reasonably attributed to a single one.

### Global requirements

#### Functional Requirements

- **FR-007**: The round MUST **NOT** introduce the skills.sh ecosystem: no `.skills/` source, no `search_skills` / `install_skill` / `remove_skill`, and no network access for skills.
- **FR-008**: The round MUST NOT change tellme's existing native tool surface, its tool **resource contract** (per-tool bounding + timeout, round 024), the persona / request-assembly semantics, the frozen class-phrase vocabulary, or the exit-code set. `list_skills` MUST honour the same per-tool bounding and timeout discipline as every other agent tool.
- **FR-009**: Skill loading and the `list_skills` tool MUST be active only on the **prompt-bearing turn path**; the offline paths MUST remain unchanged and network-free, reusing the existing plumbing (no new failure class, no new exit code).

#### Non-Functional Requirements

- **NFR-003**: tellme MUST remain **POSIX-only** (no Windows variant), consistent with its standing scope.
- **NFR-004**: There MUST be **no security/consent gate** for skills (a settled exclusion) — consistent with the existing file tools, which have no path boundary.
- **NFR-005**: The round MUST be **stdlib-only**: no new module dependency (`go.mod` / `go.sum` unchanged).
- **NFR-006**: Loading the catalog and the `list_skills` tool MUST be exercisable **hermetically** (offline, no external network) in the ordinary test suite.

### Key entities

- **Skill**: a loaded skill definition — its **name**, **description**, and (for on-demand use) its **file location**; sourced only from `<TELL_ME_HOME>/docs/skills/`. tellme does **not** persist skills and offers no install/remove — it reads them from disk.
- **Skill catalog**: the in-memory set of loaded skills for the current run (possibly empty).
- **`list_skills` tool**: the read-only agent tool that surfaces the catalog to the model.

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: With a known skills directory, a prompt asking tellme to list its skills returns exactly that set (name + description per skill) — hermetic E2E (covers US1 / FR-001, FR-002, FR-004).
- **SC-002**: Nested skill definitions are discovered, and non-skill Markdown (a skill's `rules/*.md`, `STANDARDS.md`) is **not** listed (covers FR-003).
- **SC-003**: A missing/empty directory or a malformed individual skill file does not fail the run; the catalog degrades gracefully (covers NFR-001).
- **SC-004**: With skills present but none opened, the assembled request and the pre-flight estimate contain **no** skill content — witnessed by a falsifiability check (covers FR-006).
- **SC-005**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green; the offline paths remain network-free; no new dependency; the native tool set and the class-phrase/exit-code vocabulary are unchanged (covers FR-007…FR-009, NFR-003…NFR-006).

## Assumptions

- **Skill location**: the catalog is read from `<TELL_ME_HOME>/docs/skills/`; tellme already resolves `TELL_ME_HOME` (`internal/home`), and `tellme.sh` copies `docs/` from the group template, so the directory is present in a provisioned environment.
- **What counts as a skill**: a skill is a Markdown file that declares YAML frontmatter carrying a `name` and a `description` (the reference's rule), discovered recursively. **Resolved by `/axb-technical-research` Decision 2** — the rule is "any `.md` with valid `name`+`description` frontmatter" (not a `SKILL.md`-only restriction). *(This was an open question; settled in research — no unresolved clarification remains.)*
- **Tool conventions**: `list_skills` follows tellme's existing agent-tool conventions (registered in the agent tool set, offered on prompt-bearing turns, bounded per the round-024 resource contract, with the schema-style `reason` the other tools carry). **Resolved by `/axb-technical-research` Decision 6** — the listing carries each skill's **name + description + location**, in a deterministic (path-sorted) order. *(This was an open question; settled in research — no unresolved clarification remains.)*
- **Content reading** reuses the existing `read_files` tool — no new read tool is added (NFR-002).
- **No injection**: skill content is never placed into the prompt; this round is a **load-and-list** capability only (Q1 → 1).
- **Out of scope (recorded)**: the reference's `.skills/` (skills.sh) source, the `skillssh` management tools (`search`/`install`/`remove`), cross-source merging, and relevance-based injection.
- **No new dependency**: stdlib-only; `go.mod` / `go.sum` unchanged.
- This is a **CLI-interface** capability round: `/axb-system-analysis` is expected to record the CLI end carried to `/axb-dsl-refine`; `/axb-api-plan` is **NOOP**; `/axb-data-plan` is **NOOP** (skills are read from disk, not persisted system state — the catalog is not stored); `/axb-dsl-refine` **ADDs** the executable `list_skills` interface truth; the truth change also touches `specs/truth/techstack.md` (a new "Skills" row).
- **Recorded divergence**: tellme's skills system is deliberately a **subset** of the reference's (no skills.sh, no injection) — an operator-directed simplification, recorded rather than silent (round-028 ADR-0004 precedent).
