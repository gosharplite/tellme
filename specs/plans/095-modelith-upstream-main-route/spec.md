# Round 095 — `095-modelith-upstream-main-route`

**功能分支（round branch）**: `095-modelith-upstream-main-route` (off `dev` `248846c`)

**建立日期**: 2026-09-27

**狀態**: 草稿

**輸入（user directive，2026-09-27）**: 「*We need to use the HEAD of `stacklok/modelith` main branch (current HEAD hash) for `modelith` binary.*」 ⇒ revise **all live references**.

**Anchor issue**: [#200](https://github.com/gosharplite/tellme/issues/200). **DoD = close it.**

**Type**: toolchain / record round. **No product behaviour change** — no `internal/**`, no CLI surface, no `go.mod`/`go.sum` change.

---

## 1. Why this round

`modelith` is the dev-tool binary that renders & gates `docs/domain-model/**` (ADR 0030; round 060). Its
acquisition route lives in `docs/domain-model/README.md` (**the single source**), is **quoted** by the
`Makefile`'s `$(MODELITH_INSTALL)` and by the `modelith-check` failure message, and is **cited** by
`specs/truth/techstack.md` (the *Domain model* row) and **ADR 0030 D2**.

Round 060 recorded a **fork + immutable-commit clone** route, because the `gosharplite/modelith` *fork*
declares the **upstream** module path (`github.com/stacklok/modelith`) and so is not `go install`-able from
its own GitHub path:

```sh
git clone https://github.com/gosharplite/modelith && cd modelith && git checkout b4153541cee8 && go install ./cmd/modelith
```

The operator directive (2026-09-27) is that the binary **MUST be installed from the HEAD of the upstream
`stacklok/modelith` `main` branch** — an ordinary `go install` of the upstream module's own path. The fork
+ pin route is **superseded**. This round revises every **live** reference and records the superseding
decision (**new ADR 0065**). Because it changes a `TruthArtifact` (`techstack.md`) **and** supersedes a
recorded durable decision, it is a **round** (not a docs-only hygiene pass).

**Facts verified on the host (2026-09-27)** — see the fail-first evidence in `research.md` D1:

- upstream `main` HEAD `git ls-remote` / `gh api` = **`9008354f19ff13f24273a7c71395c63698c7fbac`**
  (pseudo-version `github.com/stacklok/modelith v0.5.1-0.20260927062055-9008354f19ff`).
- `go install github.com/stacklok/modelith/cmd/modelith@main` **resolves and builds** into a temp `GOBIN`
  (egress OK), producing the **same** `9008354f19ff` build.
- The installed binary is already that build, and `make modelith-check` is **green** with it ⇒ switching
  the *documented* route does **not** red the gate.

## 2. The change (surfaces)

1. **`docs/domain-model/README.md`** (*Toolchain → Install*, the single source) — the route becomes
   `go install github.com/stacklok/modelith/cmd/modelith@main`; drop the fork/branch/commit-pin; record the
   tracked ref, the **current HEAD hash** (`9008354f19ff`, provenance of the build the committed `.md`
   were last rendered with), and the **accepted cost** of tracking `main` (below).
2. **`Makefile`** — `MODELITH_PIN := b4153541cee8` → `MODELITH_REF := main`;
   `MODELITH_INSTALL := go install github.com/stacklok/modelith/cmd/modelith@$(MODELITH_REF)`; the comment
   block rewritten; the absent-binary message prints the new route.
3. **`specs/truth/techstack.md`** (*Domain model* row, `:135`) — reconcile **in place**: drop
   `gosharplite/modelith @feat/self-domain-model` + the commit pin; state the upstream `@main` route; keep
   the dev-tool / `go.mod` unchanged / zero-tolerance `modelith-check` facts + the README single-source
   pointer.
4. **`docs/decisions/0065-modelith-install-from-upstream-main.md`** (**new ADR**) + `docs/decisions/README.md`
   — record the decision and **supersede ADR 0030 D2**; ADR 0030's **body stays verbatim**, its **index row
   gains a forward pointer** to 0065 (ADR 0026).

**Accepted cost (recorded, not fixed)** — ADR 0030 D2 chose the immutable pin precisely because the
renderer affects the committed Markdown (*"a fork move ⇒ a committed `.md` renders differently ⇒ `verify`
reds on a re-install with no repo change"* — the ADR-0012 spurious-red class). Tracking `main`
**re-introduces** that exposure, deliberately; it is recorded in the README, the Makefile comment, the truth
row, and ADR 0065. The gate behaviour is unchanged: a `main` move is a **visible red**, never silent rot.

**Frozen history (MUST NOT be edited)** — `specs/plans/060-domain-model-and-drift-gate/**`,
`docs/archives/status/2026-09-19.md`, `docs/session-summary/2026/09/19/**` (they record the `B-060-1` /
`TD-060-1` fork-era decision — history, per `plan-package-frozen` / Rule 12).

**Not modelled** — the install *route* is not a modelled entity/enum/glossary/invariant and no behaviour
changes ⇒ `docs/domain-model/*.modelith.{yaml,md}` are **NOT** touched (ADR 0041 escape hatch, recorded in
`plan.md` §5).

---

## 3. 使用者情境與測試

### 使用者故事 1 - 從上游 `main` HEAD 安裝 `modelith`，並讓所有出貨文件一致 (Priority: P1)

作為 tellme 的操作者，我希望 `modelith` dev-tool 只從 **上游 `stacklok/modelith` `main` HEAD** 安裝，
且**每一份 live 文件**都描述同一條路線，讓任何人在新主機上照文件即可取得正確的工具，不再依賴私有
fork 與 commit pin。

**為何為此優先級**: 這是本輪唯一的操作者價值；route 是所有 domain-model gate 的前置條件。

**獨立驗證方式**: 在一台沒有 `modelith` 的環境照文件安裝 → 得到上游 `main` HEAD 的建置；並以 grep
證明 live tree 不再出現 fork/pin 路線（見 §5 SC、`tasks.md` W1/W2/W3）。

**驗收情境**:

1. **Given** 出貨文件（`docs/domain-model/README.md`、`Makefile`、`specs/truth/techstack.md`），
   **When** 讀取 install route，**Then** 三處都指向 `go install github.com/stacklok/modelith/cmd/modelith@main`
   且不再出現 fork/branch/commit-pin。
2. **Given** 主機上沒有 `modelith`，**When** 執行 `make modelith-check`，**Then** gate 以非零退出並印出
   **新的** route。

**功能需求（FR）**:

- **FR-001**: 出貨文件 MUST 以 `go install github.com/stacklok/modelith/cmd/modelith@main` 作為
  `modelith` 的安裝路線，並以 `docs/domain-model/README.md` 為唯一來源（single source）。 [Verification Intent: unobservable → carried manual grep predicate (W3)]
- **FR-002**: `Makefile` MUST 以 `MODELITH_REF := main` 搭配 `MODELITH_INSTALL := go install
  github.com/stacklok/modelith/cmd/modelith@$(MODELITH_REF)`，且 MUST 移除 `MODELITH_PIN`（fork commit）。 [Verification Intent: unobservable → carried manual inspection + W1]
- **FR-003**: `make modelith-lint`/`render`/`check` 的「absent binary」訊息 MUST 印出新 route。 [Verification Intent: observable → W1 (absent-binary gate message)]
- **FR-004**: `specs/truth/techstack.md` 的 **Domain model** row MUST 就地改寫為目前 route（刪除 fork/pin
  字樣、保留 dev-tool 與 gate 事實、保留 README single-source 指向）。 [Verification Intent: unobservable → carried manual inspection (W3)]
- **FR-005**: 一個新的 **ADR 0065** MUST 記錄此決定並 **supersede ADR 0030 D2**；ADR 0030 本文 MUST 保持
  verbatim，其 **index row** MUST 加上指向 0065 的 forward pointer。 [Verification Intent: unobservable → verify-adr-index + carried manual git-diff]
- **FR-006**: **frozen history** MUST 不被修改（`specs/plans/060-*`、`docs/archives/**`、`docs/session-summary/2026/09/19/**`）。 [Verification Intent: unobservable → git-diff predicate]
- **FR-007**: `go.mod`/`go.sum` MUST 不變；`modelith` MUST 維持 dev-tool binary（非 `go.mod` 依賴）。 [Verification Intent: unobservable → git-diff predicate]

**非功能需求（NFR）**:

- **NFR-001**: route MUST 以 **HEAD of `main`**（`@main`）表述，並記錄當下 HEAD hash
  `9008354f19ff` 作為觀測／最後渲染的 provenance。 [Verification Intent: unobservable → carried manual (research D2/D3)]

### 邊界情況

- **EC-001**: 當上游 `main` 前進時，route 不變（`@main`），但可能導致某個 committed `.md` 在重裝後渲染
  不同 → `modelith-check` 在**無 repo 變更**下轉紅。 [Verification Intent: accepted-unwitnessed — the ADR-0012 spurious-red class, deliberately accepted by ADR 0065 (recorded; not mechanically witnessed)]
- **EC-002**: 當專案尚未有 `docs/domain-model/**` 時，`modelith-check` 的「no `*.modelith.yaml` found」
  之預設拒絕行為 MUST 不受本輪影響。 [Verification Intent: unobservable → existing target logic unchanged (git-diff)]

---

## 4. 需求（全域）

> 本輪無跨故事的獨立全域 FR/NFR；US-1 已收納全部 FR/NFR。

### 關鍵實體

- **Modelith dev-tool binary**: 由上游 `github.com/stacklok/modelith`（`cmd/modelith`）建置的 dev-tool；
  非 `go.mod` 依賴。

## 5. 成功標準

### 可量測成果

- **SC-001**: live tree（排除 frozen history）不再出現任何 fork/pin route 字樣
  （`gosharplite/modelith`、`feat/self-domain-model`、`b4153541`），且新 route 字串存在於
  `docs/domain-model/README.md` 與 `Makefile`。 [Verification Intent: unobservable → grep predicate (W3)]
- **SC-002**: 在沒有 `modelith` 的 `PATH` 下，`make modelith-check` 以非零退出並印出新 route。 [Verification Intent: observable → W1]
- **SC-003**: `make verify` OK（含 `verify-adr-index`：ADR 0065 恰被索引一次；`modelith-check` no drift ×3）。 [Verification Intent: unobservable → make verify]
- **SC-004**: `git diff --name-only` 不含 `go.mod`/`go.sum`，也不含任何 frozen-history 路徑。 [Verification Intent: unobservable → git-diff predicate]

## 6. 假設

- **A1**: 操作者原話的「HEAD of `main` branch（current HEAD hash）」解讀為「追蹤 `@main`（HEAD）」，
  而 `9008354f19ff` 為**當下** HEAD hash（觀測值，非 pin）。若操作者意為「pin 當下 hash」，僅需把
  `MODELITH_REF` 改為該 hash（一處）；本輪預設採 `@main`（字面）。
- **A2**: 本輪不引入 clarify——directive 已鎖定 route；剩餘（route 拼寫、ADR 形狀）為 RD-owned。
- **A3**: `modelith` 在 CI/其他主機的可得性與本輪無關（本輪只改 route 文字與 gate 訊息，不改 gate 邏輯）。
