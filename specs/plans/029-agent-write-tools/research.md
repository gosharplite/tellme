# Phase 0 Research: agent write tools — `write_file` & `replace_text` (round 029)

**Topic**: how tellme adds its **first write capability** to the agent tool surface — two tools, `write_file` (create-only, atomic) and `replace_text` (strict-unique) — behind the existing `domain/tools.Tool` port, with **no security layer** and **no undo**. Each decision supports `spec.md` (US1/US2 · FR-001..FR-014) and the operator-locked decisions from the design session (scope = the pair; `write_file` = create-only **A**; `write_file` = atomic; `replace_text` = strict-unique **A**; no security/undo; 30 s timeout).

## Decision 1: The write surface is exactly two tools, behind the existing Tool port

- **Decision**: add **two** tool adapters (`write_file`, `replace_text`) in a new `internal/infrastructure/tools/writer.go`, implementing the existing `domain/tools.Tool` port (`Name` / `Description` / `Parameters` / `Contract` / `Execute`). Register them in the CLI seam (`newToolRegistry` in `internal/cli`) alongside the reader trio and `execute_command`. No new port, no new abstraction.
- **Rationale**: the smallest surface that lets the agent create and edit files with exact, deterministic content (the dogfooding unblock, issue #60). The port already carries the resource contract (round 024) and the required-`reason` echo (round 021/022), so the write tools inherit both with no framework change.
- **Alternatives considered**:
  - A single "edit" tool with a mode flag (`create` / `replace`) — rejected: two intents in one schema is harder for the model to use correctly and muddies the contract.
  - Port the reference's full writer family (`append_text` / `undo_file_change` / `delete_path` / `create_directory`) — rejected: the shell already covers them (`>>`, `git checkout`, `rm`, `mkdir -p`); none clears the "beats bash" bar (README → *Design Intent*).

## Decision 2: `write_file` is create-only (it never overwrites)

- **Decision**: `write_file(filepath, content, reason)` creates `filepath` with exactly `content`; if `filepath` **already exists** it fails with an error and leaves the file byte-identical. There is **no** overwrite and **no** `overwrite` parameter. Editing an existing file is `replace_text`'s job.
- **Rationale**: tells the model that `write_file` means "create", not "rewrite", and — because tellme dropped the reference's backup/undo net (D5) — a silent clobber would be **irreversible**. Create-only is deterministic and fail-loud, which is exactly the property that lets a dedicated tool beat `cat > file` (which silently destroys the target). It also pushes all mutation of existing files through `replace_text`'s unique-match guarantee.
- **Alternatives considered**:
  - **Overwrite silently (reference parity)** — rejected: the reference's safety net was its `backupManager`/`undo_file_change`, which tellme does not have; without it a one-line mistake destroys a file with no recovery.
  - **Overwrite behind an explicit `overwrite: true` flag** — considered (operator option C); rejected in favour of the simplest safe rule. Regeneration is rare and `rm` + create is acceptable (the shell is available).
  - **Create-or-append-by-mode** — rejected: append is a separate intent (and out of scope, D1).

## Decision 3: `write_file` writes atomically (temp file + rename)

- **Decision**: `write_file` writes the content to a temporary file **in the target directory** and `rename`s it into place (a same-filesystem rename is atomic on POSIX); the temp file is removed on failure. A create-only presence check happens before the write.
- **Rationale**: guarantees the destination is never observed in a partial/torn state (FR-009) — a reliability property the reference's in-place `WriteFile` lacks and that the shell `>` cannot offer. Temp-in-the-same-directory keeps the rename on one filesystem (cross-device rename is not atomic). `os.CreateTemp` + `os.Rename` + `os.Remove` are stdlib.
- **Alternatives considered**:
  - **In-place `os.WriteFile` (reference parity)** — rejected: a crash/interruption mid-write leaves a truncated file.
  - **Temp file + `fsync` before rename** — considered; `fsync` adds durability (survives power loss) but not observability, at a cost disproportionate for source-file editing. Recorded as a possible future hardening, not adopted.
  - **Temp file elsewhere (e.g. `$TMPDIR`)** — rejected: a cross-filesystem rename is a copy and not atomic.

## Decision 4: `replace_text` is strict-unique

- **Decision**: `replace_text(filepath, old_text, new_text, reason)` replaces the occurrence of `old_text` only when it occurs **exactly once**; `0` occurrences → error (`old_text` not found); `>1` occurrences → error naming the count (asking for more context). It never creates the file, and an empty `old_text` is rejected.
- **Rationale**: this **is** the reason the tool exists — a deterministic, fail-loud exact-block edit with no reliable shell equivalent (`sed`/`awk` silently do the wrong thing on zero or multiple matches, and are quoting-fragile). It is the round's P1 story.
- **Alternatives considered**:
  - **Replace-first** (`>=1` → replace the first) — rejected: silently picks a target the model did not uniquely identify; loses the guarantee.
  - **Replace-all** — rejected: a broad edit on a non-unique block is exactly the foot-gun the tool exists to prevent.
  - **A line-number / range edit** — rejected: line numbers drift and are not robust to concurrent edits; block content is the stable identity.

## Decision 5: No security gate and no backup/undo

- **Decision**: neither tool calls a security manager, a path allowlist, or a consent prompt, and neither snapshots the file for undo. They are pure filesystem operations.
- **Rationale**: tellme's **settled exclusion** is "no security layer" (README → *Design Intent*; the operator always bypassed the reference's `SafePath`/consent rules, which caused repeated tool-call failures for no protection). Porting `backupManager`/`undo_file_change` would add an in-memory snapshot store for a `git checkout` alternative — out of scope by D1.
- **Alternatives considered**:
  - **Port the reference's `SecurityManager` gate** — rejected: a settled exclusion.
  - **Port `backupManager` + `undo_file_change`** — rejected: scope (D1); `git`/`execute_command` already revert.

