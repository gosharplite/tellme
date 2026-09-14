# Truth Delta: 018-post-turn-status-lines

**Plan Package**: `specs/plans/018-post-turn-status-lines`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Post-turn status lines**: add the metrics-line + `╰─⠿ Ready` chrome (CLI Application) and the per-mode `tokens.log` usage log; add the **config-only `MODELS` pricing** row (Configuration); widen the provider **usage** with `cached`/`reasoning` tokens (Reasoning & Provider Transport); extend the round-018 test coverage (unit formatters + fake-provider usage details/multi-call); retire the "post-turn metrics line not introduced" bullet. | Round-018 research Decisions 1–9: reproduce the reference's post-turn status on a hand-written `internal/ui` formatter, dependency-free, with a config-only pricing table and a per-call usage log. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. | Standalone CLI; no OpenAPI/HTTP surface of tellme's own. Round 018 only reads the provider's existing `usage` — it authors no request/response shape. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/data/data-model.dbml` | New `usage_record` table — the per-mode API-call usage log (`output/<mode>/tokens.log`): one JSON record per provider call (cached/prompt/response/thinking/total tokens + cost), append-ordered, summed for the session totals; project Note broadened to three artifacts. | Round 018 persists per-call usage so the `╰─⠿ Ready` session summary (token totals + cost) is exact across resume and resets on `--new`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature` | New feature carrying the post-turn status-line contract as atomic Rules: the metrics line of the request that just completed (M/H/C/Th; `Th` always), the three costs (request/turn/session), accumulation + reset, the session tokens + cache share, an un-priced model → zero, suppression when no usage, and the trailing position. | Round 018 — carry the acceptance journeys into executable interface truth (`FR-001`–`FR-013`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Add **4 Given rows** (usage-with-details; tool-turn-with-usage; config-only pricing; prior usage log), **1 When row** (`--new` + prompt), and **10 Then rows** (metrics presence + values; cost presence + the three relations; session tokens/cache share; zero cost; trailing position). | Round 018 — the post-turn status-line step vocabulary (its users are the `chat` features). |
| MODIFY | `specs/truth/features/cli/dsl.md` | Add **1 root Then row** `the run reports no post-turn status` — the post-turn-absence predicate is used by the `chat`, `diagnostics`, and `history` modules, so it lives at the interface root (single authority). | Round 018 — the negative boundary is cross-module; no new class phrase (still 11). |
| MODIFY | `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | Add `the run reports no post-turn status` to the version Example. | Round 018 — carry "no post-turn status on `--version`". |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | Add `the run reports no post-turn status` to the `-l` listing Example. | Round 018 — carry "no post-turn status on `-l`". |
| MODIFY | `specs/truth/features/cli/history/starting-a-fresh-session.feature` | Add `the run reports no post-turn status` to the prompt-less `--new` Example. | Round 018 — carry "no post-turn status on a prompt-less `--new`". |
