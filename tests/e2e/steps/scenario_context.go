package steps

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
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

	workDir string // the child process working directory (holds working-directory file fixtures)

	envOverrides map[string]string // env vars to set for the next run
	envUnset     map[string]bool   // env vars to remove for the next run

	args []string // CLI arguments for the next run (excluding the binary)

	stdin    string // scripted standard input for the next run (round 005)
	stdinSet bool   // whether a scripted stdin should be piped to the child
	// pacedStdin, when pacedSet, delivers the scripted input in two chunks with a
	// pause between (round 023) so bubbletea paints the editor frame into the
	// capture before the terminal key clears it (the -i witness).
	pacedStdin [2]string
	pacedSet   bool
	// stdinDevNull wires the next run's stdin to the null device — a
	// non-terminal character device (the round-012 review B1 E2E pin).
	stdinDevNull bool

	scriptedAnswer    string // the answer scripted on the fake provider (decoded), for the decoration check (grill Q4/Q6)
	scriptedAnswerSet bool   // whether scriptedAnswer was recorded
	scriptedTool      string // the tool the fake was scripted to request (round 010 ordering witness)

	exitCode int    // captured process exit code
	stdout   string // captured stdout
	stderr   string // captured stderr
	// merged holds the MERGED (stdout+stderr) capture of the last ordering run
	// (round 010 cross-stream witness). It is captured lazily by the ordering
	// Then steps and does not disturb the separate stdout/stderr fields.
	merged string
	runErr error // error from starting/waiting the process (nil on clean run)

	wsIno uint64 // inode of a pre-existing workspace dir (reuse assertion)
	wsSet bool   // whether wsIno was recorded

	// fakes holds every in-process fake provider the scenario started; they are
	// closed by afterScenario. fakeByProvider maps a provider name to its fake so
	// a Then can target the provider that must (or must not) have been hit.
	fakes          []*fakeprovider.Provider
	fakeByProvider map[string]*fakeprovider.Provider

	// lastPrompt is the prompt the last prompt-bearing run carried (When), so a
	// Then can assert the outbound request carried it.
	lastPrompt string

	// arrangedExchanges is the conversation the fixture Given arranged into the
	// session history (round 007), so a Then can compute the expected listing.
	arrangedExchanges []exchange

	// previousEstimate is the pre-flight estimate recorded by the round-011
	// `a previous run …` Given, so a Then can compare the current run against it.
	previousEstimate    int
	previousEstimateSet bool
}

// exchange is one arranged prompt/answer pair (round 007).
type exchange struct {
	prompt string
	answer string
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
	work, err := os.MkdirTemp("", "tellme-work-")
	if err != nil {
		return ctx, err
	}
	sc := &scenarioContext{
		home:         dir,
		workDir:      work,
		homeSet:      true,
		envOverrides: map[string]string{},
		// Unset every environment override the CLI honours by default, so an
		// ambient shell value cannot leak into a scenario (hermetic E2E — e.g. a
		// developer shell exporting TELL_ME_WRAP_WIDTH). A scenario that needs one
		// arranges it explicitly via setEnv, which removes it from this set.
		envUnset: map[string]bool{
			"TELL_ME_MODE":              true,
			"TELL_ME_SELECTED_PROVIDER": true,
			"TELL_ME_WRAP_WIDTH":        true,
			"MAX_TOOL_LOOP":             true,
			"MAX_HISTORY_TOKENS":        true,
			"TELL_ME_FORCE_STDIN_TTY":   true,
			"TELL_ME_FORCE_STDERR_TTY":  true,
			"TELL_ME_TUI_DEBOUNCE":      true,
		},
		args:     nil,
		exitCode: 0,
		stdout:   "",
		stderr:   "",
		runErr:   nil,
	}
	return context.WithValue(ctx, scenarioKey{}, sc), nil
}

// afterScenario removes the scenario's temporary runtime home. Godog calls this
// after every scenario, regardless of outcome.
func afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if sc := scenarioFrom(ctx); sc != nil {
		for _, f := range sc.fakes {
			f.Close()
		}
		if sc.home != "" {
			_ = os.RemoveAll(sc.home)
		}
		if sc.workDir != "" {
			_ = os.RemoveAll(sc.workDir)
		}
	}
	return ctx, nil
}