## Decision 6: Both tools slot into the round-024 tool resource contract

- **Decision**: both tools declare the optional `max_output_tokens` + `timeout` params (via the shared reader schema helper), publish a `ToolContract{DefaultTimeout: 30s}` (matching the readers), bound their (tiny) result to the resolved byte budget, and return a **nil-error timeout result** when they observe their effective deadline (FR-018). The loop stays the single enforcement point; no contract change.
- **Rationale**: uniformity — every agent tool already honours the contract; the write tools must not be a gap. The result is a short confirmation (or an inline error), so bounding is trivial and the 30 s default is generous for a filesystem write.
- **Alternatives considered**:
  - **Exempt the write tools from the contract** — rejected: two contracts to reason about; the round-024 invariant is "every tool bounds at the source".
  - **A longer/absent timeout** — rejected: a write is fast; a hung write is a bug, and 30 s bounds it.

## Decision 7: Result and failure shape

- **Decision**: success returns a short confirmation string (e.g. `File written successfully.` / `File updated successfully.` — wording finalised at `/axb-dsl-refine`); a **contract failure** (missing/ambiguous block, existing file on create, unreadable path, empty `old_text`) returns a Go **error**, which the loop surfaces as a **non-terminal** `error: …` tool result the model may recover from. This is **not** the frozen request-level class phrase (`the tool request failed`, exit 7), which is reserved for a request for an **unregistered** tool.
- **Rationale**: matches the existing tool-error semantics (round 008): a recoverable tool problem is fed back to the model, it does not abort the run. A bad `old_text` is exactly the kind of thing the model should retry with more context.
- **Alternatives considered**:
  - **Return failures as successful result text** — rejected: the loop's `err` channel is the existing, tested path for a recoverable tool failure; result-text sniffing is what round 026 deliberately avoided.
  - **Treat a bad edit as a terminal run failure** — rejected: breaks the agent loop on an ordinary, recoverable condition.

## Decision 8: No new dependency; POSIX-only; hermetic verification

- **Decision**: stdlib only (`os`, `path/filepath`, `strings`, `encoding/json`); POSIX-only; no new module. Verification is hermetic: **unit tests** pin the tool logic (create-only, atomicity, strict-unique, empty-`old_text`, missing-file, parent creation) against a temp dir, and **E2E** drives the built binary with the fake provider scripting the write tool calls (`creates a file before answering` / `edits a file before answering`), asserting the file bytes and the refusals; no pty. **Falsifiability witnesses**: revert create-only → the no-clobber Example fails; drop temp+rename → the atomicity witness fails; loosen `replace_text` to replace-first → the ambiguity Example fails.
- **Rationale**: the standing scope (stdlib, POSIX, offline) and the round-008/021/024 harness precedent; witnesses keep the guarantees falsifiable.
- **Alternatives considered**:
  - **A third-party atomic-write library** — rejected: a dependency for `CreateTemp`+`Rename`.
  - **A pty harness** — rejected: `TELL_ME_FORCE_STDIN_TTY`/scripted tool calls already cover the paths; the project forswore a pty.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: add a **Write filesystem tools** row to the agent tool surface (`write_file` create-only + atomic; `replace_text` strict-unique; no security/undo) and move `write_file`/`replace_text` **out of** the *Not Introduced Yet* write-tools bullet (leaving `append_text`/`undo_file_change` deferred; `delete_path`/`create_directory` remain non-goals).
- `specs/truth/features/cli/**` → **ADD**: an interface feature + `dsl.md` rows for the create-only / atomic / strict-unique steps (owned by `/axb-dsl-refine`).
- `specs/truth/contracts/**` and `specs/truth/data/**` → **NOOP** (no API surface; the write tools persist no state).
