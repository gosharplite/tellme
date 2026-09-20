# Technical Research — the `-h`/`-v` CLI shorthands (round 074)

**Plan Package**: `specs/plans/074-cli-help-and-version-shorthands`
**Owner**: `/axb-technical-research` (truth owner: `specs/truth/techstack.md`)
**Clarify**: resolved at specify time (**Q1 → A** `-h` **and** `--help`; **Q2 → A** help → `stdout` + exit **0**). No new question rose here.

---

## Context (grounded 2026-09-21, `dev` @ `5206f9f`)

`parseFlags` (`internal/cli/cli.go`) builds one `pflag.FlagSet("tellme", ContinueOnError)` with `SetOutput(stderr)`. Today:
- **`-h` is special-cased by pflag** (the flag is not defined): pflag calls its own `usage()` → writes `Usage of tellme:` + the flag list to **`stderr`** (the `SetOutput` writer) and returns `ErrHelp`; `parseFlags` returns `ok=false`; `run` then prints `tellme: the command-line usage is invalid` and returns the pinned usage code **2**.
- **`-v` is not a flag** → an unknown-flag parse error → the same phrase + code **2**.
- **`--version` exists** → `fmt.Fprintf(env.stdout, "tellme %s\n", version); return Success`.

Verified: `tellme -h` → the usage block on stderr + the phrase, **exit 2**; `tellme --wibble` → only the phrase, exit 2; `tellme --version` → `tellme dev` on stdout, exit 0.

---

## 決策 1 — Define both flags explicitly (do not rely on pflag's `-h` special case)

- **Decision**: register `fs.BoolVarP(&o.help, "help", "h", false, …)` and make `--version` carry the shorthand `fs.BoolVarP(&o.version, "version", "v", false, …)`; handle help in `run` (render the usage to `stdout`, return `Success`). pflag's implicit `-h` path is thereby superseded (the flag now exists, so pflag parses it normally).
- **Rationale**: the operator asked for a *successful* `-h` (Q2 → A). pflag's built-in path is unusable for that — it hard-wires the stream to the `SetOutput` writer (stderr) and the caller's `ok=false` (→ exit 2). Only an explicitly-defined flag lets tellme choose **stdout + exit 0** and keep the error path intact.
- **Alternatives considered**:
  - **Detect `pflag.ErrHelp` in `parseFlags` and special-case it** — pflag has already written the block to *stderr* by then; telling it apart from a genuine parse error and re-routing the bytes to stdout is fragile (rejected).
  - **Leave `-h` undefined and catch the error string** — brittle string matching (rejected).

## 決策 2 — The help block is the flag list, rendered by pflag (`Usage of tellme:` + `FlagUsages()`)

