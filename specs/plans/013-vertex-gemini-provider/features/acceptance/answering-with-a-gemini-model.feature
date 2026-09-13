Feature: Answering a prompt with a Gemini model

  # Acceptance only: what the operator observes when tellme drives a Gemini
  # model (the Vertex family). Today a "gemini" provider is refused, so this
  # journey is impossible; afterwards the model's answer — and nothing else —
  # lands on standard output.

  Rule: A configured Gemini model answers the operator's prompt

    Example: The operator asks a Gemini model and reads its answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "vertex-flash-3.8" that uses a Gemini model and answers with "4"
      When the operator asks tellme "What is two plus two?"
      Then tellme sends exactly one request to the provider "vertex-flash-3.8"
      And tellme prints the provider's answer "4"
      And tellme exits successfully

  Rule: The configured persona and output budget reach the Gemini model

    Example: A terse persona and a bounded output budget shape the request
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is "You are a terse assistant. Answer in one sentence."
      And the configuration selects a reachable provider "vertex-flash-3.8" that uses a Gemini model and answers directly
      And the selected provider entry allows at most 40960 output tokens
      When the operator asks tellme "What is two plus two?"
      Then what tellme sends to the model begins with the configured persona instruction
      And the request to the model allows no more than 40960 output tokens
      And the final answer is printed on standard output

  Rule: A Gemini model can ask tellme to run a tool before answering

    Example: The model reads a file with tellme's help, then answers
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "vertex-flash-3.8" that uses a Gemini model and answers using the tool results it receives
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What is the launch code? Read notes.txt to find out."
      Then tellme reads "notes.txt" using its read_files tool
      And the final answer states that the launch code is "ORANGE"
      And tellme exits successfully
