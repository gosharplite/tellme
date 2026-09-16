package mcp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// Round-032 principal-review BLOCKER [UNIT] — a chunked / delayed response body
// MUST be readable AFTER RoundTrip returns headers. The per-request deadline's
// cancel func is tied to Body.Close (not deferred inside RoundTrip), so the
// request context stays live for the whole body read.
//
// Channel-synchronised (no sleep): the handler flushes headers + the first
// chunk, then waits until the client has returned from RoundTrip before sending
// the tail. With a `defer cancel()` inside RoundTrip this test fails with
// `context canceled`.
func TestBoundedTransport_ChunkedBodyReadable(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"part":1`)
		if fl, ok := w.(http.Flusher); ok {
			fl.Flush() // headers + first chunk out; the client's RoundTrip returns
		}
		<-release // hold the tail until RoundTrip has returned
		_, _ = io.WriteString(w, `,"part":2}`)
	}))
	defer srv.Close()

	rt := &boundedTransport{base: http.DefaultTransport, timeout: 5 * time.Second}
	cl := &http.Client{Transport: rt}
	req, err := http.NewRequest(http.MethodPost, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	// Headers are in hand; let the server finish the body, then read it fully.
	close(release)
	body, rerr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if rerr != nil {
		t.Fatalf("reading the chunked body failed (the deadline was cancelled too early): %v", rerr)
	}
	if !strings.Contains(string(body), `"part":2`) {
		t.Fatalf("the body tail was not read: %q", body)
	}
}
