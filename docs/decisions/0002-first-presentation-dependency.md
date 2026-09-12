# ADR 0002 — First presentation dependency: glamour

- **Status:** Accepted
- **Date:** 2026-09-12
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** round 006 (`specs/plans/006-rendered-output-and-raw-flag`); PR #18 architectural review **M3**;
  `tell-me-go` (`internal/ui/markdown.go` = `glamour.NewTermRenderer`)

## Context

Rounds 001–005 kept tellme near-zero-dependency: stdlib plus `spf13/pflag`, `gopkg.in/yaml.v3`, and
`cucumber/godog`, with a deliberate preference for stdlib over helpers (e.g. `os.ModeCharDevice`
instead of `golang.org/x/term`; `net/http` instead of provider SDKs; a hand-rolled `${VAR}` expander).

Round 006 delivers **reference output parity**: by default `tellme` renders the provider's answer as
formatted Markdown, with `-r`/`--raw` as the plain-text inverse (Clarify Q1). The reference renders
Markdown→ANSI with **glamour**. There is **no stdlib path** that reproduces glamour's Markdown→ANSI
rendering, so the parity decision (Q1) forces a renderer dependency.

## Problem

Tellme has no presentation dependency and no precedent for adding one. Introducing glamour pulls a
non-trivial transitive tree, which — against a near-zero-dependency project — deserves an explicit,
recorded decision rather than an incidental `go get`.

## Decision

Adopt **`github.com/charmbracelet/glamour` (v1.0.0)** — the reference's renderer — as tellme's first
presentation dependency, for the default output path only.

- **Build options:** `glamour.WithStandardStyle(resolveGlamourStyle())` (style from the `GLAMOUR_STYLE`
  env, default `dark`) + `glamour.WithEmoji()`, matching the reference (`internal/ui/markdown.go`).
- **Word wrap:** `glamour.WithWordWrap(width)` when the resolved `WRAP_WIDTH` / `TELL_ME_WRAP_WIDTH`
  is `> 0`; `0`/unset uses glamour's built-in default (80).
- **Posture:** the raw (`-r`) path bypasses glamour entirely; the byte-exact/plain contract is the
  raw path (round-005 FR-007 amended: non-terminal suppression covers tellme's **own** chrome, not
  the answer's rendering).
- **Degradation (ADR-007 parity):** a glamour init/render failure is non-fatal — the sanitized raw
  text is emitted, with a single **non-class** `[WARN]` line on stderr (it deliberately does **not**
  use the reserved `tellme: ` prefix, so the frozen class-phrase vocabulary is untouched).
- **Transitive footprint:** glamour v1.0.0 pulls `lipgloss`, `termenv`, `reflow`, `chroma/v2`,
  `bluemonday`, `goldmark`, `gorilla/css`, `dlclark/regexp2`, `x/net`, `x/text` (~20 modules). The
  vulnerable `goldmark`/`x/text` versions in that tree were bumped to fixed releases
  (`goldmark v1.7.17`, `x/text v0.39.0`) so `govulncheck` stays clean.
- **pty non-adoption:** no pty dependency is added; the "stdout is a terminal" branch stays a **named
  pin** (the reference ships no pty either).

## Alternatives considered

1. **Zero-dependency stdlib formatting** (word-wrap only, no ANSI/Markdown styling) — preserves the
   near-zero-dependency discipline but cannot reproduce glamour's rendering → **partial parity**;
   rejected against the Q1 parity decision.
2. **`goldmark` + hand-rolled ANSI** — avoids glamour but re-implements its styling with worse
   fidelity and more code; rejected.
3. **glamour + a pty test dependency** — rejected (Q2): the reference ships no pty; adding one
   diverges from its footprint for behaviour it does not itself verify.

## Consequences

- tellme now carries a real presentation dependency tree; future dependency-hygiene work must track
  glamour's transitive tree (notably the `goldmark`/`x/net` families) for vulnerabilities and updates.
- The default output is no longer byte-identical to the answer; the **raw** (`-r`) path is the
  byte-exact contract, and rendered output is a presentation zone (outside the byte-determinism
  contract).
- The renderer is isolated behind the `answerRenderer` seam in `internal/cli` (the concrete
  `*ui.Renderer` lives in `internal/ui`), so the mode selection and width resolution remain
  unit-testable without glamour.
- Immutable once `Accepted`; a future renderer change supersedes this ADR rather than editing it.
