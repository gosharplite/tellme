# Truth Delta: 095-modelith-upstream-main-route

**Plan Package**: `specs/plans/095-modelith-upstream-main-route`
**Truth Root**: `specs/truth`

## /axb-technical-research

| 動作 | Truth 規格 | 改動摘要 | 原因 |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Build & Tooling* · the **Domain model** row (`:135`) | Reconcile the `modelith` install route **in place**: the first cell drops the `gosharplite/modelith` fork `@feat/self-domain-model` and now reads *upstream `stacklok/modelith`, tracked `@main`*; the dev-tool clause drops the *clone + pinned-build / pinned to commit `b4153541cee8`* text and now states the **upstream `main` HEAD** route (`go install github.com/stacklok/modelith/cmd/modelith@main`) + the **accepted non-hermetic cost** (**ADR 0065**). Preserved: dev-tool (not a `go.mod` dep), `go.mod`/`go.sum` unchanged, the zero-tolerance `modelith-check` gate, and the README single-source pointer. | Operator directive (2026-09-27): install from the HEAD of upstream `main`; the fork + pin route is superseded (ADR 0030 D2 → **ADR 0065**). `truth-current` requires **one** current description, not an annotated history (the round-088 F-088-1 / round-089 lesson). |

## /axb-api-plan

| 動作 | Truth 規格 | 改動摘要 | 原因 |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | 已檢查：本輪無 API 介面。 | 本輪只改 dev-tool 安裝路線與其紀錄，無 API 表面。 |

## /axb-data-plan

| 動作 | Truth 規格 | 改動摘要 | 原因 |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | 已檢查：本輪無資料/記憶體狀態變更。 | 本輪不新增或修改任何狀態。 |

## /axb-dsl-refine

| 動作 | Truth 規格 | 改動摘要 | 原因 |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/**`（CLI interface features + `dsl.md`） | 已檢查：本輪不改變任何可觀察 CLI 行為（無 step、無 DSL row）。 | 安裝路線不是使用者可見契約；`spec-by-example` 與 `dsl-refine` 皆為 NOOP。 |
