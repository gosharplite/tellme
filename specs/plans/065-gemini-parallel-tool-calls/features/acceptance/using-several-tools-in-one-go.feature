Feature: Using several tools in one go with a Gemini provider

  # Acceptance only (round 065, anchor issue #132). Business language only.
  #
  # A live check found that a Gemini/Vertex turn in which the model asks for
  # more than one tool **at once** fails: the follow-up request is rejected and
  # the turn never finishes. A single request in a turn is unaffected. This
  # journey is about a turn that asks for several tools finishing normally, and
  # about a turn that reads several pictures showing the model all of them.

  Rule: A Gemini turn that asks for several tools in one go still finishes

    Example: The model reads two files in one go and answers about both
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "first.txt" whose text is "the first note"
      And the working directory contains a file "second.txt" whose text is "the second note"
      And a configured Gemini provider "eyes" whose endpoint records the request then answers "I read the first note and the second note"
      When the operator starts tellme with the prompt "Read first.txt and second.txt, then tell me what each says"
      And the model asks, in the same step, to read the file "first.txt" and to read the file "second.txt"
      Then the Gemini provider received the answer to both read requests together
      And tellme prints the provider's answer "I read the first note and the second note"
      And tellme exits successfully

    Example: The model reads three files in one go and the turn still finishes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "a.txt" whose text is "note a"
      And the working directory contains a file "b.txt" whose text is "note b"
      And the working directory contains a file "c.txt" whose text is "note c"
      And a configured Gemini provider "eyes" whose endpoint records the request then answers "three notes read"
      When the operator starts tellme with the prompt "Read a.txt, b.txt and c.txt in one go"
      And the model asks, in the same step, to read the file "a.txt", the file "b.txt" and the file "c.txt"
      Then the Gemini provider received the answer to all three read requests together
      And tellme prints the provider's answer "three notes read"
      And tellme exits successfully

    Example: The model reads two files again in a resumed session and the turn still finishes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "first.txt" whose text is "the first note"
      And the working directory contains a file "second.txt" whose text is "the second note"
      And a configured Gemini provider "eyes" whose endpoint records the request then answers "I read the first note and the second note"
      When the operator starts tellme with the prompt "Read first.txt and second.txt, then tell me what each says"
      And the model asks, in the same step, to read the file "first.txt" and to read the file "second.txt"
      And the provider answers "I read the first note and the second note"
      When the operator asks a follow-up in the same session
      Then the Gemini provider received the answer to both read requests together again
      And tellme exits successfully

  Rule: A Gemini turn that asks for several tools in one go still shows the model every picture

    Example: The model reads two pictures in one go and the answer reflects both
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "left.png" that is a PNG picture
      And the workspace holds an image file "right.png" that is a PNG picture
      And a configured Gemini provider "eyes" that can take images and whose endpoint records the request then answers "I can see a red square and a blue square"
      When the operator starts tellme with the prompt "What is in left.png and right.png?"
      And the model asks, in the same step, to read the image file "left.png" and to read the image file "right.png"
      Then the Gemini provider received the answer to both read requests together
      And both pictures reached the Gemini provider
      And tellme prints the provider's answer "I can see a red square and a blue square"
      And tellme exits successfully

  Rule: A Gemini turn that asks for a single tool is unchanged

    Example: A single file read still finishes as before
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "only.txt" whose text is "the only note"
      And a configured Gemini provider "eyes" whose endpoint records the request then answers "I read the only note"
      When the operator starts tellme with the prompt "Read only.txt"
      And the model asks to read the file "only.txt"
      Then the Gemini provider received the answer to the read request
      And tellme prints the provider's answer "I read the only note"
      And tellme exits successfully

    Example: A single picture still reaches the model as before
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a PNG picture
      And a configured Gemini provider "eyes" that can take images and whose endpoint records the request then answers "a red square"
      When the operator starts tellme with the prompt "What is in shot.png?"
      And the model asks to read the image file "shot.png"
      Then the Gemini provider received the answer to the read request
      And tellme prints the provider's answer "a red square"
      And tellme exits successfully

  Rule: A Gemini turn that reads several pictures in one go does not lose any of them

    Example: Two pictures whose text is the same kind are both carried
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "one.png" that is a PNG picture
      And the workspace holds an image file "two.png" that is a PNG picture
      And a configured Gemini provider "eyes" that can take images and whose endpoint records the request then answers "both shown"
      When the operator starts tellme with the prompt "Compare one.png and two.png"
      And the model asks, in the same step, to read the image file "one.png" and to read the image file "two.png"
      Then the request carried two pictures to the Gemini provider
      And tellme exits successfully
