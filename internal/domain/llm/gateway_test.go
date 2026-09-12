package llm

import "testing"

// Round-007 T018 / RF-3: the Request carries the resumed prior messages ahead of
// the current prompt; a fresh conversation has an empty Messages slice.
func TestRequestCarriesPriorMessages(t *testing.T) {
	req := Request{
		Prompt: "what is my name?",
		Messages: []Message{
			{Role: "user", Content: "my name is alice"},
			{Role: "assistant", Content: "noted"},
		},
	}
	if req.Prompt != "what is my name?" {
		t.Errorf("Prompt = %q", req.Prompt)
	}
	if len(req.Messages) != 2 {
		t.Fatalf("Messages = %+v, want 2", req.Messages)
	}
	if req.Messages[0] != (Message{Role: "user", Content: "my name is alice"}) {
		t.Errorf("Messages[0] = %+v", req.Messages[0])
	}
	if req.Messages[1] != (Message{Role: "assistant", Content: "noted"}) {
		t.Errorf("Messages[1] = %+v", req.Messages[1])
	}

	var fresh Request
	if len(fresh.Messages) != 0 {
		t.Errorf("fresh Request.Messages = %+v, want empty", fresh.Messages)
	}
}
