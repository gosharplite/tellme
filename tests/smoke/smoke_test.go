// Package smoke verifies that the round-001 toolchain and dependency stack
// compile and execute under the Go test toolchain.
//
// It runs NO features and defines NO step definitions — the executable CLI
// contract is driven by the E2E suite created in later tasks. This file exists
// only to prove `go test` runs and that godog / pflag / yaml.v3 resolve.
package smoke

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func TestToolchainAndStackCompile(t *testing.T) {
	// godog: the BDD runner for the executable CLI contract (driven by tests/e2e).
	_ = godog.TestSuite{Name: "toolchain-smoke"}

	// pflag: GNU-style CLI flag parsing.
	fs := pflag.NewFlagSet("smoke", pflag.ContinueOnError)
	if fs == nil {
		t.Fatal("pflag.NewFlagSet returned nil")
	}

	// yaml.v3: configuration parsing — prove a round-trip works.
	var out struct {
		MODE string `yaml:"MODE"`
	}
	if err := yaml.Unmarshal([]byte("MODE: butler\n"), &out); err != nil {
		t.Fatalf("yaml.v3 unmarshal failed: %v", err)
	}
	if out.MODE != "butler" {
		t.Fatalf("yaml.v3 parsed MODE=%q, want %q", out.MODE, "butler")
	}
}
