Feature: Reporting a failed provider request

  # Interface truth (CLI end, `chat` module) — the provider/transport failure surface. A failed
  # request must name the frozen class phrase `the provider request failed` and exit with the pinned
  # provider error code `6`. (A reply **cut off at the provider's output limit** is a further cause of
  # the same class — see `refusing-a-cut-off-reply.feature`, round 030.)

  Rule: An unreachable provider or an error response is reported with the frozen class phrase

    Example: The provider cannot be reached
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "dead-model" whose endpoint is unreachable
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

    Example: The provider answers with an error status
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "error-model" whose endpoint answers with an error status
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

  Rule: A response that cannot be understood as an answer fails the same way

    Example: The provider returns no usable answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "garbled-model" whose endpoint answers with no usable answer
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

  Rule: A failed turn still reports the failure and leaves no progress indicator

    # Proves the absence of any indicator residue: the spinner draws one frame and then synchronously clears
    # it (AgentLoop calls the observer before the request), so the terminal-visible stream shows none.
    # The affirmative mid-wait clear-before-the-class-phrase proof is a recorded forward item.
    Example: The provider cannot be reached while the operator watches the terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "dead-model" whose endpoint is unreachable
      When the operator starts tellme with the prompt "Hello"
      Then the run shows no progress spinner
      And tellme explains on stderr that "the provider request failed"

  Rule: A transient provider failure is retried before it is reported (round 078)
    # ADR 0050: a RETRYABLE failure — a transport drop or an HTTP 429/5xx — is
    # retried at most twice (fixed delays 1 s, 3 s) before the frozen phrase + exit 6
    # fire. A non-retryable failure (4xx / decode / the round-030 truncation) is
    # reported at once, with no retry. The final failure surface is unchanged.

    Example: A momentary connection drop is absorbed
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "flaky-model" whose endpoint drops the connection once and then answers
      When the operator starts tellme with the prompt "Hello"
      Then the turn finishes with the provider's answer
      And tellme sent the request exactly twice
      And tellme announces on stderr that it is retrying the provider request
      And the retry re-sent the same request
      And the turn was recorded as a single provider call

    Example: A second drop is still absorbed
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "flaky-model" whose endpoint drops the connection twice and then answers
      When the operator starts tellme with the prompt "Hello"
      Then the turn finishes with the provider's answer
      And tellme sent the request exactly three times

    Example: A provider that never answers is failed after the bounded retries
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "dead-model" whose endpoint always drops the connection
      When the operator starts tellme with the prompt "Hello"
      Then tellme sent the request exactly three times
      And the run wrote nothing to standard output
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

    Example: A rejected request is not retried
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "strict-model" whose endpoint rejects the request outright
      When the operator starts tellme with the prompt "Hello"
      Then tellme sent the request exactly once
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code
