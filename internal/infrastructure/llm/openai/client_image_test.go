package openai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// TestMessageContentPinsTheWireShape pins round-062 (ADR 0032): a message with
// NO media serializes as a plain string `content` (byte-identical to before); a
// message WITH media serializes as a content array (a leading text part when it
// states text, then one inline base64 image_url block per media part).
func TestMessageContentPinsTheWireShape(t *testing.T) {
	if got := messageContent(llm.Message{Role: "user", Content: "hi"}); got != "hi" {
		t.Errorf("media-less content = %#v, want the plain string \"hi\"", got)
	}

	img := domaintools.MediaPart{MIMEType: "image/png", Data: []byte{0x89, 'P', 'N', 'G'}}
	got, ok := messageContent(llm.Message{Role: "user", Media: []domaintools.MediaPart{img}}).([]any)
	if !ok {
		t.Fatalf("media-bearing content is not an array: %#v", got)
	}
	if len(got) != 1 {
		t.Fatalf("parts = %d, want 1 (no text on a media-only message)", len(got))
	}
	part := got[0].(map[string]any)
	if part["type"] != "image_url" {
		t.Errorf("part type = %v, want image_url", part["type"])
	}
	url := part["image_url"].(map[string]any)["url"].(string)
	want := "data:image/png;base64," + base64.StdEncoding.EncodeToString(img.Data)
	if url != want {
		t.Errorf("data URI = %q, want %q", url, want)
	}

	// A message with BOTH text and media leads with the text part (media-first
	// ordering for the image blocks that follow).
	both := messageContent(llm.Message{Role: "user", Content: "look", Media: []domaintools.MediaPart{img}}).([]any)
	if len(both) != 2 || both[0].(map[string]any)["type"] != "text" {
		t.Errorf("text+media parts = %#v, want a leading text part", both)
	}
}

// TestRequestBodyTextPathUnchanged pins I-1: with no media anywhere, the request
// body is byte-identical to the pre-round shape (a plain string content).
func TestRequestBodyTextPathUnchanged(t *testing.T) {
	prior := []llm.Message{{Role: "user", Content: "a"}, {Role: "assistant", Content: "b"}}
	body, err := requestBody("m", "p", prior, nil, 0, "", "")
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	for i, m := range decoded.Messages {
		if !strings.HasPrefix(strings.TrimSpace(string(m.Content)), `"`) {
			t.Errorf("messages[%d].content = %s, want a plain JSON string", i, m.Content)
		}
	}
}

// TestCompleteCarriesImageOnTheWire drives Complete end to end and asserts the
// recorded request body carries the image (the red-capable carrier: removing the
// serialization fails this pin).
func TestCompleteCarriesImageOnTheWire(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"seen"}}]}`))
	}))
	defer srv.Close()

	img := domaintools.MediaPart{MIMEType: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF, 0x01}}
	c := New(Config{ProviderName: "prov", BaseURL: srv.URL, Model: "m"})
	prior := []llm.Message{
		{Role: "user", Content: "what is this?"},
		{Role: "tool", Content: "Successfully read image from shot.jpg", ToolCallID: "call_1"},
		{Role: "user", Media: []domaintools.MediaPart{img}},
	}
	if _, err := c.Complete(context.Background(), llm.Request{Messages: prior}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	wantURI := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(img.Data)
	if !strings.Contains(gotBody, wantURI) {
		t.Errorf("recorded body does not carry the image data URI:\n%s", gotBody)
	}
	var decoded struct {
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(gotBody), &decoded); err != nil {
		t.Fatal(err)
	}
	last := decoded.Messages[len(decoded.Messages)-1]
	if !strings.HasPrefix(strings.TrimSpace(string(last.Content)), "[") {
		t.Errorf("media-bearing message content = %s, want a JSON array", last.Content)
	}
}
