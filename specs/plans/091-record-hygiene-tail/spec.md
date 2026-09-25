# Round 091 — `091-record-hygiene-tail`

**Theme**: the **tail** of the 089/090 **record-hygiene** sweep — three coupled docs/records corrections on
tellme's own truth/documentation layer: (A) drop the stale **"ninth agent tool"** ordinal in
`specs/truth/techstack.md` (and any other **live** bare-ordinal claim), (C) reconcile the stale
**"the vocabulary stays ten"** figure (a **count of frozen class phrases**, not exit codes — and it has since
drifted), and (B, optional) decide whether to bind the e2e tool enumerator to `agentTools()`.
**Truth/record only** — no product behaviour change.

**Anchor issue**: [#188](https://github.com/gosharplite/tellme/issues/188). **DoD = close it.**

---

## 1. Why this round

tellme's own records must describe the system as it is (`truth-current`). The 089/090 sweep fixed the
**offered-set** row and the direction; a **tail** of small count/ordinal claims survived, because —
as recorded in **ADR 0060 §Forward RF-089-6** and **round-090 TD-090-1** — *a docs-prose claim has no
mechanical carrier*, so it drifts and is caught only by review. Three items remain:

1. **(A) A stale ordinal.** `specs/truth/techstack.md:36` (the *In-file content search tool (`search_files`)*
   row) opens *"**Round 071 (ADR 0043):** the **ninth** agent tool — the missing half of the reader trio"*.
   The ordinal is **imprecise against the shipped shape**: by **offer order** `search_files` is the **4th**
   (`list_files`, `read_files`, `get_tree`, `search_files`, …), and in the **base set** it is one of
   **eight** (round 090's reconciled count) — the capability-gated `read_image` is the only "ninth". The
   ordinal is defensible only under a *chronological add-order* reading that nothing states. The fix is to
   **drop the ordinal** (matching round 090's handling of the hand-count) and identify the tool by name +
   round/ADR.
   - `docs/decisions/0043-search-files-tool.md:23` also says *"the **ninth** agent tool"* — that is an
     **Accepted ADR** and, per `docs/decisions/README.md`, **ADRs are immutable once Accepted (supersede,
     do not edit)**. It is a historical record of the state *at decision time*; **leave it verbatim.**
   - The **live** index row `docs/decisions/README.md:69` (for ADR 0043) also says *"a ninth tool offered to
     every provider"* — that is a **live** curated surface (annotated in place since round 085), so it is in
     scope for FR-2.

2. **(C) A stale, ambiguous figure.** `specs/truth/techstack.md:105` (the *Provider request retry* row) ends
   *"… the frozen class phrase `the provider request failed` + exit code **6** (no new phrase, no new exit
   code; **the vocabulary stays ten**)."* The sentence names exit codes, so a bare "ten" reads as ten exit
   codes. In fact the count's subject is the **frozen class-phrase vocabulary**, whose live authority is the
   interface-root row `specs/truth/features/cli/dsl.md` — which enumerates **eleven** phrases — while the
   exit-code set is **seven** (`internal/cli/exitcode.go`: `0/2/3/4/5/6/7`). The figure is a **frozen-time**
   count from round 007 (`specs/plans/007-*/research.md` — "ten phrases, exit codes `0/2/3/4/5/6`") and has
   since drifted on **both** axes. The fix is to **name the subject and drop the drifting number** (the
   round-090 approach). The same stale conflation recurs on two more **live** surfaces:
   `specs/truth/techstack.md:31` and `specs/truth/features/cli/chat/dsl.md` (the round-079 note: "the
   ten-value exit-code set"). The ADR occurrences (`0051:43`, `0052:23`, `0055:45`, `0058:92`) are
   **immutable Accepted ADRs** — left verbatim on the same principle as ADR 0043.

3. **(B, optional) A mirrored list.** The offered-tool set is mirrored across ~7 live surfaces; only the
   `chat/dsl.md` `集合` cell is mechanically bound (round 090's carrier). Round-090 reviewer finding
   **R-090-1** (recorded, not required): optionally bind the e2e enumerator
   (`tests/e2e/steps/tool_usage.go` → `registeredToolNames()`) to the production assembler `agentTools()`.

## 2. The change

1. **(A) Drop the ordinal — `techstack.md` + the ADR index row.**
   - `specs/truth/techstack.md:36` — remove "the **ninth** agent tool"; keep the name + round/ADR
     (`the in-file content search tool (`search_files`) — the missing half of the reader trio`).
   - `docs/decisions/README.md:69` (the ADR 0043 index row) — replace "a ninth tool offered to every
     provider" with an ordinal-free phrasing ("offered to every provider").
   - `docs/decisions/0043-search-files-tool.md` — **unchanged** (immutable Accepted ADR, historical).
2. **(C) Name the subject, drop the stale number — the three live count surfaces.**
   - `specs/truth/techstack.md:105` — "the vocabulary stays ten" → "the class-phrase vocabulary is
     unchanged" (a name, not a drifting count).
   - `specs/truth/techstack.md:31` — "the ten-value exit-code set" → "the exit-code set".
   - `specs/truth/features/cli/chat/dsl.md` (round-079 note) — "the ten-value exit-code set" → "the
     exit-code set".
   - ADRs `0051`/`0052`/`0055`/`0058` — **unchanged** (immutable; historical).
3. **(B) Defer with a recorded reason** (§5) and home the deferral on a **live issue** (a durable home per
   Bootstrap Agent Rule 11). Rationale in §5.

No product code, no `.feature` step text, no `.feature` row semantics, no `go.mod`/`go.sum` change.

## 3. User stories

- **US-1** — As the tellme operator, I want the truth rows and the ADR index to describe the shipped tool
  set without a stale ordinal, so a reader is not misled about which tool is "ninth".
- **US-2** — As the tellme operator, I want a "no new phrase / no new exit code" invariant to name its
  subject, so a reader cannot mistake it for a count of exit codes.

## 4. Requirements

- **FR-1 (A)** — `specs/truth/techstack.md` no longer asserts an unqualified "ninth agent tool"; the
  `search_files` row identifies the tool by **name + round/ADR**, not by a bare ordinal (`truth-current`).
- **FR-2 (A)** — **no live** (non-frozen) surface carries a bare, unqualified tool ordinal that contradicts
  the offer order / the base-set count. The **immutable** Accepted ADR 0043 body is left verbatim.
- **FR-3 (C)** — every live surface that restates the round-007 "ten" figure names its subject and does not
  carry the **stale** number; the ADR occurrences are left verbatim (immutable).
- **FR-4 (B, optional)** — the e2e tool enumerator is bound to `agentTools()` **or** the item is explicitly
  deferred with a recorded reason homed on a durable surface.
- **NFR-1** — no product behaviour change; no new flag/phrase/exit code; `go.mod`/`go.sum` unchanged.
- **NFR-2** — `make verify` green (incl. `verify-adr-index`, `modelith-check`); `go test -count=1 ./...`
  green; E2E counts unchanged (round 091 adds no carrier, so no delta is expected).

## 5. Item (B) — the decision (deferred, with a recorded reason)

**Decision: defer (B).** Binding the e2e enumerator to `agentTools()` **by construction** requires one of:

- **moving the base-set composition out of the composition root** (`cmd/tellme`) into an importable package
  so both the e2e steps and the assembler share one owner — which would **reverse round 044's placement**
  ("the assembler now lives in `cmd/tellme`") and re-open the ADR 0013/0039 composition-root boundary for a
  **records** round; or
- **importing the e2e harness** (`tests/e2e/steps`, godog) into the composition root's test package — an
  inverted, heavy test dependency; or
- a **brittle Go-source-parsing** carrier.

None is proportionate to a **tail** round whose NFR-1 is "no product behaviour change", and there is **no
defect**: `registeredToolNames()` is already bound to the recorded request **behaviourally** by the E2E
(round 090's own framing of R-090-1: *"a scoped refactor, not a defect"*). The deferral is homed on a
**live issue** ([#189](https://github.com/gosharplite/tellme/issues/189), a durable home), not only in this
frozen package.

## 6. Invariants

- **I-1** — `specs/truth/**` stays the single source of truth; the `.modelith.md` files are generated and
  are **not** touched (nothing modelled changes — ADR 0041 escape hatch, recorded in `plan.md`).
- **I-2** — no **Accepted ADR** body is edited (ADR 0043 + the `0051`/`0052`/`0055`/`0058` counts are
  historical). The **ADR index** (`docs/decisions/README.md`) is a live curated surface and may be edited.
- **I-3** — frozen `specs/plans/NNN-*/**` packages (incl. 007's origin note and 089/090's) are never
  modified.
- **I-4** — the e2e/unit carrier for the offered set (round 090's `cmd/tellme/deps_offered_set_test.go`)
  is untouched and stays green.

## 7. Success criteria

- **SC-001 (A)** — `specs/truth/techstack.md:36` and `docs/decisions/README.md:69` carry **no** bare tool
  ordinal; ADR 0043's body is byte-identical to `dev`.
- **SC-002 (C)** — the three live count surfaces name their subject and carry no stale "ten"; the ADRs are
  byte-identical to `dev`.
- **SC-003 (B)** — the deferral is recorded in `plan.md`/`tasks.md` and homed on a live issue.
- **SC-004** — `make verify` green · `modelith-check` no drift · `go.mod`/`go.sum` unchanged · E2E green.

## 8. Assumptions

- **A1** — the operator's round instruction ("Open a new aixbdd round, the goal is to close #188") grants
  the intent; #188 fixes the goals and the two locked decisions (drop the ordinal; name the count's
  subject), so **no clarify escalation** is needed.
- **A2** — the truth authority for the class-phrase count is the interface-root row
  `specs/truth/features/cli/dsl.md` (enumerating the phrases); the exit-code authority is
  `internal/cli/exitcode.go`.
