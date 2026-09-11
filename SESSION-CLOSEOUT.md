# 🌙 tellme — Session Closeout

> **Repo**: `github.com/gosharplite/tellme`
> **Folder**: `~/tmp/github/gosharplite/tellme/`
> **Mission**: Disciplined BDD re-creation of `tell-me-go` driven by `aixbdd-tmg`
> **Workflow**: AIxBDD (Strict PM/RD separation, single truth, Red-Green-Refactor)
> **Companion**: start-of-session procedure is [`SESSION-BOOTSTRAP.md`](SESSION-BOOTSTRAP.md) — this file is its end-of-day mirror.

---

## 🌇 Mandatory Closeout Steps — Execute In Order

> 👤 **Humans**: run this at the end of a working day (or before handing off).
> ⛔ **Do NOT stop midway on a red gate. Do NOT leave the day uncommitted. Do NOT merge without approval — just execute the steps.**

| Step | Action | Description |
|:---|:---|:---|
| **1** | Review the working tree | `git status` + `git diff --stat`: no half-written artifacts, no stray temp files, every change belongs to the active round / plan package. |
| **2** | Run the quality gates | Execute the project's gates — currently `gofmt` + `go vet` (research D7); the full `make check`-style pipeline once code lands. Docs-only round: verify internal links, artifact consistency, and run a secret scan. **Never close out on a red gate.** |
| **3** | Update [`STATUS.md`](STATUS.md) | Refresh "Last updated", pipeline position, artifacts checklist, decisions locked, open items, and environment notes — the single live-state source the next `SESSION-BOOTSTRAP.md` reads. |
| **4** | Write / refresh the day's session summary | Create or update `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` — what was done, decisions, artifacts, commits, open items, next steps. This is the file `SESSION-BOOTSTRAP.md` **Step 8** reads. |
| **5** | Reconcile status ↔ summary | Confirm `STATUS.md` and the day's `session-summary.md` agree: same decisions, same pipeline position, same open items, same branch heads. |
| **6** | Commit the working branch | Commit with a descriptive message (e.g. `docs(<NNN>): …`). The day must end committed and pushed. |
| **7** | Propagate + hand off | If the round is at a mergeable point **and the user approves**, run the two-step merge `working → dev → main`; otherwise record the pending propagation in `STATUS.md`. State the exact next-session starting point. |

Only after Steps 1, 2, 3, 4, 5, 6, and 7 are complete and results are reported is the session closed out.

---

## ⛔ END OF FILE — EXECUTE NOW

**You just finished reading this file. Do not reply. Do not summarize. Do not ask what to do next.**

Immediately return to the step table above and execute **Step 1 → Step 2 → Step 3 → Step 4 → Step 5 → Step 6 → Step 7** in order. Report the closeout result when Steps 1–7 are complete.

---

## 🗺️ Execution Mapping — Closeout Details

### 1. Working-Tree Review (Step 1 Details)

