# Session Summary — 2026-09-21

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/beta-niffler/ait-tellme` (`$TELL_ME_HOME`); linux/amd64 host (Go 1.26.6).
**Session mode**: `butler`.
**Branches**: §1 — `073-list-role-headers-and-rendered-body` → **PR [#151](https://github.com/gosharplite/tellme/pull/151)** → merged into `dev` (`d90a79e`), propagated, tagged **`round-073`**. §2 — `074-cli-help-and-version-shorthands` → **PR [#152](https://github.com/gosharplite/tellme/pull/152)** → merged into `dev` (`fd204f4`), propagated, tagged **`round-074`**.
**Status at end of day**: round **074** `074-cli-help-and-version-shorthands` **DELIVERED / FROZEN** (the day's latest) — tellme gains a **`-h`/`--help`** flag (prints the flag list to **`stdout`**, exit **0**; offline, prompt-less, precedence before `--version`) and a **`-v`** shorthand for `--version`; **ADR 0046**; operator request (no anchor issue). *(§1 = round 073, also delivered this day.)*

---

## 1. Session 60 — round 073 `073-list-role-headers-and-rendered-body`: operator request → full pipeline → `architect` review-fold loop (2 passes, CLOSED) → **human-merged (PR #151 → `dev` `d90a79e`, merge commit)** → branch cleanup → closeout (Steps 1–8)

At session start the operator asked, after comparing tellme's `-l` output with the reference's, for `[USER] / [MODEL]` header lines (blue/magenta on a TTY stdout), the body rendered as Markdown by glamour, and a blank line between messages. A new branch **`073-list-role-headers-and-rendered-body`** was created **off `dev`**, the round ran the full AIxBDD pipeline, and **PR [#151](https://github.com/gosharplite/tellme/pull/151)** was taken through the `architect` peer's **review → fold → fold-verification** loop to **CLEARED FOR HUMAN MERGE**, then human-merged (a merge commit); the branch was deleted (local + remote) and `SESSION-CLOSEOUT.md` Steps 1–8 ran.

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | **operator request** (no anchor issue): the offline `-l` listing gains reference parity of **presentation** — `[USER]`/`[MODEL]` role headers, the model body glamour-rendered, one blank line per message, terminal-gated accents |
| Clarify (one at a time) | **Q1 → A** the listing **honours `-r`/`--raw`** (raw ⇒ model body verbatim, no accent) · **Q2 → A** the listing stays **operator-messages-only** (no `[Tool Call]`/`[Tool Response]`; round-007 contract preserved) · **Q3 → B** **only the model body is rendered** (the operator prompt is verbatim — a recorded divergence) |
| Pipeline | specify ✅ · clarify ✅ (Q1–Q3) · spec-by-example ✅ · technical-research ✅ (**ADR 0045** + `techstack.md` ×4 + domain model) · system-analysis ✅ (1 CLI end; api/data NOOP) · dsl-refine ✅ (+5 Rules; +9 DSL rows, 3 rewritten) · tasks ✅ (T001–T018) · implement ✅ |
| The change | a new domain port **`render.Listing`** (`ListingRole`/`ListingMessage`/`ListingSpec`) + a new **`internal/ui/listing.go`** adapter (`ui.NewListing()`), injected via `deps.Dependencies.NewListing`; `renderHistoryList` calls it (`listingMessages` replaces `toMessages`); the accent gate wires tellme's **first `stdout` terminal probe** (`stdoutTerminalDetector()` + `TELL_ME_FORCE_STDOUT_TTY`), scoped to the listing; the width is resolved **best-effort**; one config parse (`offlineConfigAndMode`, fold RF-8) |
| Review chain (PR #151, the `architect` peer) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; **F-1** the rendered-model-body E2E carrier was unfalsifiable · **F-2** a self-contradicting techstack row · **F-3** the ADR Consequences over-claim; plus TD-4…TD-7, RF-8/RF-9, N-1…N-3) → fold `8916807` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**; the F-1 falsifiability reproduced by mutation, 281/282) → residual homing `31caa39` |
| Merge | PR [#151](https://github.com/gosharplite/tellme/pull/151) **human-merged** into `dev` (`d90a79e`, **merge commit**); remote branch deleted by the human, then the **local branch deleted** after an ancestor check |
| Closeout | `make check` **OK** · `make test-race` green · `make verify` **OK** · topology audit **5 pre-existing, none new** (50 features · 404 module rows · 2071 steps) · diff secret scan clean · `STATUS.md` split (round-072 detail + env note → `docs/archives/status/2026-09-21.md`) · propagation `dev → main` (**no-ff**) + tag **`round-073`** · `go install` · **no anchor issue to close** |

### Decisions locked (round 073)

| # | Decision |
| --- | --- |
| **Q1 → A** | The listing **honours `-r`/`--raw`** (reference parity): `-l -r` ⇒ the model body verbatim, no accent; bare `-l` ⇒ rendered + accent on a terminal. Keeps the in-group relay recipe byte-plain. |
| **Q2 → A** | The listing stays **operator-messages-only** (the `[USER]` prompt + the `[MODEL]` answer per exchange); **no** `[Tool Call]`/`[Tool Response]` lines (the round-007 contract preserved). |
| **Q3 → B** | **Only the model answer's body is Markdown-rendered**; the operator prompt body is printed **verbatim** (a recorded divergence — the reference renders both). |
| **D1** (research) | A **dedicated `render.Listing` port** (not a widened `render.Answer`), with the **role a domain enum** so the CLI never hands the stored `user`/`assistant` naming to a presentation seam. |
| **D3** | The accent gate uses a **dedicated `stdout` terminal probe** + the `TELL_ME_FORCE_STDOUT_TTY` seam — tellme's **first stdout probe**, scoped to the listing; the answer path wires none, so **Obs 1 stays OPEN**. |
| **D4** | The colour codes are the reference's: `\033[1;34m` (`[USER]`) / `\033[1;35m` (`[MODEL]`), added to `colour.go` via the shared empty-safe `wrap`. |
| **D6** | The model body reuses the **one** glamour renderer (style / sanitizer / `WRAP_WIDTH` / degraded fallback) — one rendering policy, not two. |
| **D7** | The width is resolved **best-effort** on the `-l` path (an unreadable/invalid `WRAP_WIDTH` degrades to the renderer default; **no new failure mode**). |
| **Accent-scope (recorded)** | The gated accent is the **header** SGR pair only; a rendered model body carries glamour's own style output into a pipe exactly as the answer path does (round 006: rendering is gated by `-r` alone) — so the truth rows assert *no blue/magenta*, and the `-r` row asserts *no `\033` at all*. |

### Commits (branch `073-list-role-headers-and-rendered-body`, then merged)

| Commit | Note |
| --- | --- |
| `9e254a3` | `docs(073)`: plan package + spec |
| `69795c5` | `docs(073)`: fold clarify Q1 → A, Q2 → A, Q3 → B |
| `5ec4486` | `docs(073)`: acceptance Gherkin (listing as a conversation) |
| `db9e683` | `docs(073)`: technical research + ADR 0045 + techstack/domain-model truth |
| `a4b4070` | `docs(073)`: system analysis + dsl-refine truth + tasks |
| `c23f3af` | `feat(073)`: the `-l` listing presents messages as a conversation (role headers, rendered model body, blank separator, stdout-gated accents) |
| `bd12598` | `docs(073)`: truth-delta status + tasks record |
| `2b5483b` | `docs(073)`: STATUS — round 073 in flight (PR #151 open) |
| `8916807` | `fix(073)`: fold the architect review (F-1/F-2/F-3 + TD-4…TD-7, RF-8/RF-9, N-1, N-3) |
| `31caa39` | `docs(073)`: home the fold-verification residuals (RF-073-9/RF-073-10) in ADR 0045 §Forward |
| `d90a79e` | PR [#151](https://github.com/gosharplite/tellme/pull/151) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(073)`: day close — round 073 delivered + propagated; STATUS split + 09/21 summary |

### Artifacts / truth

- Plan package: `spec.md` (US1–US3 · FR-001…FR-009 · NFR-001…NFR-004 · I-1…I-5 · A1–A5 · SC-001…SC-004) · `checklists/requirements.md` · `features/acceptance/listing-the-session-as-a-conversation.feature` · `research.md` (D1–D9) · `plan.md` · `tasks.md` (T001–T018 + the review fold ledger) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` MODIFY ×4 (*CLI flag parsing* · *Terminal detection* · *Output rendering* · *Raw output flag* · *Session lifecycle flags*) · `specs/truth/features/cli/history/inspecting-the-session-history.feature` (+5 Rules) · `history/dsl.md` (+9 rows, 3 rewritten) · `contracts/**` + `data/**` NOOP · `docs/domain-model/tellme.modelith.{yaml,md}` (the `Session` offline-reader wording; the `Chrome` colour invariant scoped to the turn chrome) re-rendered.
- Code: `internal/domain/render/ports.go` · `internal/ui/listing.go` + `internal/ui/colour.go` · `internal/cli/cli.go` · `internal/app/deps/deps.go` · `cmd/tellme/deps.go` · tests `internal/ui/listing_test.go` + `internal/cli/list_render_test.go` + `tests/e2e/steps/step_r073_listing.go` (+ the re-pointed `-l` Thens).
- **ADR 0045** (`docs/decisions/0045-list-role-headers-and-rendered-body.md` + index).

### Falsifiability witnesses (reproduced then reverted)

Drop the separator ⇒ the unit pin + the E2E `Listing the last two messages` RED · render the operator body ⇒ the unit pin + `A listed prompt is echoed verbatim…` RED · un-gate the colour ⇒ the unit pin + `… the listing carries no accents` RED · ignore `-r` ⇒ the unit pin + `The raw listing shows the answer's source` RED. **The `architect` independently re-ran the F-1 witness** (disable the render ⇒ 281/282, the sole failure at the intended Then).

## 2. Open items (non-blocking)

- **Round-073 forward items** — **RF-073-1** render the operator prompt body too (reference parity; not adopted) · **RF-073-2** the empty-session `No history found.` sentence (not adopted) · **RF-073-3** the reference's `-l N "prompt"` list-then-chat composability + the `-b/--back` pairing (not adopted) · **RF-073-4** `[Tool Call]`/`[Tool Response]` listing parity (not adopted; Q2 → A) · **RF-073-5** the answer-path stdout-chrome observation (Obs 1) stays OPEN · **RF-073-6** the best-effort `-l` width resolution · **RF-073-7** the trailing-blank-line policy · **RF-073-8** the `ListingSpec.Warn` field vs the `Answer.WarnDegraded` method shape (review N-2) · **RF-073-9** `historyMode` is a thin wrapper with no production caller (retire/inline at closeout; fold-verification RES-FV-1) · **RF-073-10** the degraded-listing path has no dedicated pin (RES-FV-2). All in **ADR 0045 §Forward**.
- **Compaction**: the round-070 and round-071 forward-item batches are now **compacted to a single pointer row** (both ≥2 rounds old) per the curation rule.
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the pre-existing Gherkin/DSL topology-audit errors (see env notes).

## 3. Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (there is **no live issue**; the agent tool surface is settled).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 073 is fully closed out: PR #151 human-merged into `dev` (`d90a79e`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-073`** with operator approval; the installed binary refreshed from the `dev` head.)*

## 4. PM follow-ups

- None new (spec/acceptance complete; the `architect`'s F-1/F-2/F-3 were record/claim folds, not PM-owned gaps).

---

## 5. Process notes (durable)

- **The fold-verified claim must be *listing*-bound, not *arrangement*-bound** (F-1): an E2E Then that only checks "a marker-bearing answer was *arranged*" is vacuous when the `-l N` window truncates it out. Bind the assertion to the **rendered output** (a listed marker answer), and derive the expected length from the **request** (`-l N`), not the observed output (TD-5).
- **Consistency across colour axes** (F-2/F-3): when a new gate uses a different axis (here stdout, not stderr), any earlier "no X wired" claim must be *reconciled*, not merely appended to; and an ADR's Consequences must match its own Decisions.
- **The `architect` peer's tool inventory** — it exposed `execute_command`, `read_files`, `list_files`, `get_tree`, `search_files`, `write_file`, `list_skills`; one continuation failed on a hallucinated `exec_command` (exit 7). Re-staging the prompt with an explicit tool list + `execute_command` recovered it in one turn.

---

## 2. Session 60, cont. — round 074 `074-cli-help-and-version-shorthands`: operator request → full pipeline → `architect` review-fold loop (2 passes, CLOSED) → **human-merged (PR #152 → `dev` `fd204f4`, merge commit)** → branch cleanup → closeout (Steps 1–8)

The operator continued the same calendar day: after confirming round 073's merge and deleting the local branch, the operator asked *"Does tell-me-go have `-h` flag?"* (answer: yes, cobra adds `-h, --help` and `-v, --version`; tellme had neither shorthand) and then *"Please add simple `-h` and `-v` flags for tellme."* A new branch **`074-cli-help-and-version-shorthands`** was created **off `dev`**, the round ran the full AIxBDD pipeline, and **PR [#152](https://github.com/gosharplite/tellme/pull/152)** was taken through the `architect` peer's **review → fold → fold-verification** loop to **CLEARED FOR HUMAN MERGE**, then human-merged (a merge commit); the branch was deleted (local + remote) and `SESSION-CLOSEOUT.md` Steps 1–8 ran.

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | **operator request** (no anchor issue): tellme gains a **`-h`/`--help`** flag and a **`-v`** version shorthand, matching the reference's flag surface |
| Clarify (one at a time) | **Q1 → A** `-h` ships **with** the long form **`--help`** (reference parity) · **Q2 → A** help writes to **`stdout`** and exits **0** (a successful action, like `--version`) |
| Pipeline | specify ✅ · clarify ✅ (Q1/Q2) · spec-by-example ✅ · technical-research ✅ (**ADR 0046** + `techstack.md` ×3) · system-analysis ✅ (1 CLI end; api/data NOOP; domain model **not modelled**) · dsl-refine ✅ (root cross-module When row; +2 Thens; +2 Rules; a `-v` Example) · tasks ✅ (T001–T010) · implement ✅ |
| Before → after | `-h`/`--help` was pflag's **implicit** path (flag block on **stderr** + the phrase, **exit 2**); now an explicit `-h, --help` prints `Usage of tellme:` + `fs.FlagUsages()` on **stdout**, stderr empty, **exit 0**. `-v` was an unknown flag (exit 2); now `tellme {version}`, exit 0 (same path as `--version`). The unrecognized-flag refusal is unchanged (phrase on stderr, exit 2); a parse error pre-empts help (`-h -z` still refuses) |
| The change | `internal/cli/cli.go` — `parseFlags` registers `-h/--help` explicitly (superseding pflag's implicit path) and upgrades `--version` to `-v/--version`; the flag list is captured at parse time; `run` prints it to `stdout` and returns `Success`, checked **before** `--version`. Local to `internal/cli`; no new port; no new dependency |
| Review chain (PR #152, the `architect` peer) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; **F-1** the precedence-owner truth row stale · **F-2** the durable record mis-described the error path; plus TD-3/4/5, N-1…N-3, RF-1/RF-2) → fold `1185292` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**; the F-2 witness re-run on the folded head) → residual homing `75e794e` |
| Merge | PR [#152](https://github.com/gosharplite/tellme/pull/152) **human-merged** into `dev` (`fd204f4`, **merge commit**); remote branch deleted by the human, then the **local branch deleted** after an ancestor check |
| Closeout | `make check` **OK** · `make test-race` green · `make verify` **OK** · topology audit **5 pre-existing, none new** (51 features · 405 module rows · 2093 steps) · diff secret scan clean · `STATUS.md` split (round-073 detail + env note → `docs/archives/status/2026-09-21.md`) · propagation `dev → main` (**no-ff**) + tag **`round-074`** · `go install` · **no anchor issue to close** |

### Decisions locked (round 074)

| # | Decision |
| --- | --- |
| **Q1 → A** | `-h` ships **with** the long form **`--help`** (`-h, --help`), matching the reference. |
| **Q2 → A** | Help writes to **`stdout`** and exits **0** (a successful, offline, prompt-less action, like `--version`). |
| **D1** (research) | Define both flags **explicitly**, superseding pflag's implicit `-h` special case (which wrote to the `SetOutput` writer = stderr and returned `ErrHelp` → exit 2). |
| **D2** | The help block is pflag's own **flag list** (`"Usage of tellme:\n"` + `fs.FlagUsages()`) — one source, lists every flag, cannot drift. |
| **D3** | Precedence: **`--help` → `--version` → `-d` → `-l` → `-t` → `--tool-usage`**; a **parse error pre-empts help** (`-h -z` refuses, exit 2). |
| **D4** | Help emits **no** `tellme: …` phrase; the unrecognized-flag path (phrase on stderr, exit 2) is unchanged. |
| **D5** | Scope: local to `internal/cli`; **no port** (the bytes are pflag's rendering of a flag set that lives there). |
| **D7** | `-v` is purely additive — `--version` output is byte-identical. |
| **Domain model** | **Not modelled** — the round changes a CLI flag surface, not a modelled entity/invariant; recorded in `plan.md` §5 (ADR 0041's escape hatch). |

### Commits (branch `074-cli-help-and-version-shorthands`, then merged)

| Commit | Note |
| --- | --- |
| `706e27c` | `docs(074)`: plan package + spec |
| `8643ff7` | `docs(074)`: fold clarify Q1 → A, Q2 → A |
| `65eee7c` | `feat(074)`: `-h`/`--help` + `-v` (acceptance + research + ADR 0046 + truth + implementation) |
| `5a310b0` | `docs(074)`: STATUS — round 074 in flight (PR #152 open) |
| `1185292` | `fix(074)`: fold the architect review (F-1/F-2 + TD-3/4/5, N-1…N-3, RF-2) |
| `75e794e` | `docs(074)`: home the fold-verification residuals (RF-074-6/RF-074-7) in ADR 0046 §Forward |
| `fd204f4` | PR [#152](https://github.com/gosharplite/tellme/pull/152) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(074)`: day close — round 074 delivered + propagated; STATUS split + 09/21 summary §2 |

### Artifacts / truth

- Plan package: `spec.md` (US1–US3 · FR-001…FR-008 · NFR-001/NFR-002 · I-1…I-5 · A1–A4 · SC-001…SC-003) · `checklists/requirements.md` · `features/acceptance/asking-for-help-and-version.feature` · `research.md` (D1–D7) · `plan.md` · `tasks.md` (T001–T010 + the review fold ledger) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` MODIFY ×3 (*CLI flag parsing* · *Prompt input* — the precedence owner · *Version injection*) · `specs/truth/features/cli/dsl.md` (the cross-module `runs tellme with "{flag}"` row promoted from `diagnostics`) · `usage/dsl.md` (+2 Thens) · **NEW** `usage/requesting-help.feature` (+2 Rules) · `diagnostics/version-and-setup-diagnostic.feature` (+ a `-v` Example) · `contracts/**` + `data/**` NOOP · `docs/domain-model/**` **unchanged** (not modelled).
- Code: `internal/cli/cli.go` · `internal/cli/help_version_test.go` · `tests/e2e/steps/step_r074_help.go`.
- **ADR 0046** (`docs/decisions/0046-cli-help-and-version-shorthands.md` + index).

### Falsifiability witnesses (reproduced then reverted)

Drop the explicit `-h` flag (pflag's implicit path) ⇒ the E2E `The operator asks for help with "-h"` reddens at `tellme prints its flag list` **and** the unit pins · drop the `-v` shorthand ⇒ `… with the short flag` reddens at `tellme prints the build version` **and** the unit pins. The **`architect` independently re-ran the F-2 witness** on the folded head (`-h` → stdout/empty stderr/0; `-z` → phrase/2; `-v`/`--version` byte-identical; `-h -v` → help).

## 3. Open items (non-blocking)

- **Round-074 forward items** — **RF-074-1** a richer help block (reference `Usage:`/`Flags:` prose; not adopted) · **RF-074-2** the reference's `help`/`completion` subcommands (not adopted; tellme stays subcommand-free) · **RF-074-3** no full-text golden pin for the help block · **RF-074-4** no `-V`/`--Version`/`-?` aliases · **RF-074-5** `helpText` cached on the parse result · **RF-074-6** the *Session lifecycle flags* scoped subset chain (not a contradiction) · **RF-074-7** the `-v`/`--version` Examples carry no `no network access` Then. All in **ADR 0046 §Forward**.
- **Round-073 forward items** — RF-073-1…10 in ADR 0045 §Forward.
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the **5 pre-existing** Gherkin/DSL topology-audit errors.

## 4. Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (there is **no live issue**).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 074 is fully closed out: PR #152 human-merged into `dev` (`fd204f4`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-074`** with operator approval; the installed binary refreshed from the `dev` head.)*

## 5. PM follow-ups

- None new (spec/acceptance complete; the `architect`'s F-1/F-2 were truth/record folds, not PM-owned gaps).

## 6. Process notes (durable)

- **A new terminal action must update its owning truth row, not just its own** (F-1): the round added `--help` to the precedence but left the *Prompt input* row (the precedence owner) stale — one truth file asserted two chains. When a round changes a shared contract, grep the truth for the old shape and MODIFY each owner row; record every one in `truth-delta.md`.
- **Describe the pre-round mechanism exactly** (F-2): the ADR called pflag's implicit-help text "the same text today's error path prints" — the error path prints only the phrase. Name the mechanism (implicit help on the `SetOutput` writer) so a future reader cannot re-route help through the error path.
- **`-h`/`-v` are the reference's cobra conventions**; tellme reproduces the *observable contract* (help → stdout/0) without cobra — a flag, not a subcommand.

---

## 3. Session 61 — round 075 `075-skill-frontmatter-block-scalars`: operator request → full pipeline → `architect` review-fold loop (2 passes, CLOSED) → **human-merged (PR #153 → `dev` `250221d`, merge commit)** → branch cleanup → closeout (Steps 1–8)

An operator session (same calendar day): after inspecting the skills loader, the operator asked tellme itself to be fixed so a skill authored with a YAML **folded block scalar** (`description: >`) is listed by its **real** text instead of the bare indicator `>`. A new branch **`075-skill-frontmatter-block-scalars`** was created **off `dev`**, the round ran the full AIxBDD pipeline, and **PR [#153](https://github.com/gosharplite/tellme/pull/153)** was taken through the `architect` peer's **review → fold → fold-verification** loop to **CLEARED FOR HUMAN MERGE**, then human-merged (a merge commit); the branch was deleted (local + remote) and `SESSION-CLOSEOUT.md` Steps 1–8 ran.

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | **operator request** (no anchor issue): the skills loader resolves a YAML **block scalar** (`>`/`|` + chomping) `name`/`description`, so `list_skills` shows the skill's **real** text, not `>` |
| Clarify (one at a time) | **not escalated (0 questions)** — the goal/scope are unambiguous; the residual choices are technical (stdlib vs a YAML library; folding/chomping; the E2E authoring seam) → `/axb-technical-research` |
| Pipeline | specify ✅ · clarify ✅ (0 Q) · spec-by-example ✅ · technical-research ✅ (**ADR 0047** + `techstack.md` §Skills *Skills catalog (load)* MODIFY) · system-analysis ✅ (1 CLI end; api/data NOOP; dsl-refine truth) · dsl-refine ✅ (+1 Rule/Example; +2 DSL rows) · tasks ✅ (T001–T010) · implement ✅ |
| Before → after | `- golang-patterns: > (/…/SKILL.md)` → `- golang-patterns: Idiomatic Go patterns for robust code (…)`. Inline quoted/unquoted values stay byte-identical; an inline `a > b` stays literal; an indicator with no body ⇒ the file is not a skill (best-effort skip, unchanged). |
| The change | `internal/infrastructure/skills/loader.go` — `parseFrontmatter` routes `name`/`description` through `resolveFrontmatterValue`; `isBlockIndicator` (`[>|][+-]?`) + `collectBlock` + `stripBlockIndent`/`joinBlockScalar`/`applyChomping` resolve the block; **stdlib-only**, no YAML dependency; the `Skill` type + `list_skills` surface/order/bounds unchanged |
| Divergence (recorded) | the reference `tell-me-go`'s line-based `parseSkill` **shares the limitation** — a **deliberate divergence *beyond* the reference** (robustness), not parity (**ADR 0047 §Decision 6**) |
| Review chain (PR #153, the `architect` peer — init once with `SESSION-BOOTSTRAP.md`, continuations) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; **F-1** D6 overclaimed the witness set + FR-005 had no carrier · **F-2** chomping is unobservable (`TrimSpace` neutralizes it) · **F-3** an ADR context claim unverifiable · **N-1** the DSL `不該發生` clause had no carrier; RF-075-4…7 proposed) → fold `7fb0fa1` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**; the new pins proven to redden under a reverted reader) → residual homing `8126470` (R-1…R-4) |
| Merge | PR [#153](https://github.com/gosharplite/tellme/pull/153) **human-merged** into `dev` (`250221d`, **merge commit**); branch deleted (local + remote) |
| Closeout | `make check` **OK** · `make test-race` green · `make verify` **OK** · topology audit **5 pre-existing, none new** (51 features · 407 module rows · 2100 steps) · diff secret scan clean · `STATUS.md` split (round-074 detail + env note → `docs/archives/status/2026-09-21.md`) · propagation `dev → main` (**no-ff**) + tag **`round-075`** · `go install` · **no anchor issue to close** (0 open) |

### Decisions locked (round 075)

| # | Decision |
| --- | --- |
| **D1** | Resolve block scalars in the **existing stdlib reader** (`strings`), not by adopting `gopkg.in/yaml.v3` — the reference's reader is hand-rolled; no new dependency (I-4/NFR-001). |
| **D2** | The indicator grammar is exact: a value is a block scalar **iff** it is `[>|][+-]?`; anything else (incl. `a > b`) stays inline. |
| **D3** | Fold (`>`) / literal (`|`) + chomping; the resolved value is `TrimSpace`d — so trailing-newline chomping is **normalized** (`>`≡`>-`≡`>+` for the trimmed scalar). |
| **D4** | A block with no body ⇒ empty ⇒ the file is **not** a skill (best-effort skip, unchanged). |
| **D5** | The `Skill` value type + the `list_skills` surface are **unchanged** — a reader-robustness fix, one function. |
| **D6** | **Recorded divergence *beyond* the reference** (not parity): `tell-me-go`'s line-based `parseSkill` shares the limitation. |
| **Domain model** | **Not modelled** — parsing, not a modelled entity/invariant; recorded in `plan.md` §5 (ADR 0041's escape hatch). |

### Commits (branch `075-skill-frontmatter-block-scalars`, then merged)

| Commit | Note |
| --- | --- |
| `a61ec45` | `docs(075)`: plan package + spec |
| `b2df835` | `feat(075)`: resolve YAML block-scalar skill frontmatter (acceptance + research + ADR 0047 + truth + implementation) |
| `7fb0fa1` | `fix(075)`: fold the architect review (F-1/F-2/F-3 + N-1; RF-075-4…7 homed) |
| `8126470` | `docs(075)`: home the fold-verification residuals (R-1…R-4) |
| `250221d` | PR [#153](https://github.com/gosharplite/tellme/pull/153) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(075)`: day close — round 075 delivered + propagated; STATUS split + 09/21 summary §3 |

### Artifacts / truth

- Plan package: `spec.md` (US1–US2 · FR-001…FR-006 · NFR-001…NFR-003 · edge cases · SC-001…SC-003) · `checklists/requirements.md` · `features/acceptance/listing-a-block-scalar-skill.feature` · `research.md` (D1–D6) · `plan.md` · `tasks.md` (T001–T010) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` §Skills *Skills catalog (load)* MODIFY (block-scalar support + the divergence note) · `specs/truth/features/cli/chat/listing-the-available-skills.feature` (+1 Rule/Example) · `chat/dsl.md` (+2 rows + a round-075 note) · `contracts/**` + `data/**` NOOP · `docs/domain-model/**` **unchanged** (not modelled).
- Code: `internal/infrastructure/skills/loader.go` · `internal/infrastructure/skills/loader_test.go` (+6 tests) · `tests/e2e/steps/step_r075_skills_given_folded.go` · `tests/e2e/steps/step_r075_skills_then_described.go`.
- **ADR 0047** (`docs/decisions/0047-skill-frontmatter-block-scalars.md` + index).

### Falsifiability witnesses (reproduced then reverted)

Force `isBlockIndicator` false (a line-based read) ⇒ the E2E `A description written as a folded block scalar` reddens at `the listing describes the skill "golang-patterns" as "Idiomatic Go patterns for robust code"` (result `- golang-patterns: > (…)`) **and** the unit pins redden (`folded description = ">"; want "…"`). The **`architect` independently re-ran** the witness (reverted reader on an out-of-tree copy → the new Example reddens at exactly the intended Then; the other 5 Examples stay green).

## Open items (non-blocking)

- **Round-075 forward items** — **RF-075-1** the block-scalar *subset* (full YAML grammar / indentation indicators not adopted) · **RF-075-2** a block-scalar `name` has no E2E carrier (unit-only) · **RF-075-3** the reference is not fixed (recorded divergence) · **RF-075-4** a non-`name`/`description` key's block body is not consumed (pre-existing/shared with the reference) · **RF-075-5** `description: > # note` resolves literal · **RF-075-6** a more-indented block line leaks indentation · **RF-075-7** an over-then-under-indented line is kept. All in **ADR 0047 §Forward**.
- **Round-074 forward items** — RF-074-1…7 in ADR 0046 §Forward. **Round-073 forward items** — RF-073-1…10 in ADR 0045 §Forward.
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the **5 pre-existing** Gherkin/DSL topology-audit errors.

## Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (there is **no live issue**).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 075 is fully closed out: PR #153 human-merged into `dev` (`250221d`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-075`** with operator approval; the installed binary refreshed from the `dev` head.)*

## PM follow-ups

- None new (spec/acceptance complete; the `architect`'s F-1/F-2/F-3 were record/witness folds, not PM-owned gaps).

## Process notes (durable)

- **A witness that cannot fail is not a witness** (F-2): the mandatory `TrimSpace` normalizes trailing chomping, so the chomping branches are inert for the frontmatter scalar — record the normalisation (a comment + the ADR) rather than assert a chomping difference no input can make.
- **Correct the record to the pins that exist** (F-1): `research.md` D6 claimed pins (blank-line folding, quoted, CRLF) that did not exist; an unpinned MUST (FR-005) is a vacuous claim — add the pins or downgrade the claim, and never let the record overstate coverage.
- **A motivating-evidence claim must be verifiable from the tree** (F-3): naming an out-of-repo reflow the reader cannot see invites a round-trip — state what the reader does here instead.
