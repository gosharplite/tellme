# Phase 0 Research: agent write tools — `write_file` & `replace_text` (round 029)

**Topic**: how tellme adds its **first write capability** to the agent tool surface — two tools, `write_file` (create-only, atomic) and `replace_text` (strict-unique, atomic) — behind the existing `domain/tools.Tool` port, with **no security layer** and **no undo**. Each decision supports `spec.md` (US1/US2 · FR-001..FR-014) and the operator-locked decisions from the design session (scope = the pair; `write_file` = create-only **A**; every write = atomic; `replace_text` = strict-unique **A**; no security/undo; 30 s timeout).

## Decision 1: The write surface is exactly two tools, behind the existing Tool port

- **Decision**: add **two** tool adapters (`write_file`, `replace_text`) in a new `internal/infrastructure/tools/writer.go`, implementing the existing `domain/tools.Tool` port (`Name` / `Description` / `Parameters` / `Contract` / `Execute`). Register them in the CLI seam (`newToolRegistry` in `internal/cli`) alongside the reader trio and `execute_command`. No new port, no new abstraction.
- **Rationale**: the smallest surface that lets the agent create and edit files with exact, deterministic content (the dogfooding unblock, issue #60). The port already carries the resource contract (round 024) and the required-`reason` echo (round 021/022), so the write tools inherit both with no framework change.
- **Alternatives considered**:
  - A single "edit" tool with a mode flag (`create` / `replace`) — rejected: two intents in one schema is harder for the model to use correctly and muddies the contract.
  - Port the reference's full writer family (`append_text` / `undo_file_change` / `delete_path` / `create_directory`) — rejected: the shell already covers them (`>>`, `git checkout`, `rm`, `mkdir -p`); none clears the "beats bash" bar (README → *Design Intent*).

## Decision 2: `write_file` is create-only (it never overwrites)

- **Decision**: `write_file(filepath, content, reason)` creates `filepath` with exactly `content`; if `filepath` **already exists** it fails with an error and leaves the file byte-identical. There is **no** overwrite and **no** `overwrite` parameter. Editing an existing file is `replace_text`'s job. The create-only guard is **atomic** (never a `Stat`-then-`Rename` TOCTOU window — see Decision 3).
- **Rationale**: tells the model that `write_file` means "create", not "rewrite", and — because tellme dropped the reference's backup/undo net (D5) — a silent clobber would be **irreversible**. Create-only is deterministic and fail-loud, which is exactly the property that lets a dedicated tool beat `cat > file` (which silently destroys the target). It also pushes all mutation of existing files through `replace_text`'s unique-match guarantee.
- **Alternatives considered**:
  - **Overwrite silently (reference parity)** — rejected: the reference's safety net was its `backupManager`/`undo_file_change`, which tellme does not have; without it a one-line mistake destroys a file with no recovery.
  - **Overwrite behind an explicit `overwrite: true` flag** — considered (operator option C); rejected in favour of the simplest safe rule. Regeneration is rare and `rm` + create is acceptable (the shell is available).
  - **Create-or-append-by-mode** — rejected: append is a separate intent (and out of scope, D1).

## Decision 3: Every write is atomic (temp file + atomic move), and create-only is itself atomic

- **Decision**: **one** atomic-write helper serves **both** tools. It writes the full new content to a temporary file **in the target directory**, sets the temp file's mode to **`0644`**, and moves it into place:
  - **`replace_text`**: the destination already exists, so the move is a `rename` (which atomically replaces the destination) — read → replace in memory → temp → `rename`.
  - **`write_file`**: the move is **atomic create-only** — `os.Link(tmp, dest)` (an `os.Link` fails with `EEXIST` if the destination exists, giving create-only **with no stat/TOCTOU window**) followed by removing the temp; `O_CREAT|O_EXCL|O_WRONLY` on the destination directly is the alternative. No `Stat`-then-`Rename` guard.
  - The temp file is removed on any failure.
- **Rationale**: the destination is never observed partial (**FR-009, both tools**) **and** an existing file is never silently overwritten (**FR-007**) — one atomic step each, no TOCTOU. This removed the round's worst-ordered risk: the *edit* tool (no undo) would otherwise have been the only torn-write path while the *create* tool was crash-safe. Setting mode to `0644` on the temp file avoids the `0600` that `os.CreateTemp` would otherwise leave on a renamed temp (which would make every created file owner-only, unlike a shell `>` / `os.WriteFile(…, 0644)`). Missing parent directories are created with mode **`0755`**; an explicit empty `content` (`""`) creates an empty file while a **missing** `content` key is rejected — see *Recorded refinements* below.
- **Alternatives considered**:
  - **In-place `os.WriteFile` (reference parity)** — rejected: a crash/interruption mid-write leaves a truncated file (unacceptable for the no-undo P1 edit tool).
  - **`Stat`-then-`rename` for create-only** — rejected: POSIX `rename(2)` clobbers an existing destination, so create-only would rest on a non-atomic check (TOCTOU).
  - **`os.CreateTemp` with its default `0600`** — rejected: owner-only files.
  - **Temp file + `fsync` before rename** — considered; `fsync` adds durability (survives power loss) but not observability, at a cost disproportionate for source-file editing. Recorded as a possible future hardening, not adopted.
  - **Temp file elsewhere (e.g. `$TMPDIR`)** — rejected: a cross-filesystem rename is a copy and not atomic.

## Decision 4: `replace_text` is strict-unique

- **Decision**: `replace_text(filepath, old_text, new_text, reason)` replaces the occurrence of `old_text` only when it occurs **exactly once**; `0` occurrences → error (`old_text` not found); `>1` occurrences → error naming the count (asking for more context). It never creates the file, and an empty `old_text` is rejected. A **no-op** (`new_text` equals `old_text`) short-circuits to the success confirmation **without** a write.
- **Rationale**: this **is** the reason the tool exists — a deterministic, fail-loud exact-block edit with no reliable shell equivalent (`sed`/`awk` silently do the wrong thing on zero or multiple matches, and are quoting-fragile). It is the round's P1 story.
- **Alternatives considered**:
  - **Replace-first** (`>=1` → replace the first) — rejected: silently picks a target the model did not uniquely identify; loses the guarantee.
  - **Replace-all** — rejected: a broad edit on a non-unique block is exactly the foot-gun the tool exists to prevent.
  - **A line-number / range edit** — rejected: line numbers drift and are not robust to concurrent edits; block content is the stable identity.
- **Recorded hazard (review finding 6)**: `replace_text` reads the **whole** file to locate the unique block (`os.ReadFile`). The round-024 invariant "every tool bounds at the source" governs the tool's **result**, not its **input**; an unbounded input read is a **conscious, recorded** decision this round (the block must be found in the whole file). A future **input** bound (or a streaming/anchored matcher) is a forward item, noted in `techstack.md`. **Symlink case (implementation review finding 6):** the read follows a symlink, and the atomic `rename` then replaces the *link* with a regular file — a recorded limitation of `replace_text`.

## Decision 5: No security gate and no backup/undo

- **Decision**: neither tool calls a security manager, a path allowlist, or a consent prompt, and neither snapshots the file for undo. They are pure filesystem operations.
- **Rationale**: tellme's **settled exclusion** is "no security layer" (README → *Design Intent*; the operator always bypassed the reference's `SafePath`/consent rules, which caused repeated tool-call failures for no protection). Porting `backupManager`/`undo_file_change` would add an in-memory snapshot store for a `git checkout` alternative — out of scope by D1. Because there is no undo, **both** tools write atomically (D3).
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

