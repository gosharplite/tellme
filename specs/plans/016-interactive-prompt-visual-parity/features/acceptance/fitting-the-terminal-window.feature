Feature: The interactive prompt fits the terminal

  # Acceptance only: the prompt uses the width of the terminal, reflows when the
  # operator resizes it, and degrades without breaking on a very narrow terminal.
  #
  # [unit-pinned (round 016, PR #40 F1)] resize/reflow and narrow-terminal
  # degrade are pinned by the unit layer (no pty, no `WindowSizeMsg` source); the
  # observable *static* frame width is carried by
  # specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature
  # ("draws the editor within a frame bounded by the terminal width").

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
