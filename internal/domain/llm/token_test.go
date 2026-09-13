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

// TestEstimatePayloadCountsWireInputs pins the wire-faithful estimate (round-011
// research Decisions 4 & 5): it counts the persona, the tool declarations, and
// the conversation messages; it is deterministic; it grows with the wired
// payload; and it exceeds the conversation-messages-only estimate.
func TestEstimatePayloadCountsWireInputs(t *testing.T) {
	msgs := []Message{{Role: "user", Content: "hello"}}
	tools := []ToolDef{{Name: "list_files", Description: "list a directory", Parameters: []byte(`{"type":"object"}`)}}

	base := EstimatePayload("", nil, msgs)
	if base != EstimateTokens(msgs) {
		t.Errorf("empty persona + no tools must equal the message estimate: %d vs %d", base, EstimateTokens(msgs))
	}
	withPersona := EstimatePayload("be terse", nil, msgs)
	if withPersona <= base {
		t.Errorf("a persona must raise the estimate: %d vs %d", withPersona, base)
	}
	withTools := EstimatePayload("", tools, msgs)
	if withTools <= base {
		t.Errorf("tool declarations must raise the estimate: %d vs %d", withTools, base)
	}
	full := EstimatePayload("be terse", tools, msgs)
	if full <= withPersona || full <= withTools {
		t.Errorf("persona+tools must exceed each alone: %d", full)
	}
	if full <= EstimateTokens(msgs) {
		t.Errorf("the wire estimate must exceed the conversation messages alone: %d vs %d", full, EstimateTokens(msgs))
	}
	for i := 0; i < 5; i++ {
		if got := EstimatePayload("be terse", tools, msgs); got != full {
			t.Fatalf("EstimatePayload not deterministic: %d vs %d", got, full)
		}
	}
	if longer := EstimatePayload("a much longer persona instruction body here", tools, msgs); longer <= full {
		t.Errorf("a longer persona must estimate more: %d vs %d", longer, full)
	}
	if got := EstimatePayload("", nil, nil); got != 0 {
		t.Errorf("EstimatePayload empty = %d, want 0", got)
	}
}
