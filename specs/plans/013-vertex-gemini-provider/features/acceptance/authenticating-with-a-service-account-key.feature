Feature: Authenticating to a Gemini model with a service-account key

  # Acceptance only: a Gemini model is reached with the operator's configured
  # service-account key file (an "API_KEY" that points at a ".json" credential).
  # A missing or unusable key must fail clearly — never be silently ignored, and
  # never quietly fall back to another credential path.

  Rule: The configured service-account key file is what reaches the model

    Example: The operator authenticates with a key file and gets an answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a service-account key file "secrets/key.json"
      And the runtime home holds a configuration with a reachable provider "vertex-flash-3.8" that uses a Gemini model and authenticates with the key file "secrets/key.json"
      When the operator asks tellme "What is two plus two?"
      Then tellme authenticates to the provider "vertex-flash-3.8" with the configured key file
      And tellme prints the provider's answer
      And tellme exits successfully

  Rule: A missing or unusable key file fails clearly, with no silent fallback

    Example: The configured key file is missing
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a provider "vertex-flash-3.8" that uses a Gemini model whose key file "secrets/key.json" is missing
      When the operator asks tellme "What is two plus two?"
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme exits with the provider error code
