# Round 085 — `085-dsl-topology-reconciliation`

**Theme**: reconcile the CLI truth-tree **DSL topology** to **0 audit errors** (anchor issue
[#173](https://github.com/gosharplite/tellme/issues/173)) **and** correct the stale `make help`
`verify-no-network` text (issue [#174](https://github.com/gosharplite/tellme/issues/174), the
round-032 **R8d** tidy-up). **Truth-only + one Makefile string** — no product code, no behaviour
change.

**Anchor issues**: [#173](https://github.com/gosharplite/tellme/issues/173) ·
[#174](https://github.com/gosharplite/tellme/issues/174). **DoD = close both.**

---

## 1. Why this round

The `axb-gherkin-and-dsl` **topology audit** (`audit_feature_dsl_topology.py --root
specs/truth/features/cli`) enforces the AIxBDD truth-tree invariants: every feature nests one module
deep, each module owns a `dsl.md`, each phrase has **one** authoritative row (`dsl-single-authority`),
and **every Gherkin step matches exactly one row** in the merged interface-root ∪ module lookup
(`dsl-exact-one-match`: zero is a gap, more than one is an ambiguity).

The audit is a **carried check** (not a `make verify` member). It currently reports **11 errors** —
re-classified at round open (the `STATUS.md` env-note figure "6" was stale; corrected in this
round). None are ambiguity — all are **zero-match**, from two real defects plus one genuine gap:

| Class | Count | Cause |
| --- | --- | --- |
| **A — cross-module row scope** | 5 | a `history` feature step matches a row authored in `chat`/`configuration`; the merged lookup (root ∪ own module) sees zero rows |
| **B — unparsed table** | 5 | the round-083 rows in `chat/dsl.md` were appended **with no `## Given/Then (round 083)` heading and no `DSL 句型` header row**, so the audit's parser never reads them |
| **C — missing composite row** | 1 | `chat/colouring-the-session-chrome.feature:14` uses a composite `… read "{path}" with the reason "{reason}" … and reports the token usage:` Given that no row matches |

The audit failing is a **DSL-contract documentation defect**, not a behaviour defect: the godog E2E
still runs green (the Go step definitions are unchanged) — the rows are the *executable-truth
documentation* layer.

## 2. The change

1. **Class A → promote** the four shared rows to the **interface root** (`specs/truth/features/cli/dsl.md`),
   removing them from their owning module (`dsl-single-authority` — one authoritative row, now visible
   to every module; the file header already declares the root as the home for cross-module rows):
   - `a configured provider "{provider}" whose endpoint answers with "{answer}"` (was `chat/dsl.md`)
   - `the operator starts tellme with the prompt "{prompt}"` (was `chat/dsl.md`)
   - `the effective mode is "{mode}"` (was `configuration/dsl.md`)
   - `tellme exits with the configuration error code` (was `configuration/dsl.md`)
2. **Class B → restore** the round-083 block's structure in `chat/dsl.md`: add the
   `## Given (round 083)` / `## Then (round 083)` headings and the `| DSL 句型 | … |` header +
   separator rows so the five existing rows are parsed.
3. **Class C → add** the missing composite row to `chat/dsl.md`
   (`a configured provider "{provider}" whose endpoint asks tellme to read "{path}" with the reason
   "{reason}" and then answers with "{answer}" and reports the token usage:` + the
   `prompt/cached/completion/thinking` DataTable).
4. **#174 → correct** the `Makefile` help line (`build-graph capability guard: no net/net/http in
   ./cmd/tellme closure` → the shipped offline-path witness) and the stale header comment above the
   `verify-no-network` target; retire the `techstack.md` `Not Introduced Yet` **R8d** bullet and
   annotate the ADR 0012 §Forward item as resolved.

## 3. Requirements

- **FR-001** After the round, `audit_feature_dsl_topology.py --root specs/truth/features/cli` MUST
  report **0 errors** (summary: 53 features · 6 modules · 17 root rows · 455 module rows · 2328 steps).
- **FR-002** The four shared rows MUST each have **exactly one** authoritative home (the interface
  root) and MUST NOT be duplicated in a module DSL.
- **FR-003** The round-083 rows MUST be parseable — each round block in `chat/dsl.md` carries its own
  `## Given/Then (round NNN)` heading + `DSL 句型` header.
- **FR-004** The composite projection Given MUST have a row carrying the `reason` clause **and** the
  `… and reports the token usage:` tail + the `prompt/cached/completion/thinking` DataTable.
- **FR-005** The `make help` line for `verify-no-network` MUST describe the shipped target (the
  offline-path witness) — never the retired whole-binary build-graph capability guard.
- **FR-006** The `techstack.md` R8d `Not Introduced Yet` bullet MUST be retired (the tidy-up is done).

## 4. Invariants

- **I-1 — No product code / no behaviour change.** Round 085 touches only `specs/plans/**`,
  `specs/truth/features/cli/**/dsl.md`, `specs/truth/techstack.md`, `Makefile` (a help string + a
  comment), an ADR §Forward annotation, and the governance docs (`STATUS.md` / the day summary). No
  `.go` file, no `.feature` file (the step text is unchanged), no `go.mod`/`go.sum`.
- **I-2 — E2E unchanged and green.** The `.feature` files and the Go step definitions are untouched,
  so `go test -count=1 ./...` stays green with the same scenario/step counts (2328 steps).
- **I-3 — `dsl-exact-one-match` holds.** Every step still matches exactly one row; no phrase is
  duplicated across levels.
- **I-4 — `make verify` unaffected** (no gate added, removed, or re-scoped).
- **I-5 — stdlib-only / POSIX-only / hermetic** (trivially: no code, no dependency).

## 5. Scope

**In**: the DSL-row reconciliation (promote / restore / add); the `Makefile` help text; the R8d
truth/ADR records.
**Out**: any product-code change; any `.feature` or step-definition change; new ADRs (no new durable
decision — the round applies the documented `dsl-single-authority` + module-topology rules); the
`## Not Introduced Yet` items other than R8d.

## 6. Success criteria

- **SC-001** The topology audit exits **0** (0 errors).
- **SC-002** The audit's warnings do not regress (no promoted root row is used by a single module).
- **SC-003** `make check` (verify + test) is green; E2E scenario/step counts unchanged.
- **SC-004** `make help | grep verify-no-network` prints the offline-path-witness wording.

## 7. Assumptions

- **A1** The `STATUS.md` env-note "6 errors" is a stale figure; the live count is **11** (corrected
  this round; the audit *summary* line matches STATUS exactly — only the error count drifted).
- **A2** No `NEEDS CLARIFICATION`: the audit output + the two issues lock the goal and the fix
  (promotion is the only `dsl-single-authority`-legal option; duplication is forbidden).
