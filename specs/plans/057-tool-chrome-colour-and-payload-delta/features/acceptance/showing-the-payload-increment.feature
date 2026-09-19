Feature: Showing how the payload grew

  # Acceptance only (round 057). Before each model call tellme shows how big the
  # request is; the operator wants that pre-flight line to show how much it grew
  # over the previous one, instead of repeating a constant allowance it can never
  # reach.

  Rule: The pre-flight payload line shows the increase over the previous one

    Example: A larger payload shows how much it grew
      Given the operator has a runnable tellme installation
      And the previous request was about 203048 tokens and the next is about 203148
      When the operator asks tellme a question
      Then the pre-flight payload line shows an increase of 100 before the request size
      And the pre-flight payload line no longer shows the allowance
      And tellme exits successfully

  Rule: A first payload shows no increase

    Example: The very first payload of a session shows no growth
      Given the operator has a runnable tellme installation
      And no earlier payload has been shown this session
      When the operator asks tellme a question
      Then the pre-flight payload line shows an increase of zero before the request size
      And tellme exits successfully

  Rule: A payload that shrank shows a negative increase

    Example: A smaller payload shows a negative increase
      Given the operator has a runnable tellme installation
      And the previous request was about 204000 tokens and the next is about 203000
      When the operator asks tellme a question
      Then the pre-flight payload line shows a decrease of 1000 before the request size
      And tellme exits successfully

  Rule: The measured payload line is unchanged

    Example: The measured line still shows the request size and the allowance
      Given the operator has a runnable tellme installation
      And the assistant answers a question
      When the operator asks tellme a question
      Then the measured payload line still shows the request size over the allowance
      And tellme exits successfully

  Rule: The saved copy of the turn follows the same content without colour

    Example: The saved turn log shows the same payload content
      Given the operator has a runnable tellme installation
      And the previous request was about 203048 tokens and the next is about 203148
      When the operator asks tellme a question
      Then the saved turn log shows the same pre-flight payload content without any colour
      And tellme exits successfully
