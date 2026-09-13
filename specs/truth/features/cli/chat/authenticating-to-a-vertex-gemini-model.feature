Feature: Authenticating to a Vertex Gemini model

  # Interface truth (CLI end, `chat` module) — atomic rules for the service-account credential of a
  # Vertex Gemini provider. The acceptance journeys live in the plan package
  # (`features/acceptance/authenticating-with-a-service-account-key.feature`). The configured
  # service-account key file authenticates the request; a missing/unusable key fails with the frozen
  # `the provider request failed` phrase and no silent fallback.

  Rule: The configured service-account key authenticates a Vertex Gemini request

    Example: The operator authenticates with a key file and gets an answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose endpoint answers with "4"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then the request to the provider "vertex-flash-3.8" carried the service-account access token
      And tellme prints the provider's answer "4"
      And tellme exits successfully

  Rule: A missing service-account key file fails with the frozen provider phrase

    Example: The configured key file is missing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose key file is missing
      When the operator starts tellme with the prompt "What is two plus two?"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code
