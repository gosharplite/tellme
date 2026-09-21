Feature: Recovering from a momentary provider failure

  # Acceptance journey (plan-side) — round 078, operator request. A prompt turn can
  # fail because the connection to the provider momentarily dropped; re-running the
  # command usually works. tellme should absorb that blip by itself — wait and try
  # again a bounded number of times — and only then report the very same failure it
  # reports today. A failure the provider would never recover from is reported at once,
  # so a real error is never hidden behind pointless waiting.

  Rule: A momentary provider failure is retried before the run fails
    When a request fails only because the connection to the provider dropped, tellme
    waits a short while and sends the same request again — a small, fixed number of
    times — so a network blip does not cost the operator a re-run.

    Example: A single blip is absorbed and the turn finishes
      Given a configured provider whose endpoint drops the connection once and then answers
      When the operator starts tellme with the prompt "Hello"
      Then the turn finishes with the provider's answer
      And tellme sent the request exactly twice

    Example: A second blip is still absorbed
      Given a configured provider whose endpoint drops the connection twice and then answers
      When the operator starts tellme with the prompt "Hello"
      Then the turn finishes with the provider's answer
      And tellme sent the request exactly three times

  Rule: Retrying is bounded and ends in the failure the operator already knows
    When the provider keeps dropping, tellme retries a fixed, small number of times and
    then reports the failure exactly as it does today — the same message, the same exit.

    Example: A provider that never answers is failed after the bounded retries
      Given a configured provider whose endpoint always drops the connection
      When the operator starts tellme with the prompt "Hello"
      Then tellme sent the request exactly three times
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

  Rule: A failure the provider would never recover from is reported at once
    When the provider rejects the request outright rather than dropping it, tellme must
    not waste the operator's time retrying a request that cannot succeed.

    Example: A rejected request is not retried
      Given a configured provider whose endpoint rejects the request outright
      When the operator starts tellme with the prompt "Hello"
      Then tellme sent the request exactly once
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code
