Feature: Reporting the machine's resource usage

  # Acceptance only (round 059). The progress spinner's tool-phase resource
  # segment must show REAL machine figures (it read 0.0% on every macOS host
  # through round 058) and the figures must refresh at most once per second so
  # they stay readable. Business language only.

  Rule: A tool-using turn reports the machine's real resource usage

    Example: The operator watches a tool run at a terminal on macOS
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a command and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the progress spinner reports the machine's resource usage
      And the reported memory usage is not zero
      And tellme exits successfully

  Rule: The resource figures refresh at most once per second

    # The refresh cadence cannot be observed from a flat capture (the frames
    # redraw every 200 ms regardless), so this rule's carrier is a unit pin
    # (`TestSpinnerResourceSampleThrottle`) — a recorded narrowing (ADR 0029).

    Example: The operator watches a long tool run
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a slow command and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the progress spinner reports the machine's resource usage
      And the resource figures are refreshed no more than once per second
      And tellme exits successfully
