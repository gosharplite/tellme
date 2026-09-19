Feature: Calling several tools in one round (Gemini)

  # Interface truth (CLI end, `chat` module) — a turn in which the Gemini model
  # asks for more than one tool at once must complete, and a turn that reads
  # several pictures must still show the model every one of them. Acceptance
  # journey: features/acceptance/using-several-tools-in-one-go.feature (round 065;
  # anchor issue #132). The round batches a round's tool results into one
  # `functionResponse` turn (ADR 0035) — the count Vertex checks per turn.

  Rule: A Gemini turn that asks for several tools in one go still finishes

    Example: The model reads two files in one step and answers about both
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "first.txt" whose text is "the first note"
      And the working directory contains a file "second.txt" whose text is "the second note"
      And a configured Gemini provider "eyes" whose endpoint asks tellme, in one step, to read "first.txt" and "second.txt" and then answers with "both notes read"
      When the operator starts tellme with the prompt "Read first.txt and second.txt"
      Then the Gemini provider received the answer to both read requests together
      And tellme prints the provider's answer "both notes read"
      And tellme exits successfully

  Rule: A Gemini turn that asks for several pictures in one go still shows the model every picture

    Example: The model reads two pictures in one step and both reach the Gemini wire
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "left.png" that is a PNG picture
      And the workspace holds an image file "right.png" that is a PNG picture
      And a configured Gemini provider "eyes" that can take images and whose endpoint asks tellme, in one step, to read the image files "left.png" and "right.png" and then answers with "both shown"
      When the operator starts tellme with the prompt "What is in left.png and right.png?"
      Then the Gemini provider received the answer to both read requests together
      And the request carried the image file "left.png"
      And the request carried the image file "right.png"
      And tellme prints the provider's answer "both shown"
      And tellme exits successfully

  Rule: A Gemini turn that asks for a single tool is unchanged

    Example: A single file read still finishes as before
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "only.txt" whose text is "the only note"
      And a configured Gemini provider "eyes" whose endpoint asks tellme to read "only.txt" and then answers with "the only note"
      When the operator starts tellme with the prompt "Read only.txt"
      Then tellme prints the provider's answer "the only note"
      And tellme exits successfully
