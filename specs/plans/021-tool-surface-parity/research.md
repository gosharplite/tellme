# Phase 0 Research: tellme Agent Tool Surface Parity (Round 021)

Topic: align tellme's **agent tool surface** with the reference (`tell-me-go`) — **remove** the round-008 summarisation tool, **reshape** `list_files` and `read_files` to the reference's parameters/behaviour (adding a required `reason`), and **add** the reference's `get_tree`. A **tool-contract parity** round: no new system end, no new external service, no new dependency.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), and the agent tool loop (`AgentLoop`, `MAX_TOOL_LOOP`, per-tool timeout, `[tool] …` `stderr` log) were locked in rounds 001–020. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. **Settled exclusions** (out of scope): `SafePath`/consent, `-b`/`--retry`, history pinning, streaming, token-budget pruning. The **operator locked D1–D5** before `/axb-specify` (see `spec.md`); this document fixes the *how*.

Reference: `~/tmp/github/gosharplite/tell-me-go/internal/tools/workspace/{registration.go,reader.go}`.

---

## Decision 1: `read_files` becomes multi-file (`filepaths: string[]`)

- **Decision**: Replace the round-008 single `{path}` argument with the reference signature `{ "filepaths": string[] (required), "reason": string (required) }`. The result concatenates one block per requested path, **in request order**, each block framed by a header line `--- File: <path> ---`.
- **Rationale**: D1 — the reference's `read_files` reads many files in one call ("minimize LLM roundtrips"); matching the array signature is the core parity change and lets multi-file work cost one loop iteration. Request-order output is deterministic and matches the reference loop.
- **Alternatives considered**: keep a singular `path` (weaker parity — rejected by D1); accept both `path` and `filepaths` (extra surface — rejected; tellme re-creates the reference, it does not shim).

## Decision 2: `list_files` adopts the reference shape

- **Decision**: Schema `{ "path": string (optional), "reason": string (required) }`; an omitted/empty `path` defaults to `"."`. Output begins with `Contents of <path>:` and lists one entry per line prefixed `[d] ` (directory) or `[f] ` (file). Entries stay in `os.ReadDir` order (sorted by name).
- **Rationale**: D3/D5 — the reference's format; `os.ReadDir` already sorts, so the current sort is preserved with no behaviour surprise.
- **Alternatives considered**: keep the round-008 `names, dirs suffixed /` format (diverges — rejected); drop `path` (the reference defaults it — rejected).

## Decision 3: add `get_tree`

- **Decision**: Add a third read-only tool, `get_tree` — schema `{ "path": string (optional), "max_depth": integer (optional), "reason": string (required) }`; omitted `path` defaults to `"."`; omitted/`≤0` `max_depth` defaults to **2**. Output is a connector tree: each entry on its own line prefixed by `├── ` (or `└── ` for the last sibling) at the current indent, with continuation indentation `│   ` / `    `; a directory is recursed into **except `.git`**; recursion stops at `max_depth`.
- **Rationale**: D5 — the reference bundles `get_tree` with the reader; adding it makes the offered surface exactly the reference's read trio. The format is copied from the reference's `buildTree`/`writeTreeEntry`.
- **Alternatives considered**: defer `get_tree` (rejected by D5); include a dir suffix or sort differently (deviates from the reference — rejected).

## Decision 4: `reason` is required and echoed into the tool-loop log

- **Decision**: Every filesystem tool schema lists `reason` as **required**. tellme has no consent layer (D4), so the value is not acted on for authorization; instead the loop **echoes it into the existing `[tool] …` `stderr` log line**: `[tool] <name> arguments=<args> reason=<reason> result=<result>`. The loop extracts a top-level `reason` string from the tool-call arguments JSON (generic — any tool carrying `reason` gets it echoed; a tool without one omits the segment).
- **Rationale**: D2 — the operator wants `reason` required (wire parity) **and** visible. The `stderr` tool-loop log (round-008 Decision 7) is the natural, already-pinned surface; it keeps `stdout` byte-exact (FR-016) and needs no new output channel. Extracting `reason` in the loop (not per-tool) keeps the tools free of presentation concerns and works uniformly.
- **Alternatives considered**: require-but-ignore (rejected by D2); a new dedicated `reason:` line (extra surface — rejected; the existing log line is the home); a per-tool `Reason()` method on the port (widens the port — rejected; the loop already holds the raw arguments).

## Decision 5: read bounds and edge handling — reference parity

- **Decision**: `read_files` per-file behaviour matches the reference:
  - **Size cap** `maxReadSize = 100000` bytes; a longer file is cut to the cap and the block ends `... (truncated)` (`io.LimitReader(f, cap+1)` to detect truncation).
  - **Binary** file → block body `(Binary file, cannot display as text)` (a small stdlib probe — a NUL byte within the leading bytes — no new dependency).
  - **Directory** given → block body `ERROR: path is a directory, use list_files instead`.
  - **Unreadable path** → block body `ERROR: <cause>` (inline result; the loop continues — a recoverable tool result, not a tool failure).
  - **Per-call cap** `maxFilesPerCall = 50`; more → the whole result is `Error: requested too many files (N). Maximum is 50 files per call.`; empty/omitted `filepaths` → a tool argument error.
  - Each file block ends with a blank line (content + `\n\n`).
