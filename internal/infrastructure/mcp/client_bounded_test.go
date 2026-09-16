package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// T030/T032 [UNIT] — the adapter must not stall on a server that never answers:
// the per-request transport deadline (the fixed fast-fail bound) abandons the
// connect within bound + margin. The fake releases after a bounded hold so the
// httptest teardown cannot deadlock. Falsifiability (T032 witness a): removing
// the transport deadline makes Connect block until the fake releases (≈3 s),
// failing the < 1 s assertion.
func TestNewRemoteClient_NonStallOnNeverAnsweringServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
	}))
	defer srv.Close()

	start := time.Now()
	_, err := NewRemoteClient(context.Background(), srv.URL, "", 300*time.Millisecond, time.Second)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected a connect error against a never-answering server")
	}
	if elapsed > time.Second {
		t.Fatalf("Connect stalled for %v (bound 300ms); the fast-fail bound is not applied", elapsed)
	}
}
