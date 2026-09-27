# System analysis — round 095 `095-modelith-upstream-main-route`

## 專案結構

### 文件結構（本功能）

```text
specs/plans/095-modelith-upstream-main-route/
├── spec.md
├── checklists/requirements.md
├── research.md
├── plan.md
├── tasks.md
└── truth-delta.md
```

**結構決策**: a toolchain/record round — no `features/acceptance/` (spec-by-example NOOP), no contracts/data.

## 分析流程規劃

### 系統介面的盤點

本次需求共盤點出 `1` 個系統介面。

1. `modelith dev-tool install route (Makefile + docs)`
   - 端點類型：`Toolchain / Make`（非出貨使用者介面）
   - 主要介面：`Makefile`（`MODELITH_REF` / `MODELITH_INSTALL` + the comment block）＋
     `docs/domain-model/README.md`（the single source）＋ `specs/truth/techstack.md`（the *Domain model*
     truth row）＋ `docs/decisions/0065-*.md`（the decision record）
   - 需求原文依據：operator directive 2026-09-27（issue [#200](https://github.com/gosharplite/tellme/issues/200)）

### 分析流程的安排

#### Wave 1

- 平行分析介面：
  - `modelith dev-tool install route`
- 分析重點：
  - the route (`@main`), the single-source discipline, the accepted non-hermetic cost, the superseding-ADR shape
- 安排理由：單一介面；其 truth owner 為 `/axb-technical-research`（the **Techstack** owner）。

## Wave 委派

- **`/axb-technical-research`** — the **Techstack** truth owner: `specs/truth/techstack.md` (*Domain model*
  row) **MODIFY**; authored the ADR 0065 decision record (`research.md` D1–D7).
- **`/axb-api-plan`** — **NOOP** (`truth-delta.md`): no API surface.
- **`/axb-data-plan`** — **NOOP** (`truth-delta.md`): no state.
- **`/axb-ui-plan`** — **skipped**: a line-oriented CLI, and this round has no UX surface.
- **`/axb-dsl-refine`** — **NOOP** (`truth-delta.md`): no `.feature`/`dsl.md` change (no observable CLI
  contract change).

## §5 — `docs/domain-model/**` 未建模（ADR 0041 escape hatch）

本輪不改變任何**已建模行為**：dev-tool binary 的**安裝路線**不是任何模型的 entity / relationship /
attribute / enum / glossary term / scenario。故 `docs/domain-model/*.modelith.{yaml,md}`
**不更新**（`make modelith-check` 仍須 green）。此即 ADR 0041 的「或在 plan package 記錄為何未建模」情形。

**邊界澄清（F-095-3 fold）**：產品模型的 `deterministic-and-hermetic` invariant（*"`make verify` is
hermetic (ADR 0012)"*，`tellme.modelith.yaml:746-747`）**不**被本輪觸及 —— ADR 0012 的 hermetic 邊界治理的是
**ambient Go-env invocation**（ADR 0012 D1/D5；其 **R1** 已載明 *"hermeticity is a `make`-boundary property,
not a toolchain property"*），而本輪改的是 **dev-tool 版本**，落在該邊界**之外**，故 invariant 的
`(ADR 0012)` 指涉不受影響，ADR 0041 的 same-PR 規則不適用。

## Artifacts touched

- `Makefile` — the modelith comment block + `MODELITH_REF` / `MODELITH_INSTALL`（removed `MODELITH_PIN`）。
- `docs/domain-model/README.md` — *Toolchain → Install*（single source）。
- `specs/truth/techstack.md` — the **Domain model** row（MODIFY）。
- `docs/decisions/0065-modelith-install-from-upstream-main.md`（**new**）+ `docs/decisions/README.md`
  （a new index row + a forward pointer on the ADR 0030 row）。
- `STATUS.md` + `docs/session-summary/2026/09/27/session-summary.md`（live state + day log）。
