package llm

import "testing"

// TestEstimateTokensDeterministic pins the deterministic, offline estimator
// (round-009 research Decision 1): the same conversation always yields the same
// number; more content yields more.
func TestEstimateTokensDeterministic(t *testing.T) {
	msgs := []Message{{Role: "user", Content: "hello"}, {Role: "assistant", Content: "world!"}}
	first := EstimateTokens(msgs)
	for i := 0; i < 5; i++ {
		if got := EstimateTokens(msgs); got != first {
			t.Fatalf("EstimateTokens not deterministic: %d vs %d", got, first)
		}
	}
	if got := EstimateTokens(nil); got != 0 {
		t.Errorf("EstimateTokens(nil) = %d, want 0", got)
	}
	if first <= 0 {
		t.Errorf("EstimateTokens(msgs) = %d, want > 0", first)
	}
	more := EstimateTokens([]Message{{Role: "user", Content: "a substantially longer message body"}})
	less := EstimateTokens([]Message{{Role: "user", Content: "hi"}})
	if more <= less {
		t.Errorf("longer content must estimate more: %d vs %d", more, less)
	}
}
