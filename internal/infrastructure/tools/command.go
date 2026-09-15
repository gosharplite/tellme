package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// The bash-first command tool (round-024 D3/D1a/D2/D3/D8). It runs the model's
// `command` through `bash -c` (POSIX only; no Windows branch), with NO
// security/consent gate and NO pipe_commands (bash already pipes). A non-zero
// exit is a SUCCESSFUL result carrying the exit status; a command that exceeds
// its effective timeout is terminated as a PROCESS GROUP and surfaced as a
// nil-error timeout result (FR-003/FR-018); output is captured through BOUNDED
// PIPES (never a cmd.Stdout buffer, which os/exec reads to EOF); an optional
// output_file binds both streams directly to the file.

const (
	// commandDefaultTimeout is the command tool's per-tool default (round-024
	// FR-016: shell 300 s), declared upward via Contract().
	commandDefaultTimeout = 300 * time.Second
	// commandWaitDelay bounds the reap after a cancellation so an orphaned
	// descendant holding the pipe cannot wedge Wait (round-024 D1a).
	commandWaitDelay = 2 * time.Second
)

// executeCommand runs a shell command through bash -c.
type executeCommand struct{}

// Name is the wire-valid canonical identifier.
func (executeCommand) Name() string { return "execute_command" }

// Description is the model-facing summary.
func (executeCommand) Description() string { return "Run a shell command via bash -c." }

// Contract declares the command tool's per-tool default timeout (round-024 Q2).
func (executeCommand) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: commandDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments.
func (executeCommand) Parameters() json.RawMessage {
	secs := int(commandDefaultTimeout / time.Second)
	return json.RawMessage(fmt.Sprintf(`{"type":"object","properties":{"command":{"type":"string","description":"The shell command to run (via bash -c)."},"timeout":{"type":"number","description":"Optional seconds before the command's process tree is stopped and returns a timeout result; default %d."},"max_output_tokens":{"type":"integer","description":"Optional soft cap on the result size, in tokens (bytes = tokens x 4); default = the effective budget divided by 4."},"output_file":{"type":"string","description":"If set, write stdout+stderr to this file instead of returning them."},"append":{"type":"boolean","description":"Append to output_file instead of truncating it."},"reason":{"type":"string","description":"Reason for running the command."}},"required":["command","reason"]}`, secs))
}

// Execute runs the command. It returns the bounded result (or the timeout/
// truncation marker), and never an error for a non-zero exit (which is a
// successful result) nor for an observed deadline (a nil-error timeout result).
func (executeCommand) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	var args struct {
		Command    string `json:"command"`
		OutputFile string `json:"output_file"`
		Append     bool   `json:"append"`
		Reason     string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("execute_command: invalid arguments: %w", err)
	}
	if strings.TrimSpace(args.Command) == "" {
		return "", errors.New("execute_command: the command argument is required")
	}
	if args.OutputFile != "" {
		return runToFile(ctx, args.Command, args.OutputFile, args.Append)
	}
	b := int(budget)
	if b < 1 {
		b = 1
	}
	return runCaptured(ctx, args.Command, b)
}

// newCommandProcess builds a bash -c command in its own process group (round-024
// D1a). It uses exec.CommandContext so the effective timeout terminates the
// process tree on EVERY path — including output_file (review B1): when ctx is
// done the CommandContext watcher invokes Cancel, which kills the whole group,
// and WaitDelay bounds the reap so an orphaned descendant cannot wedge Wait.
func newCommandProcess(ctx context.Context, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = commandWaitDelay
	cmd.Cancel = func() error { return killGroup(cmd) }
	return cmd
}

// killGroup kills the whole process group of a started command (negative pgid),
// so bash AND its descendants die together (round-024 D1a). It is idempotent and
// safe to call from both the ctx watcher and the capture/trim abort path.
func killGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// exitStatus extracts the exit code of a finished command: 0 on success, the
// process's own code on a non-zero exit, and -1 when the failure is not an exit
// status (e.g. a signal kill).
func exitStatus(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// runToFile captures the child's stdout AND stderr directly to a file handle
// (round-024 D3): the output streams to disk and never enters memory; the result
// reports the exit status and the capture target, with NO inline preview.
func runToFile(ctx context.Context, command, path string, appendMode bool) (string, error) {
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if appendMode {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return "", fmt.Errorf("execute_command: failed to open output file: %w", err)
	}
	defer func() { _ = f.Close() }()
	cmd := newCommandProcess(ctx, command)
	cmd.Stdout = f
	cmd.Stderr = f
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("execute_command: failed to start: %w", err)
	}
	// Wait returns once ctx cancels (the watcher's Cancel kills the group) or the
	// command exits — so a hanging producer with output_file cannot wedge the turn
	// (review B1).
	werr := cmd.Wait()
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	return fmt.Sprintf("Output written to %s\nExit Code: %d\n", path, exitStatus(werr)), nil
}

