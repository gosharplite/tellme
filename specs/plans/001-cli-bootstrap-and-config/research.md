# Phase 0 Research: tellme CLI Bootstrap & Configuration (Round 001)

Topic: how to build the narrow foundation slice — CLI boot, YAML configuration load & validation,
runtime home (`TELL_ME_HOME`) + per-mode session workspace resolution, build version (`--version`),
and an offline setup diagnostic (`-d` / `-d --json`).

Scope note: the language (**Go**) is fixed by `spec.md`; this research settles the toolchain and test
stack that `spec.md` explicitly deferred to this phase. The three AIxBDD must-ask decisions were
answered by the user before writing: **system ends = one CLI end**, **BDD techstack = godog**,
**test strategy = E2E (black-box CLI)** — captured below as Decisions 4–5.

> **Review status (grill round 001 + clarify round).** These decisions were pressure-tested in an
> adversarial review — issue [#1](https://github.com/gosharplite/tellme/issues/1), verdict **proceed
> with changes** (full transcript gist:
> <https://gist.github.com/gosharplite/0b6775f580c4ec66ada984e1bbe023d6>) — and the corrections are
> folded in below, flagged **⟦grill R1⟧**. The two PM-boundary items the grill routed were then
> resolved by **`/axb-clarify`** (2026-09-10), flagged **⟦clarify⟧**: **Q1 → Option 3** (the `-d`
> diagnostic contract in Decision 2) and **Q2 → Option 1** (the round's "no `data/**` truth" scope
> lock is **released** — `/axb-data-plan` authors a minimal data truth; see the Clarify round section). **⟦grill #3 + user ratification, 2026-09-11: REVERSED — round 001 owes no `data/**` model; `/axb-data-plan` = `NOOP`.**

## Decision 1: Go module & project layout

- **Decision**: Go 1.26 with module path `github.com/gosharplite/tellme`; entrypoint at `cmd/tellme/`,
  and all non-public logic under `internal/` — specifically `internal/config`, `internal/home`,
  `internal/cli`. **⟦grill R1⟧ `internal/version` is dropped** (see rationale).
- **Rationale (⟦grill R1⟧ revised)**: the layout is justified by **separation of concerns**, not by
  unit-testability (Decision 5 is E2E-only, so "unit-testable from day one" would be self-contradictory).
  Each package owns exactly one FR cluster and therefore one **E2E-observable** responsibility, which is
  what the acceptance set actually asserts:
  - `internal/config` → FR-001/002/003/005, NFR-001; observable via the config-error exit code
    (`starting-with-a-configuration.feature`).
  - `internal/home` → FR-006/007/008/009; observable via workspace create/reuse and the
    environment-error code (`runtime-home-and-session-workspace.feature`).
  - `internal/cli` → FR-001 flag parsing + FR-014 exit-code classification + FR-011/013 diagnostic
    dispatch; observable via usage errors (`unsupported-cli-usage.feature`) and `-d`.
  "Unit-testable when the strategy broadens" survives only as a **forward benefit**, explicitly
  subordinate to Decision 5 — not the driver.
- **⟦grill R1⟧ Version is a single value, not a package.** `var version = "dev"` in
  `cmd/tellme/main.go` is the **only** version symbol and the single `-X main.version` target
  (Decision 6). A `version` package has no consumer beyond printing one string and would split the
  value across two packages (`main.version` injected, `internal/version` formatting), inviting a
  second source of truth — and it has no benchmark precedent (`tell-me-go` keeps `version` in `main`).
  **Invariant:** the injected variable is the only version symbol; any future `version` package must
  *move* the injection target (never introduce a second).
- **Alternatives considered**:
  - Flat single `main` package: simplest for one file, but does not survive the next slice.
  - A `pkg/` tree for "shared" code: premature — nothing in this module is consumed by another yet.

## Decision 2: CLI argument parsing

- **Decision**: `spf13/pflag` for GNU-style flags (`-c, --config`, `-d`, `--json`, `--version`),
  wrapped by a small `internal/cli` parser that classifies errors into the distinct exit codes required
  by FR-014 (success / usage / configuration / environment).
- **Rationale**: The spec mandates a `-c`/`--config` shorthand+long pair, which stdlib `flag` can only
  fake by double registration; pflag is the idiomatic minimal solution and is the flag layer cobra is
  built on, so adopting cobra later needs no parser rewrite. Separating parsing from `main` makes the
  usage-error path (unknown flag → distinct exit code; pflag must run `ContinueOnError` and map the
  returned error rather than its default `os.Exit(2)`) unit-testable.
- **⟦grill R1⟧ ⟦clarify⟧ Diagnostic (`-d`) dispatch — RATIFIED (clarify Q1 → Option 3).** Parsing
  precedes diagnostic dispatch: an unrecognized flag is still a usage error even with `-d`
  (`-d --wibble` ⇒ usage code). After a successful parse the classifier runs an explicit **`-d`-first
  branch** that resolves-and-reports instead of boot's fail-fast. The ratified contract:
  - `-d` is a **reporting path** — it always produces the report, whether resolution **succeeded or
    failed**. It emits resolution status (`resolved`, or `unresolved` + category ∈ {`config-missing`,
    `config-invalid`, `provider-mismatch`, `home-unset`, `home-unusable`}) to stdout, and the health
    signal in `--json` (FR-013). **The verdict lives in the exit code; the detail lives in the body** —
    mirroring `tell-me-go`'s `RunDiagnostics`, which renders the report *before* verdicting.
  - **Exit codes**: `-d` exits **0 when resolution succeeded**; exits a **dedicated, distinct non-zero
    "diagnostic: unresolved" code when a report was produced but resolution failed**; and a further
    distinct code if the report itself could not be produced.
  - **FR-014 scope (resolved by the PM):** the four `FR-014` codes (success / usage / configuration /
    environment) bind the **boot** path only. The `-d` codes are **additive** — `FR-014`'s "at least"
    is satisfied by *adding* a code rather than silently scoping one out.
  - No network on **any** `-d` path, including the failure path (FR-012/SC-004) — see Decision 5.
  - Why this and not the alternatives: **Option 1** (`-d` exits 0 on an unresolved setup) was rejected
    — the benchmark's `-d` encodes the verdict in the exit code (non-zero when not healthy), so exit-0
    would contradict it. **Option 2** (fail-fast, reuse boot codes, no report) was rejected — it is raw
    benchmark parity with the benchmark's own flaw (`-d` cannot diagnose the failures it exists for).
    Option 3 keeps the report while preserving the exit-code-as-verdict semantics.
  - The exact `--json` key schema remains an open item.
- **Alternatives considered**:
  - stdlib `flag`: zero dependencies, but no native shorthand/long pairing and awkward unknown-flag
    handling.
  - `spf13/cobra`: full subcommand framework — heavier than round 001 needs (no subcommands yet);
    deferred until `browse`/`retry` subcommands arrive.

## Decision 3: YAML configuration loading & validation

- **Decision**: Parse with `gopkg.in/yaml.v3` into a typed `Config` struct, then validate in hand-written
  code; resolve `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER` precedence explicitly in our own resolver
  (FR-003 / FR-007 / FR-015).
- **Rationale**: Configuration is a slice-local input (not truth); a thin, deterministic parser keeps the
  offline guarantees (NFR-001 / NFR-003) airtight and keeps env-over-file precedence explicit and
  testable, instead of hidden inside a framework's binding logic.
- **⟦grill R1⟧ Resolver contract (6-step ordered algorithm).** The precedence seam is the substance of
  this decision, not an `/axb-tasks` footnote; a resolver whose order is undefined is not "explicit".
  1. Resolve `TELL_ME_HOME` **first, always** — even when `-c` is given, because the workspace lives
     under it. Unset/unusable ⇒ environment error (FR-006; spec edge case).
  2. `REQ_MODE = TELL_ME_MODE` if set, else unset.
  3. Config path = `-c` if given, else `$TELL_ME_HOME/configs/${REQ_MODE:-butler}.yaml`.
  4. Load + validate the file (FR-002, FR-005).
  5. Effective mode = `REQ_MODE` if set, else the file's `MODE`, else `butler`.
  6. Workspace = `$TELL_ME_HOME/output/<effective mode>/` (FR-007).

  Refinements: (i) step 1 precedes step 3 unconditionally (home is needed on both the `-c` and default
  paths); (ii) step 3's `${REQ_MODE:-butler}` is a **filename seed only** — it never determines the
  workspace.
  - **Divergence rule (legal, non-fatal).** The default filename is only a discovery seed; the effective
    mode is authoritative from env-then-file, so filename and workspace mode may legitimately differ —
    e.g. `TELL_ME_MODE` unset and `configs/butler.yaml` declaring `MODE: coder` ⇒ file found at
    `configs/butler.yaml` while workspace = `output/coder`. This is **not an error**: FR-007 is
    authoritative ("the `TELL_ME_MODE` environment variable when set, otherwise the configuration's
    `MODE` value"), nothing requires basename == `MODE`, and erroring would reject a valid configuration.
    A non-blocking warning on divergence is deliberately **off** (bootstrap path stays silent). The
    divergence window is exactly the unset-env case; when `REQ_MODE` is set, seed == effective mode.
  - **MODE-absent fallback.** An empty/absent file `MODE` ⇒ `butler` (FR-005 lists `MODE` but does not
    require it non-empty; consistent with `spec.md`'s default-to-`butler` assumption).
- **Alternatives considered**:
  - `spf13/viper`: the benchmark's choice (env binding, watchers) — heavier, and it introduces implicit
    env coupling and non-determinism we do not want on a bootstrap path this round.
  - `sigs.k8s.io/yaml` (JSON-semantics): convenient for k8s-style schemas; unnecessary here.

## Decision 4: BDD techstack — Gherkin runner (must-ask Q2)

- **Decision**: `godog` (Cucumber for Go) — **adopted this round; first instantiated in
  `/axb-dsl-refine`**. It executes **nothing** this round: there is no `go.mod`, no Go code, and its
  input (`specs/truth/features/**`) does not exist until `/axb-dsl-refine` runs.
- **Rationale**: the project's truth model makes interface Gherkin the *executable* contract, so a runner
  that genuinely executes committed interface `.feature` files (via step definitions defined from
  `dsl.md`) is the choice that avoids drift — chosen explicitly over hand-translation.
- **⟦grill R1⟧ Executed vs carried.**
  - **Executed** (verbatim, by godog): the **interface Gherkin** under `specs/truth/features/**`
    (dsl-refine output), run E2E against the built `tellme` binary.
  - **Carried** (never executed directly): the plan-side **acceptance** journeys
    (`features/acceptance/**`). `acceptance-coverage` requires every acceptance rule to be carried by
    ≥1 `InterfaceFeature`; dsl-refine *re-expresses* acceptance rules into atomic interface scenarios +
    `DSLRow`s. (This drops the earlier "not re-translated" implication: dsl-refine *does* re-express.)
- **⟦grill R1⟧ Harness contract (pinned now — `/axb-tasks` documents it, does not invent it).**
  - Runner/step-definition location: a dedicated Go test package at **`tests/e2e/`** (benchmark parity —
  `tell-me-go`'s E2E lives there), running `godog.TestSuite` with
  `Paths: []string{"../../specs/truth/features"}`. The concrete module subpaths
  (`features/{interface}/{module}/…`) remain dsl-refine's to fix; the *package* and loading mechanism are
  decided here.
  - (a) Build once per suite: `go build -ldflags "-X main.version=<VERSION>" -o <tmp>/tellme ./cmd/tellme`;
  the binary path is passed to the suite (e.g. `TELLME_BIN`).
  - (b) Per-scenario isolation: each scenario gets a fresh `TELL_ME_HOME` (temp dir), with per-scenario
  `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER`, torn down after — no scenario shares filesystem state.
  - (c) Determinism: assert exit code + stdout + stderr + filesystem effects; forbid `time.Sleep` for
  synchronization (mirror `tell-me-go` ADR-036 / `verify-no-test-sleep`).
- **Alternatives considered**:
  - No runner — hand-translate Gherkin into stdlib `testing` table tests: fewer deps, but forfeits
    "Gherkin is executable" and invites drift.
  - `gauge`: viable, but godog is the Cucumber-native Go choice.

## Decision 5: Test strategy — E2E black-box CLI (must-ask Q3)

- **Decision**: End-to-end — build the `tellme` binary and invoke it as a subprocess, asserting exit
  code, stdout, stderr, and filesystem effects (workspace created on first run, reused thereafter).
  godog features drive the built binary; stdlib `testing` is only the host harness.
- **Rationale**: every round-001 scenario and SC-001…SC-003 is defined at the process boundary (exit
  code / streams / workspace), so black-box invocation under a controlled environment proves them.
- **⟦grill R1⟧ SC-004 (zero network) is proven by the HOST HARNESS, not by black-box observation.**
  The earlier "black-box proves everything" claim was overstated. Corrected claim: black-box
  *observation* cannot witness an **absence** claim — `And tellme performs no network access`
  (in `version-and-setup-diagnostic.feature`) is not observable in exit code / stdout / stderr /
  filesystem. The proof is a **host-harness assertion**, not a product observation:
  1. **Primary — no-egress sandbox (differential witness).** Run the built `tellme -d` and
     `tellme -d --json` with network egress blocked and assert the expected exit code + byte-identical
     structured output, so a path that *requires* network produces a genuine **FAIL** (non-zero exit /
     changed output / timeout). Two mechanisms, in preference order:
     - **Privileged netns** — an emptied network namespace (Linux `unshare -n`; macOS `sandbox-exec`
       deny-network profile). **⟦verified 2026-09-10⟧ NOT available on the local dev host**: `unshare -n`
       and `unshare -rn` both fail with `Operation not permitted`, so this is a CI / privileged-Linux
       mechanism.
     - **Unprivileged fallback (portable)** — a **hostile network environment**: an unroutable DNS
       resolver (`resolv.conf` → loopback/blackhole) plus `HTTP(S)_PROXY` pointed at a closed port, so
       any real egress fails fast. No privileges required; usable on the local host.
     **⟦clarify⟧** both cover the ready case (exit 0) **and** the `-d` **failure path** (the dedicated
     non-zero "unresolved" code), per Decision 2.
  2. **Backstop — build-graph capability guard.** A `verify`-style Makefile gate
     (`go list -deps ./cmd/tellme` / `go tool nm`) asserting no network-capable package (`net`,
     `net/http`, provider SDKs) is in the diagnostic binary's dependency closure. This closes the case the
     sandbox cannot: a best-effort call that swallows errors still *links* a network package, so
     capability-absence is decisive and platform-independent.
  3. **Epistemic grade (stated honestly).** The sandbox is a *necessary-condition / differential*
     witness — it can produce a real FAIL, it cannot *prove* absence. The build-graph guard supplies the
     *capability-absence* witness. Neither alone is sufficient; together they are. On the local dev host
     the privileged netns is unavailable (above), so local SC-004 rests on the **unprivileged hostile-env
     fallback + the build-graph guard**; if no differential mechanism can run, that witness is
     **SKIPPED — never passed green-by-skip** (the build-graph guard still holds).
- **Alternatives considered**:
  - Unit-first: cheaper, but does not exercise the acceptance journeys.
  - White-box injection (an injected network client that fails the test if invoked): rejected — it forces
    the product to carry a DI/network seam this round has not committed to (the exact "testing internals
    locks in choices" risk), whereas a sandbox manipulates the environment *around* the unchanged binary.
  - Mixed (E2E + unit for the resolver): revisit in a later slice once resolver logic grows beyond what
    the CLI boundary can observe — recorded as a forward risk.

## Decision 6: Build & version injection

- **Decision**: `go build -ldflags "-X main.version=$(VERSION)" -o tellme ./cmd/tellme`; a Makefile
  `build` target with `VERSION ?= dev`. `--version` prints the linked value.
- **Rationale**: FR-010 needs the build version at runtime, and the version is only known at build time;
  `-ldflags -X` is the standard, zero-runtime-cost mechanism (same approach as the benchmark Makefile).
- **⟦grill R1⟧ Falsifiable `--version` assertion.** Both ends default to `dev`, so a build where `-X`
  silently missed — the very binding mistake that matters — still prints `dev`, and an assertion against
  `dev` is a **false green on the one bug the test exists to catch**. The suite therefore builds with a
  **distinctive sentinel** `VERSION=0.0.0-harness` (unrelated to `dev`) and the `--version` step asserts
  **that exact string** on stdout. Single injection target is explicitly **`main.version`**. Each failure
  mode becomes a genuine FAIL: (i) wrong `-X` target → `dev` ≠ sentinel; (ii) dropped ldflag → `dev`;
  (iii) hard-coded source constant ≠ sentinel; (iv) stale binary from a build that dropped the flag.
  Optional hardening: a per-run `0.0.0-harness-<runid>` (ULID); the default stays the fixed sentinel
  (deterministic, reusable). `VERSION ?= dev` is only the local default — the harness overrides it and
  the release path sets `VERSION=x.y.z`, so `dev` never appears in an asserted run. The falsifiable
  assertion lives at the **interface/step-def level**, not in the business-worded acceptance feature.
- **Alternatives considered**:
  - `go:embed` / generated file: extra build step and source churn for no gain.
  - Hard-coded constant: cannot reflect the actual build (and is now explicitly caught by the sentinel).

## Decision 7: Development tooling (fmt / vet / lint)

- **⟦grill R1⟧ Adoption rule (replaces "disproportionate to surface").** Adopt a gate now **iff** it is
  either **(a) requirement-proving** — it directly proves a round-001 requirement — or **(b) mechanism-only**
  — a direct analyzer that runs on its default analysis with **no curated policy artifact** to author.
  - **Adopt (b):** `gofmt` (via `go fmt`), `go vet` (toolchain-native), and **`staticcheck`** — the direct
    analyzer supplying `unused` (U1000) plus the SA/S classes over the three new dependencies
    (`pflag`, `yaml.v3`, `godog`) that `go vet` does not watch; it runs on defaults with no config file.
  - **Adopt (a):** the no-network sandbox + build-graph guard (Decision 5) and `verify-no-test-sleep`
    parity (Decision 4) — **stated plainly: these are adopted because they *prove SC-004* (and
    determinism), not because the surface is rich.**
  - **Defer, with reasons:** `golangci-lint` and `govulncheck` are deferred **not** for install cost
    (both are present in `$GOPATH/bin`) but because each requires a **curated governance artifact** —
    `.golangci.yml` linter selection for the former, a triage/policy posture for the latter — a next-slice
    policy decision. Named as the intended next-slice aggregator + security gate (benchmark parity:
    `tell-me-go`'s `lint:` is `golangci-lint run ./...` with `.golangci.yml`, bundled into `check`).
  - **Named gap (honesty):** `staticcheck` does **not** cover `errcheck`'s unchecked-error class, and
    `errcheck` is not installed here. Since the round does file/YAML I/O (`os.ReadFile`, `yaml.Unmarshal`,
    `MkdirAll`), unchecked errors are a live (if small) risk — recorded as an explicit round-001 residual,
    closed next slice by `golangci-lint` with `errcheck` enabled.
  - **Environment caveat:** the Makefile must resolve these tools explicitly (`command -v` + a
    `go run`/`go install` fallback, as `tell-me-go` does) because they live in `$GOPATH/bin` and may not
    be on `PATH`; `staticcheck` becomes a documented one-line prereq
    (`go install honnef.co/go/tools/cmd/staticcheck@latest`).
- **Alternatives considered**:
  - Adopt `golangci-lint` now: stronger, but requires authoring a `.golangci.yml` policy artifact now —
    deferred as above.
  - No lint/verify target at all: forfeits a cheap correctness gate (rejected — `staticcheck` is free).

## Truth-hygiene rule (⟦grill R1⟧)

`techstack.md` records decisions that are **both settled and technological**. A three-state discipline
keeps truth honest:
1. **Settled + present** → in `techstack.md`, plainly.
2. **Decided, not yet instantiated** → in `techstack.md` **only** under an explicit bucket
   ("Adopted, Not Yet Instantiated"), so a reader never reads aspiration as exercise (e.g. `godog`).
3. **Proposed / unratified** → **never** in truth; it lives here in plan-side `research.md` until ratified.
   (The `-d` exit-code semantics, Decision 2, were *proposed* here and are now **ratified** — see below.)

## Clarify round (2026-09-10) — PM-boundary rulings ⟦clarify⟧

Grill round #1 routed two PM-boundary items to `/axb-clarify`; both were ratified before
`/axb-system-analysis`:

- **Q1 → Option 3 — `-d` diagnostic semantics.** `-d` always reports (resolved *or* unresolved) and
  exits a **dedicated non-zero "diagnostic: unresolved"** code on failure; the `FR-014` codes bind the
  boot path only (Decision 2).
- **Q2 → Option 1 — `/axb-data-plan`.** **⟦grill #3 + user ratification, 2026-09-11: REVERSED — `data-model.dbml` was authored then deleted within the round; the round owes no `data/**` model and `/axb-data-plan` = `NOOP`. This ruling is kept only as history.⟧** The round's "no `data/**` truth" scope lock is **released**;
  the owner authors a minimal data truth — the config input contract (`configs/<mode>.yaml`:
  `MODE`/`PERSON`/`SELECTED_PROVIDER`/`PROVIDERS` + `TELL_ME_*` precedence) and the workspace/state
  lifecycle (`output/<mode>/`).

Both rulings are **behaviour/scope**, not technology — so neither changes `techstack.md` (they land in
spec/acceptance and `data/**`, which are PM / truth-owner territory).

## Residual risks / forward links

- **⟦grill R1⟧ SC-004 is proven by the host harness** (no-network sandbox + build-graph capability guard)
  — *decided*, not open (supersedes the earlier "black-box proves everything"). The godog runner package
  and harness contract are now pinned at `tests/e2e/`; only the concrete module subpaths remain
  dsl-refine's.
- **⟦grill R1⟧ No net-new user value this round.** Round 001 carries no user-value acceptance: the
  executable-Gherkin / `acceptance-coverage` thesis is first exercised on **infrastructure contracts**
  (config resolution, exit-code taxonomy, harness provenance), not product behaviour — unproven on user
  value until a later round. The horizontal foundation slice is *permitted but not affirmed* by the
  methodology. **Candidate smallest vertical addition (PM-owned proposal, no spec edit):** a mutating
  operator journey **`tellme init`** — scaffold `$TELL_ME_HOME/configs/<mode>.yaml` + the
  `output/<mode>/` workspace — turning "tellme configured itself" into "tellme set me up"; a lighter
  read-only alternative is **`tellme config show`** (print the effective mode/provider/home/workspace),
  which overlaps `-d` and so adds less. Adding user value is a PM decision, not an RD one.
- **⟦grill R1⟧ ⟦clarify⟧ ⟦grill #3⟧ `/axb-data-plan` — RULED, then REVERSED.** ⟦grill #3 + user ratification (2026-09-11): `data-model.dbml` was authored then deleted; round 001 owes **no** `data/**` model; `/axb-data-plan` = `NOOP`. (Original ruling:) The round's "no `data/**`
  truth" scope lock is **released**. `aixbdd-tmg/README.md`'s CLI rule triggers `/axb-data-plan` for a
  CLI that manages persistent configuration (`configs/<mode>.yaml`) and system-owned state
  (`output/<mode>/`); the benchmark (`tell-me-go`) likewise persists and models these surfaces. The
  `/axb-data-plan` owner (delegated by `/axb-system-analysis`) authors a **minimal** data truth — the
  config input contract + the workspace/state lifecycle — which later slices extend. Research surfaces
  and routes; it does **not** write `specs/truth/data/**`.
- **⟦grill R1⟧ Unchecked-error coverage** (Decision 7) is an explicit round-001 residual.
- **⟦clarify⟧ `-d` / FR-014 scope — RESOLVED** (clarify Q1 → Option 3; Decision 2). The `FR-014` codes
  bind the boot path; `-d` gets its own "diagnostic: unresolved" code and always emits the report.
  **PM follow-ups remain** (spec/acceptance are PM-owned): add an edge case + an acceptance Example for
  `-d` on a broken/unresolved setup, and update `spec.md` to reflect the released data lock (Q2) — **done (PM-3), then further reversed by grill #3: `spec.md` now states there is no `data/**` model.**
- **Acceptance-coverage gap (PM-routed).** No acceptance Example exercises the no-`-c` + `MODE≠butler`
  default-path discovery (FR-007 already defines the requirement; the missing artifact is an acceptance
  Example → `/axb-spec-by-example`, and downstream an `InterfaceFeature`). RD raises it; the PM authors it.
- Exact exit-code numeric values (only distinctness is required) and the `-d --json` output schema remain
  open items in `STATUS.md`; implementation decisions for `/axb-tasks`/`/axb-implement`.
- `pflag` vs `cobra` should be revisited when subcommands (`browse`, `retry`) arrive.
- E2E is sufficient for round 001; revisit "mixed" when a later slice adds logic not observable at the
  CLI boundary.
