Feature: Labelling the progress indicator

  # Acceptance only: the indicator names the step the run is waiting on — the
  # model it is waiting for, or the tools it is running — plus the machine's
  # resource usage while tools run.

  Rule: The indicator names the model while the run waits for it

    Example: The operator waits for an answer from a named model
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      When the operator runs tellme with the prompt "hi"
      Then the progress indicator names the model it is waiting for
      And the progress indicator shows how long the run has waited

  Rule: The indicator names the tools and the resource usage while tools run

    Example: The turn runs a single tool
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      And the model asks for the tool "read the notes" before answering
      When the operator runs tellme with the prompt "read the notes"
      Then the progress indicator names the tool it is running
      And the progress indicator reports the machine's resource usage

    Example: The turn runs several tools
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      And the model asks for the tools "read the notes" and "list the folder" before answering
      When the operator runs tellme with the prompt "read the notes and list the folder"
      Then the progress indicator names every tool it is running
      And the progress indicator reports the machine's resource usage
