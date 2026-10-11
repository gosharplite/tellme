# Session Summary — 2026-10-11

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `$TELL_ME_HOME` on the new **`kiosk`** env (a **GKE container managed by ConfigHub**); linux/amd64; Go **1.26.6** system, **1.26.9** user-scoped.
**Session mode**: `butler`.
**Branch**: `096-dependency-vuln-hygiene` (off `dev` `99f1114`).
**Status at end of session**: **round 096 `096-dependency-vuln-hygiene` IN FLIGHT** — a **dependency / toolchain vulnerability-hygiene** round (bump `golang.org/x/net ≥ v0.60.0` + record the `vulncheck` gate's toolchain floor ≥ 1.26.9; **ADR 0066**); **PR [#202](https://github.com/gosharplite/tellme/pull/202) OPEN** (awaiting a human merge). **Last delivered round remains 095.**

---

## 1. Session 88 — bootstrap → new `kiosk` env → `make check-full` triage → round 096

The session began with `SESSION-BOOTSTRAP.md` Steps 1–5 (round 095 delivered/frozen; active branch `dev`, clean). The operator flagged the **new working environment (`booth` → `kiosk`)** and asked whether all tools were available; a `make check-full` run surfaced a red `vulncheck` gate; the operator directed a toolchain/dependency bump; that became **round 096**.

### At a glance

| Area | Outcome |
| --- | --- |
| Round | **096 `096-dependency-vuln-hygiene`** (operator value; **ADR 0066**) |
| Tool inventory (kiosk) | **All required present** (`go`/`gofmt`/`goimports`/`staticcheck`/`golangci-lint`/`govulncheck`/`modelith`); **`deadcode` missing** (advisory only) → **installed** (`go install golang.org/x/tools/cmd/deadcode@v0.47.0`; ADR 0064) |
| `make check-full` on Go 1.26.6 | **stopped at `vulncheck`** — 9 stdlib advisories + `x/net@v0.38.0` |
| Toolchain | Go **1.26.9** provisioned **user-scoped** (`~/sdk/go1.26.9` + a `~/.local/bin/go` wrapper pinning `GOROOT`, since the container pins `GOROOT=/usr/local/go` + `/usr/local/go` is root-only) |
| Module bump | `golang.org/x/net v0.38.0 → v0.60.0` (+ transitive `x/sys`·`x/term`·`x/sync`·`x/text`) |
| Gates (post-fix) | **`make check-full` exit 0** · `make verify` **OK** · `govulncheck` **"No vulnerabilities found."** · E2E **green (counts unchanged)** · `make test-race` no races |
| Delivery | branch `096-dependency-vuln-hygiene` → **PR [#202](https://github.com/gosharplite/tellme/pull/202) open into `dev`** (no Copilot review; only a human merges) |

### Environment facts (durable)

- **`kiosk` = a GKE container managed by ConfigHub.** `/home/pos` is a **PD-backed PVC** (`kiosk-home`, `premium-rwo`, `ReadWriteOnce`, zone `asia-east1-a`, reclaim `Delete`) ⇒ **durable across container restarts/redeploys**; only a PVC deletion loses it. `/` is an **overlay** (ephemeral).
- The container **pins `GOROOT=/usr/local/go`** (via `/etc/profile.d/k8s_env.sh` + the container ENV) and `/usr/local/go` is **root-only** (sudo needs a password) ⇒ the 1.26.9 bump is **user-scoped**; a system-wide upgrade needs a ConfigHub/image change (recorded; not a repo artifact).
- Tool inventory verified on PATH; the only gap was the **advisory** `deadcode` (installed).

### Decisions locked (round 096)

| # | Decision |
| --- | --- |
| **D1** | **Bump the module floor** `golang.org/x/net ≥ v0.60.0` (clears the gate's module half); `go.mod`/`go.sum` is a repo artifact. |
| **D2** | **The toolchain floor (≥ 1.26.9) is an environment prerequisite** (ConfigHub/image), recorded in `specs/truth/techstack.md` — **not** a `go.mod` `go` directive. |
| **D3** | The `go` directive stays `1.26`; the `go mod tidy` `1.26 → 1.26.0` rewrite is **inert**. |
| **D4** | **ADR 0066** records the gate's toolchain-sensitivity + the two floors; it **cites** ADR 0012 R1 (the toolchain is outside the make-hermetic boundary) and amends **no** ADR body. |
| **D5** | The round is **not modelled** (ADR 0041) — no `docs/domain-model/**` change. |

### Commits (branch `096-dependency-vuln-hygiene`)

| Commit | Note |
| --- | --- |
| *(this round)* | `feat(096)`: dependency/toolchain vuln hygiene — bump `golang.org/x/net` ≥ v0.60.0 + record the `vulncheck` toolchain floor (ADR 0066) |

### Open items (non-blocking)

- **PR for round 096** awaits a human review/merge → then the closeout (`SESSION-CLOSEOUT.md`): propagate `dev → main` (no-ff), tag **`round-096`**, refresh the binary.
- **ADR 0066 §Forward** RF-066-1…4 (the floor is recorded, not machine-checked · the environment provisioning is out of repo · the module floor is a `≥` · `x/net` stays indirect).

### Next steps

1. Human reviews + merges the round-096 PR; then the closeout (propagate `dev → main` no-ff, tag `round-096`).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `096-dependency-vuln-hygiene` until merged, then `dev`).

### PM follow-ups

- None (no PM-owned requirement gap; the round is a dependency/toolchain hygiene reconciliation).

### Process notes (durable)

- **A dependency bump is a *round*, not a docs-only hygiene commit.** It rides a `make verify` gate (`vulncheck`, zero-tolerance) and carries no precedent for a non-round `go.mod` change ⇒ treated as a full AIxBDD round (the round-095 "a dev-tool/gate fact is a truth clause" precedent).
- **The gate has two axes.** `govulncheck`'s red splits into an **ambient toolchain** half (fix in the environment) and a **module** half (fix in `go.mod`) — only the latter is a repo artifact. Record the *floor* so the next host does not re-discover it via a red gate.

---

## 2. Round 096 closeout (2026-10-11) — **DELIVERED / FROZEN** (`SESSION-CLOSEOUT.md` Steps 1–8)

Round 096 was **human-merged** (PR [#202](https://github.com/gosharplite/tellme/pull/202) → `dev` **`ed75e4f`**, **merge commit**); the remote branch was already gone (`git fetch --prune` → `[deleted] origin/096-dependency-vuln-hygiene`), the local branch tip (`465e5c7`) was an **ancestor of `origin/dev`**, so the **local branch was deleted** (`git branch -d 096-dependency-vuln-hygiene`, was `465e5c7`), `dev` was fast-forwarded to the merge, and `SESSION-CLOSEOUT.md` Steps 1–8 ran.

| Step | Outcome |
| --- | --- |
| **1 — working tree** | `dev` clean at start; `dev == origin/dev == ed75e4f`; no delivered `specs/plans/**` package touched after freeze; no stray temp files; round branch already deleted (local + remote) |
| **2 — gates** | **`make check` OK** (`make verify` OK + `go test -count=1 ./...` green) · `make test-race` **no data races** · `govulncheck` **"No vulnerabilities found."** · diff-level secret grep clean (no credential patterns) |
| **3 — STATUS.md** | header → round 096 **DELIVERED / FROZEN**; **Rule-12 split**: the **round-095 delivered-round detail** relocated **verbatim** into [`docs/archives/status/2026-10-11.md`](../../../../archives/status/2026-10-11.md); round-096 section added; delivered-rounds pointer → 001–096; roadmap candidates unchanged; the round-096 env note + the round-close-tags line (`round-096` at this propagation) refreshed |
| **4 — daily summary** | this §2 (closeout) appended (the §1 record preserved) |
| **5 — reconcile** | `STATUS.md` ↔ this summary agree: no round in flight, `dev` active, branch heads match, tracker → live seeds #194–#197 open |
| **6 — commit** | working `dev` committed + pushed |
| **7 — propagate + hand off** | `dev -> main` (**no-ff**), tagged **`round-096`** (verified `main^{tree} == dev^{tree}`); installed binary refreshed (`go install ./cmd/tellme`; `--version` → `dev`) |
| **8 — issue tracker** | **no round anchor issue** (operator-value round) — verified the live seeds [#194](https://github.com/gosharplite/tellme/issues/194)–[#197](https://github.com/gosharplite/tellme/issues/197) are **open** and untouched ⇒ left as-is |

### Commits (branch `096-dependency-vuln-hygiene`, then merged by the human)

| Commit | Note |
| --- | --- |
| `531d61b` | `feat(096)`: bump `golang.org/x/net ≥ v0.60.0` + record the `vulncheck` toolchain floor (ADR 0066) |
| `465e5c7` | `docs(096)`: name round PR #202 in STATUS + the day log + tasks |
| `ed75e4f` | PR [#202](https://github.com/gosharplite/tellme/pull/202) merge into `dev` (by the human, merge commit) |
| *(this closeout, on `dev`)* | `docs(096)`: day close — round 096 delivered + propagated; STATUS split + 10/11 summary §2 |

### Open items (non-blocking)

- **None new.** `STATUS.md` carries no open-items index: a deferred item lives in its `ADR 00NN §Forward` or a live GitHub issue. **ADR 0066 §Forward RF-066-1…4** are disclosures, not tasking.
- **Issue tracker**: **#194–#197 open** (live round seeds); round 096 opened/closed no issue.

### Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value or a live issue** (live seeds [#194](https://github.com/gosharplite/tellme/issues/194)–[#197](https://github.com/gosharplite/tellme/issues/197); Bootstrap Agent Rule 11).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

### PM follow-ups

- None (no PM-owned requirement gap; a dependency/toolchain hygiene round).

*(Round 096 is fully closed out: PR #202 human-merged into `dev` (`ed75e4f`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-096`**; the installed binary refreshed.)*
