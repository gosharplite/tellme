package cli

import (
	"errors"
	"strings"
	"testing"
)

// T015 — fast, isolated unit tests for the flag parsing and the input/output
// mode selection (round-005 research Decisions 3 & 4; folds issue #14). These
// complement the black-box E2E acceptance path.

func TestCombinePrompt(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		piped string
		want  string
	}{
		{name: "args only", args: []string{"hi"}, want: "hi"},
		{name: "multiple args joined with single spaces", args: []string{"a", "b", "c"}, want: "a b c"},
		{name: "piped only", piped: "body", want: "body"},
		{name: "instruction then piped content", args: []string{"Summarize"}, piped: "package main", want: "Summarize\npackage main"},
		{name: "surrounding whitespace is trimmed", args: []string{"x"}, piped: "  y\n", want: "x\n  y"},
		{name: "empty pipe and no args", want: ""},
		{name: "empty pipe with an arg", args: []string{"only"}, want: "only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var piped []byte
			if tt.piped != "" {
				piped = []byte(tt.piped)
			}
			if got := combinePrompt(tt.args, piped); got != tt.want {
				t.Fatalf("combinePrompt(%v, %q) = %q, want %q", tt.args, tt.piped, got, tt.want)
			}
		})
	}
}

func TestResolvePromptModeSelection(t *testing.T) {
	alwaysTTY := func(any) bool { return true }
	neverTTY := func(any) bool { return false }

	tests := []struct {
		name  string
		args  []string
		stdin string
		isTTY func(any) bool
		want  string
	}{
		{name: "terminal stdin is not read", stdin: "ignored", isTTY: alwaysTTY, want: ""},
		{name: "terminal stdin keeps the positional arg", args: []string{"ask"}, stdin: "ignored", isTTY: alwaysTTY, want: "ask"},
		{name: "piped content with no arg", stdin: "body", isTTY: neverTTY, want: "body"},
		{name: "instruction combined with piped content", args: []string{"ask"}, stdin: "body", isTTY: neverTTY, want: "ask\nbody"},
		{name: "empty pipe with no arg yields no prompt", isTTY: neverTTY, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePrompt(tt.args, strings.NewReader(tt.stdin), tt.isTTY)
			if err != nil {
				t.Fatalf("resolvePrompt error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolvePrompt = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvePromptPropagatesReadError(t *testing.T) {
	_, err := resolvePrompt(nil, failingReader{}, func(any) bool { return false })
	if err == nil {
		t.Fatal("resolvePrompt = nil error, want a read error")
	}
}

func TestReadPipedStdinIsBounded(t *testing.T) {
	big := strings.Repeat("a", maxStdinBytes+100)
	got, err := readPipedStdin(strings.NewReader(big))
	if err != nil {
		t.Fatalf("readPipedStdin error: %v", err)
	}
	if len(got) != maxStdinBytes {
		t.Fatalf("readPipedStdin read %d bytes, want %d (bounded)", len(got), maxStdinBytes)
	}
}

func TestDefaultIsTerminalNonFileIsFalse(t *testing.T) {
	if defaultIsTerminal(strings.NewReader("x")) {
		t.Error("defaultIsTerminal(non-file) = true, want false")
	}
}

func TestParseFlags(t *testing.T) {
	opts, args, ok := parseFlags([]string{"-c", "/tmp/x.yaml", "hello", "world"})
	if !ok {
		t.Fatal("parseFlags ok = false, want true")
	}
	if opts.configPath != "/tmp/x.yaml" {
		t.Errorf("configPath = %q, want /tmp/x.yaml", opts.configPath)
	}
	if opts.diagnostic || opts.version {
		t.Errorf("diagnostic/version = %v/%v, want false/false", opts.diagnostic, opts.version)
	}
	if strings.Join(args, " ") != "hello world" {
		t.Errorf("positional args = %v, want [hello world]", args)
	}

	opts, args, ok = parseFlags([]string{"-d"})
	if !ok || !opts.diagnostic || len(args) != 0 {
		t.Errorf("parseFlags(-d) = (%+v, %v, %v), want diagnostic with no args", opts, args, ok)
	}

	if _, _, ok := parseFlags([]string{"--wibble"}); ok {
		t.Error("parseFlags(--wibble) ok = true, want false (unrecognized flag)")
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("boom") }
