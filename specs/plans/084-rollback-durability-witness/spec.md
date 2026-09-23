# Feature Specification: Rollback durability witness — the `fsync` guarantee becomes falsifiable (round 084)

**Feature Branch**: `084-rollback-durability-witness`

**Created**: 2026-09-23

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the theme is anchored by issue [#169](https://github.com/gosharplite/tellme/issues/169) and the operator's directive *"Open a new aixbdd round, the goal is to close #169"*, and the issue's *Proposed resolution (a)/(b)* + the grill-round corrections are routed to `/axb-technical-research`)

**Input (anchor issue [#169](https://github.com/gosharplite/tellme/issues/169), 2026-09-23; operator directive)**: *"Round 081 (`-b`/`--back`, ADR 0053) asserts a durability/atomicity guarantee for the session-history rollback on six live surfaces, but only two of its three named mechanisms are falsifiable. Removing the file `fsync` or the directory `fsync` leaves the whole suite green — no input reddens."* — so the `fsync` clause must become **falsifiable**, or be **recorded as accepted-unwitnessed** and not asserted as truth. The fix is a **fresh round** — never a re-open of the frozen 081 package.

**Observed context (`dev` @ `030e142`, round-083 era)**: `internal/infrastructure/history/file_store.go` `writeRaw` (`:202`) writes the surviving lines to `history.jsonl.tmp`, calls `f.Sync()` (`:218`), `os.Rename`s it over `history.jsonl` (`:227`), then calls `syncDir(filepath.Dir(...))` (`:231`) whose error is swallowed (`_ = d.Sync()`, `:242`). The claim "*durable and atomic: temp file + `fsync` + atomic rename*" is asserted on **six live surfaces**: `specs/truth/techstack.md` (*Session history store* row) · `specs/truth/features/cli/history/dsl.md` (prologue) · `docs/decisions/0053-back-rollback-turns.md` (D3 + Consequences) · `docs/decisions/README.md` (index row 0053) · `docs/domain-model/tellme.modelith.{yaml,md}` (`history-rollback-removes-complete-turns`) · the frozen `specs/plans/081-back-rollback-turns/**`. The only pins on the rollback path today are `TestFileStore_Rollback_FailureLeavesPriorHistoryIntact` (fold `F-081-3` — reddens only for the temp+rename half) and `TestFileStore_Rollback_DoesNotRewriteSurvivorBytes` (`TD-081-2`). Removing `f.Sync()` or `syncDir(...)` keeps the suite green (measured 2026-09-23).

**Behaviour intent**: **MODIFY (the rollback store's witness surface — not its observable behaviour)** — make the durability guarantee **falsifiable** by binding it to a **mechanism-seam** witness (a unit pin that reddens when the file `fsync`-before-rename is removed), and **calibrate the asserted prose** so every live surface asserts only what is witnessed, with the best-effort directory `fsync` recorded as an **accepted-unwitnessed limit**. The observable rollback contract (`history.jsonl` shape, clamp, no-op, archive-not-touched, exit codes) is **unchanged**. **Governed by `aixbdd-tmg` ADR 0006** (*no normative clause may be promoted to truth without a witness; an unwitnessable claim is a decision — recorded in the ADR, not asserted as truth*). **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **No observable behaviour change (I-1).** The round adds/repairs a **witness** and **prose calibration**; it MUST NOT change the rollback's observable contract — the `history.jsonl` / `history.Entry` / `history.Step` shapes, the `-b`/`--back` semantics (clamp, no-op, archive-not-touched, exit codes, phrases), `--new`, the offline readers, or the ten-value exit-code set. A reader of `-b` output cannot tell this round shipped.
- **The clause must be witnessed or recorded, never left as prose (I-2).** Per ADR 0006, the durability guarantee is either carried by a **falsifier that reddens** (a mechanism-seam unit pin — see below) **or** demoted out of truth into an **accepted-unwitnessed** record on a decision surface; it MUST NOT remain asserted as truth while no input can falsify it.
- **Effect-strength calibration (I-3).** The asserted guarantee MUST name exactly the witnessed effect — the file `fsync` precedes the atomic `rename`. It MUST NOT claim a stronger effect (e.g. full power-loss durability) than the witnessed mechanism supports. A mechanism (`syncDir`) named on **no** asserted surface is **not** an asserted claim (the grill-round Q10 correction: the unit of obligation is the atomic *effect claim*, not the mechanism list).
- **The witness must discriminate (I-4).** The demonstration MUST be the claim's **discriminating** mutation (remove the file `fsync`), the failure MUST be **attributed** to the pin's intended assertion, and the mutation MUST then be **reverted and re-green** — the ADR 0006 three-gate DoD.
- **The mechanism seam is injectable, stdlib-only, POSIX-only, hermetic (I-5).** The witness rides an injected sync seam (no real crash is simulated); no new dependency; `verify-no-network` intact; the seam is test-visible only.
- **The six surfaces must agree (I-6).** Every live surface that asserts the guarantee (techstack row · `dsl.md` prologue · domain-model invariant · the ADR record) MUST assert the **same, calibrated** clause; the frozen 081 package is **history** and MUST NOT be edited.
- **The claim→witness ledger carries the clause (I-7).** `tasks.md` records the durability clause with its witness pointer (the ADR-0006 Claim→Witness ledger), and the round's `[WITNESS]` task is done only when the falsifier is shown to redden.

---

## Grounded in the current system *(measured 2026-09-23, `dev` @ `030e142`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/history/file_store.go` (`writeRaw`, `:202`–`:234`) | builds `buf` → `os.OpenFile(activePath+".tmp", O_CREATE|O_TRUNC|O_WRONLY)` → `f.Write(buf)` → **`f.Sync()`** (`:218`) → `f.Close()` → **`os.Rename(tmp, activePath)`** (`:227`) → `syncDir(filepath.Dir(activePath))` (`:231`) → `nil`. A sync error aborts, removes the temp file, returns the error. |
| `internal/infrastructure/history/file_store.go` (`syncDir`, `:235`–`:243`) | `os.Open(dir)`, `_ = d.Sync()`, `_ = d.Close()` — **best-effort** (a directory fsync is unsupported on some filesystems); its error is deliberately swallowed. |
| `internal/infrastructure/history/file_store_rollback_test.go` | the round-081 pins: `…_Normal` · `…_Clamp` · `…_NonPositiveIsNoop` · `…_MissingFile` · `…_DoesNotTouchArchive` · `…_LeavesNoTempFile` · `…_FailureLeavesPriorHistoryIntact` (F-081-3) · `…_DoesNotRewriteSurvivorBytes` (TD-081-2) · `…_DecodeFailureLeavesFileIntact` · `…_LeavesAPreExistingArchiveByteIdentical`. **None** asserts that the file `fsync` runs (removing `f.Sync()` keeps them all green). |
| `specs/truth/techstack.md` — *Session history store* row | "…the surviving entries are written to a temp file + `fsync` + atomic `rename`, so a crash mid-rollback leaves the prior `history.jsonl` intact". |
| `specs/truth/features/cli/history/dsl.md` — prologue | "The store's durable `Rollback` (temp-file + fsync + rename) is the single owner of the truncation." |
| `docs/decisions/0053-back-rollback-turns.md` | D3 + Consequences assert "**durable and atomic** … a crash mid-rollback leaves the prior history intact". |
| `docs/domain-model/tellme.modelith.{yaml,md}` (`history-rollback-removes-complete-turns`) | "it is **durable** (temp-file + fsync + atomic rename, so a crash mid-rollback leaves the prior history intact)". |
| `GAPS.md` §2/§3/§7 · `aixbdd-tmg#15` (closed) → `#16` (merged, **ADR 0006**) | the "claim-without-a-tripwire" class record; the obligation this round discharges. |

---

## Design (locked by the issue/operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **L-1** | The round's deliverable is a **falsifier** for the durability clause (or an **accepted-unwitnessed** record), with **no** observable behaviour change. | **locked** (issue Summary / Non-goals) |
| **L-2** | The **fresh round** — never a re-open of the frozen 081 package. | **locked** (`plan-package-frozen`) |
| **L-3** | The round MUST use the ADR-0006 mechanism: classify the clause, bind it to a witness tier (**mechanism-seam** or **accepted-unwitnessed**), and record the claim→witness mapping in `tasks.md`. | **locked** (ADR 0006) |
| **L-4** | The **observable** rollback contract is **unchanged**; the exit-code set stays **ten**; the frozen phrases are unchanged. | **locked** (issue Non-goals) |
| **L-5** | The **grill-round corrections** apply: unit of obligation = the atomic *effect claim* (not the mechanism list); the directory `fsync` is not an asserted claim; effect-strength calibration is required. | **locked** (issue §Grill-round corrections) |
| **S-1** | **The resolution** — **(a)** make the clause executable via a **mechanism seam** (an injected `syncer`/`syncDirFn` on the store) with a unit pin asserting the file `fsync` runs **before** the `rename`; vs **(b)** **demote** the clause out of every asserted surface into an accepted-unwitnessed `ADR §Forward` record. | research decision (D-x); **proposed: (a) — witness it** (keeps the guarantee asserted *and* guarded), with the **directory** `fsync` demoted to a recorded best-effort limit. |
| **S-2** | **The exact seam shape** — a `syncer` interface (e.g. `Sync(*os.File) error` + `SyncDir(string) error`) injected into `fileStore` with an `os` default; a recording fake in the test; where the seam is constructed (the adapter vs `deps`). | research decision (D-x) |
| **S-3** | **Which surfaces change** — the `techstack.md` *Session history store* row, the `history/dsl.md` prologue, the domain-model invariant `history-rollback-removes-complete-turns`, **and** a new **ADR** (amending/annotating ADR 0053); whether `history.jsonl`/`history.Store`/contracts/data are NOOP. | research decision (D-x); **proposed: MODIFY the three prose surfaces + a new ADR; annotate ADR 0053; `contracts/**` + `data/**` NOOP; the domain model **is** modelled (an invariant wording change — re-render).** |
| **S-4** | **Scope excludes** — no mutation-testing framework, no coverage threshold, no CI gate; no change to `-b` semantics; the grill-flagged **byte-unchanged** sub-clause is either pinned or explicitly recorded (a D-x choice). | research decision (D-x) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — No observable behaviour change.** Rollback semantics, shapes, exit codes, and phrases are unchanged.
- **I-2 — Witnessed or recorded.** The durability clause is carried by a falsifier that reddens, or homed as accepted-unwitnessed — never left as unguarded truth.
- **I-3 — Effect-strength calibration.** Asserted guarantee = the witnessed effect (file `fsync` before `rename`); no stronger claim.
- **I-4 — The witness discriminates.** Remove-the-file-`fsync` reddens the pin; failure attributed; revert-and-re-green.
- **I-5 — Injectable, stdlib-only, POSIX-only, hermetic.**
- **I-6 — The asserted surfaces agree.** All live surfaces assert the identical calibrated clause; the frozen 081 package is untouched.
- **I-7 — The claim→witness ledger** in `tasks.md` carries the clause + its witness pointer.

---

## 使用者情境與測試 *(必填)*

### 使用者情境 1 - 回滾的耐久性保證可被否證 (Priority: P1)

作為維護 tellme 的 RD，當我閱讀任何斷言「回滾是耐久且原子」的表面時，我希望該保證有一個**能變紅的見證**（移除檔案 `fsync` 會讓一個測試失敗），因為一個無法被否證的保證會讓後續回合「從它推論」（durability is handled），而靜默的回歸無法被看見。

**為何為此優先級**: 這是 issue [#169](https://github.com/gosharplite/tellme/issues/169) 的核心缺陷與唯一交付面 —— 讓 `fsync` 子句可被否證（或明確記錄為 accepted-unwitnessed）；價值明確、風險受限（只加見證與 seam，不動可觀察行為）。

**獨立驗證方式**: 以注入的 sync seam（recording fake）執行 `Rollback`，斷言 store **先對暫存檔 `fsync` 再 `rename`**；並以**移除檔案 `fsync` 的 mutation** 驗證該 pin **變紅**（復現後回退、再變綠）。

**驗收情境**:

1. **Given** a session history with N complete turns, **When** the operator runs a rollback (`-b 1`), **Then** the store flushes the temporary file (`fsync`) **before** it atomically renames it over `history.jsonl`, and the rollback still removes exactly the last turn.
2. **Given** the file `fsync` is removed from the rollback path, **When** the witness runs, **Then** the witness **reddens** naming the missing durability step (and, reverted, the suite is green again).

**功能需求（FR）**:

- **FR-001**: 回滾路徑在原子 `rename` 之前 MUST 對暫存檔執行 `fsync`；此機制 MUST 可被注入（a sync seam）以便測試斷言其**順序**。[Verification Intent: unobservable → mechanism-seam]
- **FR-002**: 系統 MUST 提供一個單元見證（unit pin），在**移除檔案 `fsync`**（該宣稱的 discriminating mutation）時**變紅**。[Verification Intent: unobservable → mechanism-seam]
- **FR-003**: 回滾的可觀察契約 MUST 維持不變 —— 移除最後 N 個完整回合（clamp）、`n ≤ 0` 為 no-op、archive 不被觸碰、缺檔視為 0 removed、結束碼集合維持十個（I-1）。[Verification Intent: observable → 既有 round-081 驗收 Examples]

**非功能需求（NFR）**:

- **NFR-001**: 見證 MUST 為決定性、hermetic、stdlib-only、POSIX-only —— 不新增依賴；MUST NOT 真的模擬當機或觸網（`verify-no-network` MUST NOT 被破壞）。[Verification Intent: unobservable → unit pin]

---

### 使用者情境 2 - 斷言的文字與實際被見證的效果一致 (Priority: P2)

作為一個閱讀 tellme truth 的 RD，當我讀到耐久性保證時，我希望它斷言的**恰好是被見證的效果**（暫存檔 + `fsync`-before-`rename` + 原子 `rename`），而不是更強的宣稱（例如完整的斷電耐久性）；被 best-effort 實作的目錄 `fsync` MUST NOT 被斷言為保證，而 MUST 被記錄為一項限制。

**為何為此優先級**: 這是 US1 的一致面 —— 見證若成立，但文字仍過度宣稱（或反之），缺陷只是換了位置；它在 US1（見證）之上，故為 P2。

**獨立驗證方式**: 檢視每一個斷言耐久性的表面（`techstack.md` row、`history/dsl.md` prologue、domain-model invariant、ADR 紀錄），確認它們斷言**同一、且僅止於**被見證的子句；並確認目錄 `fsync` 的 best-effort 限制被記錄（accepted-unwitnessed）。

**驗收情境**:

1. **Given** the durability clause is asserted on the truth/domain-model surfaces, **When** the surfaces are read against the witnessed mechanism, **Then** each asserts exactly the witnessed effect (temp file + `fsync`-before-rename + atomic rename) and names no stronger guarantee.
2. **Given** the directory `fsync` is best-effort (its error swallowed), **When** the decision surface is read, **Then** it is recorded as an explicit accepted-unwitnessed limit and is **not** asserted as a guarantee on any truth surface.

**功能需求（FR）**:

- **FR-004**: 任何斷言耐久性的表面 MUST 斷言**恰好**被見證的子句（暫存檔 + `fsync`-before-`rename` + 原子 `rename`）（I-3）；MUST NOT 斷言超出見證機制的更強保證。[Verification Intent: unobservable → doc review + ADR record]
- **FR-005**: best-effort 的目錄 `fsync` MUST NOT 被斷言為保證；它 MUST 以 accepted-unwitnessed 限制記錄於決策面（ADR §Forward）並在 surviving truth row 上留下被見證半邊的指標。[Verification Intent: unobservable → ADR record]

---

### 邊界情況

- **EC-001**: 當暫存檔的 `fsync` 回傳錯誤時，系統 MUST 中止回滾（不 `rename`）、移除暫存檔、回傳錯誤，並讓先前的 `history.jsonl` **保持不變**。[Verification Intent: unobservable → unit pin]
- **EC-002**: 當目錄 `fsync` 不受支援（回傳錯誤）時，系統 MUST NOT 讓回滾失敗（best-effort）——此限制 MUST 被記錄，而非被斷言為保證。[Verification Intent: unobservable → unit pin + ADR record]
- **EC-003**: 當回滾成功時，系統 MUST NOT 留下暫存檔（既有 `…_LeavesNoTempFile` 契約不變）。[Verification Intent: observable → 既有 round-081 驗收 Examples]
- **EC-004**: 當倖存行的位元「MUST 不被更動」時，該子句若無專屬承載（grill 標記的未解子句），MUST 明確被 pin（一則新的單元見證）**或**被記錄為 accepted-unwitnessed —— 不得留為無承載的 prose。[Verification Intent: unobservable → unit pin 或 ADR record（由研究決議 D-x）]

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 本輪 MUST 在 `tasks.md` 的 Claim→Witness 盤點對照表中，以**原子效果宣稱**為單位記錄耐久性子句及其見證指標（測試 id 或 `none → ADR NNNN §Forward`）（I-7）。[Verification Intent: unobservable → tasks.md ledger]
- **FR-007**: 本輪 MUST NOT 修改凍結的 `specs/plans/081-back-rollback-turns/**`；MUST NOT 改變 `-b`/`--back` 語意、`history.Entry`/`history.Step` 形狀、互動提示、即時 turn chrome、或結束碼/片語集合（I-1/I-6）。[Verification Intent: observable → 既有 round-081 驗收 Examples]

#### 非功能需求

- **NFR-002**: 見證與 seam MUST 為決定性且 hermetic；MUST NOT 新增任何依賴；`go.mod`/`go.sum` MUST 維持不變。[Verification Intent: unobservable → unit pin]
- **NFR-003**: 結束碼集合 MUST 維持十個；provider 片語詞彙 MUST 維持凍結；MUST NOT 因本輪新增任何片語或結束碼。[Verification Intent: observable → `TestExitCodesMatchPinnedContract`]

### 關鍵實體 *(若功能涉及資料，必填)*

- **回滾耐久性宣稱（rollback durability claim）**: 一個原子效果宣稱 —— 「倖存行被寫入暫存檔、在原子 `rename` 前 `fsync`、再 `rename` 覆蓋 `history.jsonl`，故回滾中途當機時先前的 `history.jsonl` 保持完整」。單位 = **效果宣稱**（非機制清單）。
- **sync seam（注入點）**: 迴圈/store 可注入的同步邊界 —— 檔案 `Sync` 與目錄 `SyncDir`；預設為 `os` 實作；測試以 recording fake 斷言呼叫**順序**。
- **`history.jsonl` / `history.Entry` / `history.Step`**: 既有形狀 —— **不變**；本輪只加見證與校準文字。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 存在一個見證，在**移除檔案 `fsync`**（mutation 復現後回退）時**變紅**；在 shipped 程式碼上為綠。[Verification Intent: unobservable → unit pin]
- **SC-002**: `make check` 全綠；E2E 契約**不變**（無可觀察行為改變，scenario/step 數不因本輪上升或下降）。[Verification Intent: observable → `make check`]
- **SC-003**: 每一個斷言耐久性的存活表面（`techstack.md` row、`history/dsl.md` prologue、domain-model invariant）斷言**同一、且僅止於**被見證的子句；目錄 `fsync` 被記錄為 accepted-unwitnessed 限制。[Verification Intent: unobservable → doc review + ADR record]
- **SC-004**: `tasks.md` 的 Claim→Witness 盤點表列出耐久性宣稱及其見證指標（I-7）。[Verification Intent: unobservable → tasks.md ledger]
- **SC-005**: `go.mod`/`go.sum` 不變；`make verify` 通過（含 `modelith-check` 無 drift）；`TestExitCodesMatchPinnedContract` 綠。[Verification Intent: observable → `make verify`]

## 假設

- **A1**: 本輪 anchor = issue [#169](https://github.com/gosharplite/tellme/issues/169)；**DoD = 關閉它**。修復 MUST 為 **fresh round**（`plan-package-frozen`）。
- **A2**: 核心目標 —— 「耐久性子句 MUST 被見證或明確記錄為 accepted-unwitnessed；可觀察行為不變；以原子效果宣稱為單位；目錄 fsync 非斷言宣稱」—— 由 issue 與 `aixbdd-tmg` ADR 0006 鎖定；issue 的 *Proposed resolution (a)/(b)* 與 grill 修正（S-1…S-4：解析方式、seam 形狀、受影響表面、範圍排除、未解子句）由 `/axb-technical-research` 決議，故本輪 clarify **未升級（0 題）**。
- **A3**: 回滾的**可觀察**行為在 `dev` @ `030e142` 已正確且已由 round-081 的驗收 Examples 覆蓋；本輪不新增可觀察驗收（無新的 Gherkin 行為），見證住在單元層（ADR 0006：unobservable clause → unit pin）。
- **A4**: 現行套件在移除檔案 `fsync` 或目錄 `fsync` 時**全綠**（2026-09-23 量測）—— 本輪的相鄰性缺陷目前**無任何測試可變紅**（`GAPS.md` / `aixbdd-tmg#15` 類）。本輪 MUST 補上該見證。
- **A5**: 預期新增一枚 **ADR**（記錄 seam、見證層級、effect-strength 校準、目錄 fsync 的 accepted-unwitnessed 限制）並在 ADR 0053 上加註承接指標；MODIFY `techstack.md` *Session history store* row、`history/dsl.md` prologue、domain-model invariant（re-render）；`contracts/**` 與 `data/**` 預期 NOOP；是否新增 interface Gherkin 由 `/axb-dsl-refine` 判定（預期 NOOP 或極薄，因無可觀察行為）。
- **A6**: 本輪**不**實作 mutation-testing 框架、coverage 門檻或 CI gate；MUST NOT 以真實當機模擬見證（改以注入 seam）。
