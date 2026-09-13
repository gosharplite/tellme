# Spec Quality Checklist: tellme Vertex AI / Gemini Provider Support (round 013)

**Created**: 2026-09-13

**Feature Directory**: `specs/plans/013-vertex-gemini-provider`

**Spec Path**: `specs/plans/013-vertex-gemini-provider/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear
- [x] No implementation technology, framework, or code detail is written as a requirement (the transport mechanism and the OAuth2 mechanics are stated as observable constraints; the exact wire mapping is left to research/DSL)
- [x] Edge cases cover the main high-risk scenarios (non-Vertex URL, anthropic/unknown families, empty persona, zero limits, unset `${VAR}`, unreachable token endpoint, tool-call-only turn, absent usage, non-transport paths)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (answer path P1 before credential handling P2, which depends on it)
- [x] Each user story is independently verifiable
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries (OpenAI-family invariance, no-regression, family mapping must not widen)
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify` (scope, credential shape, dependency posture)
- [x] This round's clarify questions were held to the budget (Round 1 = 3 questions)
- [x] Low-risk undecided details are disclosed via assumptions (A1–A7) rather than asked
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain — all resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (Clarify Round 1, 2026-09-13)

- **Q1 → Option 1 (scope)** — Vertex AI only; the Gemini API + ADC are deferred (`FR-001`, edge cases, `A1`).
- **Q2 → Option 1 (credential)** — `.json`-suffix service-account detection, no schema change (`FR-006`, `A2`).
- **Q3 → Option 1 (dependency)** — stdlib-only OAuth2; no new module (`FR-010`, `NFR-005`, `A3`).

### Folded assumptions (accepted, not re-asked)

- **A4** — credential read at turn time; missing/unreadable → `the provider request failed` (exit 6); `-d`/boot/`-l` unchanged.
- **A5** — Vertex request shape follows the reference.
- **A6** — provider `URL` + token endpoint are injectable for hermetic E2E.
- **A7** — rounds 001–012 semantics unchanged.

### Still open (non-blocking — later-phase determinations)

- **Exact wire mapping** — field names for `generationConfig`/`thinkingConfig`/`systemInstruction`/`tools[].functionDeclarations`, and the response extraction paths (`candidates[0].content.parts[]`, `functionCall`, `usageMetadata`) — a `/axb-technical-research` / `/axb-dsl-refine` determination.
- **Token cache lifetime** — process-scoped in-memory cache is assumed (`FR-008`); a persisted cache is a possible later refinement — `/axb-technical-research`.
- **Credential-failure taxonomy placement** — whether an unreadable key is `provider request failed` (exit 6, chosen, `A4`) vs `provider configuration is invalid` (exit 3); revisit if the implementation reveals a cleaner boundary — `/axb-technical-research`/DSL.
- **`google` TYPE alias** — whether `TYPE: "google"` is also accepted (reference parity) or only `gemini` — a `/axb-technical-research`/DSL determination (the spec admits both in `FR-001`).

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (Vertex answer path P1 · service-account auth P2). All clarify items are resolved; only research/DSL-level determinations remain. `/axb-spec-by-example` and `/axb-technical-research` can proceed in parallel.
