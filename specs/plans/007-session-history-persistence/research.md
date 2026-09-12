# Phase 0 Research: tellme Session History Persistence (Round 007)

Topic: the round-007 session slice — persist each completed turn, **auto-resume** the conversation on the next run, and add the **`--new`** (fresh session) and **`-l N`** (inspect) surfaces — per Clarify Round 1 (Q1 auto-resume always; Q2 include both `--new` and `-l`; Q3 reuse the environment class phrase for a history I/O failure). The contract must follow `tell-me-go`'s persistent-`Session` behaviour while honouring tellme's settled exclusions.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution), testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), and output rendering (glamour) were locked in rounds 001–006. The system still has **one CLI end** (the operator terminal) and adds **no new system end** and **no new external service** and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the CLI Gherkin; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. **Settled exclusions** (out of scope, not deferred): `-b`/`--retry`, history pinning, streaming responses, token-budget pruning, `SafePath`. **Deferred** (not excluded): summarisation — it lands later, with the agent-tools round.

---

## Decision 1: History store — an append-only JSON-Lines file under the session workspace

- **Decision**: Persist the session as a single **append-only JSON-Lines** file, `history.jsonl`, inside the existing per-mode session workspace (`$TELL_ME_HOME/output/<mode>/`, round-001 truth). One line per **completed turn**, a JSON object with the two fields `prompt` and `answer`, serialized with `encoding/json` over a fixed struct (stable field order ⇒ deterministic bytes). A line is written **only after** the turn completes (FR-001/FR-003).
- **Rationale**: Matches the reference's `output/<mode>/history.jsonl` JSON-Lines store (`internal/infrastructure/history`) while dropping what tellme excludes (patch/pin lines, compaction). The workspace already exists and is per-mode, so the file needs no new directory. Append-after-complete makes an interrupted turn invisible (the reference's `history-persisted-after-turn` invariant, kept), and a single whole-line write avoids partial records.
- **Alternatives considered**:
  - **SQLite (the reference's task store)**: gives queries/atomicity, but is a new dependency and a database where a small append-only log suffices — rejected (dependency-free discipline; the store is read wholesale on resume).
  - **A single JSON array file / pretty-printed**: rewrites the whole file on each turn and loses the crash-safety of append-only — rejected.
  - **Store the rendered answer**: binds persisted content to a presentation choice — rejected; the store holds the provider's answer text (Decision 7).

## Decision 2: Resume — load the store and carry prior turns as conversation context

- **Decision**: On a prompt run, read `history.jsonl` before contacting the provider and expand each stored turn into an ordered pair of conversation messages (`user` = prompt, `assistant` = answer). Extend the domain port `llm.Request` to carry this prior conversation (`Messages []Message{Role, Content}` alongside the current `Prompt`); the OpenAI adapter sends `history… + {role:"user", content: prompt}` as the `messages` array. With **empty** history the payload is byte-for-byte the round-004 single-user-message request (no behaviour change on a fresh workspace).
- **Rationale**: Q1 (auto-resume always) requires the earlier turns to reach the provider; the adapter already builds a `messages` array, so the change is a widening of the request value type plus an assembly step, not a transport change. Keeping `Prompt` as the current-turn field preserves the round-004/005/006 call sites and their fakes.
- **Alternatives considered**:
  - **Flatten history into a single prompt string**: loses role structure the OpenAI wire shape expects and degrades answer quality — rejected.
  - **A separate `Conversation` port**: heavier seam than a field on the existing request — rejected.
  - **Resume only under an explicit flag**: rejected by Q1.

## Decision 3: `--new` — archive the active history into a single deterministic archive file

- **Decision**: Add a boolean `--new`. When it runs, move the current session aside by **appending** the active `history.jsonl`'s lines to `history.archive.jsonl` in the same workspace (creating it if absent), then removing the active file. The subsequent session (or the same invocation's prompt) starts with **no** prior context; the archived turns are retained, not destroyed (FR-005/FR-006).
- **Rationale**: A single, fixed-name archive file keeps the file name **deterministic** (no timestamped backup directory), preserves every prior conversation across repeated resets (append, unlike an overwrite), and mirrors the reference's "archive the history" concept (`session.archive`) without its timestamped backup machinery.
- **Alternatives considered**:
  - **Timestamped backup file per `--new`** (`history.archive.<ts>.jsonl`, the reference's backups): non-deterministic filenames make black-box assertions awkward — rejected.
  - **Overwrite a single archive file**: loses the *oldest* conversation on a second `--new` — violates "retained" — rejected.
  - **Delete the active file outright**: destroys history — rejected (Q2/FR-006).

## Decision 4: `-l N` — print the last N messages, plain, with no provider request

- **Decision**: Add an integer `-l N`. It reads `history.jsonl` and prints the last **N messages** (a message per line, `role: content`) to standard output, then exits — **no provider request** (FR-007/FR-008). When fewer than N messages exist it prints all available (FR-009); under `--new`-fresh / empty history it prints nothing and exits successfully. N MUST be a **positive integer**; a missing, zero, or negative N is a usage error (the existing usage-error class phrase + code). Output is **plain text** — history **rendering** is deferred to the later rendered-history slice.
- **Rationale**: Mirrors the reference's `-l <n>` last-messages view at the message granularity users expect, and stays offline/deterministic. Requiring a positive N removes the need to invent a default, keeping the contract exact and the acceptance unambiguous.
- **Alternatives considered**:
  - **Default count when `-l` is bare**: invents a number with no source — rejected; a bare `-l` is a `pflag` missing-argument usage error.
  - **Print turns, not messages**: diverges from the reference's message count — rejected.
  - **Render the listed history (glamour)**: the rendered-history slice belongs to a later round — rejected.

## Decision 5: Flag surface and dispatch precedence

- **Decision**: Keep the existing surface (`-c/--config`, `-d/--diagnostics`, `--version`, `-r/--raw`) and add **only** `--new` and `-l N`. Dispatch precedence: `--version` → `-d` → `-l` → (optional `--new`) prompt turn → boot. `-l` is a terminal reporting command (prints and exits). `--new` with a prompt runs that prompt in the fresh session; `--new` without a prompt boots. The prompt-less boot remains the empty-run fall-through.
- **Rationale**: Extends the round-004 precedence (`--version` → `-d` → prompt → boot) with the two new report/session commands in the natural order, and keeps the round-005 stdin read on the prompt path only (the version/diagnostic/`-l` paths never read stdin).
- **Alternatives considered**:
  - **`-l` as a subcommand** (`tellme history`): the reference uses a flag; a subcommand framework (`cobra`) is a deferred dependency — rejected.
  - **Let `--new` imply a prompt-less boot only**: needlessly forbids "start fresh **and** ask" — rejected.

## Decision 6: Failure contract — reuse the environment class phrase

- **Decision**: A failure reading, appending, or archiving the session history is reported on standard error with the existing environment class phrase `tellme: the runtime home is not usable` (contract-free trailing detail) and exits with the **environment error code (4)** (FR-011).
- **Rationale**: Q3 — the history lives inside the session workspace beneath the runtime home, so this is an environment/workspace condition, exactly the round-005 precedent (which reused the same phrase for the stdin-read path). It keeps the frozen class-phrase vocabulary at **ten** — no new phrase, no new exit code.
- **Alternatives considered**:
  - **A new `the session history is not usable` phrase + code**: grows the frozen vocabulary for a workspace condition the environment phrase already covers — rejected (Q3).
  - **Non-fatal degrade (warn and answer anyway)**: silently loses memory the operator expects to persist — rejected for a *write* failure.

## Decision 7: Persisted content is the provider's answer text (not the rendered form)

- **Decision**: The `answer` field stores the provider's normalized answer text (`llm.Response.Text`) verbatim. Rendering (glamour) and the `-r` flag are output concerns and never touch the store.
- **Rationale**: Keeps the store content-stable and independent of presentation, so `-l` (and the future rendered-history slice) can present it however they like and the raw bytes are deterministic (NFR-002).
- **Alternatives considered**:
  - **Store the rendered/ANSI answer**: couples persisted content to a terminal-dependent presentation — rejected.
  - **Store the full `llm.Response`**: carries no additional content today — rejected (YAGNI).

## Decision 8: Testing — record the sent conversation in the fake provider; unit-test the store

- **Decision**:
  - Extend the **local fake provider** (`net/http/httptest`) to decode and **record the request's `messages` array**, so a scenario can assert that prior turns were actually carried as context (the resume witness), and extend the E2E runner with **filesystem assertions** over `history.jsonl` / `history.archive.jsonl`.
  - Add **pure-helper / unit tests** for: history file path resolution, append (one turn ⇒ one line), reload (lines ⇒ ordered messages), the `--new` archive move, and the `-l` mode selection + positive-count validation.
  - Extend the **offline no-network** scope to include `-l` and a prompt-less `--new` (they must leave the recording sink at zero connections); the prompt-bearing `--new` turn is a chat path and may dial, exactly like the round-004 prompt turn.
- **Rationale**: The resume claim is only falsifiable if the fake records what the client sent; the store mechanics are pure and unit-testable cheaply; the offline guarantee must cover the two new non-prompt commands (NFR-001).
- **Alternatives considered**:
  - **Assert resume only via the answer text** (a scripted reply that echoes context): weaker and provider-script-dependent — rejected in favour of recording the wire request.
  - **Skip the offline extension**: would silently widen the network surface — rejected.

## Residual risks / forward links

- **Round-004 truth MODIFY expected**: the chat turn now also **persists** and **carries prior turns**; `/axb-dsl-refine` is expected to record a **MODIFY** against `specs/truth/features/cli/chat/answering-a-single-prompt.feature` (+ its `dsl.md`) so the single-turn Examples note that a run resumes prior context when history is present (a fresh workspace is unchanged). The `reporting-a-failed-provider-request` feature is unaffected.
- **Class-phrase vocabulary unchanged**: ten phrases, exit codes `0/2/3/4/5/6` — no change this round (Decision 6).
- **Determinism**: stored lines carry no timestamp or ID (only `prompt`/`answer`), so `history.jsonl` bytes are reproducible for a given conversation (NFR-002); the archive file preserves line order.
- **`-l` output format** (`role: content`, one per line) is a plain-text presentation choice this round; the rendered-history slice may revisit it.
- **No new dependency**: the round is stdlib-only (`encoding/json`, `os`, `bufio`); `go.mod` is untouched.
- **Deferred, not excluded**: history summarisation and token-budget pruning wait for the agent-tools round; pinning, `-b`/`--retry`, streaming, and `SafePath` remain out of tellme scope.
