# System Analysis Plan

## Project structure

### Document structure (this feature)

```text
specs/023-interactive-prompt-teardown/
├── spec.md
├── research.md
├── plan.md
├── features/acceptance/
│   ├── handing-back-the-terminal.feature
│   └── continuing-on-the-standard-surface.feature
├── ui/
│   ├── ui-plan.md
│   └── screens/{entry,10-submitted,20-abandoned}.txt
├── truth-delta.md
└── tasks.md
```

### Source structure (repository root)

```text
cmd/tellme/
internal/
├── cli/                     (runTUIPrompt → the -i submit resumes the standard surface + echoes the prompt)
├── ui/                      (turn chrome / spinner / post-turn status emitters — reused, unchanged)
└── ui/tui/prompt/           (model.go: clear the frame on submit/abort; run.go — unchanged driver)
tests/e2e/                   (steps + fake provider; the teardown/chrome witness)
specs/truth/features/cli/chat/  (interface truth: the -i surface rules + DSL rows)
specs/truth/techstack.md     (MODIFY — already recorded)
```

**Structure decision**: tellme has a single **CLI end**. This round is a presentation-surface change confined to the `-i` interactive prompt's submit/abort transition and the CLI's post-submit handoff to the already-existing standard turn surface (`internal/ui` + `internal/cli`). No new package, endpoint, or stored state.

## Analysis process planning

### System interface inventory

This round inventories `1` system interface.

1. `CLI end — the -i interactive prompt and its submitted-turn surface`
   - Endpoint type: `cli`
   - Primary interface: the `-i` interactive prompt's submit/abort transition (the editor frame clears on `Ctrl+S`/`Alt+Enter` and `Esc`/`Ctrl+C`) and the standard turn surface it hands to (the echoed prompt · input-capture acknowledgement · `─` rule + `╭─⠿ Turn <N>` header · live spinner · post-turn status).
   - Requirement evidence: `spec.md` US1/US2 · FR-001..FR-014 (the editor releases the terminal on submit; the submit continues on the standard surface; the prompt is echoed).

### Analysis process waves

#### Wave 1

- Interfaces analysed in parallel:
  - `CLI end`
- Analysis focus:
  - the `-i` model's submit/abort transition (clear the frame, no residue) and the CLI's post-submit wiring (chrome flag on, the echoed prompt on `stderr` before the input-capture line, the spinner gate now satisfied);
  - the review of the PM's terminal-mode `ui/**` (the entry frame + the cleared post-submit handoff + the abandoned frame) for落地性 — reviewed, not re-planned;
  - the truth-rule rewrite the round implies (the round-017 "no turn chrome" rule and the round-019 "`-i` excluded" spinner rule).
- Rationale: a single CLI end with no cross-interface dependency; one wave suffices.

### Planner delegation

- `/axb-api-plan` → **NOOP** (`specs/truth/contracts/**`; tellme has no OpenAPI surface).
- `/axb-data-plan` → **NOOP** (`specs/truth/data/**`; no stored-state change).
- `/axb-ui-plan` → **terminal mode**, produced by the PM (this package's `ui/ui-plan.md` + `ui/screens/*.txt`); reviewed here for可落地性, not re-planned or re-done.
- `/axb-dsl-refine` → the **CLI end's contract owner** carries the end forward: it rewrites the affected `specs/truth/features/cli/chat/**` rules and `dsl.md` rows (the `-i` teardown + echo; the turn-chrome and spinner re-scoping).
- `/axb-technical-research` → already delivered (`research.md` + `specs/truth/techstack.md` MODIFY).
