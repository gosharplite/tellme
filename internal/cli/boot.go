package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/home"
)

// sessionRef names a command's session target (round 098; ADR 0068): the
// TELL_ME_HOME override and the `-c` configuration path. Nearly every offline
// command carries this pair, and `homeDir, configPath string` was adjacent-
// identical on the ones the parameter-shape gate reds — including renderTurn's
// three-string `homeDir, configPath, prompt` head.
type sessionRef struct {
	home   string
	config string
}

// resolve is the single resolution algorithm shared by the boot, diagnostic, and
// turn paths: home → config path → load/validate → effective selected provider →
// effective mode → workspace. On failure it returns a *resolveError carrying the
// pinned reason category; the callers differ only in how they render it.
func resolve(target sessionRef) (resolution, *resolveError) {
	res := resolution{Home: target.home, Path: target.config, Explicit: target.config != ""}

	// Step 1 — resolve TELL_ME_HOME first, always (FR-006).
	if target.home == "" {
		return res, &resolveError{Reason: reasonHomeUnset}
	}

	// Step 3 — the config path: -c when given, else the default for the mode seed.
	if !res.Explicit {
		res.Path = defaultConfigPath(target.home)
	}

	// Step 4 — load + validate the file (FR-002, FR-005).
	cfg, err := config.Load(res.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return res, &resolveError{Reason: reasonConfigMissing, Err: err}
		}
		return res, &resolveError{Reason: reasonConfigInvalid, Err: err}
	}

	// Step 4b — resolve and validate the rendered width (round-006 FR-006): the
	// helper owns resolve+validate, so a non-integer override or a negative value
	// from either source is a configuration error. This step runs before provider
	// resolution, so a config that is both width-invalid and provider-invalid
	// reports the width error (recorded precedence).
	width, werr := cfg.EffectiveWrapWidth(os.Getenv("TELL_ME_WRAP_WIDTH"))
	if werr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: werr}
	}
	res.WrapWidth = width

	// Step 4c — resolve the tool-loop bound (round-008 FR-006): MAX_TOOL_LOOP
	// (env/config, default 1000). A non-integer or negative value is a
	// configuration error.
	maxLoops, lerr := cfg.EffectiveMaxToolLoop(os.Getenv("MAX_TOOL_LOOP"))
	if lerr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: lerr}
	}
	res.MaxToolLoop = maxLoops

	// Step 4d — resolve the payload budget (round-009 FR-010): MAX_HISTORY_TOKENS
	// (env/config, default 1000000). A non-integer or negative value is a
	// configuration error.
	budget, bErr := cfg.EffectiveMaxHistoryTokens(os.Getenv("MAX_HISTORY_TOKENS"))
	if bErr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: bErr}
	}
	res.MaxHistoryTokens = budget

	// Step 4d.1 — expand ${VAR} / ${VAR:-default} in the MCP_SERVERS string
	// fields, best-effort (round-032 SC-002 / issue #67): a `TOKEN:
	// "${GITHUB_TOKEN}"` entry authenticates instead of being sent literally (and
	// then warn+skipped). An unresolved ${VAR} keeps its literal text AND emits a
	// non-fatal diagnostic warning naming the field/variable, so the cause is
	// visible instead of an opaque "could not be reached"; the load never fails.
	// Expansion runs BEFORE validation so the validator sees the resolved values.
	res.MCPWarnings = append(res.MCPWarnings, cfg.ExpandMCPServers()...)

	// Step 4e — validate the MCP_SERVERS registry (round-032 FR-002/FR-013): a
	// malformed REMOTE entry reuses the configuration-invalid class phrase (a
	// stable, classed failure); a COMMAND (stdio) entry is warn+skipped, surfaced
	// on the prompt path. The registry is carried on the resolution so the
	// prompt-path discovery reads the exact validated set.
	if mcpVal, mcpErr := cfg.ValidateMCPServers(); mcpErr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: mcpErr}
	} else {
		res.MCPWarnings = append(res.MCPWarnings, mcpVal.Warnings...)
	}
	res.MCPServers = cfg.MCPServers

	// Step 5 — the effective selected provider must be in the registry (FR-003).
	res.Selected = cfg.EffectiveSelectedProvider(os.Getenv("TELL_ME_SELECTED_PROVIDER"))
	if !cfg.ProviderInRegistry(res.Selected) {
		return res, &resolveError{Reason: reasonProviderMismatch}
	}

	// Step 5b — resolve the selected provider entry: expand ${VAR} placeholders
	// FIRST, then validate the RESOLVED state (FR-001..FR-009).
	prov := cfg.Providers[res.Selected]
	if err := prov.Expand(); err != nil {
		return res, &resolveError{Reason: reasonProviderInvalid, Err: err}
	}
	if err := prov.Validate(); err != nil {
		return res, &resolveError{Reason: reasonProviderInvalid, Err: err}
	}
	cfg.Providers[res.Selected] = prov
	res.Provider = prov
	res.Pricing, res.Priced = cfg.PricingFor(prov.Model)

	// Step 5c — the run-static effective budget (round 024 FR-015): the tool
	// resource bound cap, min(MAX_HISTORY_TOKENS, the model's configured context
	// window). With no window configured it is MAX_HISTORY_TOKENS (so the payload
	// line and the bound are byte-identical to the model-blind default). The
	// window is the optional per-model MODELS.<model>.CONTEXT_WINDOW.
	res.EffectiveBudget = res.MaxHistoryTokens
	if window, ok := cfg.ContextWindowFor(prov.Model); ok {
		if window < res.EffectiveBudget {
			res.EffectiveBudget = window
		}
	} else {
		logNoWindowOnce(prov.Model)
	}

	// Step 6 — effective mode + prepare the session workspace (FR-007/008/009).
	res.Person = cfg.Person
	res.Mode = cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))
	workspace, err := home.EnsureWorkspace(target.home, res.Mode)
	res.Workspace = workspace.Path
	if err != nil {
		return res, &resolveError{Reason: reasonHomeUnusable, Err: err}
	}
	return res, nil
}

// renderBoot runs the boot path and reports readiness or an actionable error.
func renderBoot(target sessionRef, env runtimeEnv) int {
	res, rerr := resolve(target)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	_, _ = fmt.Fprintln(env.stdout, "configuration: ready")
	_, _ = fmt.Fprintln(env.stdout, "session workspace: "+res.Workspace)
	return Success
}
