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
	"github.com/gosharplite/tellme/internal/infrastructure/llm/openai"
)

// supportedFamilies are the OpenAI-compatible provider TYPE labels this slice
// adapts (round-004 Clarify Q2). Gemini/Vertex and Anthropic are deferred.
var supportedFamilies = map[string]bool{
	"openai":   true,
	"deepseek": true,
	"kimi":     true,
}

// NewGateway returns the concrete domainllm.Gateway for the resolved provider's
// family. An unsupported family yields a *llm.ProviderError (rendered as the
// frozen provider class phrase + exit code 6) instead of a malformed request.
func NewGateway(prov config.Provider, name string) (domainllm.Gateway, error) {
	if !supportedFamilies[strings.ToLower(strings.TrimSpace(prov.Type))] {
		return nil, &domainllm.ProviderError{
			Provider: name,
			Err:      fmt.Errorf("unsupported provider family %q (supported: openai, deepseek, kimi)", prov.Type),
		}
	}
	return openai.New(openai.Config{
		ProviderName:  name,
		BaseURL:       prov.URL,
		APIKey:        prov.APIKey,
		Model:         prov.Model,
		MaxTokens:     prov.MaxTokens,
		Headers:       prov.Headers,
		ThinkingLevel: prov.ThinkingLevel,
	}), nil
}
