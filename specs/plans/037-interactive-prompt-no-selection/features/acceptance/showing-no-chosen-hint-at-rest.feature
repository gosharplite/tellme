Feature: Showing no chosen hint at rest

  # PM acceptance (round 037). When the interactive prompt opens it shows the candidate
  # hints, but NONE of them is "chosen" until the operator navigates. This SUPERSEDES the
  # round-016 assumption (frozen in specs/plans/016-interactive-prompt-visual-parity/**)
  # that the first suggestion is pre-selected at rest: the reference (tell-me-go) opens
  # with no selection, and this round aligns tellme to it. Business language only — no
  # terminal/ANSI/implementation detail.

  Rule: The prompt pre-selects no suggestion

    Example: The prompt opens with every hint unselected
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging with version 037"
      When the operator opens the interactive prompt
      Then the interactive prompt marks no suggestion as the current choice
      And tellme exits successfully

  # The companion behavior — the first navigation chooses the FIRST suggestion and an
  # accepted choice is inserted into the editor — is already carried by the existing
  # accept journey in specs/truth/features/cli/chat/prompting-with-suggestions.feature
  # ("The operator accepts the current suggestion into the editor"), so no second rule is
  # minted here (keeps acceptance-coverage exact).
