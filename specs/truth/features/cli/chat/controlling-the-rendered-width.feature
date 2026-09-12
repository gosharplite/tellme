Feature: Controlling the rendered width

  # Interface truth (CLI end, `chat` module) — the rendered width (config `WRAP_WIDTH` /
  # `TELL_ME_WRAP_WIDTH`) wraps rendered output only (round-006 Clarify Q3, reference parity).

  Rule: Rendered output wraps at the configured width

    Example: The operator sets a narrow rendered width
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
      And the rendered width is "40"
      When the operator starts tellme with the prompt "Tell me something long"
      Then the captured standard output is wrapped so that no line is wider than 40 columns
      And tellme exits successfully

  Rule: The rendered width applies only to rendered output

    Example: Raw output is not wrapped
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
      And the rendered width is "40"
      When the operator starts tellme with the prompt "Tell me something long" and the raw flag
      Then the captured standard output contains the answer on a single line
      And tellme exits successfully
