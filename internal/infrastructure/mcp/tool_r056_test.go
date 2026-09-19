package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// T016 [UNIT] — round 056 (ADR 0025 D1): the offered declaration is tellme's OWN
// envelope — a required `reason` plus `MCP_PAYLOAD` whose subschema is the
// server's advertised input schema carried VERBATIM (the server's definition is
// never mutated).
func TestTool_ParametersOffersReasonEnvelopeWithVerbatimServerSchema(t *testing.T) {
	serverSchema := json.RawMessage(`{"type":"object","properties":{"sku":{"type":"string"}},"required":["sku"]}`)
	tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "lookup_price", InputSchema: serverSchema}, &recordingClient{}, time.Second)

	var params struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(tool.Parameters(), &params); err != nil {
		t.Fatalf("the offered parameters are not a JSON object: %v", err)
	}
	if params.Type != "object" {
		t.Fatalf("the envelope root must be an object, got %q", params.Type)
	}
	if len(params.Required) != 1 || params.Required[0] != ReasonKey {
		t.Fatalf("the envelope must require exactly %q, got %v", ReasonKey, params.Required)
	}
	if _, ok := params.Properties[ReasonKey]; !ok {
		t.Fatalf("the envelope must declare a %q property", ReasonKey)
	}
	payload, ok := params.Properties[MCPPayloadKey]
	if !ok {
		t.Fatalf("the envelope must declare a %q property", MCPPayloadKey)
	}
	// The server's advertised schema must appear VERBATIM inside MCP_PAYLOAD.
	var got, want any
	_ = json.Unmarshal(payload, &got)
	_ = json.Unmarshal(serverSchema, &want)
	gb, _ := json.Marshal(got)
	wb, _ := json.Marshal(want)
	if string(gb) != string(wb) {
		t.Fatalf("MCP_PAYLOAD subschema must be the server's schema verbatim:\n got %s\nwant %s", gb, wb)
	}
	// required ⊆ properties (the round-031/#64 invariant the envelope must keep).
	for _, r := range params.Required {
		if _, ok := params.Properties[r]; !ok {
			t.Fatalf("envelope required %q without declaring it", r)
		}
	}
}

// T015 [UNIT] — round 056 (ADR 0025 D2): Execute forwards ONLY the MCP_PAYLOAD
// object to the server; the envelope's outer keys never reach it.
func TestTool_ExecuteForwardsOnlyPayload(t *testing.T) {
	c := &recordingClient{text: "$42"}
	tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "lookup_price"}, c, time.Second)

	got, err := tool.Execute(context.Background(), `{"reason":"check the price","MCP_PAYLOAD":{"sku":"A1"}}`, 0)
	if err != nil {
		t.Fatalf("a valid envelope must not error: %v", err)
	}
	if got != "$42" {
		t.Fatalf("expected the server's result, got %q", got)
	}
	if c.args == nil {
		t.Fatal("the server must receive a non-nil payload object")
	}
	if _, ok := c.args["reason"]; ok {
		t.Fatalf("the `reason` must NOT be forwarded to the server: %v", c.args)
	}
	if _, ok := c.args[MCPPayloadKey]; ok {
		t.Fatalf("the `MCP_PAYLOAD` wrapper must NOT be forwarded: %v", c.args)
	}
	if c.args["sku"] != "A1" {
		t.Fatalf("the payload contents must be forwarded verbatim: %v", c.args)
	}
}

// T015 [UNIT] — an absent MCP_PAYLOAD (with a reason) is a legitimate EMPTY
// payload ({}).
func TestTool_ExecuteAbsentPayloadIsEmptyObject(t *testing.T) {
	c := &recordingClient{text: "$42"}
	tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "x"}, c, time.Second)
	if _, err := tool.Execute(context.Background(), `{"reason":"why"}`, 0); err != nil {
		t.Fatal(err)
	}
	if c.args == nil || len(c.args) != 0 {
		t.Fatalf("an absent MCP_PAYLOAD must be an empty object, got %v (nil=%v)", c.args, c.args == nil)
	}
}

// T015 [UNIT] — round 056 (ADR 0025 D2): a shape violation is REFUSED — the
// server is never contacted — and the result is a recoverable nil-error text.
func TestTool_ExecuteRefusesShapeViolationWithoutContactingServer(t *testing.T) {
	cases := map[string]string{
		"stray top-level key":   `{"reason":"why","sku":"A1"}`,
		"non-object payload":    `{"reason":"why","MCP_PAYLOAD":"not an object"}`,
		"null payload":          `{"reason":"why","MCP_PAYLOAD":null}`,
		"unparseable arguments": `not json`,
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			c := &recordingClient{text: "$42"}
			tool := NewTool("shop", domaintools.MCPToolDefinition{Name: "x"}, c, time.Second)
			got, err := tool.Execute(context.Background(), args, 0)
			if err != nil {
				t.Fatalf("a refusal must be a nil-error recoverable result; got %v", err)
			}
			if c.args != nil {
				t.Fatalf("the server must NOT be contacted on a shape violation; got args=%v", c.args)
			}
			if got == "" {
				t.Fatal("the refusal must carry a recoverable result text")
			}
		})
	}
}
