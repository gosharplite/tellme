Feature: Choosing between the interactive prompt and the plain input

  # Acceptance only: the rich interactive prompt is opt-in. The plain reader
  # stays the default at a terminal, and any piped / non-terminal input keeps
  # its existing behaviour — upgrading tellme never changes existing workflows.

  Rule: The interactive prompt is opt-in and the plain reader stays the default

    Example: A bare terminal invocation keeps the plain reader
      Given the operator has a runnable tellme installation
      When the operator starts tellme at a terminal with no prompt
      Then tellme announces the plain multi-line reader
      And the interactive prompt is not shown

    Example: Enabling the interactive prompt opens the rich prompt instead
      Given the operator has a runnable tellme installation
      When the operator starts tellme at a terminal with the interactive prompt enabled
      Then the interactive prompt is shown
      And the plain multi-line reader is not announced

  Rule: A non-terminal input never opens the interactive prompt

    Example: A piped invocation keeps the piped behaviour
      Given the operator has a runnable tellme installation
      When the operator pipes "summarize the changelog" into tellme
      Then tellme uses the piped prompt
      And the interactive prompt is not shown
      And tellme prints the model's answer

    Example: Enabling the interactive prompt with piped input falls back safely
      Given the operator has a runnable tellme installation
      When the operator pipes "summarize the changelog" into tellme with the interactive prompt enabled
      Then the interactive prompt is not shown
      And tellme uses the piped prompt
      And tellme prints the model's answer
