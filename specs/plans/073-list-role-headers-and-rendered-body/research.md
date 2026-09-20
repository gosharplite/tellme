# Technical Research — the `-l` listing presentation parity (round 073)

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`
**Owner**: `/axb-technical-research` (truth owner: `specs/truth/techstack.md`)
**Clarify**: **not escalated** (3 questions were resolved by `/axb-clarify` at specify time: Q1 → A · Q2 → A · Q3 → B). No new question rose here — every residual was technical.

---

## Context (grounded 2026-09-21, `dev` @ `7ab84f6`)

`renderHistoryList` (internal/cli/cli.go:1048) today emits `fmt.Fprintf(env.stdout, "%s: %s\n", m.Role, m.Content)` per message from `toMessages` (`user` / `assistant`, no header, no rendering, no separator). The three operator-locked changes (role heads, glamour body, blank separator) plus the clarifications (`-r` honoured; operator-messages-only; only the model body rendered) mean the listing needs a **presentation owner** that can (a) render Markdown, (b) gate colour, (c) be driven hermetically. `internal/cli` may not import `internal/ui` (the layer-discipline gate, ADR 0011), so the seam matters.

---

## 決策 1 — A new domain port `render.Listing` for the listing, not a reuse of `render.Answer`

- **Decision**: add `render.Listing` (+ `ListingMessage`, `ListingRole`, `ListingSpec`) to `internal/domain/render/ports.go`, implemented by a new `internal/ui/listing.go` adapter (`ui.NewListing()`), injected through `deps.Dependencies.NewListing func() render.Listing`. `internal/cli` assembles `[]render.ListingMessage` from `toMessages` and calls the port.
- **Rationale**: the bytes of every presentation surface are **single-owned** in `internal/ui` (round 051 / ADR 0020); `internal/cli` may not import it (ADR 0011 RULE-B/E). `render.Answer` is insufficient — it renders one blob with no role/colour/separator contract — and widening it would give the answer path a knob it never uses. A dedicated port keeps the role header, the colour gate, and the separator owned by one adapter and hermetic-testable in isolation.
- **Alternatives considered**:
  - **Reuse `render.Answer` + do the header/separator in `internal/cli`** — the CLI would emit `[USER]`/`[MODEL]` and the escapes/sanitization policy, splitting one presentation contract across two layers (rejected: violates the byte-ownership rule and re-opens the layer baseline).
  - **A func-typed seam (`NewListing func(...) string`)** — workable (the `LoopFactory`/`ProgressFactory` precedent) but a single-method interface reads as the `Lines`/`Answer` family and keeps the signature evolvable; chosen the interface.

## 決策 2 — The role is a domain enum, not the stored `user`/`assistant` string

- **Decision**: the port carries `ListingRole` (`ListingOperator`, `ListingModel`) and the CLI maps `user → ListingOperator`, `assistant → ListingModel`. The header text (`[USER]`/`[MODEL]`) is produced by the adapter.
- **Rationale**: the operator asked for `[USER]`/`[MODEL]`; tellme persists `prompt`/`answer` and maps them to `user`/`assistant`. Passing the stored/derived wire role into the presentation port would leak a *storage* naming into a *presentation* contract and would make the `[MODEL]` divergence invisible in the port. A domain enum makes the names the adapter's business and the mapping one explicit table.
- **Alternatives considered**:
  - **Pass `"user"/"assistant"` strings** — the adapter uppercases them, yielding `[ASSISTANT]`, not the operator-requested `[MODEL]` (rejected).
  - **A `bool isOperator`** — collapses if a third role (a future system/thought block) ever appears (rejected).

## 決策 3 — A dedicated `stdout` terminal probe + a `TELL_ME_FORCE_STDOUT_TTY` seam

- **Decision**: add `stdoutTTY func(any) bool` to `runtimeEnv` (built by a new `stdoutTerminalDetector()` reading `TELL_ME_FORCE_STDOUT_TTY`, mirroring the stderr detector), plus a `runtimeEnv.stdoutIsTerminal()` accessor. The listing's colour gate is `env.stdoutIsTerminal() && !raw`.
- **Rationale**: tellme's colour gates are per-stream and honest — the chrome gate is the **stderr** probe, and the listing writes to **stdout**, so its gate must be the stdout probe (the reference's own axis: `useColor = isTTY(stdout) && !rawOutput`). A dedIcated probe + seam keeps the three-stream model uniform (stdin/stderr already have one each) and lets the E2E assert the colour on/off branches **without a pty**. Reusing the shared `isTTY` probe would tie the listing colour to the `TELL_ME_FORCE_STDIN_TTY` seam — a misleading cross-axis coupling.
- **Rationale (Obs 1)**: this wires a **stdout terminal probe for the listing colour only**. The round-006 / PR #16 **Obs 1** ("tellme's own *stdout chrome*" on a prompt turn) is a **different** behaviour (no stdout chrome is emitted on the prompt path) and **stays OPEN**; this round neither adds nor closes it. Recorded so the two are not conflated.
- **Alternatives considered**:
  - **Reuse `env.isTTY(env.stdout)`** (the stdin-oriented probe/seam) — the E2E would force `TELL_ME_FORCE_STDIN_TTY`, conflating the stdin seam with the stdout gate (rejected).
  - **No probe ⇒ colour always on** — would inject escapes into redirects (violates I-3) (rejected).
  - **Gate on the stderr probe** — wrong stream; a stderr-terminal with a piped stdout would colour a redirect (rejected).

## 決策 4 — Colour codes: add the reference's `colorBlue` / `colorMagenta`

- **Decision**: add `colorBlue = "\033[1;34m"` (operator header) and `colorMagenta = "\033[1;35m"` (model header) to `internal/ui/colour.go`, reusing the existing `wrap(s, code, enabled)` (empty never wrapped; disabled ⇒ verbatim).
- **Rationale**: the reference's palette is the source (`tell-me-go/internal/ui/colors.go`); the element set is tellme's own (a recorded divergence, per the round-054 precedent for the green accents). `wrap` already gives the empty-safe / byte-plain-when-off rule the listing needs.
- **Alternatives considered**:
  - **Plain 8-colour blue/magenta (`\033[0;34m`)** — diverges from the reference's bright variants the operator named ("blue/magenta") (rejected).
  - **A per-call inline SGR literal** — bypasses the shared `wrap` invariant (rejected).

## 決策 5 — One blank line after **every** message (the reference's shape)

- **Decision**: the adapter writes, per message: the header line, the body normalised to end in exactly one `\n`, then one additional `\n` (the separator). So the listing ends with a blank line, and consecutive messages are separated by exactly one blank line.
- **Rationale**: the reference emits `Fprintln` after **every** content block (`tell-me-go/internal/ui/history.go`) — uniform, trivially correct, and deterministic. It cannot produce a double gap even when a body already ends in `\n` (the normalisation), satisfying the spec's "no double gap" edge case.
- **Alternatives considered**:
  - **Separator *between* messages only** — a conditional "not the last" branch for a cosmetic saving (rejected: a second code path, and the acceptance's "answer's body separated from the end" reads either way).
  - **No normalisation (raw body + `\n`)** — a body ending in `\n` would yield a double blank (rejected: the edge case).

## 決策 6 — The model body reuses the one Markdown renderer (glamour) + the one width resolution

- **Decision**: the adapter holds a `ui.Renderer` (`NewRenderer()`) and calls `Render(body, width)` — the same `GLAMOUR_STYLE` resolution, the same LaTeX sanitizer, the same `WithWordWrap(width)`, and the same degraded fallback (sanitized raw text) as the answer path. The width is the **same resolved** `WRAP_WIDTH`/`TELL_ME_WRAP_WIDTH` value. On degrade the adapter writes the same one-time `[WARN] markdown rendering degraded, falling back to raw text` line to the injected diagnostic writer.
- **Rationale**: the operator asked for "rendered as Markdown by glamour"; there must be **one** rendering policy, not two (FR-009). Reusing the type + the resolution keeps the listing's bytes identical in spirit to the answer path and adds no dependency (glamour is already direct).
- **Alternatives considered**:
  - **A second renderer/config** — a second style/width policy to keep in sync (rejected).
  - **Print the answer's already-rendered bytes** — the render never runs on the `-l` path (there is no turn), so nothing is cached to reuse (rejected).

## 決策 7 — The width is resolved **best-effort** on the listing path (no new failure)

- **Decision**: `renderHistoryList` resolves the width via the existing `config.Load` + `cfg.EffectiveWrapWidth(env)` composition, but **best-effort**: on any load/validate error it falls back to `width = 0` (the renderer's built-in default). The mode resolution (round 053) is unchanged and remains the only hard requirement.
- **Rationale**: `-l` is an offline reader whose only hard config dependency is the mode (round 053 / ADR 0022). Today an unreadable **default** config already degrades to `butler` (`historyMode`), but an unreadable **explicit** `-c` is a hard error. Turning a *width* problem into a hard `-l` failure would be a **new failure mode** the operator did not ask for (I-1: the listing stays a read-only reporter). Best-effort honours `WRAP_WIDTH` when it is readable and never blocks the listing otherwise.
- **Alternatives considered**:
  - **Hard-fail on an invalid width** — the prompt path's behaviour, but it would make `-l` refuse to list on a broken `WRAP_WIDTH` where it lists today (rejected: a new failure mode).
  - **Ignore the width (always 0)** — the operator's rendering request would silently ignore `WRAP_WIDTH` (rejected).

## 決策 8 — `-r` is honoured by the listing (a third raw consumer)

- **Decision**: the listing reads the existing `-r`/`--raw` flag; when set, the model body is emitted **verbatim** (no glamour, no LaTeX sanitizer) and **no** colour is emitted; the operator body is verbatim in both cases. The flag's help text gains the listing (it currently reads as answer-only).
- **Rationale**: clarify **Q1 → A** (reference parity); the reference's `-l` is `-r`-aware. It also keeps the in-group relay recipe (`… -l 1 -r -c "<target>.yaml"`) byte-plain — the round-053 rationale that motivated round 053 itself.
- **Alternatives considered**:
  - **`-r` irrelevant to `-l`** (the pre-round status quo) — the relay recipe would start receiving Markdown/ANSI (rejected: Q1 → A).

## 決策 9 — The acceptance is carried by the E2E over a forced-stdout-terminal seam + unit pins on the adapter

- **Decision**: the E2E drives the listing through `TELL_ME_FORCE_STDOUT_TTY` (colour positive) and a plain pipe (colour negative); the deterministic per-message **byte** shape (header / rendered body / separator) is pinned in `internal/ui` unit tests and in the E2E over a small arranged history. The existing round-007/053 `-l` scenarios are re-pointed at the new per-message bytes.
- **Rationale**: the suite captures `stdout` as a pipe (never a terminal), so a colour positive requires the seam (the round-019 `TELL_ME_FORCE_STDERR_TTY` precedent). The per-message bytes are the contract; a flat capture can and should pin them.
- **Alternatives considered**:
  - **A pty harness** — out of scope (the repo is pty-free by design) (rejected).
  - **Only a unit pin** — would leave the new truth rows uncarried end-to-end (rejected: `acceptance-coverage`).

---

## Residual risks / forward items (for the ADR §Forward)

- **RF-073-1** — the reference renders **both** role bodies; tellme renders only the model body (Q3 → B). A future "render the prompt too" request is a one-line change gated on this record.
- **RF-073-2** — the reference's `No history found.` empty-session sentence is **not** adopted (A2); the listing stays silent on an empty session.
- **RF-073-3** — the reference's `-l N "prompt"` composability (list **and** chat) and its `-b/--back` pairing stay unadopted (A1).
- **RF-073-4** — the listing's tool-activity omission (Q2 → A) keeps the round-007 operator-only contract; a future `[Tool Call]`/`[Tool Response]` parity request would supersede it.
- **RF-073-5** — the `stdout` probe is wired for the listing colour only; the answer-path **Obs 1** (stdout chrome) stays OPEN (決策 3).
- **RF-073-6** — best-effort width resolution (決策 7): a broken `WRAP_WIDTH` silently degrades to the renderer default on the `-l` path.