// scenarioFrom returns the scenario's state, or nil if none is attached.
func scenarioFrom(ctx context.Context) *scenarioContext {
	sc, _ := ctx.Value(scenarioKey{}).(*scenarioContext)
	return sc
}

// pipeStdin arranges a scripted standard input for the next run (a pipe, so the
// CLI sees a non-terminal and reads it).
func (sc *scenarioContext) pipeStdin(content string) {
	sc.stdin = content
	sc.stdinSet = true
}

// devNullStdin arranges the next run's standard input to be the null device — a
// character device that is NOT a terminal (the round-012 review B1 E2E pin).
func (sc *scenarioContext) devNullStdin() {
	sc.stdinDevNull = true
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
	var res harness.RunResult
	switch {
	case sc.stdinDevNull:
		res = harness.RunInWithDevNull(sc.workDir, sc.args, sc.runEnv(), sc.unsetNames())
	case sc.pacedSet:
		res = harness.RunInWithPacedStdin(sc.workDir, sc.args, sc.pacedStdin[0], sc.pacedStdin[1], sc.runEnv(), sc.unsetNames())
	case sc.stdinSet:
		res = harness.RunInWithStdin(sc.workDir, sc.args, sc.stdin, sc.runEnv(), sc.unsetNames())
	default:
		res = harness.RunIn(sc.workDir, sc.args, sc.runEnv(), sc.unsetNames())
	}
	sc.exitCode = res.ExitCode
	sc.stdout = res.Stdout
	sc.stderr = res.Stderr
	sc.runErr = res.Err
	sc.merged = "" // defensive: drop any merged capture from a prior run
}

// captureMerged runs the currently arranged command with stdout and stderr
// MERGED into one ordered buffer (the round-010 cross-stream ordering witness)
// and stores the merged bytes in sc.merged. It deliberately does NOT disturb the
// separate stdout/stderr/exitCode capture, so the presence assertions keep
// working.
//
// To leave NO trace on the scenario state — the re-run auto-resumes and appends
// the session history, and a future mutating tool would touch the working
// directory — the merged run executes against COPY-ONLY copies of the home and
// working directory, and each fake is restored to its pre-capture script cursor
// and recorded-request count afterwards. A later Then therefore sees the same
// history, files, and RequestCount() as before the capture.
func (sc *scenarioContext) captureMerged() {
	homeCopy, err := os.MkdirTemp("", "tellme-merged-home-")
	if err != nil {
		return
	}
	defer func() { _ = os.RemoveAll(homeCopy) }()
	workCopy, err := os.MkdirTemp("", "tellme-merged-work-")
	if err != nil {
		return
	}
	defer func() { _ = os.RemoveAll(workCopy) }()
	if err := copyTree(sc.home, homeCopy); err != nil {
		return
	}
	if err := copyTree(sc.workDir, workCopy); err != nil {
		return
	}

	// Snapshot each fake, re-arm its script cursor so the merged run replays the
	// same scripted exchange, and restore it afterwards (no request-count leak).
	snaps := make([]fakeSnapshot, len(sc.fakes))
	for i, f := range sc.fakes {
		served, requests := f.Snapshot()
		snaps[i] = fakeSnapshot{served: served, requests: requests}
		f.Reset()
	}
	defer func() {
		for i, f := range sc.fakes {
			f.Restore(snaps[i].served, snaps[i].requests)
		}
	}()

	env := sc.runEnv()
	env["TELL_ME_HOME"] = homeCopy
	var res harness.RunResult
	if sc.stdinSet {
		res = harness.RunInMergedWithStdin(workCopy, sc.args, sc.stdin, env, sc.unsetNames())
	} else {
		res = harness.RunInMerged(workCopy, sc.args, env, sc.unsetNames())
	}
	sc.merged = res.Stdout
}

// fakeSnapshot is a fake provider's restorable state (round-010 merged capture).
type fakeSnapshot struct {
	served   int
	requests int
}

