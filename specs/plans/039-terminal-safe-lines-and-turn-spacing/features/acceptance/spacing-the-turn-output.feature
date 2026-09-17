Feature: Spacing the turn output

  # Acceptance journey (business language). A tool-using turn's output is a dense wall of lines.
  # The operator wants it grouped: each tool call's report starts a fresh, blank-separated block;
  # the trailing summary of the round's reasons is set off by a blank line; and the turn's closing
  # status (the measured payload line and the session summary) is set off by a blank line too.

  Rule: Each tool call's report starts a blank-separated block

    Example: A round of two calls separates each call's report
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint shows the folder tree and then reads "notes.txt" and then answers with "the launch code is ORANGE"
      When the operator starts tellme with the prompt "Survey the project, then read notes.txt."
      Then each tool call's report begins after a blank line
      And tellme exits successfully

    Example: A call that states no reason still starts a fresh block
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" without stating a reason and then answers with "done"
      When the operator starts tellme with the prompt "Read notes.txt."
      Then the action of the call without a reason begins after a blank line
      And tellme exits successfully

  Rule: The turn's trailing summary and closing status are set off by a blank line

    Example: The round's reasons and the closing status each follow a blank line
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the trailing reason summary follows a blank line
      And the turn's closing status follows a blank line
      And tellme exits successfully
