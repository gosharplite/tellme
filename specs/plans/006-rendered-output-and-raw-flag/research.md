# Phase 0 Research: tellme Rendered Output & Raw Flag (Round 006)

Topic: the round-006 output-presentation slice — render the provider's answer as formatted Markdown **by default**, add the **`-r`/`--raw`** flag, and expose the rendered **wrap width** — per Clarify Round 1 (Q1 reference parity; Q2 glamour renderer, no pty; Q3 include `WRAP_WIDTH`/`TELL_ME_WRAP_WIDTH`). The contract must match `tell-me-go`'s output behaviour.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution), testing harness (`godog` + stdlib `testing`), and the provider transport (stdlib `net/http`) were locked in rounds 001–005. The system still has **one CLI end** (the operator terminal) and adds **no new system end** and **no new external service**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the CLI Gherkin; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers.

The byte-level behaviours the spec deferred here are settled below (Decisions 1–7), each grounded in the reference source.

---

## Decision 1: Renderer = glamour — the reference's renderer

- **Decision**: Adopt `github.com/charmbracelet/glamour` (v1.0.0, matching the reference) as the Markdown→ANSI renderer for the default output path. It is the round's **first presentation dependency**; it pulls `github.com/charmbracelet/lipgloss`, `github.com/muesli/termenv`, and `github.com/muesli/reflow` **transitively** (glamour v1.0.0 `go.mod`).
- **Rationale**: Q1 (reference parity) requires the reference's rendering, and the reference's renderer *is* glamour (`internal/ui/markdown.go` → `glamour.NewTermRenderer`). No stdlib path reproduces Markdown→ANSI rendering, so a zero-dependency option cannot satisfy the parity decision. The transitive set matches the reference's presentation stack; the reference's **direct** `lipgloss` is used only by its TUI (`internal/ui/tui/**`), which tellme does not have.
- **Alternatives considered**:
  - **Zero-dependency stdlib formatting** (word-wrap only; no ANSI/Markdown styling): preserves the zero-dep discipline but cannot reproduce glamour's rendering — **partial parity**, rejected against Q1.
  - **`goldmark` + hand-rolled ANSI**: avoids glamour but re-implements its styling with worse fidelity and more code — rejected.
  - **glamour + a pty test dependency**: rejected (Q2) — the reference itself ships **no pty**; adding one diverges from its dependency footprint for behaviour it does not itself verify.

## Decision 2: Renderer build options — `WithStandardStyle` + `WithEmoji`; style from `GLAMOUR_STYLE`

- **Decision**: Build the renderer with `glamour.WithStandardStyle(resolveGlamourStyle())` followed by `glamour.WithEmoji()`, where `resolveGlamourStyle()` reads the `GLAMOUR_STYLE` environment variable and falls back to a default style when unset.
- **Rationale**: These are the reference's exact options (`internal/ui/markdown.go`), and `GLAMOUR_STYLE` is the reference's style-override mechanism (`internal/ui/renderer.go:resolveGlamourStyle`). Matching both keeps rendered output equivalent.
- **Alternatives considered**:
  - **A fixed style (no `GLAMOUR_STYLE`)** — simpler, but diverges from the reference's configurability — rejected.
  - **Explicit `WithDarkStyle`/`WithLightStyle`** — not the reference's mechanism — rejected.

## Decision 3: The render/raw gate is `-r` alone; the stdout terminal probe gates only tellme's own presentation

- **Decision**: Add a boolean `-r`/`--raw` flag (`pflag` `BoolVarP(&raw, "raw", "r", false, …)`). Rendering is suppressed **iff `-r` is set** — never by whether stdout is a terminal (FR-001/FR-002/FR-004). The **stdout** terminal probe is **NOT wired this round**: tellme ships no *own* presentation chrome (color/spinner/labels) for it to gate, so — exactly as round 005 recorded — it stays a **named pin**, and PR #16 Final-Review **Obs 1** therefore remains **OPEN** (to be wired when tellme gains its own presentation chrome). Rendering itself is gated by `-r` only.
- **Rationale**: Matches the reference exactly: `renderTextLocked(ui, part, raw)` gates rendering on `raw`; `session_manager.go` computes `UseColor = isTTY && !RawOutput`, consumed only by the renderer's own color helpers. It also **amends round-005 FR-007**: the non-terminal suppression now covers tellme's *own* presentation, not the answer's rendering — the byte-exact/plain piped stream is obtained via `-r`.
- **Alternatives considered**:
  - **A terminal gate on rendering** (auto-plain when stdout is not a terminal) — the round-005 posture; rejected for parity (Q1) and because it would re-render round 005's pipeline contract as automatic rather than `-r`-driven.
  - **No stdout probe at all** — rejected: it would leave tellme's own presentation ungated and keep PR #16 Obs 1 open.

## Decision 4: Rendered vs raw byte handling

- **Decision**:
  - **Rendered path**: render the answer through glamour, then **trim leading and trailing newlines** from the rendered result; if the result is non-empty, write it followed by a trailing `"\n\n"`.
  - **Raw path** (`-r`): write the answer text **verbatim**, followed by **exactly one CLI-appended** `"\n"` (unconditional — matches round-005 FR-006 and the DSL `is exactly` row; M2 reconciliation).
