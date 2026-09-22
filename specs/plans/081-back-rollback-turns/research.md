# Technical Research: Roll back the last N turns of the session history (`-b`/`--back`) (round 081)

**Plan Package**: `specs/plans/081-back-rollback-turns`
**Created**: 2026-09-22
**Anchor**: issue [#163](https://github.com/gosharplite/tellme/issues/163)
**Grounding**: `dev` @ `0008fd0`

> Decision-driven research (no product code). The behaviour is anchored by the issue and the operator's directive; its *Open design decisions* (S-1…S-8) are resolved below (D1–D9). The design is tellme's own; the reference is the behaviour reference but is not asserted byte-for-byte.

---

## 0. The problem, precisely

`tellme` has **no way to undo a turn**. `parseFlags` registers `-c -d -h -i -l --new -r --tool-usage -t -v`; `tellme -b` is an unrecognized flag → `tellme: the command-line usage is invalid` + exit **2**. The reference ships `-b`/`--back [N]` *"Go back / delete the last N turns from history"* with `NoOptDefVal = "1"`, and its `bb` alias is a frequently-used shorthand.

Three structural facts make this a clean, small feature in `tellme`:

1. **A turn is exactly one `history.Entry` line.** `file_store.go` writes one JSON line per completed turn `{prompt, answer, calls, steps[]}`; the tool steps are **inside** that line's `steps[]`. So a rollback of `N` turns is the **removal of the last `N` lines** — no message-pair arithmetic, and **no** risk of splitting a `FunctionCall`/`FunctionResponse` pair (the reference, which stores one *message* per line, must drop `N × 2` messages and repair odd parity; tellme does not).
2. **The active file holds only complete turns.** Rounds 079/080 (ADR 0051/0052) close a partial/failed turn with a synthetic assistant answer **before** writing it, so the file is always a sequence of complete `user … assistant` turns; a truncation at a line boundary is therefore always a clean cut that still replays on **both** provider families.
3. **The offline session selector already exists.** `resolveWorkspace` + `offlineConfigAndMode` resolve the session `mode` for `-l`/`-t`/prompt-less `--new` (mode = `TELL_ME_MODE` → the `-c` config's `MODE` → the default config's `MODE` → `butler`), and an explicit unreadable `-c` refuses. `-b` reuses this verbatim.

The one missing capability is at the **domain port**: `internal/domain/history.Store` exposes only `Load()`, `Append(Entry)`, `Archive()` — there is no truncate/rewrite. And the adapter's `Archive()` moves the *whole* active file (that is `--new`, not a rollback).

## 1. Decisions

### D1 — The flag: `-b`/`--back [N]`, optional count, default 1 (S-1 / L-1)

`fs.IntVarP(&o.back, "back", "b", 0, "Roll back the last N turns of the session history (defaults to 1 when the value is omitted).")` with `fs.Lookup("back").NoOptDefVal = "1"`. Mirroring `-l` (round 054 / ADR 0023), the argv pre-pass consumes an **adjacent integer** (`-b 3` → `-b=3`) and leaves a **non-integer** token untouched, so `tellme -b "new prompt"` means `-b 1` **plus** the positional prompt `"new prompt"` (L-1; the issue's `-b "new prompt"` ask). A bare `-b` means `-b 1`.

### D2 — Standalone is offline; with a prompt it runs (L-2 / FR-003 / FR-006)

`tellme -b` / `tellme -b N` (no positional prompt) is an **offline** action: resolve the workspace, roll back, print a confirmation to `stdout`, exit **0** — **no** provider request. It joins the offline-path set of the `tellme performs no network access` truth row. `tellme -b [N] "p"` rolls back **first**, then runs the prompt as a normal reasoning turn (chrome + spinner + provider request permitted) — mirroring the reference's 3-phase order (`-l` render → `-b` rollback → chat).

### D3 — The Store capability: a durable `Rollback` on the domain port (S-1) — **option (a)**

Extend `internal/domain/history.Store` with:

```go
// Rollback removes the last n complete turns (entries) from the active
// history, writing the surviving entries durably. It returns the number of
// turns actually removed (clamped to the available turns) and the number
// remaining. n <= 0 is a no-op. The archive is never touched.
Rollback(n int) (removed int, err error)
```

*Rationale:* a **named, intent-revealing** capability (the reference's `RollbackTurns`) beats a generic `Rewrite([]Entry)` — it states the one operation the CLI needs, keeps the CLI free of "load → trim → re-save" bookkeeping, and concentrates the durability logic in the adapter. *Rejected:* (b) a generic `Rewrite` — a wider, less-meaningful surface; (c) load-trim-rewrite at the CLI layer — spreads history knowledge into the CLI and re-implements durability there.

**Durability (NFR-001 / I-3):** the adapter writes the surviving entries to a temp file in the **same directory**, `fsync`s it, then `os.Rename`s it over `history.jsonl` (atomic on POSIX), and `fsync`s the directory best-effort. A crash mid-rollback leaves the **prior** `history.jsonl` intact (the live file is never truncated in place before the new bytes are durable). A missing active file ⇒ 0 removed (an empty session, tolerated). A decode failure ⇒ returned as an error (never a partial write).

### D4 — Cardinality: clamp when `N` exceeds the available turns (S-2)

`removed = min(N, len(entries))`. `tellme -b 999` on a 2-turn session empties it and reports **2 removed**, exit 0. *Rationale:* reference parity, and it is the least surprising "go back as far as you can". *Rejected:* refusing with an error — it would make an idempotent "clear the session" impossible and adds a failure surface for no value.

### D5 — `N ≤ 0` is a usage error (S-3)

`-b 0` / `-b -1` → `tellme: the command-line usage is invalid` + exit **2**, symmetric with `-l ≤ 0` (round 007). *Rationale:* a rollback of zero turns is meaningless as a *command*; a silent no-op would hide an operator typo. *Rejected:* the reference's silent `backN ≤ 0` no-op. *(Note: the clamp in D4 covers the "more than available" case; the usage error covers the "not a positive count" case.)*

### D6 — Composition & precedence (S-4)

`-b` runs as a **phase before the prompt phase**, and **after** the offline reporters `-l` / `-t` / `--tool-usage` (which are terminal). Concretely, in `dispatchReporting` order, an **offline `-b`** (no positional prompt) sits **after `-l`**: `tellme -l 2 -b` lists the last 2 messages, then rolls back 1 (reference order).

- **`-b` × `-l`** — supported (list then roll back).
- **`-b` × `-t` / `--tool-usage`** — a rollback has no meaningful composition; `-b` is handled **after** `-l` and **before** `-t`? Decision: keep `-b` in the **same precedence tier as `-l`** (both are session-history actions), ordered **after `-l`** and **before `-t`**. If `-b` and `-t`/`--tool-usage` are given together with no prompt, `-b` runs first (a history mutation precedes a read of the *current* trace) — recorded; not a scenario the round must over-engineer.
- **`-b` × `--new`** — **decided: `-b` then `--new`** when a prompt or `--new` is present is **refused**? No — keep it simple and honest: a prompt-less `-b` rolls back and exits (it does **not** also archive). If `--new` is **also** present with no prompt, the two are contradictory session actions; **decision: `-b` takes precedence and `--new` is ignored for the run** is *not* acceptable (silent). Instead: **`-b` + `--new` (no prompt) is a usage error** (exit 2) — the operator must choose "undo the last N" *or* "start fresh". Recorded as a deliberate choice. *(A prompt-bearing `-b "p" --new` — archive then rollback is nonsense — also refuses.)* This is the smallest honest rule; the scenario is unlikely and the refusal is deterministic.
- **Prompt-bearing `-b [N] "p"`** — rolls back, then runs `"p"` (D2). `--new` is not combined here.

### D7 — The confirmation line (S-5)

One **plain** line to **`stdout`** (the offline-report channel, like `-l`), **no** `tellme: ` prefix, **not** in `turns.log`:

```text
Rolled back 2 turns. History now holds 1 turn.
```

Pluralisation: `turn`/`turns` for both figures (0 ⇒ `0 turns`). The wording is **contract-free beyond the effect** — the E2E asserts the removal (file bytes) primarily and may assert a substring of the line secondarily; no frozen phrase is added (NFR-003). A **failure** (an unreadable/undecodable `history.jsonl`, or a rename failure) reuses the existing environment-class phrase via `emitHistoryError` (exit **4**) — no new phrase.

### D8 — The optional-int pre-pass is generalised (FR-009, resolving ADR 0023 **RF-54-4**)

`consumeListValue(args)` becomes a shared `consumeOptionalIntArgs(args, names ...string)` (or an equivalent named helper) that consumes an **adjacent integer** for **any** of the optional-int flags (`--list`/`-l`, `--back`/`-b`), leaves a non-integer token as a positional, and stops at `--`. The recorded boundary cases carry over verbatim (positional-agnostic; combined short flags such as `-rl 5` / `-rb 2` are **not** split — recorded, unchanged). This **resolves** ADR 0023 §Forward **RF-54-4**.

### D9 — Records (S-7): a new ADR + truth + domain model + CLI feature

- **ADR 0053** (`docs/decisions/0053-back-rollback-turns.md` + index row), **resolving ADR 0023's RF-54-4**.
- `specs/truth/techstack.md`: **MODIFY** *CLI flag parsing* (register `-b`/`--back [N]`; the generalised pre-pass); **MODIFY** *Session history store* (add the durable rollback capability); **MODIFY** *Session lifecycle flags* (the offline `-b` action + the session selection); the offline-path scope of the `tellme performs no network access` row gains the standalone `-b`.
- `docs/domain-model/tellme.modelith.{yaml,md}`: **MODIFY** the `History` entity — add a **rollback** action (a session's last N complete turns may be removed atomically) and, if warranted, an invariant (the archive is untouched; the schema is unchanged; only complete turns exist). Re-rendered; `modelith-check` green (ADR 0041 same-PR).
- `specs/truth/features/cli/history/inspecting-the-session-history.feature` (the `history` module) gains rollback Rules/Examples; `history/dsl.md` gains the Given/When/Then rows; `specs/truth/data/**` is **NOOP** (the `history_entry`/`history_step` shapes are unchanged — a rollback removes whole lines and adds none).

## 2. Rejected alternatives (recorded — not forward items)

| Alternative | Why rejected |
| --- | --- |
| A generic `Store.Rewrite([]Entry)` | Wider surface than the one intent; the CLI would carry history bookkeeping (D3). |
| Load-trim-rewrite in the CLI | Spreads history/durability knowledge out of its adapter (D3). |
| Refuse when `N` > available | Blocks the idempotent "clear the session"; adds a needless failure surface (D4). |
| Silent no-op for `N ≤ 0` (reference) | Hides an operator typo; asymmetric with `-l ≤ 0` (D5). |
| Writing the archive on rollback | Conflates rollback with `--new` (I-3b) — the archive is for "fresh session", not "undo". |
| A TUI/interactive confirmation | Out of scope; the round is a plain line-oriented CLI action (S-9). |
| `--retry` in the same round | A distinct semantic (roll back the last user message + resend); routed to a sibling round (S-9). |

## 3. Residual risks (→ §Forward of ADR 0053)

- **RF-081-1** — the confirmation wording is not a frozen contract (the effect, not the text, is asserted).
- **RF-081-2** — a rollback does **not** trim `turns.log` / `tokens.log` (they are a trace and a usage history; D-S6 no).
- **RF-081-3** — no `flock`/`ModeLocker`: a concurrent tellme process could race the rename (consistent with the repo's standing "no `flock`" policy).
- **RF-081-4** — the `-b` × `--new` refusal (D6) is a tellme-specific rule; the reference has no `--new` conflict handling.
- **RF-081-5** — `-b` combined with `-t`/`--tool-usage` runs before them (recorded order); the reference has no such combination.
- **RF-081-6** — the offline `-b` is a *read-modify-write* offline action (unlike the read-only `-l`/`-t`); a `-c`-named mode with no workspace creates the workspace (the round-053 `EnsureWorkspace` behaviour).
- **RF-081-7** — `--retry` (roll back the last user message + resend) remains unimplemented (S-9).
