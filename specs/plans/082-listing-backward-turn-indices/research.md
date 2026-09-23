# Technical Research — 082-listing-backward-turn-indices

**Plan Package**: `specs/plans/082-listing-backward-turn-indices`
**Spec**: [`spec.md`](spec.md)
**Anchor issue**: [#165](https://github.com/gosharplite/tellme/issues/165)
**Status**: complete — decisions **D1–D8** below are locked; residuals live in **ADR 0054 §Forward**.

> Decision-driven research for the round. The spec (behaviour) is fixed by the issue + the operator directive; this document decides the *technical* shape: where the turn index is carried, the exact label, the truth/record surfaces, and the scope guard.

---

## 1. Problem & grounding

`tellme -l [N]` (round 073; ADR 0045) prints, per message, a role header line (`[USER]` / `[MODEL]`) followed by the body. `tellme -b [N]` (round 081; ADR 0053) removes the last N complete turns. The operator must **count turns upward from the bottom** of the `-l` output to choose `-b 1` / `-b 2` / `-b 3`. The round appends a **backward-counting turn index** to each header so the printed suffix is exactly the `-b` argument.

Measured at the round start (`dev` @ `f9aa96f`):

| Site | Shape |
| --- | --- |
| `internal/domain/render/ports.go` | `ListingMessage{Role ListingRole; Body string}` — no index. |
| `internal/cli/cli.go` `listingMessages(entries, n)` | emits (Operator, Model) per entry, then `msgs = msgs[len(msgs)-n:]`. |
| `internal/ui/listing.go` `header(role, colour)` | `blue("[USER]", colour)` / `magenta("[MODEL]", colour)`. |
| `internal/cli/cli.go` `renderHistoryList` | `newListing().Render(env.stdout, listingMessages(entries, n), ListingSpec{Colour: env.stdoutIsTerminal(), Raw: raw, Width: res.WrapWidth, Warn: env.stderr})`. |

`entries` is oldest→newest; `entries[len(entries)-1]` is the most recent turn (distance `1`), `entries[i]` has distance `len(entries) - i`.

---

## 2. Decisions

### D1 — The turn index is a field on the domain listing message, computed by the CLI

`render.ListingMessage` gains a `TurnIndex int` field. `listingMessages` computes it from the **entry's position in the full loaded history** (`len(entries) - i`, `i` = 0-based entry index) **before** the message-count truncation, and stamps it on **both** messages of that entry. The `internal/ui` adapter then formats the header from `m.TurnIndex`.

**Why not compute in the adapter from the slice length:** the adapter sees only the **truncated** slice, so a slice-length-derived index would renumber a lone leading `[MODEL]` to `- 1` (spec **FR-004** / invariant I-3 requires the **true** distance). The CLI holds the full `entries`, so it is the only site that can compute the true distance. `TurnIndex` is a **domain-owned shape** (the port already carries the role + body; the index is the same kind of presentation input), so the CLI still names no `internal/ui` type (ADR 0020 / RULE-E is untouched).

**Why not a richer type / a per-entry call:** a single `int` on the existing value type is the least surface and keeps the adapter's `Render` signature stable.

### D2 — The exact label: `[USER] - N` / `[MODEL] - N`, with a bare-label fallback for a non-positive index

The header becomes `fmt.Sprintf("[%s] - %d", roleWord, m.TurnIndex)` where `roleWord` is `USER` / `MODEL` — the exact shape the issue's Example Output shows (ASCII hyphen, single spaces: `[USER] - 2`). When `m.TurnIndex <= 0` the adapter falls back to the existing bare `[USER]` / `[MODEL]` label, so a non-history caller (a future direct `render.Listing` use) never prints `- 0`. The historical caller always passes `>= 1`, so the shipped `-l` output always carries the suffix.

**Alternatives rejected:** an en-dash `–` (the issue's examples use the ASCII `-`); a bare integer with no brackets (breaks the round-073 header contract); a `TurnIndex uint` (cannot express "absent" without a sentinel).

### D3 — The colour unit is the whole label

The accent wraps the **formatted** label (`blue("[USER] - 2")` / `magenta("[MODEL] - 2")`), the empty-safe `wrap` being unchanged: on a terminal `stdout` with `-r` off the whole label is accented; under `-r`/`--raw` **or** a non-terminal `stdout` it is plain with no escapes. The round-073 stdout-terminal gate (`env.stdoutIsTerminal()` + `TELL_ME_FORCE_STDOUT_TTY`) is reused **unchanged** (no new gate). This satisfies spec **FR-005/FR-006**.

### D4 — Truth & records

- **ADR 0054** (`docs/decisions/0054-list-backward-turn-indices.md`) — **amends ADR 0045** (the listing presentation owner), recording the header label change + the colour-unit decision + the scope guard; indexed in `docs/decisions/README.md`.
- **`specs/truth/techstack.md`** — MODIFY the **Session lifecycle flags** row (the `-l` listing presentation owner): the round-082 clause.
- **`specs/truth/features/cli/history/**`** — the `-l` listing gains a Rule for the header label (the owning interface truth) + `history/dsl.md` rows; **`/axb-dsl-refine`** owns this.
- **`docs/domain-model/**`** — **NOT modelled.** The listing is a **presentation surface** (the `render.Listing` port and its value types are not domain-model entities; the model's `Session`/`History` offline invariants are unaffected — the stored history is read, never changed). ADR 0041 escape hatch: this is recorded in `plan.md` §5.
- **`specs/truth/contracts/**`** + **`specs/truth/data/**`** — **NOOP** (no API surface; no persisted-state shape change).

### D5 — The `-l` selection semantics are unchanged

`-l N` still selects the last **N messages** (not turns); `TurnIndex` is computed on the full entry list before truncation, so the **message-count** selection and the **true-distance** index are independent. This keeps the round-073 selection contract (and the round-054 default) byte-identical in *selection*, changing only the header label.

### D6 — Scope guard (excludes)

Absolute / session-start numbering (`[USER] #3`), a `--json` listing, a config toggle, any change to `-b`/`--back`, `history.Store`, `history.jsonl`, the interactive prompt, or the live turn chrome (`╭─⠿ Turn N`). No new exit code; no new class phrase; the exit-code set stays **ten**.

### D7 — Determinism, hermeticity, no new dependency

The change is pure formatting arithmetic over the already-loaded slice; stdlib-only (`fmt`); POSIX-only; no new dependency; `go.mod`/`go.sum` unchanged. The E2E arranges history via the existing `the session history already holds the exchanges:` Given and asserts the header text on `stdout` — no pty, no network.

### D8 — Verification strategy

- **Unit (`internal/ui/listing_test.go`)**: the adapter formats `[USER] - N` / `[MODEL] - N`; the whole label is the colour unit; the `TurnIndex <= 0` fallback prints the bare label.
- **Unit (`internal/cli/list_render_test.go` or a sibling)**: `listingMessages` stamps the same index on both messages of an entry and the **true** distance on a partial slice (the odd-`-l` case) — the load-bearing arithmetic.
- **E2E (`tests/e2e/steps/`)**: extend the round-073 listing stepdefs — assert the header sequence for the full and partial listings (SC-001…SC-003) and the accent/plain discipline (SC-004).
- **Falsifiability witnesses** (reproduced then reverted): (a) drop the `TurnIndex` stamp ⇒ the E2E header-sequence Then reds; (b) compute the index from the truncated slice ⇒ the odd-`-l` Then reds (the leading `[MODEL]` renumbered to `- 1`); (c) accent the bare word (not the whole label) ⇒ the terminal-accent Then reds.

---

## 3. Residual risks (homed in ADR 0054 §Forward)

- **RF-082-1** — the label is not rune-capped (it is tellme-authored, bounded by the index magnitude — a cosmetic forward item).
- **RF-082-2** — the `TurnIndex` field is public on the port value type; a future caller must set it (the `<= 0` fallback is the guard).
- **RF-082-3** — no absolute/`--json` numbering (the issue's explicit scope).
- **RF-082-4** — the header format is asserted by a **predicate** (the round-073 precedent), not a byte-golden of the whole listing; the `-r` path stays byte-exact.
- **RF-082-5** — the index is derived from the loaded history only; a rollback (`-b`) then `-l` shows the trimmed numbering (correct, and already covered by the round-081 `-l`×`-b` compose).
