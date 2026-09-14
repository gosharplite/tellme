Feature: Prompting with live suggestions

  # Interface truth (CLI end, `chat` module). The `-i` interactive TUI prompt offers live
  # suggestions from three sources — recent prompts (the shared log + session), workspace paths, and
  # registered tools — while the operator types. Acceptance journey:
  # features/acceptance/composing-a-prompt-with-live-suggestions.feature. Driven end-to-end through
  # the `TELL_ME_FORCE_STDIN_TTY` seam with a scripted key sequence; the suggestion list is asserted
  # as presence in the captured output (not exact ANSI bytes).

  Rule: The interactive prompt suggests a recent prompt that matches what the operator types

    Example: A recorded recent prompt is offered while typing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging with version 015"
      When the operator opens the interactive prompt and types "deploy"
      Then the interactive prompt offers the recent prompt "deploy to staging with version 015"
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
