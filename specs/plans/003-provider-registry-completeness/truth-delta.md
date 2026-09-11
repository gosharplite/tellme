# Truth Delta: 003-provider-registry-completeness

**Plan Package**: `specs/plans/003-provider-registry-completeness`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | (i) **Configuration** section: expanded `Provider entry schema` to model full typed attributes (`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`); added **Variable expansion** row specifying stdlib regex-based `${VAR}` and `${VAR:-default}` expansion in credentials, URLs, and headers; updated **Effective-value resolution & validation** row with active provider validation and frozen class phrase `tellme: the provider configuration is invalid`. (ii) **Testing & Verification** section: updated `Pure-helper unit tests` row to include `${VAR}` expansion and provider validation pure helpers. | Round-003 research Decisions 1–4: implements Clarify Round 1 Q1 (Option 1 - core request set), Q2 (Option 1 - targeted expansion with error on unset), and Q3 (Option 1 - exit 3 with dedicated class phrase). Prepares provider configuration model for Slice 004 LLM requests. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no OpenAPI/HTTP surface, so there is no API contract to author or change this round. | `contract-authoritative` holds vacuously. Provider registry schema changes are CLI-end input behavior owned by `/axb-dsl-refine` (`specs/truth/features/cli/**`). |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Configuration is input data loaded at startup; no new database tables, state entities, or in-memory persistence models are introduced. | `data-model-covers-all-state` holds vacuously. Handed off to `/axb-dsl-refine` as CLI input contract behavior. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/dsl.md` (interface root) | Added `the provider configuration is invalid` to the frozen class phrase vocabulary in the `tellme explains on stderr that "{reason}"` row (now eight frozen class phrases). | Round-003 FR-007 / Clarify Round 1 Q3: establishes the dedicated failure class phrase for provider validation failures while reusing configuration error exit code 3. |
| MODIFY | `specs/truth/features/cli/configuration/dsl.md` | Added 8 module-specific Given rows: (1) `a well-formed configuration "{config_path}" where provider "{provider}" specifies:` (DataTable), (2) `the provider "{provider}" in configuration "{config_path}" includes custom headers:` (DataTable), (3) `a well-formed configuration "{config_path}" where provider "{provider}" specifies only mandatory fields:` (DataTable), (4) `the environment variable "{var_name}" is set to "{var_value}"`, (5) `the environment variable "{var_name}" is unset`, (6) `a configuration "{config_path}" where selected provider "{provider}" is missing "{field}"`, (7) `a configuration "{config_path}" where selected provider "{provider}" has negative "{field}"`, (8) `a well-formed configuration "{config_path}" where provider "{provider}" references unset variable "{var_name}"`. | Round-003 spec FR-001..FR-009: supports executable step definitions for complete provider attributes, optional field omission, `${VAR}` and `${VAR:-default}` dynamic expansion, missing required fields, negative limits, and unresolvable variable references. |
| MODIFY | `specs/truth/features/cli/configuration/starting-with-a-configuration.feature` | Added 4 Rules (6 Examples): (1) `Rule: A run proceeds when the selected provider declares complete typed attributes` (all core fields / omit optional fields), (2) `Rule: Dynamic environment variable expressions in the selected provider must expand at load time` (`${VAR}` and `${VAR:-default}`), (3) `Rule: A run stops when the selected provider entry is malformed or missing mandatory fields` (missing `MODEL` / negative `MAX_TOKENS`), (4) `Rule: A run stops when an environment variable referenced without a default is unset`. | Carries round-003 acceptance criteria (Stories 1, 2, and 3) into the CLI executable contract. Mechanical topology audit passed (165 steps). |