// copyTree copies the directory tree src into dst (files + subdirectories),
// preserving file modes. It is the isolation primitive for the merged capture.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
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

// newFake starts an in-process fake provider owned by the scenario (closed by
// afterScenario).
func (sc *scenarioContext) newFake() *fakeprovider.Provider {
	f := fakeprovider.Start()
	sc.fakes = append(sc.fakes, f)
	return f
}

// registerFake records the fake backing a named provider so a Then can assert
// whether that provider was (or was not) contacted.
func (sc *scenarioContext) registerFake(provider string, f *fakeprovider.Provider) {
	if sc.fakeByProvider == nil {
		sc.fakeByProvider = map[string]*fakeprovider.Provider{}
	}
	sc.fakeByProvider[provider] = f
}

// writeDefaultConfig writes the default configuration for the effective mode
// (configs/butler.yaml) selecting `selected`, with each provider pointing at the
// given endpoint.
func (sc *scenarioContext) writeDefaultConfig(selected string, urls map[string]string) error {
	return sc.writeFile("configs/butler.yaml", []byte(fakeprovider.ConfigYAML(selected, urls)))
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

// unresolvedCategories are the pinned reason categories the diagnostic reports
// for an unresolved setup (specs/truth/features/cli/diagnostics/dsl.md).
var unresolvedCategories = []string{"config-missing", "config-invalid", "provider-mismatch", "home-unset", "home-unusable", "provider-invalid"}

// blockedRun reruns the current command under a hostile network environment so a
// caller can assert the outcome is unchanged (differential no-egress witness).
// The hostile-env definition lives once in the leaf harness.
func (sc *scenarioContext) blockedRun() harness.RunResult {
	return harness.RunInWithBlockedNetwork(sc.workDir, sc.args, sc.runEnv(), sc.unsetNames())
}

// writeWorkFile writes a working-directory fixture the child can read (e.g. a
// file for the read_files tool). The name is relative to the child's working
// directory.
func (sc *scenarioContext) writeWorkFile(name, content string) error {
	p := filepath.Join(sc.workDir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), 0o644)
}

// historyDir returns the session workspace directory the session commands
// resolve to in the E2E scenarios (no configuration, no TELL_ME_MODE → the
// default mode "butler").
func (sc *scenarioContext) historyDir() string {
	return filepath.Join(sc.home, "output", "butler")
}

// historyFilePath is the active session-history file the fixture writes.
func (sc *scenarioContext) historyFilePath() string {
	return filepath.Join(sc.historyDir(), "history.jsonl")
}

// historyArchivePath is the archived session-history file (`--new`).
func (sc *scenarioContext) historyArchivePath() string {
	return filepath.Join(sc.historyDir(), "history.archive.jsonl")
}

// recordExchange records an arranged exchange so a listing Then can compute the
// expected messages.
func (sc *scenarioContext) recordExchange(prompt, answer string) {
	sc.arrangedExchanges = append(sc.arrangedExchanges, exchange{prompt: prompt, answer: answer})
}

// onlyFake returns the scenario's single fake provider, or nil when none.
func (sc *scenarioContext) onlyFake() *fakeprovider.Provider {
	if len(sc.fakes) == 0 {
		return nil
	}
	return sc.fakes[0]
}

// setPersona writes PERSON into the default configuration (configs/butler.yaml),
// which the provider Given must have created (round-011 Given: persona config).
func (sc *scenarioContext) setPersona(persona string) error {
	return setPersonaAt(sc.home, persona)
}

// setPersonaAt writes PERSON into the default configuration under home.
func setPersonaAt(home, persona string) error {
	p := filepath.Join(home, "configs", "butler.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("the default configuration must exist before setting the persona: %w", err)
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	cfg["PERSON"] = persona
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, out, 0o644)
}

