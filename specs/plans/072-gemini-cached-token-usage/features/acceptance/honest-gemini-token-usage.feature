# Acceptance — honest token usage on a Gemini/Vertex turn (round 072)

**Plan Package**: `specs/plans/072-gemini-cached-token-usage`
**Feature under acceptance**: the metrics line's `H` / `Th` figures on the Gemini/Vertex family

> PM-readable acceptance journeys in business language (no API/tool/technical detail).
> `axb-dsl-refine` later splits these into executable interface truth under
> `specs/truth/features/cli/chat/**`; coverage requires every Rule here to be carried by
> at least one interface Rule.

## Feature: The assistant reports what a turn really cost

### Rule: A reused conversation prefix is reported as cached, not re-billed

#### Example: A turn that reuses most of its input reports the reused part as a cache hit

- **Given** the assistant is talking to a model that offers reused-conversation caching
- **And** the model reports that most of a turn's input was reused from the previous turn
- **When** that turn completes
- **Then** the assistant reports the reused portion as **cached (hit)** tokens, not as fresh input
- **And** the reported cost reflects the cheaper cached rate for that reused portion

#### Example: A turn whose entire input was reused reports no fresh input

- **Given** the model reports that **all** of a turn's input was reused
- **When** that turn completes
- **Then** the assistant reports **zero** fresh-input tokens for that turn
- **And** the reported cost reflects only the cached rate for the reused portion

### Rule: The assistant reports reasoning tokens when the model uses them

#### Example: A turn where the model reasons reports those reasoning tokens

- **Given** the model reports that a turn used reasoning (thinking) tokens
- **When** that turn completes
- **Then** the assistant reports that reasoning total for the turn

### Rule: The session totals agree with the turn figures

#### Example: A multi-turn session sums the per-turn figures

- **Given** a session with several completed turns
- **When** the assistant reports the session's running totals
- **Then** the cached and fresh-input totals equal the sum of the per-turn figures