- **Decision**: the help output is the pflag-rendered `FlagUsages()` (pflag's own `PrintDefaults` text — one aligned line per flag, short form and long form) . We write `"Usage of tellme:\n"` + `fs.FlagUsages()` to `stdout`, and return `Success`. This is the *same text* today's error path prints to stderr (verified above) — just on the success stream and with code 0.
- **Rationale**: the operator said **"simple"**, and this is exactly the block tellme already emits — one source (`fs.FlagUsages()`), no hand-maintained list to drift. It also lists **every** flag with its short form, satisfying FR-002, and it cannot go stale when a later round adds a flag.
- **Alternatives considered**:
  - **A hand-written help string** — a second, drift-prone source for the same flags (rejected).
  - **cobra-style `Usage:`/`Flags:`/`Available Commands`** — tellme has no subcommands; adopting cobra for one block is a non-starter (rejected; also A2).

## 決策 3 — Precedence: `--help` → `--version` → `-d` → `-l` → `-t` → `--tool-usage`

- **Decision**: help is checked **first** in `run` (before the version path), then the existing order is unchanged. So `tellme -h --version` prints help and exits 0 (help wins), and `tellme -v` alone prints the version. Help stays a **prompt-less terminal action** — it never falls through to the boot/prompt path.
- **Rationale**: `--help` conventionally outranks other terminal actions (a help request is the most general "show me the surface" intent); it also matches cobra's behaviour. Both help and version are offline and prompt-less, so neither can consume a prompt or dial a provider (I-2). Keeping `--version` where it is preserves its byte-for-byte contract (I-4).
- **Alternatives considered**:
  - **Version first** — an odd surface (`-h` would be shadowed by `-v`); rejected.
  - **Fold help into `dispatchReporting`** — that function is the *offline reporting* batch; help precedes even `--version`, so it belongs in `run`'s head (rejected).

## 決策 4 — Help is a success: `stdout`, code 0, no `tellme: ` phrase

- **Decision**: help writes to **`stdout`** and returns **`Success` (0)**; it prints **no** `tellme: …` line (nothing on `stderr`). The usage-error path (unrecognized flag) is untouched: pflag's error → the phrase on `stderr` → code **2**.
- **Rationale**: clarify **Q2 → A** and the frozen phrase vocabulary (I-3): `tellme: …` is reserved for the class phrases; help is not a failure, so it must not emit one. Keeping the two apart is what makes `tellme -h` script-friendly (stdout is the requested artifact, empty stderr, exit 0).
- **Alternatives considered**:
  - **`stderr` + a dedicated help code** — clarify Q2 → B, not taken; it diverges from the reference and the norm (rejected).
  - **Help printing the phrase `the command-line usage is invalid`** — would double the class phrase for a non-error (rejected).

## 決策 5 — The change is local to `internal/cli/cli.go` (+ tests + truth); no new seam

- **Decision**: the two flags and the help render live entirely in `parseFlags` + `run` (`internal/cli`); `env.stdout` carries the help; the existing `flags` struct gains `help bool`. No new port, no deps field, no truth-model change beyond the CLI-flag truth rows.
- **Rationale**: this is a flag-surface change on one already-injected stream; a new port for "print the flag usages" would be ceremony with exactly one caller (contrast round 073, where the *listing bytes* warranted a port because `internal/ui` owns presentation).
- **Alternatives considered**:
  - **A `render.Lines`/port method for help** — the help text is pflag's own rendering of the flag set, which lives in `internal/cli`; routing it through a domain port would leak pflag into the domain (rejected).

## 決策 6 — Witnesses: an E2E per flag + a unit pin for the precedence; a mutation for the stream

- **Decision**: E2E covers `-h`/`--help` (flag list present, empty stderr, exit 0), `-v`/`--version` (version on stdout, exit 0), and the unchanged unrecognized-flag refusal (exit 2). A unit pin asserts the precedence (`-h` beats `-v`; help never reaches the boot path). The falsifiability witness mutates the help stream to `stderr`… see D7.
- **Rationale**: the acceptance-level carriers are the flags' observable contracts; the precedence/terminal-action claim is a unit-level ordering fact (a flat capture can show it, but a unit pin is cheaper and exact).

## 決策 7 — `-v` is purely additive; `--version` output is byte-identical

- **Decision**: `--version` prints `tellme {version}\n` to `stdout` and exits 0 **exactly as today**; `-v` is the same code path. The version string is the injected build version (`dev` in this repo).
- **Rationale**: I-4 (no regression). The shorthand only widens flag acceptance.
- **Alternatives considered**:
  - **A distinct `-v` output shape** — the operator asked for a *shorthand*, not a second format (rejected).

---

## Residual risks / forward items (for the ADR §Forward)

- **RF-074-1** — the help block is pflag's flag list; the richer reference-style (`Usage:`/`Flags:` sections, an example line) is **not** adopted (A2, "simple").
- **RF-074-2** — tellme stays **subcommand-free** (no `help`/`completion` commands the reference has); help is a flag, not a command.
- **RF-074-3** — the help text is derived at run time from `fs.FlagUsages()`; it has **no** dedicated golden test beyond "the flags appear" (a full-text pin would couple the test to pflag's alignment).
- **RF-074-4** — no `-V`/`--Version`/`-?` aliases; only the two shorthands the operator named.
