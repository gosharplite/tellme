# ADR 0053 — `-b`/`--back`: roll back the last N turns of the session history

- **Status:** Accepted — its durability clause is **witnessed and calibrated** by [ADR 0056](0056-rollback-durability-witness.md) (round 084; the file `fsync`-before-rename gains a mechanism-seam pin, the best-effort directory `fsync` is recorded accepted-unwitnessed)
- **Date:** 2026-09-22
- **Deciders:** tellme owner
- **Related:** [ADR 0023](0023-list-default-and-chrome-colour.md) (the `-l` optional-int pre-pass whose **RF-54-4** this ADR resolves by generalising it), [ADR 0022](0022-offline-session-config-and-turns-log.md) (the offline session-command selection this reuses), [ADR 0051](0051-interrupted-turn-partial-persistence.md) / [ADR 0052](0052-failed-turn-partial-persistence.md) (the guarantee that `history.jsonl` holds only complete turns, which makes a rollback a clean line cut), [ADR 0011](0011-layer-discipline-gate.md) (the layer ranking the adapter respects); issue [#163](https://github.com/gosharplite/tellme/issues/163); round 081 (`specs/plans/081-back-rollback-turns`)

## Context

`tellme` shipped no way to undo a turn. The reference (`tell-me-go`) ships `-b`/`--back [N]` (*"Go back / delete the last N turns from history"*, `NoOptDefVal = "1"`) and its `bb` alias is a frequently-used shorthand; `tellme -b` today is an unrecognized flag (exit 2).

`tellme`'s history model differs from the reference in a way that makes this **smaller and safer** here:

- **One `history.Entry` line per turn** (`{prompt, answer, calls, steps[]}`) — the tool steps live *inside* the line, so a rollback of N turns is the removal of the **last N lines**. The reference (one *message* per line) must drop `N × 2` messages and repair odd parity; tellme needs neither, and can never split a `FunctionCall`/`FunctionResponse` pair.
- **Only complete turns exist in the active file** — rounds 079/080 (ADR 0051/0052) close a partial/failed turn with a synthetic assistant answer before writing it — so a line-boundary truncation always leaves a valid `user … assistant` sequence on **both** provider families.
- **The offline session selector already exists** (`resolveWorkspace`/`offlineConfigAndMode`, round 053 / ADR 0022), so `-b` reuses it verbatim.

The one missing capability is at the domain port: `internal/domain/history.Store` exposes only `Load`/`Append`/`Archive` — there is no truncate/rewrite. The adapter's `Archive()` moves the *whole* active file (that is `--new`).

## Decision

**D1 — the flag.** Register `-b`/`--back [N]` (`IntVarP`, `NoOptDefVal = "1"`). The argv pre-pass consumes an **adjacent integer** (`-b 3` → `-b=3`) and leaves a **non-integer** token as a positional prompt, so `tellme -b "p"` means "roll back 1, then run `p`".

**D2 — the two forms.** `tellme -b [N]` (no prompt) is an **offline** action: roll back, print a plain `stdout` confirmation, exit 0, **no** provider request (it joins the offline-path set). `tellme -b [N] "p"` rolls back **then** runs `"p"` as a normal reasoning turn (chrome + provider request permitted) — the reference's 3-phase `-l` → `-b` → chat order.

**D3 — a durable `Store.Rollback` (domain port, option (a)).** Extend `internal/domain/history.Store`:

```go
Rollback(n int) (removed int, err error)
```

removing the last **n complete turns** (whole entries, including their embedded `steps[]`). It is **durable and atomic**: the surviving entries are written to a temp file in the same directory, `fsync`ed, then `os.Rename`d over `history.jsonl` (the live file is never truncated in place); a crash mid-rollback leaves the prior history intact. A missing active file ⇒ 0 removed. A decode failure ⇒ an error (never a partial write). *Rejected:* a generic `Rewrite([]Entry)` (a wider, less-meaningful surface) and load-trim-rewrite at the CLI (spreads history/durability knowledge out of its adapter).

**D4 — clamp.** `removed = min(n, len(entries))`; `tellme -b 999` on a 2-turn session empties it and reports 2. *Rejected:* refusing when `n` > available (blocks the idempotent "clear the session").

**D5 — `n ≤ 0` is a usage error** (exit 2), symmetric with `-l ≤ 0`. *Rejected:* the reference's silent no-op.

**D6 — the archive is never touched.** A rollback drops the last N turns and keeps the rest; it **MUST NOT** write `history.archive.jsonl` (rollback ≠ `--new`).

**D7 — composition / precedence.** `-b` is a phase after `-l` and before `-t` (`tellme -l 2 -b` lists then rolls back). `-b` combined with `--new` **refuses** (usage error, exit 2) — "undo the last N" and "start fresh" are contradictory session actions; a silent precedence would be dishonest.

**D8 — the confirmation line.** One plain line to **`stdout`**, no `tellme: ` prefix, not in `turns.log`: `Rolled back N turns. History now holds M turn(s).` The effect (the file bytes) is the contract; the wording is not frozen. A failure (unreadable history / rename failure) reuses the environment-class phrase via `emitHistoryError` (exit 4) — no new phrase; the exit-code set stays **ten**.

**D9 — the optional-int pre-pass is generalised** (resolving ADR 0023 **RF-54-4**). The round-054 `consumeListValue` becomes one helper covering `-l`/`--list` **and** `-b`/`--back`; the recorded boundary cases carry over (positional-agnostic; combined short flags such as `-rl 5` / `-rb 2` are **not** split).

**D10 — records.** ADR 0053 (+ index); `techstack.md` *CLI flag parsing* / *Session history store* / *Session lifecycle flags* MODIFY; the `history` CLI feature + `history/dsl.md` rows; the `docs/domain-model` `History` entity gains a **rollback** action (same-PR, ADR 0041); `data/**` **NOOP** (schema unchanged).

## Consequences

- `tellme -b` ≡ `tellme -b 1`; `tellme -b N` drops the last N turns; `tellme -b "p"` undoes the last turn and re-asks; `tellme -l 2 -b` lists then undoes.
- The rollback is durable: a crash mid-operation cannot corrupt `history.jsonl`.
- `history.Entry` / `history.Step` JSON shapes are unchanged; no id/timestamp is introduced.
- No new dependency; POSIX-only; `go.mod`/`go.sum` unchanged.
- **Recorded divergence:** the reference drops message pairs (`N × 2`) and prints `⏪ Rolled back …`; tellme drops whole turn lines and prints its own plain line. The reference treats `backN ≤ 0` as a no-op; tellme refuses.

## Alternatives considered

- A generic `Store.Rewrite([]Entry)`, load-trim-rewrite at the CLI — rejected (D3).
- Refusing when `N` > available — rejected (D4).
- Silent `N ≤ 0` no-op — rejected (D5).
- Writing the archive on rollback — rejected (D6; conflates with `--new`).
- An interactive confirmation, a TUI `--back`, `--json` — out of scope.

## Forward items

- **RF-081-1** the confirmation wording is not a frozen contract (the effect is asserted).
- **RF-081-2** a rollback does not trim `turns.log` / `tokens.log` (a trace and a usage history).
- **RF-081-3** no `flock`/`ModeLocker` — a concurrent process could race the rename (the repo's standing no-`flock` policy).
- **RF-081-4** the `-b` × `--new` refusal is a tellme-specific rule (the reference has no such combination).
- **RF-081-5** `-b` with `-t`/`--tool-usage` runs before them (recorded order); the `-b` validation sits below the `-d` tier (symmetric with `-l`), so `-d -b 0` prints the diagnostic report rather than refusing, and a *valid* `-b` combined with `-d` is dropped by `-d`'s precedence (the same behaviour as `-l`).
- **RF-081-6** the offline `-b` is a read-modify-write offline action; a `-c`-named mode with no workspace creates the workspace (`EnsureWorkspace`, round 053).
- **RF-081-7** `--retry` (roll back the last user message + resend) remains unimplemented.
- **RF-081-8** `rawNonEmptyLines` copies each survivor line **trimmed** (the same `strings.TrimSpace` `Load` applies), so a hand-edited line with surrounding whitespace is normalised by a rollback rather than copied byte-for-byte; for a history this binary wrote the line is already minimal, so the copy is byte-verbatim (FR-010 holds structurally for tool-written history). A future raw-bytes-preserving copy would need its own line-splitting that tolerates padding.
- **RF-081-9** the prompt-bearing rollback's post-rollback `Load`-error branch (and `turnOptions.rollbackTurns` generally) has no unit witness — the in-package `fakeStore.loadErr` cannot fail only a second `Load` call; a cheap pin is available later if a fault-injecting store double is introduced.
