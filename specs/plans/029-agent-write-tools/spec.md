# Feature Specification: agent write tools (round 029)

**Feature Branch**: `029-agent-write-tools`

**Created**: 2026-09-16

**Status**: Draft — scope **operator-directed** (design session 2026-09-16): add the two first-class **write** tools to tellme's agent tool surface — `write_file` (create-only, atomic) and `replace_text` (strict-unique) — so the agent can create and edit files reliably without resorting to shell heredocs.

**Input**: Operator design session, verbatim: *"Let's do (b)."* → *"A"* (create-only) → *"replace_text = A (strict-unique)"*.

**Scope note**: an **agent-tool-surface** round — the first write capability beyond the existing read-only reader trio (`list_files` / `read_files` / `get_tree`) and the shell tool (`execute_command`). It adds exactly **two** tools, both with **no security/consent gate** (tellme's settled exclusion) and both portable to tellme's existing tool contract (resource params + required `reason`). It does **not** change the reader trio, `execute_command`, the round-024 tool resource contract, the CLI dispatch, or any non-tool behaviour. It deliberately **omits** the reference's other write tools (`append_text`, `undo_file_change`, `delete_path`, `create_directory`) — the shell already covers them, and they fail the "beats bash" bar. It is the **first round of the dogfooding-enablement track** (umbrella issue #60).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The agent edits a file with an exact, uniquely-identified block (Priority: P1)

As an operator driving tellme on its own repository, I want the agent to replace a specific block of a file exactly — and to *fail loudly* when the block is missing or not unique — so that edits are deterministic and never silently corrupt a file the way `sed` can.

**Why this priority**: `replace_text` is the **irreplaceable** write primitive — its strict-unique semantics have no reliable shell equivalent (bash `sed`/`awk` silently miscarry on zero or multiple matches). It is the single capability that justifies a dedicated write tool at all, so it is the core of the round.

**Independent verification**: with only `replace_text` implemented, drive an agent turn whose tool call carries a unique `old_text`; confirm exactly that block changed and the rest of the file is byte-identical. Then drive a call whose `old_text` is absent, and one whose `old_text` occurs twice; confirm both **fail** (a non-terminal tool error) and leave the file **byte-identical**.

**Acceptance Scenarios**:

1. **Given** a file whose text contains the block `old_text` exactly once, **When** the agent calls `replace_text` with that `old_text` and a `new_text`, **Then** that single occurrence is replaced by `new_text` and every other byte of the file is unchanged.
2. **Given** a file that does **not** contain `old_text`, **When** the agent calls `replace_text`, **Then** the tool fails with an error (no such block) and the file is left byte-identical.
3. **Given** a file in which `old_text` occurs **more than once**, **When** the agent calls `replace_text`, **Then** the tool fails with an error naming the match count (asking for more context) and the file is left byte-identical.

**Functional Requirements (FR)**:

- **FR-001**: `replace_text(filepath, old_text, new_text, reason)` MUST replace **the single occurrence** of `old_text` with `new_text` in the file at `filepath`.
- **FR-002**: If `old_text` does **not** occur in the file, the tool MUST fail with an error and MUST leave the file unchanged.
- **FR-003**: If `old_text` occurs **more than once**, the tool MUST fail with an error that names the occurrence count and MUST leave the file unchanged.
- **FR-004**: If `filepath` does not exist (or is unreadable / not a regular file), the tool MUST fail with an error; it MUST NOT create the file.
- **FR-005**: An **empty** `old_text` MUST be rejected with an error (it matches everywhere and cannot be uniquely identified).

---

### User Story 2 - The agent creates a new file with exact content (Priority: P2)

As an operator driving tellme on its own repository, I want the agent to create a new file whose content is exactly what the model specifies — creating parent directories as needed and never clobbering an existing file — so that file creation is deterministic and safe, unlike a shell `>` redirect that silently destroys whatever was there.

**Why this priority**: `write_file` is the **replaceable** half — a shell heredoc can already create a file, so its marginal value is lower than `replace_text`'s, and creation is a smaller share of file work than editing. It is P2 because `replace_text` can ship its value without it, but `write_file` is still needed to create the files a round produces, so it is in scope (and its *safety* — create-only + atomic — is what lets it clear the "beats bash" bar).

**Independent verification**: with only `write_file` implemented, drive an agent turn whose tool call names a path that does not exist (including a missing parent directory); confirm the file is created with byte-exact `content` and the parent directory exists. Then call `write_file` again on the same path; confirm it **fails** and the file is left byte-identical (create-only, no clobber).

**Acceptance Scenarios**:

1. **Given** a `filepath` that does not exist (whose parent directory may also be missing), **When** the agent calls `write_file` with `content`, **Then** the parent directories are created and the file is created with **exactly** `content`.
2. **Given** a `filepath` that already exists, **When** the agent calls `write_file`, **Then** the tool fails with an error and the existing file is left **byte-identical**.
3. **Given** any `write_file` call, **When** the write does not complete (e.g. it is interrupted or fails mid-write), **Then** the destination path is either **absent** or holds the **complete** `content` — never a partial/torn file.

**Functional Requirements (FR)**:

- **FR-006**: `write_file(filepath, content, reason)` MUST create the file at `filepath` with **exactly** `content` (byte-for-byte).
- **FR-007**: If `filepath` already **exists**, the tool MUST fail with an error and MUST NOT modify or truncate the existing file (**create-only**; there is no overwrite, and no `overwrite` parameter).
- **FR-008**: The tool MUST create any missing **parent directories** (`MkdirAll`-equivalent) before creating the file.
- **FR-009**: The write MUST be **atomic**: at no observable point does the destination hold a partial file — the implementation writes to a temporary file in the target's directory and `rename`s it into place (a same-filesystem rename is atomic on POSIX), cleaning up the temp file on failure.
- **FR-010**: An **empty** `content` MUST be accepted (it creates an empty file) — the create-only rule (FR-007) still applies.

---

## Requirements *(mandatory)*

> The per-story FR are attached under each story above; this section holds only requirements that constrain both stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-011**: Both tools MUST be registered in the agent tool registry and offered to the model in a **deterministic order** (so the round-026 `--tool-usage` report enumerates them deterministically). Neither tool requires consent (tellme has no security layer — a settled exclusion).
- **FR-012**: Both tools MUST require `reason` and the loop MUST echo it into the tool-loop log line (the round-021/022 convention), and both MUST declare the round-024 **resource-contract** params (`max_output_tokens`, `timeout`) with a per-tool **default timeout of 30 s** (matching the readers).
- **FR-013**: Both tools MUST bound their result to the resolved byte budget and MUST return a **nil-error timeout result** when they observe their effective deadline (the round-024 FR-018 convention), never an `error:`-class result.
- **FR-014**: The round MUST NOT add `append_text`, `undo_file_change`, `delete_path`, or `create_directory`, and MUST NOT add any security/consent gate, path allowlist, or backup/undo store.

#### Non-Functional Requirements

- **NFR-001**: The round MUST add **no new third-party module** (Go standard library only) and MUST stay **POSIX-only** (tellme's standing scope).
- **NFR-002**: Verification MUST be **hermetic** — a temp working directory, injected streams, the scripted fake provider, and no pty (the round-015/016/017 precedent); a tool failure MUST surface as a **non-terminal** tool result (the model may recover), not a process failure.

### Key entities

- **None.** This round adds **behaviour** (two agent tools), not persisted data; it introduces no new file format, table, or record. (A created/edited file's content is model-supplied and not modelled as an entity.)

## Edge cases

- **`old_text` spans a line boundary / contains newlines**: MUST be matched byte-for-byte (the block may be multi-line) — no trimming or normalisation.
- **`new_text` equals `old_text`**: a no-op replace is permitted (the tool still rewrites the file; the content is unchanged).
- **`write_file` where a parent path component is a regular file**: MUST fail with an error (the directory cannot be created); no partial file is left (FR-009).
- **`write_file`/`replace_text` observing their effective timeout**: returns the nil-error timeout result (FR-013); no partial/torn file (FR-009 for `write_file`).
- **Concurrent/duplicate calls**: tool execution stays **sequential** (a settled exclusion — tool-call concurrency is declined, issue #47), so no write-write race is introduced.
- **A path outside any "workspace"**: allowed (no path/safety boundary — a settled exclusion).

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: An agent `write_file` call with new content creates the file (and parents) with byte-exact content; a **second** `write_file` on the same path fails and leaves the file byte-identical — i.e. create-only holds (covers FR-006, FR-007, FR-008).
- **SC-002**: After a `write_file` failure/interruption, the destination is either absent or complete — the atomicity witness shows no partial file (covers FR-009).
- **SC-003**: An agent `replace_text` call edits exactly one uniquely-matched block; with **zero** or **multiple** matches it fails and leaves the file byte-identical (covers FR-001, FR-002, FR-003).
- **SC-004**: Both tools appear in the `--tool-usage` report's registry order, both require `reason`, and both honour the resource contract (covers FR-011, FR-012, FR-013).
- **SC-005**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with **falsifiability witnesses**: reverting create-only → the no-clobber Example fails; removing the temp+rename → the atomicity witness fails; loosening `replace_text` to replace-first → the ambiguity Example fails.

## Assumptions

- The agent tool loop (round 008), the tool resource contract + FR-018 timeout result (round 024), the required-`reason` echo (round 021/022), and the registry/CLI seam already exist; this round only adds two tool adapters behind the existing `domain/tools.Tool` port and registers them.
- The reference (`tell-me-go`) implements `write_file`/`replace_text` behind a security gate + backup snapshots; tellme **drops both** (no security layer — settled; no undo — `undo_file_change` is deliberately out of scope), so the tools reduce to pure filesystem operations.
- **Create-only is a tellme improvement over the reference** (which clobbers): with no backup/undo net, a silent overwrite would be irreversible, so `write_file` refuses to overwrite. Editing an existing file is `replace_text`'s job.
- **Atomic write is a tellme improvement over the reference** (which writes in place, non-atomically): temp-file + `rename` guarantees no partial file.
- Both tools are a first step of the **dogfooding-enablement track** (issue #60); whether `write_file` earns its place is to be **measured** via the round-026 `--tool-usage` report over real usage and revisited later (its value is admittedly lower than `replace_text`'s). This is a recorded intent, not part of this round's scope.
- The actual truth changes (`specs/truth/techstack.md` agent-tool-surface row; a new `features/cli/**` interface feature + `dsl.md` rows; `contracts/**` and `data/**` expected NOOP) are made by the truth-owner skills; this plan package records the intent only (`fresh-package-per-round`).
