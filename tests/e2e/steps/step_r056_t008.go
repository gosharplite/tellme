package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// T008 [BDD-RED] — Then: the request offered the tool "{tool}" from the MCP server "{server}" with a reason and the server's declared schema
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered the tool "([^"]*)" from the MCP server "([^"]*)" with a reason and the server's declared schema$`, thenOfferedMCPToolWithReasonAndSchema)
	})
}

// thenOfferedMCPToolWithReasonAndSchema (必查 呈現結果): the offered definition
// named `mcp_{server}_{tool}` declares a REQUIRED top-level `reason` and carries
// the server's advertised input schema VERBATIM as the `MCP_PAYLOAD` property's
// subschema — the server's definition is never mutated (round 056 / ADR 0025 D1).
func thenOfferedMCPToolWithReasonAndSchema(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no provider fake recorded a request")
	}
	var body struct {
		Tools []struct {
			Function struct {
				Name       string          `json:"name"`
				Parameters json.RawMessage `json:"parameters"`
			} `json:"function"`
		} `json:"tools"`
	}
	if err := json.Unmarshal([]byte(f.LastBody()), &body); err != nil {
		return fmt.Errorf("could not decode the recorded request tools: %w", err)
	}
	want := "mcp_" + server + "_" + tool
	for _, t := range body.Tools {
		if t.Function.Name != want {
			continue
		}
		var params struct {
			Type       string                     `json:"type"`
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		}
		if err := json.Unmarshal(t.Function.Parameters, &params); err != nil {
			return fmt.Errorf("the offered definition %q has unparseable parameters: %w", want, err)
		}
		if !containsString(params.Required, "reason") {
			return fmt.Errorf("the offered definition %q did not require a reason; required=%v", want, params.Required)
		}
		if _, ok := params.Properties["reason"]; !ok {
			return fmt.Errorf("the offered definition %q did not declare a reason property; properties=%v", want, params.Properties)
		}
		payload, ok := params.Properties[mcpPayloadKey]
		if !ok {
			return fmt.Errorf("the offered definition %q did not declare an %s property", want, mcpPayloadKey)
		}
		// The advertised (default freeform) schema must appear verbatim.
		var got, wantSchema any
		if err := json.Unmarshal(payload, &got); err != nil {
			return fmt.Errorf("the %s subschema is unparseable: %w", mcpPayloadKey, err)
		}
		_ = json.Unmarshal([]byte(`{"type":"object","properties":{}}`), &wantSchema)
		if !jsonEqual(got, wantSchema) {
			return fmt.Errorf("the offered %s subschema is not the server's declared schema; got=%s", mcpPayloadKey, string(payload))
		}
		return nil
	}
	return fmt.Errorf("the request did not offer %q; offered tools=%v", want, toolNamesOf(body))
}

func toolNamesOf(body struct {
	Tools []struct {
		Function struct {
			Name       string          `json:"name"`
			Parameters json.RawMessage `json:"parameters"`
		} `json:"function"`
	} `json:"tools"`
}) []string {
	var out []string
	for _, t := range body.Tools {
		out = append(out, t.Function.Name)
	}
	return out
}

// jsonEqual reports deep JSON equality of two decoded values.
func jsonEqual(a, b any) bool {
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(ab) == string(bb)
}
