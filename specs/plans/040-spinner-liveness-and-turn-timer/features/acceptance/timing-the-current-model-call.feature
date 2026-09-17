Feature: Timing the current model call

  # Acceptance journey (business language). The progress indicator shows one figure today: the whole
  # seconds since the prompt was captured. The operator wants a second figure alongside it — how long
  # the current model call has been running — so a slow single call can be told apart from a slow series of
  # calls. The first figure keeps counting from the prompt; the second restarts for each model call.

  Rule: The indicator shows the total wait and the current model call's time

    Example: A single-call answer shows both times
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi"
      Then the progress indicator shows the time since the prompt and the current model call's time
      And tellme exits successfully

    Example: A tool-using answer restarts the model-call time for each call
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint reads "notes.txt" and then answers with "all good"
      When the operator starts tellme with the prompt "read the notes"
      Then the progress indicator shows the time since the prompt and the current model call's time
      And the current model call's time restarts when a new model call begins
      And tellme exits successfully

  Rule: Both times are shown while tools run, alongside the resource usage

    Example: The tool phase shows both times and the machine's resource usage
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint reads "notes.txt" and then answers with "all good"
      When the operator starts tellme with the prompt "read the notes"
      Then the progress indicator shows the time since the prompt and the current model call's time
      And the progress indicator reports the machine's resource usage
      And tellme exits successfully
