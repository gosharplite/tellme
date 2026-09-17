Feature: Keeping the prompt open on an empty submit

  # Acceptance journey (business language). At the interactive prompt, pressing the submit keys
  # (`Ctrl+S` / `Alt+Enter`) on an empty box must do nothing — the operator stays in the prompt and
  # can keep composing. (The reference behaves this way; an accidental empty submit must not end the
  # session.) A non-empty submit is unchanged.

  Rule: An empty submit does not end the prompt

    Example: An accidental empty submit is ignored and the next submit runs
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator submits an empty prompt then submits the prompt "hi" at the interactive prompt
      Then the interactive prompt is shown
      And tellme sends exactly one request to the provider "test-model"
      And tellme prints the provider's answer "ok"
      And tellme exits successfully
