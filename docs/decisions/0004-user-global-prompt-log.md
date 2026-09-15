# ADR 0004 — User-global interactive prompt log (`~/.tellme/global_prompts.jsonl`)

- **Status:** Accepted
- **Date:** 2026-09-15
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** round 028 (`specs/plans/028-user-global-prompt-log`); PR #59 architectural review (TD-1/TD-3, RF-1);
  round 015 (`specs/plans/015-interactive-tui-prompt` — the original environment-scoped placement);
  round 026 (`~/.tellme/tools-count.jsonl` — the user-global root precedent)

## Context

Round 015 introduced tellme's `-i` interactive prompt log at the **environment-scoped**
`<TELL_ME_HOME>/output/global_prompts.jsonl` — an append-only JSON-Lines file holding one
`{"timestamp":"<RFC3339>","prompt":"<text>"}` per line, read newest-first (deduped) to seed suggestions and
written only under `-i`. It was deliberately placed at the `output/` **root** and kept **byte-for-byte
compatible with `tell-me-go`**, so every persona tool sharing a `TELL_ME_HOME` saw one pool of recent prompts.

That scoping has two consequences the operator rejected:
1. the history is **per-`TELL_ME_HOME`** — starting a new environment/repository resets the suggestions, so a
   prompt typed in one environment is not offered in another;
2. it is shared **with `tell-me-go`**, whose writer/reader contract tellme must then never break.

Round 026 had already established a **user-global** durable-state root for a different log
(`~/.tellme/tools-count.jsonl`, a machine-wide tool-usage log), deliberately outside the `TELL_ME_HOME`
namespace. Round 028 (operator request) extends that root to the prompt log.

## Problem

The operator wants the interactive prompt history to **follow the user** across every tellme
environment/repository/mode (one pool of recent prompts), rather than being scoped to a single
`TELL_ME_HOME`. The change must also **not lose** the prompts already recorded in the environment-scoped
file, and must not silently split the history across two files.

## Decision

Move the `-i` shared prompt log to a **per-user** file at **`~/.tellme/global_prompts.jsonl`**, and add a
**first-use seed**:

- **Location.** The log lives at `<user-home>/.tellme/global_prompts.jsonl` (the user-global root the
  round-026 tool-usage log already uses). The path is resolved via the **CLI-injected user-home resolver**
  (`userHomeDir = os.UserHomeDir`), mirroring the round-026 `newToolUsageStore` seam, so CLI unit tests stay
  hermetic. Both the suggestion **read** and the `-i` **record write** target this file; the
  environment-scoped `output/global_prompts.jsonl` is **no longer read or written**.
- **Seed-on-absent migration.** When the user-global file is **absent**, tellme copies the current
  environment-scoped `<TELL_ME_HOME>/output/global_prompts.jsonl` into it **verbatim** (a copy, not a move;
  the source is left in place). The seed is an **explicit `Seed(ctx) error`** invoked once at the composition
  root (not a constructor side effect), is **best-effort** (any I/O error is swallowed so the prompt is never
  aborted), and **never overwrites** an existing destination (a missing source starts empty).
- **Shape unchanged.** The record shape, the append-only write, the newest-first-deduped read, and the
  compaction policy are unchanged; the `-i`-only write rule is unchanged.
- **Recorded divergence — the `tell-me-go` sharing contract is dropped.** After this ADR, tellme **no longer
  shares** the prompt log with `tell-me-go` (which keeps the environment-scoped file). tellme also never read
  `tell-me-go`'s legacy locations (`.tellmego/prompts.jsonl`, `<home>/global_prompts.jsonl`) — those were a
  reference-side migration path, not a tellme behaviour — so this is a recording, not new skip logic.

## Alternatives considered

1. **Keep the environment-scoped `output/` location** — rejected: it is exactly the per-`TELL_ME_HOME`
   scoping the operator asked to remove.
2. **An XDG path (`~/.config/tellme/` or `$XDG_STATE_HOME`)** — rejected: `~/.tellme/` is tellme's
   established user-global root (round 026); introducing XDG would split the convention for no gain.
3. **Merge/dedupe across every environment's log on first use** — rejected: ambiguous source precedence and
   more complexity than needed; the operator chose a single verbatim copy source.
4. **Keep writing the environment-scoped file too (so `tell-me-go` still sees tellme's prompts)** — rejected:
   two writers, guaranteed drift, and the operator directed a single read/write target.

## Consequences

- The interactive prompt history is now **per-user** (shared across every tellme env/repo/mode) — the
  operator's goal — at the cost of **no longer sharing with `tell-me-go`**. This is an **operator-accepted**
  trade-off.
- The user-global `~/.tellme/` root now hosts **two** append-only logs (this one and the round-026
  tool-usage log); the relocation **escalates** the materiality of the existing "no `flock`" forward item —
  the file is now shared by every process on the machine, so two writers can race the ≈150 KiB compaction.
  Compaction/rotation for the user-global root remains a forward item.
- The first `-i` run after the upgrade seeds the user-global file from whichever environment is active then;
  other environments' env-scoped logs are **not** merged (a disclosed, operator-accepted single-source seed).
- `specs/truth/techstack.md` and `specs/truth/data/data-model.dbml` (`prompt_log_entry`) reflect the new
  location + lifecycle; the `chat` interface truth (`recording-the-shared-prompt-log.feature` +
  `carrying-over-the-environment-prompt-log.feature` + `dsl.md`) names the new path and the seed.
- Immutable once `Accepted`; a future change of the log location supersedes this ADR rather than editing it.
