# Technical research — round 090 `090-record-hygiene-offered-set-and-direction`

**Topic**: reconcile four record-hygiene surfaces, and give the offered-set claim a **single-sourced,
mechanically checked carrier**.

**Owner**: `axb-technical-research` (owns the `techstack.md` `cobra` note; no new technology, no new
dependency).

---

## D1 — The defect (A): a superseded truth row with no carrier

`specs/truth/features/cli/chat/dsl.md:208` documents the offered set as **seven** tools and omits
`search_files`. The sibling truth surface `chat/offering-the-agent-tools.feature:3,14` says **eight** and
names `search_files` (round 071 / ADR 0043). The **live** registry the step asserts — via
`tests/e2e/steps/tool_usage.go` `registeredToolNames()`, which mirrors the production assembler — already
returns **eight**. So the row's prose is **superseded** and **contradicts** a sibling truth surface (the
round-088 **F-088-1** class), and **nothing reddens**: the topology audit parses the row's *pattern*, and
the E2E step enumerates the *live* registry (so it passes regardless of the row's prose). This is the
round-089 **RF-089-6** limit (a docs-prose claim with no mechanical carrier).

## D2 — Decision: the offered-set claim is single-sourced to the code, with a reddening carrier

Operator decision (issue #186): **(A) → single-source carrier** — the claim is *checked*, not hand-copied.

**The single source** = the **production** agent-tool assembler. `cmd/tellme` `agentTools()`
(`= assembleAgentTools(deps.ToolSetSpec{})`, the base set — no `vision`) is the non-overridable
production registry the CLI builds; it is already the anchor of the round-031 schema gate
(`cmd/tellme/deps_test.go`). *(The e2e `registeredToolNames()` mirrors it in a package that cannot import
`cmd/tellme` — package `main`; the carrier uses the production assembler directly, the stronger source.)*

**The carrier** (test-only): a doc-consistency test **in `cmd/tellme`** that
1. reads `specs/truth/features/cli/chat/dsl.md` (relative `../../specs/…`, the e2e precedent),
2. locates the row line containing `the request offered exactly the agent tools`,
3. extracts the row's `集合` cell backticked tool names, and
4. asserts **set-equality** with the names from `agentTools()`.

A mutation that adds/removes a tool (or edits the documented list) **reddens** it (W-A). It rides
`go test` (not a `make verify` member), like the round-031/061 schema gates. Rationale for the parse
target: the `集合` cell is the row's structured enumeration (pipe-delimited, individually backticked), so
the parse is bounded and not a whole-file scan.

**Prose reconciliation** (`chat/dsl.md:208`): fix the `集合` cell to the **eight** base tools (add
`search_files`), **drop the stale hand-count**, and name the owner (the live registry). Keep the
`不該發生` clause (no summarisation tool, no `pipe_commands`). The `.feature` is already current (D1) — no
change.

## D3 — (B) README surface enumeration

`README.md`'s *"deliberately small tool surface"* paragraph lists the reader family + `search_files` +
the write pair, omitting **`list_skills`** (round 033) and **`read_image`** (round 062, capability-gated).
Fix: reconcile to the shipped set (readers + `search_files` + write pair + `list_skills`, with
`execute_command` the primitive and `read_image` noted as capability-gated). (B) sits inside the same
section as (C).

## D4 — (C) the direction belongs in an ADR

tellme's operator-declared direction (no security layer · no Windows · bash-first · POSIX-only) is
declared in `README.md` ("Design Intent & Direction"), echoed in `STATUS.md`, and **not** in any ADR;
`README.md` even asserts *"Direction changes are recorded here…"*. Operator decision: **the direction
belongs in an ADR; the README is not the source of truth for directions.** Records per
`docs/decisions/README.md` (*"a decision that other artifacts or future rounds depend on and must be able
to cite"*) — exactly this. **ADR 0061** records: the three directions (+ POSIX-only), the rationale (the
reference's always-bypassed guardrails were pure overhead; the Windows cross-platform tax; bash-first as
the universal primitive), the **consequences** (a deliberately small tool surface; an **accepted
operator risk** — destructive/out-of-tree actions are explicit, not oversights), and a forward note
(direction changes go through a superseding ADR, not a README edit). Then the README paragraph becomes a
**summary that points at ADR 0061** (and its "recorded here" line is corrected), and `STATUS.md`'s
*Direction* line points at the ADR.

## D5 — (D) the `cobra` note

`specs/truth/techstack.md:155` reads *"deferred until subcommands (`browse`, `retry`) exist."* `browse` is
a **subcommand**; the reference's `retry` is a **flag** (`--retry`). Fix the wording to a subcommand
example + further flag surfaces, keeping the `cobra` deferral rationale.

## D6 — No product behaviour; the carrier is the only `.go`

The round changes **no** production `.go` (the carrier is `*_test.go`, the step comment a comment). No
`.feature` step text, no `go.mod`/`go.sum`, no `make verify` member. `docs/domain-model/**` is **not**
modelled (a doc *count* fix + a decision record + a comment touch no modeled entity). ADR 0061 supersedes
nothing (it is the first record of the direction).

## D7 — The witness (falsifiability)

- **W-A (real carrier)** — the carrier reddens under a tool add/remove or a documented-list edit;
  reproduced + reverted.
- **W-B / W-D (manual, carried)** — README enumeration == the shipped set; `techstack.md:155` no longer
  calls `retry` a subcommand (no mechanical home; inspection, recorded honestly as such — the round does
  **not** overclaim a reddening test for these).
- **W-C (durable)** — `make verify-adr-index` (standing member) proves ADR 0061 is indexed once/unique;
  README/STATUS point at it.

## D8 — Scope guard

No product code, no `.feature` step text, no `go.mod`/`go.sum`, no new dependency, no `make verify`
change. Frozen `specs/plans/NNN-*/**` untouched. The docs-prose *gate* question (a general carrier for
such claims) stays out of scope (ADR 0060 RF-089-6) — this round scopes the carrier to the offered-set
claim only.
