Feature: Watching the tool loop work

  # Interface truth (CLI end, `chat` module) — the tool loop reports its activity on the diagnostic
  # output in the reference's decomposed shape (round 034). Acceptance journeys:
  # features/acceptance/showing-the-tool-calls.feature, framing-the-turn-per-call.feature,
  # streaming-the-command-output.feature.

  Rule: A tool-using run reports each call's step, reason, action, and result

    Example: The run reports the step marker, the reason, the action, and the result
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the tool step marker
      And the run reported the reason "checking the launch code" for the tool call "read_files"
      And the run reported the action for the tool call "read_files" without the reason
      And the run reported the result for the tool call "read_files"
      And tellme exits successfully

  Rule: A tool-using run reports the round's reasons again at the call's end

    Example: The reasons reappear grouped before the measured payload status
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the reasons again before the measured payload status
      And tellme exits successfully

  Rule: A tool-using run reports every call in the order it was made

    Example: Two tools used in sequence are reported in order
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint shows the folder tree and then reads "notes.txt" and then answers with "The launch code is ORANGE."
      When the operator starts tellme with the prompt "Survey the project, then read notes.txt."
      Then the run reported the action for the tool call "get_tree" before the action for the tool call "read_files"
      And tellme exits successfully

  Rule: A tool-using run streams a command's output live

    Example: A shell command's output is streamed as it runs
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command and then answers with "done"
      When the operator starts tellme with the prompt "Run the command."
      Then the run streamed the command's output on its diagnostic output
      And tellme exits successfully

    Example: A command that writes to a file streams no output block
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that writes its output to a file and then reads it and then answers with "done"
      When the operator starts tellme with the prompt "Run the command."
      Then the run streamed no command output block for the writing command
      And tellme exits successfully