- **Rationale**: D3 — "functionality as close as possible." The inline `ERROR:` handling keeps a bad path in one file from failing the whole read, matching the reference. The 100000 cap supersedes the round-008 1 MiB placeholder (a tellme-local number, not an external contract) and still satisfies NFR-001 (bounded result ⇒ cannot exhaust the context).
- **Alternatives considered**: keep 1 MiB (diverges — rejected by D3); return a tool error on a bad path (fails the whole call — rejected).

## Decision 6: remove `summarize_history`

- **Decision**: Delete the summarisation adapter (`internal/infrastructure/tools/summarize.go`) and stop appending it in the registry factory (`internal/cli/cli.go` `newToolRegistry`), so a prompt run offers **exactly** `list_files`, `read_files`, `get_tree`. Remove the corresponding truth (the `chat` interface feature + its DSL rows). The loop, the registry port, `MAX_TOOL_LOOP`, and the `the tool request failed`/exit-7 contract are **unchanged** — a request for the now-unknown tool therefore fails under the existing contract.
- **Rationale**: FR-010/FR-011 (D5-adjacent, operator's first instruction). The tool was an LLM-backed reader whose behaviour overlaps the reference's *different* `summarize_history` (context compaction), which tellme does not have — removing it leaves a coherent, single-purpose read surface.
- **Alternatives considered**: keep it (contradicts the operator instruction — rejected); replace it with the reference's context-compacting version (a different, larger capability — out of scope).

## Decision 7: testing — unit-test the tools + extend the fake provider; no pty

- **Decision**:
  - **Unit** (stdlib `testing`, table-driven): the three tools' schemas (incl. required `reason`), `list_files` `Contents of …`/`[d]`/`[f]` output + default `path`, `read_files` multi-file framing + request order + the 100000 cap + binary + directory + the ≤50 cap + empty-arguments error, `get_tree` connector output + default `max_depth` + `.git` skip, and the loop's `reason` echo in the `[tool] …` line.
  - **E2E** (godog): extend the fake provider to script `read_files`/`list_files`/`get_tree` tool calls and to record the offered tool definitions; add/refresh the `chat` interface scenarios; **remove** the summarise scenario and its step files. Assert the offered set is exactly the three, `stdout` stays byte-exact, and the `reason` appears on `stderr`.
  - No pty; the existing hermetic harness and the working-directory file fixtures suffice.
- **Rationale**: The tools are pure (path in → text out) and cheaply unit-testable; the offered-tool-set and loop behaviour are only falsifiable E2E via the recording fake. Removing the summarise steps keeps the suite green.
- **Alternatives considered**: assert only via answer text (weaker — rejected); keep the summarise scenario (would fail — rejected).

## Decision 8: no new dependency; no security layer

- **Decision**: stdlib-only (`os`, `path/filepath`, `io`, `sort`, `strings`, `encoding/json`); `go.mod`/`go.sum` untouched. No `SafePath`, no `UserInteractor`, no new failure class/exit code — the tools read whatever path the model gives (D4).
- **Rationale**: parity needs nothing beyond the standard library; the reference's `IsPathSafe`/binary helper have stdlib equivalents; re-opening the security layer is a separate slice.
- **Alternatives considered**: import the reference tool layer (impossible/unwanted — rejected); add `SafePath` (out of scope — rejected).

---

## Residual risks / forward links

- **Instruction vs. semantics of `reason`**: the reference uses `reason` for consent; tellme has none, so `reason` is required-and-echoed, not enforced (D2/D4). Recorded so it is not mistaken for a consent gate.
- **Instruction/behaviour split**: `/axb-dsl-refine` MUST update the `chat` interface truth — MODIFY the `read_files` Given/Then rows to the multi-file signature, ADD the `get_tree` feature + rows, DELETE the `summarising-the-conversation` feature + rows, and pin the `reason` echo and the `list_files`/`read_files` output formats.
- **Data truth**: the persisted tool-step shape (`{tool, arguments, result[, signature]}`) is unchanged — only the `read_files` **arguments content** now carries `filepaths[]` (expected `data/**` NOOP).
- **Context growth**: `read_files` results are now up to `50 × 100000` bytes — larger than the old single-file 1 MiB ceiling. This is the reference's own bound and is still finite; the loop re-sends the growing conversation each iteration (pruning remains excluded), so a multi-hundred-KB read raises per-turn input cost — a known characteristic, bounded by `MAX_TOOL_LOOP`.
- **`get_tree` depth**: the reference prints entries for `max_depth` levels below the root (its `depth > maxDepth` cut); the tests must pin the exact level count to the reference, not an off-by-one.
- **`.git`**: `get_tree` lists `.git` but does not recurse into it (reference behaviour) — pinned so a future "skip dotdirs" change is deliberate.
- **Deferred, still out of scope**: write/shell tools, tool-call concurrency, `SafePath`, the reference's context-compacting `summarize_history`.
