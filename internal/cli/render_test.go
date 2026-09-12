package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
)

// T013 — render/raw mode selection and the invalid-width configuration error
// (round-006 research Decisions 3 & 7).

func TestWriteAnswer_Raw(t *testing.T) {
	var out, errOut bytes.Buffer
	writeAnswer(&out, &errOut, "plain answer", true, 0, &stubRenderer{})
	if out.String() != "plain answer\n" {
		t.Fatalf("stdout = %q, want %q", out.String(), "plain answer\n")
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestWriteAnswer_RawAlwaysAppendsOneNewline(t *testing.T) {
	var out, errOut bytes.Buffer
	writeAnswer(&out, &errOut, "with newline\n", true, 0, &stubRenderer{})
	if out.String() != "with newline\n\n" {
		t.Fatalf("stdout = %q, want %q (answer verbatim + exactly one CLI-appended newline)", out.String(), "with newline\n\n")
	}
}

func TestWriteAnswer_Rendered(t *testing.T) {
	var out, errOut bytes.Buffer
	r := &stubRenderer{out: "\nRENDERED\n\n"}
	writeAnswer(&out, &errOut, "**Bold**", false, 40, r)
	if out.String() != "RENDERED\n\n" {
		t.Fatalf("stdout = %q, want %q (trimmed + one blank line)", out.String(), "RENDERED\n\n")
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestWriteAnswer_DegradedFallsBackToRaw(t *testing.T) {
	var out, errOut bytes.Buffer
	r := &stubRenderer{degraded: true}
	writeAnswer(&out, &errOut, "raw answer", false, 0, r)
	if out.String() != "raw answer\n" {
		t.Fatalf("stdout = %q, want the raw fallback %q", out.String(), "raw answer\n")
	}
	if !r.warned {
		t.Error("degradation did not emit the one-time warning")
	}
}

func TestResolveRejectsNegativeWrapWidth(t *testing.T) {
	t.Setenv("TELL_ME_SELECTED_PROVIDER", "")
	t.Setenv("TELL_ME_MODE", "butler")
	t.Setenv("TELL_ME_WRAP_WIDTH", "")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "MODE: butler\n"+
		"SELECTED_PROVIDER: prov\n"+
		"WRAP_WIDTH: -5\n"+
		"PROVIDERS:\n"+
		"  prov:\n"+
		"    TYPE: openai\n"+
		"    MODEL: gpt-5.5\n"+
		"    URL: https://example.test/v1\n")

	_, rerr := resolve(dir, cfgPath)
	if rerr == nil {
		t.Fatal("resolve() = nil error, want config-invalid for a negative WRAP_WIDTH")
	}
	if rerr.Reason != reasonConfigInvalid {
		t.Fatalf("resolve() reason = %q, want %q", rerr.Reason, reasonConfigInvalid)
	}
	if !errors.Is(rerr.Err, config.ErrInvalidValue) {
		t.Fatalf("resolve() error = %v, want a config.ErrInvalidValue wrap", rerr.Err)
	}
}

func TestResolveRejectsNonIntegerWrapWidthOverride(t *testing.T) {
	t.Setenv("TELL_ME_SELECTED_PROVIDER", "")
	t.Setenv("TELL_ME_MODE", "butler")
	t.Setenv("TELL_ME_WRAP_WIDTH", "not-a-number")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "MODE: butler\n"+
		"SELECTED_PROVIDER: prov\n"+
		"PROVIDERS:\n"+
		"  prov:\n"+
		"    TYPE: openai\n"+
		"    MODEL: gpt-5.5\n"+
		"    URL: https://example.test/v1\n")

	_, rerr := resolve(dir, cfgPath)
	if rerr == nil || rerr.Reason != reasonConfigInvalid {
		t.Fatalf("resolve() = %+v, want config-invalid for a non-integer TELL_ME_WRAP_WIDTH", rerr)
	}
}
