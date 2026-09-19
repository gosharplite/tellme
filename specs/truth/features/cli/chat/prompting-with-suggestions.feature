Feature: Prompting with live suggestions

  # Interface truth (CLI end, `chat` module). The `-i` interactive TUI prompt offers live
  # suggestions from three sources — recent prompts (the user-global shared log), workspace
  # paths (path-like queries only), and registered tools — while the operator types. (Round
  # 064 / ADR 0034: the shared log is the only prompt source; there is no separate session
  # source.) Acceptance journeys:
  # features/acceptance/composing-a-prompt-with-live-suggestions.feature and
  # features/acceptance/finding-a-recent-prompt-beyond-the-shallow-window.feature. Driven
  # end-to-end through the `TELL_ME_FORCE_STDIN_TTY` seam with a scripted key sequence; the
  # suggestion list is asserted as presence in the captured output (not exact ANSI bytes).

  Rule: The interactive prompt suggests a recent prompt that matches what the operator types

    Example: A recorded recent prompt is offered while typing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging with version 015"
      When the operator opens the interactive prompt and types "deploy"
      Then the interactive prompt offers the recent prompt "deploy to staging with version 015"
      And tellme exits successfully

  Rule: The interactive prompt searches a deep window of recent prompts

    # Round 064 (ADR 0034; operator request): the recent-prompt candidate pool is the
    # newest 50 distinct prompts (the reference's LoadTopN(ctx, 50)) — not 10 — so a
    # match older than the newest 10 is offered again; the surfaced list stays capped
    # at 10 (carried executably by the Rule below). Acceptance journey:
    # features/acceptance/finding-a-recent-prompt-beyond-the-shallow-window.feature.
    # [unit-pinned (round 064)] the engine's requested depth and the ≤10 cap arithmetic
    # are additionally pinned at the unit layer (internal/app/suggestions).

    Example: A recent prompt older than the newest ten is still offered
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "review the last two commits"
      And the shared prompt log holds 12 newer prompts about other topics
      When the operator opens the interactive prompt and types "commits"
      Then the interactive prompt offers the recent prompt "review the last two commits"
      And tellme exits successfully

  Rule: The prompt never shows more than ten suggestions

    # Round 064 (ADR 0034; review TD-1): the surface cap (10) is an executable
    # acceptance rule, not a comment — with more matching prompts than the cap,
    # the surplus is dropped. The fixture appends twenty distinct matching prompts;
    # read newest-first, the cap offers the newest ten and drops the rest, so
    # "commit note 20" (newest) is offered and "commit note 10" (the eleventh) is not.
    # [unit-pinned] the exact ≤10 arithmetic is additionally pinned at the unit layer
    # (internal/app/suggestions: the accumulator caps at maxSuggestions).

    Example: The ten nearest matches are offered and the surplus is dropped
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log holds 20 newer prompts that each mention "commit note"
      When the operator opens the interactive prompt and types "commit"
      Then the interactive prompt offers the recent prompt "commit note 20"
      And the interactive prompt does not offer the recent prompt "commit note 10"
      And tellme exits successfully

  Rule: The interactive prompt suggests a workspace entry for a path-like query

    Example: A local file is offered as the operator types a path
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator opens the interactive prompt and types "./"
      Then the interactive prompt offers the workspace entry "notes.txt"
      And tellme exits successfully

  Rule: The interactive prompt suggests an available tool for a matching query

    Example: A registered tool is offered while typing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt and types "read"
      Then the interactive prompt offers the available tool "read_files"
      And tellme exits successfully

  Rule: Accepting a suggestion inserts it into the editor

    # [unit-pinned (round 016, PR #40 F2)] the last-token replacement (a multi-word
    # line + a single-token suggestion) is pinned by the unit layer; this E2E Example
    # uses a whole-line replacement (a single-word query).

    Example: The operator accepts the current suggestion into the editor
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging with version 016"
      When the operator opens the interactive prompt, types "deploy", and accepts the current suggestion
      Then the interactive prompt holds the accepted suggestion "deploy to staging with version 016"
      And tellme exits successfully

  Rule: An over-long suggestion is not offered

    Example: A recent prompt spanning many lines is not offered
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "one\ntwo\nthree\nfour"
      When the operator opens the interactive prompt and types "one"
      Then no suggestion offered to the operator spans more than three lines
      And tellme exits successfully
