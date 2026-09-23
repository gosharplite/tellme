# ADR 0054 — Backward-counting turn indices in the `-l`/`--list` role headers

- **Status:** Accepted
- **Date:** 2026-09-23
- **Deciders:** tellme owner
- **Related:** [ADR 0045](0045-list-role-headers-and-rendered-body.md) (the `-l` listing role headers + the stdout-terminal colour gate this **amends**), [ADR 0053](0053-back-rollback-turns.md) (the `-b`/`--back [N]` rollback this aligns with), [ADR 0023](0023-list-default-and-chrome-colour.md) (the `-l` optional value + the chrome colour policy), [ADR 0020](0020-cli-ui-decoupling.md) (the `render.*` domain ports the CLI uses — the new field keeps the CLI free of `internal/ui` types); issue [#165](https://github.com/gosharplite/tellme/issues/165); round 082 (`specs/plans/082-listing-backward-turn-indices`)

## Context

`tellme -l [N]` (round 073; ADR 0045) prints each listed message as a role header line (`[USER]` / `[MODEL]`) plus a body; `tellme -b [N]` (round 081; ADR 0053) removes the last N **complete turns** of the session history. To choose the rollback count an operator currently counts turns upward from the bottom of the `-l` output — error-prone and divorced from the stored model.

The listing is a pure presentation of the loaded `history.Entry` slice (oldest → newest): `entries[i]` is the `len(entries) - i`-th turn counting back from the end. The `-l N` selection takes the last **N messages** (a message is a role + body), so an **odd** N can lead with a `[MODEL]` whose partner `[USER]` is outside the printed window.

## Decision

**D1 — a backward-counting turn index on each role header.** Each listed message's header gains a trailing ` - N`: `[USER] - N` / `[MODEL] - N`, where `N = 1` is the **most recent** `history.Entry` and `N` increments toward older turns. Both messages of one entry carry the **same** `N` (a turn is atomic under `-b`, so a split label would invite a half-turn rollback).

**D2 — the index is the *true* distance, carried on the domain listing value.** `render.ListingMessage` gains a `TurnIndex int` field. The CLI's `listingMessages` computes it from the entry's position in the **full loaded history** (`len(entries) - i`) **before** the message-count truncation, and stamps it on both messages of that entry. The `internal/ui` adapter formats the header from `m.TurnIndex`; it never derives the index from the (truncated) slice length, so a lone leading `[MODEL]` keeps its true distance (e.g. `[MODEL] - 2`) rather than being renumbered `- 1`.

**D3 — the label format and the fallback.** The header is `fmt.Sprintf("[%s] - %d", roleWord, m.TurnIndex)` (`roleWord` = `USER` / `MODEL`; ASCII hyphen; single spaces — the issue's Example Output shape). When `m.TurnIndex <= 0` the adapter falls back to the existing bare `[USER]` / `[MODEL]` label, so a non-history caller never prints `- 0`. The historical `-l` caller always passes `>= 1`.

**D4 — the colour unit is the whole label.** The accent wraps the formatted label (`blue("[USER] - 2")` / `magenta("[MODEL] - 2")`) via the existing empty-safe `wrap`; on a terminal `stdout` with `-r` off the whole label is accented, and under `-r`/`--raw` or a non-terminal `stdout` it is plain with no escapes. The round-073 stdout-terminal gate (`env.stdoutIsTerminal()` + `TELL_ME_FORCE_STDOUT_TTY`) is reused unchanged — no new gate, no new probe.

**D5 — the `-l` selection semantics are unchanged.** `-l N` still selects the last **N messages**; the index is computed on the full entry list before truncation, so the message-count selection and the true-distance index are independent.

**D6 — presentation-only.** No change to `history.jsonl`, `history.Store`, `-b`/`--back`, the interactive prompt, or the live turn chrome (`╭─⠿ Turn N`). No new exit code (the set stays **ten**); no new class phrase. The listing stays offline (no provider request, no stdin read).

**D7 — scope guard.** Excluded: absolute / session-start numbering (`[USER] #3`), a `--json` listing, a config toggle, and any change to `-b` or the session stores.

**D8 — records.** ADR 0054 (+ index) **amends ADR 0045**; `techstack.md` *Session lifecycle flags* MODIFY (the `-l` listing presentation owner) + *Raw output flag* cross-note; the `history` CLI feature + `history/dsl.md` rows (`/axb-dsl-refine`); `docs/domain-model/**` **NOT modelled** (a presentation surface — the `render.Listing` port is not a model entity; ADR 0041 escape hatch, recorded in `plan.md` §5); `contracts/**` + `data/**` **NOOP**.

## Consequences

- `tellme -l` prints `[USER] - K` / `[MODEL] - K`, giving a 1:1 visual correspondence to `tellme -b K` (index 1 = the turn `-b 1` removes).
- A partial listing (`-l` an odd count) shows a leading message at its true distance, so the operator can still read the right `-b` count.
- The change is byte-local to the header lines: the body rendering, the blank-line separator, the `-l` last-N-**messages** selection, the `-r` behaviour, and the `[Tool …]`/turn chrome are unchanged.
- No new dependency; stdlib-only (`fmt`); POSIX-only; `go.mod`/`go.sum` unchanged.
- **Recorded divergence:** the reference's `-l` prints no turn index at all (the operator counts by hand); tellme adds the backward index as an operator-value enhancement beyond parity.

## Alternatives considered

- Compute the index in the `internal/ui` adapter from the slice length — rejected (D2: renumbers a lone leading message, violating the true-distance requirement).
- A bare integer with no brackets, or an en-dash separator — rejected (D3: breaks the round-073 header contract / departs from the issue's shape).
- Absolute (session-start) numbering — rejected (D7: does not align with `-b [N]`, the issue's explicit exclusion).
- A separate `-l` turn-count flag / a config toggle — rejected (D5/D7: scope creep; the index is free information already in the listing).
