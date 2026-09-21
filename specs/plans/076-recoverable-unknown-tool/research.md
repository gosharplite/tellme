# Technical Research: A recoverable, per-turn-bounded unknown tool name (round 076)

**Plan Package**: `specs/plans/076-recoverable-unknown-tool`
**Truth Root**: `specs/truth`
**Anchor**: [#154](https://github.com/gosharplite/tellme/issues/154)
**Grounded on**: `dev` @ `86d0deb` (2026-09-22)

> Decision-driven research for the tool-loop change. Owner: `/axb-technical-research`. Updates `specs/truth/techstack.md` (the *Agent tool loop* row + the *Tool-usage accounting* parenthetical) and writes **ADR 0048**.

## Problem

An unknown tool name is **terminal** (`internal/agent/agentloop.go:133-136`): `Registry.Lookup` misses → `return ErrIncomplete{"tool %q is not available"}` → the CLI renders the frozen phrase and exits **7**. Observed live (butler / `deepseek-flash` on `misc`): a model asked for the GitHub MCP tool by its bare upstream name `get_me` (the callable wire name is `mcp_github_get_me`); the whole run aborted. A single naming slip kills the turn — while a **real** tool's call-time error is already folded back non-terminally (`agentloop.go:171-176`).

## Decisions

- **D1 — An unknown tool name is a recoverable fold-back (not terminal).** Reuse the existing recoverable pattern: `logResult` + append a `tool`-role message carrying `ToolCallID` + `continue` — exactly the shape the reason-less refusal (`refuseReasonless`, round 056 / ADR 0025 D3) already uses. The unknown call **runs no tool** and records **no `history.Step`/`Signature`** and **no** tool-usage record (nothing executed). *Rejected:* keeping it terminal (the defect); a provider-side retry (wrong layer).
- **D2 — The fold-back message names the unknown tool and a bounded list of the available wire names.** `error: no tool named "<name>"; available tools: <n1>, <n2>, …` (built from `Registry.Tools()` in offer order; `no tools are available` when the registry is empty). A **specific** message is what breaks a fixation on the first retry; a vague one makes it sticky. The list is **not** separately capped — it is bounded only by the **fixed registry size** (today ~9 wire names); the loop's byte clamp `clampBytes` does **not** run on the fold-back path (it applies only to an executed tool's result — corrected per review F-4). *Rejected:* a bare `unknown tool` (sticky) — the reference-ecosystem "did you mean" suggestion machinery is out of scope (#155).
- **D3 — The recoverable path is bounded PER TURN (the cap).** `maxUnknownToolFolds = 3`: after 3 unknown-name fold-backs in one turn, the loop stops folding back and returns `ErrIncomplete` (frozen phrase + exit 7). **Why mandatory:** tellme has **no** repetition/runaway detection (the reference's SHA-256 / tool-repetition guards were deliberately not re-created — verified absent in `internal/`), and the only other bound is `MaxLoops` (`MAX_TOOL_LOOP`, default **1000**; each round is a **paid** provider call). An unbounded fold-back would turn a fast fatal error into an up-to-~1000-call spend. *N = 3* balances "one genuine slip self-corrects" against "a fixation is stopped fast"; the value is a named constant pinned by a unit test, adjustable in one place. *Rejected:* relying on `MaxLoops` alone.
- **D4 — Counter home: a loop-local variable in `Run`.** `Run` is one turn (the loop is constructed per turn via `newloop.go`); a local `unknownFolds := 0` therefore resets every turn with no state to leak, no `TurnState` field, no domain-model change. *Rejected:* a struct field (would need per-turn reset plumbing) or a `TurnState` field (a domain shape change for a runtime-only counter).
- **D5 — Ordering: the unknown-name check stays BEFORE the reason gate.** Today `Lookup` precedes `refuseReasonless`. An unknown call that also lacks a reason is therefore reported **as unknown** (one round trip buys the name correction); since the unknown call never executes, the *no reason, no go* rule (which gates **execution**) is not violated. Single-owned classification, documented. *Rejected:* reason-first (two round trips to learn the name is unknown).
- **D6 — The round-008 FR-010 contract is NARROWED, not removed.** The frozen phrase `the tool request failed` + **exit 7** still cover the genuine incomplete cases: the tool-loop **bound reached**, the per-turn **cap exhausted**, and **`no tools are registered`**. Only the "one unknown name → immediate abort" case changes. Recorded in **ADR 0048** + the truth row.
- **D7 — tellme's own robustness, not reference parity.** Whether the reference aborts on an unknown name is not asserted here; this is a **tellme-side robustness improvement** recorded as such (the truth row notes it as a divergence beyond what the reference is known to do — cf. round 075's framing). No reference behaviour is claimed.
- **D8 — Not modelled (ADR 0041 escape hatch).** The change is an error-handling classification in the loop; it introduces no modelled entity, attribute, relationship, or invariant. `docs/domain-model/**` is **unchanged**; the reason is recorded in `plan.md` §5.

## Candidates considered

| Candidate | Verdict |
| --- | --- |
| Keep terminal (today) | **Rejected** — the defect (#154). |
| Fold back, unbounded | **Rejected** — a fixation can spend up to `MaxLoops` paid calls (D3). |
| Fold back, bounded by a NEW global repetition/runaway detector | **Rejected in scope** — that is a larger, separate decision (the reference's runaway protection). Out of scope; do not solve it here. |
| Message: bare `unknown tool` | **Rejected** — sticky (D2). |
| Message: unknown name + bounded available list | **Adopted** (D2). |
| Suggestion/near-miss machinery | **Deferred** — overlaps #155; not needed to break the loop. |

## Risks / mitigations

- **Fixation spend (D3).** Mitigated by the per-turn cap; `MaxLoops` remains the outer bound.
- **Unpaired calls → a Gemini/Vertex 400 (the round-065 / [#132](https://github.com/gosharplite/tellme/issues/132) class).** Mitigated by appending a result for **every** unknown call in the round (I-1); the terminal cap path `return`s before any request, so no unpaired wire state is ever sent.
- **Two-rule ambiguity (reason × unknown).** Resolved by D5 (single owner: unknown-name first) and pinned by a unit test.

## Verification plan

- **Unit (`internal/agent`):** an unknown name folds back and the run continues; the per-turn cap terminates (`maxUnknownToolFolds+1` provider calls); the reason gate ordering.
- **E2E:** a `misc`-shaped journey (unknown name → read → answer, exit 0) and the cap journey (always unknown → phrase + exit 7).
- **Gates:** `gofmt`/`go vet`/`go build`; `go test -count=1 ./...`; `make verify`; `make test-race`.
