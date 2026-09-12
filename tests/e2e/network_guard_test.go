package e2e

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// TestOfflinePathsDoNotContactProvider is the offline-path witness (round-004
// research Decision 6; extended round 007 with `-l` and a prompt-less `--new`):
// the offline paths (`--version`, `-d`, a prompt-less boot, `-l`, and a
// prompt-less `--new`) must leave a recording sink untouched AND complete
// identically under blocked egress. It replaces the retired whole-binary
// capability guard, which no longer holds now that the prompt-bearing chat path
// links net/http.
func TestOfflinePathsDoNotContactProvider(t *testing.T) {
	var hits int64
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"should not be reached"}}]}`))
	}))
	defer sink.Close()

	home := t.TempDir()
	cfg := "MODE: butler\n" +
		"PERSON: \"e2e persona\"\n" +
		"SELECTED_PROVIDER: sink-prov\n" +
		"PROVIDERS:\n" +
		"  sink-prov:\n" +
		"    TYPE: deepseek\n" +
		"    MODEL: deepseek-v4-flash\n" +
		"    URL: " + sink.URL + "\n" +
		"    API_KEY: test-key\n"
	if err := os.MkdirAll(filepath.Join(home, "configs"), 0o755); err != nil {
		t.Fatalf("mkdir configs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, "configs", "butler.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	env := map[string]string{"TELL_ME_HOME": home}
	unset := []string{"TELL_ME_MODE", "TELL_ME_SELECTED_PROVIDER"}

	for _, args := range [][]string{{"--version"}, {"-d"}, {"-l", "5"}, {"--new"}, nil} {
		base := harness.Run(args, env, unset)
		if base.Err != nil || base.ExitCode != 0 {
			t.Fatalf("offline %v exited %d (err=%v): stderr=%q", args, base.ExitCode, base.Err, base.Stderr)
		}
		blocked := harness.RunWithBlockedNetwork(args, env, unset)
		if blocked.Err != nil {
			t.Fatalf("blocked run %v error: %v", args, blocked.Err)
		}
		if blocked.ExitCode != base.ExitCode || blocked.Stdout != base.Stdout {
			t.Errorf("offline %v changed under blocked egress: base(exit=%d out=%q) blocked(exit=%d out=%q)",
				args, base.ExitCode, base.Stdout, blocked.ExitCode, blocked.Stdout)
		}
	}

	if got := atomic.LoadInt64(&hits); got != 0 {
		t.Errorf("an offline path contacted the provider sink %d time(s)", got)
	}
}
