Feature: Piping a prompt into tellme

  # Interface truth (CLI end, `chat` module) — atomic rules for prompt input through standard input.
  # The acceptance journeys live in the plan package (`features/acceptance/**`).
  # Combining an instruction with piped content follows the reference's main chat path
  # (round-005 Clarify Q2): the positional argument(s), a newline, then the piped content.

  Rule: A prompt piped on standard input is sent to the selected provider and the answer printed

    Example: The operator pipes a question with no instruction
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "List packages with go list"
      When the operator pipes "How do I list Go packages?" into tellme
      Then tellme sends exactly one request to the provider "test-model"
      And the request carried the piped content "How do I list Go packages?"
      And tellme prints the provider's answer "List packages with go list"
      And tellme exits successfully

    Example: The operator pipes content together with an instruction
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "The file defines the program entry point"
      When the operator pipes "package main" into tellme with the instruction "Summarize this"
      Then tellme sends exactly one request to the provider "test-model"
      And the request carried the instruction "Summarize this" followed by the piped content "package main"
      And tellme prints the provider's answer "The file defines the program entry point"
      And tellme exits successfully

  Rule: A pipe without prompt content makes no provider request

    Example: The operator pipes nothing and gives no instruction
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "unused"
      When the operator pipes nothing into tellme
      Then tellme performs no network access
      And tellme exits successfully

    Example: An explicit diagnostic flag takes precedence over piped input
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "unused"
      When the operator runs tellme's diagnostic with "Hello" piped in
      Then tellme performs no network access
      And tellme exits successfully
