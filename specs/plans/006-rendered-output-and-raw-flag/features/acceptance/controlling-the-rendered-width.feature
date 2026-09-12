Feature: Controlling the rendered width

  # Acceptance only: the operator can choose the column width at which rendered output wraps,
  # through configuration or the environment. Reference parity (Clarify Round 1, Q3).

  Rule: Rendered output wraps at the configured width

    Example: The operator sets a narrow wrap width through the environment
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
      When the operator sets the wrap width to 40 columns and asks tellme "Tell me something long"
      Then the displayed answer is wrapped so that no line is wider than about 40 columns
      And tellme exits successfully

    Example: With no wrap width configured, the default width is used
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
      When the operator asks tellme "Tell me something long" and no wrap width is configured
      Then the displayed answer is wrapped at tellme's default width
      And tellme exits successfully

  Rule: The wrap width applies only to rendered output

    Example: Raw output is left unwrapped
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
      When the operator sets the wrap width to 40 columns and asks tellme "Tell me something long" with "-r" and captures its standard output
      Then the captured output contains the provider's answer on a single line, unwrapped
      And tellme exits successfully

  Rule: An invalid wrap width is refused

    Example: The operator configures a negative wrap width
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" and the wrap width set to a negative value
      When the operator asks tellme "Hello"
      Then tellme reports that the configuration is invalid
      And tellme exits with the configuration-error code
