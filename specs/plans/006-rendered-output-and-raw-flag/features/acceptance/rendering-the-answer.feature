Feature: Rendering the provider's answer

  # Acceptance only: what the operator observes on the output side once tellme formats the answer.
  # By default the answer is displayed as formatted text; the operator can ask for the raw answer with "-r".
  # tellme never adds terminal decoration of its own when the answer is redirected. Reference parity (Clarify Round 1, Q1).

  Rule: The provider's answer is displayed as formatted text by default

    Example: The operator asks a question whose answer uses Markdown
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "**Bold** and _italic_."
      When the operator asks tellme "Format this"
      Then the displayed answer shows the words "Bold" and "italic" without the raw Markdown markers "**" and "_"
      And tellme exits successfully

    Example: The answer is still displayed as formatted text when it is redirected
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "**Bold**."
      When the operator asks tellme "Format this" and captures its standard output
      Then the captured output shows the word "Bold" without the raw Markdown marker "**"
      And tellme exits successfully

  Rule: tellme adds none of its own terminal decoration when the answer is redirected

    Example: The operator redirects the answer to a file and reads it back
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "Plain answer."
      When the operator asks tellme "Say hi" and captures its standard output
      Then the captured output carries no terminal decoration added by tellme
      And the captured output shows the provider's answer "Plain answer."
      And tellme exits successfully

  Rule: The operator can ask for the raw answer with "-r"

    Example: The operator requests raw output and captures it
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "**Bold** and _italic_."
      When the operator asks tellme "Format this" with "-r" and captures its standard output
      Then the captured output contains the provider's answer "**Bold** and _italic_." as plain text
      And the captured output carries no formatting added by tellme
      And tellme exits successfully

    Example: The long form of the raw flag behaves the same
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "**Bold**."
      When the operator asks tellme "Format this" with "--raw" and captures its standard output
      Then the captured output contains the provider's answer "**Bold**." as plain text
      And tellme exits successfully
