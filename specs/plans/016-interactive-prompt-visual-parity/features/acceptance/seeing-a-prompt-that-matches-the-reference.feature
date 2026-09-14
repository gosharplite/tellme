Feature: The interactive prompt matches tell-me-go

  # Acceptance only: the `-i` interactive prompt presents the same surface as
  # tell-me-go — a framed multi-line editor above a list of suggestions, and no
  # session metrics header.

  Rule: The interactive prompt presents a framed editor above a list of suggestions

    Example: The operator opens the interactive prompt
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      Then the prompt is framed around a multi-line editor
      And the prompt lists suggestions beneath the editor
      And exactly one suggestion is marked as the current choice

  Rule: The interactive prompt shows no session metrics header

    Example: The operator opens the interactive prompt
      Given the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      Then the prompt shows no provider, token-usage, or turn-count header

  Rule: The editor keeps a multi-line prompt without line numbers

    Example: The operator composes a prompt across several lines
      Given the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      And the operator types a prompt across several lines
      Then the prompt keeps every line the operator typed
      And the prompt shows no line numbers
