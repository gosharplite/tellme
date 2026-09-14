Feature: The post-turn lines stay bounded

  # Acceptance only: the metrics and the summary appear only when there is usage
  # to report, and only outside the answer; the non-prompt commands keep their
  # existing output.

  Rule: The lines appear only when there is usage to report

    Example: The provider reports no usage
      Given the operator has a runnable tellme installation
      And the provider reports no usage
      When the operator runs tellme with the prompt "hi"
      Then the run reports the answer without the post-turn lines

  Rule: The lines stay outside the answer

    Example: The operator pipes the prompt in and reads the answer
      Given the operator has a runnable tellme installation
      When the operator pipes "summarise the notes" into tellme
      Then the answer is unchanged by the post-turn lines
      And the post-turn lines appear outside the answer

  Rule: The non-prompt commands keep their existing output

    Example: The operator asks for the version
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the version flag
      Then the run prints the version without reporting post-turn lines

    Example: The operator lists the recent conversation
      Given the operator has a runnable tellme installation
      When the operator lists the recent conversation
      Then the run prints the conversation without reporting post-turn lines
