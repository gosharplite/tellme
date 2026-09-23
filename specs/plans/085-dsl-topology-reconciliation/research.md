# Technical research — round 085 `085-dsl-topology-reconciliation`

**Topic**: how to reconcile the CLI truth-tree's DSL topology to 0 audit errors, and the R8d help
text, while changing **no** behaviour.

**Owner**: `axb-technical-research` (no new technology; no new dependency).

---

## D1 — The audit is the oracle; the failure is a row-lookup defect, not behaviour

The `axb-gherkin-and-dsl` **topology audit** (`audit_feature_dsl_topology.py`) merges the interface
root `dsl.md` with the feature's **own** module `dsl.md`, then requires **exactly one** matching row
per step (`dsl-exact-one-match`). It is a **carried check** — not a `make verify` member — so its
findings were recorded, not gated (issue #173).

Grounding: the live run on `dev` @ `2502133` reports **11 errors** (the `STATUS.md` env-note "6" is
stale — the audit *summary* line matches STATUS exactly, only the error count drifted; corrected
this round). The godog E2E is **green** — the Go step definitions are untouched: the rows are the
*executable-truth documentation* layer, so a row-lookup gap never reddens the suite.

## D2 — Class A (5): promote cross-module rows to the interface root

Cause: a `history` feature step matches a row authored in a **different** module (`chat` /
`configuration`). The merged lookup (root ∪ own module) therefore sees **zero** rows.

Options:
- **(a) Duplicate the row** into the `history` DSL — **rejected**: violates `dsl-single-authority`
  (one authoritative row; a duplicate is an ambiguity error, the audit's own duplicate rule).
- **(b) Re-phrase the history step** to a `history`-scoped row — **rejected**: changes the E2E step
  text (an I-2 violation) and the assertions are genuinely cross-module.
- **(c) Promote the row to the interface root** — **adopted**: the root file's header already
  declares it the home for "cross-module rows — sentence patterns used identically by two or more CLI
  modules". The `DSL.move_row` action (semantics unchanged) is exactly this. Root rows are visible to
  **every** module, so both the owning module and the second consumer resolve.

The four rows (each used by ≥ 2 modules, so no new single-module-root warning):

| Phrase | Was | Modules |
| --- | --- | --- |
| `a configured provider "{provider}" whose endpoint answers with "{answer}"` | `chat` | chat + history |
| `the operator starts tellme with the prompt "{prompt}"` | `chat` | chat + history |
| `the effective mode is "{mode}"` | `configuration` | configuration + history |
| `tellme exits with the configuration error code` | `configuration` | configuration + history |

## D3 — Class B (5): restore the round-083 table structure

Cause: the round-083 rows in `chat/dsl.md` were appended after a prose blockquote **without** a
`## Given/Then (round 083)` heading and **without** the `| DSL 句型 | … |` header + separator rows.
The audit's `parse_dsl` only collects rows inside a table that begins with the `DSL 句型` header, so
those five rows are never read — they exist in the file but are invisible to the checker.

Fix: add the headings + header rows (structure only; the row text is already correct). Convention:
each round's rows carry their own `## Given/Then (round NNN)` heading + the `DSL 句型` header, with the
round note **after** the table (the round-071…080 precedent).

## D4 — Class C (1): add the missing composite row

`chat/colouring-the-session-chrome.feature:14` uses
`a configured provider "{provider}" whose endpoint asks tellme to read "{path}" with the reason
"{reason}" and then answers with "{answer}" and reports the token usage:` — a **composite** (a
`read_files` tool call **carrying a reason**, then an answer **carrying a usage block**). No existing
row carries all three clauses at once (the `… read "{path}" with the reason "{reason}" …` row has no
usage tail; the `… read "{path}" … reports the token usage:` row has no reason clause). The round
(046/027) chrome-colouring scenario legitimately needs the composite, so **add** it (semantics: the
fake returns a `read_files` tool call with `reason` on the first request, then answers carrying the
JSON `usage` block).

## D5 — #174 (R8d): the help text

The `make help` line (`Makefile:104`) still describes the **round-001 whole-binary build-graph
capability guard** (`build-graph capability guard: no net/net/http in ./cmd/tellme closure`), which
round 004 **retired** (the prompt-bearing chat path legitimately links `net/http`). The shipped target
is the **offline-path no-contact witness** (recording sink + differential, `TestOfflinePathsDoNotContactProvider`).
Fix the help line + the target's header comment; retire the `techstack.md` R8d bullet; annotate the
ADR 0012 **index row** — the declared immutability carve-out; the §Forward body is left verbatim (so ADR 0012 is byte-identical to `dev`). Docs/comment-only.

## D6 — No ADR owed

The round makes **no new durable decision**: the promotion is forced by `dsl-single-authority` (the
only legal option), the table-structure convention is already the audit's parser contract + the skill
STANDARDS, and R8d is a record tidy-up + a comment fix. Recorded in this `research.md` + the
`truth-delta.md` (the durable homes within the round); **no ADR 0057** is written (an ADR would record
a rule that already exists).

## D7 — The witness (falsifiability)

The **audit** is the falsifiable witness: it reports **11 → 0** errors. Reproduced pre-fix (11) and
re-run post-fix (0). The E2E is the **behaviour guard**: scenario/step counts unchanged and green
(proving the reconciliation is documentation-only, I-2).

## D8 — Scope guard

No product code, no `.feature`, no step definition, no `go.mod`/`go.sum`, no new dependency, no
`make verify` change.
