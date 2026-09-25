# Round 090 — `090-record-hygiene-offered-set-and-direction`

**Theme**: four coupled **record-hygiene** fixes on tellme's own documentation/truth layer — reconcile the
stale **"seven agent tools"** truth row (and give the offered-set claim a **single-sourced, mechanically
checked carrier**), fix the `README.md` tool-surface enumeration, **relocate the operator-declared
direction into an ADR** (README demoted to a pointer), and correct the `cobra` note that calls `retry` a
subcommand. **Truth/record + a test-only carrier** — no product behaviour change.

**Anchor issue**: [#186](https://github.com/gosharplite/tellme/issues/186). **DoD = close it.**

---

## 1. Why this round

tellme's own records must describe the system as it is (`truth-current`). Four surfaces drifted or were
never given a durable home:

1. **(A) A stale truth row.** `specs/truth/features/cli/chat/dsl.md:208` (the row for the step
   `the request offered exactly the agent tools`) documents **seven** tools and omits **`search_files`**
   — but the sibling truth surface `chat/offering-the-agent-tools.feature` says **eight** (naming
   `search_files`, round 071 / ADR 0043), and the **live registry** the step asserts
   (`tests/e2e/steps/tool_usage.go` → `registeredToolNames()`, mirroring the production assembler)
   already returns **eight**. The row's prose is **superseded** and **self-contradicts a sibling truth
   surface** — the round-088 **F-088-1** class. Nothing reddens today (the topology audit parses the
   row's *pattern*; the E2E asserts the *live registry*), so it rotted silently — the round-089
   **RF-089-6** limit. The operator's decision: give the claim a **single-sourced carrier** so it is
   *checked*, not hand-copied.
2. **(B) An incomplete README enumeration.** `README.md`'s *"deliberately small tool surface"* paragraph
   lists the reader family + `search_files` + the write pair, **omitting `list_skills` and `read_image`**
   (both shipped).
3. **(C) An unrecorded direction.** tellme's **operator-declared direction** (no security layer · no
   Windows · bash-first · POSIX-only) is declared in `README.md` — which even claims to be its home —
   and **echoed in `STATUS.md`**, with **no ADR**. Operator decision: the direction belongs in an
   **ADR**; **`README.md` is not the source of truth for directions**.
4. **(D) An imprecise note.** `specs/truth/techstack.md:155` lists `retry` as a *subcommand* alongside
   `browse`; the reference's `retry` is a **flag** (`--retry`).

(B) sits inside the same README section as (C), so they are one edit.

## 2. The change

1. **(A) Reconcile + carry the offered set.**
   - `specs/truth/features/cli/chat/dsl.md:208` — fix the `集合` cell to the **eight** base tools (add
     `search_files`), drop the stale hand-count, and name the **single owner** (the live agent-tool
     registry). The `不該發生` clause (no summarisation tool, no `pipe_commands`) is kept.
   - `specs/truth/features/cli/chat/offering-the-agent-tools.feature` — already correct; confirmed to
     agree (`truth-current`).
   - **The carrier** (test-only): a doc-consistency test in `cmd/tellme` (the package that owns the
     production assembler `agentTools()`) reads the `chat/dsl.md` offered-set cell, extracts its
     backticked tool names, and asserts **set-equality** with `agentTools()` — so adding/removing a tool,
     or editing the documented list, **reddens** the check (W-A). Rides `go test` (not a `make verify`
     member), like the round-031/061 schema gates.
   - **The rider**: correct the stale history comment in
     `tests/e2e/steps/step_r021_t026_chat_then_offered_tools.go` (record round 071's `search_files`).
2. **(B) `README.md`** — reconcile the *"deliberately small tool surface"* enumeration to the shipped
   set (the reader family + `search_files` + the write pair + `list_skills`, with `execute_command` the
   primitive and the capability-gated `read_image` noted).
3. **(C) Direction → ADR.** Write **ADR 0061** (`docs/decisions/0061-operator-declared-direction.md`)
   recording the direction (no security layer — an accepted operator decision, with the accepted-risk
   note; no Windows; bash-first; POSIX-only) with rationale and consequences; add its index row.
   **Demote `README.md`**: the direction section becomes a **summary that points at ADR 0061** (README no
   longer claims to be the source; the "Direction changes are recorded here" line is corrected).
   Reconcile `STATUS.md`'s *Direction* line to point at the ADR.
4. **(D) `specs/truth/techstack.md:155`** — correct the `cobra` note (a subcommand example, e.g. `browse`,
   plus further flag surfaces; `retry` is a flag), keeping the deferral rationale.

No product behaviour change; no new flag/phrase/exit code; `go.mod`/`go.sum` unchanged.

## 3. Requirements

- **FR-001** (A) `specs/truth/features/cli/chat/dsl.md` no longer asserts a stale count/set; the offered-set
  claim is **single-sourced** and **mechanically checked** against the live agent-tool registry — a carrier
  that **reddens** if a tool is added/removed (or the documented list is edited) without reconciliation.
- **FR-002** (A) `chat/dsl.md` and `chat/offering-the-agent-tools.feature` agree (one current description).
- **FR-003** (A rider) the stale history comment is corrected (no false history).
- **FR-004** (B) `README.md`'s tool-surface enumeration matches the shipped set.
- **FR-005** (C) a new **ADR** records the operator-declared direction (rationale + consequences + the
  accepted-risk note); `README.md` **points at** the ADR and no longer claims to be the direction's source;
  `STATUS.md`'s direction line agrees.
- **FR-006** (D) `techstack.md`'s `cobra` note no longer calls `retry` a subcommand.
- **NFR-001** no product behaviour change; no new flag/phrase/exit code; `go.mod`/`go.sum` unchanged.
- **NFR-002** `make verify` green (incl. `verify-adr-index`, `modelith-check`); `go test -count=1 ./...`
  green — E2E counts **unchanged** (the carrier is a **unit** test in `cmd/tellme`, adding no scenario).

## 4. Invariants

- **I-1 — No product code.** The only `.go` change is a **test-only** carrier in `cmd/tellme`
  (`*_test.go`) plus a **comment** fix in `tests/e2e/steps/**`. No production `.go`, no `.feature` step text.
- **I-2 — `truth-current` holds.** After the round, the offered-set claim is current (eight) and agrees
  with the sibling feature; the direction has a durable home (ADR 0061) referenced by README + STATUS.
- **I-3 — E2E unchanged.** No `.feature` change and no step-behaviour change ⇒ E2E scenario/step counts
  are unchanged (330 · 2487).
- **I-4 — Frozen history untouched.** `specs/plans/NNN-*/**` (incl. round 071's package) is never edited.
- **I-5 — `make verify` unaffected** (no gate added/removed/re-scoped); stdlib-only; POSIX-only.

## 5. Scope

**In**: the `chat/dsl.md` offered-set reconciliation (owner `/axb-dsl-refine`) + the test-only carrier + the
step comment; the `README.md` surface enumeration; **ADR 0061** + its index row + the README/STATUS
direction pointers; the `techstack.md` `cobra` note (owner `/axb-technical-research`).
**Out**: any product behaviour; introducing summarisation/pruning/pinning; porting `--retry`/`browse`;
relitigating the settled direction (this round **records** it); a general docs-prose gate (ADR 0060
RF-089-6 — scoped here to the offered-set claim only); frozen plan packages.

## 6. Success criteria

- **SC-001** The carrier test **reddens** under the mutation "remove a tool from `agentTools()` " or
  "edit the documented list" (W-A), and passes at head.
- **SC-002** `chat/dsl.md` names **eight** base tools incl. `search_files`; the feature agrees.
- **SC-003** `make verify-adr-index` is green; ADR 0061 is indexed once and unique.
- **SC-004** `README.md`'s enumeration lists the shipped set; the direction paragraph points at ADR 0061.
- **SC-005** `techstack.md:155` no longer calls `retry` a subcommand.
- **SC-006** `make check` green; E2E counts unchanged (330 · 2487); `go.mod`/`go.sum` unchanged.

## 7. Assumptions

- **A1** The operator's round instruction grants the two locked decisions in #186 ((A) single-source
  carrier; direction → ADR). No `/axb-clarify` needed (0 questions).
- **A2** The **only** stale offered-set surface is `chat/dsl.md:208`; the feature is already current
  (verified). The carrier's single source is the **production** assembler (`cmd/tellme.agentTools()`).
- **A3** No `NEEDS CLARIFICATION`.
