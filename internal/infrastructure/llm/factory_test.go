package llm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
	domainllm "github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/infrastructure/llm/gemini"
	"github.com/gosharplite/tellme/internal/infrastructure/llm/openai"
)

func TestNewGateway_SupportedFamilies(t *testing.T) {
	for _, family := range []string{"openai", "deepseek", "kimi", "OpenAI", "  DeepSeek  "} {
		gw, err := NewGateway(config.Provider{Type: family, URL: "https://x", Model: "m"}, "p", "")
		if err != nil {
			t.Fatalf("family %q: unexpected error %v", family, err)
		}
		if _, ok := gw.(*openai.Client); !ok {
			t.Errorf("family %q: gateway = %T, want *openai.Client", family, gw)
		}
	}
}

func TestNewGateway_GeminiFamily(t *testing.T) {
	keyFile := writeTestServiceAccount(t)
	for _, family := range []string{"gemini", "google", "  Gemini  "} {
		gw, err := NewGateway(config.Provider{
			Type:   family,
			URL:    "http://127.0.0.1:1/v1/projects/e2e/locations/global/publishers/google/models",
			APIKey: keyFile,
			Model:  "gemini-3.8-flash",
		}, "vertex-flash-3.8", "")
		if err != nil {
			t.Fatalf("family %q: unexpected error %v", family, err)
		}
		if _, ok := gw.(*gemini.Client); !ok {
			t.Errorf("family %q: gateway = %T, want *gemini.Client", family, gw)
		}
	}
}

func TestNewGateway_GeminiFamily_MissingCredential(t *testing.T) {
	gw, err := NewGateway(config.Provider{Type: "gemini", URL: "https://x", APIKey: "/no/such/key.json", Model: "m"}, "prov", "")
	if err == nil {
		t.Fatalf("expected a credential error, got gateway %T", gw)
	}
	var perr *domainllm.ProviderError
	if !errors.As(err, &perr) {
		t.Fatalf("error = %T, want *llm.ProviderError", err)
	}
	if perr.Provider != "prov" {
		t.Errorf("ProviderError.Provider = %q, want prov", perr.Provider)
	}
}

func TestNewGateway_UnsupportedFamily(t *testing.T) {
	for _, family := range []string{"anthropic", "", "cohere"} {
		gw, err := NewGateway(config.Provider{Type: family, URL: "https://x"}, "prov", "")
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

// writeTestServiceAccount writes a valid service-account JSON key file (a fresh
// RSA key) into a temp dir and returns its path.
func writeTestServiceAccount(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	sa := map[string]string{
		"type":         "service_account",
		"client_email": "e2e@example.iam.gserviceaccount.com",
		"private_key":  string(pemBytes),
		"token_uri":    "https://oauth2.googleapis.com/token",
	}
	data, err := json.Marshal(sa)
	if err != nil {
		t.Fatalf("marshal service account: %v", err)
	}
	path := filepath.Join(t.TempDir(), "key.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write service account: %v", err)
	}
	return path
}
