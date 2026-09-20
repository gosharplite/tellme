package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// Round 074 (ADR 0046) — the `-h`/`--help` and `-v`/`--version` flag surface.

// TestParseFlagsHelpShorthand pins that `-h` and `--help` set the help flag and
// capture the rendered flag list (which must list both new flags).
func TestParseFlagsHelpShorthand(t *testing.T) {
	for _, form := range []string{"-h", "--help"} {
		opts, args, ok := parseFlags([]string{form}, io.Discard)
		if !ok {
			t.Fatalf("parseFlags(%q) ok = false, want true", form)
		}
		if !opts.help {
			t.Errorf("parseFlags(%q): help = false, want true", form)
		}
		if len(args) != 0 {
			t.Errorf("parseFlags(%q): args = %v, want []", form, args)
		}
		for _, want := range []string{"-h, --help", "-v, --version"} {
			if !strings.Contains(opts.helpText, want) {
				t.Errorf("help text for %q is missing %q; got %q", form, want, opts.helpText)
			}
		}
	}
}

// TestParseFlagsVersionShorthand pins that `-v` sets the same version flag as
// `--version` (one code path — no separate shape).
func TestParseFlagsVersionShorthand(t *testing.T) {
	for _, form := range []string{"-v", "--version"} {
		opts, _, ok := parseFlags([]string{form}, io.Discard)
		if !ok {
			t.Fatalf("parseFlags(%q) ok = false, want true", form)
		}
		if !opts.version {
			t.Errorf("parseFlags(%q): version = false, want true", form)
		}
	}
}

// TestParseFlagsUnknownFlagStillRefused pins I-1: the new flags do not weaken the
// usage-error path — an unrecognized flag still returns ok=false (the caller
// then prints the pinned phrase and exits 2).
func TestParseFlagsUnknownFlagStillRefused(t *testing.T) {
	for _, form := range []string{"-z", "--wibble", "-H", "--Help"} {
		if _, _, ok := parseFlags([]string{form}, io.Discard); ok {
			t.Errorf("parseFlags(%q) ok = true, want false (an unrecognized flag must be refused)", form)
		}
	}
}

// TestRunHelpPrecedenceAndStream pins the round-074 contract: `-h` beats `-v`
// (help is checked first), help writes to stdout and exits 0, and stderr stays
// empty.
func TestRunHelpPrecedenceAndStream(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-h", "-v"}, "0.0.0-test", testOptions(), env(&out, &errOut, &stubRenderer{}))
	if code != Success {
		t.Fatalf("run(-h -v) code = %d, want Success (help precedes version)", code)
	}
	if !strings.Contains(out.String(), "-h, --help") {
		t.Errorf("stdout must carry the flag list; got %q", out.String())
	}
	if strings.Contains(out.String(), "0.0.0-test") {
		t.Errorf("help must not print the version; got %q", out.String())
	}
	if errOut.String() != "" {
		t.Errorf("help must carry nothing on stderr; got %q", errOut.String())
	}
}

// TestRunVersionShorthand pins that `-v` prints the version to stdout and exits 0.
func TestRunVersionShorthand(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-v"}, "9.9.9-sentinel", testOptions(), env(&out, &errOut, &stubRenderer{}))
	if code != Success {
		t.Fatalf("run(-v) code = %d, want Success", code)
	}
	if !strings.Contains(out.String(), "9.9.9-sentinel") {
		t.Errorf("stdout must carry the version; got %q", out.String())
	}
	if errOut.String() != "" {
		t.Errorf("version must carry nothing on stderr; got %q", errOut.String())
	}
}
