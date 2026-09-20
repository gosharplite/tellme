# Acceptance — the session listing reads like a conversation (round 073)

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`
**Feature under acceptance**: the offline `-l` / `--list` history listing's presentation

> PM-readable acceptance journeys in business language (no API/tool/technical detail).
> `axb-dsl-refine` later splits these into executable interface truth under
> `specs/truth/features/cli/history/**`; coverage requires every Rule here to be carried by
> at least one interface Rule.

## Feature: Listing a session shows who said what

### Rule: Each listed message is introduced by who said it

#### Example: A listed exchange names the operator and the model

- **Given** the session holds one completed exchange — the operator's prompt and the model's answer
- **When** the operator lists the session
- **Then** the operator's prompt is introduced by a line naming the **operator**
- **And** the model's answer is introduced by a line naming the **model**

#### Example: The naming is the same whether or not the output is viewed in a terminal

- **Given** the session holds one completed exchange
- **When** the operator lists the session with the output redirected to a file
- **Then** the operator's message is still introduced as the operator's
- **And** the model's message is still introduced as the model's

### Rule: The model's answer is presented as formatted prose, not as its raw source

#### Example: An answer containing emphasis reads as emphasis, not as markup

- **Given** the session holds an answer written in formatted prose
- **When** the operator lists the session
- **Then** the answer is presented as formatted prose
- **And** the answer's raw formatting marks are **not** shown

#### Example: An answer is shown as its raw source when the operator asks for raw output

- **Given** the session holds an answer written in formatted prose
- **When** the operator lists the session asking for raw output
- **Then** the answer's raw formatting marks **are** shown, unchanged

### Rule: The operator's own prompt is shown exactly as it was written

#### Example: A prompt is echoed verbatim, never reformatted

- **Given** the session holds a prompt written in formatted prose
- **When** the operator lists the session
- **Then** the operator's prompt is shown **exactly** as written
- **And** the prompt's raw formatting marks **are** shown, unchanged

### Rule: Consecutive messages are separated so the listing reads as a conversation

#### Example: A blank line separates the messages

- **Given** the session holds one completed exchange
- **When** the operator lists the session
- **Then** a blank line separates the operator's prompt from the model's message
- **And** no other extra blank lines are inserted

### Rule: The role lines are accented only when the output is a terminal

#### Example: A terminal listing accents the role lines

- **Given** the session holds one completed exchange
- **When** the operator lists the session to a terminal
- **Then** the operator's role line carries a distinct accent and the model's carries another

#### Example: A redirected listing carries no accents

- **Given** the session holds one completed exchange
- **When** the operator lists the session with the output redirected
- **Then** the listing carries **no** accent and the role lines are readable as plain text

### Rule: The listing stays private and offline

#### Example: Listing sends no request and does not change the session

- **Given** the session holds one completed exchange
- **When** the operator lists the session
- **Then** no request is sent to any model
- **And** the session's stored history is left unchanged