// previousRunEstimate runs the current prompt once against a throwaway copy of
// the runtime home with the given persona, parses the pre-flight `~<n>` estimate,
// and stores it on the scenario (round-011 Given: a previous run …). The real
// home (history, config) is left untouched.
func (sc *scenarioContext) previousRunEstimate(persona, prompt string) error {
	homeCopy, err := os.MkdirTemp("", "tellme-prev-home-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(homeCopy) }()
	// Snapshot every fake so this arrange-run leaves NO request-count / script
	// trace (round-011 TD-3, mirroring the round-010 merged-capture witness).
	snaps := make([]fakeSnapshot, len(sc.fakes))
	for i, f := range sc.fakes {
		served, requests := f.Snapshot()
		snaps[i] = fakeSnapshot{served: served, requests: requests}
	}
	defer func() {
		for i, f := range sc.fakes {
			f.Restore(snaps[i].served, snaps[i].requests)
		}
	}()

	// Copy FIRST, then set the persona on the copy — the real home stays
	// untouched (round-011 TD-1; the Given is hermetic).
	if err := copyTree(sc.home, homeCopy); err != nil {
		return err
	}
	if err := setPersonaAt(homeCopy, persona); err != nil {
		return err
	}
	env := sc.runEnv()
	env["TELL_ME_HOME"] = homeCopy
	res := harness.RunIn(sc.workDir, []string{prompt}, env, sc.unsetNames())
	n, ok := estimatedPayloadValue(res.Stderr)
	if !ok {
		return fmt.Errorf("the previous run reported no estimated payload status; stderr=%q", res.Stderr)
	}
	sc.previousEstimate = n
	sc.previousEstimateSet = true
	return nil
}

// writeServiceAccountKey generates a fresh RSA service-account key file at a
// home-relative path (its token_uri pointing at tokenURI) and returns its
// absolute path — the gemini provider's API_KEY target (round-013 T003).
func (sc *scenarioContext) writeServiceAccountKey(rel, tokenURI string) (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}
	sa := map[string]string{
		"type":         "service_account",
		"client_email": "e2e@example.iam.gserviceaccount.com",
		"private_key":  string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})),
		"token_uri":    tokenURI,
	}
	data, err := json.Marshal(sa)
	if err != nil {
		return "", err
	}
	if err := sc.writeFile(rel, data); err != nil {
		return "", err
	}
	return sc.homePath(rel), nil
}

// writeGeminiConfig writes the default configuration selecting a `gemini`
// provider whose Vertex-shaped URL points at the fake and whose API_KEY is the
// given service-account key file path (round-013 T003).
func (sc *scenarioContext) writeGeminiConfig(provider, model, fakeBase, keyPath string, maxTokens int) error {
	url := strings.TrimRight(fakeBase, "/") + "/v1/projects/e2e/locations/global/publishers/google/models"
	cfg := fmt.Sprintf("MODE: butler\n"+
		"PERSON: \"e2e persona\"\n"+
		"SELECTED_PROVIDER: %s\n"+
		"PROVIDERS:\n"+
		"  %s:\n"+
		"    TYPE: gemini\n"+
		"    MODEL: %s\n"+
		"    URL: \"%s\"\n"+
		"    API_KEY: \"%s\"\n"+
		"    MAX_TOKENS: %d\n"+
		"    THINKING_BUDGET: 32768\n"+
		"    THINKING_LEVEL: HIGH\n",
		provider, provider, model, url, keyPath, maxTokens)
	return sc.writeFile("configs/butler.yaml", []byte(cfg))
}

// setSelectedProviderMaxTokens sets the effective config's selected provider
// entry MAX_TOKENS to n (round-013 Given: the configured Gemini provider entry
// allows at most N output tokens).
func (sc *scenarioContext) setSelectedProviderMaxTokens(n int) error {
	p := sc.homePath("configs/butler.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("the default configuration must exist before setting MAX_TOKENS: %w", err)
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	sel, _ := cfg["SELECTED_PROVIDER"].(string)
	providers, _ := cfg["PROVIDERS"].(map[string]any)
	if providers == nil {
		return fmt.Errorf("the default configuration has no PROVIDERS registry")
	}
	entry, _ := providers[sel].(map[string]any)
	if entry == nil {
		return fmt.Errorf("the selected provider %q is not in the registry", sel)
	}
	entry["MAX_TOKENS"] = n
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, out, 0o644)
}
