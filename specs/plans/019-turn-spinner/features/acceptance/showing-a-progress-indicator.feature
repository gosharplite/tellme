Feature: Showing a progress indicator while the run works

  # Acceptance only: on a terminal, a prompt-bearing run shows a live progress
  # indicator from the moment the input is captured until the operator regains
  # control, so a slow model or tool never looks like a hang.

  Rule: A prompt-bearing run shows a live indicator while it works

    Example: The operator runs a prompt at a terminal
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      When the operator runs tellme with the prompt "hi"
      Then the run shows a progress indicator while it works
      And the progress indicator keeps changing while the run waits
      And the answer is unchanged by the progress indicator

    Example: The operator's turn takes a while to answer
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      And the model takes a while to answer
      When the operator runs tellme with the prompt "hi"
      Then the run shows a progress indicator for as long as it waits

  Rule: The progress indicator makes room for the conversation

    Example: The operator's turn runs a tool before answering
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      And the model asks for a tool before answering
      When the operator runs tellme with the prompt "read the notes"
      Then the answer appears without the progress indicator covering it
      And the progress indicator is gone once the run finishes
