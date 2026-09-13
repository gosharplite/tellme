package gemini

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// writeTestKey writes a valid service-account JSON key file (a fresh RSA key)
// whose token_uri points at tokenURI, and returns its path.
func writeTestKey(t *testing.T, tokenURI string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	sa := map[string]string{
		"type":         "service_account",
		"client_email": "e2e@example.iam.gserviceaccount.com",
		"private_key":  string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})),
		"token_uri":    tokenURI,
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

func TestNewTokenSource_Errors(t *testing.T) {
	if _, err := newTokenSource("", nil); err == nil {
		t.Error("empty path: expected an error")
	}
	if _, err := newTokenSource("/no/such/key.json", nil); err == nil {
		t.Error("missing file: expected an error")
	}
	if _, err := newTokenSource("/no/such/key.txt", nil); err == nil {
		t.Error("non-.json path: expected an error")
	}
	dir := t.TempDir()
	bad := filepath.Join(dir, "key.json")
	_ = os.WriteFile(bad, []byte("not json"), 0o600)
	if _, err := newTokenSource(bad, nil); err == nil {
		t.Error("invalid JSON: expected an error")
	}
	empty := filepath.Join(dir, "empty.json")
	_ = os.WriteFile(empty, []byte(`{"client_email":"","private_key":""}`), 0o600)
	if _, err := newTokenSource(empty, nil); err == nil {
		t.Error("missing client_email/private_key: expected an error")
	}
}

func TestTokenSource_MintCacheAndReuse(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
			t.Errorf("grant_type = %q", r.Form.Get("grant_type"))
		}
		if !strings.HasPrefix(r.Form.Get("assertion"), "ey") {
			t.Errorf("assertion is not a JWT: %q", r.Form.Get("assertion"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-123","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	ts, err := newTokenSource(writeTestKey(t, srv.URL), srv.Client())
	if err != nil {
		t.Fatalf("newTokenSource: %v", err)
	}
	for i := 0; i < 3; i++ {
		tok, err := ts.Token(context.Background())
		if err != nil {
			t.Fatalf("Token: %v", err)
		}
		if tok != "tok-123" {
			t.Fatalf("token = %q, want tok-123", tok)
		}
	}
	if h := atomic.LoadInt32(&hits); h != 1 {
		t.Errorf("token endpoint hit %d times, want 1 (cache reuse within the run)", h)
	}
}

func TestAssertion_IsRS256JWT(t *testing.T) {
	sa := serviceAccount{ClientEmail: "svc@example.iam.gserviceaccount.com"}
	keyFile := writeTestKey(t, "https://oauth2.googleapis.com/token")
	raw, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("read key: %v", err)
	}
	if err := json.Unmarshal(raw, &sa); err != nil {
		t.Fatalf("unmarshal key: %v", err)
	}
	assertion, err := sa.assertion(tsNow(), "https://oauth2.googleapis.com/token")
	if err != nil {
		t.Fatalf("assertion: %v", err)
	}
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		t.Fatalf("assertion has %d parts, want 3", len(parts))
	}
	headerJSON, err := b64urlDecode(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var header map[string]any
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	if header["alg"] != "RS256" {
		t.Errorf("alg = %v, want RS256", header["alg"])
	}
}
