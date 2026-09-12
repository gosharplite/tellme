// Package harness is a leaf test-support package for the tellme CLI E2E suite.
//
// It builds the tellme binary once per test process with the sentinel version
// injected, and runs it as a black-box subprocess, fully capturing the exit
// code, stdout, and stderr. It imports nothing from the steps/e2e packages, so
// the test graph stays acyclic: suite -> steps -> harness.
package harness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// SentinelVersion is the distinctive build version the E2E binary is linked
// with. It must never equal the local `dev` default, so a build where the
// -X main.version injection silently missed fails the --version assertion
// instead of passing green against `dev` (research.md Decision 6).
const SentinelVersion = "0.0.0-harness"

// pipedRunTimeout bounds every piped-stdin run so a hang (e.g. the CLI blocking
// on a terminal) fails with an explicit deadline message instead of being
// inferred from the suite's own hang-to-timeout (grill Q6). It is a generous
// falsification ceiling, not a synchronization sleep.
const pipedRunTimeout = 60 * time.Second

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// BinaryPath returns the path to the tellme binary, building it once per test
// process with `-ldflags "-X main.version=<SentinelVersion>"`. It never uses
// `make build` (whose VERSION defaults to `dev`).
func BinaryPath() (string, error) {
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "tellme-bin-")
		if err != nil {
			buildErr = fmt.Errorf("create binary temp dir: %w", err)
			return
		}
		binPath = filepath.Join(dir, "tellme")
		buildErr = build(binPath)
	})
	return binPath, buildErr
}

// build compiles ./cmd/tellme with the sentinel version into out.
func build(out string) error {
	cmd := exec.Command("go", "build",
		"-ldflags", "-X main.version="+SentinelVersion,
		"-o", out,
		"./cmd/tellme",
	)
	cmd.Dir = repoRoot()
	if combined, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build ./cmd/tellme: %w\n%s", err, combined)
	}
	return nil
}

// repoRoot resolves the module root from this source file's location
// (tests/e2e/harness/cmd_helper.go -> three parents up).
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// RunResult is the captured outcome of a black-box tellme invocation.
type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error // nil when the process started and exited (even non-zero)
}

// Run invokes the (once-built) tellme binary with args, under an environment
// derived from the current process plus `set` overrides and minus `unset` names.
// The child's stdin is the null device (the default: no piped input).
func Run(args []string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, args, nil, set, unset, 0)
}

// RunWithStdin is Run with a scripted standard input: the child's stdin is an
// os.Pipe carrying `stdin` (a non-terminal, so the CLI reads it), then EOF. The
// run is bounded by pipedRunTimeout so a hang fails explicitly (grill Q6).
func RunWithStdin(args []string, stdin string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, args, strings.NewReader(stdin), set, unset, pipedRunTimeout)
}

// RunBinary is Run against an explicit binary path.
func RunBinary(bin string, args []string, set map[string]string, unset []string) RunResult {
	return runExec(bin, args, nil, set, unset, 0)
}

// runExec runs bin with args, wiring stdout/stderr (and stdin when non-nil) into
// buffers and capturing the exit code. A non-zero timeout bounds the run via
// exec.CommandContext; on deadline expiry the process is killed and Err carries
// an explicit deadline message (distinct from a normal non-zero exit).
func runExec(bin string, args []string, stdin io.Reader, set map[string]string, unset []string, timeout time.Duration) RunResult {
	var (
		cmd *exec.Cmd
		ctx context.Context
	)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, bin, args...)
	} else {
		cmd = exec.Command(bin, args...)
	}
	cmd.Env = buildEnv(set, unset)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := RunResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err == nil {
		return res // ExitCode 0
	}
	if ctx != nil && ctx.Err() == context.DeadlineExceeded {
		res.ExitCode = -1
		res.Err = fmt.Errorf("tellme did not finish within %s (possible hang awaiting input)", timeout)
		return res
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.ExitCode = ee.ExitCode()
		return res
	}
	res.ExitCode = -1
	res.Err = err
	return res
}

// buildEnv derives a child environment from the current process: entries named
// in `unset` are removed; names in `set` override or add.
func buildEnv(set map[string]string, unset []string) []string {
	drop := make(map[string]bool, len(unset))
	for _, name := range unset {
		drop[name] = true
	}
	env := make([]string, 0, len(os.Environ())+len(set))
	seen := make(map[string]bool, len(set))
	for _, kv := range os.Environ() {
		eq := strings.IndexByte(kv, '=')
		if eq <= 0 {
			env = append(env, kv)
			continue
		}
		name := kv[:eq]
		if drop[name] {
			continue
		}
		if v, ok := set[name]; ok {
			env = append(env, name+"="+v)
			seen[name] = true
			continue
		}
		env = append(env, kv)
	}
	for name, v := range set {
		if !seen[name] {
			env = append(env, name+"="+v)
		}
	}
	return env
}
