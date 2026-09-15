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
// The child's stdin is an empty pipe — a non-terminal giving immediate EOF (the
// default: no piped input).
func Run(args []string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, "", args, nil, set, unset, 0, false)
}

// RunWithStdin is Run with a scripted standard input: the child's stdin is an
// os.Pipe carrying `stdin` (a non-terminal, so the CLI reads it), then EOF. The
// run is bounded by pipedRunTimeout so a hang fails explicitly (grill Q6).
func RunWithStdin(args []string, stdin string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, "", args, strings.NewReader(stdin), set, unset, pipedRunTimeout, false)
}

// RunBinary is Run against an explicit binary path.
func RunBinary(bin string, args []string, set map[string]string, unset []string) RunResult {
	return runExec(bin, "", args, nil, set, unset, 0, false)
}

// RunIn is Run with the child's working directory set to dir, so a scenario's
// working-directory fixtures (e.g. a file for the read_files tool) are visible
// to the child.
func RunIn(dir string, args []string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, dir, args, nil, set, unset, 0, false)
}

// RunInWithStdin is RunWithStdin with the child's working directory set to dir.
func RunInWithStdin(dir string, args []string, stdin string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, dir, args, strings.NewReader(stdin), set, unset, pipedRunTimeout, false)
}

// markerDeadline bounds the wait for the editor frame to paint, so a child that
// never paints still terminates (a generous falsification ceiling, not a
// synchronization sleep).
const markerDeadline = 10 * time.Second

// RunInWithSyncedStdin runs the child with a scripted stdin delivered in two
// chunks, but the second chunk (the terminal key) is written only AFTER the
// child's stderr shows `marker` (the editor frame painted) — an
// output-synchronized handshake instead of a wall-clock sleep (PR #51
// implementation-review TD1; the ADR-036 determinism discipline). bubbletea
// COALESCES frames when all keys are available at once, so without the handshake
// only the final frame would render and the editor box would never reach the
// capture (the round-016/015 presence assertions + the teardown witness).
func RunInWithSyncedStdin(dir string, args []string, compose, key, marker string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExecSynced(bin, dir, args, compose, key, marker, set, unset, pipedRunTimeout)
}

// runExecSynced is runExec with the output-synchronized stdin handshake: it
// drains stderr concurrently, writes `compose`, waits until the stream carries
// `marker` (bounded by markerDeadline), then writes `key` and closes stdin.
func runExecSynced(bin, dir string, args []string, compose, key, marker string, set map[string]string, unset []string, timeout time.Duration) RunResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = buildEnv(set, unset)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Start(); err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}

	// Drain stderr concurrently, signalling once the marker appears.
	painted := make(chan struct{})
	var once sync.Once
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, rerr := stderrPipe.Read(buf)
			if n > 0 {
				stderr.Write(buf[:n])
				if marker != "" && strings.Contains(stderr.String(), marker) {
					once.Do(func() { close(painted) })
				}
			}
			if rerr != nil {
				once.Do(func() { close(painted) })
				return
			}
		}
	}()

	go func() {
		_, _ = io.WriteString(stdin, compose)
		select {
		case <-painted:
		case <-time.After(markerDeadline):
		case <-ctx.Done():
		}
		_, _ = io.WriteString(stdin, key)
		_ = stdin.Close()
	}()

	err = cmd.Wait()
	<-done
	return finishResult(RunResult{Stdout: stdout.String(), Stderr: stderr.String()}, err, ctx, timeout)
}

// finishResult maps a cmd.Wait error to a RunResult (shared by runExec and
// runExecSynced): nil → success, a context deadline → an explicit hang message,
// an ExitError → its code, otherwise the raw error.
func finishResult(res RunResult, err error, ctx context.Context, timeout time.Duration) RunResult {
	if err == nil {
		return res
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

// RunInWithDevNull is RunIn with the child's stdin wired to the null device
// (os.DevNull) — a *character device* that is NOT a terminal. It pins the
// round-012 review BLOCKER B1 fix at the E2E layer: a real isatty probe must
// treat /dev/null as non-interactive, so tellme takes the boot path and never
// engages the interactive multi-line reader. (A bare os.ModeCharDevice probe
// reported /dev/null as a terminal, which is what this helper now guards.)
func RunInWithDevNull(dir string, args []string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	f, err := os.Open(os.DevNull)
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	defer func() { _ = f.Close() }()
	return runExec(bin, dir, args, f, set, unset, 0, false)
}

// RunInMerged is RunIn with the child's stdout and stderr MERGED into a single
// ordered buffer — the round-010 cross-stream ordering witness. Both streams are
// wired to the SAME comparable writer, so os/exec serializes the writes and the
// captured bytes preserve the child's write order (the `2>&1` view a real
// terminal shows). The merged bytes are returned in Stdout; Stderr is empty.
func RunInMerged(dir string, args []string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, dir, args, nil, set, unset, 0, true)
}

// RunInMergedWithStdin is RunInMerged with a scripted standard input.
func RunInMergedWithStdin(dir string, args []string, stdin string, set map[string]string, unset []string) RunResult {
	bin, err := BinaryPath()
	if err != nil {
		return RunResult{ExitCode: -1, Err: err}
	}
	return runExec(bin, dir, args, strings.NewReader(stdin), set, unset, pipedRunTimeout, true)
}

// runExec runs bin with args, wiring stdout/stderr (and stdin when non-nil) into
// buffers and capturing the exit code. When merge is true, stdout and stderr
// share one ordered buffer (the round-010 witness); otherwise they are captured
// separately. A non-zero timeout bounds the run via exec.CommandContext; on
// deadline expiry the process is killed and Err carries an explicit deadline
// message (distinct from a normal non-zero exit).
func runExec(bin, dir string, args []string, stdin io.Reader, set map[string]string, unset []string, timeout time.Duration, merge bool) RunResult {
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
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = buildEnv(set, unset)
	if stdin != nil {
		cmd.Stdin = stdin
	} else {
		// Default: an empty pipe — a NON-terminal — so the CLI's char-device
		// terminal probe does not mistake the null device for an interactive
		// terminal (round 012). An empty reader gives the child an immediate EOF
		// (the "no piped input" default), exactly as the null device did, but the
		// child's stdin is a pipe, not a character device.
		cmd.Stdin = strings.NewReader("")
	}

	var stdout, stderr bytes.Buffer
	if merge {
		// Merged witness (round 010): both streams share ONE comparable writer, so
		// os/exec serializes the writes and `stdout` preserves the child's
		// interleaved write order. Stderr stays empty.
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

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

// StripANSI removes ANSI SGR/CSI/OSC escape sequences so a step can assert the
// visible text of a rendered answer. The glamour renderer's exact codes are
// terminal/profile-dependent, so rendered assertions key on visible text
// (round-006 research residual risks).
func StripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			// CSI: ESC [ ... <final byte 0x40-0x7e>
			if i+1 < len(s) && s[i+1] == '[' {
				j := i + 2
				for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
					j++
				}
				i = j
				continue
			}
			// OSC: ESC ] ... BEL or ESC \
			if i+1 < len(s) && s[i+1] == ']' {
				j := i + 2
				for j < len(s) && s[j] != 0x07 {
					if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
						j++
						break
					}
					j++
				}
				i = j
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