## Decision 8: No new dependency; POSIX-only; atomicity witness; hermetic verification

- **Decision**: stdlib only (`os`, `path/filepath`, `strings`, `encoding/json`); POSIX-only; no new module. Verification is hermetic: **unit tests** pin the tool logic against a temp dir — create-only, **atomicity**, strict-unique, empty-`old_text`, missing-file, parent creation, **and the created-file mode `0644`**; the atomicity guarantee is **unit-tier** (no E2E fault injection exists) and is pinned by a witness: **inject a writer whose `Write` fails after N bytes, then assert the destination is absent or byte-identical and no `*.tmp` residue remains**. **E2E** drives the built binary with the fake provider scripting the write tool calls (`creates a file before answering` / `edits a file before answering`), asserting the file bytes and the refusals; no pty. **Falsifiability witnesses**: revert create-only → the no-clobber Example fails; disable the temp+rename path → the unit-tier atomicity witness fails; loosen `replace_text` to replace-first → the ambiguity Example fails.
- **Rationale**: the standing scope (stdlib, POSIX, offline) and the round-008/021/024 harness precedent; witnesses keep the guarantees falsifiable. Atomicity is verified where it is observable (unit), not dressed up as an E2E contract it cannot be.
- **Alternatives considered**:
  - **A third-party atomic-write library** — rejected: a dependency for `CreateTemp` + `Link`/`Rename`.
  - **An E2E atomicity Example** — rejected (review finding 4): atomicity is not observable end-to-end without fault injection; a vacuous E2E Example would imply a contract the acceptance layer cannot execute. The guarantee is unit-tier.
  - **A pty harness** — rejected: `TELL_ME_FORCE_STDIN_TTY`/scripted tool calls already cover the paths; the project forswore a pty.

