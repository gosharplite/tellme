Feature: Reporting the machine's resource usage

  # Acceptance only (round 059). The progress spinner's tool-phase resource
  # segment must show REAL machine figures (it read 0.0% on every macOS host
  # through round 058). The once-per-second refresh is a documented narrowing —
  # see the note under the second Rule. Business language only.

  Rule: A tool-using turn reports the machine's real resource usage

    # The non-zero assertion lives in the interface carrier: the `chat` row
    # `the progress spinner reports the machine's resource usage` now requires a
    # non-zero memory figure (round 059). These Examples are built from EXISTING
    # sentences so they are not vacuous (the F-058-1 / R-058-c class).

    Example: The operator watches a tool run at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a command and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the progress spinner reports the machine's resource usage
      And tellme exits successfully

    Example: The operator watches a command stream at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a command producing a great deal of output and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the progress spinner reports the machine's resource usage
      And tellme exits successfully

  Rule: The resource figures refresh at most once per second

    # Documented narrowing (R-059-c): the once-per-second refresh cadence cannot
    # be observed from a flat capture (the frames redraw every 200 ms), so its
    # carrier is the unit pin `TestSpinnerResourceSampleThrottle` (internal/ui) —
    # not an Example. Recorded as a PM follow-up so a future PM pass may mirror
    # it into the acceptance journey if a runner-owned form appears.
