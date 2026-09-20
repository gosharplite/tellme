# ADR 0043 — `search_files`: a bounded, deterministic in-file content search tool

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0025](0025-mcp-tool-call-reason.md) (the universal `reason` gate — unchanged), round 024 (`specs/plans/024-*` — the tool **resource contract** this tool bounds itself with), round 021 (`specs/plans/021-*` — the reader family shape), [ADR 0011](0011-layer-discipline-gate.md) (layer discipline), round 071 (`specs/plans/071-search-files-tool` — this ADR's round), issue [#147](https://github.com/gosharplite/tellme/issues/147) (the candidate inventory)

## Context

`tellme`'s agent surface is deliberately small (`README.md` *Design Intent & Direction*): a tool earns its place only if it **beats bash on a real axis** — context-boundedness, a deterministic/testable contract, or reliability. The reader family (`list_files`, `read_files`, `get_tree`) can **find and read** a file but has **no way to locate content**; today that means an unbounded `grep -rn`, which can flood the model's context window and whose output is not a testable contract. Issue [#147](https://github.com/gosharplite/tellme/issues/147) records `search_files` as the **one Filesystem candidate** — "the missing half of the reader trio".

The reference (`tell-me-go/internal/tools/workspace/search.go` + `internal/pkg/concurrentsearch/`) provides `search_files(path, query, is_regex)`; but it differs from tellme in **three deliberate divergences** (1–3 below) plus **three further recorded differences** (4–6; the review-A1 fold, TD-071-4):

1. a **`SafePath` `PathValidator` gate** (tellme has **no** security layer — a settled exclusion) — *deliberate divergence*;
2. a **`defaultWorkspacePolicy` directory ignore list** (a *secret-scanning* concern tellme does not own) — *deliberate divergence*;
3. a **worker-pool result order** (non-deterministic — a tool's output is executable truth and must be reproducible) — *deliberate divergence*;
4. the reference **appends `" (truncated)"`** when it cuts a line at 500; tellme cuts **silently** (recorded as `RF-071-3`);
5. the reference **skips any file > 1 MiB**; tellme scans files of any size (the **byte budget** bounds the *result*, the **10 MB** line token bounds a *line*);
6. (minor) the reference's binary probe reads **1024 B**; tellme probes **8000 B** (with `io.ReadFull`, so the probe is deterministic).

## Decision

**D1 — ADD `search_files` as the ninth agent tool, offered to every provider** (capability gate `none` — not vision-gated). It is a sibling of the reader family (`internal/infrastructure/tools/search.go`), built with the shared `resourceSchema` (so `required ⊆ properties` holds — round 031) and declaring the reader default timeout.

**D2 — Search mode: literal by default, `is_regex` opt-in** (clarify Q1 → A). A literal query uses `strings.Contains`; `is_regex: true` compiles an RE2 pattern. An **invalid pattern is a stable tool error** naming it (the tool-error path — a non-terminal, recoverable result), never the loop's terminal request failure. A missing/empty `query` is a stable tool error (the required argument).

**D3 — The output is bounded at the source (the round-024 resource contract)** (clarify Q3 → A). **Every** return path — including the no-match result — is bounded by the resolved **byte budget** (`truncateToBudget`, the shared `TruncationMarker`); the tool never lets a caller-controlled string (the echoed `query`/`path`) exceed the budget. The reported set is the first **100** matches (path order) **of the first 1000 candidates collected in walk order** (`searchScanCap`; the walk is deterministic, so the selection is too — but the reported 100 are *not* the 100 globally-smallest paths on a tree larger than the scan cap). Each match line is trimmed at **500 bytes** (a byte cut, rune-boundary safe via `ToValidUTF8` — the readers' bounds are byte budgets; `grep -n` semantics). Each match is rendered `path:line: <trimmed line>`. Matches are sorted **path asc, then line asc** — the ordering divergence from the reference (whose order is worker-dependent; divergence #3 above). A no-match search is a **result** (`0 matches found …`), not an error. A deadline observed returns the **nil-error timeout result** (FR-018).

**D4 — Walk scope: binary files skipped, a 10 MB max-line token, NO directory ignore list** (clarify Q2 → A). The walk recurses (`os.ReadDir`, sorted by name), skips binary files (an 8000-byte NUL probe, read with `io.ReadFull`) and unreadable entries **best-effort**, and searches a file argument directly. There is deliberately **no** directory ignore list — tellme owns no `WorkspacePolicy`; the caller scopes with `path`. This follows the sibling readers' "read whatever you are given" stance. **Disclosure:** an unreadable **root** (or a vanished directory) yields the same `0 matches found …` as a genuine miss — a best-effort skip, so the model cannot distinguish "nothing matched" from "could not read it" (a recorded consequence of the Q2 answer, not a defect).

**D5 — No `SafePath`/consent gate.** The tool reads the path it is given — the settled no-security-layer direction (round 008 Clarify Q3 / round 021 D4).

**D6 — The tool joins the recordable union.** `search_files` is added to the agent assembler (`cmd/tellme`), so it is offered in every prompt run, enumerated by the round-031 schema gate, and listed in the offline `--tool-usage` report.

## Consequences

- The agent can locate content in a directory subtree **without** an unbounded shell `grep`; the result is bounded, deterministic, and carried by executable truth (`specs/truth/features/cli/chat/searching-file-contents.feature` + `chat/dsl.md`).
- **No new dependency** (stdlib `regexp`/`bufio`/`path/filepath`/`sort`); POSIX-only; hermetic.
- **No config change, no CLI-flag change**; the reason gate, the resource contract, the per-call timeout, and the observer hooks are unchanged.
- `thenOfferedTools`'s single-sourced expected set (`registeredToolNames`) grows to **eight** base tools; the `--tool-usage` recordable union follows.
- `README.md`'s tool-surface prose and the domain model's `Tool` entity (nine tools incl. the vision-gated `read_image`) are updated in the same round ([ADR 0041](0041-domain-model-drift-guard.md) — the model is load-bearing).
- **Folded from review A1** (2026-09-20): **TD-071-1** (bound the no-match path), **TD-071-2** (a unit pin for the FR-018 timeout result), **TD-071-3** (the e2e interface carrier for determinism), **TD-071-4** (the divergence inventory above), and the NITs (byte-vs-char wording · the scan-cap selection semantics · the unreadable-root disclosure · the file-argument pin + the deterministic binary probe).

## Verification

- `go test -count=1 ./...` green, incl. the new E2E journey (`searching-file-contents.feature`: 4 scenarios) and the unit pins (`internal/infrastructure/tools/search_test.go`).
- **Falsifiability witnesses** (reproduced then reverted): remove the deterministic sort ⇒ the determinism pin reds; drop the match cap ⇒ the capped-marker pin reds; treat a literal as a regex ⇒ the literal-default pin reds.
- `make verify` green (layer gate 0 · `modelith-check` · `verify-fmt` · `verify-adr-index` · lint 0 · govulncheck); `go.mod`/`go.sum` unchanged.
- The topology audit is unchanged (the new feature's steps match their `dsl.md` rows).

## Forward

- **RF-071-1** — the reference's *filename*-pattern tool (`find_file`) stays an **excluded** candidate ([#146](https://github.com/gosharplite/tellme/issues/146)); `search_files` is content search only.
- **RF-071-2** — the Go/AST symbol tools ([#147](https://github.com/gosharplite/tellme/issues/147)'s 24) are a separate, larger candidate cluster; not opened here.
- **RF-071-3** — the walk reports **no** per-line truncation marker (the line is silently capped at 500 **bytes**); the reference appends `" (truncated)"` at that cut — a marked variant is deferred (divergence #4 above).
- **RF-071-4** — the 10 MB max-line token and the 1000-candidate scan cap are fixed constants; a configurable form is deferred (the scan-cap selection semantics are disclosed in **D3**).
- **RF-071-5** — `searchBinProbe = 8000` duplicates `binary.go`'s NUL-probe magnitude; hoisting it into a single owner is deferred.
