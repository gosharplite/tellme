# Requirements checklist — round 087 `087-mcp-tool-cache`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/087-mcp-tool-cache/`) — `fresh-package-per-round`
      (never a re-open of the frozen round-032 package).
- [x] Anchor: live issue **#180** (DoD = close it); the remedy is the recorded forward item
      `techstack.md` → *Not Introduced Yet → Cross-invocation MCP tool caching*.
- [x] Clarify **not escalated** — the issue fixes the goals (option A recommended, FR-1…7,
      NFR-1…3, W1…W6); the residual choices are RD-owned (`/axb-technical-research`) and do
      not change the story split or the acceptance logic (A5).
- [x] Every normative clause (FR / NFR / EC / SC) carries a **Verification Intent**.

## Requirements → judgeable carriers

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-001 | warm+fresh ⇒ zero dials, per-tool declarations offered | E2E `the MCP server has received no further connections` + `the request offered the tool "mcp_<s>_<t>"`; unit pin (fake cache + recording factory) |
| FR-002 | cold/corrupt ⇒ one bounded discovery + write | E2E cold Then (`the MCP server has been dialed` + `the MCP tool cache holds an entry for "github"`); unit pin |
| FR-003 | stale ⇒ cached served (no pre-dial) + post-answer refresh; failed refresh keeps prior | E2E stale-with-never-answering Then (offered from cache); unit pin on `Refresh` |
| FR-004 | changed decl ⇒ cold key | unit pin (declaration mismatch) |
| FR-005 | warm offer set/order == live | unit determinism pin (warm vs live) + the E2E offered-name Then |
| FR-006 | dropped/renamed tool ⇒ round-076 fold-back | E2E recoverable fold-back Then |
| FR-007 | server down at call ⇒ recoverable `error: …` | E2E recoverable-error Then |
| FR-008 | best-effort / no new failure mode | unit pins (corrupt file, write error) |
| FR-009 | no credential persisted | cache-file content Then + unit pin (`TOKEN` absent) |
| FR-010 | `--new` does not clear the cache | E2E `--new` Then |
| NFR-001 | offline paths network-free | the existing `make verify-no-network` + offline E2E Thens |
| NFR-002 | stdlib-only / no new dependency | `git diff go.mod go.sum` empty + `make lint`/`vet` |
| EC-001 | declaration mismatch ⇒ cold | unit pin (same as FR-004) |
| EC-002 | stale + never-answering server ⇒ cached tool still offered | E2E |
| EC-003 | entry for a server not in config ⇒ ignored | unit pin |
| EC-004 | no MCP_SERVERS ⇒ no cache effect | unit pin |
| SC-001…SC-005 | the four behaviours + gates green | the E2E/unit carriers above + `make check` |

## Boundaries

- **In**: the cache store and its JSON shape; `DiscoverCached` + the lazy client; the
  `deps.Discovery` refresh hook; the `MCPDiscoverer` home parameter; the composition-root
  wiring + `mcpToolCacheTTL`; the CLI post-answer refresh; the `techstack.md` / `chat`
  truth rows; ADR 0058 (+ index); the unit + E2E carriers.
- **Out**: a re-open of round 032; the stdio transport; the MCP `-d` diagnostic;
  MEMORY/PLUR; the tool-call envelope/naming/schema/auth/call-time-error contracts; an
  explicit refresh flag; a detached refresh helper; `go.mod`/`go.sum`.

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the falsifiable witnesses exist on day one (the
zero-connection warm Then; the cold-discovery Then; the stale-served Then; the
offer-set equality pin; the recoverable call-failure Then).
