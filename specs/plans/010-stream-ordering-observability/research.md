# Phase 0 Research: tellme Stream-Ordering Observability (Round 010)

Topic: make **cross-stream output ordering** — the interleave of the diagnostic stream (`stderr`) with the answer stream (`stdout`) — a **first-class, checkable contract**. Round 009 shipped an ordering defect (the post-turn payload line printed *before* the answer) that **every gate was structurally unable to see**: no layer could *represent* "a `stderr` line trails the `stdout` answer". This round (i) pins the required orderings, (ii) adds the **missing oracle** — a **merged-stream witness** in the E2E harness — and (iii) asserts the orderings at **both** the unit and E2E layers. It changes **no** observable output *content* and adds **no** dependency.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution), testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), output rendering (glamour), session history (append-only JSON-Lines), the agent tool loop (round 008), and the payload status line (round 009) were locked in rounds 001–009. The system still has **one CLI end**, adds **no new system end**, **no new external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. Clarify Round 1 (in `spec.md`) already locked the round's three high-impact decisions: Q1 scope = payload-status **+ tool-loop** ordering (degrade warning incidental); Q2 witness = **merged single-buffer capture**; Q3 = assert at **both** layers. **IN this round**: the witness, the ordering contract, and the layered assertions. **OUT**: any change to the observable output content — the payload-status order is already correct after `7bcb2d3`.

---

## Decision 1: The E2E witness — a merged single-buffer capture (the `2>&1` equivalent)

- **Decision**: Add a harness run variant that points the child's `stdout` and `stderr` at the **same** `io.Writer` (a single `*bytes.Buffer`), so the two streams are merged into one **ordered** stream — the exact view a real terminal shows and the one that exposed round 009's defect. `os/exec` serializes writes to a shared, comparable writer, so the merged order reflects the child's write order; tellme writes sequentially within a turn, so the order is deterministic.
- **Rationale**: The harness today captures `stdout` and `stderr` into two **separate** buffers (`harness.RunResult{Stdout, Stderr}`), which structurally cannot represent an interleaved write — that is the missing oracle. A merged buffer needs **no** product change, **no** new dependency, and **no** pty.
- **Alternatives considered**:
  - **Shell `2>&1`** — the harness invokes `os/exec` directly (no shell intervenes) — rejected.
  - **A timestamped tee / write-recorder with stream labels** — preserves stream identity but adds a process/seam and a nondeterminism risk — rejected.
  - **A pty harness** — a new dependency the project forswore (round 005 grill Q6; reaffirmed round 006 Q2) — rejected.

## Decision 2: Ordering is pinned as executable interface truth, not prose

- **Decision**: The required ordering becomes an explicit, assertable facet of the CLI interface truth (`specs/truth/features/cli/chat/**` — DSL rows plus the feature assertions), carried by an interface Rule — **not** left as acceptance prose as in round 009.
- **Rationale**: Round 009's ordering intent existed only as acceptance prose and was dropped from the executable DSL (an `acceptance-coverage` leak — the DSL row pinned *presence* on `stderr`, never the ordering). Pinning it in the executable contract is what makes it falsifiable.
- **Alternatives considered**:
  - **Keep it in prose** — the status quo that failed — rejected.
  - **Rely on the unit test alone** — leaves the E2E blind spot open — rejected.

## Decision 3: Which orderings are required vs incidental

- **Decision**: **Required** — (a) the payload status brackets the answer (`pre-flight < answer < measured`); (b) the tool-loop log precedes the answer. **Incidental** (not asserted, documented as not guaranteed) — the round-006 degraded-render warning relative to the answer.
- **Rationale**: (a) and (b) are tellme-owned, deliberate `stderr` diagnostics that must bracket the answer they describe — the exact bug class. The degrade warning is a startup fallback whose micro-ordering relative to the answer the contract does not own; asserting it would be brittle.
- **Alternatives considered**:
  - **Payload-status only** — leaves the sibling tool-loop interleave unobserved (the same class) — rejected.
  - **Assert every interleave** — over-constrains an incidental fallback — rejected.

## Decision 4: Assert at both layers (unit + E2E)

