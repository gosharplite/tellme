Feature: Capturing a multi-line prompt at the terminal

  # Acceptance only: at an interactive terminal with no prompt argument, tellme
  # reads the operator's prompt across multiple lines and sends it on Ctrl+D. The
  # hint is a diagnostic; the answer stays on standard output.

  Rule: At a terminal with no prompt, tellme reads a multi-line prompt

    Example: The operator composes a two-line prompt and sends it
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator starts tellme with no prompt, types a two-line prompt, then sends it
      Then tellme announces that it is reading multi-line input
      And tellme asks the model the typed two-line prompt
      And the final answer is printed on standard output

  Rule: The reading announcement is a diagnostic and never disturbs the answer

    Example: The announcement is kept off the answer stream
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator starts tellme with no prompt, types a prompt, then sends it
      Then the reading announcement is reported on the diagnostic output, not standard output
      And the final answer is printed on standard output

  Rule: A piped prompt never enters the interactive reader

    Example: A piped prompt is handled as before
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator pipes "What is two plus two?" into tellme
      Then no reading announcement is reported
      And the final answer is printed on standard output
