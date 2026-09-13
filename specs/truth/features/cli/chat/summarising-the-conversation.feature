Feature: Summarising the conversation with an agent tool

  # Interface truth (CLI end, `chat` module) — the deferred summarisation capability arrives as an
  # on-demand agent tool, not as automatic context pruning. Acceptance journey:
  # features/acceptance/summarising-the-conversation.feature.

  Rule: The model can summarise earlier history through a tool

    Example: A long conversation is condensed on demand
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      And a configured provider "test-model" whose endpoint asks tellme to summarise the conversation and then answers with "Summary: the operator is Alice."
      When the operator starts tellme with the prompt "Summarise our conversation so far."
      Then tellme summarised the earlier conversation using its summarise tool
      And the earlier conversation records are unchanged
      And tellme exits successfully
