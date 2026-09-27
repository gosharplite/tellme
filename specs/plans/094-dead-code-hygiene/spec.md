# Round 094 — `094-dead-code-hygiene`

**Theme**: two things in one PR — (1) **cleanup**: remove the small, measured set of
**genuinely-dead exported symbols** that no shipped gate can see; (2) **carrier**: add an
**advisory** `make dead-code` target (never fails) so the exported-dead class has a mechanical
surface — plus **ADR 0064** recording the advisory decision and superseding the clause that says
tellme ships no `dead-code`. No product behaviour change.

**Anchor issue**: [#198](https://github.com/gosharplite/tellme/issues/198). **DoD = close it.**

> ⚠ **Governance — this round REOPENS a settled decision under explicit operator intent.**
> Bootstrap **Agent Rule 11 addendum (b)** forbids reopening a declined decision without explicit
> operator intent. This round **supersedes ADR 0042 §D5's "no `dead-code`" wording** (its **§D4
> percentage-coverage decline stands**; the **§D4 *reachability-orphan* half is answered** by this
> round's carrier) and **corrects the `specs/truth/techstack.md` reachability clause** that cited
> round 068. The retired **ADR 0038 §Forward `RF-068-1`** — which is about the *unpaired-call
> diagnostic's E2E carrier*, **not** reachability — stays **retired and inert**: this round neither
> fires nor unfires its trigger, and does not reopen it. The operator filed #198 as a round candidate
> and directed *"Open a new aixbdd round, the goal is to close #198"* — the explicit intent. The
> **carrier form is deliberately NOT the reference's heavy `cmd/deadcode`**.

---

## 1. Why this round

tellme's quality posture includes **no coverage tooling** (ADR 0042 §D4; `techstack.md` *Coverage
tooling — DECLINED*): a coverage *percentage* is misleading because the E2E contract runs the built
binary as a subprocess. That stance left **one** class a profile uniquely adds — a **reachability
orphan** — *asserted* (round 068) to be **caught by reasoning**. The operator's
measurement (issue #198, 2026-09-27) **falsifies that assertion**: a whole-program
reachability pass finds a small set of **genuinely-dead exported symbols** the standard `unused`
linter cannot see (it treats any **exported** symbol as used), and tellme has **no carrier** for
this class at all.

Re-measured this round (2026-09-27, session 84; see §2), on the round branch's base `dev` `ae300e9`:

- `golangci-lint run ./...` and `staticcheck ./...` both report **0 issues** — the `unused`
  linter's `exported-is-used` default hides the class; the config **cannot** be widened to catch it
  (tested `exported-is-used: false` + `whole-program: true`; it never reports an unused **exported**
  symbol).
- **Vanilla** `golang.org/x/tools/cmd/deadcode@v0.47.0` (pinned) with **`-test`** reports **7**
  items on `dev` `ae300e9` (see §3) — 4 of the issue's 7 plus 3 the issue's inventory did **not**
  list (`cli.noopCallObserver.OnCallBegin/OnCallEnd`, `ui.sharedSource.Suggest`). Without `-test` it
  reports **1151** items (head `77db4c1`; **1157** on `dev` `ae300e9`) — test-reachable code reads as
  dead — so `-test` is **mandatory**.
- **Provenance (measured, load-bearing).** The PATH binary named `deadcode` on the documented dev
  host is the reference's **heavy** `tell-me-go/cmd/deadcode` (`go version -m $(command -v deadcode)`
  → `path github.com/gosharplite/tell-me-go/cmd/deadcode`), which prints `[PRIVATE]`/`[DEAD]` noise
  (284 lines, exit 0) — **not** tellme's carrier. The target therefore **probes the binary's
  provenance** and **skips** (exit 0, naming the vanilla install route) if the PATH binary is not
  `golang.org/x/tools/cmd/deadcode` (ADR 0064 D5 / RF-064-3).

## 2. The re-verification (do not trust a tool verdict alone)

Each candidate was re-verified by **grep** (zero callers) at implementation time, not by the tool
alone:

| Symbol | Live references (grep) | Verdict |
| --- | --- | --- |
| `config.(Config).EffectiveUseTUIPrompt` | definition only (the rule is re-implemented inline in `cli.tuiRequested`) | **remove** |
| `openai.NewWithHTTPClient` | definition only | **remove** |
| `agent.DefaultToolTimeout` | definition; a **doc comment** in `tool_contract.go` names it | **remove** (+ fix the comment) |
| `di.TokenResolver` (`= mcp.TokenSource`) | definition only (`NewGhTokenResolver` returns `mcp.TokenSource` directly) | **remove** |
| `harness.RunWithStdin` | definition; a **doc comment** in `step_t014…go` names it | **remove** (+ fix the comment) |
| `harness.RunBinary` | definition only | **remove** |
| `mcptest.SchemaWithProperty` | definition only (`schemaJSON` stays — used by `AnnotatedSchema`) | **remove** |
| **`cli.noopCallObserver`** (type + `OnCallBegin` + `OnCallEnd` + the conformance line) | **never instantiated** — `compositeObserver` is built with a real renderer (`cli.go:875`) and test doubles; the only reference is its own `var _ agentport.CallObserver = …` | **remove** — a **tool-verified + grep-verified genuinely-dead production symbol** the issue's inventory missed (recorded divergence, §7) |

