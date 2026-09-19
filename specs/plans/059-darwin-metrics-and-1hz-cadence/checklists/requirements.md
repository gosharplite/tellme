# 規格品質檢查清單：real macOS CPU/MEM + a 1 Hz sample cadence (round 059)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/059-darwin-metrics-and-1hz-cadence`

**Spec 路徑**: `specs/plans/059-darwin-metrics-and-1hz-cadence/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用現況常數／檔案位置僅為可追蹤性）
- [x] 邊界情況已涵蓋主要高風險情境
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 real macOS CPU/MEM — P1；US2 1 Hz cadence — P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（US1: FR-001…FR-005；US2: FR-006…FR-008）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；S-1…S-6 為跨故事設計前提；I-1…I-3 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（Q1 CPU 來源、Q2 依賴、Q3 節流範圍、Q4 MEM 定義、Q5 平台範圍）
- [x] 本輪 clarify 題數控制在 1 至 5 題（5 題，一題一問）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A6）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **無殘留缺口**，Q1–Q5 皆已拍板，可進入後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **本輪修兩個 operator 回報缺陷**：(1) macOS 上 `[CPU: 0.0% | MEM: 0.0%]` — CPU 為 hardcoded stub、MEM 為 `syscall.Sysctl` NUL-trim 解碼錯誤；(2) CPU/MEM 每 200 ms frame 重新取樣（5 Hz），應為 1 Hz。
- **對照 tell-me-go**：`darwin && cgo` Mach `host_statistics64` + `darwin && !cgo` `runtime/metrics`；sysctl 走 `golang.org/x/sys/unix`；`updateSystemMetrics` 以 `metrics_shouldSample`（≥ 1 s）節流、braille 仍 200 ms。
- **cgo 決策**：允許 darwin cgo（Q1 → 1）；`make verify-cross-compile` 維持 `CGO_ENABLED=0`，此後僅覆蓋 darwin **nocgo** 腿（A1）。
- **依賴**：新增 `golang.org/x/sys`（Q2 → 1，A2）；port 維持百分比簽章（A3）。
- **A5** — 新增 **ADR 0029** 記錄 darwin sampler 分裂 + 1 Hz cadence。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: Q1–Q5 皆已拍板（一題一問），無殘留缺口；可進入 `/axb-spec-by-example`（US2 的 cadence 規則）與 `/axb-technical-research`（ADR 0029 + techstack row）。
