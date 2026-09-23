Feature: Inspecting several pictures in one step

  # Plan-side acceptance journey — round 083, anchor issue
  # https://github.com/gosharplite/tellme/issues/167. The OpenAI-compatible
  # provider family refuses a request whose assistant tool-call block is
  # interrupted by a non-tool message: it requires the tool results to be
  # answered together, before any other message. When the agent inspects SEVERAL
  # pictures in ONE step, it must deliver the pictures together, AFTER every tool
  # result, so the turn completes. Inspecting ONE picture is already correct and
  # must stay unchanged. This layer is PM-readable business language; the
  # executable interface truth lives under specs/truth/features/cli/**.

  Rule: A turn that inspects several pictures in one step completes

    Example: The model inspects three pictures in one step
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "a.png" that is a PNG picture
      And the workspace holds an image file "b.png" that is a PNG picture
      And the workspace holds an image file "c.png" that is a PNG picture
      And a configured provider "eye" that can take images and whose endpoint asks tellme, in one step, to read "a.png", "b.png" and "c.png" and then answers with "three squares"
      When the operator starts tellme with the prompt "What do these pictures show?"
      Then the tool results of the step were answered together, before the pictures were shown
      And tellme prints the provider's answer "three squares"
      And tellme exits successfully

    Example: The three pictures ride one message after the results
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "a.png" that is a PNG picture
      And the workspace holds an image file "b.png" that is a PNG picture
      And the workspace holds an image file "c.png" that is a PNG picture
      And a configured provider "eye" that can take images and whose endpoint asks tellme, in one step, to read "a.png", "b.png" and "c.png" and then answers with "three squares"
      When the operator starts tellme with the prompt "What do these pictures show?"
      Then the step showed the three pictures together after the results
      And tellme exits successfully

    Example: A resumed session replays the step and completes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "a.png" that is a PNG picture
      And the workspace holds an image file "b.png" that is a PNG picture
      And the workspace holds an image file "c.png" that is a PNG picture
      And the session history already holds a turn in which the agent read "a.png", "b.png" and "c.png"
      And a configured provider "eye" that can take images whose endpoint answers with "still three squares"
      When the operator starts tellme with the prompt "And now?"
      Then the replayed step answered every tool result together
      And tellme prints the provider's answer "still three squares"
      And tellme exits successfully

  Rule: One picture keeps today's shape

    Example: Reading a single picture is unchanged
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a PNG picture
      And a configured provider "eye" that can take images and whose endpoint asks tellme to read "shot.png" and then answers with "a red square"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request carried the image file "shot.png"
      And the image was attached as a "image/png" picture
      And tellme prints the provider's answer "a red square"
      And tellme exits successfully

  Rule: The Gemini family keeps its placement

    Example: A Gemini round that inspects two pictures completes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "a.png" that is a PNG picture
      And the workspace holds an image file "b.png" that is a PNG picture
      And a configured Gemini provider "eye" that can take images and whose endpoint asks tellme, in one step, to read the image files "a.png" and "b.png" and then answers with "two squares"
      When the operator starts tellme with the prompt "What do these pictures show?"
      Then the Gemini provider received the answer to both read requests together
      And tellme prints the provider's answer "two squares"
      And tellme exits successfully
