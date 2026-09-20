# ADR 0046 — The `-h`/`--help` and `-v`/`--version` CLI shorthands

**Status**: Accepted (round 074)

**Date**: 2026-09-21

**Related**: round 006 (the flag surface / `-r`) · round 053 / ADR 0022 (the offline session commands) · round 002 (the pinned exit-code contract; the `--json` removal) · `specs/truth/features/cli/usage/unsupported-cli-usage.feature` (the usage-error contract) · `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` (the `--version` contract) · the reference (`tell-me-go` is cobra-based and exposes `-h, --help` / `-v, --version`).

## Context

tellme ships one `pflag.FlagSet` with no `-h`/`--help` and only the long `--version`. The operator compared the flag surface with the reference and asked: *"Please add simple `-h` and `-v` flags for tellme."*

Grounded behaviour before the round (`dev` @ `5206f9f`, verified on the built binary):

- `tellme -h` → pflag's **implicit** help path (the flag is undefined): it writes `Usage of tellme:` + the flag list to the `SetOutput` writer (**`stderr`**) and returns `ErrHelp`; `run` then prints `tellme: the command-line usage is invalid` and exits **2**.
- `tellme -v` → an unknown-flag parse error → the same phrase, exit **2**.
- `tellme --version` → `tellme dev` on **`stdout`**, exit **0**.

Two clarify questions were settled (one at a time):

- **Q1 → A** — `-h` ships **with** the long form **`--help`** (`-h, --help`), matching the reference.
- **Q2 → A** — help writes to **`stdout`** and exits **0** (a successful action, like `--version`).

## Decision

1. **Two explicit flags.** `parseFlags` registers `fs.BoolVarP(&o.help, "help", "h", …, false)` and upgrades `--version` to `fs.BoolVarP(&o.version, "version", "v", …, false)`. Defining `-h` explicitly **supersedes pflag's implicit `-h` special case**, so tellme controls the stream and the exit code (D1).

2. **The help block is pflag's own flag list.** On `-h`/`--help`, tellme writes `"Usage of tellme:\n"` + `fs.FlagUsages()` to **`stdout`** and returns `Success`. This is the text pflag's **implicit help path** wrote to the `SetOutput` writer (`stderr`) *before this round* (`tellme -h` → exit 2) — one source (`fs.FlagUsages()`), so it lists **every** flag with its short form and cannot drift when a later round adds a flag (D2). *(The unrecognized-flag path is a different one: it prints only the class phrase, no block — see §Context.)*

3. **Precedence: `--help` → `--version` → `-d` → `-l` → `-t` → `--tool-usage`.** Help is checked **first** in `run`; then the existing order is unchanged. Both help and version are **offline, prompt-less, terminal** actions — they never fall through to the boot/prompt path, read no stdin, and dial no provider. `tellme -h --version` prints help and exits 0 (help wins) (D3). **One exception, and it is the right one:** a **parse error pre-empts help** — `tellme -h -z` (an unrecognized flag) fails in `parseFlags` *before* `run`, so it refuses with the class phrase on `stderr` and exit **2**, not help. The precedence above applies only among the flags the parser accepts.

4. **Help is a success; the error path is untouched.** Help emits **no** `tellme: …` line (nothing on `stderr`) and exits **0**. The frozen class-phrase vocabulary is unchanged: an **unrecognized** flag still refuses with `tellme: the command-line usage is invalid` on **`stderr`** and exit **2** (the pinned usage code) (D4, I-1/I-3).

5. **`-v` is purely additive.** `-v` runs the same code path as `--version`: `tellme {version}\n` on `stdout`, exit **0**. The long form's output is byte-identical to before (I-4).

6. **Scope: one file + tests + truth.** The change is local to `internal/cli/cli.go` (`parseFlags`, `run`, the `flags.help` field); no new port, no deps field, no domain-model change — the help text is pflag's rendering of the flag set, which lives in `internal/cli` (D5). No new dependency; stdlib + the existing pflag; POSIX-only; hermetic (I-5).

## Consequences

- `tellme -h` / `tellme --help` are script-friendly: the flag list is the requested artifact on `stdout`, `stderr` is empty, exit 0. `tellme -v` matches `--version`.
- tellme **stays subcommand-free** — the reference's `help`/`completion` commands are **not** adopted; help is a flag, not a command (RF-074-2).
- The help block is the **flag list** — tellme has no subcommands, so the reference's `Available Commands`/`completion` sections are absent (its first line *is* `Usage of tellme:`, matching the reference's own `Usage:` line); the richer reference prose rendering is **not** adopted ("simple", A2) (RF-074-1).
- The help text is derived at run time from `fs.FlagUsages()`; there is **no** dedicated golden test beyond "the flags appear" — a full-text pin would couple the test to pflag's alignment (RF-074-3).
- No `-V`/`--Version`/`-?` aliases — only the two shorthands the operator named (RF-074-4).

## Forward (non-blocking)

- **RF-074-1** — a richer help block (reference parity: `Usage:`/`Flags:` sections, an example line); not adopted (A2).
- **RF-074-2** — the reference's `help` / `completion` subcommands; not adopted (tellme stays subcommand-free).
- **RF-074-3** — no full-text golden pin for the help block (only "the flag names appear").
- **RF-074-4** — no `-V` / `--Version` / `-?` aliases.
- **RF-074-5** — `flags.helpText` caches the rendered flag list on the parse result; the cleaner shape is to hand the `*pflag.FlagSet` (or a render func) back to `run` and render there (review RF-1; non-blocking, single caller).
