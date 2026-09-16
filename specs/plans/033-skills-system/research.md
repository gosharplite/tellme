# Phase 0 Research: skills system — load & list pre-loaded skills (round 033)

**Topic**: how tellme gains a **minimal skills system** — it **loads** the skill definitions under `<TELL_ME_HOME>/docs/skills/` and surfaces them to the agent **on demand** through a read-only **`list_skills`** tool. Skill **content is not auto-injected**; the agent reads a skill's file with the existing `read_files` tool. The reference's **skills.sh** ecosystem (`.skills/`, `search/install/remove_skill`) and its **relevance-based injection** are explicitly out of scope. Each decision supports `spec.md` (US1/US2 · FR-001…FR-009 · NFR-001…NFR-006) and the operator-locked clarifications (**Q1 → on-demand only**; **Q2 → moot**; **Q3 → a `list_skills` tool**).

**Must-ask questions (settled).** Per the AIxBDD three must-asks, all three are already written in the existing `specs/truth/techstack.md` and are **unchanged** this round: the system has a **single CLI end** (no web frontend, no HTTP server — skills are read from local files, exposing no new end); the BDD techstack is **`godog`** driving the interface Gherkin E2E against the built binary; the test strategy is **E2E** for the acceptance path plus fast unit tests for pure helpers. No re-ask is warranted (must-ask Rule 2). This round adds **no** new end and does **not** change the test strategy.

## Decision 1: The skills source is `<TELL_ME_HOME>/docs/skills/` only (single source)

- **Decision**: the skills catalog is loaded from **one** directory — `<TELL_ME_HOME>/docs/skills/` — resolved through the existing runtime-home resolver (`internal/home`) and a CLI-injected user/home seam (mirroring the round-026/028 pattern). There is **no** `.skills/` source and no cross-source merging.
- **Rationale**: the operator scoped the round to "load skills in `ait-*/docs/skills`" and to drop the `skillssh` complication. A single source keeps the loader trivial and the acceptance a pure function of the environment's `docs/skills` (which `tellme.sh` copies from the group template). tellme already resolves `TELL_ME_HOME`; the skills dir is just one more consumer.
- **Alternatives considered**:
  - **Add the reference's second source (`.skills/`)** — rejected (Q1): that *is* the skills.sh complication the operator asked to drop; it also needs the `skillssh` install/remove tools to be useful.
  - **A user-global `~/.tellme/skills/`** — rejected: the operator explicitly wants the environment-scoped `ait-*/docs/skills`; the user-global root is reserved for the round-026/028 logs.

## Decision 2: A skill is a Markdown file with valid `name`/`description` frontmatter; discovery is recursive

