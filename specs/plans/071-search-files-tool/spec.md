# Feature Specification: `search_files` — a bounded, deterministic in-file search tool (round 071)

**Feature Branch**: `071-search-files-tool`

**Created**: 2026-09-20

**Status**: Draft (specified — clarify **folded**, Q1–Q3 locked; see §Clarify strategy)

**Anchor**: issue [#147](https://github.com/gosharplite/tellme/issues/147) — the candidate inventory. `search_files` is the **one** Filesystem candidate ("the missing half of the reader trio"); a candidate round is opened **from operator value**, which the operator gave this session.

**Input (operator, 2026-09-20, this session)**:

> *"Refer to search_files of tell-me-go, why it is worth porting to tellme?"* → (answer) → *"yes"* — open the round; the **ignore-policy / bounds / ordering** divergences are the first `/axb-clarify` questions.

**Behaviour intent**: **ADD (a new agent tool)** — add `search_files` to the read-only reader family: a **bounded, deterministic, own-contract** in-file content search over a directory subtree. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first — the bar this tool must clear, and what is deliberately NOT copied

- **The design-intent bar (`README.md`)**: a dedicated tool earns its place only if it beats bash on a real axis. `search_files` clears **context-boundedness** (an unbounded `grep -rn` can blow the window), **determinism / testable contract** (a tool's output is *executable truth*; a shell one-liner's is not), and it is the **missing half of the reader trio** (`list_files` / `read_files` / `get_tree` can find and read a file but have no way to locate content).
- **Not copied — the reference's security gate.** `tell-me-go`'s `search_files` runs behind a `SafePath` `PathValidator` (`sp.IsPathSafe`). tellme has **no security layer** (settled exclusion) — the tool reads **whatever path it is given**, like its sibling readers.
- **Not copied — the reference's ignore policy.** `tell-me-go` skips directories matched by its `defaultWorkspacePolicy` (`.git`, `node_modules`, `vendor`, `bin`, `obj`, `output`, `dist`, `testdata`, `configs`, `secrets`, plus any hidden dir) — that policy serves a **secret-scanning** concern. tellme has no `WorkspacePolicy`; this tool's own scope is **Q2**.
- **Not copied — the reference's non-determinism.** Its result order is worker/channel order (a ≤8 worker pool). A tool whose output is executable truth must be **deterministic** (**I-4**; a recorded tellme divergence).
- **Recorded further differences (review A1 / TD-071-4).** The reference **appends `" (truncated)"`** when it cuts a line at 500 (tellme cuts **silently** — `RF-071-3`), **skips any file > 1 MiB** (tellme scans any size; the byte budget bounds the *result*), and probes binaries at **1024 B** vs tellme's **8000 B**. The divergence ledger is **five** deliberate/recorded differences, not three.

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `417dacd`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/tools/filesystem.go` | the reader family (`listFiles`, `readFiles`, `getTree`) — each built with the shared `resourceSchema` builder, **bounded at source** by the resolved byte budget, returning a **nil-error timeout result** on its deadline (FR-018). |
| `internal/domain/tools/tools.go` | the `Tool` port `Execute(ctx, arguments string, budget ByteBudget) (string, error)` + `ToolContract{DefaultTimeout}`; the `MediaTool` capability (round 070). |
| `internal/domain/tools/timeout.go` · `outputsink.go` | the round-024 **resource contract** (budget + timeout + ceiling). |
| `cmd/tellme/deps.go` | `agentTools()` / `assembleAgentTools(spec)` — the **non-overridable** production assembler (the round-031 schema gate iterates it). |
| `cmd/tellme` (recordable union) | the `--tool-usage` recordable set = union(base, capability-gated) — every recordable tool must be listed (round 062 F-062-1). |
| `internal/infrastructure/tools/filesystem.go` `readMaxPerCall` | the ≤50 files/call **degenerate cap** (round 021/024) — the reader family's precedent for a hard cap **alongside** the byte budget. |
| `README.md` · `specs/truth/techstack.md` | the *deliberately small tool surface* (the reader trio + write pair + `execute_command` + `list_skills` + the vision-gated `read_image`) — tellme's agent surface is these **eight** today. |

**Reference behaviour** (measured in the local `tell-me-go` tree — `internal/tools/workspace/registration.go:112`, `search.go`, `internal/pkg/concurrentsearch/concurrentsearch.go`): args `{path=".", query, is_regex=false, reason}`; a recursive walk over `path`; per match `path:line: <line trimmed>`; the line is trimmed and cut at **500** chars; a hard **100**-match cap appends `\n... (truncated)`; no match → `N matches found …` (a **result**, not an error); **skips binary** files and files **> 1 MB** (tellme does **not** skip by size — the byte budget bounds the result); a **10 MB** max line token.

---

## Design (the shape is a `/axb-technical-research` decision; the clarify items are marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **ADD** a ninth tool `search_files` to the reader family — a **bounded, own-contract** in-file content search over a directory subtree. | locked (round goal) |
| **S-2** | It is **bounded at the source** — the round-024 resource contract (the resolved **byte budget** + the per-call **timeout** + a **nil-error timeout result**). | locked |
| **S-3** | It is **deterministic** — a fixed, reproducible result order (not the reference's worker order). | locked (I-4) |
| **S-4** | It carries the **mandatory `reason`** (ADR 0025) and the shared resource schema (round-031 `required ⊆ properties`). | locked (I-1/I-7) |
| **S-5** | **No `SafePath` / consent** — the tool reads the path it is given (the settled no-security-layer exclusion). | locked |
| **S-6** | **Search mode** — literal substring by default with an `is_regex: true` RE2 opt-in. | **locked (Q1 → A)** |
| **S-7** | **Scope / ignore policy** — skip binary files + a 10 MB max-line token; **no** directory ignore list; unreadable skipped best-effort. | **locked (Q2 → A)** |
| **S-8** | **Bound & shape** — the byte budget (primary, **every** return path incl. no-match) + a hard 100-match cap (first 100 by path of the first 1000 walk-order candidates) + a 500-**byte** per-line trim; `path:line: <trimmed line>`, sorted path-then-line. | **locked (Q3 → A)** |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The universal `reason` gate is unchanged** (ADR 0025) — every call MUST carry a renderable reason.
- **I-2 — The round-024 resource contract** — the tool bounds its own output at the source and returns a **nil-error timeout result** on its deadline (never the `error:` path).
- **I-3 — No security layer · no new dependency · stdlib-only · POSIX-only · hermetic.**
- **I-4 — Deterministic output** — a fixed, reproducible order (testability is part of the tool's justification).
- **I-5 — Offered to every provider** — the gate is `none` (it is **not** vision-gated like `read_image`).
- **I-6 — `--tool-usage` recordable** — the tool joins the recordable union (report order preserved).
- **I-7 — Schema gate** — `required ⊆ properties` holds for the tool's schema (round 031).
- **I-8 — No config change · no UX change** to the existing surface.

---

## Clarify strategy

**FOLDED — Q1/Q2/Q3 answered one at a time (this session).** Three gaps changed the tool's **formal acceptance criteria** — its schema and its observable output *are* the acceptance contract — so they were interviewed, never assumed. Locked answers:

- **Q1 → A — search mode**: a **literal** substring by default, with an **`is_regex: true`** opt-in for an RE2 pattern.
- **Q2 → A — scope / ignore**: skip **binary** files + a **10 MB max-line token**, **no** directory ignore list, unreadable paths skipped best-effort; the caller scopes with `path`.
- **Q3 → A — bound & shape**: the round-024 **byte budget is the primary bound**, plus a hard degenerate **100-match cap** and a **500-byte per-line trim**; format `path:line: <trimmed line>`, sorted **path asc, then line asc**; the truncation marker is the shared `TruncationMarker`.

**No `NEEDS CLARIFICATION` remains.** (`/axb-clarify` was delegated by `/axb-specify`; the operator's standing *"keep going unless you need to ask"* instruction applied to Q3 after Q1/Q2 were answered — Q3's **recommended** option A was therefore locked.)

**Technical (not clarify) decisions deferred to `/axb-technical-research`**: the walk/sort mechanics, the scan cap, and the file layout (all in `research.md` D1–D7).

---

## User Stories (proposed — subject to the clarify answers)

### US1 — The agent can locate content in a directory subtree (Priority: P1)

A caller (the model via the tool) asks for matches of a literal string or pattern under a directory; the tool returns the matching lines, **bounded** and **deterministic**.

**Why P1**: it is the round's whole goal — the missing half of the reader trio (find content, where the sibling tools only list/read/tree).

**Independent verification**: a hermetic E2E (scripted fake provider → one `search_files` call) against a fixture tree containing a known string in ≥2 files; assert the returned lines name the matches.

**Acceptance (proposed)**:

1. **Given** a fixture tree where the query appears in several files, **When** `search_files` runs with that query, **Then** the result names each match as `path:line: text` in a **deterministic** order.
2. **Given** a query with **no** match, **When** `search_files` runs, **Then** it returns a **"0 matches"** result (a result, **not** an error).

### US2 — The output is bounded at the source (Priority: P1)

The result never blows the model's context: it is bounded by the round-024 resource budget, with a truncation marker; an oversized scan stops at the bound.

**Why P1**: context-boundedness is the axis that earns the tool its place (the README bar).

**Independent verification**: a fixture tree with far more matches than a small injected budget; assert the result stays within the budget and ends with the truncation marker; a separate case asserts the nil-error timeout result.

**Acceptance (proposed)**:

1. **Given** a small injected byte budget and a query matching far more, **When** `search_files` runs, **Then** the result stays within the resolved budget and carries the truncation marker.
2. **Given** a scan that exceeds the effective per-call timeout, **When** it runs, **Then** it returns the **nil-error timeout result** (FR-018).

### US3 — The search mode is explicit (Priority: P1 — the set is fixed by Q1)

**Why P1**: the literal-vs-regex choice is the usable contract; the tool must not silently treat a literal as a regex (or vice versa).

**Acceptance (proposed)**:

1. **Given** a **literal** query, **When** `search_files` runs, **Then** it matches **substrings** (regex metacharacters are literal).
2. **Given** an **invalid** pattern in regex mode, **When** `search_files` runs, **Then** it returns a **stable tool error** naming the bad pattern.

*(The exact surface — an `is_regex` flag vs a required `mode` — is **Q1**.)*

---

## Functional Requirements (proposed)

- **FR-001** — The system MUST offer a `search_files` agent tool (a **ninth** base tool) with args `{path?, query, …}` plus the mandatory `reason`.
- **FR-002** — The tool MUST search **file contents** under `path` (default `.`), **recursively**.
- **FR-003** — The tool MUST bound its own output **at the source** (the round-024 byte budget) and return a **nil-error timeout result** on its deadline (I-2).
- **FR-004** — The tool MUST produce a **deterministic** result order (I-4).
- **FR-005** — The tool MUST carry the mandatory `reason`, the shared resource schema (`required ⊆ properties`), and the resource contract (I-1/I-2/I-7).
- **FR-006** — The tool MUST join the `--tool-usage` **recordable union** (I-6).
- **FR-007** — The tool MUST express the search mode **explicitly** (literal vs regex — per Q1), never coercing silently.
- **FR-008** — The tool MUST be **offered to every provider** — the capability gate is `none` (I-5).
- **FR-009** — An **empty** `query` MUST be a stable tool error (the query is required), matching the reference's `query argument is required`.
- **FR-010** — The tool's **observable output** (the match-line format, the cap/truncation marker, the "0 matches" text, and the bounds) MUST be single-owned and testable — recorded in `specs/truth/features/cli/**` + `dsl.md` (the *reader family* truth rows) at `/axb-dsl-refine`, and in the `techstack.md` reader row at `/axb-technical-research`.

## Success Criteria (proposed)

- **SC-001** — `search_files` is offered by the production assembler (`agentTools`) and appears in the `--tool-usage` report; the round-031 schema gate passes for it.
- **SC-002** — A hermetic E2E journey covers the bounded, deterministic search (plus the "0 matches" and invalid-pattern cases); the tool suite stays green.
- **SC-003** — `make verify` green (layer gate 0); `go.mod`/`go.sum` **unchanged** (stdlib-only).
- **SC-004** — The tool **beats bash on the boundedness axis, demonstrably**: a scan whose match volume far exceeds the budget returns a result within the budget with a stable marker (a falsifiability witness against an unbounded `grep`).

---

## Edge cases (proposed)

- A query matching **nothing** → a "0 matches" **result** (not an error).
- An **invalid regex** (in regex mode) → a stable tool error naming the bad pattern.
- **Binary** files → skipped (per Q2).
- **Unreadable** paths / dangling symlinks → skipped, a bounded best-effort result (per Q2).
- A **very large file** or a **very long line** → bounded (a max-line token + a trim; per Q2/Q3).
- `path` naming a **file**, not a directory → a stable result/error (per the eventual design).
- **Empty `query`** → the required-argument error (FR-009).
- The walk reaching the **timeout** → the nil-error timeout result (I-2).

## Key entities

`search_files` (the new tool) · the `Tool` port + `ToolContract` · the reader family (`NewFilesystemTools`) · the agent assembler (`agentTools` / `assembleAgentTools`) · the `--tool-usage` recordable union · the reader truth rows (`specs/truth/techstack.md` + `specs/truth/features/cli/chat/**` + `dsl.md`).

## Assumptions

- **A1** — The change is **infra / truth-local** (`internal/infrastructure/tools` + `cmd/tellme` [+ `internal/domain/tools` if the contract changes] + the truth rows); the agent loop, the adapters, the CLI, and the config are untouched.
- **A2** — The tool is **text-content** search only; a **filename**-pattern tool (`find_file`) is **out of scope** (it is an excluded candidate in [#146](https://github.com/gosharplite/tellme/issues/146)).
- **A3** — The reference is a **behaviour/architecture reference**; its **security gate** and its **ignore policy** are **not** copied (the design-intent bar + the no-security-layer direction).

## Out of scope

- The **Go/AST** symbol tools ([#147](https://github.com/gosharplite/tellme/issues/147)'s 24) — a separate, larger candidate cluster.
- **Filename** search (`find_file`) — excluded (#146).
- The **enterprise / MCP** candidates — shell-reachable or better as MCP servers.