- **Rationale**: Reproduces the reference's byte behaviour (`renderMarkdownWithUILocked` trims then appends `"\n\n"`; `renderTextLocked`'s raw branch prints then `Fprintln`s; tellme appends unconditionally per round-005 FR-006). Keeping the raw path's single-terminating-newline rule preserves round-005 FR-006 for the raw mode.
- **Alternatives considered**:
  - **Write glamour output as-is (no trim/append)** — rejected: produces unstable leading/trailing blank lines that diverge from the reference.
  - **Always append `"\n"` on the raw path** — rejected: would double-terminate answers that already end in a newline and break the verbatim rule.

## Decision 5: LaTeX→Unicode sanitization before rendering

- **Decision**: Apply the reference's `sanitizeForTerminal` replacement map (e.g. `$\rightarrow$`→`→`, `$\times$`→`×`, `$\dots$`→`...`) to the answer text **before** rendering, and also on the degraded raw fallback (Decision 6).
- **Rationale**: The reference sanitizes before rendering (`renderTextLocked` → `sanitizeForTerminal`); omitting it would visibly diverge on LaTeX-bearing answers.
- **Alternatives considered**:
  - **Skip sanitization** — rejected: breaks rendered parity for common LLM LaTeX output.
  - **A different/newer replacement set** — rejected: parity means the reference's map.

## Decision 6: Graceful degradation when the renderer fails to initialize (ADR-007 parity)

- **Decision**: If `glamour.NewTermRenderer` fails to initialize, keep the turn working: fall back to **raw (unrendered)** output and emit at most **one** `[WARN] markdown rendering degraded, falling back to raw text` line on standard error; initialization failure MUST NOT fail the turn.
- **Rationale**: The reference treats glamour failures as non-fatal (ADR-007) and rate-limits the warning (`sync.Once`). Parity + robustness.
- **Alternatives considered**:
  - **Fail the turn on renderer-init failure** — rejected: a cosmetic dependency must not break reasoning.
  - **Warn on every render** — rejected: floods stderr; the reference rate-limits to once.

## Decision 7: Wrap width — `WRAP_WIDTH` config + `TELL_ME_WRAP_WIDTH` env; `>= 0`; `0` = default; rendered-only

- **Decision**: Add an integer `WRAP_WIDTH` configuration (YAML key `WRAP_WIDTH`) with a `TELL_ME_WRAP_WIDTH` environment override, honouring the existing `TELL_ME_*` **environment-over-file** precedence. Validate `>= 0` (a negative value is a configuration error — the existing configuration-error contract). Apply the width with `glamour.WithWordWrap(width)` when `width > 0`; `0`/unset uses glamour's default (80). The width applies to **rendered** output only and is ignored under `-r`.
- **Rationale**: Identical to the reference (`internal/domain/config/config.go` `WrapWidth`; `internal/infrastructure/config/config.go` `BindEnv("WRAP_WIDTH","TELL_ME_WRAP_WIDTH")`; `SetWordWrap` → `glamour.WithWordWrap`).
- **Alternatives considered**:
  - **Config-only (no `TELL_ME_WRAP_WIDTH`)** — rejected (Q3): the env override is part of the reference surface.
  - **Defer wrap width entirely** — rejected (Q3): it is part of the same rendered-output facet.
  - **Clamp a negative value instead of rejecting** — rejected: the reference validates `>= 0` and errors.

## Residual risks / forward links

- **pty verification deferred (named pin)**: the "stdout is a terminal" branch of FR-003 is **pinned, not pty-verified** (Q2 — no pty dependency; the reference does not verify it either). The harness covers the branch *logic* via stand-ins, as in round 005.
- **round-005 FR-007 amendment**: `/axb-dsl-refine` is expected to record a **MODIFY** against `specs/truth/features/cli/chat/piping-the-answer-out.feature` and its `dsl.md` rows — the "redirected answer is the answer text alone / carries no decoration" Examples move their byte-exact assertion to the `-r` path.
- **ANSI-dependent assertions**: glamour's exact output is terminal/`TERM`- and profile-dependent; acceptance assertions must key on **marker presence/absence** (e.g. the literal `**` absent when rendered, present under `-r`) rather than exact byte equality. Implementation note for `/axb-tasks` — the E2E runner needs an ANSI-aware predicate (extend, don't replace, the round-005 stdout capture).
- **go.mod growth**: the round adds `github.com/charmbracelet/glamour` (+ transitive `lipgloss`/`termenv`/`reflow`); `go mod tidy` must be run and the module graph reviewed. This is the project's first deliberate presentation dependency and supersedes the round-005 "Not Introduced Yet" note for a renderer.
- **Exit codes / class phrases**: unchanged this round except that the configuration-error path now also covers an invalid `WRAP_WIDTH`; the round-004 table `0/2/3/4/5/6` and all frozen phrases stand (FR-008/FR-009).