## Recorded refinements (PR #61 reference cross-check)

- **R2 — missing vs explicit-empty `content`**: a **missing** `content` key is rejected (recoverable tool error); only an explicit `""` writes an empty file (`FR-010`). Mirrors the reference's `write_file` key-presence guard and closes a silent-0-byte-write path. **Partial mitigation only** — it does not catch a mid-string truncation (see R1).
- **R3 — directory mode**: missing parent directories are created with mode **`0755`** (`FR-008`).

## Forward item (R1) — provider transports can silently truncate large tool-call arguments

Round 029 is the first round with multi-KB tool arguments (`write_file.content`, `replace_text.new_text` — the largest, last-emitted keys). tellme's transports surface **no** finish reason (`grep finish` → 0 transport hits; `maxOutputTokens`/`max_tokens` emitted only when `MAX_TOKENS` is set), so a `MAX_TOKENS`/`length` truncation mid-tool-call is indistinguishable from a complete one and can yield a silently truncated/empty write. The reference learned this the hard way (Anthropic's 4096 default burning retry dollars; Gemini's `FinishReasonMaxTokens`). **Recorded as a deliberate forward item — tracked in issue #62** — because the guard (surface a mid-tool-call `MAX_TOKENS` finish as a **terminal** provider error in both adapters + `Classify`) is a **transport-layer** change and belongs in its own round, not the write-tools round. Silent omission is not acceptable; deliberate deferral is.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: add a **Write filesystem tools** row to the agent tool surface (`write_file` **atomic create-only** via `os.Link`/`EEXIST`; `replace_text` strict-unique + **atomic** via `rename`, with a no-op short-circuit; created-file mode `0644`; **missing** `content` rejected; parent-dir mode `0755`; no security/undo; the `replace_text` whole-file-read hazard; the #62 provider-truncation forward item) and move `write_file`/`replace_text` **out of** the *Not Introduced Yet* write-tools bullet (leaving `append_text`/`undo_file_change` deferred; `delete_path`/`create_directory` remain non-goals).
- `specs/truth/features/cli/**` → **ADD** the write-tools interface feature + `dsl.md` rows (owned by `/axb-dsl-refine`); **MODIFY** `offering-the-agent-tools.feature` prose (four → **six** tools).
- `specs/truth/contracts/**` and `specs/truth/data/**` → **NOOP** (no API surface; the write tools persist no state).
