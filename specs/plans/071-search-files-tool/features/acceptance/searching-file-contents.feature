# Acceptance — searching file contents for a pattern (round 071)

**Plan Package**: `specs/plans/071-search-files-tool`
**Feature under acceptance**: a new agent tool `search_files`

> PM-readable acceptance journeys in business language (no API/tool/technical detail).
> `axb-dsl-refine` later splits these into executable interface truth under
> `specs/truth/features/cli/chat/searching-file-contents.feature`; coverage requires
> every Rule here to be carried by at least one interface Rule.

## Feature: Searching the project's files for a pattern

### Rule: A search reports where a pattern occurs

#### Example: A pattern that occurs in several files is reported with its locations

- **Given** a project whose files hold a repeated phrase in more than one file
- **When** the operator asks where that phrase is used
- **Then** the answer names each file (and the line) where the phrase occurs

#### Example: A pattern that occurs nowhere is reported as none

- **Given** a project whose files do not contain a particular phrase
- **When** the operator asks where that phrase is used
- **Then** the assistant reports that there were no matches — without failing

### Rule: A search never floods the conversation

#### Example: A search with very many matches reports only what fits, and says so

- **Given** a project where a phrase occurs far more often than one answer can carry
- **When** the operator asks where that phrase is used
- **Then** the assistant reports only what fits and marks the answer as trimmed

### Rule: A malformed pattern is reported plainly

#### Example: A bad pattern is reported as a mistake, not a crash

- **Given** the operator asks the assistant to search using a pattern that is not well formed
- **When** the assistant runs the search
- **Then** the assistant reports that the pattern was invalid — the conversation continues
