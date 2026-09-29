# 🤖 tellme — Session Bootstrap

> **Repo**: `github.com/gosharplite/tellme`
> **Folder**: `~/tmp/github/gosharplite/tellme/`
> **Mission**: Disciplined BDD re-creation of `tell-me-go` driven by `aixbdd-en`
> **Workflow**: AIxBDD (Strict PM/RD separation, single truth, Red-Green-Refactor)
> **Companion**: end-of-day procedure is [`SESSION-CLOSEOUT.md`](SESSION-CLOSEOUT.md) — this file is its start-of-session mirror.

---

## ⚡ Mandatory First Reads — Execute In Order

> 👤 **Humans**: Start with [`README.md`](README.md) instead. This section is for AI agents.
>
> ⛔ **Do NOT stop to summarize. Do NOT ask questions. Do NOT acknowledge — just execute.**
> After reading this file, you are a machine consuming a checklist. Run Step 1 immediately.

| Step | Action | Description |
|:---|:---|:---|
| **1** | Read [`README.md`](README.md), then tellme's **three domain models** | Repo overview — vision, BDD methodology, core roles, and CLI workflow roadmap; then, **right after the README**, read tellme's own domain model: [`docs/domain-model/tellme.modelith.md`](docs/domain-model/tellme.modelith.md) (the shipped product) · [`docs/domain-model/quality.modelith.md`](docs/domain-model/quality.modelith.md) (the quality process) · [`docs/domain-model/environment-management.modelith.md`](docs/domain-model/environment-management.modelith.md) (the environment manager). They are **descriptive docs, not truth** — on conflict, `specs/truth/**` wins. Modelith sources: the sibling `*.modelith.yaml`; regenerate with `make modelith-render` |
| **2** | List all pre-load skills | Inventory and inspect all pre-loaded skills in the current session context to establish operational capabilities and governance boundaries |
| **3** | List all agents you can talk to in current shell env | Discover peer agents and personas in the current workspace (`$TELL_ME_HOME/configs/*.yaml`), identify self (`$TELL_ME_MODE`), and map available conversational targets per `tmg-chat-ingroup` |
| **4** | Read [`STATUS.md`](STATUS.md), then align to the active branch | Live session state — current round / active plan package and its pipeline position, the branch model (`main` → `dev` → session branch), decisions locked so far, artifact progress, and open items. Then run `git branch --show-current`; if it is **not** the **Active branch** named in `STATUS.md`, `git checkout` that branch before doing any work, so the round's artifacts are present |
| **5** | Read the **session summary of the last 5 days** | Session continuity — read `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` for the current day and the preceding 4 calendar days to inherit what recent sessions did, decisions locked, artifact progress, and open items. Skip any calendar day with no summary file |

> **The two external reference trees are no longer bootstrap reads.** `tell-me-go` and `aixbdd-en` are now consulted **on demand** (see *On-demand reference repos* below) — `README.md`'s *Essential References* is the pointer. `tellme` is stable; the session inherits its capability model from `tellme`'s **own** domain model (Step 1) and its BDD discipline from the already-loaded skills (Step 2).

Only after Steps 1–5 are complete and results are reported may the agent respond to user tasking.

---

## ⛔ END OF FILE — EXECUTE NOW

**You just finished reading this file. Do not reply. Do not summarize. Do not ask what to do next.**

Immediately return to the step table at the top and execute **Step 1 → Step 2 → Step 3 → Step 4 → Step 5** in order. Report results when Steps 1–5 are complete.

---

## 🗺️ Execution Mapping & On-Demand References

### 0. tellme's own domain model (Step 1 Details)

Right after `README.md`, read **tellme's own** domain model (round 060; **ADR 0030**) — the shipped system, its quality process, and its environment manager:

| File | Focus & Key Insights |
|---|---|
| [`docs/domain-model/tellme.modelith.md`](docs/domain-model/tellme.modelith.md) | **Product** model — entities (`Session`, `Turn`, `Provider`, `Tool`, `ToolCall`, `Context`, `History`, `Skill`, `MCP` server/tool, `Config`, `Persona`, `Chrome`, `PromptInput`, …), their invariants, and execution scenarios |
| [`docs/domain-model/quality.modelith.md`](docs/domain-model/quality.modelith.md) | **Quality process** model — the `QualityPipeline` gates (`make verify`), the E2E contract, the topology audit, ADR governance, and the triage loop. Records tellme's **no-`NonFixCatalog`** divergence |
| [`docs/domain-model/environment-management.modelith.md`](docs/domain-model/environment-management.modelith.md) | **Environment management** model — environments/groups/personas/provisioning/hot-swap; models the **external** Niffler manager (`tellme.sh`) |

