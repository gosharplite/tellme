package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
)

// TestConsumeOptionalIntArgs covers the round-081 generalisation of the round-054
// `-l` pre-pass to also cover `-b`/`--back` (ADR 0023 RF-54-4).
func TestConsumeOptionalIntArgs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"adjacent int is consumed (-l)", []string{"-l", "5"}, []string{"-l=5"}},
		{"adjacent int is consumed (-b)", []string{"-b", "3"}, []string{"-b=3"}},
		{"bare -b is left for NoOptDefVal", []string{"-b"}, []string{"-b"}},
		{"non-integer adjacent is left as a prompt", []string{"-b", "hello"}, []string{"-b", "hello"}},
		{"-- stops the scan", []string{"-b", "--", "5"}, []string{"-b", "--", "5"}},
		{"long forms too", []string{"--back", "2"}, []string{"--back=2"}},
		{"-l still works", []string{"-l", "hello"}, []string{"-l", "hello"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := consumeOptionalIntArgs(tc.args, "-l", "--list", "-b", "--back")
			if strings.Join(got, "|") != strings.Join(tc.want, "|") {
				t.Fatalf("consumeOptionalIntArgs(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

// TestParseFlagsBack covers the flag surface: a bare `-b` means 1; `-b 3` means 3;
// `-b "p"` means 1 plus the positional prompt `p`.
func TestParseFlagsBack(t *testing.T) {
	t.Parallel()
	if f, _, ok := parseFlags([]string{"-b"}, io.Discard); !ok || !f.backSet || f.back != 1 {
		t.Fatalf("parseFlags(-b): ok=%v backSet=%v back=%d, want true/true/1", ok, f.backSet, f.back)
	}
	if f, _, ok := parseFlags([]string{"-b", "3"}, io.Discard); !ok || !f.backSet || f.back != 3 {
		t.Fatalf("parseFlags(-b 3): ok=%v backSet=%v back=%d, want true/true/3", ok, f.backSet, f.back)
	}
	f, args, ok := parseFlags([]string{"-b", "rephrase"}, io.Discard)
	if !ok || !f.backSet || f.back != 1 || len(args) != 1 || args[0] != "rephrase" {
		t.Fatalf("parseFlags(-b rephrase): ok=%v backSet=%v back=%d args=%v, want true/true/1/[rephrase]", ok, f.backSet, f.back, args)
	}
}

// TestRenderRollback_ConfirmsAndClamps pins the offline rollback action: it rolls
// back the last N turns and prints the plain confirmation to stdout.
func TestRenderRollback_ConfirmsAndClamps(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	st := &fakeStore{entries: []history.Entry{
		{Prompt: "one", Answer: "a1"},
		{Prompt: "two", Answer: "a2"},
		{Prompt: "three", Answer: "a3"},
	}}
	var out, errOut bytes.Buffer
	code := renderRollback(home, "", runtimeEnv{stdout: &out, stderr: &errOut}, 1, func(string) history.Store { return st })
	if code != Success {
		t.Fatalf("code = %d, want Success; stderr=%q", code, errOut.String())
	}
	if got := out.String(); got != "Rolled back 1 turn. History now holds 2 turns.\n" {
		t.Fatalf("stdout = %q, want the rollback confirmation", got)
	}
	if len(st.rolled) != 1 || st.rolled[0] != 1 {
		t.Fatalf("Rollback calls = %v, want [1]", st.rolled)
	}
}

// TestRenderRollback_ClampsOverCount pins the clamp: -b 9 on a 1-turn session
// removes 1 and reports 1.
func TestRenderRollback_ClampsOverCount(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	st := &fakeStore{entries: []history.Entry{{Prompt: "only", Answer: "a"}}}
	var out, errOut bytes.Buffer
	if code := renderRollback(home, "", runtimeEnv{stdout: &out, stderr: &errOut}, 9, func(string) history.Store { return st }); code != Success {
		t.Fatalf("code = %d, want Success", code)
	}
	if got := out.String(); got != "Rolled back 1 turn. History now holds 0 turns.\n" {
		t.Fatalf("stdout = %q, want a clamped 1-turn report", got)
	}
}

// TestRollbackConfirmation_Pluralisation covers the turn/turns wording.
func TestRollbackConfirmation_Pluralisation(t *testing.T) {
	t.Parallel()
	if got := rollbackConfirmation(0, 0); got != "Rolled back 0 turns. History now holds 0 turns.\n" {
		t.Fatalf("(0,0) = %q", got)
	}
	if got := rollbackConfirmation(1, 1); got != "Rolled back 1 turn. History now holds 1 turn.\n" {
		t.Fatalf("(1,1) = %q", got)
	}
}
