Feature: Driving a Vertex Gemini model

  # Interface truth (CLI end, `chat` module) — atomic rules for driving a Gemini model (the Vertex AI
  # family) through one reasoning turn. The acceptance journeys live in the plan package
  # (`features/acceptance/answering-with-a-gemini-model.feature`). A `TYPE: "gemini"` provider is
  # routed to the Vertex `:generateContent` transport instead of being refused; its answer, its tool
  # round-trip, its persona, and its output budget behave exactly like the OpenAI-compatible family.

  Rule: A prompt driven through a Vertex Gemini provider prints the model's answer

    Example: The operator asks a Gemini model and reads its answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose endpoint answers with "4"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then tellme sends exactly one request to the provider "vertex-flash-3.8"
      And tellme prints the provider's answer "4"
      And tellme exits successfully

  Rule: The configured persona is carried into a Vertex Gemini request

    Example: A configured persona precedes the prompt on a Gemini request
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose endpoint answers with "4"
      And the runtime home holds a configuration whose persona is "You are a terse assistant. Answer in one sentence."
      When the operator starts tellme with the prompt "What is two plus two?"
      Then the request carried the persona "You are a terse assistant. Answer in one sentence."
      And tellme prints the provider's answer "4"

  Rule: The configured output budget is carried into a Vertex Gemini request

    Example: A bounded output budget is requested from the Gemini model
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose endpoint answers with "4"
      And the configured Gemini provider entry allows at most 40960 output tokens
      When the operator starts tellme with the prompt "What is two plus two?"
      Then the request to the provider "vertex-flash-3.8" allows at most 40960 output tokens

  Rule: A Vertex Gemini model can ask tellme to run a tool before answering

    Example: The model reads a file with tellme's help, then answers
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose endpoint asks tellme to read "notes.txt" and then answers with "the launch code is ORANGE"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator starts tellme with the prompt "What is the launch code? Read notes.txt to find out."
      Then tellme read "notes.txt" using its read_files tool
      And tellme prints the provider's answer "the launch code is ORANGE"
