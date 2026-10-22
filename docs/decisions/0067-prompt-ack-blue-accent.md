# ADR 0067 — The input-capture acknowledgement is the fifth chrome accent, in bright blue

- **Status:** Accepted
- **Date:** 2026-10-22
- **Deciders:** tellme owner (operator directive, 2026-10-22; colour chosen by the operator)
- **Related:** **ADR 0023** (the chrome-colour policy — its **D2** gate is reused unchanged and its **D3**
  operator-locked element set gains a **fifth** element here; the 0023 **body stays verbatim**, its index
  row gains a forward pointer) · **ADR 0027** (the tool-chrome accents that crowded the palette) ·
  **ADR 0045**/**ADR 0054** (`[USER]` bright blue / `[MODEL]` bright magenta in the `-l` listing — the
  operator-voice precedent this element follows) · **ADR 0041** (the model is load-bearing; this round
  **is** modelled) · **ADR 0006** (falsifiability) · [issue #204](https://github.com/gosharplite/tellme/issues/204)
  · `internal/ui/colour.go` · `internal/ui/turn.go` · `internal/ui/render_ports.go` · `internal/cli/cli.go`

## Context

Every prompt-bearing turn opens with one tellme-owned diagnostic line on `stderr`:

```
[13:02:41] Input captured. Processing...
```

It is the moment the operator's prompt enters the session, and it sat between two things that are
visually settled: the terminal's own echo above it and the turn frame below it (`─` rule,
`╭─⠿ Turn N - <mode>` header, and — three lines later — a `Payload:` line whose MODE token is already
green). It was the **only** chrome line still **plain**: `internal/cli/cli.go:858` passed
`dp.NewLines(false)` — a leftover from before round 054 (when no chrome line carried colour), not a
deliberate exclusion.

**The obvious candidate was rejected on evidence, and that rejection is what selects this line.**
Colouring the operator's **prompt text** was investigated first (issue #204 revision 1) and fails,
because the prompt text is presented by four different owners:

| Surface | The operator's prompt text | Owner |
| --- | --- | --- |
| `-i` while typing | **white** — the editor sets no foreground (`internal/ui/tui/prompt/textarea.go`) | tellme's TUI, unstyled |
| `-i` after submit | tellme's re-echo (`internal/cli/cli.go:855`) | tellme |
| Ctrl+D multi-line | the **terminal driver's** echo — `internal/cli/prompt.go:116-122` never reads a character | **the tty** |
| positional | not echoed at all (truth row: *the prompt is not echoed*) | — |

Colouring it would make one concept appear blue / white / absent, and inside a single `-i` interaction
flip white → blue seconds apart; making it *uniformly* blue would require disabling tty echo and
writing a bespoke line editor (backspace, arrows, UTF-8/IME, bracketed paste, `Ctrl+C` cancel vs
`Ctrl+D` send, the 1 MiB bound) — a regression risk on the operator's working path, for a colour.

**Green was the next candidate and was rejected by the operator** (*"not green, it's for dollar"*), which
`internal/ui/metrics.go:38-48` confirms literally: the green on the `╰─⠿ Ready` line wraps
`green(fmt.Sprintf("$%.4f", sessionCost), …)`. Green is also the palette's busiest accent (four
meanings) and its other instances sit in the **same** `stderr` block, three lines below the ack.

**Blue** is the palette's least-loaded accent (one meaning: `[USER] - N` in the `-l` listing,
`internal/ui/listing.go:84`), it means the **same concept** (the operator's side of the conversation),
and its existing use is on a **different surface** (`stdout`, the offline listing) — the two never share
a visual block.

## Decision

**D1 — the fifth element: the input-capture acknowledgement, accented in bright blue.** On a
colour-enabled surface the **whole** line — timestamp included — is wrapped in the reference's
bright-blue SGR pair `\033[1;34m … \033[0m`. The colour is `colorBlue`; the mechanism is the existing
`blue()` / `wrap()` in `internal/ui/colour.go`. **No palette growth**: no new SGR constant and no new
wrapping helper.

**D2 — the gate is ADR 0023 D2's, unchanged.** The accent appears only when the diagnostic stream
(`stderr`) is a **terminal**, `-r` is **off**, and the surface is prompt-bearing — the single
`chromeColour` predicate (`internal/cli/cli.go:1116`), shared with the spinner. Redirected/piped
`stderr`, `-r`, the offline paths, `stdout`, and `turns.log` stay **byte-identical** to the plain form.
No second colour predicate may be introduced. **It is the *resolved* gate that is consumed:** the call site
reads the turn's once-resolved `colourOn` (`internal/cli/cli.go:808`, round 054 fold **RF-54-2** — *"resolve
the shared terminal gate ONCE per turn so the consumers cannot disagree"*), never re-running `chromeColour`
at the ack site. Re-probing would let the ack disagree with the spinner/payload/`Ready` accents **within one
turn**, which is exactly the disagreement the RF-54-2 fold exists to make impossible — so this element is a
**consumer** of the shared verdict, not a sixth opinion on it.

**D3 — the colour lives in the adapter (ADR 0023 D5's rule, applied).** `ui.Lines.InputCaptured` gains
the value receiver `(l Lines)` and returns the colour variant; the pure `FormatInputCaptured` stays the
single plain entry point (byte-identical, no signature change) and gains a `formatInputCapturedColour`
sibling — the `formatPayloadEstimateColour` / `formatToolReasonColour` shape. The bytes remain owned by
`internal/ui`.

**D4 — uniformity is the element's justification, and it is structural.** The accent lands on **every**
prompt-bearing surface — positional, piped, Ctrl+D multi-line, `-b <N> "prompt"`, and `-i` submit —
because all five reach the single `if opts.chrome` block that emits the ack. If a future change colours
only some surfaces, this ADR's premise is void and the element should be withdrawn rather than
half-applied. **The witnessing unit is the `turnOptions` construction site, not the reader label:** that
value is built in exactly **three** places — `:295` (`-i`), `:429` (positional · piped · `-b` · a
prompt-bearing `--new`), and `:479` (the Ctrl+D reader) — so one colouring Example per site is
**exhaustive** (piped and `-b` are the same construction as positional). All three are witnessed.

**D5 — under `-r` the line stays and only the colour drops.** The acknowledgement is gated on
`opts.chrome`, **not** on `opts.raw`; `-r` suppresses the accent, never the line.

**D6 — recorded divergence: the reference greens this line; tellme blues it.** `tell-me-go` passes
`colorGreen` to `printFeedback` for this exact message (`tell-me-go/internal/ui/capture.go:237-238`).
tellme adopts the reference's **gate** (its `!options.Raw && IsTTY(stderr)` is the same expression as
`chromeColour`) but **not** its colour — consistent with ADR 0023's own *Recorded divergence* clause:
*"tellme reuses the reference's **code** on **tellme's** element set."* Two further reference behaviours
are deliberately **not** adopted: the `"\r" + EL` prefix before the coloured line
(`capture.go:285-291`) — it changes **non-colour** bytes and interacts with spinner state — and the
separate `disableEscapeSequences` hatch (`-r` already suppresses colour entirely).

**D7 — modelled.** The `Chrome` entity and the `chrome-colour-terminal-gated` invariant in
`docs/domain-model/tellme.modelith.yaml` name the acknowledgement as a fifth accent (bright blue); the
rendered Markdown is regenerated with `make modelith-render` (ADR 0041 — `modelith-check` is a
`verify` member).

## Consequences

- The operator's prompt moment is marked on **every** surface by one tellme-owned line, uniformly, with
  no dependence on which reader collected the prompt.
- Blue now carries **two** meanings — `[USER] - N` (operator-authored text, `-l`) and the
  acknowledgement (the operator's channel, turn chrome). Both are "the operator's side", and they never
  appear in the same block (`stdout` offline listing vs `stderr` turn chrome). Green — the alternative —
  would have carried a **fifth**, unrelated meaning inside the same block as its `$` instance.
- Off the gate nothing changes: the plain path is byte-identical by construction (`wrap()` returns its
  input verbatim when disabled), so the E2E contract (which runs off a terminal) and the `turns.log`
  artifact are untouched.
- **Palette pressure is now recorded**, not merely felt: green 4 meanings, yellow 2, grey 2, blue 2,
  magenta 1, red 1. A future round that wants a further accent must justify it against this table.

## Forward (non-blocking)

- **RF-067-1** — the multi-line hint (`MultiLineHint` / `TUIHint`) remains **plain**; the reference
  colours its hint **yellow on `stdout`** (`tell-me-go/internal/ui/capture.go:270`). Admitting it would
  need its own decision (a different element on a different stream) and is **not** implied by this ADR.
- **RF-067-2** — the reference's `"\r" + EL` prefix is **not** adopted (D6): a `\r` + clear-line write
  before the ack would alter non-colour bytes and can race the spinner's own line-clearing. If a future
  round wants the reference's caret behaviour, it must carry its own witness.
- **RF-067-3** — blue's second meaning is accepted but **untested for theme legibility**: `1;34`
  (bright blue) is widely readable, but tellme cannot assert how a colour renders on an arbitrary
  terminal. A future terminal-profile concern would be a **new** decision, not a defect here.
- **RF-067-4** — the palette count in §Consequences is a **snapshot**: each accent-bearing round updates
  it. There is no mechanical carrier for "green now means five things" — the ADR set is the record.
- **RF-067-5** — **this round also fired ADR 0066's RF-066-1.** Running `make verify` surfaced a newly
  published stdlib advisory, **`GO-2026-6809`** (`crypto/x509`, reached via `fmt.Fprintf` →
  `x509.Certificate.Verify`), fixed only in **`go1.26.10`**; on the then-provisioned `go1.26.9` the
  `vulncheck` gate reds with **no repo change** — exactly the "recorded, not machine-checked" limit
  RF-066-1 named. Cleared per **ADR 0066 D3** (module half in `go.mod` — nothing to do, no module
  advisory; toolchain half in the environment: the user-scoped SDK is now `~/sdk/go1.26.10`), and the
  recorded floor on the gate's `specs/truth/techstack.md` row was reconciled **≥ 1.26.9 → ≥ 1.26.10**.
  ADR 0066's **body stays verbatim**; its `docs/decisions/README.md` index row gains the forward pointer.
  **The underlying limit stands:** the floor is still recorded rather than machine-checked, so it will
  drift again at the next advisory — a future `make` guard parsing `go env GOVERSION` is the standing
  (unclaimed) remedy.