- **Descriptive docs, not truth**: on any conflict with [`specs/truth/**`](specs/truth), the **truth wins** and the model is corrected (ADR 0030 §D4).
- **Source vs. rendered**: edit the `*.modelith.yaml`; the `*.modelith.md` is **generated** — never hand-edit. Regenerate with `make modelith-render`; `make modelith-check` (a `make verify` member) fails on drift.

### 1. On-demand reference repos — `tell-me-go` / `aixbdd-en` (NOT bootstrap reads)

`tellme` is stable and development is slowing, so the bootstrap no longer reads the two external reference trees. They remain **on-demand references**: consult one **only when a round actually needs it**, and treat [`README.md`](README.md) → *Essential References* as the pointer. Neither is truth — on any conflict, [`specs/truth/**`](specs/truth) wins.

| Reference | Local path | Consult it when… |
|---|---|---|
| **`tell-me-go`** (capability & architecture) | `~/tmp/github/gosharplite/tell-me-go/` | a round needs the target capability/architecture behaviour — `README.md`, `Makefile`, `docs/domain-model/{tell-me-go,quality}.modelith.md`, `docs/architect/environments/**`, `docs/architect/INTENTIONAL_NON_FIXES.md` |
| **`aixbdd-en`** (BDD discipline) | `~/tmp/github/gosharplite/aixbdd-en/` | a round needs the formal BDD methodology — `domain-model/aixbdd.modelith.md` (entities, `truth-single-owner`, `plan-package-frozen`, `fresh-package-per-round`, `acceptance-coverage`, `dsl-exact-one-match`) and `README.md` (phase pipeline, CLI streamlining) |

The **skills** are already loaded in the session (Step 2 — `list_skills`); read a chosen skill's `SKILL.md` on demand with `read_files`, rather than reading a whole reference tree at bootstrap.

### 2. Pre-loaded Skills Verification (Step 2 Details)

Identify and enumerate all pre-loaded skills injected into the session context.

### 3. In-Group Agent Discovery (Step 3 Details)

Inspect the current shell environment and discover available peer agents using the `tmg-chat-ingroup` protocol:
1. Identify the current workspace root via `$TELL_ME_HOME`.
2. Enumerate all agent configurations in `$TELL_ME_HOME/configs/*.yaml` and extract their corresponding `MODE` names and personas (`PERSON`).
3. Identify the active agent identity via `$TELL_ME_MODE`.
4. Report all peer agents that can be reached (every mode except self; never message your own mode to avoid session self-pollution).

### 4. Session Status Verification (Step 4 Details)

