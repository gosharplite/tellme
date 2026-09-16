Feature: Refusing a reply cut off at the output limit

  # Interface truth (CLI end, `chat` module) — the provider output-cap truncation guard (round 030). A
  # provider that cuts the model's reply off at its output limit returns a partial reply; tellme must
  # refuse it (the frozen provider class phrase + the provider error code) rather than act on a
  # half-finished tool call or present a half-finished answer. **Both transport families are pinned**
  # (OpenAI-compatible `finish_reason:"length"` and Vertex/Gemini `finishReason:"MAX_TOKENS"`), on both
  # the tool-call and the text-only truncation sites. Acceptance journey:
  # features/acceptance/refusing-a-cut-off-reply.feature.

  Rule: A reply cut off while asking tellme to change a file is refused, and the file is untouched

    Example: A cut-off file creation writes nothing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains no file "notes.txt"
      And a configured provider "test-model" whose reply is cut off at the output limit while creating the file "notes.txt" with the content "hello"
      When the operator starts tellme with the prompt "Create notes.txt containing exactly the text 'hello'."
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme creates no file "notes.txt"
      And tellme exits with the provider error code

    Example: A cut-off file edit leaves the file as it was
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "config.txt" whose lines are:
        | alpha |
        | BETA  |
        | gamma |
      And a configured provider "test-model" whose reply is cut off at the output limit while editing the file "config.txt" replacing "BETA" with "beta"
      When the operator starts tellme with the prompt "In config.txt, replace the line 'BETA' with 'beta'."
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And the lines of "config.txt" are still:
        | alpha |
        | BETA  |
        | gamma |
      And tellme exits with the provider error code

    Example: A cut-off Gemini file creation writes nothing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains no file "docs.txt"
      And a configured Gemini provider "vertex-flash-3.8" whose reply is cut off at the output limit while creating the file "docs.txt" with the content "hi"
      When the operator starts tellme with the prompt "Create docs.txt containing exactly the text 'hi'."
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme creates no file "docs.txt"
      And tellme exits with the provider error code

  Rule: A reply cut off before the answer is finished is refused, not shown as complete

    Example: A cut-off answer is not presented as the answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose reply is cut off at the output limit before finishing its answer
      When the operator starts tellme with the prompt "What is the launch code? Answer in detail."
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme prints no answer
      And tellme exits with the provider error code

    Example: A cut-off Gemini answer is not presented as the answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "vertex-flash-3.8" whose reply is cut off at the output limit before finishing its answer
      When the operator starts tellme with the prompt "What is two plus two? Answer in detail."
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme prints no answer
      And tellme exits with the provider error code
