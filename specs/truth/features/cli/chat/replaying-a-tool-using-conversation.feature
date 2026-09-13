Feature: Replaying a tool-using conversation

  # Interface truth (CLI end, `chat` module) — a resumed conversation that used a tool on a
  # Vertex/Gemini provider replays the earlier tool step carrying its persisted provider token, so the
  # resume completes instead of failing at the provider. Acceptance journey:
  # features/acceptance/replaying-a-tool-using-conversation.feature.

  Rule: A resumed conversation replays the earlier tool step carrying its provider token

    Example: A Gemini file-read step is replayed with its token when the session resumes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose endpoint answers with "Noted"
      And the session history already holds a tool-using exchange carrying the provider token "sig-abc"
      When the operator starts tellme with the prompt "What did you find?"
      Then the request replayed the earlier tool step "read_files" carrying the provider token "sig-abc"
      And tellme exits successfully

  Rule: A resumed conversation on a provider that needs no token is unaffected

    Example: A tool-using step is replayed without a token when the session resumes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "Noted"
      And the session history already holds a tool-using exchange with no provider token
      When the operator starts tellme with the prompt "What did you find?"
      Then the request replayed the earlier tool step "read_files"
      And tellme exits successfully
