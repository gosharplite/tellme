package steps

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// scenarioKey is the private context key under which a scenario's state is stored.
type scenarioKey struct{}

// scenarioContext carries the isolated per-scenario state:
//   - a fresh temporary runtime home (the TELL_ME_HOME stand-in);
//   - the environment overrides / unsets a scenario arranges;
//   - the CLI arguments and the captured result of the last tellme invocation;
//   - bookkeeping for the workspace reuse assertion.
//
// It is created once per scenario by beforeScenario and torn down by afterScenario.
type scenarioContext struct {
	home    string // the actual resolved runtime home directory (temp dir)
	homeSet bool   // whether TELL_ME_HOME should be exported for a run

	envOverrides map[string]string // env vars to set for the next run
	envUnset     map[string]bool   // env vars to remove for the next run

	args []string // CLI arguments for the next run (excluding the binary)

	exitCode int    // captured process exit code
	stdout   string // captured stdout
	stderr   string // captured stderr
	runErr   error  // error from starting/waiting the process (nil on clean run)

	wsIno uint64 // inode of a pre-existing workspace dir (reuse assertion)
	wsSet bool   // whether wsIno was recorded
}

// beforeScenario creates an independent scenarioContext backed by a fresh temp
// runtime home, and injects it into the scenario context. TELL_ME_MODE and
// TELL_ME_SELECTED_PROVIDER are unset by default so the ambient shell cannot
// leak into a scenario; a step may arrange them explicitly.
func beforeScenario(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	dir, err := os.MkdirTemp("", "tellme-e2e-")
	if err != nil {
		return ctx, err
	}
	sc := &scenarioContext{
		home:         dir,
		homeSet:      true,
		envOverrides: map[string]string{},
		envUnset:     map[string]bool{"TELL_ME_MODE": true, "TELL_ME_SELECTED_PROVIDER": true},
		args:         nil,
		exitCode:     0,
		stdout:       "",
		stderr:       "",
		runErr:       nil,
	}
	return context.WithValue(ctx, scenarioKey{}, sc), nil
}

// afterScenario removes the scenario's temporary runtime home. Godog calls this
// after every scenario, regardless of outcome.
func afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if sc := scenarioFrom(ctx); sc != nil && sc.home != "" {
		_ = os.RemoveAll(sc.home)
	}
	return ctx, nil
}

// scenarioFrom returns the scenario's state, or nil if none is attached.
func scenarioFrom(ctx context.Context) *scenarioContext {
	sc, _ := ctx.Value(scenarioKey{}).(*scenarioContext)
	return sc
}

// setEnv arranges an environment override for the next run.
func (sc *scenarioContext) setEnv(name, value string) {
	if sc.envOverrides == nil {
		sc.envOverrides = map[string]string{}
	}
	sc.envOverrides[name] = value
	delete(sc.envUnset, name)
}

// runEnv returns the environment override map for the next run (including
// TELL_ME_HOME when the home is set).
func (sc *scenarioContext) runEnv() map[string]string {
	env := make(map[string]string, len(sc.envOverrides)+1)
	for k, v := range sc.envOverrides {
		env[k] = v
	}
	if sc.homeSet {
		env["TELL_ME_HOME"] = sc.home
	} else {
		delete(env, "TELL_ME_HOME")
	}
	return env
}

// unsetNames returns the environment names to remove for the next run.
func (sc *scenarioContext) unsetNames() []string {
	names := make([]string, 0, len(sc.envUnset)+1)
	for k := range sc.envUnset {
		names = append(names, k)
	}
	if !sc.homeSet {
		names = append(names, "TELL_ME_HOME")
	}
	return names
}

// run executes the currently arranged command and records the captured result.
func (sc *scenarioContext) run() {
	res := harness.Run(sc.args, sc.runEnv(), sc.unsetNames())
	sc.exitCode = res.ExitCode
	sc.stdout = res.Stdout
	sc.stderr = res.Stderr
	sc.runErr = res.Err
}

// homePath resolves a home-relative path (e.g. a {config_path}) under the home.
func (sc *scenarioContext) homePath(rel string) string {
	return filepath.Join(sc.home, filepath.FromSlash(rel))
}

// workspacePath resolves an operator-facing {home}-rooted path (e.g.
// "ait-tmg/output/butler") to the actual directory under the scenario home. The
// leading path segment is the {home} name and is replaced by the actual home.
func (sc *scenarioContext) workspacePath(op string) string {
	trimmed := strings.TrimPrefix(filepath.ToSlash(op), "/")
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		trimmed = trimmed[i+1:]
	} else {
		trimmed = ""
	}
	return filepath.Join(sc.home, filepath.FromSlash(trimmed))
}

// expectedWorkspace is the workspace the product should resolve for the
// scenario's effective mode (TELL_ME_MODE env, else "butler").
func (sc *scenarioContext) expectedWorkspace() string {
	mode := "butler"
	if m, ok := sc.envOverrides["TELL_ME_MODE"]; ok && m != "" {
		mode = m
	}
	return filepath.Join(sc.home, "output", mode)
}

// writeFile writes content at a home-relative path, creating parent dirs.
func (sc *scenarioContext) writeFile(rel string, content []byte) error {
	p := sc.homePath(rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, content, 0o644)
}

// inodeOf returns the inode of a path (Linux), or 0 when unavailable.
func inodeOf(path string) (uint64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return st.Ino, nil
	}
	return 0, nil
}

// wellFormedConfig returns a resolvable YAML configuration whose MODE is `mode`
// and whose SELECTED_PROVIDER `selected` is present in its PROVIDERS registry.
func wellFormedConfig(mode, selected string) string {
	return "MODE: " + mode + "\n" +
		"PERSON: \"e2e persona\"\n" +
		"SELECTED_PROVIDER: " + selected + "\n" +
		"PROVIDERS:\n" +
		"  " + selected + ":\n" +
		"    TYPE: deepseek\n" +
		"    MODEL: deepseek-v4-flash\n" +
		"    URL: https://api.deepseek.com\n"
}

// emptyRegistryConfig returns a YAML configuration whose PROVIDERS registry is
// empty (selected-provider validation must fail).
func emptyRegistryConfig(mode string) string {
	return "MODE: " + mode + "\n" +
		"PERSON: \"e2e persona\"\n" +
		"SELECTED_PROVIDER: deepseek-flash\n" +
		"PROVIDERS: {}\n"
}

// networkCapabilityViolation reports the first network-capability indicator in
// the built binary, if any (the build-graph capability guard). It delegates to
// the leaf harness so the guard has a single implementation.
func networkCapabilityViolation() (string, error) {
	return harness.NetworkCapabilityViolation()
}

// unresolvedCategories are the pinned reason categories the diagnostic reports
// for an unresolved setup (specs/truth/features/cli/diagnostics/dsl.md).
var unresolvedCategories = []string{"config-missing", "config-invalid", "provider-mismatch", "home-unset", "home-unusable"}

// blockedRun reruns the current command under a hostile network environment so a
// caller can assert the outcome is unchanged (differential no-egress witness).
// The hostile-env definition lives once in the leaf harness.
func (sc *scenarioContext) blockedRun() harness.RunResult {
	return harness.RunWithBlockedNetwork(sc.args, sc.runEnv(), sc.unsetNames())
}
