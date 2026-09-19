// Package mcptest is the SDK-built hermetic fake MCP server used by the tellme
// E2E suite (round-032 T008, review fold B2). It lives INSIDE the confined
// internal/infrastructure/mcp tree so the verify-mcp-sdk-confinement gate (which
// covers production AND test files) is satisfied by construction, and it is
// built with the SDK's server API — no hand-rolled Streamable-HTTP wire, so the
// tests exercise tellme's real client against a real (in-process) MCP server.
package mcptest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// neverAnswerHold bounds how long a "never answers" fake holds a request before
// releasing it, so the httptest teardown cannot deadlock if a client abandons a
// request without closing its socket. It is well past the client's fixed 3 s
// fast-fail bound, so the non-stall behaviour is still exercised.
const neverAnswerHold = 5 * time.Second

// Options configures a fake MCP server's scripted behaviour.
type Options struct {
	// Tool is the tool name the server advertises ("" = it offers no tools).
	Tool string
	// Result is the text the tool returns when called.
	Result string
	// Description is the advertised tool description.
	Description string
	// Schema is the advertised input schema. Nil selects a freeform object schema
	// ({"type":"object","properties":{}}). A MALFORMED schema is a JSON object
	// with `required` entries not present in `properties` (the #64-mirror shape).
	Schema any
	// RequiredToken, when non-empty, makes the server demand
	// `Authorization: Bearer <RequiredToken>` on every request.
	RequiredToken string
	// NeverAnswer, when true, makes the server accept connections but never
	// respond (it holds each request until the client cancels) — the non-stall
	// witness.
	NeverAnswer bool
	// ToolError, when true, makes the tool return an MCP-level tool error
	// (isError: true).
	ToolError bool
	// TransportFail, when true, makes a tools/call request fail at the transport
	// layer (HTTP 500).
	TransportFail bool
}

// Server is a scriptable in-process MCP server.
type Server struct {
	opts Options
	mcp  *sdk.Server
	srv  *httptest.Server

	mu          sync.Mutex
	connections int
	auths       []string
	calls       []string
	callArgs    map[string][]string
	allArgs     []string
	closeOnce   sync.Once
}

// defaultSchema is the freeform object schema advertised when Options.Schema is
// nil.
func defaultSchema() any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

// Start launches the fake MCP server on a loopback listener.
func Start(opts Options) *Server {
	s := &Server{opts: opts}
	s.mcp = sdk.NewServer(&sdk.Implementation{Name: "fake-mcp", Version: "test"}, nil)
	if opts.Tool != "" {
		schema := opts.Schema
		if schema == nil {
			schema = defaultSchema()
		}
		desc := opts.Description
		if desc == "" {
			desc = "a fake MCP tool"
		}
		s.mcp.AddTool(&sdk.Tool{Name: opts.Tool, Description: desc, InputSchema: schema}, s.handleTool)
	}
	sdkHandler := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return s.mcp },
		&sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.recordConnection(r.Header.Get("Authorization"))
		if opts.NeverAnswer {
			// Hold the request past the client's fast-fail bound without answering,
			// so the "never answers" case is exercised — but release after a bounded
			// window so httptest.Server.Close() (teardown) cannot deadlock if the
			// client abandons without closing the socket.
			select {
			case <-r.Context().Done():
			case <-time.After(neverAnswerHold):
			}
			return
		}
		if opts.RequiredToken != "" && r.Header.Get("Authorization") != "Bearer "+opts.RequiredToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if opts.TransportFail && isToolCall(r) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sdkHandler.ServeHTTP(w, r)
	}))
	return s
}

// URL returns the base endpoint a configuration should point at.
func (s *Server) URL() string { return s.srv.URL }

// Close stops the server (idempotent).
func (s *Server) Close() {
	s.closeOnce.Do(func() { s.srv.Close() })
}

// ConnectionCount returns how many HTTP requests the server received.
func (s *Server) ConnectionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connections
}

// ReceivedAuthorization reports whether ANY request carried the given bearer
// token (`Authorization: Bearer <token>`).
func (s *Server) ReceivedAuthorization(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	want := "Bearer " + token
	for _, a := range s.auths {
		if a == want {
			return true
		}
	}
	return false
}

// Called reports whether the named tool was called.
func (s *Server) Called(tool string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, name := range s.calls {
		if name == tool {
			return true
		}
	}
	return false
}

func (s *Server) recordConnection(auth string) {
	s.mu.Lock()
	s.connections++
	s.auths = append(s.auths, auth)
	s.mu.Unlock()
}

func (s *Server) recordToolCall(name string) {
	s.mu.Lock()
	s.calls = append(s.calls, name)
	s.mu.Unlock()
}

// ReceivedArguments returns the JSON-marshalled arguments the fake received for
// each call to the named tool, in call order (round 056). It lets a Then assert
// exactly what tellme forwarded to the server — e.g. that the envelope's outer
// keys never reached it (the payload-purity witness).
func (s *Server) ReceivedArguments(tool string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.callArgs[tool]...)
}

// recordToolArguments appends one call's marshalled arguments under the tool.
func (s *Server) recordToolArguments(name string, args any) {
	b, err := json.Marshal(args)
	if err != nil {
		b = []byte("{}")
	}
	s.mu.Lock()
	if s.callArgs == nil {
		s.callArgs = map[string][]string{}
	}
	s.callArgs[name] = append(s.callArgs[name], string(b))
	s.allArgs = append(s.allArgs, string(b))
	s.mu.Unlock()
}

// AllReceivedArguments returns every recorded tool-call's marshalled arguments,
// in call order (round 056 — the payload-purity witness does not need to key by
// tool name).
func (s *Server) AllReceivedArguments() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.allArgs...)
}

// handleTool is the low-level SDK tool handler: it records the call and returns
// either the scripted result or an MCP-level tool error.
func (s *Server) handleTool(_ context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
	if req != nil && req.Params != nil {
		s.recordToolCall(req.Params.Name)
		s.recordToolArguments(req.Params.Name, req.Params.Arguments)
	}
	if s.opts.ToolError {
		return &sdk.CallToolResult{
			IsError: true,
			Content: []sdk.Content{&sdk.TextContent{Text: "the tool failed"}},
		}, nil
	}
	return &sdk.CallToolResult{
		Content: []sdk.Content{&sdk.TextContent{Text: s.opts.Result}},
	}, nil
}

// isToolCall reports whether the JSON-RPC request body is a tools/call method
// (the transport-failure mode's trigger). It restores the request body.
func isToolCall(r *http.Request) bool {
	if r.Body == nil {
		return false
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return bytes.Contains(body, []byte(`"tools/call"`))
}

// MalformedSchema returns the #64-mirror malformed input schema — a JSON object
// whose `required` names an undeclared property — for a scenario that needs a
// tool the normalizer must skip.
func MalformedSchema() any {
	return map[string]any{"type": "object", "required": []any{"missing_property"}}
}

// schemaJSON is a small helper used by tests to build a raw schema object from a
// JSON string; it panics on an invalid literal (test-only).
func schemaJSON(s string) any {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		panic(err)
	}
	return v
}

// SchemaWithProperty builds a well-formed object schema declaring one required
// string property named name — a helper for tests asserting the offered schema.
func SchemaWithProperty(name string) any {
	return schemaJSON(`{"type":"object","properties":{"` + strings.TrimSpace(name) + `":{"type":"string"}},"required":["` + strings.TrimSpace(name) + `"]}`)
}
