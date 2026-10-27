package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/home"
)

// resolveWorkspace resolves only the runtime home + effective mode + session
// workspace (no configuration/provider), for the session commands `-l`, `-t`,
// `--new`, and `-b` that must work offline. Round 053 (ADR 0022): the mode comes
// from the `-c` configuration when the env override is unset, and an explicit
// `-c` that cannot be honoured is a resolve error. The returned resolution
// carries the home/config path so a failure renders an actionable message.
func resolveWorkspace(target sessionRef) (resolution, *resolveError) {
	res := resolution{Home: target.home, Path: target.config, Explicit: target.config != ""}
	if res.Path == "" {
		res.Path = defaultConfigPath(target.home)
	}
	if target.home == "" {
		return res, &resolveError{Reason: reasonHomeUnset}
	}
	// ONE configuration parse (round 073 fold RF-8): the mode (round 053) and the
	// rendered width (round 073) both come from the same load.
	cfg, mode, err := offlineConfigAndMode(target)
	if err != nil {
		reason := reasonConfigInvalid
		if errors.Is(err, os.ErrNotExist) {
			reason = reasonConfigMissing
		}
		return res, &resolveError{Reason: reason, Err: err}
	}
	ws, err := home.EnsureWorkspace(target.home, mode)
	res.Workspace = ws.Path
	if err != nil {
		return res, &resolveError{Reason: reasonHomeUnusable, Err: err}
	}
	// Round 073: the offline listing's rendered width, resolved best-effort from
	// the same configuration (an unreadable/invalid width degrades to 0).
	res.WrapWidth = offlineWidth(cfg)
	return res, nil
}

// historyMode resolves the effective mode for a session command: the
// TELL_ME_MODE override when set; else the `-c` configuration's MODE (or the
// default configuration's MODE when no `-c` was given); else "butler".
//
// Round 053 (ADR 0022): the `-c` path is honoured (the previous shape read only
// TELL_ME_MODE / the default config, so `-l 1 -c architect.yaml` silently read
// the default session). An explicit `-c` that cannot be read/parsed is a hard
// error (Q2 → (A)); an absent default configuration still degrades to "butler"
// (round-007 tolerance).
func historyMode(target sessionRef) (string, error) {
	_, mode, err := offlineConfigAndMode(target)
	return mode, err
}

// offlineConfigAndMode loads the configuration ONCE for the offline session
// commands and derives the effective session MODE (round 053 / ADR 0022): the
// `TELL_ME_MODE` override when set; else the `-c` configuration's MODE (or the
// default configuration's MODE when no `-c` was given); else "butler". It also
// returns the parsed configuration so the offline listing resolves its rendered
// width (round 073) WITHOUT a second parse (review fold RF-8).
//
// An explicit `-c` that cannot be read/parsed is a hard error (Q2 → (A)) UNLESS
// the env override names the mode — then the file is used for the width only,
// best-effort. An absent default configuration still degrades to "butler"
// (round-007 tolerance).
func offlineConfigAndMode(target sessionRef) (*config.Config, string, error) {
	envMode := os.Getenv("TELL_ME_MODE")
	explicit := target.config != ""
	path := target.config
	if !explicit {
		path = defaultConfigPath(target.home)
	}
	cfg, err := config.Load(path)
	if err != nil {
		if explicit && envMode == "" {
			return nil, "", err
		}
		if envMode != "" {
			return nil, envMode, nil
		}
		return nil, "butler", nil
	}
	if envMode != "" {
		return cfg, envMode, nil
	}
	return cfg, cfg.EffectiveMode(""), nil
}

// offlineWidth resolves the rendered width for the offline listing (round 073),
// reusing the round-006 resolution (config WRAP_WIDTH + the TELL_ME_WRAP_WIDTH
// override). It is BEST-EFFORT: the listing is a read-only reporter, so an
// unreadable/invalid width degrades to 0 (the renderer's built-in default)
// rather than failing the listing (a new failure mode the round refuses).
func offlineWidth(cfg *config.Config) int {
	if cfg == nil {
		return 0
	}
	width, err := cfg.EffectiveWrapWidth(os.Getenv("TELL_ME_WRAP_WIDTH"))
	if err != nil {
		return 0
	}
	return width
}

// writeAnswer writes the provider's answer to the environment's stdout
// (round-006 FR-001/FR-004): the answer bytes verbatim under -r/--raw, or the
// Markdown-rendered form by default. Rendering is gated by -r ALONE — never by
// whether stdout is a terminal. On renderer degradation the renderer's sanitized
// fallback text is written and a one-time non-class warning goes to stderr (the
// frozen `tellme: {phrase}` vocabulary is untouched).
func (e runtimeEnv) writeAnswer(answer string, raw bool, width int) {
	if raw {
		e.writeRawAnswer(answer)
		return
	}
	rendered, degraded := e.renderer.Render(answer, width)
	if degraded {
		e.renderer.WarnDegraded(e.stderr)
		e.writeRawAnswer(rendered) // research D5: the degraded fallback is the sanitized text
		return
	}
	if trimmed := strings.Trim(rendered, "\n"); trimmed != "" {
		_, _ = fmt.Fprint(e.stdout, trimmed+"\n\n")
	}
}

// writeRawAnswer prints text verbatim followed by exactly one CLI-appended
// terminating newline (round-005 FR-006). It serves both the raw (-r) path and
// the sanitized degraded fallback (round-006 research D5).
func (e runtimeEnv) writeRawAnswer(answer string) {
	_, _ = fmt.Fprintln(e.stdout, answer)
}
