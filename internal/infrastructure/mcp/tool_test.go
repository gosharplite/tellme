package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// recordingClient is a test double for the tools.MCPClient port.
type recordingClient struct {
	args map[string]interface{}
	err  error
	text string
}

func (c *recordingClient) ListTools(context.Context) ([]domaintools.MCPToolDefinition, error) {
	return nil, nil
}

func (c *recordingClient) CallTool(_ context.Context, _ string, args map[string]interface{}) (domaintools.ToolResult, error) {
	c.args = args
	if c.err != nil {
		return domaintools.ToolResult{}, c.err
	}
	return domaintools.ToolResult{Text: c.text}, nil
}

func (c *recordingClient) Close() error { return nil }

// T028(f) [UNIT] — nil args are normalised to a non-nil empty object (TD2/R2).
func TestTool_ExecuteNormalizesNilArgsToEmptyObject(t *testing.T) {
	c := &recordingClient{}
	tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "x"}, c, time.Second)
	if _, err := tool.Execute(context.Background(), "", 0); err != nil {
		t.Fatal(err)
	}
	if c.args == nil {
		t.Fatal("nil args must be normalised to a non-nil {} (a strict server rejects null)")
	}
	if len(c.args) != 0 {
		t.Fatalf("args must be empty, got %v", c.args)
	}
}

// T028(b) [UNIT] — a call-time failure is a recoverable nil-error result
// carrying the error text (TD1/R3): the loop never aborts on it.
func TestTool_ExecuteCallTimeErrorIsRecoverable(t *testing.T) {
	c := &recordingClient{err: errors.New("transport boom")}
	tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "x"}, c, time.Second)
	got, err := tool.Execute(context.Background(), `{}`, 0)
	if err != nil {
		t.Fatalf("a call-time failure must be a nil-error result; got err=%v", err)
	}
	if !strings.Contains(got, "transport boom") {
		t.Fatalf("the error text must be carried in the result: %q", got)
	}
}

// T028(d) [UNIT] — the adapter names its tool within the 64-byte budget.
func TestTool_NameIsBounded(t *testing.T) {
	server := strings.Repeat("s", 24)
	def := domaintools.MCPToolDefinition{Name: strings.Repeat("t", 200)}
	tool := NewTool(server, def, &recordingClient{}, time.Second)
	if len(tool.Name()) > 64 {
		t.Fatalf("name %q exceeds 64 bytes", tool.Name())
	}
}

// T028(e) [UNIT] — timeout default 300 s, server TIMEOUT honoured, clamped to
// the fixed 7200 s ceiling (TD5/FR-021).
func TestResolveMCPTimeout(t *testing.T) {
	if got := ResolveMCPTimeout(0); got != 300*time.Second {
		t.Fatalf("default = %v, want 300s", got)
	}
	if got := ResolveMCPTimeout(5); got != 5*time.Second {
		t.Fatalf("server TIMEOUT = %v, want 5s", got)
	}
	if got := ResolveMCPTimeout(99999); got != 7200*time.Second {
		t.Fatalf("clamp = %v, want 7200s", got)
	}
}

// T028 [UNIT] — the adapter clamps its own output at the source (FR-014).
func TestTool_ExecuteClampsAtSource(t *testing.T) {
	c := &recordingClient{text: strings.Repeat("a", 100)}
	tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "x"}, c, time.Second)
	got, err := tool.Execute(context.Background(), `{}`, domaintools.ByteBudget(10))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) > 10+len(domaintools.TruncationMarker) || !strings.HasSuffix(got, domaintools.TruncationMarker) {
		t.Fatalf("output not clamped to the source budget: %q", got)
	}
}