Read the repo-root `STATUS.md` to establish live session state, then align the working branch:
1. Current round / active plan package and its position in the phase pipeline.
2. The **Active branch** and the branch model (`main` → `dev` → session branch).
3. **Align the working tree**: run `git branch --show-current`; if it is not the **Active branch** named in `STATUS.md`, check that branch out before doing any work (otherwise the round's artifacts are absent).
4. Decisions locked so far. (An open / future item lives in its `ADR 00NN §Forward` or a live issue — `STATUS.md` carries no open-items index.)
5. Keep it current: update `STATUS.md` at each pipeline phase gate and whenever a decision is locked.

### 5. Recent Session Summaries (Step 5 Details)

Read the per-day session summaries for the **last 5 days** to inherit recent session context:
1. Locate summaries under `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` (e.g. `docs/session-summary/2026/09/10/session-summary.md`).
2. Read the summary for the **current day and the preceding 4 calendar days** — newest first is fine.
3. Extract, per day: what was done, decisions locked, artifacts produced, commits, and open items; reconcile against `STATUS.md` (they should agree).
4. **Skip any calendar day with no `session-summary.md` file** — do not treat a missing day as an error.
5. Keep it current: maintain the daily summary (`docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md`) alongside `STATUS.md` so the next session inherits accurate state.

### 6. Vendor API specs — on-demand (DeepSeek / Gemini)

[`docs/model-specs.md`](docs/model-specs.md) is a **descriptive, repo-side reference** (not truth) for fetching the **live** DeepSeek and Gemini/Vertex API specs and reproducing a call. Consult it **whenever a provider/API issue appears** — a wire decode, a usage/cost field, the image/media shape, auth, or a transport failure — instead of concluding from memory. Its rule: **test egress → fetch the live spec → reproduce the call** (a spec says what is *allowed*, not what `tellme` *uses*; the defect is as often in tellme's hand-rolled adapter, `${VAR}` config expansion, auth, or the sandbox). It is **not** a mandatory read — an **on-demand** reference, subordinate to `specs/truth/**` on conflict, and it does **not** relax the repo's offline/hermetic **test** posture.

---

## ⚠️ Agent Rules

1. **No Vibe Coding**: Never write speculative code directly. All product code in `tellme` must be driven by an active `PlanPackage` with executable tests.
2. **Strict BDD Pipeline**: Always follow the sequence: Requirements (`spec.md`) → Acceptance Gherkin (`features/acceptance/`) → Truth Delta (`truth-delta.md`) → Executable DSL (`specs/truth/features/**`) → Tasks (`tasks.md`) → TDD Implementation.
3. **Reference, Never Copy Blindly**: `tell-me-go` is the benchmark for capability, behavior, and architecture; `aixbdd-en` is the benchmark for development discipline. Both are **on-demand references**, not bootstrap reads (they are no longer read at session start — see *On-Demand References*). Clean architecture, testability, and determinism take precedence over legacy shortcuts.
4. **Frozen History**: Never modify delivered `specs/plans/NNN-<slug>/` directories. Always create a new package for new iterations or modifications.
5. **Truth Integrity**: Keep `specs/truth/**` as the single source of truth for current system behavior. Record all modifications through `truth-delta.md`.
6. **Skill Awareness**: Verify pre-loaded skills before taking action; follow the specific SOP and invariants defined in each active skill.
7. **In-Group Protocol**: Respect peer agent boundaries and messaging rules defined in `tmg-chat-ingroup` (clear `TELL_ME_MODE`, sequential dispatch, and never message self).
8. **Session Status Discipline**: Read `STATUS.md` at bootstrap (Step 4) and keep it current — update it at every pipeline phase gate and whenever a decision is locked, so the next session inherits accurate state.
9. **Session History Continuity**: Read the **session summaries of the last 5 days** at bootstrap (Step 5) — `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` — to inherit recent context, decisions, artifact progress, and open items before tasking; reconcile them with `STATUS.md` and skip missing days.
10. **Domain Model is Descriptive Docs**: tellme's domain model (`docs/domain-model/**`, Step 1) is **subordinate** to truth — on any conflict, `specs/truth/**` wins and the model is corrected. Edit the `*.modelith.yaml` source only; the `*.modelith.md` is generated (`make modelith-render`), and `make modelith-check` (a `make verify` member) fails on drift (**ADR 0030**). The model is **load-bearing** (**ADR 0041**): a round that changes **modelled behaviour** updates `docs/domain-model/**` **in the same PR** (or records, in its plan package, why the change is not modelled), and the **advisory** `make modelith-drift` (never fails; **not** a `verify` member) surfaces a modeled entity whose concept has left the code.
11. **Disclosures live in their durable home, not on `STATUS.md`**: `STATUS.md` carries **no** open-items index. A *forward item* — a decision *deferred to a trigger* — lives in its **`ADR 00NN §Forward`** (the authority) or a **live issue**; it is a disclosure, **not** tasking. Do **not** re-raise, re-litigate, or propose one as a round theme unless its **trigger has fired**; a **⚠ trigger-gated** item is inert. Pick a round theme from the **roadmap / a live issue**, **never** from an `ADR §Forward` item. **Two addenda (session 52):** (a) **a forward-item cluster must never generate a round theme** — a theme comes from operator value or a live issue, never from "this would clear the backlog"; (b) **a settled / declined decision is never reopened — even to fire a trigger — without explicit operator intent.** **Third addendum (session 55):** when answering "what is next?" / listing candidates, name **only** (i) **live issues** and (ii) **operator-value themes** — **never** an `ADR §Forward` item, **not even labelled "disclosure, not a candidate"**. Putting a disclosure in a candidates list *is* surfacing it (it invites exactly the "what is this?" round-trip the curation rule exists to prevent). If asked about a forward item, answer **from its ADR** and **do not add it to any next-steps list**. *(The `RF-063-10`/`RF-068-1`/`RF-069-2..4` lesson: an un-curated deferral sitting on a bootstrap-read surface becomes a permanent muse — the recurrence, not the item, is the bug.)* **Fourth addendum (2026-09-24):** do **not** use the word **"carried"** as a status on a bootstrap-read surface — it collides with *"carried-forward item"* (a deferral) and reads as open work. Use the quality-model terms instead: an **advisory** check (reported, never fails) vs a **gate** (a `make verify` member / zero-tolerance), and name the invariant (`topology-audit-not-a-gate`) when describing a non-gating check.

12. **Vendor API specs are live-fetchable, not memorized**: when a DeepSeek or Gemini/Vertex issue appears, do **not** conclude from memory — use [`docs/model-specs.md`](docs/model-specs.md) (**§6 reference**): **test egress → fetch the live spec** (the Gemini/Vertex *discovery document*; DeepSeek's HTML docs **+** the OpenAI Chat Completions spec for its deltas) → **reproduce the call** with the real credential (presence-only checks; **never print a secret**). Reading the spec is necessary but **not sufficient** — the defect may be in tellme's hand-rolled adapter, `${VAR}` config expansion, auth, or the sandbox. The reference is **descriptive, subordinate to `specs/truth/**`**, and does **not** change the offline/hermetic **test** posture; any fix still needs an executable witness (ADR 0006).
