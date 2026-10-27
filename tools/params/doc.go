// Package params holds tellme's compile-blind argument-swap gate: a signature
// scan that reds when a production function or method under internal/** can have
// two adjacent parameters swapped at a call site and still compile.
//
// The gate is a build-tagged (//go:build quality) test, so it never contributes
// to the default `go test ./...` run. It is executed by the `verify-params`
// Makefile target, which is a member of `make verify`:
//
//	go vet -tags=quality ./tools/params
//	golangci-lint run --build-tags=quality ./tools/params/...
//	go test -count=1 -tags=quality -run TestVerifyParameterShapes ./tools/params
//
// `-count=1` is kept — and its rationale is DIFFERENT from the architecture gate's.
// `tools/arch` shells a `go list` CHILD PROCESS, which the Go test cache genuinely
// cannot see; THIS gate reads the tree in-process (`filepath.WalkDir` +
// `parser.ParseFile`), and Go keys the cache on exactly the files a test opens
// (`go help test`: "Tests that open files within the package's module … only match
// future runs in which the files … are unchanged"). So here the cache does see the
// tree, and `-count=1` is belt-and-braces + parity with `verify-architecture` —
// not the difference between a witness and a vacuous pass. Recording the
// distinction is the point: an inherited rationale is not a verified one.
//
// The four arms are (a) n>=4 with an adjacent identical-type run, (b) an
// adjacent identical-type run of >=3, (c) an adjacent run of >=2 func-typed
// parameters (a DI-idiom rule, not a swap rule), and (d) a same-package named
// struct composite literal with >=2 unkeyed elements — the call-site half of the
// hazard, which `go vet`'s composites analyser covers only cross-package.
//
// The allow-list (tools/params/allowlist.txt) is a fail-on-stale ratchet: an
// entry that no longer matches a violation fails the gate, and
// `-update-allowlist` refuses to GROW it (the ADR 0020 fold R-51-1 rule).
//
// Policy: docs/decisions/0068-parameter-shape-gate.md.
// Truth: specs/truth/techstack.md (Build & Tooling, "Parameter-shape gate").
package params
