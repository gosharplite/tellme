# tellme — Status

**Last updated**: 2026-09-16 (day close, session 15: **round 031 `031-tool-schema-wellformedness` DELIVERED / FROZEN** — PR [#65](https://github.com/gosharplite/tellme/pull/65) merged by `thptcnec` (frozen head `06d574f` → `dev` `a7de7b7`)). The round fixed the tool-schema `required ⊆ properties` defect ([#64](https://github.com/gosharplite/tellme/issues/64)) that made a strict provider (Vertex/Gemini) 400 **every** request. Round-030 detail lives in the archive (Rule 12); rounds 001–030 live in the archives. Open issues: **#60** (dogfooding-enablement umbrella) · **#13** (coverage tooling).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (all rounds through 031 delivered; the next round `032-*` starts off `dev`)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–030).

## Round 031 — `031-tool-schema-wellformedness` (DELIVERED / FROZEN)

**Status**: ✅ **DELIVERED / FROZEN** — branch `031-tool-schema-wellformedness` (off `dev`); PR [#65](https://github.com/gosharplite/tellme/pull/65) carried through an **Architectural Review** of the plan+truth half (**APPROVED** + **5 folds** ARCH-1…ARCH-5) → **FOLD ACCEPTED** → residue nit folded → **review loop CLOSED** → implementation review **APPROVED** (one required truth fold + a vacuous-gate guard) → **principal-architect review CERTIFIED READY TO MERGE** → **merged** by `thptcnec` (2026-09-16T01:52:05Z) as `dev` `a7de7b7`; frozen head **`06d574f`** (10 commits). **Resolves issue [#64](https://github.com/gosharplite/tellme/issues/64)** (now closed). A **tool-schema correctness + verification-gate** round — the durable value is the gate that would have caught the defect.

**Scope**: the shared `resourceSchema` builder now declares the mandatory **`reason` property**, so all 5 builder-backed tools (`list_files`, `read_files`, `get_tree`, `write_file`, `replace_text`) satisfy `required ⊆ properties`; `execute_command` (inline, already compliant) is untouched. A new hermetic gate (`TestAgentToolSchemasAreWellFormed`) iterates the **non-overridable production assembler `agentTools()`** (never the `newToolRegistry` DI seam — ARCH-1) and asserts the invariant for **every** tool, plus edge-case pins (zero-`required` vacuous pass; non-object/unparseable fails), an assembler↔registry name pin, and a vacuous-gate guard. **No** new dependency (stdlib-only); the offered set (six tools), class-phrase vocabulary, exit codes, and `stdout`/`stderr` are unchanged.

**Pipeline**: specify ✅ · spec-by-example (skipped — no new journey) · research ✅ · system-analysis ✅ (0 interfaces) · api/data/dsl-refine (NOOP) · tasks ✅ (T001–T007) · implement ✅ · **merged ✅** · **propagated ✅**.

**Decisions locked (round 031)**: **Q1 → 1** (hermetic **tool-schema well-formedness gate** + a **manual** live Vertex/Gemini closeout check; keeps `make verify` offline). **Q2 → 1** (minimal fix — the shared `resourceSchema` declares the `reason` property; `execute_command` left inline; plus a for-every-registered-tool `required ⊆ properties` gate). **ARCH-1…ARCH-5** (gate reads `agentTools()`; flat-schema precondition + `#60` recursive-walk forward item; T002 not a second hand-enumerated six; STATUS reconciled; acceptance-carrier divergence recorded).

**Review trail (PR #65)**: plan+truth **APPROVED** with ARCH-1…ARCH-5 folds (`e10b131`) → **FOLD ACCEPTED** → residue nit folded (`4a29cde`) → **review loop CLOSED** → implementation `989dd20` → **impl review APPROVED** (required fold: the **Agent tool-schema gate** truth row + bullet must name `agentTools()`, not the registry var — `truth-current`; + a vacuous-gate guard) → fold `c1190e9` → plan-side residual folded (`06d574f`) → **principal review CERTIFIED READY TO MERGE** (`06d574f`) → fold-review **no known residual**.

**Commits**: `0180f9c` (spec) · `cec72db` (research + techstack) · `f3dc75d` (system-analysis) · `c7efcfe` (tasks) · `e10b131` (plan+truth review fold) · `4a29cde` (fold-review nit) · `989dd20` (implementation T001–T007) · `c1190e9` (impl-review fold) · `7790bea` (STATUS note) · `06d574f` (impl-fold-review residual) → PR [#65](https://github.com/gosharplite/tellme/pull/65) merge into `dev` (`a7de7b7`).

**Propagation**: ✅ **DONE (no-ff)** — `031-tool-schema-wellformedness → dev` (PR #65, `a7de7b7`) `→ main`.

---

## Delivered rounds (index)

| Round | Branch | Delivered via |
| --- | --- | --- |
| 001 | `001-cli-bootstrap-and-config` | PR [#6](https://github.com/gosharplite/tellme/pull/6) |
| 002 | `002-followup-cleanups` | PR [#7](https://github.com/gosharplite/tellme/pull/7) |
| 003 | `003-provider-registry-completeness` | PR [#11](https://github.com/gosharplite/tellme/pull/11) |
| 004 | `004-first-reasoning-turn` | PR [#12](https://github.com/gosharplite/tellme/pull/12) |
| 005 | `005-stdin-piping` | PR [#16](https://github.com/gosharplite/tellme/pull/16) |
| 006 | `006-rendered-output-and-raw-flag` | PR [#19](https://github.com/gosharplite/tellme/pull/19) |
| 007 | `007-session-history-persistence` | PRs [#20](https://github.com/gosharplite/tellme/pull/20)/[#21](https://github.com/gosharplite/tellme/pull/21) |
| 008 | `008-agent-tools-and-tool-call-loop` | PRs [#24](https://github.com/gosharplite/tellme/pull/24)/[#25](https://github.com/gosharplite/tellme/pull/25) |
| 009 | `009-payload-status-line` | PRs [#26](https://github.com/gosharplite/tellme/pull/26)/[#27](https://github.com/gosharplite/tellme/pull/27) |
| 010 | `010-stream-ordering-observability` | PR [#29](https://github.com/gosharplite/tellme/pull/29) |
| 011 | `011-persona-and-payload-estimate` | PR [#30](https://github.com/gosharplite/tellme/pull/30) |
| 012 | `012-interactive-multiline-prompt` | PR [#31](https://github.com/gosharplite/tellme/pull/31) |
| 013 | `013-vertex-gemini-provider` | PR [#33](https://github.com/gosharplite/tellme/pull/33) |
| 014 | `014-session-replay-fidelity` | PR [#35](https://github.com/gosharplite/tellme/pull/35) |
| 015 | `015-interactive-tui-prompt` | PR [#38](https://github.com/gosharplite/tellme/pull/38) |
| 016 | `016-interactive-prompt-visual-parity` | PR [#40](https://github.com/gosharplite/tellme/pull/40) |
| 017 | `017-turn-chrome-parity` | PR [#41](https://github.com/gosharplite/tellme/pull/41) |
| 018 | `018-post-turn-status-lines` | PR [#43](https://github.com/gosharplite/tellme/pull/43) |
| 019 | `019-turn-spinner` | PR [#44](https://github.com/gosharplite/tellme/pull/44) |
| 020 | `020-cross-compile-gate` | PR [#46](https://github.com/gosharplite/tellme/pull/46) |
| 021 | `021-tool-surface-parity` | PR [#48](https://github.com/gosharplite/tellme/pull/48) |
| 022 | `022-tool-loop-log-line` | PR [#50](https://github.com/gosharplite/tellme/pull/50) |
| 023 | `023-interactive-prompt-teardown` | PR [#51](https://github.com/gosharplite/tellme/pull/51) |
| 024 | `024-tool-resource-contract-and-execute-command` | PR [#54](https://github.com/gosharplite/tellme/pull/54) |
| 025 | `025-spinner-width-safety` | PR [#56](https://github.com/gosharplite/tellme/pull/56) |
| 026 | `026-tool-usage-accounting` | PR [#57](https://github.com/gosharplite/tellme/pull/57) |
| 027 | `027-ai-call-turn-counter` | PR [#58](https://github.com/gosharplite/tellme/pull/58) |
| 028 | `028-user-global-prompt-log` | PR [#59](https://github.com/gosharplite/tellme/pull/59) |
| 029 | `029-agent-write-tools` | PR [#61](https://github.com/gosharplite/tellme/pull/61) |
| 030 | `030-provider-truncation-guard` | PR [#63](https://github.com/gosharplite/tellme/pull/63) |
| 031 | `031-tool-schema-wellformedness` | PR [#65](https://github.com/gosharplite/tellme/pull/65) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–030 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); **031 is the most recent delivered round**.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `031-tool-schema-wellformedness` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history. |
| (next) `032-*` | not started | The next round's working branch off `dev`. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026/027/028/029/030/031 — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **030** ([#62](https://github.com/gosharplite/tellme/issues/62)) = the provider-transport truncation guard; **031** ([#64](https://github.com/gosharplite/tellme/issues/64)) = the tool-schema well-formedness fix + gate.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–031** | — | Provider-registry completeness → … → the tool-schema well-formedness fix + gate. | ✅ **Delivered** (see the delivered-rounds index) |
| **future slices (candidates)** | [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)) skills/context rounds — incl. the round-031 typed-schema + recursive-walk forward items; **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)); plus the carried forward items below. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Issue tracker (2026-09-16 closeout, session 15)** — **[#64](https://github.com/gosharplite/tellme/issues/64) CLOSED (completed)** — delivered by round 031 (PR [#65](https://github.com/gosharplite/tellme/pull/65) merged `a7de7b7`); **[#60](https://github.com/gosharplite/tellme/issues/60)** open (dogfooding-enablement umbrella); **[#13](https://github.com/gosharplite/tellme/issues/13)** open (coverage tooling; still accurate). `#47` `not_planned`, `#53`/`#55` completed (earlier); `#62` closed by round 030.
- **Round-031 forward items (recorded on [#60](https://github.com/gosharplite/tellme/issues/60))** — (a) **typed schema construction** (replace the `fmt.Sprintf` string interpolation in `resourceSchema` with a typed struct serialized via `encoding/json`) — [#60 · 5690778272](https://github.com/gosharplite/tellme/issues/60#issuecomment-5690778272); (b) **recursive schema walk** for nested `required`/`properties` (the gate is flat/root-only today) — [#60 · 5690604951](https://github.com/gosharplite/tellme/issues/60#issuecomment-5690604951); (c) **SC-002** — the **manual** live Vertex/Gemini confirmation (plain + tool-using prompt, no `400`) is the round's one open, non-gating closeout step.
- **Round-030 forward items** — (a) a truncation failure **accounts no usage** (TD-1); (b) the guard runs **before** the generic content-empty "no usable answer" path (TD-3 divergence); (c) Vertex/Gemini `MALFORMED_FUNCTION_CALL` out of scope (TD-4); (d) the guard matches `"length"`/`"MAX_TOKENS"` **exactly**. (Round-030 detail → [`2026-09-16.md`](docs/archives/status/2026-09-16.md).)
- **Round-029 forward items** — (a) `write_file` survival via `--tool-usage` after dogfooding; (b) `replace_text` reads the **whole** file (an unbounded input read); (c) **TD-3** — `replace_text` through a symlink replaces the link. (Round-029 detail → [`2026-09-16.md`](docs/archives/status/2026-09-16.md).)
- **Older forward items** — rounds 018–028 forward items are recorded per-round in the archives (`2026-09-15.md` for 020–026; `2026-09-16.md` for 027–029).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` — **refreshed this session from `06d574f`** (the merged round-031 head, so it carries the tool-schema `reason`-property fix + gate). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Real-endpoint note (now resolved)**: bug [#64](https://github.com/gosharplite/tellme/issues/64) — the tool-schema defect that made a Vertex/Gemini provider 400 **every** request — is **fixed by round 031**; a Vertex/Gemini provider is expected to work against a real endpoint again (SC-002's **manual** live confirmation remains a closeout step). The hermetic fakes still do **not** schema-validate (a recorded `#60` forward item).
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020). The other dev host (`…/mbp-johndoe-niffler/`) is macOS and mirrors the opposite.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` — the vendored skill tree; read for `/home/pos/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — **clean this closeout**.
