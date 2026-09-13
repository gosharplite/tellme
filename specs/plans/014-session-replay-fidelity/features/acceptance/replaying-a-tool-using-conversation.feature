Feature: Replaying a tool-using conversation across runs

  # Acceptance only: what the operator experiences when tellme resumes a
  # conversation that used a tool on a Gemini model. Today such a resume fails
  # at the provider ("the provider request failed"); afterwards the earlier tool
  # step is replayed faithfully and the follow-up is answered.

  Rule: A resumed conversation replays the earlier tool step and is answered

    Example: A Gemini model reads a file, the operator quits, then resumes and asks again
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "vertex-flash-3.8" that uses a Gemini model and answers using the tool results it receives
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What is the launch code? Read notes.txt to find out."
      And the operator starts tellme again and asks "Restate the launch code you found."
      Then the resumed run replays the earlier tool step that read "notes.txt"
      And the final answer states that the launch code is "ORANGE"
      And tellme exits successfully