**KEEP — do NOT delete (false positive):**

| Symbol | Why |
| --- | --- |
| `ui.sharedSource.Suggest` (`internal/ui/tuiprompt_test.go`) | **Interface-conformance-only** (`var _ domaintui.Source = sharedSource{}` + `var _ prompt.Source = sharedSource{}`) — the assertions pin that `domaintui.Source` and `prompt.Source` have **identical method sets** (the load-bearing assignability the adapter relies on). Structural typing the analyzer cannot see; removing the method would delete a **contract pin**, not dead code. |
| `llm.(ProviderError).Unwrap` (+ the `ErrIncomplete`/`resolveError` unwrap methods) | Required by `errors.Is`/`errors.As` unwrap chains — structural-typing usage. **Not currently surfaced** by vanilla `deadcode -test` (the RTA sees the interface call), so it needs **no** filter term. The predicate deliberately does **not** carry a name-wide `.*\.Unwrap$` alternative: measured (F-094-3), that swallows a **genuinely-dead new** `Unwrap` method — a NEW finding must surface. If the RTA ever reports a *live* unwrap FP, it is added as a **recorded symbol** then (ADR 0064 RF-064-2), never a name-wide alternative. |

## 3. The cleanup inventory (remove)

Delete exactly the **8** entries of §2's **remove** rows (10 symbols counting the two
`noopCallObserver` methods): `EffectiveUseTUIPrompt` · `NewWithHTTPClient` ·
`agent.DefaultToolTimeout` · `di.TokenResolver` · `harness.RunWithStdin` · `harness.RunBinary` ·
`mcptest.SchemaWithProperty` · `cli.noopCallObserver` (+ its `var _` conformance). The two
**comment** references (the `agent` alias doc, the harness doc) are corrected to name the surviving
symbols. No other file changes.

## 4. User stories

- **US-1 (P1)** — As the tellme operator, I want the genuinely-dead exported symbols removed, so the
  tree carries no unreachable surface the shipped gates cannot see.
- **US-2 (P1)** — As the tellme operator, I want an **advisory** `make dead-code` carrier, so the
  exported-dead class has a mechanical surface that surfaces **new** findings while printing
  **nothing** on the clean tree.

## 5. Requirements

- **FR-1** — each **remove** row of §2/§3 is absent from the tree (grep: no reference except the
  removed definition). [Verification Intent: unobservable → grep + `make check` green]
- **FR-2** — `make dead-code` exists and is **advisory, never fails** (exit 0 always), and is **not**
  a `make verify` member. [Verification Intent: unobservable → the target's exit code + a grep of
  the `verify:` aggregate line]
- **FR-3** — the target invokes the **vanilla** `golang.org/x/tools/cmd/deadcode`, pinned
  (`v0.47.0`), as a **PATH dev-tool binary** (like `modelith`/`golangci-lint`) — **no**
  `go.mod`/`go.sum` change — with **`-test`**; an **absent tool prints an install hint + exits 0**
  (never fail). [Verification Intent: unobservable → the target recipe + a run with a scrubbed PATH]
- **FR-4** — the target **filters the recorded FP class** (the interface-conformance-only class —
  measured member `ui.sharedSource.Suggest`; **not** a name-wide `Unwrap` alternative, which would
  swallow a genuinely-dead new `Unwrap` method — F-094-3) so the **findings body** on the **clean**
  tree is **empty** (the run still prints its banner + `✓` + `advisory done` lines) and only **new**
  findings appear. [Verification Intent: unobservable → a clean-tree run reports no findings; W1
  positive control prints the injected symbol]
- **FR-7** — the target **skips** (exit 0, naming the vanilla install route) when the PATH `deadcode`
  is **not** the vanilla `golang.org/x/tools/cmd/deadcode` binary (a provenance probe), so a wrong
  binary is a **no-op**, never a silently re-meant advisory; and it **distinguishes a tool failure**
  (non-zero exit — a compile error etc.) from a clean tree, so a quiet run always means *clean*.
  [Verification Intent: unobservable → EC-003 + EC-004 runs]
