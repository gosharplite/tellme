package llm

import (
	"errors"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
	domainllm "github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/infrastructure/llm/openai"
)

func TestNewGateway_SupportedFamilies(t *testing.T) {
	for _, family := range []string{"openai", "deepseek", "kimi", "OpenAI", "  DeepSeek  "} {
		gw, err := NewGateway(config.Provider{Type: family, URL: "https://x", Model: "m"}, "p")
		if err != nil {
			t.Fatalf("family %q: unexpected error %v", family, err)
		}
		if _, ok := gw.(*openai.Client); !ok {
			t.Errorf("family %q: gateway = %T, want *openai.Client", family, gw)
		}
	}
}

func TestNewGateway_UnsupportedFamily(t *testing.T) {
	for _, family := range []string{"gemini", "anthropic", "", "google"} {
		gw, err := NewGateway(config.Provider{Type: family, URL: "https://x"}, "prov")
		if err == nil {
			t.Fatalf("family %q: expected an error, got gateway %T", family, gw)
		}
		var perr *domainllm.ProviderError
		if !errors.As(err, &perr) {
			t.Fatalf("family %q: error = %T, want *llm.ProviderError", family, err)
		}
		if perr.Provider != "prov" {
			t.Errorf("family %q: ProviderError.Provider = %q, want prov", family, perr.Provider)
		}
	}
}
