Feature: Preserving existing conversations across runs

  # Acceptance only: the tool-replay change is additive — a conversation that
  # used no tool, and a tool-using conversation on a provider that needs no
  # token, both resume exactly as before.

  Rule: A conversation that used no tool resumes unchanged

    Example: A plain exchange with a Gemini model resumes and is answered
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "vertex-flash-3.8" that uses a Gemini model and answers using the conversation it receives
      When the operator asks tellme "My name is Alice."
      And the operator starts tellme again and asks "What is my name?"
      Then the resumed run carries the earlier exchange "My name is Alice."
      And tellme prints the model's answer
      And tellme exits successfully

  Rule: A tool-using conversation on a provider that needs no token is unaffected

    Example: A tool-using exchange with an OpenAI-compatible model resumes and is answered
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the tool results it receives
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What is the launch code? Read notes.txt to find out."
      And the operator starts tellme again and asks "Restate the launch code you found."
      Then the resumed run replays the earlier tool step that read "notes.txt"
      And tellme exits successfully
