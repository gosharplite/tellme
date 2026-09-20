# Tasks — the media channel in-band (round 070)

**Plan Package**: `specs/plans/070-media-channel-in-band`
**Inputs**: `spec.md` · `research.md` (D1–D7) · `plan.md` · `truth-delta.md`
**Anchor**: issue [#142](https://github.com/gosharplite/tellme/issues/142)

**Scope reminder**: a **pure internal-shape refactor**. No config change, no UX change, no truth-behaviour change.

---

## Phase 1 — Setup

_None._ (No new dependency; the media type is relocated, not added.)

## Phase 2 — Foundational

- [x] **T001** [CODE] `internal/domain/tools/media.go` **NEW** — `MediaPart{MIMEType, Data}` (relocated from `domain/llm`).
- [x] **T002** [CODE] `internal/domain/tools/tools.go` — add the optional `MediaTool` capability interface (embedding `Tool`; `ExecuteMedia(ctx, args, budget) (text, []MediaPart, err)`).

## Phase 3 — Test Alignment & Implementation

- [x] **T003** [CODE] `internal/domain/llm/gateway.go` — `Message.Media []tools.MediaPart` (+ import); **delete** `internal/domain/llm/media.go` (the collector + the old `MediaPart`).
- [x] **T004** [CODE] `internal/infrastructure/tools/image.go` — implement `ExecuteMedia` (in-band media) + a **delegating** `Execute`; **drop** the `internal/domain/llm` import.
- [x] **T005** [CODE] `internal/agent/agentloop.go` — type-assert `tools.MediaTool`; **drop** the per-call collector; the media fold is unchanged.
- [x] **T006** [CODE] `internal/infrastructure/llm/{gemini,openai}/client.go` — serialize `[]tools.MediaPart`.
- [x] **T007** [TEST ALIGN] Media-type references updated (`internal/domain/llm/{token,unpaired}_test.go`, `internal/infrastructure/llm/{gemini,openai}/*_image*_test.go`); `internal/infrastructure/tools/image_test.go` rewritten to `ExecuteMedia` (**no assertion weakened**).

## Phase 4 — Verification & Regression

- [x] **T008** [TRUTH] `specs/truth/techstack.md` MODIFY (*Image filesystem tool*) + **ADR 0040** + the index row; **annotate ADR 0032** (`Status`, `D7a`, `RF-062-10`) and **ADR 0039** (`Status`, `RF-069-1`).
- [x] **T009** [GATES] `gofmt`/`go vet`/`go build` clean; `go test -count=1 ./...` green; `make verify` **OK**; `go.mod`/`go.sum` unchanged.
- [x] **T010** [WITNESS] Structural witnesses + falsifiability witnesses.

---

## Implementation ledger

### Structural witnesses (SC-001)

```
$ grep -rn 'WithMediaCollector\|AttachMedia' --include=*.go .
(no matches)                                  # the collector is gone
$ grep -rn 'internal/domain/llm' internal/infrastructure/tools/*.go | grep -v _test
(no matches)                                  # infra/tools no longer imports the conversation model
$ grep -n 'tools.MediaTool' internal/agent/agentloop.go
if mt, ok := tool.(tools.MediaTool); ok {     # the loop asserts the capability
```

### Behaviour-identity (SC-002) — no assertion weakened

The only test edits are the **media-type relocation** (`llm.MediaPart` → `tools.MediaPart`) and the rewrite of `image_test.go`'s three collector-based cases to `ExecuteMedia` (same assertions: one `image/png`, exact bytes; the not-a-picture refusal with no media; the ceiling boundary). The full suite (incl. `reading-a-local-image` + `calling-several-tools-in-one-round`) is **green**.

### Falsifiability witness (T010, reproduced then reverted)

- Disable the loop's in-band fold (`if len(media) > 0` → `if false`): the E2E **reds** — `the request did not carry the image file "…" (0 image block(s) recorded)` for five image scenarios. **Reproduced then reverted** → green. ⇒ the in-band `MediaTool` return is the **carrier** (the collector's removal did not orphan the media).

### Gates (T009)

```
gofmt -l .                      (clean)
go vet ./...                    (clean)
go build ./...                  (clean)
go test -count=1 ./...          ok (all packages incl. tests/e2e)
make verify                     verify: OK
topology audit                  49 features · 6 modules · 16 root + 384 module rows · 1989 steps · 5 pre-existing, none new
git diff --stat go.mod go.sum   (unchanged)
```
