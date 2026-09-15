# Research: user-global interactive prompt log (round 028)

Topic: relocate tellme's round-015 `-i` **shared prompt log** from the environment-scoped
`<TELL_ME_HOME>/output/global_prompts.jsonl` to a **per-user** `~/.tellme/global_prompts.jsonl`, with a
**one-time seed** from the old location when the new file is first needed. Operator-directed (session
2026-09-15); the target path and the seed rule were given verbatim.

> **AIxBDD must-ask questions** — this is **not** a starting project: `specs/truth/techstack.md` already
> pins the system's ends (a single CLI end), the CLI BDD techstack (`godog`), and the test strategy
> (E2E acceptance + fast pure-helper units). None is changed this round, so no clarify is needed (the
> round-004/013/027 precedent).

## Decision 1: The log moves to a per-user file at `~/.tellme/global_prompts.jsonl`

- **Decision**: Resolve the log path at `<user-home>/.tellme/global_prompts.jsonl` via the **CLI-injected
  user-home resolver** (`userHomeDir = os.UserHomeDir`, passed into the adapter — mirroring the round-026
  `newToolUsageStore` seam; the adapter keeps the path join). Both the suggestion **read** and the `-i`
  **record write** target it, replacing `<TELL_ME_HOME>/output/global_prompts.jsonl`.
