Feature: Composing a multi-line prompt

  # Plan-side acceptance journey — round 093, anchor issue #191.
  #
  # The operator composes a prompt across several lines in the `-i` interactive
  # prompt. The prompt must keep the typed lines apart — each on its own editor
  # row — and the run must end cleanly. Previously the two lines could collapse
  # onto one row (`line oneline two`), a defect the existing assertion could not
  # see. This is the PM-readable acceptance journey; the executable interface
  # truth lives under specs/truth/features/cli/chat/**.

  Rule: The editor keeps each typed line on its own row

    Example: The operator composes a prompt across several lines
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator opens the interactive prompt and types "line one\nline two"
      Then the interactive prompt keeps each typed line on its own row
      And tellme exits successfully
