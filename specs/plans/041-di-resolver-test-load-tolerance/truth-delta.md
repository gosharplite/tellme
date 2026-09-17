# Truth Delta: 041-di-resolver-test-load-tolerance

**Plan Package**: `specs/plans/041-di-resolver-test-load-tolerance`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md` (Testing & Verification): record the resolver harness's **load-tolerance** change — the positive test is decoupled from a tight hardcoded budget that is not its subject (a generous bound), the bounded test keeps the tight bound + raised ceiling as the falsifiability carrier, and the shim uses a **dominant** PATH. **ADD** a new ADR (`docs/decisions/0010-…`) recording the rule *"a test must not hardcode a tight wall-clock budget that is not its subject; real-time assertions must clear a host-speed margin"* (+ the `docs/decisions/README.md` index row).
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change; the round touches a unit-test fixture only).
> - `/axb-dsl-refine` — **NOOP** (no new/changed CLI interface truth; no new Gherkin or `dsl.md` row — the carrier is the unit test + the recorded rule).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Host test harness** row (Testing & Verification) | Extends the determinism clause: in addition to "no `time.Sleep` for synchronization (ADR-036 parity)", it now records **"no test hardcoding a tight wall-clock budget that is not its subject (ADR 0010)"** — a non-bound test takes a generous **test-local** deadline; a wall-clock **ceiling** clears a host-speed margin while staying below the unbounded-case measurement; a bound-asserting test also asserts **non-vacuity** by the failure's shape (signal-killed child: `ExitCode() == -1`); a PATH-shadowing shim gets a **dominant** `PATH`. | Round 041 FR-006/FR-007; `research.md` D1–D3/D7 — the recorded project rule (the durable home for the round-035/040 flake class). |
| MODIFY | `specs/truth/techstack.md` — **Pure-helper unit tests** row (Testing & Verification) | Appends `plus (round 041)`: the **bounded `gh`-token-resolver harness** in `internal/infrastructure/di` — the **dominant-PATH** `gh` shim (shadow the target, resolve its children; retires the round-032 N1 in-shim PATH restoration), the **generous test-local bound** on the trimming test (decoupled from a tight hardcoded budget that is not its subject), and `TestNewGhTokenResolver_Bounded` as the **sole** carrier of the resolver's boundedness (200 ms bound + an exit-code non-vacuity pin (the child was signal-killed by the deadline: `ExitCode() == -1`) + 2 s ceiling below the ≈3 s unbounded case). | Round 041 FR-001–FR-004; `research.md` D1–D4/D7 — the harness's load tolerance. |
| ADD | `docs/decisions/0010-test-deadline-decoupling.md` (+ the `docs/decisions/README.md` index row) | Records the test-deadline rule (D1 decouple non-bound tests from production constants; D2 ceilings clear a host-speed margin with falsifiability preserved; D3 a bound-under-test asserts non-vacuity; D4 dominant-PATH shims), with the measured ≈14–17× host-speed factor as evidence. | Round 041 Q5/FR-006; `research.md` D7 — a durable, citable home (the round-035 session lesson), not a frozen plan package. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface. This round changes a unit-test fixture + records a test rule. | `contract-authoritative` holds vacuously; `spec.md` FR-008; `research.md` D8. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` record shapes | No persisted-state change: the round touches a Go unit-test fixture and documentation only. | `spec.md` FR-008/FR-009; `research.md` D8. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | No user-facing CLI interface behaviour changes — the carrier is a Go unit test plus the recorded rule; no feature Rule, Example, step, or `DSLRow` is added or changed (the Gherkin/DSL topology audit is unchanged). | `spec.md` A6; `research.md` D8 — the round-020/031 non-BDD-tooling precedent. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0010-test-deadline-decoupling.md` (+ the `docs/decisions/README.md` index row) | Records the four test-deadline rules (decouple · ceiling margin · non-vacuity · dominant PATH) with the measured host-speed evidence; no existing ADR is superseded. | A project-level convention that other artifacts/rounds must be able to cite needs a durable home (`adr-index-consistent`); round 041 Q5. |
