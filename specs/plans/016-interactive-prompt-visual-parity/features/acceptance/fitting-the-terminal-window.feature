Feature: The interactive prompt fits the terminal

  # Acceptance only: the prompt uses the width of the terminal, reflows when the
  # operator resizes it, and degrades without breaking on a very narrow terminal.

  Rule: The prompt uses the width of the terminal and reflows on resize

    Example: The operator resizes the terminal while composing
      Given the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      And the operator widens the terminal
      Then the prompt fills the new width
      And the editor and the suggestion list stay aligned

    Example: The terminal is narrower than the prompt needs
      Given the operator is working at a very narrow interactive terminal
      When the operator opens the interactive prompt
      Then the prompt is shown without breaking the terminal
