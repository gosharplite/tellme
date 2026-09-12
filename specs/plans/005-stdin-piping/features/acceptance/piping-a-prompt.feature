Feature: Piping a prompt into tellme

  # Acceptance only: what the operator observes when the prompt is fed through standard input.
  # This round makes tellme composable in a pipeline — e.g. `cat file | tellme "summary"` — while keeping
  # the existing offline behavior when nothing is piped. The provider request shape is unchanged from round 004.
  # Combining an instruction with piped content follows tell-me-go's main path (Clarify Round 1, Q2).

  Rule: A prompt piped on standard input must be sent to the selected provider and the answer printed

    Example: The operator pipes a question with no instruction
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "Use go list ./... to list packages."
      When the operator pipes "How do I list Go packages?" into tellme
      Then tellme sends exactly one request carrying the piped question to the provider "test-model"
      And tellme prints the provider's answer "Use go list ./... to list packages."
      And tellme exits successfully

    Example: The operator pipes content together with an instruction
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "The file defines the program entry point."
      When the operator pipes "package main" into tellme with the instruction "Summarize this"
      Then tellme sends exactly one request carrying the instruction followed by the piped content to the provider "test-model"
      And tellme prints the provider's answer "The file defines the program entry point."
      And tellme exits successfully

  Rule: A pipe without prompt content must keep the existing offline behavior

    Example: The operator pipes nothing and gives no instruction
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      When the operator pipes nothing into tellme with no instruction
      Then tellme reports that the configuration is ready
      And tellme performs no network access
      And tellme exits successfully

    Example: An explicit diagnostic flag takes precedence over piped input
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      When the operator runs tellme's diagnostic "-d" with "Hello" piped in
      Then tellme reports that the configuration resolved
      And tellme performs no network access
      And tellme exits successfully
