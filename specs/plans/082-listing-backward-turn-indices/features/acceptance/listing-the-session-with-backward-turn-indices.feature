# Acceptance — the session listing shows how far back each turn is (round 082)

**Plan Package**: `specs/plans/082-listing-backward-turn-indices`
**Feature under acceptance**: the offline `-l` / `--list` history listing's role-header labels
**Anchor issue**: [#165](https://github.com/gosharplite/tellme/issues/165)

> PM-readable acceptance journeys in business language (no API/tool/technical detail).
> `axb-dsl-refine` later splits these into executable interface truth under
> `specs/truth/features/cli/history/**`; coverage requires every Rule here to be carried by
> at least one interface Rule.

## Feature: Listing a session shows how far back each turn is, matching what a rollback would remove

### Rule: Each listed message is introduced by its role and how far back its turn is

#### Example: The most recent exchange is numbered 1

- **Given** the session holds two completed exchanges
- **When** the operator lists the last four messages
- **Then** the newest exchange's prompt is introduced as the operator, one turn back
- **And** the newest exchange's answer is introduced as the model, one turn back
- **And** the earlier exchange's prompt is introduced as the operator, two turns back

#### Example: Only the last message is listed when the operator asks for one

- **Given** the session holds two completed exchanges
- **When** the operator lists the last message
- **Then** the single listed message is introduced as the model, one turn back

### Rule: The operator and the model of one exchange carry the same turn distance

#### Example: A tool-using exchange is numbered as one turn for both of its messages

- **Given** the session holds an exchange whose turn used a tool
- **When** the operator lists that exchange
- **Then** its operator message and its model message carry the **same** turn distance

### Rule: A partial listing keeps each message's true distance from the end

#### Example: A leading answer without its prompt keeps its true distance

- **Given** the session holds two completed exchanges
- **When** the operator lists the last three messages
- **Then** the leading model message is introduced as two turns back
- **And** the following two messages are introduced as one turn back

### Rule: The turn distance is accented together with the role only on a terminal

#### Example: A terminal listing accents the whole role label

- **Given** the session holds two completed exchanges
- **When** the operator lists the last two messages to a terminal
- **Then** the operator's turn-distance label carries the operator accent
- **And** the model's turn-distance label carries the model accent

#### Example: A redirected listing carries the label as plain text

- **Given** the session holds two completed exchanges
- **When** the operator lists the last two messages with the output redirected
- **Then** the turn-distance labels are readable as plain text and carry no accent

### Rule: The listing stays private and offline

#### Example: Listing sends no request and does not change the session

- **Given** the session holds two completed exchanges
- **When** the operator lists the last four messages
- **Then** no request is sent to any model
- **And** the session's stored history is left unchanged
