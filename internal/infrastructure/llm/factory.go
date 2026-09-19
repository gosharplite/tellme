// Package llm is the provider-transport composition seam: it maps a resolved
// provider entry to the concrete llm.Gateway adapter for its wire family, so the
// presentation layer never couples to a concrete adapter constructor and an
// un-adapted family fails with an actionable client error (review finding #1).
package llm

import (
	"fmt"
	"strings"

	"github.com/gosharplite/tellme/internal/config"
	domainllm "github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/infrastructure/llm/gemini"
	"github.com/gosharplite/tellme/internal/infrastructure/llm/openai"
)

// supportedFamilies are the OpenAI-compatible provider TYPE labels.
var supportedFamilies = map[string]bool{
	"openai":   true,
	"deepseek": true,
	"kimi":     true,
}

// geminiFamilies are the Vertex/Gemini provider TYPE labels (round 013).
var geminiFamilies = map[string]bool{
	"gemini": true,
	"google": true,
}

// Family returns the transport family for a provider TYPE label — "openai" for a
// supported OpenAI-compatible label, "gemini" for a Vertex/Gemini label, and ""
// for an unsupported one. It reads the SAME tables NewGateway dispatches on, so
// the label→family classification has one owner (round 063: the composition root
// uses it to resolve the family-aware image ceiling; internal/cli never
// classifies a label itself).
func Family(typeLabel string) string {
	f := strings.ToLower(strings.TrimSpace(typeLabel))
	switch {
	case supportedFamilies[f]:
		return "openai"
	case geminiFamilies[f]:
		return "gemini"
	}
	return ""
}

// NewGateway returns the concrete domainllm.Gateway for the resolved provider's
// family. `openai`/`deepseek`/`kimi` use the OpenAI-compatible adapter;
// `gemini`/`google` use the Vertex/Gemini adapter (round 013); any other family
// keeps the existing unsupported-provider failure — the mapping never widens
// silently (round-013 FR-013).
func NewGateway(prov config.Provider, name, persona string) (domainllm.Gateway, error) {
	family := strings.ToLower(strings.TrimSpace(prov.Type))
	switch {
	case supportedFamilies[family]:
		return openai.New(openai.Config{
			ProviderName:  name,
			BaseURL:       prov.URL,
			APIKey:        prov.APIKey,
			Model:         prov.Model,
			MaxTokens:     prov.MaxTokens,
			Headers:       prov.Headers,
			ThinkingLevel: prov.ThinkingLevel,
			Persona:       persona,
		}), nil
	case geminiFamilies[family]:
		gw, err := gemini.New(gemini.Config{
			ProviderName:   name,
			BaseURL:        prov.URL,
			APIKey:         prov.APIKey,
			Model:          prov.Model,
			MaxTokens:      prov.MaxTokens,
			Headers:        prov.Headers,
			ThinkingBudget: prov.ThinkingBudget,
			ThinkingLevel:  prov.ThinkingLevel,
			Persona:        persona,
		})
		if err != nil {
			return nil, err // already a *domainllm.ProviderError (credential failure)
		}
		return gw, nil
	default:
		return nil, &domainllm.ProviderError{
			Provider: name,
			Err:      fmt.Errorf("unsupported provider family %q (supported: openai, deepseek, kimi, gemini, google)", prov.Type),
		}
	}
}