1. Run `git status --porcelain=v1 -b` and `git diff --stat` on the working branch.
2. Confirm the working branch matches `STATUS.md`'s **Active branch**; flag any drift.
3. Confirm no artifacts of a **delivered** `specs/plans/NNN-<slug>/` package were touched (frozen history).
4. Remove or commit stray temp files (e.g. `/tmp` staging, editors' leftovers); nothing untracked should be left ambiguous.

### 2. Quality Gates (Step 2 Details)

Run the gates the project currently owns — no more, no less:

- **Code rounds**: `gofmt -l .` clean, `go vet ./...` clean (research D7 — `golangci-lint` is deferred until adopted). Once the round-001 implementation exists, this widens to the `make`-style pipeline (build, vet, tests, coverage) referenced by `tell-me-go`'s Makefile — adapted, never copied blindly.
- **Docs / pre-implementation rounds**: verify Markdown links resolve, artifact cross-references agree (`STATUS.md` ↔ daily log ↔ `truth-delta.md`), and run a **secret scan** over the diff before committing.
- **Always**: confirm no secrets, credentials, or `secrets/`-style files are staged.

A red gate **stops closeout** — fix, re-run, then continue. Never record a passing status over a failing gate.

### 3. `STATUS.md` Update (Step 3 Details)

Keep `STATUS.md` the single live-state source the next bootstrap reads:

1. Bump **"Last updated"** to today (add the end-of-day marker).
2. Update the **Active branch** / branch model table (heads, tracking, propagation status).
3. Update **Current round** — scope, and the **Artifacts** checklist (`[x]` only what actually landed).
4. Update **Pipeline position** (which `axb-*` skill is `done` vs `next`) and any **pending decision**.
5. Record **decisions locked this session**, **PM follow-ups**, and **Open items (non-blocking)**.
6. Refresh **Environment notes** (provider upgrades, sandbox/host limitations, tool availability).

### 4. Daily Session Summary (Step 4 Details)

Write or extend `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` (e.g. `docs/session-summary/2026/09/10/session-summary.md`):

1. Header: project, repo, workspace (`$TELL_ME_HOME`), session mode, branches, **status at end of day**.
2. If the day had **multiple sessions**, number them (e.g. §1–§7 earlier, §8–§14 later) and say so — mirror the existing daily-log style.
3. Per session: at-a-glance table, artifacts produced, skills run, decisions, the grill/clarify outcomes if any, feasibility results.
4. **Decisions log** and a **commits** table (working branch unless noted).
5. **Open items (non-blocking)** and **Next steps** (the exact skill/step to resume at).
6. **PM follow-ups** (spec/acceptance are PM-owned — record, do not write them here).
7. Link the summary from `STATUS.md` and keep the **back-link** from the summary to `STATUS.md`.

### 5. Status ↔ Summary Reconciliation (Step 5 Details)

The next bootstrap reads **both** (`STATUS.md` at Step 7, the day summaries at Step 8) — they must not disagree:

- Same **pipeline position** (`done` vs `next`).
- Same **decisions locked** and **open items**.
- Same **branch heads** and **propagation status**.
- Same **artifact checklist** (`[x]` set).

On any mismatch, fix the stale one (usually the summary) before committing.

### 6. Commit (Step 6 Details)

1. Stage only the round's artifacts + the closeout edits (`STATUS.md`, the daily summary).
2. Commit with a scoped, descriptive message (e.g. `docs(001): day close — system-analysis plan + status`).
3. Push the working branch (tracking `origin/<branch>`); confirm the push succeeded.
4. **Never** auto-commit speculative or half-finished work — closeout commits the state you are willing to hand off.

### 7. Branch Propagation & Handoff (Step 7 Details)

1. Propagation is the **two-step merge** `working → dev → main`; it lands on `dev` (integrated) before `main` (stable).
2. Run it **only when the round is at a green, mergeable point and the user approves** — otherwise leave it pending and record "Propagation pending" in `STATUS.md`.
3. Handoff check — state the **next-session starting point**: active branch, pipeline position, the exact next skill/step, and any pending decision.
4. Confirm `SESSION-BOOTSTRAP.md` will still run cleanly: `STATUS.md` current, the day summary present, active branch correct.

---

## ⚠️ Closeout Rules

1. **Never close out on a red gate** — fix the failure first; a green status must be earned.
2. **End committed and pushed** — the working branch must be clean, committed, and pushed before the session ends.
3. **Status ↔ summary must agree** — `STATUS.md` and the day's `session-summary.md` are read together at the next bootstrap; keep them in sync.
4. **Frozen history** — never modify a delivered `specs/plans/NNN-<slug>/` package during closeout (or ever); new work starts a fresh package.
5. **Truth integrity** — closeout **records** truth changes, it never authors them; `specs/truth/**` changes only through their owner skills (recorded in `truth-delta.md`).
6. **Record every locked decision** — a decision not written into both `STATUS.md` and the day summary is lost to the next session.
7. **No secrets committed** — scan the diff before committing; credentials/`secrets/`-style files never enter history.
8. **Propagate only with approval** — the two-step merge `working → dev → main` runs on a green, user-approved round; otherwise record it as pending.
9. **Hand off explicitly** — always name the active branch, pipeline position, next skill/step, and any pending decision for the next session.
10. **Keep it current and linked** — maintain the daily log and its back-link so `SESSION-BOOTSTRAP.md` Step 8 always has a fresh, accurate summary to read.
