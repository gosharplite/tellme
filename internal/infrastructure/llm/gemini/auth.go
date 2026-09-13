// This file implements the Vertex AI service-account OAuth2 flow with the Go
// standard library only (round-013 research Decision 3, Clarify Q3 = stdlib-only):
// parse the operator's service-account JSON, sign a JWT-RS256 assertion, exchange
// it at the credential's token endpoint for an access token, and cache that token
// in memory with a 60-second expiry safety margin (research Decision 5 / review
// D3). No `golang.org/x/oauth2`, no provider SDK, no new module.
package gemini

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	tokenScope        = "https://www.googleapis.com/auth/cloud-platform"
	defaultTokenURI   = "https://oauth2.googleapis.com/token"
	expirySafetyGap   = 60 * time.Second
	maxTokenRespBytes = 1 << 20 // 1 MiB
	assertionLifetime = time.Hour
)

// serviceAccount is the subset of a GCP service-account JSON key the adapter needs.
type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// tokenSource mints and caches a Vertex access token from a service-account key
// file, concurrency-safely (round-013 research Decision 5 / review D3).
type tokenSource struct {
	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time
	cred        serviceAccount
	tokenURI    string
	http        *http.Client
	now         func() time.Time
}

// newTokenSource reads and parses the service-account key file at path. A
// missing / unreadable / non-`.json` / invalid credential is an error (surfaced
// by the adapter constructor as the frozen provider-failure class).
func newTokenSource(path string, httpClient *http.Client) (*tokenSource, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("a service-account key file is required (API_KEY must point at a .json credential)")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		return nil, fmt.Errorf("the gemini credential must be a service-account .json key file, got %q", path)
	}
	// config: the credential is read at turn time (research Decision 4).
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read the service-account key file %q: %w", path, err)
	}
	var sa serviceAccount
	if err := json.Unmarshal(raw, &sa); err != nil {
		return nil, fmt.Errorf("the service-account key file %q is not valid JSON: %w", path, err)
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return nil, fmt.Errorf("the service-account key file %q is missing client_email or private_key", path)
	}
	uri := sa.TokenURI
	if uri == "" {
		uri = defaultTokenURI
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &tokenSource{cred: sa, tokenURI: uri, http: httpClient, now: time.Now}, nil
}

// Token returns a valid access token, minting one on first use or near expiry.
// The token is obtained once and reused within the run (round-013 FR-008).
func (ts *tokenSource) Token(ctx context.Context) (string, error) {
	ts.mu.RLock()
	if ts.accessToken != "" && ts.now().Add(expirySafetyGap).Before(ts.expiresAt) {
		tok := ts.accessToken
		ts.mu.RUnlock()
		return tok, nil
	}
	ts.mu.RUnlock()

	ts.mu.Lock()
	defer ts.mu.Unlock()
	// Double-check under the write lock so a concurrent minter is not duplicated.
	if ts.accessToken != "" && ts.now().Add(expirySafetyGap).Before(ts.expiresAt) {
		return ts.accessToken, nil
	}
	return ts.mintLocked(ctx)
}

// mintLocked signs an assertion and exchanges it for a fresh access token. The
// caller must hold the write lock.
func (ts *tokenSource) mintLocked(ctx context.Context) (string, error) {
	assertion, err := ts.cred.assertion(ts.now(), ts.tokenURI)
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ts.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxTokenRespBytes))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token endpoint returned status %d", resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("unreadable token response: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("token response carried no access_token")
	}
	ts.accessToken = out.AccessToken
	ttl := out.ExpiresIn
	if ttl <= 0 {
		ttl = 3600
	}
	ts.expiresAt = ts.now().Add(time.Duration(ttl) * time.Second)
	return ts.accessToken, nil
}

// assertion builds and RS256-signs the JWT bearer assertion (pure-ish helper).
func (sa serviceAccount) assertion(now time.Time, aud string) (string, error) {
	key, err := parseRSAPrivateKey(sa.PrivateKey)
	if err != nil {
		return "", err
	}
	hb, err := json.Marshal(map[string]any{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	cb, err := json.Marshal(map[string]any{
		"iss":   sa.ClientEmail,
		"scope": tokenScope,
		"aud":   aud,
		"iat":   now.Unix(),
		"exp":   now.Add(assertionLifetime).Unix(),
	})
	if err != nil {
		return "", err
	}
	signingInput := b64url(hb) + "." + b64url(cb)
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return signingInput + "." + b64url(sig), nil
}

// parseRSAPrivateKey parses a PEM RSA private key (PKCS#8 or PKCS#1).
func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("the service-account private_key is not PEM-encoded")
	}
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		key, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("the service-account private_key is not an RSA key")
		}
		return key, nil
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("could not parse the service-account RSA private key")
}

// b64url is base64url without padding (the JWT wire encoding).
func b64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
