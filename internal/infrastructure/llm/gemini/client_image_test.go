package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// TestCompleteRefusesMediaLoudly pins round-062 (ADR 0032 D4): the Gemini family
// has no inline_data image path this round, so a media-bearing message is a LOUD
// *llm.ProviderError — never a silent drop. The check fires before any network
// I/O, so no service-account credential is needed for this pin.
func TestCompleteRefusesMediaLoudly(t *testing.T) {
	c := &Client{cfg: Config{ProviderName: "vertex"}}
	_, err := c.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Media: []llm.MediaPart{{MIMEType: "image/png", Data: []byte{0x89}}}}},
	})
	var perr *llm.ProviderError
	if err == nil || !errors.As(err, &perr) {
		t.Fatalf("err = %v, want *llm.ProviderError", err)
	}
	if !strings.Contains(perr.Err.Error(), "cannot carry images") {
		t.Errorf("error = %q, want the loud media refusal", perr.Err.Error())
	}
}
