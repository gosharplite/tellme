package agent

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gosharplite/tellme/internal/domain/tools"
)

// The pure tool-resource-contract resolver (round-024 D4/D5/FR-015/FR-016).
// Kept in its own file — out of Run — so the resolve/convert/clamp policy is
// unit-testable in isolation and Run stays under the cyclop gate (review R4,
// mirroring the internal/ui pure-formatter + observer-seam pattern).

// BytesPerToken is the contract-owned token→byte conversion factor (round-024
// Q1). It is deliberately a SEPARATE constant from the token estimator's
// heuristic ratio (internal/domain/llm) — this one is a contract term, that one
// is not — so the two are free to move independently.
const BytesPerToken = 4

// TimeoutCeiling is the hard upper bound on a tool call's effective timeout
// (round-024 FR-016): a `timeout` param above it is clamped, never rejected. It
// aliases the shared domain constant so the loop and the tool adapters cannot
// drift (round-032 implementation-review F4).
const TimeoutCeiling = tools.TimeoutCeiling

// resolveBound resolves a call's effective token bound in three tiers
// (round-024 D5): the call's `max_output_tokens` param when positive, else the
// default; then clamped to the ceiling. A non-positive ceiling is treated as
// unbounded. A floor of 1 keeps the bound positive.
func resolveBound(param, ceiling, def int) int {
	b := param
	if b <= 0 {
		b = def
	}
	if ceiling > 0 && b > ceiling {
		b = ceiling
	}
	if b < 1 {
		b = 1
	}
	return b
}

// clampBytes applies the loop's raw-byte-length backstop (round-024 Q1/D4): the
// result is compared by len — never by the estimator (whose +4 per-message
// overhead would double-trim). A compliant tool bounds at the source to the same
// byte budget (reserving marker space), so this backstop is inert for compliant
// tools; it exists so a misbehaving tool cannot overflow the window.
func clampBytes(result string, byteBudget int) string {
	if byteBudget <= 0 || len(result) <= byteBudget {
		return result
	}
	return strings.ToValidUTF8(result[:byteBudget], "") + tools.TruncationMarker
}

// resolveTimeout resolves a call's effective timeout in three tiers
// (round-024 FR-016): the call's `timeout` param when positive, else the tool's
// declared default (toolDefault, from its Contract), else DefaultToolTimeout;
// then clamped to TimeoutCeiling. It delegates to the shared domain resolver so
// the loop and the tool adapters share one implementation (F4).
func resolveTimeout(param, toolDefault time.Duration) time.Duration {
	return tools.ResolveTimeout(param, toolDefault)
}

// defaultEffectiveBudget is the fallback effective budget (tokens) when the loop
// is constructed without one — so a directly-constructed loop (unit tests) still
// resolves a sane bound. It matches config.DefaultMaxHistoryTokens.
const defaultEffectiveBudget = 1000000

// toolArgInt reads a top-level integer argument, or 0 when absent/unparseable.
func toolArgInt(arguments, key string) int {
	var probe map[string]json.RawMessage
	if json.Unmarshal([]byte(arguments), &probe) != nil {
		return 0
	}
	var n int
	if json.Unmarshal(probe[key], &n) != nil {
		return 0
	}
	return n
}

// toolArgSeconds reads a top-level numeric argument as a Duration, or 0 when
// absent/unparseable.
func toolArgSeconds(arguments, key string) time.Duration {
	var probe map[string]json.RawMessage
	if json.Unmarshal([]byte(arguments), &probe) != nil {
		return 0
	}
	var f float64
	if json.Unmarshal(probe[key], &f) != nil {
		return 0
	}
	if f <= 0 {
		return 0
	}
	return time.Duration(f * float64(time.Second))
}

// callTimeout resolves one call's effective timeout from its arguments and the
// tool's declared default (round-024 FR-016).
func (a *AgentLoop) callTimeout(tool tools.Tool, arguments string) time.Duration {
	return resolveTimeout(toolArgSeconds(arguments, "timeout"), tool.Contract().DefaultTimeout)
}

// callByteBudget resolves one call's effective BYTE budget (round-024 FR-015):
// the token bound (param -> clamped to the ceiling -> else the default), times
// the contract-owned bytes-per-token factor.
//
// The divisors here — the ceiling `eb/2` and the default `eb/4` — are MIRRORED by
// the `max_output_tokens` description strings in `resourceSchema`
// (internal/infrastructure/tools) and the command schema. If these divisors ever
// change, update those two description strings with them (review nit 3: makes the
// drift greppable without a shared constant).
func (a *AgentLoop) callByteBudget(arguments string) int {
	eb := a.EffectiveBudget
	if eb <= 0 {
		eb = defaultEffectiveBudget
	}
	tok := resolveBound(toolArgInt(arguments, "max_output_tokens"), eb/2, eb/4)
	return tok * BytesPerToken
}
