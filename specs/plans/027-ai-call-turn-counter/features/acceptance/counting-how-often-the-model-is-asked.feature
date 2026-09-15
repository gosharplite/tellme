Feature: Counting how often the model is asked

  # Acceptance only: the number in a turn's opening header counts how many times
  # the model has actually been asked, not how many prompts the operator typed.
  # A prompt the model answers in one request advances the number by one; a
  # prompt that first consults a tool before answering advances it by two (or
  # more). The number reads the same as tell-me-go's counter.

  Rule: The number counts every request the model has been asked so far

    Example: The very first prompt of a new conversation
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers without consulting a tool
      When the operator asks tellme "What is two plus two?"
      Then the turn opens at "Turn 1" for the active mode
      And the operator sees the answer
      And tellme exits successfully

    Example: A plain prompt right after another plain prompt
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers without consulting a tool
      When the operator asks tellme "hi"
      Then the turn opens at "Turn 1" for the active mode
      When the operator asks tellme "hello again"
      Then the turn opens at "Turn 2" for the active mode
      And tellme exits successfully

    Example: The prompt that follows a turn which consulted a tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What does notes.txt say?"
      Then the turn opens at "Turn 1" for the active mode
      When the operator asks tellme "And now?"
      Then the turn opens at "Turn 3" for the active mode
      And tellme exits successfully

  Rule: A turn opens a single time, however many times the model is asked

    Example: A prompt that consults a tool still opens the turn once
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What does notes.txt say?"
      Then the turn opens exactly once for the run
      And the operator sees the answer
      And tellme exits successfully
