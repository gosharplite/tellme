Feature: Piping the answer out

  # Acceptance only: what the operator observes on the output side when tellme is used in a pipeline.
  # A piped run must produce the plain answer alone (no terminal decoration) and must never wait for an
  # interactive terminal. Adopting tell-me-go's TTY-aware posture (Clarify Round 1, Q3).

  Rule: A redirected answer must carry only the answer text

    Example: The operator redirects the answer to a file and reads it back
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "Use go list ./... to list packages."
      When the operator runs tellme with the prompt "How do I list Go packages?" and captures its standard output
      Then the captured output contains exactly the provider's answer "Use go list ./... to list packages."
      And the captured output carries no terminal decoration
      And tellme exits successfully

  Rule: A piped run must not wait for interactive terminal input

    Example: The operator pipes a prompt in a non-interactive environment
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "Hello back."
      When the operator pipes "Hello" into tellme
      Then tellme completes the turn without waiting for terminal input
      And tellme prints the provider's answer "Hello back."
      And tellme exits successfully