- **FR-5** — **ADR 0064** records the advisory decision (tool pin, advisory-not-a-gate, the FP
  policy, **not-a-catalog**, **not** the reference's `cmd/deadcode`, the provenance hazard) and
  states it **supersedes ADR 0042 §D5's "no `dead-code`" wording**; **ADR 0042** gets a back-pointer
  (its **index row** + a §D4/§D5 forward annotation; the **§D4/D5 body left verbatim**).
  [Verification Intent: unobservable → `make verify-adr-index` + inspection]
- **FR-6** — `specs/truth/techstack.md` gains a row for the advisory target (mirroring the
  `modelith-drift` row); `docs/domain-model/quality.modelith.yaml` adds `dead-code` as an
  **advisory** `QualityGate` member alongside `modelith-drift`; `make modelith-render` regenerates
  the `.md`; `make modelith-check` is drift-free; `docs/decisions/README.md` gains the ADR 0064
  index row. [Verification Intent: unobservable → `modelith-check` + `verify-adr-index`]
- **NFR-1** — **no product behaviour change** (the removed symbols have zero callers); no new
  flag/phrase/exit code; `go.mod`/`go.sum` unchanged; no new dependency. [Verification Intent:
  unobservable → `go build ./...` + `make check` + `git diff --stat`]
- **NFR-2** — `make verify`'s member list is **unchanged** (no new gate); `make check` green;
  `go test -count=1 ./...` green with **E2E counts unchanged** (the round adds no carrier).
  [Verification Intent: unobservable → the `verify:` aggregate + the E2E run]

## 6. Edge cases

- **EC-001** — the `deadcode` binary is **absent** from PATH ⇒ the target prints the install hint
  and **exits 0**. [Verification Intent: unobservable → a scrubbed-PATH run]
- **EC-002** — a **new** genuine dead export is introduced ⇒ the target **reports it** (and still
  exits 0). [Verification Intent: unobservable → W1 positive control]
- **EC-003** — the PATH binary is the **heavy** `tell-me-go/cmd/deadcode` (name collision) ⇒ the
  **provenance probe** skips it (exit 0, naming the vanilla install route), so the advisory never
  silently re-means. [Verification Intent: unobservable → the probe's run on the dev host]
- **EC-004** — a **tool failure** (the tool exits non-zero — e.g. the tree does not compile) ⇒ the
  target prints `analysis did not run (tool exit N)` and still exits 0, so a quiet output never
  means *unknown*. [Verification Intent: unobservable → a broken-tree run]

## 7. Invariants

- **I-1** — every removed symbol has **zero** callers (grep-verified); removal is
  **behaviour-neutral**.
- **I-2** — **no** Accepted-ADR **body** is edited (§D4/D5 of ADR 0042 annotated only via its index
  row + a forward pointer; ADRs are immutable — supersede, do not edit).
- **I-3** — frozen `specs/plans/NNN-*/**` packages (incl. 068's) are never modified.
- **I-4** — the FP filter is **one documented exclusion predicate** (a class), **not** a growing
  catalog and **not** a `NonFixCatalog`.

## 8. Success criteria

- **SC-001 (cleanup)** — the §3 symbols are absent; `make check` green; behaviour-neutral.
  [Verification Intent: unobservable → grep + `make check`]
- **SC-002 (carrier)** — `make dead-code` exits 0 on a clean tree with **no findings reported**;
  with an injected synthetic export it reports it and still exits 0. [Verification Intent:
  unobservable → W1 positive control + W2]
- **SC-003 (advisory, not a gate)** — the `verify:` aggregate line is unchanged. [Verification
  Intent: unobservable → grep]
- **SC-004 (records)** — ADR 0064 indexed; ADR 0042 index row + forward pointer; `techstack.md`
  row; `quality.modelith.yaml` advisory member + re-render; `modelith-check` no drift.
- **SC-005 (no regression)** — `make verify` green; `go test -count=1 ./...` green (E2E **330
  scenarios · 2487 steps — unchanged**); `go.mod`/`go.sum` unchanged.

## 9. Assumptions

- **A1** — the operator's instruction grants the intent to reopen ADR 0042 §D4/§D5 (Rule 11(b)).
- **A2** — the two locked operator decisions hold: **D1** = carrier (ii) — filter the recorded FPs
  so a clean tree prints nothing; **D2** = `config.EffectiveUseTUIPrompt` is **deleted** (its inline
  duplicate in `cli.tuiRequested` is **not** a drop-in — it returns `true` for `-i` without loading
  config) and recorded as a single-owner candidate in the ADR.
- **A3** — the tool pin is `golang.org/x/tools/cmd/deadcode@v0.47.0` (the `go.sum`-indirect
  version; no `go.mod` change).