- **Decision**: the loader **walks** `<home>/docs/skills/` **recursively** and treats a file as a skill iff its content **starts with a `---` YAML frontmatter block that declares both `name` and `description`** (the reference's `parseSkill` rule: CRLF-normalized; a leading `---\n` and a closing `---\n`; `name:` and `description:` parsed from the block). A file without a valid frontmatter block is **not** a skill and is skipped silently (e.g. a skill's `rules/*.md`, `STANDARDS.md`, `examples.md`). A skill's `TokenCount` heuristic (`len(content)/4`) is computed but unused this round. Duplicate names: the **first discovered wins**, later ones are skipped with a warning. A missing/empty/unreadable directory → an **empty catalog**, no error. A malformed individual file → skip best-effort (never fail the run).
- **Rationale**: this precisely matches the reference's behavior (`fileSkillRepository.parseSkill`), so the same `docs/skills` layout (a directory per skill containing `SKILL.md` plus `rules/`/`templates/`) yields exactly the `SKILL.md` set without special-casing the filename. It is robust to a skill authored under a differently-named file (any frontmatter-bearing Markdown), and it avoids the "restrict to `SKILL.md`" trap where a nested `rules/x.md` could be misread if it ever carried frontmatter. Best-effort degradation mirrors the reference (`fileSkillRepository.reload` skips a duplicate with a warning) and tellme's standing "never break a turn on a read" discipline.
- **Alternatives considered**:
  - **Strictly `SKILL.md` files only** — considered (the round-033 spec's open `NEEDS CLARIFICATION`); rejected: less faithful to the reference and breaks if a skill is authored in a file not named `SKILL.md`. The frontmatter rule subsumes it for the shipped layout (only `SKILL.md` carries valid frontmatter).
  - **Any `.md` file (no frontmatter requirement)** — rejected: would list a skill's `rules/*.md` and `STANDARDS.md` as skills (they are not), producing a wrong catalog.

## Decision 3: On-demand surface only — a read-only `list_skills` tool; no injection

- **Decision**: tellme surfaces the catalog **only** through a **read-only** agent tool, `list_skills`. It does **not** inject any skill content into the prompt (no system-message block, no relevance selector, no per-turn budget). The agent uses a skill by opening its listed file with the existing `read_files`.
- **Rationale**: this is the operator's locked choice (Q1 → 1, Q3 → 1). It is also the smallest correct design: it needs no selection heuristic, adds **zero** per-turn prompt cost, keeps the pre-flight estimate unchanged, and leverages machinery tellme already ships (`read_files`, the tool port). "List all pre-loaded skills" then becomes a deterministic model call, and "use a skill" is an ordinary read the model chooses to make.
- **Alternatives considered**:
  - **Auto-inject relevant skills (the reference's `SkillSelector` + `ContextTransformer`)** — rejected (Q1): adds a ranking heuristic and per-turn prompt bloat, and tellme has no context-transformer pipeline to hang it on.
  - **Inject the full catalog manifest every turn** — rejected: spends budget every turn for the axb-* set's descriptions and still does not give the content.

## Decision 4: The catalog is loaded on the prompt-bearing turn path; the offline paths are untouched

- **Decision**: tellme loads the catalog when it assembles a **prompt-bearing turn** (positional/piped prompt, the round-012 reader, the `-i` submit) — the same surface where the agent tools are offered — and **not** on the offline paths (`--version`, `-d`, `-l`, `--tool-usage`, prompt-less `--new`, boot). The load is a local filesystem read; it makes **no** network contact.
- **Rationale**: the `list_skills` tool is only meaningful where tools are offered, so the load belongs on that path — consistent with how MCP discovery is gated (round 032 FR-016) and with the tool-usage report being a separate offline surface. Keeping the offline paths byte-identical preserves tellme's long-standing offline guarantee.
- **Alternatives considered**:
  - **Load once at process boot** — rejected: the boot path is shared with the offline flags; loading there would touch `docs/skills` on `--version`/`-d` for no benefit, and `--version` has no `TELL_ME_HOME`.
  - **Cache the catalog for the process lifetime** — rejected: unnecessary; loading is a cheap directory walk and a fresh read per run is the deterministic, testable choice (a mid-run edit is picked up next run).

## Decision 5: A minimal domain type + a file loader; no repository framework

- **Decision**: add a small domain value type (`Skill{Name, Description, Location}`) in a new `internal/domain/skills`, plus a tiny loader in `internal/infrastructure/skills` that walks the directory and returns `[]Skill` (and any warnings). There is **no** `SkillRepository` interface, **no** `CompositeRepository`, **no** `Refresh`, and **no** `SkillSelector`. The `list_skills` tool is constructed with the loaded catalog (or a `func() ([]Skill, error)` seam) so it stays hermetic and testable.
- **Rationale**: tellme's minimalism ("a dedicated tool only earns its place if it beats bash on a real axis"); the full reference framework exists to support two sources + on-demand refresh + relevance selection, none of which this round has. A value type + a loader + a tool is the smallest shape that satisfies the spec without a speculative abstraction.
- **Alternatives considered**:
  - **Port the reference's `domain/skills` + `infrastructure/skills` framework (ports, composite, selector)** — rejected (Q1): the extra interfaces have exactly one implementation and no second source to compose, so they add indirection with no payoff.
  - **Inline the load inside the tool's `Execute`** — rejected: couples the tool to the filesystem/home and makes the tool harder to unit-test; a load-at-assembly seam keeps it hermetic (the round-026 `ToolUsageSink` precedent).

## Decision 6: `list_skills` output = per skill `name` + `description` + `location`, in a deterministic order

- **Decision**: the tool returns one entry per loaded skill — its **name**, its **description**, and its **readable location** (the on-disk path, so the agent can open it with `read_files`) — in a **deterministic order** (sorted by location path, i.e. by file path). An empty catalog returns a clear "no skills are loaded" result (not an error).
- **Rationale**: name + description is what "list the pre-loaded skills" must answer (FR-002); the location is what makes the on-demand flow work (FR-005) — it is the join to `read_files`. A path sort is deterministic and stable across runs (tellme's determinism value), independent of directory walk order. Returning empty-as-a-result (not an error) lets the agent answer truthfully (FR-004).
- **Alternatives considered**:
  - **Name + description only (no path)** — rejected: the agent would have to guess the file location, breaking the on-demand read; and the reference's `list_skills` is an enumeration aid, whereas tellme's must be the entry point to reading.
  - **Group by source (reference shape)** — inapplicable: there is only one source this round.
  - **Discovery/walk order** — rejected: filesystem order is not guaranteed; a path sort is deterministic and testable.

## Decision 7: `list_skills` is an ordinary agent tool (resource contract + `reason`), registered in `agentTools()`

- **Decision**: `list_skills` is a `domain/tools.Tool` adapter in `internal/infrastructure/tools/` (e.g. `skills.go`), built with the shared `resourceSchema` helper so it carries the round-024 resource params (`max_output_tokens`, `timeout`) and the mandatory `reason` property (round 031 `required ⊆ properties`). It declares a `ToolContract` default timeout in the **local-reader class (30 s)** (a local directory read, like the reader trio), bounds its own output to the resolved byte budget at the source, and returns a **nil-error timeout result** when it observes its deadline (FR-018). It is registered in the production assembler `agentTools()` alongside the reader trio, the write pair, and `execute_command`, so the round-031 well-formedness gate covers it automatically.
- **Rationale**: uniformity — every agent tool already honours the resource contract and the `reason` schema; a skills tool must not be a gap. Registering it in `agentTools()` means the existing hermetic gate (which iterates the production assembler — round-031 ARCH-1) validates its schema for free, so no new gate is needed.
- **Alternatives considered**:
  - **A parameterless tool (reference parity) that bypasses the shared builder** — rejected: it would be the lone tool outside the `reason`/resource contract and outside the well-formedness gate; the reference has no such contract, tellme does.
  - **A separate, non-tool surface (e.g. only an offline flag)** — rejected (Q3): the operator chose the model-callable tool.
  - **The networked-tool timeout class (300 s)** — rejected: `list_skills` does a local directory read, so the 30 s reader class is the honest fit.

## Decision 8: No skills.sh ecosystem and no injection — a recorded divergence from the reference

- **Decision**: the round adds **no** `.skills/` source, **no** `search_skills`/`install_skill`/`remove_skill`, and **no** automatic skill injection. This is a deliberate, recorded divergence from `tell-me-go` (which ships all of these), not an omission.
- **Rationale**: the operator named the `skillssh` toolkit as the "complication" to avoid. Under the on-demand model, an install/remove surface is unnecessary: the `docs/skills` directory is managed by the environment (the Niffler `tellme.sh` group template / git), exactly as the reference's *local* skills are. Recording the divergence (round-028 ADR-0004 precedent) keeps the truth honest and prevents a future reader from assuming parity.
- **Alternatives considered**:
  - **Keep `list_skills` + `search_skills` but drop install/remove** — rejected: `search_skills` hits the GitHub network and belongs to the skills.sh ecosystem; the operator asked to drop it.
  - **Silently omit** — rejected: a recorded divergence is required (the reference ships these; the difference must be explicit).

## Decision 9: Hermetic verification; stdlib-only; POSIX-only

- **Decision**: verification is hermetic. **Unit tests** cover the loader/parser (frontmatter parsing, recursive discovery, non-skill `.md` skipped, duplicate-name first-wins, missing/empty dir → empty, CRLF handling) and the tool's output format/order/empty case/`reason` echo. **E2E** drives the built binary through the existing godog harness with a known `docs/skills` set placed under the per-scenario `TELL_ME_HOME`, and asserts the `list_skills` result (the fake provider scripts a `list_skills` call and its result is asserted). **Falsifiability witnesses**: (a) drop the frontmatter rule → the non-skill-`.md` scenario fails; (b) inject skill content into the request → the no-injection assertion fails. stdlib-only (`os`, `path/filepath`, `strings`); `go.mod`/`go.sum` unchanged; POSIX-only.
- **Rationale**: matches tellme's standing test strategy (godog E2E + fast unit tests, ADR-036 determinism) and its scope (stdlib, POSIX). The no-injection property (FR-006) is a negative that needs an explicit witness (assert the request carries no skill content), mirroring round-010's ordering witnesses.
- **Alternatives considered**:
  - **Unit-only** — rejected: the capability is end-to-end observable (a prompt → a listed catalog), so E2E is the honest acceptance layer.
  - **A live skills.sh network check** — rejected: out of scope (no network); the loader is local-only.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: add a **Skills** category to the CLI Application rows — the **skills catalog (load)** (single source `<TELL_ME_HOME>/docs/skills/`, recursive frontmatter parse, best-effort, no injection) and the **`list_skills` agent tool** (read-only; name+description+location; honours the resource contract) — and move the reference's **skills.sh** ecosystem + **automatic skill injection** into *Not Introduced Yet* as explicit non-goals/divergences; update the **Agent tool schemas** row (the shared builder now also backs `list_skills`), the **Agent tool-schema gate** row (six → seven tools), and the **Tool-usage accounting** row (the report enumerates the live registry, now seven tools).
- `specs/truth/features/cli/**` → **ADD** expected — the executable `list_skills` interface truth (the catalog is listed on demand; no content is injected), authored by `/axb-dsl-refine`.
- `specs/truth/contracts/**` → **NOOP** (single CLI end; no OpenAPI surface).
- `specs/truth/data/**` → **NOOP** (skills are read from disk each run — the catalog is **not** persisted system state; no install/remove means nothing to store).

## Residual risks / forward items (not blocking)

- **Loader fidelity to the reference is by rule, not by code**: tellme re-implements `parseSkill` (stdlib), so a future change to the reference's rule is not automatically inherited — acceptable, since the rule is small and unit-pinned.
- **A very large catalog** is bounded by the round-024 resource contract on the tool's **result** (it does not page); listing an enormous skills tree is not a use case here (the `axb-*`/`golang-*` set is small).
- **`list_skills` result ordering** is fixed to a path sort; a future need to group/sort differently is a small, local change.
- **No `.skills/` / injection** — recorded as out of scope (Q1/Q3); a future round could add either, but neither is planned.
- **Tool count in the truth** — the techstack tool rows are updated to the current set; the round-026 report's live-registry enumeration (now seven tools) is corrected in the same edit.
