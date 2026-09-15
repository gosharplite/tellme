# Feature Specification: user-global interactive prompt log (round 028)

**Feature Branch**: `028-user-global-prompt-log`

**Created**: 2026-09-15

**Status**: Draft — scope **operator-directed** (session 2026-09-15): the round-015 `-i` shared prompt log moves from the environment-scoped `<TELL_ME_HOME>/output/global_prompts.jsonl` to the **per-user** `~/.tellme/global_prompts.jsonl` (a fixed location that does **not** depend on `TELL_ME_HOME`), with a one-time seed from the old location when the new file does not yet exist.

**Input**: Operator request, verbatim: *"tellme will read from and save to ~/.tellme/global_prompts.jsonl. If ~/.tellme/global_prompts.jsonl does not exist, the existing output/global_prompts.jsonl file will be copy to ~/.tellme/global_prompts.jsonl."*

**Scope note**: a **state-location** round. It relocates the **read** (seeding `-i` suggestions) and the **write** (recording an `-i` submission) of the shared prompt log to a user-global file, and adds a migration that copies the existing environment-scoped log into it the first time the user-global file is needed. It does **not** change the log's record shape, the `-i`-only write rule, the round-015/016 TUI chrome, the non-`-i` behaviours, or the round-026 tool-usage log. It **does** modify the round-015/016 truth that the log is shared, byte-for-byte, with `tell-me-go`: after this round the sharing unit changes from *one environment* to *one user* (see Assumptions → recorded divergence).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The interactive prompt log is user-global (Priority: P1)

As an operator, I want tellme's `-i` prompt history to live at `~/.tellme/global_prompts.jsonl` — a per-user file, the same way the round-026 tool-usage log lives at `~/.tellme/tools-count.jsonl` — so that my recent-prompt suggestions follow me across every tellme environment, repository and mode, instead of being scoped to a single `TELL_ME_HOME`.

**Why this priority**: it is the entire slice. Both the read path (seeding the interactive suggestions) and the write path (recording an `-i` submission) must point at the user-global file; until that happens nothing else in the round matters.

**Independent verification**: point two different runtime homes (`TELL_ME_HOME=A`, `TELL_ME_HOME=B`) at two separate home directories and drive an `-i` submission in each. Confirm (a) a prompt recorded under one home is offered as a suggestion under the other (both read the *same* user-global file), (b) the recorded line lands in `~/.tellme/global_prompts.jsonl` and **not** in `<TELL_ME_HOME>/output/global_prompts.jsonl`, and (c) a non-`-i` run writes nothing to either file.

**Acceptance Scenarios**:

1. **Given** the operator submits a prompt at the interactive (`-i`) prompt, **When** the submission is recorded, **Then** exactly one line in the shape `{"timestamp":"<RFC3339>","prompt":"<text>"}` is appended to `~/.tellme/global_prompts.jsonl`, and nothing is appended to `<TELL_ME_HOME>/output/global_prompts.jsonl`.
2. **Given** `~/.tellme/global_prompts.jsonl` already holds prompts, **When** the operator opens the interactive prompt, **Then** suggestions are seeded from that file, regardless of which `TELL_ME_HOME` is active.
3. **Given** the operator runs tellme **without** `-i` (a positional prompt, a piped prompt, `-d`, `-l`, `--version`, `--tool-usage`), **When** the run completes, **Then** `~/.tellme/global_prompts.jsonl` is neither read for suggestions nor written.

**Functional Requirements (FR)**:

