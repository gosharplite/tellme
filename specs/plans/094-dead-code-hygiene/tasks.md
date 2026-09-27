# Tasks — round 094 `094-dead-code-hygiene`

One task per unit of work. A `[WITNESS]` task pins a non-observable claim; `[X]` marks a verified
task. The round reopens **ADR 0042 §D4/§D5** + retired **`RF-068-1`** under explicit operator intent
(issue #198).

## Setup

- [X] **T001** — Create the round branch `094-dead-code-hygiene` off `dev` `ae300e9`; create the plan
  package `specs/plans/094-dead-code-hygiene/`.

## Foundational

- [X] **T002** — Re-verify each candidate by **grep** (zero callers) and by the tool
  (`golang.org/x/tools/cmd/deadcode@v0.47.0 -test ./...`): 7 items + `cli.noopCallObserver`
  (uninstantiated production type) ⇒ remove; `ui.sharedSource.Suggest` (interface-conformance-only)
  ⇒ filter.

## Remove the genuinely-dead exports (FR-1)

- [X] **T003** — `internal/config/config.go` — delete `EffectiveUseTUIPrompt` (+ its doc comment).
- [X] **T004** — `internal/infrastructure/llm/openai/client.go` — delete `NewWithHTTPClient` (+ doc).
- [X] **T005** — `internal/agent/agentloop.go` — delete the `DefaultToolTimeout` alias (+ doc); fix
  the `tool_contract.go` comment to name `tools.DefaultToolTimeout`.
- [X] **T006** — `internal/infrastructure/di/mcp_factory.go` — delete the `TokenResolver` type alias
  (+ doc).
- [X] **T007** — `internal/cli/composite_observer.go` — delete `noopCallObserver` (type + 2 methods +
  the `var _` conformance).
- [X] **T008** — `tests/e2e/harness/cmd_helper.go` — delete `RunWithStdin` + `RunBinary` (+ doc);
  fix the `step_t014…go` doc comment.
- [X] **T009** — `internal/infrastructure/mcp/mcptest/server.go` — delete `SchemaWithProperty` (+ doc).

## Add the advisory carrier (FR-2/FR-3/FR-4)

- [X] **T010** — `Makefile` — add the `dead-code` target (vanilla `deadcode@v0.47.0`, `-test`, the
  recorded FP filter, exit 0 always, absent-tool hint) + `.PHONY` + the `help` line; the `verify:`
  aggregate line is **unchanged**.

## Records (FR-5/FR-6)

- [X] **T011** — `docs/decisions/0064-dead-code-advisory-carrier.md` (**NEW**) — the advisory decision.
- [X] **T012** — `docs/decisions/README.md` — add the 0064 index row + annotate the 0042 row with a
  forward pointer (0042 body verbatim).
- [X] **T013** — `specs/truth/techstack.md` — the *Dead-code reachability (advisory)* row + the
  *Coverage tooling — DECLINED* pointer.
- [X] **T014** — `docs/domain-model/quality.modelith.yaml` — add `dead-code` as an advisory
  `QualityGate` member; `make modelith-render`; `modelith-check` no drift.

## Witnesses

- [X] **T015** — **[WITNESS] W1** (positive control) — inject a synthetic dead export ⇒ `make
  dead-code` **reports it** and **exits 0**; revert ⇒ clean (empty output).
- [X] **T016** — **[WITNESS] W2** (advisory, not a gate) — the `verify:` aggregate line is
  **unchanged** (grep); `make dead-code` exits **0** with findings.
- [X] **T017** — **[WITNESS] W3** (cleanup behaviour-neutral) — `go build ./...` + `make check` green;
  `gofmt`/`goimports`/`go vet`/`staticcheck`/`golangci-lint` clean; each removed symbol absent (grep);
  `modelith-check` no drift; EC-001 (scrubbed PATH ⇒ hint + exit 0).

## Verification

- [X] **T018** — `make verify` **OK** (incl. `modelith-check`, `verify-adr-index`); `go test -count=1
  ./...` **green** (E2E **330 scenarios · 2487 steps — unchanged**); `make test-race` no data races;
  `go.mod`/`go.sum` unchanged.

## Delivery

- [X] **T019** — Commit + push the round branch; open the round PR (no Copilot review; only a human
  merges).

---

## Claim → Witness ledger

| Claim | Carrier / witness | Kind | Status |
| --- | --- | --- | --- |
| **CLM-001** — each §3 symbol is genuinely dead (zero callers) and removed | grep per symbol (only its definition before; **nothing** after) + `go build ./...` green | mechanism/grep | verified |
| **CLM-002** — `make dead-code` is advisory + never fails | **W1/W2**: clean tree ⇒ exit **0** + empty output; injected synthetic export ⇒ exit **0** + reported; `verify:` aggregate line unchanged (grep) | mechanism | verified |
| **CLM-003** — the carrier runs the **vanilla** pinned tool with `-test` | the target recipe (`deadcode -test ./...`, `DEADCODE_PIN=v0.47.0`); the temp-GOBIN build's `go version -m` shows `path golang.org/x/tools/cmd/deadcode` | mechanism | verified |
| **CLM-004** — the FP filter makes the clean tree print nothing | clean-tree run prints **empty**; the predicate = the interface-conformance-only class (`sharedSource.Suggest` + the unwrap-class) | mechanism | verified |
| **CLM-005** — absent tool ⇒ hint + exit 0 | **EC-001**: a scrubbed-PATH run prints the install hint and exits 0 | mechanism | verified |
| **CLM-006** — the round adds **no** `verify` member + **no** dependency | grep the `verify:` aggregate; `go.mod`/`go.sum` unchanged (`git diff`) | mechanism | verified |
| **CLM-007** — records consistent | `make verify-adr-index` (0064 indexed once/unique); `modelith-check` no drift | mechanical | verified |
| **CLM-008** — no product behaviour change | `go build ./...` + `make check` green; E2E **330 · 2487** unchanged; each removed `.go` symbol is a whole-symbol deletion (no logic touched) | mechanism | verified |

## Fold ledger

_(populated during the `architect` review-fold loop)_
