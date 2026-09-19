Feature: Keeping more of an action's argument

  # Acceptance only (round 057). When the assistant calls a tool, tellme shows the
  # call's arguments; a long value is shortened so the line stays readable. The
  # operator wants more of a long value kept than before.

  Rule: A long argument value is kept up to five hundred characters

    Example: A value of exactly five hundred characters is shown whole
      Given the operator has a runnable tellme installation
      And the assistant calls a tool with an argument value of five hundred characters
      When the operator asks tellme to do it
      Then the action line shows the whole value with nothing left out
      And tellme exits successfully

    Example: A value of five hundred and one characters is shortened to five hundred
      Given the operator has a runnable tellme installation
      And the assistant calls a tool with an argument value of five hundred and one characters
      When the operator asks tellme to do it
      Then the action line shows the value shortened to five hundred characters ending in a single ellipsis
      And tellme exits successfully

  Rule: A value with accented or wide characters is still cut cleanly

    Example: A value of mixed characters is shortened without breaking a character
      Given the operator has a runnable tellme installation
      And the assistant calls a tool with an argument value of mixed characters longer than five hundred
      When the operator asks tellme to do it
      Then the action line shows the value shortened to five hundred readable characters ending in a single ellipsis
      And tellme exits successfully
