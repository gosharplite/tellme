Feature: Composing with settled suggestions

  # Acceptance only: as the operator types, the prompt settles its suggestions
  # after a short pause, accepts a suggestion into the editor, and never offers
  # an over-long suggestion.

  Rule: Suggestions settle after the operator pauses typing

    Example: The operator types a prompt and the suggestions settle
      Given the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      And the operator types a prompt
      Then the prompt offers suggestions for the typed text after a short pause

  Rule: Accepting a suggestion inserts it into the editor

    Example: The operator accepts a suggestion to complete a line
      Given the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      And the operator types "deploy "
      And the operator accepts the current suggestion "deploy to staging with version 016"
      Then the editor holds "deploy to staging with version 016"

  Rule: An over-long suggestion is never offered

    Example: The operator's recent prompts include one that spans many lines
      Given the operator is working at an interactive terminal
      And the operator's recent prompts include one that spans many lines
      When the operator opens the interactive prompt
      Then no suggestion offered to the operator spans more than three lines