- **Decision**: Assert each required ordering at the **unit** layer (the turn's emit order — extending round-009's `TestRunTurn_PostTurnStatusFollowsAnswer`) **and** the **E2E** merged-stream layer (Decision 1).
- **Rationale**: Layered defence — the unit layer is cheap and deterministic and pins the emit order; the E2E layer is the true black-box oracle that round 009 lacked.
- **Alternatives considered**:
  - **E2E only** — drops the cheap deterministic floor — rejected.
  - **Unit only** — leaves the E2E blind spot the round exists to close — rejected.

## Decision 5: No product change — the round is contract + oracle

- **Decision**: The product (`internal/cli` emit order) is already correct after `7bcb2d3`; round 010 introduces **no** output-content change. The work is (i) pinning the contract in the CLI truth, (ii) the merged-stream witness, (iii) the layered assertions.
- **Rationale**: `7bcb2d3` reordered `runTurn` (write the answer, then emit the post-turn line); the round makes that ordering a **guarantee** rather than an accident.
- **Alternatives considered**:
  - **Re-implement an ordering mechanism in the product** — unnecessary; the writes are already ordered — rejected.
  - **Move content between streams** — forbidden; `stdout` must stay byte-exact — rejected.

## Decision 6: Determinism — no sleeps; shared-buffer serialization

- **Decision**: The merged witness is deterministic without `time.Sleep`; the harness keeps its existing hermeticity (unset ambient `TELL_ME_*` / `MAX_*` overrides) and bounded runs.
- **Rationale**: Sequential writes plus a single shared writer yield a stable order; ADR-036 discipline forbids `time.Sleep` for synchronization.
- **Alternatives considered**:
  - **Timestamped writes with sleeps** — nondeterministic and forbidden — rejected.

## Decision 7: No new dependency — stdlib-only witness

- **Decision**: The witness uses Go stdlib `os/exec` only; `go.mod` / `go.sum` are untouched.
- **Rationale**: Merging two streams into one buffer is a plain `exec.Cmd` field assignment — nothing beyond the standard library.
- **Alternatives considered**:
  - **A `2>&1` helper shell / a pty library** — needless dependency for a test-only mechanism — rejected.

## Decision 8: Testing & BDD techstack unchanged (must-ask answers settled)

- **Decision**: No new system end and no change to the BDD techstack or strategy. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are not re-decided. The round reuses the round-008 fake (which already serves a tool-call then an answer) and the round-009 fake (usage / no-usage), and adds the merged-stream capture (Decision 1) plus the unit emit-order assertion (Decision 4).
- **Rationale**: The round adds no dependency and reuses the existing fake; only the harness gains a capture variant and a unit assertion.
- **Alternatives considered**:
  - **A new runner/framework** — a game changer with no need — rejected.

---

## Residual risks / forward links

- **Stream identity under a merged capture**: the merged buffer loses which fd a chunk came from, so assertions key on **content** (the DSL's status-line / tool-log patterns), not on a stream label. This is intentional — the merged view is the contract view.
- **Cross-fd ordering scope**: a merged order is guaranteed only for tellme's **own sequential writes**; if a stream were buffered elsewhere (e.g. a shell), the observed interleave could differ. The E2E harness invokes the binary **directly** (`os/exec`), so no intervening buffer exists.
- **No-final-answer turns**: the ordering contract scopes to "the answer" when present; a turn that ends without an answer (tool-loop bound hit / tool error) asserts only that the tool-loop lines precede any terminal outcome — the "measured after the answer" requirement is vacuous there (`/axb-dsl-refine` wording).
- **Incidental interleave**: the round-006 degraded-render warning is documented as **not guaranteed** and is **not** asserted (Decision 3).
- **Class-phrase contract**: the ordering rows must not introduce the reserved `tellme: ` prefix — the frozen class-phrase vocabulary is unchanged; `/axb-dsl-refine` publishes the ordering semantics as CLI truth rows.
- **Deferred, still out of scope**: streaming, pinning, `-b`/`--retry`, token-budget pruning, MCP, memory, and any write/shell tools.
