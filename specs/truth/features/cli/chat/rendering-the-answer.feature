Feature: Rendering the answer

  # Interface truth (CLI end, `chat` module) — the default output path renders the answer as formatted
  # Markdown on all streams (round-006 Clarify Q1, reference parity). The E2E suite always captures
  # stdout as a non-terminal pipe, so this pins that a redirected answer is still rendered.

  Rule: A successful turn renders the answer as formatted text by default

    Example: The answer's Markdown is rendered
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "**Bold** and _italic_"
      When the operator starts tellme with the prompt "Format this"
      Then the captured standard output is the rendered answer, not its raw Markdown
      And tellme exits successfully
