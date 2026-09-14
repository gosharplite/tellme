Feature: Presenting the interactive prompt

  # Interface truth (CLI end, `chat` module). The `-i` interactive TUI prompt presents the
  # tell-me-go surface — a bordered multi-line editor above a suggestion list, with the keybinding
  # hints in the placeholder and NO session metrics header. Acceptance journeys:
  # features/acceptance/seeing-a-prompt-that-matches-the-reference.feature and
  # features/acceptance/fitting-the-terminal-window.feature. Driven end-to-end through the
  # `TELL_ME_FORCE_STDIN_TTY` seam with a scripted key sequence; the chrome Thens assert
  # presence/absence in the captured output (not exact ANSI bytes).

  Rule: The interactive prompt frames a multi-line editor above a suggestion list

    Example: The prompt presents the reference surface
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging with version 016"
      When the operator opens the interactive prompt
      Then the interactive prompt is framed around a multi-line editor
      And the interactive prompt lists suggestions beneath the editor
      And the interactive prompt marks one suggestion as the current choice
      And tellme exits successfully

  Rule: The interactive prompt shows no session metrics header

    Example: The prompt shows no dashboard header
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      Then the interactive prompt shows no session metrics header
      And tellme exits successfully

  Rule: The editor advertises the submit and abort keys

    Example: The prompt advertises the keys in its placeholder
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      Then the interactive prompt advertises the submit and abort keys
      And tellme exits successfully

  Rule: The editor keeps a multi-line prompt and shows no line numbers

    Example: The operator composes a prompt across several lines
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt and types "line one\nline two"
      Then the interactive prompt keeps the typed text "line one\nline two"
      And the interactive prompt shows no line numbers
      And tellme exits successfully

  Rule: The interactive prompt draws the editor within a frame bounded by the terminal width

    Example: The prompt frames the editor
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      Then the interactive prompt frames the editor within the terminal width
      And tellme exits successfully