// runCaptured reads the child's stdout and stderr through BOUNDED pipes,
// combined into the one result budget (round-024 D4/D8). On reaching the byte
// budget it stops reading and kills the process group (the "trimmed" outcome);
// on the deadline it kills the process group and returns the nil-error timeout
// result (the "stopped" outcome) — the pinned T1 order stop -> close read-ends ->
// kill(-pgid) if alive -> Wait.
func runCaptured(ctx context.Context, command string, budget int) (string, error) {
	cmd := newCommandProcess(ctx, command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("execute_command: stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("execute_command: stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("execute_command: failed to start: %w", err)
	}
	// Reserve marker space so the whole result (bytes + terminator) stays within
	// the byte budget and the loop's backstop stays inert (round-024 Q1/D4).
	limit := budget - len(capMarker)
	if limit < 1 {
		limit = 1
	}
	buf := &boundedBuffer{limit: limit, full: make(chan struct{})}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(buf, stdout) }()
	go func() { defer wg.Done(); _, _ = io.Copy(buf, stderr) }()
	finished := make(chan struct{})
	go func() { wg.Wait(); close(finished) }()

	trimmed := false
	select {
	case <-finished:
		// The producer finished on its own within budget (or closed its pipes).
	case <-buf.full:
		// Trim wins over a simultaneous deadline (the byte budget was reached):
		// stop -> close read-ends -> kill -> drain (review B2).
		trimmed = true
		abortCapture(cmd, stdout, stderr, finished)
	case <-ctx.Done():
		abortCapture(cmd, stdout, stderr, finished)
	}
	werr := cmd.Wait()

	if trimmed {
		// The byte budget was reached: a distinct "trimmed" outcome. The
		// signal-kill we caused (141/137) is NOT presented as the command's own
		// exit status (round-024 D7/T1).
		return string(buf.bytes()) + capMarker, nil
	}
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	out := string(buf.bytes())
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return fmt.Sprintf("%sExit Code: %d\n", out, exitStatus(werr)), nil
}

// abortCapture terminates the process tree and CLOSES the pipe read ends so a
// producer that escaped the process group (setsid / nohup-style) cannot wedge the
// drain — the pinned "close the read ends" step (round-024 D8/T1; review B2). The
// drain is additionally bounded by commandWaitDelay.
func abortCapture(cmd *exec.Cmd, stdout, stderr io.ReadCloser, finished <-chan struct{}) {
	_ = killGroup(cmd)
	_ = stdout.Close()
	_ = stderr.Close()
	select {
	case <-finished:
	case <-time.After(commandWaitDelay):
	}
}

// boundedBuffer accumulates up to limit bytes and signals (once) when the limit
// is reached, so the reader can stop and the process group can be killed. It is
// safe for the concurrent stdout/stderr readers.
type boundedBuffer struct {
	mu        sync.Mutex
	buf       []byte
	limit     int
	truncated bool
	full      chan struct{}
	closeOnce sync.Once
}

// Write appends p while it fits, signalling the full channel the first time the
// limit is reached; it never returns an error (the reader stops on the signal,
// not on a write error).
func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	room := b.limit - len(b.buf)
	if room <= 0 {
		b.markFull()
		return 0, nil
	}
	if len(p) > room {
		b.buf = append(b.buf, p[:room]...)
		b.markFull()
		return room, nil
	}
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// markFull signals the full channel once (caller holds the lock).
func (b *boundedBuffer) markFull() {
	b.truncated = true
	b.closeOnce.Do(func() { close(b.full) })
}

// bytes returns the accumulated bytes.
func (b *boundedBuffer) bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf
}

// NewCommandTool returns the bash-first command tool (round-024).
func NewCommandTool() domaintools.Tool { return executeCommand{} }

// Compile-time port conformance.
var _ domaintools.Tool = executeCommand{}
