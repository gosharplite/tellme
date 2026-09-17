// Package arch holds tellme's layer-discipline gate: an import-direction check
// over the pinned layer ranking.
//
// The gate is a build-tagged (//go:build arch) test, so it never contributes to
// the default `go test ./...` run. It is executed by the `verify-architecture`
// Makefile target, which is a member of `make verify`:
//
//	go vet -tags=arch ./tools/arch
//	go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch
//
// `-count=1` is required: the gate's verdict depends on the whole module (via
// `go list`), which the Go test **cache cannot see** — a cached result would let
// an added illegal import pass vacuously.
//
// The committed baseline (tools/arch/baseline.txt) is a fail-on-stale ratchet:
// it is generated from the gate's own output — regenerate it with
// `make verify-architecture-update` — never hand-edit it.
//
// Policy: docs/decisions/0011-layer-discipline-gate.md.
// Truth: specs/truth/techstack.md (Build & Tooling, "Layer-discipline gate").
package arch
