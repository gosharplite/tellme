package steps

import "strings"

// Round-015 TUI key-sequence convention shared by the interactive-prompt step
// definitions and the product runner (round-015 T034). The scripted keys are the
// raw control bytes the TUI decodes from the injected stdin; the forced-terminal
// seam (TELL_ME_FORCE_STDIN_TTY) makes the piped stdin look like a terminal, so
// the prompt is drivable end-to-end without a real pty (research Decision 6).
//
//	Ctrl+C (\x03) — abort the prompt (no request)
//	Ctrl+S (\x13) — submit the composed prompt
//	Enter  (\r)   — a line break inside the multi-line editor (round 093)
//
// A key sequence is the operator's keystrokes concatenated.
//
// Kept in ONE file so the per-sentence step files stay independent (Zero Shared
// Edits), mirroring wire_tools.go / payload_status.go.
const (
	tuiKeyAbort  = "\x03" // Ctrl+C
	tuiKeySubmit = "\x13" // Ctrl+S
	tuiKeyAccept = "\t"   // Tab — accept the current suggestion (round 016)
	// tuiKeyEnter is the byte a REAL terminal delivers for the Enter key (CR) —
	// NOT a line feed. bubbletea decodes LF as `ctrl+j`, which the bubbles
	// textarea does not bind to `insert newline` (it binds CR/`enter` and
	// `ctrl+m`), so a raw \n was silently dropped and the typed lines collapsed
	// onto one row (round 093; issue #191).
	tuiKeyEnter = "\r"
)

// typeText encodes a scripted text keystroke sequence: each logical line break
// (`\n`, the readable Gherkin escape) becomes the terminal Enter byte so the
// product inserts a newline and keeps the lines apart (round 093).
func typeText(text string) string { return strings.ReplaceAll(text, "\n", tuiKeyEnter) }

// launchTUI arranges the next run as `tellme -i` against the forced-terminal seam
// with a scripted key sequence, then runs it (怎麼做 for the interactive-prompt
// When steps).
func launchTUI(sc *scenarioContext, keys string) { launchTUIArgs(sc, keys, []string{"-i"}) }

// launchTUIFresh arranges the next run as `tellme --new -i` (round 027): the same
// forced-terminal + scripted-key arrangement, but the run starts a fresh session
// first — so `--new` on the `-i` surface archives the active history before the
// submitted turn.
func launchTUIFresh(sc *scenarioContext, keys string) {
	launchTUIArgs(sc, keys, []string{"--new", "-i"})
}

func launchTUIArgs(sc *scenarioContext, keys string, args []string) {
	sc.setEnv("TELL_ME_FORCE_STDIN_TTY", "1")
	sc.setEnv("TELL_ME_TUI_DEBOUNCE", "0") // refresh synchronously (round 016)
	sc.pipeStdin(keys)                     // keep the plain stdin (the merged capture reruns it)
	// Round 023: deliver the compose keys, then the terminal key (submit/abort)
	// ONLY after the editor frame paints — an output-synchronized handshake
	// (harness.RunInWithSyncedStdin). bubbletea coalesces frames when all keys
	// arrive at once, so without the handshake the editor box would never reach
	// the capture (the round-016/015 presence assertions) and the teardown witness
	// (T006) would be vacuous.
	compose, final := keys, ""
	if len(keys) > 0 {
		compose, final = keys[:len(keys)-1], keys[len(keys)-1:]
	}
	sc.syncedStdin = [2]string{compose, final}
	sc.syncedSet = true
	sc.args = args
	sc.run()
}

// tuiKeysOpenAndAbort opens the prompt and aborts.
func tuiKeysOpenAndAbort() string { return tuiKeyAbort }

// tuiKeysTypeAndAbort opens the prompt, types q, then aborts.
func tuiKeysTypeAndAbort(q string) string { return typeText(q) + tuiKeyAbort }

// tuiKeysTypeAndSubmit opens the prompt, types text, then submits.
func tuiKeysTypeAndSubmit(text string) string { return typeText(text) + tuiKeySubmit }

// tuiKeysTypeAcceptAbort opens the prompt, types q, accepts the current
// suggestion (Tab), then aborts (round 016).
func tuiKeysTypeAcceptAbort(q string) string { return typeText(q) + tuiKeyAccept + tuiKeyAbort }

// renderedOutput returns the captured output a suggestion/dashboard presence
// assertion searches: the rendered interactive frame plus the answer stream. The
// TUI renders to the diagnostic stream (stderr); the union keeps the presence
// checks faithful to whichever stream carries the frame ("captured output").
func renderedOutput(sc *scenarioContext) string { return sc.stdout + "\n" + sc.stderr }