- **Rationale**: it reuses the **round-026 user-global root** (`~/.tellme/`, already the home of
  `tools-count.jsonl`), so tellme gains no new base-directory convention; it makes the prompt history
  **follow the operator** across every environment/repository/mode (the operator's stated goal), instead
  of being reset per `TELL_ME_HOME`.
- **Alternatives considered**:
  - Keep `<TELL_ME_HOME>/output/global_prompts.jsonl` — rejected: the operator wants a user-global log.
  - An XDG path (`~/.config/tellme/` or `$XDG_STATE_HOME`) — rejected: `~/.tellme/` is tellme's established
    user-global root (round 026); introducing XDG would split the convention for no gain.

## Decision 2: Both read and write move; the environment-scoped file becomes a seed source only

- **Decision**: After this round the adapter reads and writes **only** the user-global file.
  `<TELL_ME_HOME>/output/global_prompts.jsonl` is **not** read or written by tellme; it is used solely as
  the seed source (Decision 3).
- **Rationale**: a single log avoids split-brain suggestions (two files that can diverge) and keeps the
  "read newest-first + deduped" semantics against one source.
- **Alternatives considered**:
  - Dual-write (keep the env file in sync for tell-me-go) — rejected: two writers, guaranteed drift, and
    the operator directed a single read/write target.

## Decision 3: Seed-on-absent — a verbatim copy from the environment-scoped log

- **Decision**: When `~/.tellme/global_prompts.jsonl` is **absent**, copy
  `<TELL_ME_HOME>/output/global_prompts.jsonl` into it **verbatim** (byte-identical), when that source
  exists. It is a **copy** (the source is left in place), it **never overwrites** an existing destination,
  and a **missing source** yields an empty log with no error. A **blank runtime home** (`TELL_ME_HOME` unset
  → `home == ""`, which the `-i` path tolerates because `runTUIPrompt` proceeds past a `resolve()` error) is
  treated as "no source": the seed is **skipped**, so it never reads a cwd-relative `output/global_prompts.jsonl`
  (PR #59 review micro-note).
- **Rationale**: preserves the operator's existing prompt history across the upgrade (no silent loss);
  deterministic and trivial; exactly the rule the operator stated ("the existing output/global_prompts.jsonl
  file will be copy to ~/.tellme/...").
- **Alternatives considered**:
  - Merge/dedupe across multiple environments' logs — rejected: ambiguous source precedence, more
    complexity than the operator asked for; only one environment's file is the verbatim source.
  - Move (rename) the file — rejected: destroys the environment's copy and surprises anyone still reading
    the old shape.

## Decision 4: The seed runs in the history adapter, best-effort, before the first suggestion read

- **Decision**: Implement the seed as an **explicit `Seed(ctx) error`** on the infrastructure history
  adapter (`internal/infrastructure/history`), invoked **once** at the composition root **before** the
  interactive read (not a constructor side effect — the tracker is built **twice** per `-i` run: once to
  seed suggestions, once to `Append` the submission), and make it **best-effort**: any I/O error
  (unreadable source, unwritable destination) is swallowed and the prompt continues with an empty/partial
  log.
- **Rationale**: the tracker is only built on the `-i` path, so the seed naturally fires only there (the
  log is never touched by a non-`-i` run — preserving the round-015/027 `-i`-only contract); best-effort
  matches the round-026 tool-usage log and the round-028 NFR-001.
- **Alternatives considered**:
  - A separate CLI migration step / subcommand — rejected: extra surface to run, and it would have to be
    remembered by every environment.
  - Seed lazily on the first **write** instead — rejected: a read-only session would not seed, so the
    operator would not see their carried-over suggestions on first open.

## Decision 5: An unresolvable or unwritable home degrades to a no-op

- **Decision**: If `os.UserHomeDir()` fails, or `~/.tellme/global_prompts.jsonl` cannot be created/written,
  the prompt log degrades to a **no-op** (no suggestions, no record) and the `-i` prompt proceeds normally.
- **Rationale**: the prompt log is a convenience; it must never block or fail the interactive prompt
  (NFR-001). This mirrors the round-026 tool-usage log's best-effort posture.
- **Alternatives considered**:
  - A hard error — rejected: breaking the prompt over a non-essential log is worse than losing suggestions.
  - Fall back to the environment-scoped path — rejected: a second location contradicts the operator's
    single user-global target and would resurface the old split.

## Decision 6: The tell-me-go sharing contract is dropped (recorded divergence)

- **Decision**: Record that tellme **no longer shares** the prompt log with `tell-me-go` (which keeps
  `<TELL_ME_HOME>/output/global_prompts.jsonl`). tellme **never** read tell-me-go's legacy locations
  (`.tellmego/prompts.jsonl`, `<home>/global_prompts.jsonl`) — that was a **reference-side** migration
  path, not a tellme behaviour — so this is a recording of the reference, **not** new skip logic. The
  record **shape** stays unchanged (one `{"timestamp","prompt"}` per line) — it no longer needs to
  round-trip with tell-me-go, but keeping the shape avoids a needless format change.
- **Rationale**: inherent to the operator-directed relocation (Decision 1/2); the round-015/016
  "byte-for-byte with the other personas" contract is superseded.
- **Alternatives considered**:
  - Keep writing the environment-scoped file too so tell-me-go still sees tellme's prompts — rejected: the
    operator chose a single user-global read/write target; the divergence is intended.

## Decision 7: No new dependency; POSIX-only; hermetic verification

- **Decision**: Go standard library only (`os.UserHomeDir()`, `encoding/json`, `os`); the E2E suite points
  `HOME` at a per-scenario temp directory **and** a temp runtime home (the round-026 precedent), so the
  operator's real `~/.tellme/` is never touched; the seed is pinned by unit tests (absent → verbatim copy;
  present → no-op; missing source → empty) and end-to-end (seed-on-absent, no-overwrite).
- **Rationale**: tellme's standing scope (stdlib-only, POSIX-only) and the round-026 hermetic-home pattern.
- **Alternatives considered**:
  - Add a config key for the log path — rejected: the operator fixed the location; no configuration surface
    is needed.

## Residual risks / forward items

- **Unbounded `~/.tellme/` growth** — the user-global root now hosts two append-only logs (the round-026
  `tools-count.jsonl` and this prompt log); compaction/rotation for the user-global root remains a shared
  forward item (the prompt log keeps its round-015 in-file compaction policy, unchanged here).
- **Cross-process contention (escalated by the relocation, TD-3)** — the round-015 log was per-`TELL_ME_HOME`
  (one writer per environment); after the relocation **every** environment/process on the machine shares
  one file, while the adapter still takes **no `flock`**. `O_APPEND` keeps single-line appends safe, but two
  concurrent processes can race the ≈150 KiB **compaction** (optimistic size-checked). This **increases** the
  materiality of the existing "no `flock`" forward item; recorded here and qualified in `techstack.md`. No
  code change this round. The **first-use seed is likewise not atomic**: `Seed` creates the destination with
  `O_EXCL` and then writes the bytes in a single `Write`, so a concurrent other-process reader can observe
  the file mid-copy (empty/partial) for that one read — acceptable for a one-time migration (and `O_EXCL` is
  the right no-overwrite primitive), recorded so a future reader knows the seed publish is not atomic
  (PR #59 implementation-review note).
- **Single-source seed** — only the first interactive run after the move (in whichever environment is
  active while the destination is absent) seeds the log; other environments' env-scoped logs are not
  merged. Operator-accepted (verbatim copy), disclosed in `spec.md` → edge cases.
- **Round-015/016 truth churn** — the `tell-me-go` sharing clause and the path must be updated in
  `specs/truth/techstack.md` (this owner, below), `specs/truth/data/data-model.dbml` (`/axb-data-plan`), and
  the `specs/truth/features/cli/chat/**` rows plus the module note (`/axb-dsl-refine`).
