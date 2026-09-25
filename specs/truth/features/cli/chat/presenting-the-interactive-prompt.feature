Feature: Presenting the interactive prompt

  # Interface truth (CLI end, `chat` module). The `-i` interactive TUI prompt presents the
  # tell-me-go surface — a bordered multi-line editor above a suggestion list, with the keybinding
  # hints in the placeholder and NO session metrics header. Acceptance journeys:
  # features/acceptance/seeing-a-prompt-that-matches-the-reference.feature,
  # features/acceptance/showing-no-chosen-hint-at-rest.feature, and
  # features/acceptance/fitting-the-terminal-window.feature. Driven end-to-end through the
  # `TELL_ME_FORCE_STDIN_TTY` seam with a scripted key sequence; the chrome Thens assert
  # presence/absence in the captured output (not exact ANSI bytes).
  #
  # Round 037: the suggestion list opens with NO suggestion pre-selected (the selection starts at a
  # no-choice sentinel and resets on every refresh), so the at-rest assertion is "marks NO suggestion
  # as the current choice". This SUPERSEDES round 016's "marks one suggestion as the current choice"
  # (which followed the reference's mockup; the reference's code never pre-selects). The first `Tab`
  # still selects the first suggestion — carried by the accept journey in prompting-with-suggestions.feature.
  #
  # Round 093 (issue #191): a scripted line break is delivered as the terminal Enter byte (CR), so the
  # product inserts a newline and the typed lines stay apart; and "keeps the typed text" now requires each
  # line of a multi-line value on its OWN editor row — the pre-093 substring check passed vacuously on a
  # joined row (`line oneline two`). The scripted keys are delivered after the composed frame paints (a
  # content-aware handshake), so the composed frame is always captured.

  Rule: The interactive prompt frames a multi-line editor above a suggestion list

    Example: The prompt presents the reference surface
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging with version 016"
      When the operator opens the interactive prompt
      Then the interactive prompt is framed around a multi-line editor
      And the interactive prompt lists suggestions beneath the editor
      And the interactive prompt marks no suggestion as the current choice
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

  # Round 093 carrier: the typed two-line value must be kept on two editor rows (the joined-row defect
  # `line oneline two` reddens this Example's Then — see the DSL row for the separate-rows clause).
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
