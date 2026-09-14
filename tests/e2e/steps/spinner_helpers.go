package steps

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Round-019 spinner helpers. Kept in ONE file so the per-sentence step files stay
// independent (Zero Shared Edits), mirroring payload_status.go / wire_tools.go.
//
// The spinner writes a carriage-return-redrawn line to the diagnostic stream:
// `\r` + erase-to-end-of-line + `{frame}{status} ({n}s){resource}`. The frame
// presence assertions read the RAW stream; the negative ("no residue")
// assertion simulates the redraw so a synchronously-drawn frame that was cleared
// leaves no visible residue (round-019 acceptance: "no progress indicator
// remains"; interface comment: "no indicator residue").

// reSpinnerFrame matches a braille frame followed by a phase status.
var reSpinnerFrame = regexp.MustCompile(`[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏] (?:Thinking|Executing)`)

// reSpinnerLine matches a full spinner line: a braille frame + a phase status +
// an `(<n>s)` elapsed segment.
var reSpinnerLine = regexp.MustCompile(`[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏] (?:Thinking|Executing)[^\r\n]*\([0-9]+s\)`)

// reSpinnerElapsed matches a spinner frame carrying an `(<n>s)` elapsed segment.
var reSpinnerElapsed = regexp.MustCompile(`[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏][^\r\n]*\([0-9]+s\)`)

// hasSpinnerFrame reports whether s carries a spinner frame followed by a phase
// status (the raw presence check).
func hasSpinnerFrame(s string) bool { return reSpinnerFrame.MatchString(s) }

// spinnerVisible simulates the carriage-return redraw: within each
// newline-delimited line only the text after the last `\r` survives, and the
// ANSI erase-to-end-of-line is dropped. A frame that was cleared synchronously
// therefore leaves no visible residue; a frame that was never cleared does.
func spinnerVisible(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, line := range strings.Split(s, "\n") {
		if i := strings.LastIndexByte(line, '\r'); i >= 0 {
			line = line[i+1:]
		}
		line = strings.ReplaceAll(line, "\x1b[K", "")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// spinnerSelectedModel reads the effective selected provider's configured MODEL
// from the default configuration the provider Given wrote (the label the spinner
// must name — reference parity: the MODEL attribute, not the registry key).
func spinnerSelectedModel(sc *scenarioContext) (string, error) {
	data, err := os.ReadFile(filepath.Join(sc.home, "configs", "butler.yaml"))
	if err != nil {
		return "", err
	}
	var cfg struct {
		SelectedProvider string `yaml:"SELECTED_PROVIDER"`
		Providers        map[string]struct {
			Model string `yaml:"MODEL"`
		} `yaml:"PROVIDERS"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.Providers[cfg.SelectedProvider].Model, nil
}