- **FR-001**: The `-i` interactive prompt log MUST be **read from and written to** `~/.tellme/global_prompts.jsonl` (resolved from the operating-system user's home directory), and MUST NOT be read from or written to `<TELL_ME_HOME>/output/global_prompts.jsonl`.
- **FR-002**: The log's **record shape and read semantics** MUST be unchanged: one `{"timestamp":"<RFC3339>","prompt":"<text>"}` record per line (both fields strings), append-only; read newest-first and deduplicated; the existing bounded compaction behaviour is unchanged.
- **FR-003**: The log MUST still be **written only under `-i`** — a one-shot (positional/piped) run and every non-prompt run write nothing to it; it is read only to seed interactive suggestions.
- **FR-004**: If the directory that holds the user-global log (`~/.tellme/`) does not exist, tellme MUST create it before writing.

---

### User Story 2 - First use seeds the user-global log from the existing environment-scoped log (Priority: P2)

As an existing tellme operator, I do not want to lose the prompts I have already recorded when the log moves, so on the first interactive use after the move — while `~/.tellme/global_prompts.jsonl` does not yet exist — tellme copies the current `<TELL_ME_HOME>/output/global_prompts.jsonl` into it, leaving the original in place.

**Why this priority**: it is the continuity safeguard for the relocation in US1. Without it the move would silently discard the operator's existing prompt history; with it the upgrade is seamless. It is P2 because US1 (the relocation itself) is the deliverable and US2 only matters for operators who already have an env-scoped log.

**Independent verification**: with an existing `<TELL_ME_HOME>/output/global_prompts.jsonl` and no `~/.tellme/global_prompts.jsonl`, run `-i` and confirm the user-global file is created with **byte-identical** content while the source is left untouched; run a second `-i` and confirm no re-copy occurs (destination present); run with neither file present and confirm tellme starts with an empty log without error.

**Acceptance Scenarios**:

1. **Given** `<TELL_ME_HOME>/output/global_prompts.jsonl` exists and `~/.tellme/global_prompts.jsonl` does not, **When** the operator runs the interactive prompt, **Then** `~/.tellme/global_prompts.jsonl` is created with the same bytes as the source, and `<TELL_ME_HOME>/output/global_prompts.jsonl` is left unchanged.
2. **Given** `~/.tellme/global_prompts.jsonl` already exists, **When** the operator runs the interactive prompt, **Then** it is **not** overwritten or re-seeded.
3. **Given** neither file exists, **When** the operator runs the interactive prompt, **Then** tellme starts with an empty log and exits/records normally, without error.

**Functional Requirements (FR)**:

- **FR-005**: When `~/.tellme/global_prompts.jsonl` is **absent**, tellme MUST copy the current `<TELL_ME_HOME>/output/global_prompts.jsonl` into it **verbatim** (byte-identical), when that source exists.
- **FR-006**: The seed MUST be a **copy**, not a move — the source file is left in place — and MUST NOT overwrite an existing `~/.tellme/global_prompts.jsonl`.
- **FR-007**: When the source `<TELL_ME_HOME>/output/global_prompts.jsonl` does **not** exist, tellme MUST start with an empty user-global log and MUST NOT fail.

**Non-Functional Requirements (NFR)**:

- **NFR-001**: The seed MUST be **best-effort** and MUST NOT break the interactive prompt: a seed failure (unreadable source, unwritable destination) degrades to an empty or partial user-global log rather than an error that aborts the prompt.

---

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-008**: The round MUST NOT change any of: the log's record shape and read/compaction semantics (FR-002); the `-i`-only write rule; the round-015/016 interactive-prompt chrome and keybindings; the non-`-i` behaviours (positional/piped output, `stdout` byte-exactness, the `-d` / `-l` / `--version` / `--tool-usage` surfaces); or the round-026 tool-usage log at `~/.tellme/tools-count.jsonl`.

#### Non-Functional Requirements

- **NFR-002**: The round MUST add **no new third-party module** (Go standard library only) and MUST stay **POSIX-only** (tellme's standing scope).
- **NFR-003**: Verification MUST be **hermetic** — a temp `HOME`, a temp runtime home, injected streams and clock, and the scripted fake provider; **no pty** (the round-015/016/017 precedent).

### Edge cases

- **Multiple environments, one user-global log**: only the **first** interactive run after the move seeds the file — that is, the run in whichever environment is active while `~/.tellme/global_prompts.jsonl` is still absent. Later interactive runs in other environments do **not** re-seed (the destination already exists). Consequence: exactly **one** environment's env-scoped history is carried over (a verbatim copy — no merge/dedupe across environments), which the operator accepted.
- **`~/.tellme/` missing**: tellme creates it (FR-004).
- **User-global log deleted by the operator**: because the seed is "on absent", the next interactive run re-seeds from the active environment's source if that source still exists, else starts empty.
- **An `-i` submission on a fresh session** (`--new`): the recorded prompt still lands in the single user-global log; the log is not session-scoped, so `--new` does not clear it (unchanged from round 015).

### Key entities

- **User-global prompt log** — `~/.tellme/global_prompts.jsonl`: the shared, append-only record of interactive prompts (one `{"timestamp","prompt"}` per line), read to seed `-i` suggestions and written only on an `-i` submission. After this round this is the **only** log tellme reads and writes.
- **Environment-scoped prompt log (seed source)** — `<TELL_ME_HOME>/output/global_prompts.jsonl`: the round-015 location. After this round it is **not** read or written by tellme; it is used solely as a one-time verbatim copy source when the user-global log is absent.

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: An `-i` submission appends exactly one `{"timestamp","prompt"}` line to `~/.tellme/global_prompts.jsonl` and appends nothing to `<TELL_ME_HOME>/output/global_prompts.jsonl`; a non-`-i` run appends to neither (covers FR-001, FR-003).
- **SC-002**: With two different runtime homes, a prompt recorded under one is offered as a suggestion under the other — both read the same `~/.tellme/global_prompts.jsonl` (covers FR-001).
- **SC-003**: With an existing env-scoped log and no user-global log, the first interactive run creates the user-global log byte-identical to the source and leaves the source intact; an already-present user-global log is never overwritten; a missing source yields an empty log without error (covers FR-005, FR-006, FR-007).
- **SC-004**: `make verify`, the E2E suite and the Gherkin/DSL topology audit are green, with a falsifiability witness for the relocation (the write and read targets move to the user-global file; the seed fires only while the destination is absent; the destination is never overwritten).

## Assumptions

- The user-global base is **`~/.tellme/`**, matching the round-026 tool-usage log (`~/.tellme/tools-count.jsonl`); this round does not introduce a new base directory convention, it reuses the existing one.
- The seed is a **verbatim copy** (no dedupe or merge), triggered only while the destination is absent, sourced from the single active `<TELL_ME_HOME>/output/global_prompts.jsonl`.
- The interactive tracker (and therefore the seed) is only engaged on the **`-i`** path — a non-`-i` run never reads, writes or seeds the log.
- **Recorded divergence (operator-directed)**: after this round tellme no longer shares the prompt log with `tell-me-go`. `tell-me-go` continues to use `<TELL_ME_HOME>/output/global_prompts.jsonl`; tellme neither reads nor writes that file, and tellme also skips `tell-me-go`'s legacy locations (`.tellmego/prompts.jsonl`, `<home>/global_prompts.jsonl`). The sharing unit changes from *per-`TELL_ME_HOME` (shared with `tell-me-go`)* to *per-user (shared across every tellme environment, not with `tell-me-go`)*. This supersedes the round-015/016 "same file the other personas use" contract; it is a deliberate operator decision, recorded here so the truth-owner skills can MODIFY that contract explicitly.
- The round **modifies in place** the round-015/016 truth (`specs/truth/data/data-model.dbml` `prompt_log_entry` location, `specs/truth/techstack.md` *Shared global prompt log* row, and the `specs/truth/features/cli/chat/**` DSL/feature rows that name the path) and records the `tell-me-go` divergence; the actual truth changes are made by the truth-owner skills. No frozen plan package is touched (`fresh-package-per-round`).
