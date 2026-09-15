Feature: Starting the count over on a fresh session

  # Acceptance only: the running number belongs to the current conversation, so
  # starting a fresh session begins it again at the first turn.

  Rule: Starting a fresh conversation begins the number again

    Example: A fresh session after a conversation that consulted a tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What does notes.txt say?"
      Then the turn opens at "Turn 1" for the active mode
      When the operator starts a fresh session with "--new" and the prompt "hi"
      Then the turn opens at "Turn 1" for the active mode
      And tellme exits successfully

    Example: A fresh session requested from the interactive prompt
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What does notes.txt say?"
      Then the turn opens at "Turn 1" for the active mode
      When the operator starts a fresh session at the interactive prompt with the prompt "hi"
      Then the turn opens at "Turn 1" for the active mode
      And tellme exits successfully
